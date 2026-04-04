package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	apperr "github.com/sreagent/sreagent/internal/pkg/errors"
	"github.com/sreagent/sreagent/pkg/types"
)

// Success returns a successful JSON response.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, types.Response{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// SuccessPage returns a successful paginated JSON response.
func SuccessPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, types.Response{
		Code:    0,
		Message: "ok",
		Data: types.PageData{
			List:     list,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Error returns an error JSON response.
func Error(c *gin.Context, err error) {
	if appErr, ok := err.(*apperr.AppError); ok {
		c.JSON(appErr.HTTPStatus(), types.Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, types.Response{
		Code:    50000,
		Message: "internal server error",
	})
}

// ErrorWithMessage returns an error response with a custom message.
func ErrorWithMessage(c *gin.Context, code int, message string) {
	status := http.StatusBadRequest
	if code >= 50000 {
		status = http.StatusInternalServerError
	}
	c.JSON(status, types.Response{
		Code:    code,
		Message: message,
	})
}

// GetPageQuery extracts pagination parameters from the request.
func GetPageQuery(c *gin.Context) types.PageQuery {
	pq := types.DefaultPageQuery()
	if page, err := strconv.Atoi(c.DefaultQuery("page", "1")); err == nil && page > 0 {
		pq.Page = page
	}
	if pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "20")); err == nil && pageSize > 0 && pageSize <= 100 {
		pq.PageSize = pageSize
	}
	return pq
}

// GetIDParam extracts an ID parameter from the URL path.
func GetIDParam(c *gin.Context, paramName string) (uint, error) {
	idStr := c.Param(paramName)
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, apperr.ErrInvalidParam
	}
	return uint(id), nil
}

// GetCurrentUserID gets the authenticated user's ID from context.
func GetCurrentUserID(c *gin.Context) uint {
	if id, exists := c.Get("user_id"); exists {
		if uid, ok := id.(uint); ok {
			return uid
		}
	}
	return 0
}
