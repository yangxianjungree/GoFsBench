package common

import (
	"os"
	"sync"
	"testing"
)

var LOGGING_ONCE sync.Once

func init_log() {
	SetLogName("log.log_benchmark")
	err := InitLog(0)
	if err != nil {
		LOG_STD("Init log failed, error: ", err)
		os.Exit(-1)
	}
}

func BenchmarkLogging(b *testing.B) {
	LOGGING_ONCE.Do(init_log)

	for i := 0; i < b.N; i++ {
		DBG("count: ------------------------------ ", i)
	}
}
