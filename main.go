package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

type Filename struct {
	Name string `uri:"name" binding:"required"`
}

func main() {
	connString := "postgres://postgres:postgres@localhost:5432/postgres"

	dc, err := CreateController(connString)
	if err != nil {
		log.Fatal(err)
	}
	defer dc.Close()

	r := gin.Default()

	r.GET("/:name", dc.GetFile)
	r.POST("/", dc.CreateFile)
	r.DELETE("/:name", dc.DeleteFile)

	r.Run() // listen and serve on 0.0.0.0:8080
}
