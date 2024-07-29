package status

import (
	"memento/utils"
	"net/http"
)

type StatusResponse struct {
	HTTPStatusCode int    `json:"-"`
	StatusText     string `json:"status"`
	ErrorText      string `json:"error,omitempty"`
}

func (s *StatusResponse) Render(w http.ResponseWriter) {
	utils.JsonEncode(w, s.HTTPStatusCode, s)
}

func ErrNotFound(err error) *StatusResponse {
	return &StatusResponse{
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     "Not Found",
		ErrorText:      err.Error(),
	}
}

func ErrBadRequest(err error) *StatusResponse {
	return &StatusResponse{
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     "Bad Request",
		ErrorText:      err.Error(),
	}
}

func ErrInternal(err error) *StatusResponse {
	return &StatusResponse{
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     "Internal Server Error",
		ErrorText:      err.Error(),
	}
}

func StatusOK(status string) *StatusResponse {
	return &StatusResponse{
		HTTPStatusCode: http.StatusOK,
		StatusText:     status,
	}
}

func StatusCreated(status string) *StatusResponse {
	return &StatusResponse{
		HTTPStatusCode: http.StatusCreated,
		StatusText:     status,
	}
}
