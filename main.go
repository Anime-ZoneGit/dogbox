package main

import (
	"context"
	"log"
	"path/filepath"
	"time"

	"github.com/Fekinox/dogbox-main/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type Filename struct {
	Name string `uri:"name" binding:"required"`
}

func main() {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/postgres")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)

	q := db.New(conn)

	r := gin.Default()
	r.GET("/:name", func(c *gin.Context) {
		var filename Filename
		if err := c.ShouldBindUri(&filename); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"file": filename.Name,
		})
	})

	r.POST("/", func(c *gin.Context) {
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

		i, err := q.UploadFile(ctx, db.UploadFileParams{
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
	})

	r.DELETE("/:name", func(c *gin.Context) {
		var filename Filename
		if err := c.ShouldBindUri(&filename); err != nil {
			c.JSON(400, gin.H{"msg": err.Error()})
			return
		}

		c.JSON(200, gin.H{
			"deleted": filename.Name,
		})
	})

	r.Run() // listen and serve on 0.0.0.0:8080
}
