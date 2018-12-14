package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/levigross/grequests"
)

// FaceVerifyResp : face verify resp
type FaceVerifyResp struct {
	Result       string  `json:"result"`
	TimeUserd    int     `json:"time_used"`
	Score        float64 `json:"score"`
	ErrorMessage string  `json:"errorMessage"`
}

func main() {
	loadFaceDir("/test/face_inner_pic")
}

func loadFaceDir(faceHomeDir string) {
	files, err := ioutil.ReadDir(faceHomeDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	fp, _ := os.OpenFile("/test/face_vfy.csv", os.O_RDWR|os.O_CREATE, 0755)
	fw := csv.NewWriter(fp)

	for _, file := range files {
		if file.IsDir() {
			// do Dir
			loadFaceFile(fw, fmt.Sprintf("%s/%s", faceHomeDir, file.Name()))
		} else {
			// do File
			fmt.Printf("no face dir : %s\n", file.Name())
		}
	}

	fw.Flush()
	fp.Close()
}

func loadFaceFile(fw *csv.Writer, faceDir string) {
	files, err := ioutil.ReadDir(faceDir)
	if err != nil {
		log.Fatal(err)
		return
	}

	var sensetimeRate string
	var faceFileNames []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".jpg") {
			faceFileNames = append(faceFileNames, file.Name())
		} else {
			sensetimeRate = file.Name()
		}
	}
	if len(faceFileNames) != 2 {
		log.Printf("pic file num is invalid, num=%d\n", len(faceFileNames))
		return
	}

	oneFileName := fmt.Sprintf("%s/%s", faceDir, faceFileNames[0])
	twoFileName := fmt.Sprintf("%s/%s", faceDir, faceFileNames[1])
	yituRate := verifyFace(oneFileName, twoFileName)
	log.Printf("st_rate: %s, yt_rate: %f, %s -> %s\n", sensetimeRate, yituRate, oneFileName, twoFileName)

	fw.Write([]string{sensetimeRate, fmt.Sprintf("%f", yituRate), oneFileName, twoFileName})
}

func verifyFace(oneFileName, twoFileName string) float64 {
	oneReader, _ := os.Open(oneFileName)
	twoReader, _ := os.Open(twoFileName)
	var images = []grequests.FileUpload{
		{
			FieldName:    "imageOne",
			FileName:     oneFileName,
			FileMime:     "multipart/form-data",
			FileContents: ioutil.NopCloser(oneReader),
		},
		{
			FieldName:    "imageTwo",
			FileName:     twoFileName,
			FileMime:     "multipart/form-data",
			FileContents: ioutil.NopCloser(twoReader),
		},
	}

	url := "http://192.168.2.166:2180/yitu-atom/verify/face/verification"
	resp, err := grequests.Post(url, &grequests.RequestOptions{Files: images})
	if err != nil {
		log.Println("Cannot post: ", err)
		return 0.0
	}
	if resp.Ok != true {
		log.Println("Request did not return OK")
		return 0.0
	}
	log.Println(resp.String())

	var ret FaceVerifyResp
	err = json.Unmarshal(resp.Bytes(), &ret)
	if err != nil {
		log.Fatal(err.Error())
		return 0.0
	}
	if ret.Result == "success" {
		return ret.Score
	}
	return 0.0
}
