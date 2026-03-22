package ginx

// ApiResponse is the unified API response format
type ApiResponse struct {
	Code int    `json:"code" example:"0"`
	Msg  string `json:"msg" example:"成功"`
	Data any    `json:"data"`
}

// PageResponse is a generic paginated response
type PageResponse[T any] struct {
	Total    int64 `json:"total" example:"100"`
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"page_size" example:"20"`
	List     []T   `json:"list"`
}
