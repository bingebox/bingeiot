package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"github.com/levigross/grequests"
)

type sensetimeCompareResponse struct {
	Result   string  `json:"result"`
	Score    float32 `json:"score"`
	TimeUsed int     `json:"time_used"`
}

const (
	sensetimeServerURL = "http://192.168.2.14:9001"
)

type SenseTimeEngine struct {
}

func (self *SenseTimeEngine) DisplayName() string {
	return "商汤"
}

func (self *SenseTimeEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("SenseTimeEngine begin")
	imgData1, err := base64.StdEncoding.DecodeString(img1)
	if err != nil {
		return 0, err
	}

	imgData2, err := base64.StdEncoding.DecodeString(img2)
	if err != nil {
		return 0, err
	}

	imgBuf1 := ioutil.NopCloser(bytes.NewBuffer(imgData1))
	imgBuf2 := ioutil.NopCloser(bytes.NewBuffer(imgData2))

	resp, err := grequests.Post(sensetimeServerURL+"/verify/face/verification", &grequests.RequestOptions{
		Files: []grequests.FileUpload{
			{"imageOne.jpg", imgBuf1, "imageOne", "image/jpeg"},
			{"imageTwo.jpg", imgBuf2, "imageTwo", "image/jpeg"},
		},
	})
	if err != nil {
		return 0, err
	}

	var respBody sensetimeCompareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}

	if respBody.Result != "success" {
		return 0, fmt.Errorf("Operation failed: %s", respBody.Result)
	}

	return respBody.Score * 100, nil
}
