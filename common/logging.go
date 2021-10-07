package common

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync/atomic"
)

var (
	log_hander           *log.Logger = nil
	cur_log_cnt          int64       = 0
	cur_log_index        int64       = 0
	log_file_name_prefix             = "log"
)

const (
	DBG_MODE = 1
)

func LOG_STD(a ...interface{}) (int, error) {
	return fmt.Println(a...)
}

func SetLogName(name string) {
	log_file_name_prefix = name
}

func InitLog(i int64) error {
	cur_log_cnt = i
	log_name := "./" + log_file_name_prefix + ".log" + strconv.FormatInt(i, 10)
	f, err := os.OpenFile(log_name, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0666)
	if err != nil {
		LOG_STD("Open log file failed, error: ", err)
		return err
	}

	log_hander = log.New(f, "["+log_file_name_prefix+"]", log.Ldate|log.Ltime)
	return nil
}

func ERR(a ...interface{}) {
	if log_hander == nil {
		return
	}

	old := atomic.AddInt64(&cur_log_index, 1)
	loop_cnt := 0
	for old >= LOG_CNT_PER_FILE {
		loop_cnt += 1
		if atomic.CompareAndSwapInt64(&cur_log_index, old, 0) {
			cur_log_cnt += 1
			InitLog(cur_log_cnt)
		}
		old = atomic.AddInt64(&cur_log_index, 1)
		if loop_cnt > 10 {
			LOG_STD("Had CAS for ", loop_cnt, " times.")
		}
	}
	log_hander.Println(a...)
}

func DBG(a ...interface{}) {
	if DBG_MODE != 1 {
		return
	}
	ERR(a...)
}
