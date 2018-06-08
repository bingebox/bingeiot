// +build ignore

package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/go-opencv/go-opencv/opencv"
)

// #cgo CPPFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -Wl,-rpath,${SRCDIR}/lib -L${SRCDIR}/lib -lFX
// #include "fx_api.h"
import "C"

func decodeImage(b64 string) (*C.RS_IMG, *opencv.IplImage, error) {
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, nil, err
	}

	iplImg := opencv.DecodeImageMem(data)
	if iplImg == nil {
		return nil, nil, errors.New("DecodeImageMem failed")
	}

	return &C.RS_IMG{
		C.int(iplImg.Width()),
		C.int(iplImg.Height()),
		C.int(iplImg.WidthStep()),
		C.int(iplImg.ImageSize()),
		C.PIX_FORMAT_BGR888,
		0,
		(*C.uchar)(iplImg.ImageData()),
	}, iplImg, nil
}

type ReadSenseEngine struct {
	mu sync.Mutex
}

func (self *ReadSenseEngine) DisplayName() string {
	return "阅面"
}

func (self *ReadSenseEngine) Compare(img1, img2 string) (float32, error) {
	rsImg1, iplImg1, err := decodeImage(img1)
	if err != nil {
		return 0, err
	}
	defer iplImg1.Release()

	rsImg2, iplImg2, err := decodeImage(img2)
	if err != nil {
		return 0, err
	}
	defer iplImg2.Release()

	self.mu.Lock()
	defer self.mu.Unlock()

	var feature1 [512]C.float
	var featureVersion1 C.int = 0
	rc := C.fxExtractFeature(*rsImg1, &feature1[0], &featureVersion1)
	if rc < 0 {
		return 0, fmt.Errorf("fxExtractFeature(img1) failed: %d", rc)
	}

	var feature2 [512]C.float
	var featureVersion2 C.int = 0
	rc = C.fxExtractFeature(*rsImg2, &feature2[0], &featureVersion2)
	if rc < 0 {
		return 0, fmt.Errorf("fxExtractFeature(img2) failed: %d", rc)
	}

	return float32(C.fxFaceVerification(&feature1[0], &feature2[0])), nil
}

func init() {
	rc := C.fxInit()
	if rc < 0 {
		log.Fatalf("fxInit() failed: %d", rc)
	}

	rc = C.fxSetRunMode(C.FX_RUNMODE_ENROLL)
	if rc < 0 {
		log.Fatalf("fxSetRunMode() failed: %d", rc)
	}
}
