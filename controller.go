package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/Fekinox/dogbox-main/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type DogboxController struct {
	db *db.Queries
	conn *pgx.Conn
}

func CreateController(connString string) (*DogboxController, error) {
	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, err
	}

	q := db.New(conn)

	return &DogboxController{
		db: q,
		conn: conn,
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

	c.JSON(200, gin.H{
		"file": filename.Name,
	})
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
		c.JSON(400, gin.H{"msg": err.Error()})
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
