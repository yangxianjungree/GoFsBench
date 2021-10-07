package main

import (
	. "common"
	"net/http"
	_ "net/http/pprof"
)

var (
	server_addr = ":9999"
	// server = "127.0.0.1:9999"
	// server = "192.168.1.4:9999"
)

func main() {
	// debug.SetMaxThreads(20)

	if !Init() {
		ERR("Init failed.")
		return
	}

	go func() {
		LOG_STD("raise pprof http server.....")
		http.ListenAndServe("0.0.0.0:6060", nil)
	}()

	file_server(server_addr)
}
