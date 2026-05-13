package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Code    string      `json:"code,omitempty"`
}

type PaginatedResponse struct {
	Success  bool        `json:"success"`
	Data     interface{} `json:"data"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

type ListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Total   int64       `json:"total"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

type MFARequiredResponse struct {
	Success     bool   `json:"success"`
	MFARequired bool   `json:"mfaRequired"`
	Message     string `json:"message"`
	Code        string `json:"code"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
	})
}

func SuccessPaginated(c *gin.Context, data interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, PaginatedResponse{
		Success:  true,
		Data:     data,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func SuccessList(c *gin.Context, data interface{}, total int64) {
	c.JSON(http.StatusOK, ListResponse{
		Success: true,
		Data:    data,
		Total:   total,
	})
}

func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: message,
	})
}

func ErrorWithCode(c *gin.Context, statusCode int, message, code string) {
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: message,
		Code:    code,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, message)
}

func ServiceUnavailable(c *gin.Context, message string) {
	Error(c, http.StatusServiceUnavailable, message)
}

func MFARequired(c *gin.Context, message string) {
	c.JSON(http.StatusOK, MFARequiredResponse{
		Success:     false,
		MFARequired: true,
		Message:     message,
		Code:        "MFA_REQUIRED",
	})
}

// SafeError logs the full error internally and returns a safe message to the client.
// Use this for errors from services, databases, or other internal operations
// that may contain sensitive implementation details.
func SafeError(c *gin.Context, statusCode int, err error, context string) {
	log.Printf("[ERROR] %s: %v", context, err)
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: "an internal error occurred",
	})
}

// SafeErrorWithCode is like SafeError but allows specifying a custom error code.
func SafeErrorWithCode(c *gin.Context, statusCode int, err error, context string, code string) {
	log.Printf("[ERROR] %s: %v", context, err)
	c.JSON(statusCode, ErrorResponse{
		Success: false,
		Message: "an internal error occurred",
		Code:    code,
	})
}
