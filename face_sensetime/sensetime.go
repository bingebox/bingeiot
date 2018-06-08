package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// 新建目标库
	addTargetLibraryURL string = "http://%s:%s/verify/target/add"

	// 删除目标库
	deleteTargetLibraryURL = "http://%s:%s/verify/target/deletes"

	// 清空目标库
	clearTargetLibraryURL = "http://%s:%s/verify/target/clear"

	// 获取目标库信息
	getTargetLiraryInfoURL = "http://%s:%s/verify/target/gets"

	// 图片入库
	addImageSyncURL = "http://%s:%s/verify/face/synAdd"

	// 删除单张目标库图片
	deleteImageURL = "http://%s:%s/verify/face/deletes"

	// 获取单张目标库图片
	verifyFaceGet = "http://%s:%s/verify/face/gets"

	// 单张图片人脸检测
	imageDetectQualityURL = "http://%s:%s/verify/face/detectAndQuality"

	// 单张图片人脸特征提取
	getImageFeatureURL = "http://%s:%s/verify/feature/gets"

	// 批量图片人脸特征提取
	getImagesFeatureURLBatch = "http://%s:%s/verify/feature/batchGet"

	// 单张图片人脸属性提取
	getImageAttributeURL = "http://%s:%s/verify/attribute/gets"

	// 1:N 人脸搜索(图片)
	imageSearch = "http://%s:%s/verify/face/search"

	// 1:N 人脸搜索(特征)
	featureSearch = "http://%s:%s/verify/feature/search"

	// 将两张人脸图片进行对比，输出者的相似百分
	verifyFaceImageURL = "http://%s:%s/verify/face/verification"

	// http 正常返回状态
	httpNormalStatus = 200

	// 成功result
	successResult = "success"

	// 失败result
	errorResult = "error"
)

var (
	serverIP      = flag.String("ip", "192.168.2.14", "server ip")
	serverPort    = flag.String("port", "9001", "server tcp port")
	dbName        = flag.String("db", "", "face DB name")
	command       = flag.String("c", "", "command: version|detail|addLib|getLib|clearLib|deleteLib|getAttribute|compareImageLoop|compareFeatureLoop|addImageLoop|searchImageLoop|getFeature|getFeatureLoop|searchFeature|searchFeatureLoop")
	isPrint       = flag.Bool("isprint", false, "is printed for resp content")
	loop          = flag.Int("loop", 1, "routine count")
	count         = flag.Int("count", 1, "execute count")
	imageOnePath  = flag.String("f1", "", "the first image file path")
	imageTwoPath  = flag.String("f2", "", "the second image file path")
	imageFileName = flag.String("f", "", "imange file path")
	feature       = flag.String("feature", "", "face feature value")
	topNum        = flag.String("topNum", "100", "top number for search result")
	score         = flag.String("score", "0.0", "lower score for search result")
	fastSearch    = flag.String("fastSearch", "0", "fastSearch: 0|1")
)

// MessageResp : message fo response
type MessageResp struct {
	Result       string  `json:"result"`
	ErrorMessage string  `json:"errorMessage"`
	Score        float32 `json:"score"`
	TimeUsed     int32   `json:"time_used"`
}

type faceAttributeResp struct {
	Result       string        `json:"result"`
	ErrorMessage string        `json:"errorMessage"`
	TimeUsed     int32         `json:"time_used"`
	Data         faceAttribute `json:"data"`
}

type faceAttribute struct {
	Age        float32 `json:"age"`
	Attractive float32 `json:"attractive"`
	Smile      int     `json:"smile"`
	Gender     int     `json:"gender"`
	Eyeglass   int     `json:"eyeglass"`
	Sunglass   int     `json:"sunglass"`
	Mask       int     `json:"mask"`
	Race       int     `json:"race"`
	EyeOpen    int     `json:"eyeOpen"`
	MouthOpen  int     `json:"mouthOpen"`
	Beard      int     `json:"beard"`
}

// SenseTimeAPI : SenseTime API
type SenseTimeAPI struct {
	httpClient *http.Client
}

// Create a http client
func (senseTimeAPI *SenseTimeAPI) createHTTPClient() {
	senseTimeAPI.httpClient = &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   60 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     time.Duration(90) * time.Second,
		},
		Timeout: 20 * time.Second,
	}
}

// 生成URL编码字串
func makeURL(args map[string]string) (string, io.ReadCloser) {
	values := url.Values{}
	for key, val := range args {
		values.Add(key, val)
	}
	contentType := "application/x-www-form-urlencoded"
	return contentType, ioutil.NopCloser(strings.NewReader(values.Encode()))
}

// 生成multipart
func makeMultipart(args map[string]string) (string, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	for key, val := range args {
		w.WriteField(key, val)
	}
	return w.FormDataContentType(), buf
}

// 生成multipart, 带文件的
func makeMultipart2(args map[string]string, files map[string]string) (string, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	for key, val := range args {
		w.WriteField(key, val)
	}

	for key, filepath := range files {
		fw, err := w.CreateFormFile(key, filepath)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		fd, err := os.Open(filepath)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		defer fd.Close()
		_, err = io.Copy(fw, fd)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
	}
	return w.FormDataContentType(), buf
}

// HTTP post tool
func (senseTimeAPI *SenseTimeAPI) httpPost(url, contentType string, buf io.Reader) ([]byte, error) {
	resp, err := senseTimeAPI.httpClient.Post(url, contentType, buf)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return body, nil
}

// HTTP Get Tool
func (senseTimeAPI *SenseTimeAPI) httpGet(url string) string {
	resp, err := senseTimeAPI.httpClient.Get(url)
	if err != nil {
		log.Fatal(err.Error())
	} else if resp == nil {
		log.Fatal("resp == nil")
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("read body fatal: %s", err.Error())
		} else {
			bodyStr := string(body)
			log.Printf("status=%s, body=%s", resp.Status, bodyStr)
			return bodyStr
		}
	}
	return ""
}

// 查询版本号信息
func (senseTimeAPI *SenseTimeAPI) getVersion() {
	url := fmt.Sprintf("http://%s:%s/verify/version", *serverIP, *serverPort)
	senseTimeAPI.httpGet(url)
}

// 获取库容量信息
func (senseTimeAPI *SenseTimeAPI) getDetail() {
	url := fmt.Sprintf("http://%s:%s/verify/detail", *serverIP, *serverPort)
	senseTimeAPI.httpGet(url)
}

// 创建目标库
func (senseTimeAPI *SenseTimeAPI) addTargetLibrary(dbName, fastSearch *string) {
	args := make(map[string]string)
	args["dbName"] = *dbName
	args["fastSearch"] = *fastSearch
	contentType, buf := makeMultipart(args)
	url := fmt.Sprintf("http://%s:%s/verify/target/add", *serverIP, *serverPort)
	result, err := senseTimeAPI.httpPost(url, contentType, buf)
	if err == nil {
		log.Println(string(result))
	}
}

// 读取目标库信息
func (senseTimeAPI *SenseTimeAPI) getTargetLibraryInfo() {
	url := fmt.Sprintf(getTargetLiraryInfoURL, *serverIP, *serverPort)
	senseTimeAPI.httpGet(url)
}

// 清空目标库: 清空单个目标库中的所有图片
func (senseTimeAPI *SenseTimeAPI) clearTargetLibraryImages(dbName *string) {
	args := make(map[string]string)
	args["dbName"] = *dbName
	contentType, buf := makeURL(args)
	url := fmt.Sprintf(clearTargetLibraryURL, *serverIP, *serverPort)
	result, err := senseTimeAPI.httpPost(url, contentType, buf)
	if err == nil {
		log.Println(string(result))
	}
}

// 删除目标库
func (senseTimeAPI *SenseTimeAPI) deleteTargetLibrary(dbName *string) {
	args := make(map[string]string)
	args["dbName"] = *dbName
	contentType, buf := makeURL(args)
	url := fmt.Sprintf(deleteTargetLibraryURL, *serverIP, *serverPort)
	result, err := senseTimeAPI.httpPost(url, contentType, buf)
	if err == nil {
		log.Println(string(result))
	}
}

// 获取单张人脸图片属性
func (senseTimeAPI *SenseTimeAPI) getFaceAtrribute(imageFileName string) {
	files := make(map[string]string)
	files["imageData"] = imageFileName

	contentType, buf := makeMultipart2(nil, files)
	url := fmt.Sprintf("http://%s:%s/verify/attribute/gets", *serverIP, *serverPort)
	respBody, err := senseTimeAPI.httpPost(url, contentType, buf)
	if err == nil {
		log.Println(string(respBody))
		var resp faceAttributeResp
		err = json.Unmarshal(respBody, &resp)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			log.Printf("%s, %s, %d\n", resp.Result, resp.ErrorMessage, resp.TimeUsed)
			log.Printf("\n age-年龄: %f\n attractive-颜值: %f\n smile-是否微笑(1-是,0-否): %d\n gender-性别(0-男,1-女): %d\n beard-胡子(1-是,0-否): %d\n eyeglass-戴眼镜(1-是,0-否): %d\n eyeOpen-睁眼(1-是,0-否): %d\n mask-戴口罩(1-是,0-否): %d\n mouthOpen-张嘴(1-是,0-否): %d\n race-人种(0-黄,1-黑,2-白): %d\n sunglass-戴太阳镜(1-是,0-否): %d\n",
				resp.Data.Age, resp.Data.Attractive, resp.Data.Smile, resp.Data.Gender, resp.Data.Beard,
				resp.Data.Eyeglass, resp.Data.EyeOpen, resp.Data.Mask, resp.Data.MouthOpen, resp.Data.Race,
				resp.Data.Sunglass)
		}
	}
}

// 1:N人脸搜索(图片)
func (senseTimeAPI *SenseTimeAPI) searchImageLoop(dbName, imageFileName, topNum, score string) {
	log.Printf("dbName=%s, imageFileName=%s, topNum=%s, score=%s, count=%d\n",
		dbName, imageFileName, topNum, score, *count)
	start := time.Now()
	var i int
	for i = 0; i < *count; i++ {
		args := make(map[string]string)
		args["dbName"] = dbName
		args["topNum"] = topNum
		args["score"] = score

		files := make(map[string]string)
		files["imageData"] = imageFileName

		contentType, buf := makeMultipart2(args, files)
		url := fmt.Sprintf(imageSearch, *serverIP, *serverPort)
		respBody, err := senseTimeAPI.httpPost(url, contentType, buf)
		if err == nil {
			var m MessageResp
			err = json.Unmarshal(respBody, &m)
			if err != nil {
				log.Fatal(err.Error())
				break
			} else {
				log.Printf("%s, %f, %s, %d\n", m.Result, m.Score, m.ErrorMessage, m.TimeUsed)
			}
		}
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("excCount: %d, excTime: %dnans, %ds\n", i, elapsed, elapsed.Nanoseconds()/1000000000)
}

// searchFeatureLoop : 1:N人脸搜索(特征)--压测总控
func (senseTimeAPI *SenseTimeAPI) searchFeatureLoop(dbName, feature, topNum, score *string) {
	fd, err := os.Open(*feature)
	if err == nil {
		defer fd.Close()
		content, err := ioutil.ReadAll(fd)
		if err != nil {
			log.Println(err.Error())
			return
		}
		*feature = string(content)
	}

	start := time.Now()
	chs := make([]chan int, *loop)
	for i := 0; i < *loop; i++ {
		chs[i] = make(chan int)
		go senseTimeAPI.searchFeatureRoutine(chs[i], dbName, feature, topNum, score)
		chs[i] <- *count
	}

	for _, ch := range chs {
		<-ch
	}

	end := time.Now()
	elapsed := end.Sub(start)
	elapsedSec := (elapsed.Nanoseconds() / 1000000000)
	qps := 0
	if int(elapsedSec) == 0 {
		qps = (*loop) * (*count)
	} else {
		qps = (*loop) * (*count) / int(elapsedSec)
	}
	log.Printf("thread: %d, exec count: %d, total elapsed time: %dnans, %ds, qps: %d\n", *loop, *count, elapsed, elapsedSec, qps)
}

// 1:N人脸搜索(特征)--压测协程
func (senseTimeAPI *SenseTimeAPI) searchFeatureRoutine(ch chan int, dbName, feature, topNum, score *string) {
	count := <-ch
	for i := 0; i < count; i++ {
		senseTimeAPI.searchFeature(dbName, feature, topNum, score, *isPrint)
	}
	ch <- 0
}

// 1:N人脸搜索(特征)
func (senseTimeAPI *SenseTimeAPI) searchFeature(dbName, feature, topNum, score *string, isPrint bool) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	w.WriteField("dbName", *dbName)
	w.WriteField("feature", *feature)
	w.WriteField("topNum", *topNum)
	w.WriteField("score", *score)

	url := fmt.Sprintf(featureSearch, *serverIP, *serverPort)
	resp, err := senseTimeAPI.httpClient.Post(url, w.FormDataContentType(), buf)
	if err != nil {
		log.Println(err.Error())
	} else if resp.Body == nil {
		log.Println("boyd is nil")
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
		} else {
			if isPrint {
				log.Println(string(body))
			}
		}
	}
}

func (senseTimeAPI *SenseTimeAPI) addImageSync(ch chan int, dbName, imageFileName string) {
	log.Printf("dbName=%s, imageFileName=%s\n", dbName, imageFileName)
	count := <-ch
	start := time.Now()
	var i int
	for i = 0; i < count; i++ {
		buf := new(bytes.Buffer)
		w := multipart.NewWriter(buf)
		defer w.Close()

		// dbName
		w.WriteField("dbName", dbName)
		w.WriteField("getFeature", "0")

		// imageDatas
		fw, err := w.CreateFormFile("imageDatas", imageFileName)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		fd, err := os.Open(imageFileName)
		if err != nil {
			log.Fatal(err.Error())
			break
		}
		defer fd.Close()
		_, err = io.Copy(fw, fd)
		if err != nil {
			log.Fatal(err.Error())
			break
		}

		url := fmt.Sprintf(addImageSyncURL, *serverIP, *serverPort)
		resp, err := senseTimeAPI.httpClient.Post(url, w.FormDataContentType(), buf)
		if err != nil {
			log.Fatal(err.Error())
			break
		} else if resp.Body == nil {
			log.Fatal("boyd is nil")
			break
		} else {
			defer resp.Body.Close()
			jsonbuf, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err.Error())
				break
			} else {
				log.Println(string(jsonbuf))
			}
		}
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("excCount: %d, excTime: %dnans, %ds\n", i, elapsed, elapsed.Nanoseconds()/1000000000)
	ch <- i
}

func (senseTimeAPI *SenseTimeAPI) addImageSyncLoop(dbName, imageFileName string) error {
	start := time.Now()
	chs := make([]chan int, *loop)
	for i := 0; i < *loop; i++ {
		chs[i] = make(chan int)
		go senseTimeAPI.addImageSync(chs[i], dbName, imageFileName)
		chs[i] <- *count
	}

	for _, ch := range chs {
		<-ch
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("thread: %d, exec count: %d, total elapsed time: %dnans, %ds\n", *loop, *count, elapsed, elapsed.Nanoseconds()/1000000000)
	return nil
}

func getFileMultipart(imageOne, imageTwo string) (string, *bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	// imageOne
	fw, err := w.CreateFormFile("imageOne", imageOne)
	if err != nil {
		return "", nil, err
	}
	fd, err := os.Open(imageOne)
	if err != nil {
		return "", nil, err
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		return "", nil, err
	}

	// imageTwo
	fw2, err := w.CreateFormFile("imageTwo", imageTwo)
	if err != nil {
		return "", nil, err
	}
	fd2, err := os.Open(imageTwo)
	if err != nil {
		return "", nil, err
	}
	defer fd2.Close()
	_, err = io.Copy(fw2, fd2)
	if err != nil {
		return "", nil, err
	}

	return w.FormDataContentType(), buf, nil
}

// Done : 执行参数体
type Done struct {
	executeTotal int
	executeCount int
	usedTime     int64
}

func (senseTimeAPI *SenseTimeAPI) verifyFaceImage(imageOne, imageTwo string) {
	url := fmt.Sprintf(verifyFaceImageURL, *serverIP, *serverPort)

	start := time.Now()
	chs := make([]chan Done, *loop)
	for i := 0; i < *loop; i++ {
		chs[i] = make(chan Done)
		go senseTimeAPI.verifyFaceImageExecute(chs[i], url, imageOne, imageTwo)
		chs[i] <- Done{*count, 0, 0}
	}

	var excTime int64
	for _, ch := range chs {
		done := <-ch
		excTime = excTime + done.usedTime
	}

	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("thread: %d, exec count: %d, total elapsed time: %dnans, %ds\n", *loop, *count, elapsed, elapsed.Nanoseconds()/1000000000)
}

func (senseTimeAPI *SenseTimeAPI) verifyFaceImageExecute(ch chan Done, url, imageOne, imageTwo string) {
	done := <-ch
	start := time.Now()
	var i int
	for i = 0; i < done.executeTotal; i++ {
		contentType, buf, err := getFileMultipart(imageOne, imageTwo)
		if err != nil {
			log.Fatal(err.Error())
			return
		}

		resp, err := senseTimeAPI.httpClient.Post(url, contentType, buf)
		if err != nil {
			log.Fatal(err.Error())
		} else {
			defer resp.Body.Close()
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal(err.Error())
			} else {
				bodyStr := string(body)

				b := []byte(bodyStr)
				var m MessageResp
				err = json.Unmarshal(b, &m)
				if err != nil {
					log.Fatal(err.Error())
				} else {
					log.Printf("%s, %f, %s, %d\n", m.Result, m.Score, m.ErrorMessage, m.TimeUsed)
				}
			}
		}
	}
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("excCount: %d, excTime: %dnans, %ds\n", i, elapsed, elapsed.Nanoseconds()/1000000000)
	ch <- Done{done.executeTotal, done.executeCount, elapsed.Nanoseconds()}
}

// 单张图片人脸特征提取--压测总控
func (senseTimeAPI *SenseTimeAPI) getImageFeatureLoop(imagePath string) {
	log.Printf("loop=%d, count=%d\n", *loop, *count)
	start := time.Now()
	chs := make([]chan int, *loop)
	for i := 0; i < *loop; i++ {
		chs[i] = make(chan int)
		go senseTimeAPI.getImageFeatureRoutine(chs[i], imagePath)
		chs[i] <- *count
	}

	for _, ch := range chs {
		<-ch
	}

	end := time.Now()
	elapsed := end.Sub(start)
	elapsedSec := (elapsed.Nanoseconds() / 1000000000)
	qps := 0
	if int(elapsedSec) == 0 {
		qps = (*loop) * (*count)
	} else {
		qps = (*loop) * (*count) / int(elapsedSec)
	}
	log.Printf("thread: %d, exec count: %d, total elapsed time: %dnans, %ds, qps: %d\n", *loop, *count, elapsed, elapsedSec, qps)
}

// 单张图片人脸特征提取--压测协程
func (senseTimeAPI *SenseTimeAPI) getImageFeatureRoutine(ch chan int, imagePath string) {
	count := <-ch
	for i := 0; i < count; i++ {
		senseTimeAPI.getImageFeature(imagePath, *isPrint)
	}
	ch <- 0
}

// getImageFeature : 单张图片人脸特征提取
func (senseTimeAPI *SenseTimeAPI) getImageFeature(imagePath string, isPrint bool) {
	log.Printf("imagePath=%s, isPrint=%v\n", imagePath, isPrint)
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	// imageData
	fw, err := w.CreateFormFile("imageData", imagePath)
	if err != nil {
		log.Println(err.Error())
		return
	}
	fd, err := os.Open(imagePath)
	if err != nil {
		log.Println(err.Error())
		return
	}
	defer fd.Close()
	_, err = io.Copy(fw, fd)
	if err != nil {
		log.Println(err.Error())
		return
	}

	url := fmt.Sprintf(getImageFeatureURL, *serverIP, *serverPort)
	resp, err := senseTimeAPI.httpClient.Post(url, w.FormDataContentType(), buf)
	if err != nil {
		log.Println(err.Error())
		return
	} else if resp.Body == nil {
		log.Println("boyd is nil")
		return
	} else {
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err.Error())
		} else {
			if isPrint {
				log.Println(string(body))
			}
		}
	}
}

// DisplayName : 模块名
func (senseTimeAPI *SenseTimeAPI) DisplayName() string {
	return "商汤"
}

func readFile(fileName string) (string, error) {
	fd, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	defer fd.Close()

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		log.Fatal(err.Error())
		return "", err
	}
	return string(content), nil
}

// CompareFeatureLoop : 单张图片人脸特征提取--压测总控
func (senseTimeAPI *SenseTimeAPI) CompareFeatureLoop(featureFileName1, featureFileName2 string) {
	log.Printf("-f1 %s -f2 %s -isprint %v -loop %d -count %d\n", featureFileName1, featureFileName2,
		*isPrint, *loop, *count)
	feature1, _ := readFile(featureFileName1)
	feature2, _ := readFile(featureFileName2)

	start := time.Now()
	chs := make([]chan int, *loop)
	for i := 0; i < *loop; i++ {
		chs[i] = make(chan int)
		go senseTimeAPI.CompareFeatureRoutine(chs[i], feature1, feature2, *isPrint)
		chs[i] <- *count
	}

	for _, ch := range chs {
		<-ch
	}

	end := time.Now()
	elapsed := end.Sub(start)
	elapsedSec := (elapsed.Nanoseconds() / 1000000000)
	qps := 0
	if int(elapsedSec) == 0 {
		qps = (*loop) * (*count)
	} else {
		qps = (*loop) * (*count) / int(elapsedSec)
	}
	log.Printf("thread: %d, exec count: %d, total elapsed time: %dnans, %ds, qps: %d\n",
		*loop, *count, elapsed, elapsedSec, qps)
}

// CompareFeatureRoutine : 单张图片人脸特征提取--压测协程
func (senseTimeAPI *SenseTimeAPI) CompareFeatureRoutine(ch chan int, feature1, feature2 string, isPrint bool) {
	count := <-ch
	for i := 0; i < count; i++ {
		senseTimeAPI.CompareFeature(feature1, feature2, isPrint)
	}
	ch <- 0
}

// CompareFeature : 1:1人脸特征比对
func (senseTimeAPI *SenseTimeAPI) CompareFeature(feature1, feature2 string, isPrint bool) (float32, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	defer w.Close()

	w.WriteField("feature1", feature1)
	w.WriteField("feature2", feature2)

	url := fmt.Sprintf("http://%s:%s/verify/feature/compare", *serverIP, *serverPort)
	resp, err := senseTimeAPI.httpClient.Post(url, w.FormDataContentType(), buf)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return 0, err
	}
	bodyStr := string(body)
	if isPrint {
		log.Printf("status=%s, body=%s", resp.Status, bodyStr)
	}
	return 0, nil
}

func main() {
	flag.Parse()
	log.Printf("serverIP=%s, serverPort=%s\n", *serverIP, *serverPort)

	senseTimeAPI := SenseTimeAPI{}
	senseTimeAPI.createHTTPClient()
	if *command == "getLib" {
		senseTimeAPI.getTargetLibraryInfo()
	} else if *command == "addLib" {
		senseTimeAPI.addTargetLibrary(dbName, fastSearch)
	} else if *command == "clearLib" {
		senseTimeAPI.clearTargetLibraryImages(dbName)
	} else if *command == "deleteLib" {
		senseTimeAPI.deleteTargetLibrary(dbName)
	} else if *command == "version" {
		senseTimeAPI.getVersion()
	} else if *command == "detail" {
		senseTimeAPI.getDetail()
	} else if *command == "compareImageLoop" {
		senseTimeAPI.verifyFaceImage(*imageOnePath, *imageTwoPath)
	} else if *command == "compareFeatureLoop" {
		senseTimeAPI.CompareFeatureLoop(*imageOnePath, *imageTwoPath)
	} else if *command == "addImageLoop" {
		senseTimeAPI.addImageSyncLoop(*dbName, *imageFileName)
	} else if *command == "searchImageLoop" {
		senseTimeAPI.searchImageLoop(*dbName, *imageFileName, *topNum, *score)
	} else if *command == "getFeature" && *imageFileName != "" {
		senseTimeAPI.getImageFeature(*imageFileName, true)
	} else if *command == "getFeatureLoop" && *imageFileName != "" {
		senseTimeAPI.getImageFeatureLoop(*imageFileName)
	} else if *command == "searchFeature" {
		senseTimeAPI.searchFeature(dbName, feature, topNum, score, true)
	} else if *command == "searchFeatureLoop" {
		senseTimeAPI.searchFeatureLoop(dbName, feature, topNum, score)
	} else if *command == "getAttribute" {
		senseTimeAPI.getFaceAtrribute(*imageFileName)
	} else {
		const usage = `Usage: ./face_sensetime [cmd...]
>face_sensetime -c version -ip=ip_addr -port=tcp_port
>face_sensetime -c detail -ip=ip_addr -port=tcp_port
>face_sensetime -c addLib -ip=ip_addr -port=tcp_port -db=dbname -fastSearch=[0|1]
>face_sensetime -c getLib -ip=ip_addr -port=tcp_port
>face_sensetime -c clearLib -ip=ip_addr -port=tcp_port -db=dbname 
>face_sensetime -c deleteLib -ip=ip_addr -port=tcp_port -db=dbname 
>face_sensetime -c compareImageLoop -ip=ip_addr -port=tcp_port -f1=imageFileName1 -f2=imageFileName2 -loop=M -count=N
>face_sensetime -c compareFeatureLoop -ip=ip_addr -port=tcp_port -f1=featureFileName1 -f2=featureFileName2 -isprint=[true|false] -loop=M -count=N
>face_sensetime -c addImageLoop -ip=ip_addr -port=tcp_port -db=dbname -f=imageFileName -loop=M -count=N
>face_sensetime -c searchImageLoop -ip=ip_addr -port=tcp_port -db=dbname -f=imageFileName -topNum=X -score=Y -count=N
>face_sensetime -c getFeature -ip=ip_addr -port=tcp_port -f=imageFileName 
>face_sensetime -c getFeatureLoop -ip=ip_addr -port=tcp_port -f=imageFileName -isprint=[true|false] -loop=M -count=N
>face_sensetime -c searchFeature -ip=ip_addr -port=tcp_port -db=dbname -feature=featureContent -topNum=X -score=Y
>face_sensetime -c searchFeatureLoop -ip=ip_addr -port=tcp_port -db=dbname -feature=featureFileName -topNum=X -score=Y -isprint=[true|false] -loop=M -count=N
>face_sensetime -c getAttribute -ip=ip_addr -port=tcp_port -f=imageFileName 
		`
		log.Println(usage)
	}
}
