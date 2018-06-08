package main

import (
	"fmt"
	"log"

	"github.com/levigross/grequests"
)

type compareResponse struct {
	ErrorMessages string  `json:"error_message"`
	Score         float32 `json:"confidence"`
	TimeUsed      int     `json:"time_used"`
	RequestID     string  `json:"request_id"`
}

const (
	serverURL = "https://api-cn.faceplusplus.com"
)

// FaceplusplusEngine :
type FaceplusplusEngine struct {
}

// DisplayName : 模块名
func (face *FaceplusplusEngine) DisplayName() string {
	return "Face++"
}

// Compare : 比对
func (face *FaceplusplusEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("FaceplusplusEngine begin")

	resp, err := grequests.Post(serverURL+"/facepp/v3/compare", &grequests.RequestOptions{
		Data: map[string]string{
			"api_key":        "PVKLhUYStvjYwXLrGMIsGGLPNzPHVVYA",
			"api_secret":     "rZ6NgnzqdPsujVidBjChDD2v8-4JSJS9",
			"image_base64_1": img1,
			"image_base64_2": img2,
		},
	})
	if err != nil {
		return 0, err
	}

	var respBody compareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}
	log.Println(respBody)

	if respBody.ErrorMessages != "" {
		return 0, fmt.Errorf("Operation failed: %s", respBody.ErrorMessages)
	}

	return respBody.Score, nil
}
