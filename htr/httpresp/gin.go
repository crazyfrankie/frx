package httpresp

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type data struct {
	Code    int32  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

func GinError(c *gin.Context, err error) {
	c.JSON(http.StatusOK, parseError(c.Request.Context(), err))
}

func GinSuccess(c *gin.Context, data any) {
	c.JSON(http.StatusOK, parseSuccess(data))
}
