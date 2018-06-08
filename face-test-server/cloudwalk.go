package main

import (
	"fmt"
	"log"
	"github.com/levigross/grequests"
)

type cloudwalkCompareResponse struct {
	Result int     `json:"result"`
	Score  float32 `json:"score"`
	Info   string  `json:"info`
}

const (
	cloudwalkServerURL = "http://120.24.168.188:7000"
)

type CloudwalkEngine struct {
}

func (self *CloudwalkEngine) DisplayName() string {
	return "云从"
}

func (self *CloudwalkEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("CloudwalkEngine begin")
	resp, err := grequests.Post(cloudwalkServerURL+"/face/tool/compare", &grequests.RequestOptions{
		Data: map[string]string{
			"app_id":     "2441039@qq.com",
			"app_secret": "2813275897364251927",
			"imgA":       img1,
			"imgB":       img2,
		},
	})
	if err != nil {
		return 0, err
	}

	var respBody cloudwalkCompareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}

	if respBody.Result != 0 {
		return 0, fmt.Errorf("Operation failed: %d %s", respBody.Result, respBody.Info)
	}

	return respBody.Score * 100, nil
}
