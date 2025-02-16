package router

import (
	"github.com/gin-gonic/gin"
	"github.com/xuewenG/songlist-proxy/handler"
)

func Bind(r *gin.Engine) {
	r.POST("/songlist/getView", handler.GetView)
}
