package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Health(context *gin.Context) {
	context.String(http.StatusOK, "ok")
}
