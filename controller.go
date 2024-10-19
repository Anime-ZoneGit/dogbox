package main

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Fekinox/dogbox-main/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const DATA_DIR = "data"

type DogboxController struct {
	db *db.Queries
	conn *pgx.Conn

	pwd string
}

func (dc *DogboxController) getImagePath(name string) string {
	return filepath.Join(dc.pwd, DATA_DIR, "images", name)
}

func CreateController(connString string) (*DogboxController, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return &DogboxController{
		db: q,
		conn: conn,
		pwd: wd,
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
	ident := strings.TrimSuffix(name, filepath.Ext(name))

	p, err := dc.db.GetFile(c.Request.Context(), ident)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": err.Error()})
	}

	imPath := dc.getImagePath(p.Filename)
	
	c.File(imPath)
}

func (dc *DogboxController) CreateFile(c *gin.Context) {
	data, err := c.FormFile("data")
	if err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	_ = data

	ts := time.Now()

	ident := uuid.NewString()
	del := uuid.NewString()
	ext := filepath.Ext(data.Filename)
	filename := ident + ext

	imPath := dc.getImagePath(filename)
	err = os.MkdirAll(filepath.Dir(imPath), 0755)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	err = c.SaveUploadedFile(data, imPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	i, err := dc.db.UploadFile(c.Request.Context(), db.UploadFileParams{
		Filename: filename,
		Identifier: ident,
		Uploaddate: pgtype.Timestamptz{
			Time: ts,
			InfinityModifier: pgtype.Finite,
			Valid: true,
		},
		Deletetoken: del,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": err.Error()})
		return
	}

	c.JSON(201, gin.H{
		"message": i,
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
