package apierror

import "fmt"

type APIError struct {
	Code string `json:"code"`
	Desc string `json:"desc"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Desc)
}

func New(code string) APIError {
	return APIError{Code: code}
}

func WithDesc(code, desc string) APIError {
	return APIError{Code: code, Desc: desc}
}
