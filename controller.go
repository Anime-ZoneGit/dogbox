package main

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	db "github.com/Fekinox/dogbox-main/db/sqlc"
	store "github.com/Fekinox/dogbox-main/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/sqids/sqids-go"
)

type DogboxController struct {
	db     *db.Queries
	cfg    Config
	router *gin.Engine
	conn   *pgx.Conn
	sqids  *sqids.Sqids

	store store.Store

	pwd string
}

var (
	BadRequestError = errors.New("Bad request")
	NotFoundError   = func(name string) error {
		return errors.New(fmt.Sprintf("Not found: %s", name))
	}
)

func (dc *DogboxController) getImagePath(name string) string {
	return filepath.Join("images", name)
}

func (dc *DogboxController) genDeletionKey(id int64) (string, error) {
	return dc.sqids.Encode([]uint64{uint64(id)})
}

func CreateController(cfg Config) (*DogboxController, error) {
	var engine *gin.Engine
	if cfg.Environment == "test" || cfg.Environment == "release" {
		gin.SetMode(gin.ReleaseMode)
		engine = gin.New()
	} else {
		engine = gin.Default()
	}

	conn, err := pgx.Connect(context.Background(), cfg.GetDBUrl())
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	s, err := sqids.New(
		sqids.Options{
			MinLength: 10,
		},
	)
	if err != nil {
		return nil, err
	}

	store := store.MakeLocalStore(cfg.DogboxDataDir)

	return &DogboxController{
		db:     q,
		cfg:    cfg,
		conn:   conn,
		router: engine,
		pwd:    wd,
		sqids:  s,

		store: store,
	}, nil
}

func (dc *DogboxController) MountHandlers() {
	api := dc.router.Group("/api")

	posts := api.Group("/posts")
	// posts.Use(Timeout(60 * time.Second))
	posts.Use(ErrorHandler(&dc.cfg))

	posts.GET(
		":name",
		RateLimiter(100, 25),
		dc.GetFile,
	)
	posts.POST(
		"",
		ApiKeyMiddleware(&dc.cfg),
		RateLimiter(20, 5),
		dc.CreateFile,
	)
	posts.DELETE(
		":name",
		ApiKeyMiddleware(&dc.cfg),
		RateLimiter(20, 5),
		dc.DeleteFile,
	)
}

func (dc *DogboxController) Start(addr string) error {
	return dc.router.Run(addr)
}

func (dc *DogboxController) Router() *gin.Engine {
	return dc.router
}

func (dc *DogboxController) Close() error {
	return dc.conn.Close(context.Background())
}

func (dc *DogboxController) GetFile(c *gin.Context) {
	name := c.Param("name")
	sq := dc.sqids.Decode(strings.TrimSuffix(name, filepath.Ext(name)))
	if len(sq) == 0 {
		c.AbortWithError(http.StatusBadRequest, BadRequestError)
		return
	}
	id := int64(sq[0])

	p, err := dc.db.GetPost(c.Request.Context(), id)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, NotFoundError(name))
		return
	}

	modTimeMd5 := md5.Sum([]byte(p.UpdatedAt.Time.String()))
	modTimeString := fmt.Sprintf("%x", modTimeMd5)

	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("Etag", modTimeString)

	if match := c.GetHeader("If-None-Match"); match != "" {
		if strings.Contains(match, modTimeString) {
			c.Status(http.StatusNotModified)
			return
		}
	}

	imPath := dc.getImagePath(p.Filename)

	reader, err := dc.store.Retrieve(imPath)
	if err != nil {
		c.AbortWithError(http.StatusNotFound, NotFoundError(name))
		return
	}

	tmpFile, err := store.CreateTempFile(filepath.Join(
		dc.cfg.DogboxDataDir,
		"images",
		"tmp",
		p.Filename,
	))
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer tmpFile.Cleanup()

	_, err = io.Copy(tmpFile, reader)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	http.ServeContent(
		c.Writer,
		c.Request,
		p.Filename,
		p.UpdatedAt.Time,
		tmpFile,
	)
}

func (dc *DogboxController) CreateFile(c *gin.Context) {
	data, err := c.FormFile("data")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, BadRequestError)
		return
	}

	final, err := dc.uploadToStore(c.Request.Context(), data, dc.store)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, BadRequestError)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": final,
	})
}

func (dc *DogboxController) DeleteFile(c *gin.Context) {
	name := c.Param("name")

	c.JSON(http.StatusNoContent, gin.H{
		"deleted": name,
	})
}

func (dc *DogboxController) uploadToStore(
	ctx context.Context,
	data *multipart.FileHeader,
	st store.Store,
) (*db.Post, error) {
	dataFilename := filepath.Base(data.Filename)
	tx, err := dc.conn.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	qtx := dc.db.WithTx(tx)

	i, err := qtx.CreatePost(ctx, db.CreatePostParams{
		Filename:    "newfile",
		DeletionKey: "delkey",
		Hash:        "hash",
	})
	if err != nil {
		return nil, err
	}

	ident, err := dc.sqids.Encode([]uint64{uint64(i.ID)})
	if err != nil {
		return nil, err
	}

	ext := filepath.Ext(dataFilename)
	filename := ident + ext

	imPath := dc.getImagePath(filename)

	srcFile, err := data.Open()
	if err != nil {
		return nil, err
	}
	defer srcFile.Close()

	dstWriter := store.NewWriter(st, imPath)
	defer dstWriter.Close()

	hasher := sha256.New()

	w := io.MultiWriter(
		dstWriter,
		hasher,
	)

	if _, err := io.Copy(w, srcFile); err != nil {
		return nil, err
	}

	hashString := hex.EncodeToString(hasher.Sum(nil))
	dKey, err := dc.genDeletionKey(i.ID)
	if err != nil {
		return nil, err
	}

	final, err := qtx.UpdatePost(ctx, db.UpdatePostParams{
		Filename:    &filename,
		DeletionKey: &dKey,
		Hash:        &hashString,
		ID:          i.ID,
	})
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}

	return final, nil
}
