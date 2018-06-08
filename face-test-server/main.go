package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

type CompareRequest struct {
	Image1 string `form:"image1" json:"image1"`
	Image2 string `form:"image2" json:"image2"`
}

type Engine interface {
	DisplayName() string
	Compare(img1, img2 string) (float32, error)
}

var (
	allEngines = []Engine{
		&CloudwalkEngine{},
		&YituEngine{},
		//&DswdEngine{},
		&SenseTimeEngine{},
		&FaceplusplusEngine{},
		&BaiduEngine{},
		&YoutuEngine{},
	}
)

func main() {
	exePath, err := os.Executable()
	log.Println(exePath)
	if err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.Static("/static", filepath.Join(filepath.Dir(exePath), "static"))

	router.POST("/api/compare", func(ctx *gin.Context) {
		log.Println("------compare------")
		var reqBody CompareRequest
		err := ctx.ShouldBind(&reqBody)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		results := make(map[string]interface{})

		for _, engine := range allEngines {
			similarity, err := engine.Compare(reqBody.Image1, reqBody.Image2)
			if err == nil {
				results[engine.DisplayName()] = fmt.Sprintf("%.2f%%", similarity)
			} else {
				results[engine.DisplayName()] = err.Error()
			}
		}

		ctx.JSON(http.StatusOK, results)
	})

	router.Run()
}
