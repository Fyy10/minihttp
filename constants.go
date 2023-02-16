package minihttp

import "time"

const (
	// status code and text
	StatusOK     = 200
	StatusOKText = "OK"

	StatusMovedPermanently     = 301
	StatusMovedPermanentlyText = "Moved permanently"

	StatusBadRequest     = 400
	StatusBadRequestText = "Bad request"

	StatusNotFound     = 404
	StatusNotFoundText = "Not found"

	StatusInternalServerError     = 500
	StatusInternalServerErrorText = "Internal Server Error"

	// receive buffer size
	RecvBufSize = 1024

	// receive timeout
	RecvTimeout time.Duration = 5 * time.Second
)
