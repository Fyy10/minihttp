package minihttp

import (
	"fmt"
)

type Response struct {
	Proto      string
	StatusCode int
	StatusText string
	Headers    map[string]string
	Body       string
}

func (resp *Response) String() string {
	headers := resp.Proto + " " + fmt.Sprint(resp.StatusCode) + " " + resp.StatusText + "\r\n"
	for k, v := range resp.Headers {
		headers += k + ": " + v + "\r\n"
	}
	return headers + "\r\n" + resp.Body
}

// 301 Moved Permanently
func ResponseMovedPermanently(proto, location string, close bool) *Response {
	resp := new(Response)
	resp.Headers = make(map[string]string)

	resp.Proto = proto
	resp.StatusCode = StatusMovedPermanently
	resp.StatusText = StatusMovedPermanentlyText
	if close {
		resp.Headers["Connection"] = "close"
	}
	resp.Headers["Location"] = location
	resp.Body = ""
	resp.Headers["Content-Length"] = fmt.Sprint(len(resp.Body))
	return resp
}

// 400 Bad Request
func ResponseBadRequest(proto string) *Response {
	resp := new(Response)
	resp.Headers = make(map[string]string)

	resp.Proto = proto
	resp.StatusCode = StatusBadRequest
	resp.StatusText = StatusBadRequestText
	resp.Headers["Connection"] = "close"
	resp.Body = ""
	resp.Headers["Content-Length"] = fmt.Sprint(len(resp.Body))
	return resp
}

// 404 Not Found
func ResponseNotFound(proto string, close bool) *Response {
	resp := new(Response)
	resp.Headers = make(map[string]string)

	resp.Proto = proto
	resp.StatusCode = StatusNotFound
	resp.StatusText = StatusNotFoundText
	if close {
		resp.Headers["Connection"] = "close"
	}
	resp.Body = ""
	resp.Headers["Content-Length"] = fmt.Sprint(len(resp.Body))
	return resp
}

// 500 Internal Server Error
func ResponseInternalServerError(proto string) *Response {
	resp := new(Response)
	resp.Headers = make(map[string]string)

	resp.Proto = proto
	resp.StatusCode = StatusInternalServerError
	resp.StatusText = StatusInternalServerErrorText
	resp.Headers["Connection"] = "close"
	resp.Body = ""
	resp.Headers["Content-Length"] = fmt.Sprint(len(resp.Body))
	return resp
}
