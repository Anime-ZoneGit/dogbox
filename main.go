package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type Filename struct {
	Name string `uri:"name" binding:"required"`
}

func main() {
	v := viper.New()
	config := LoadConfig(v, ".env")

	dc, err := CreateController(config)
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
