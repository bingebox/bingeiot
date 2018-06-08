package main

import (
	"crypto/md5"
	"encoding/hex"
	"flag"
	"log"
	"time"

	"github.com/levigross/grequests"
	"github.com/sunfmin/fanout"
)

type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Rtn       int    `json:"rtn"`
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

type RetrieveRequest_Retrieval struct {
	FaceImageID   string   `json:"face_image_id"`
	RepositoryIDs []string `json:"repository_ids"`
	Threshold     float64  `json:"threshold"`
}

type RetrieveRequest_Condition struct {
}

type RetrieveRequest_Order struct {
	Similarity int `json:"similarity"`
}

type RetrieveRequest struct {
	Retrieval RetrieveRequest_Retrieval `json:"retrieval"`
	Condition RetrieveRequest_Condition `json:"condition"`
	Order     RetrieveRequest_Order     `json:"order"`
	Start     int                       `json:"start"`
	Limit     int                       `json:"limit"`
}

type RetrieveResponse struct {
	Rtn     int    `json:"rtn"`
	Message string `json:"message"`
	Total   int    `json:"total"`
}

var (
	flagServer       = flag.String("server", "http://192.168.2.10:11180", "")
	flagUser         = flag.String("user", "admin99", "")
	flagPassword     = flag.String("password", "admin99", "")
	flagFaceImageID  = flag.String("face-image-id", "844424930131973@DEFAULT", "")
	flagRepositoryID = flag.String("repository-id", "3@DEFAULT", "")
	flagIterations   = flag.Int("iterations", 10000, "")
	flagConcurrency  = flag.Int("concurrency", 1, "")
)

func login() string {
	passwordMD5 := md5.Sum([]byte(*flagPassword))

	resp, err := grequests.Post(*flagServer+"/business/api/login", &grequests.RequestOptions{
		JSON: &LoginRequest{
			Name:     *flagUser,
			Password: hex.EncodeToString(passwordMD5[:]),
		},
	})
	if err != nil {
		log.Panic(err)
	}

	var body LoginResponse
	err = resp.JSON(&body)
	if err != nil {
		log.Panic(err)
	}

	if body.Rtn != 0 {
		log.Panicf("Request failed: %d %s", body.Rtn, body.Message)
	}

	return body.SessionID
}

func testRetrieval(sessionID string) {
	inputs := make([]interface{}, *flagIterations)
	for i := 0; i < *flagIterations; i++ {
		inputs[i] = i
	}

	startTime := time.Now()

	_, err := fanout.ParallelRun(*flagConcurrency, fanout.Worker(func(input interface{}) (interface{}, error) {
		resp, err := grequests.Post(*flagServer+"/business/api/retrieval", &grequests.RequestOptions{
			JSON: &RetrieveRequest{
				Retrieval: RetrieveRequest_Retrieval{
					FaceImageID:   *flagFaceImageID,
					RepositoryIDs: []string{*flagRepositoryID},
					Threshold:     90,
				},
				Order: RetrieveRequest_Order{
					Similarity: -1,
				},
				Start: 0,
				Limit: 100,
			},
			Headers: map[string]string{
				"session_id": sessionID,
			},
		})
		if err != nil {
			log.Panic(err)
		}

		var body RetrieveResponse
		err = resp.JSON(&body)
		if err != nil {
			log.Panic(err)
		}

		if body.Rtn != 0 {
			log.Panicf("Request failed: %d %s", body.Rtn, body.Message)
		}

		log.Printf("Retrieval#%08d: %d record(s) found", input.(int), body.Total)
		return input, nil
	}), inputs)

	if err != nil {
		log.Panic(err)
	}

	endTime := time.Now()
	log.Printf("Retrieval: %.3f qps", float64(*flagIterations)/endTime.Sub(startTime).Seconds())
}

func main() {
	flag.Parse()

	sessionID := login()
	testRetrieval(sessionID)
}
