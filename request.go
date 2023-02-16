package minihttp

import (
	"fmt"
	"log"
	"net/textproto"
	"strings"
)

type Request struct {
	Method string
	URL    string
	Proto  string

	Host    string
	Close   bool
	Headers map[string]string
	Body    string
}

type InvalidRequestError struct {
	Method      string
	URL         string
	Description string
}

func (e *InvalidRequestError) Error() string {
	return fmt.Sprintf("invalid request: %s %s %s", e.Method, e.URL, e.Description)
}

func (req *Request) Parse(reqString string) error {
	remaining := reqString
	lineCount := 0
	for strings.Contains(remaining, "\r\n") {
		lineCount++
		idx := strings.Index(remaining, "\r\n")
		line := remaining[:idx]
		remaining = remaining[idx+2:]

		// parse a request line
		if lineCount == 1 {
			// start line
			// FIXME: cannot handle multiple spaces
			elements := strings.Split(line, " ")
			if len(elements) != 3 {
				return &InvalidRequestError{"start", "line:", line}
			}
			switch elements[0] {
			case "GET":
				req.Method = "GET"

				// check valid URL
				req.URL = elements[1]
				if !strings.HasPrefix(req.URL, "/") {
					// URL does not contain '/' as prefix
					return &InvalidRequestError{req.Method, req.URL, "invalid URL"}
				}

				if strings.HasSuffix(req.URL, "/") {
					req.URL += "index.html"
				}

				req.Proto = elements[2]
				if req.Proto != "HTTP/1.1" {
					return &InvalidRequestError{req.Method, req.URL, req.Proto}
				}
			default:
				log.Println("Request contains unsupported method", elements[0])
				return &InvalidRequestError{req.Method, req.URL, "unsupported method"}
			}
		} else {
			// headers
			key, value, found := strings.Cut(line, ":")
			// check valid header
			if !found || key == "" {
				// ":" not found in the header line
				// or header name is empty (header starts with ":")
				return &InvalidRequestError{req.Method, req.URL, "invalid header: " + fmt.Sprint(line)}
			}

			// canonicalize the key
			key = textproto.CanonicalMIMEHeaderKey(key)
			// deal with prefix and suffix spaces in the value
			value = strings.Trim(value, " ")

			switch key {
			case "Host":
				req.Host = value
			case "Connection":
				if value == "close" {
					req.Close = true
				} else {
					req.Close = false
				}
			default:
				req.Headers[key] = value
			}
		}
	}

	if req.Host == "" {
		return &InvalidRequestError{req.Method, req.URL, "unspecified host"}
	}
	if len(remaining) != 0 {
		// this will never happen
		return &InvalidRequestError{req.Method, req.URL, "incomplete request, remaining: " + remaining}
	}

	return nil
}
