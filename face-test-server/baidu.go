package main

import (
	"fmt"
	"log"

	"github.com/levigross/grequests"
)

type accessTokenResponse struct {
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int    `json:"expires_in"`
	Scope            string `json:"scope"`
	SessionKey       string `json:"session_key"`
	AccessToken      string `json:"access_token"`
	SessionSecret    string `json:"session_secret"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type baiduCompareRequest struct {
	Image           string `json:"image"`
	ImageType       string `json:"image_type"`
	FaceType        string `json:"face_type"`
	QualityControl  string `json:"quality_control"`
	LivenessControl string `json:"liveness_control"`
}

type baiduCompareResponse struct {
	ErrorCode int                `json:"error_code"`
	ErrorMsg  string             `json:"error_msg"`
	Result    baiduCompareResult `json:"result"`
}

type baiduCompareResult struct {
	Score float32 `json:"score"`
}

const (
	baiduServerURL = "https://aip.baidubce.com"
)

// BaiduEngine :
type BaiduEngine struct {
}

// DisplayName : 模块名
func (face *BaiduEngine) DisplayName() string {
	return "百度"
}

func (face *BaiduEngine) getAccessToken() (string, error) {
	APIKey := "wO6N1G4dZGeRT0yyqNv41pLE"
	secretKey := "0eFUbwA1LydNFR74irpmPzZ1ERqpvWwN"
	url := baiduServerURL + "/oauth/2.0/token?grant_type=client_credentials&client_id=" + APIKey + "&client_secret=" + secretKey
	resp, err := grequests.Post(url, &grequests.RequestOptions{
		Headers: map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
		},
	})
	if err != nil {
		return "", err
	}

	var respBody accessTokenResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return "", err
	}
	log.Println(respBody)
	return respBody.AccessToken, nil
}

// Compare : 比对
func (face *BaiduEngine) Compare(img1, img2 string) (float32, error) {
	log.Println("BaiduEngine begin")

	accessToken, err := face.getAccessToken()
	if err != nil {
		return 0, err
	}
	log.Printf("accessToken=%s\n", accessToken)

	var compareRequests []baiduCompareRequest
	compareRequests = append(compareRequests, baiduCompareRequest{
		Image:           img1,
		ImageType:       "BASE64",
		FaceType:        "LIVE",
		QualityControl:  "LOW",
		LivenessControl: "LOW",
	})
	compareRequests = append(compareRequests, baiduCompareRequest{
		Image:           img2,
		ImageType:       "BASE64",
		FaceType:        "LIVE",
		QualityControl:  "LOW",
		LivenessControl: "LOW",
	})

	url := baiduServerURL + "/rest/2.0/face/v3/match?access_token=" + accessToken
	resp, err := grequests.Post(url, &grequests.RequestOptions{
		JSON: compareRequests,
		Headers: map[string]string{
			"Content-Type": "application/json; charset=UTF-8",
		},
	})
	if err != nil {
		return 0, err
	}
	log.Println(resp.String())

	var respBody baiduCompareResponse
	err = resp.JSON(&respBody)
	if err != nil {
		return 0, err
	}
	log.Println(respBody)

	if respBody.ErrorCode != 0 {
		return 0, fmt.Errorf("Operation failed:%d, %s", respBody.ErrorCode, respBody.ErrorMsg)
	}

	return respBody.Result.Score, nil
}
