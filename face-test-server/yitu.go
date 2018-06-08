package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"github.com/levigross/grequests"
)

type yituLoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type yituLoginResponse struct {
	Rtn       int    `json:"rtn"`
	Message   string `json:"message"`
	SessionId string `json:"session_id"`
}

type yituCompareRequest struct {
	ImageBase641 string `json:"image_base64_1"`
	ImageBase642 string `json:"image_base64_2"`
	ImageType1   int    `json:"image_type_1"`
	ImageType2   int    `json:"image_type_2"`
}

type yituCompareResponse struct {
	Rtn        int     `json:"rtn"`
	Message    string  `json:"message"`
	Similarity float32 `json:"similarity"`
}

const (
	yituServerURL = "http://192.168.2.10:11180/business/api"
	yituUser      = "admin99"
	yituPassword  = "admin99"
)

type YituEngine struct {
	sessionId string
}

func (self *YituEngine) DisplayName() string {
	return "依图"
}

func (self *YituEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("YituEngine begin")
	if self.sessionId == "" {
		err := self.login()
		if err != nil {
			return 0, err
		}
	}

	resp, err := grequests.Post(yituServerURL+"/face/verify", &grequests.RequestOptions{
		JSON: &yituCompareRequest{
			ImageBase641: img1,
			ImageBase642: img2,
		},
		Headers: map[string]string{
			"session_id": self.sessionId,
		},
	})
	if err != nil {
		return 0, err
	}

	var respBody yituCompareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}

	if respBody.Rtn != 0 {
		return 0, fmt.Errorf("Operation failed: %s", respBody.Message)
	}

	return respBody.Similarity, nil
}

func (self *YituEngine) login() error {
	passwordMD5 := md5.Sum([]byte(yituPassword))

	resp, err := grequests.Post(yituServerURL+"/login", &grequests.RequestOptions{
		JSON: &yituLoginRequest{
			Name:     yituUser,
			Password: hex.EncodeToString(passwordMD5[:]),
		},
	})
	if err != nil {
		return err
	}

	var respBody yituLoginResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return err
	}

	if respBody.Rtn != 0 {
		return fmt.Errorf("Operation failed: %d %s", respBody.Rtn, respBody.Message)
	}

	self.sessionId = respBody.SessionId
	return nil
}
