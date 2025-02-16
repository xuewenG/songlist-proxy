package cache

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/xuewenG/songlist-proxy/config"
	"github.com/xuewenG/songlist-proxy/decompress"
)

type requestRemoteServerData struct {
	Url string `json:"url"`
	Uid string `json:"uid"`
}

type cacheItem struct {
	Data     []byte
	CreateAt int64
}

var cacheMutex sync.Mutex
var cacheItems = make(map[string]cacheItem)
var updateMutex sync.Mutex
var updateChans = make(map[string]chan struct{})

func GetSonglistFromCache(path string, header http.Header) []byte {
	// 检查缓存是否存在
	cacheMutex.Lock()
	item, found := cacheItems[path]
	cacheMutex.Unlock()

	if found {
		log.Println("Cache found")
		go updateCacheData(path, header)

		return item.Data
	} else {
		log.Println("Cache not forund")
		data := updateCacheData(path, header)

		return data
	}
}

func updateCacheData(path string, header http.Header) []byte {
	// 检查是否有正在更新的请求
	updateMutex.Lock()
	updateChan, updating := updateChans[path]
	if !updating {
		log.Println("Start updating cache")

		// 没有更新操作，创建更新通道并开始更新
		updateChan = make(chan struct{})
		updateChans[path] = updateChan
		updateMutex.Unlock()

		// 获取数据
		data := fetchData(header)

		// 更新缓存
		cacheMutex.Lock()
		cacheItems[path] = cacheItem{
			Data:     data,
			CreateAt: time.Now().UnixMilli(),
		}
		cacheMutex.Unlock()

		// 更新完成后关闭标志
		close(updateChan)
		updateMutex.Lock()
		delete(updateChans, path)
		updateMutex.Unlock()

		return data
	} else {
		log.Println("Wait for updating cache")

		// 如果缓存正在更新，等待更新完成
		updateMutex.Unlock()
		<-updateChan

		// 返回缓存数据
		cacheMutex.Lock()
		item := cacheItems[path]
		cacheMutex.Unlock()

		return item.Data
	}
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
		var finalAvatarUrl string
		avatarCacheKey := "VIRTUAL_KEY_AVATAR"

		faceField, ok := dataField["face"].(string)
		if !ok || faceField == "" {
			log.Println("No avatar, check cache")

			cacheMutex.Lock()
			item, found := cacheItems[avatarCacheKey]
			cacheMutex.Unlock()

			if found {
				log.Println("Cache found")
				finalAvatarUrl = string(item.Data)
			} else {
				log.Println("Cache not found, use default url")
				finalAvatarUrl = config.Config.Bili.Avatar
			}

			dataField["face"] = finalAvatarUrl
		} else {
			finalAvatarUrl = faceField
		}

		cacheMutex.Lock()
		cacheItems[avatarCacheKey] = cacheItem{
			Data:     []byte(finalAvatarUrl),
			CreateAt: time.Now().UnixMilli(),
		}
		cacheMutex.Unlock()
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
