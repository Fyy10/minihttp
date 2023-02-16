package minihttp

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"os"
	"path"
	"strings"
	"time"
)

type Server struct {
	// Addr specifies the TCP address for the server to listen on,
	// in the form "host:port". It shall be passed to net.Listen()
	// during ListenAndServe().
	Addr string // e.g. ":0"

	// VirtualHosts contains a mapping from host name to the docRoot path
	// (i.e. the path to the directory to serve static files from) for
	// all virtual hosts that this server supports
	VirtualHosts map[string]string
}

// ListenAndServe listens on the TCP network address s.Addr and then
// handles requests on incoming connections.
func (s *Server) ListenAndServe() error {
	// validate docRoots
	for _, v := range s.VirtualHosts {
		fileInfo, err := os.Stat(v)
		// path not exist or is not a directory
		if err != nil || !fileInfo.IsDir() {
			return err
		}
	}

	listener, err := net.Listen("tcp", s.Addr)
	if err != nil {
		log.Println("cannot create listener:", err)
		return err
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("failed accepting connection:", err)
			continue
		}
		go s.handleConnection(conn)
	}
}

/*
Continuously read requests from connection, there can be multiple requests.
Close connection if timeout or:
 1. request has the "Connection: closed" header
 2. bad request
*/
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	remaining := ""
	buf := make([]byte, RecvBufSize)

	req := new(Request)
	req.Headers = make(map[string]string)

	var resp *Response // nil

	// initialize timeout
	conn.SetReadDeadline(time.Now().Add(RecvTimeout))
	for {
		for strings.Contains(remaining, "\r\n\r\n") {
			idx := strings.Index(remaining, "\r\n\r\n")
			reqString := remaining[:idx+2]
			remaining = remaining[idx+4:]

			err := req.Parse(reqString)
			if err != nil {
				resp = ResponseBadRequest("HTTP/1.1")
			} else {
				switch req.Method {
				case "GET":
					resp = s.handleGET(req)
				default:
					resp = ResponseInternalServerError(req.Proto)
				}
			}

			// send response
			conn.Write([]byte(resp.String()))

			// close connection if requested
			if resp.Headers["Connection"] == "close" {
				// FIXME: if connection is closed right after write, the other side may not receive the response (? not really)
				conn.Close()
				break
			}

			// refresh timeout for new requests
			conn.SetReadDeadline(time.Now().Add(RecvTimeout))
		}

		n, err := conn.Read(buf)
		if err != nil {
			// EOF (receive stream finished or connection closed by client)
			if err == io.EOF {
				conn.Close()
				break
			}

			// request timeout
			if os.IsTimeout(err) {
				if remaining != "" {
					// incomplete request
					log.Println("incomplete request until timeout, error:", err)
					log.Println("remaining:", remaining)
					conn.Write([]byte(ResponseBadRequest("HTTP/1.1").String()))
				}
				// FIXME: if connection is closed right after write, the other side may not receive the response (? not really)
				conn.Close()
				break
			}

			// connection already closed (from server side)
			if errors.Is(err, net.ErrClosed) {
				break
			}

			// other
			log.Println("failed receiving request:", err)
			conn.Close()
			break
		}

		remaining += string(buf[:n])
	}
}

func (s *Server) handleGET(req *Request) *Response {
	// handle a GET request
	resp := new(Response)
	resp.Headers = make(map[string]string)
	resp.Proto = req.Proto
	resp.StatusCode = StatusOK
	resp.StatusText = StatusOKText
	resp.Headers["Date"] = time.Now().UTC().Format(time.RFC1123)
	if s.VirtualHosts[req.Host] == "" {
		// the requested host is not hosted in virtual hosting
		log.Println("the requested host " + req.Host + " does not exist")
		return ResponseNotFound(req.Proto, req.Close)
	}

	filePath := path.Clean(path.Join(s.VirtualHosts[req.Host], req.URL))

	// check invalid file path
	// filePath should not escape docRoot
	if !strings.Contains(filePath, s.VirtualHosts[req.Host]) {
		log.Println("trying to escape docRoot, rejected path:", filePath)
		return ResponseNotFound(req.Proto, req.Close)
	}

	fileInfo, err := os.Stat(filePath)

	if err != nil {
		log.Println("file not exist, path:", filePath)
		return ResponseNotFound(req.Proto, req.Close)
	}

	// filePath is a directory
	if fileInfo.IsDir() {
		return ResponseMovedPermanently(req.Proto, req.URL+"/", req.Close)
	}

	resp.Headers["Content-Length"] = fmt.Sprint(fileInfo.Size())
	fileContent, _ := os.ReadFile(filePath)
	resp.Body = string(fileContent)
	resp.Headers["Last-Modified"] = fileInfo.ModTime().UTC().Format(time.RFC1123)
	resp.Headers["Content-Type"] = mime.TypeByExtension(path.Ext(filePath))

	if req.Close {
		resp.Headers["Connection"] = "close"
	}

	return resp
}
