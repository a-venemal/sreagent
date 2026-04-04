package types

// Response is the unified API response structure.
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData wraps paginated data.
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// PageQuery contains pagination parameters.
type PageQuery struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}

// DefaultPageQuery returns a PageQuery with default values.
func DefaultPageQuery() PageQuery {
	return PageQuery{Page: 1, PageSize: 20}
}
