package main

import (
	"log"

	"github.com/Fyy10/minihttp"
)

func main() {
	hosts := make(map[string]string)
	hosts["localhost:8080"] = "cmd/demo/"
	s := minihttp.Server{
		Addr:         "localhost:8080",
		VirtualHosts: hosts,
	}

	log.Fatal(s.ListenAndServe())
}
