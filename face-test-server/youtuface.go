package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/ochapman/youtu"
)

const (
	appID     uint32 = 10135041
	secretID  string = "AKIDA2kZIqBNsHUWNEvcP8i7KMI8qwxs6Hk2"
	secretKey string = "Gf1IUWM1GcqtkRwqgt5KNFvsxrZ9Zp9W"
	userID    string = "youtu_75405_20180604115524_514"
)

// YoutuEngine :
type YoutuEngine struct {
	yt *youtu.Youtu
}

// DisplayName : 模块名
func (face *YoutuEngine) DisplayName() string {
	return "腾讯优图"
}

// Compare : 比对
func (face *YoutuEngine) Compare(imgBase641, imgBase642 string) (float32, error) {
	log.Println("YoutuEngine begin")
	err := face.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadFile() failed: %s\n", err)
		return 0, err
	}

	resp, err := face.yt.CompareBase64(imgBase641, imgBase642)
	if err != nil {
		return 0, err
	}

	if resp.ErrorCode != 0 {
		return 0, fmt.Errorf("Operation failed: %s", resp.ErrorMsg)
	}

	return resp.Similarity, nil
}

// Init :
func (face *YoutuEngine) Init() error {
	as, err := youtu.NewAppSign(appID, secretID, secretKey, userID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewAppSign() failed: %s\n", err)
		return err
	}

	//yt := youtu.Init(as, youtu.TencentYunHost)
	face.yt = youtu.Init(as, youtu.DefaultHost)
	return nil
}

// Detect :
func (face *YoutuEngine) Detect() {
	err := face.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadFile() failed: %s\n", err)
		return
	}

	imgData, err := ioutil.ReadFile("./pic/a_fullview.jpg")
	if err != nil {
		fmt.Fprintf(os.Stderr, "ReadFile() failed: %s\n", err)
		return
	}

	df, err := face.yt.DetectFace(imgData, false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DetectFace() failed: %s", err)
		return
	}
	fmt.Printf("df: %#v\n", df)
}
