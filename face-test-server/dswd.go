package main

import (
	"fmt"
	"log"
	"github.com/levigross/grequests"
)

type dswdCompareRequest struct {
	ImageBase64_1 string `json:"image_base64_1"`
	ImageBase64_2 string `json:"image_base64_2"`
}

type dswdCompareResponse struct {
	Rtn        int     `json:"rtn"`
	Message    string  `json:"message"`
	Similarity float32 `json:"similarity"`
}

const (
	dswdServerURL = "http://119.29.147.210:16781"
)

type DswdEngine struct {
}

func (self *DswdEngine) DisplayName() string {
	return "先达"
}

func (self *DswdEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("DswdEngine begin")
	resp, err := grequests.Post(dswdServerURL+"/face/verify", &grequests.RequestOptions{
		JSON: &dswdCompareRequest{
			ImageBase64_1: img1,
			ImageBase64_2: img2,
		},
	})
	if err != nil {
		return 0, err
	}

	var respBody dswdCompareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}

	if respBody.Rtn != 0 {
		return 0, fmt.Errorf("Operation failed: %d %s", respBody.Rtn, respBody.Message)
	}

	return respBody.Similarity, nil
}
