package minihttp

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"
)

func startServer() {
	s := Server{
		Addr: "localhost:8081",
		VirtualHosts: map[string]string{
			"website1": "cmd/demo/example",
			"website2": "cmd/demo",
		},
	}
	go s.ListenAndServe()
}

func Test200OK404NotFound(t *testing.T) {
	startServer()
	t.Parallel()

	reqList := []string{
		"GET / HTTP/1.1\r\nhost: website1\r\n\r\n",
		"GET /notfound HTTP/1.1\r\nhost:website1\r\nconnection: close\r\n\r\n",
	}

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		t.Fatal("cannot connect to server:", err)
	}

	for _, req := range reqList {
		conn.Write([]byte(req))
	}

	// wait for the server to close the connection
	time.Sleep(time.Second)

	// connection already closed, but could still read from system receive buffer
	// FIXME: does bufio read from conn completely?
	respReader := bufio.NewReader(conn)

	resp, err := http.ReadResponse(respReader, nil)
	if err != nil {
		t.Fatal("failed parsing response:", err)
	}

	if resp.Proto != "HTTP/1.1" {
		t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
	}

	if resp.StatusCode != 200 {
		t.Errorf("Expected response code of 200 but got: %v\n", resp.StatusCode)
	}

	resp.Body.Close()

	resp, err = http.ReadResponse(respReader, nil)
	if err != nil {
		t.Fatal("failed parsing response:", err)
	}

	if resp.Proto != "HTTP/1.1" {
		t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
	}

	if resp.StatusCode != 404 {
		t.Errorf("Expected response code of 404 but got: %v\n", resp.StatusCode)
	}

	if resp.ContentLength != 0 {
		t.Error("response content length is not 0, get", resp.ContentLength)
	}

	resp.Body.Close()
}

func Test301MovedPermanently(t *testing.T) {
	t.Parallel()

	req := "GET /example HTTP/1.1\r\nHost: website2\r\nconnection: close\r\n\r\n"

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		t.Fatal("cannot connect to server:", err)
	}

	conn.Write([]byte(req))

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal("failed parsing response:", err)
	}

	if resp.Proto != "HTTP/1.1" {
		t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
	}

	if resp.StatusCode != 301 {
		t.Errorf("Expected response code of 301 but got: %v\n", resp.StatusCode)
	}

	url, err := resp.Location()
	if err != nil {
		t.Fatal("Expected response with header \"Location\" but not found")
	}

	if url.Path != "/example/" {
		t.Errorf("Expected response location with \"/example/\" but got: %v\n", url.Path)
	}

	resp.Body.Close()
}

func Test400BadRequest(t *testing.T) {
	t.Parallel()

	req := "GET / HTTP/1.1\r\n\r\n"

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		t.Fatal("cannot connect to server:", err)
	}

	conn.Write([]byte(req))

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal("failed parsing response:", err)
	}

	if resp.Proto != "HTTP/1.1" {
		t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected response code of 400 but got: %v\n", resp.StatusCode)
	}

	if !resp.Close {
		t.Error("Expected response with header \"Connection: close\" but not found")
	}

	resp.Body.Close()
}

func Test404NotFound(t *testing.T) {
	t.Parallel()

	reqList := []string{
		"GET /../ HTTP/1.1\r\nhost: website1\r\n\r\n",
		"GET / HTTP/1.1\r\nhost: website2\r\n\r\n",
		"GET /../main.go HTTP/1.1\r\nhost:website1\r\n\r\n",
		"GET / HTTP/1.1\r\nhost:not:a:good:host\r\n\r\n",
		"GET /example/notfound HTTP/1.1\r\nhost:website2\r\nconnection:close\r\n\r\n",
	}

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		t.Fatal("cannot connect to server:", err)
	}

	for _, req := range reqList {
		conn.Write([]byte(req))

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			t.Fatal("failed parsing response:", err)
		}

		if resp.Proto != "HTTP/1.1" {
			t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
		}

		if resp.StatusCode != 404 {
			t.Errorf("Expected response code of 404 but got: %v\n", resp.StatusCode)
		}

		resp.Body.Close()
	}
}

func TestTimeout(t *testing.T) {
	t.Parallel()

	req := "GET / HTTP/1.1"

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		t.Fatal("cannot connect to server:", err)
	}

	conn.Write([]byte(req))

	resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
	if err != nil {
		t.Fatal("failed parsing response:", err)
	}

	if resp.Proto != "HTTP/1.1" {
		t.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
	}

	if resp.StatusCode != 400 {
		t.Errorf("Expected response code of 400 but got: %v\n", resp.StatusCode)
	}

	if !resp.Close {
		t.Error("Expected response with header \"Connection: close\" but not found")
	}

	resp.Body.Close()
}

func BenchmarkGETSingleConnection(b *testing.B) {
	req0 := fmt.Sprint(
		"GET / HTTP/1.1\r\n",
		"Host: website1\r\n",
		"User-Agent: gotest\r\n",
		"\r\n",
	)
	req1 := fmt.Sprint(
		"GET / HTTP/1.1\r\n",
		"Host: website1\r\n",
		"User-Agent: gotest\r\n",
		"Connection: close\r\n",
		"\r\n",
	)

	conn, err := net.Dial("tcp", "localhost:8081")

	if err != nil {
		b.Fatal("cannot connect to server:", err)
	}

	req := req0
	for i := 0; i < b.N; i++ {
		if i == b.N-1 {
			req = req1
		}

		conn.Write([]byte(req))

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			b.Fatal("failed parsing response:", err)
		}

		if resp.Proto != "HTTP/1.1" {
			b.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
		}

		if resp.StatusCode != 200 {
			b.Errorf("Expected response code of 200 but got: %v\n", resp.StatusCode)
		}

		resp.Body.Close()
	}
}

func BenchmarkGETMultipleConnection(b *testing.B) {
	req := fmt.Sprint(
		"GET / HTTP/1.1\r\n",
		"Host: website1\r\n",
		"User-Agent: gotest\r\n",
		"Connection: close\r\n",
		"\r\n",
	)

	for i := 0; i < b.N; i++ {
		conn, err := net.Dial("tcp", "localhost:8081")

		if err != nil {
			b.Fatal("cannot connect to server:", err)
		}

		conn.Write([]byte(req))

		resp, err := http.ReadResponse(bufio.NewReader(conn), nil)
		if err != nil {
			b.Fatal("failed parsing response:", err)
		}

		if resp.Proto != "HTTP/1.1" {
			b.Errorf("Expected HTTP/1.1 but got a version: %v\n", resp.Proto)
		}

		if resp.StatusCode != 200 {
			b.Errorf("Expected response code of 200 but got: %v\n", resp.StatusCode)
		}

		resp.Body.Close()
	}
}
