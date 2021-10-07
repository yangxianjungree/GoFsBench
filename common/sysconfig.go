package common

import (
	"syscall"
)

func SetMaxOpenFiles(max, cur uint64) bool {
	var rLimit syscall.Rlimit

	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		LOG_STD("Error Getting Rlimit ", err)
		return false
	}

	DBG(rLimit)

	rLimit.Max = max
	rLimit.Cur = cur

	err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		LOG_STD("Error Setting Rlimit ", err)
		return false
	}

	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
	if err != nil {
		LOG_STD("Error Getting Rlimit ", err)
		return false
	}

	DBG("Set success, Rlimit Final", rLimit)
	return true
}
