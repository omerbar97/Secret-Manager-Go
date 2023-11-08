package types

import "net/http"

type ApiError struct {
	Err    string
	Status int
}

type ApiHandlerFunc func(rw http.ResponseWriter, r *http.Request) error

func (e *ApiError) Error() string {
	return e.Err
}
