package main

import (
	. "common"
)

var (
	max_open_file = 102400
)

func Init() bool {
	SetLogName("log.client")
	err := InitLog(0)
	if err != nil {
		LOG_STD("Init log failed, error: ", err)
		return false
	}

	if !SetMaxOpenFiles(uint64(max_open_file), uint64(max_open_file)) {
		LOG_STD("Set max open files failed.")
		return false
	}

	return true
}
