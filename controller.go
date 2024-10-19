package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	db "github.com/Fekinox/dogbox-main/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/sqids/sqids-go"
)

const DATA_DIR = "data"

type DogboxController struct {
	db *db.Queries
	cfg Config
	conn *pgx.Conn
	sqids *sqids.Sqids

	pwd string
}

func (dc *DogboxController) getImagePath(name string) string {
	return filepath.Join(dc.pwd, DATA_DIR, "images", name)
}

func (dc *DogboxController) genDeletionKey(id int64) (string, error) {
	return dc.sqids.Encode([]uint64{uint64(id)})
}

func CreateController(cfg Config) (*DogboxController, error) {
	conn, err := pgx.Connect(context.Background(), cfg.DBUrl)
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

	return &DogboxController{
		db: q,
		cfg: cfg,
		conn: conn,
		pwd: wd,

		sqids: s,
	}, nil
}

func (dc *DogboxController) Close() error {
	return dc.conn.Close(context.Background())
}

func (dc *DogboxController) GetFile(c *gin.Context) {
	var filename Filename
	if err := c.ShouldBindUri(&filename); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	name := filename.Name
	sq := strings.TrimSuffix(name, filepath.Ext(name))
	id := int64(dc.sqids.Decode(sq)[0])

	p, err := dc.db.GetPost(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": err.Error()})
	}

	imPath := dc.getImagePath(p.Filename)
	
	c.File(imPath)
}

func (dc *DogboxController) CreateFile(c *gin.Context) {
	fmt.Println(c.ContentType())
	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	tx, err := dc.conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"msg": err.Error()})
		return
	}
	defer tx.Rollback(c.Request.Context())
	qtx := dc.db.WithTx(tx)

	i, err := qtx.CreatePost(c.Request.Context(), db.CreatePostParams{
		Filename: "newfile",
		DeletionKey: "delkey",
		Hash: "hash",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	ident, err := dc.sqids.Encode([]uint64{uint64(i.ID)})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	ext := filepath.Ext(data.Filename)
	filename := ident + ext

	imPath := dc.getImagePath(filename)

	err = os.MkdirAll(filepath.Dir(imPath), 0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	srcFile, err := data.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	dstFile, err := os.Create(imPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	hasher := sha256.New()

	w := io.MultiWriter(
		dstFile,
		hasher,
	)

	if _, err := io.Copy(w, srcFile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	hashString := hex.EncodeToString(hasher.Sum(nil))
	dKey, err := dc.genDeletionKey(i.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	final, err := qtx.UpdatePost(c.Request.Context(), db.UpdatePostParams{
		Filename: &filename,
		DeletionKey: &dKey,
		Hash: &hashString,
		ID: i.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	if err = tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": final,
	})
}

func (dc *DogboxController) DeleteFile(c *gin.Context) {
	var filename Filename
	if err := c.ShouldBindUri(&filename); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"deleted": filename.Name,
	})
}
