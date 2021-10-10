package main

import (
	. "common"
	"net/http"
	_ "net/http/pprof"
)

var (
	server_addr = ":" + GetGlobalConfigIns().Port
)

func main() {
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
