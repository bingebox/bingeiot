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

// LoginRequest : 登录请求
type LoginRequest struct {
	Name     string `json:"name"`
	Password string `json:"password"`
}

// LoginResponse : 登录应答
type LoginResponse struct {
	Rtn       int    `json:"rtn"`
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// RetrieveRequestRetrieval :
type RetrieveRequestRetrieval struct {
	FaceImageID   string   `json:"face_image_id"`
	RepositoryIDs []string `json:"repository_ids"`
	Threshold     float64  `json:"threshold"`
}

// RetrieveRequestCondition :
type RetrieveRequestCondition struct {
}

// RetrieveRequestOrder :
type RetrieveRequestOrder struct {
	Similarity int `json:"similarity"`
}

// RetrieveRequest : 查询请求
type RetrieveRequest struct {
	Retrieval RetrieveRequestRetrieval `json:"retrieval"`
	Condition RetrieveRequestCondition `json:"condition"`
	Order     RetrieveRequestOrder     `json:"order"`
	Start     int                      `json:"start"`
	Limit     int                      `json:"limit"`
}

// RetrieveResponse : 查询应答
type RetrieveResponse struct {
	Rtn     int    `json:"rtn"`
	Message string `json:"message"`
	Total   int    `json:"total"`
}

// ConditionQueryReq : 条件查询请求
type ConditionQueryReq struct {
	ExtraFields []string                `json:"extra_fields"`
	Condition   conditionQueryCondition `json:"condition"`
	Order       order                   `json:"order"`
	Start       int                     `json:"start"`
	Limit       int                     `json:"limit"`
}

type conditionQueryCondition struct {
	RepositoryID repositoryID `json:"repository_id"`
	Timestamp    timestamp    `json:"timestamp"`
}

type repositoryID struct {
	In []string `json:"$in"`
}

type timestamp struct {
	BeginTime int64 `json:"$gte"`
	EndTime   int64 `json:"$lte"`
}

type order struct {
	Timestamp int `json:"timestamp"`
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
				Retrieval: RetrieveRequestRetrieval{
					FaceImageID:   *flagFaceImageID,
					RepositoryIDs: []string{*flagRepositoryID},
					Threshold:     90,
				},
				Order: RetrieveRequestOrder{
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

// 用人脸检索接口导出历史纪录列表
func searchFaceList(sessionID string) {
	var beginTime int64 = 1
	var endTime int64 = 1527578920
	resp, err := grequests.Post(*flagServer+"/business/api/condition/query", &grequests.RequestOptions{
		JSON: &ConditionQueryReq{
			ExtraFields: []string{"place_code", "capture_time"},
			Condition: conditionQueryCondition{
				RepositoryID: repositoryID{
					In: []string{"353@DEFAULT", "4@DEFAULT"},
				},
				Timestamp: timestamp{
					BeginTime: beginTime,
					EndTime:   endTime,
				},
			},
			Order: order{Timestamp: -1},
			Start: 0,
			Limit: 1,
		},
		Headers: map[string]string{
			"session_id": sessionID,
		},
	})
	if err != nil {
		log.Panic(err)
	}
	log.Println(string(resp.Bytes()))
}

func _main() {
	flag.Parse()

	sessionID := login()
	//testRetrieval(sessionID)
	searchFaceList(sessionID)
}
