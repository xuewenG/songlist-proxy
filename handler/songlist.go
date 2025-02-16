package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xuewenG/songlist-proxy/cache"
)

func GetView(c *gin.Context) {
	log.Println("Receive request")

	r := c.Request
	w := c.Writer

	data := cache.GetSonglistFromCache(r.URL.Path, r.Header)

	if len(data) != 0 {
		log.Println("Sending data")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	} else {
		log.Println("Data is empty")
		w.WriteHeader(http.StatusNotFound)
	}
}
