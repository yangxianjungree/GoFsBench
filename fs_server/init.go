package main

import (
	. "common"
)

func Init() bool {
	SetLogName("log.server")
	err := InitLog(0)
	if err != nil {
		LOG_STD("Init log failed, error: ", err)
		return false
	}

	if !InitMaxOpenFiles() {
		LOG_STD("Set max open files failed.")
		return false
	}

	InitCPool()

	return true
}
