package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const RequestIDKey = "request_id"

type StatusBlock struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalItems int `json:"total_items"`
	TotalPages int `json:"total_pages"`
	PrevPage   int `json:"prev_page"`
	NextPage   int `json:"next_page"`
}

type Response struct {
	Status StatusBlock `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Meta   interface{} `json:"meta,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func write(c *gin.Context, code int, message string, data, meta interface{}) {
	c.JSON(code, Response{
		Status: StatusBlock{Code: code, Message: message},
		Data:   data,
		Meta:   meta,
	})
}

func writeError(c *gin.Context, code int, message, errDetail string) {
	c.JSON(code, Response{
		Status: StatusBlock{Code: code, Message: message},
		Error:  errDetail,
	})
}

func Success(c *gin.Context, message string, data interface{}) {
	write(c, http.StatusOK, message, data, nil)
}

// Pagination sekarang strongly typed, bukan interface{}
func SuccessWithPagination(c *gin.Context, message string, data interface{}, meta Pagination) {
	write(c, http.StatusOK, message, data, meta)
}

func Created(c *gin.Context, message string, data interface{}) {
	write(c, http.StatusCreated, message, data, nil)
}

func BadRequest(c *gin.Context, errDetail string) {
	writeError(c, http.StatusBadRequest, "Bad Request", errDetail)
}

func Unauthorized(c *gin.Context, errDetail string) {
	writeError(c, http.StatusUnauthorized, "Unauthorized", errDetail)
}

func Forbidden(c *gin.Context, errDetail string) {
	writeError(c, http.StatusForbidden, "Forbidden", errDetail)
}

func NotFound(c *gin.Context, errDetail string) {
	writeError(c, http.StatusNotFound, "Not Found", errDetail)
}

func InternalServerError(c *gin.Context, errDetail string) {
	writeError(c, http.StatusInternalServerError, "Internal Server Error", errDetail)
}

func ServiceUnavailable(c *gin.Context, errDetail string) {
	writeError(c, http.StatusServiceUnavailable, "Service Unavailable", errDetail)
}
