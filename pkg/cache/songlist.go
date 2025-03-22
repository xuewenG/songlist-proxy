package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/xuewenG/songlist-proxy/pkg/config"
	"github.com/xuewenG/songlist-proxy/pkg/decompress"
)

var cache []byte

var mutex sync.Mutex
var wg sync.WaitGroup

func GetSonglistFromCache(header http.Header) []byte {
	if len(cache) != 0 {
		log.Println("Cache is available")
		go refreshCache(header)
		return cache
	}

	log.Println("Cache is not available")
	refreshCache(header)
	wg.Wait()
	log.Println("Cache is updated")

	return cache
}

func refreshCache(header http.Header) {
	log.Println("Try to refresh cache")

	if mutex.TryLock() {
		log.Println("Start refreshing cache")
		wg.Add(1)
		cache = fetchData(header)
		wg.Done()
		mutex.Unlock()
		log.Println("Refreshing cache finished")
	} else {
		log.Println("Cache is refreshing")
	}
}

type requestRemoteServerData struct {
	Url string `json:"url"`
	Uid string `json:"uid"`
}

func fetchData(header http.Header) []byte {
	log.Println("Fetching data from remote server")

	// 使用固定参数
	reqBody, err := json.Marshal(requestRemoteServerData{
		Url: config.Config.Bili.Url,
		Uid: config.Config.Bili.Uid,
	})
	if err != nil {
		log.Fatalf("Failed to encode data, %v\n", err)
		return make([]byte, 0)
	}

	// 创建新的请求，将用户的请求转发到原始地址
	targetURL := "https://api.starlwr.com/songlist/getView"
	req, err := http.NewRequest(http.MethodPost, targetURL, bytes.NewReader(reqBody))
	if err != nil {
		log.Fatalf("Failed to create request, %v\n", err)
		return make([]byte, 0)
	}

	// 复制用户请求的 Header
	req.Header = header
	req.Header.Set("Accept-Encoding", "gzip")

	// 发起请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Failed to fetch data from target server, %v\n", err)
		return make([]byte, 0)
	}
	defer resp.Body.Close()

	// 读取响应数据
	var data []byte
	if strings.Contains(resp.Header.Get("Content-Encoding"), "gzip") {
		data, err = decompress.Gzip(resp.Body)
		if err != nil {
			log.Fatalf("Failed to decompress data from target server, %v\n", err)
			return make([]byte, 0)
		}
	} else {
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalf("Failed to read data from target server, %v\n", err)
			return make([]byte, 0)
		}
	}

	// 处理头像数据
	data = checkAvatar(data)

	log.Println("Fetching data from remote finished")
	return data
}

func checkAvatar(originData []byte) []byte {
	var res map[string]interface{}
	err := json.Unmarshal(originData, &res)
	if err != nil {
		log.Fatalf("Failed to decode data, %v\n", err)
		return originData
	}

	if dataField, ok := res["data"].(map[string]interface{}); ok {
		faceField, ok := dataField["face"].(string)
		if !ok || faceField == "" {
			log.Println("No avatar, use default")
			dataField["face"] = config.Config.Bili.Avatar
		}
	} else {
		log.Fatalln("Failed to decode dataField")
	}

	modifiedData, err := json.Marshal(res)
	if err != nil {
		log.Fatalf("Failed to encode data, %v\n", err)
		return originData
	}

	return modifiedData
}
