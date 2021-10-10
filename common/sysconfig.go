package common

import (
	"encoding/json"
	io "io/ioutil"
	"os"
	"sync"
	"syscall"
)

const (
	d_DEFAULT_FILE_SERVER_PORT    = "9999"
	d_DEFAULT_IO_THREADS          = 16
	d_DEFAULT_PRIOR_IO_THREADS    = 3
	d_DEFAULT_USE_PROTO_MSG       = true
	d_DEFAULT_PAGE_SIZE           = 4096
	d_DEFAULT_MAX_OPEN_FILES      = 102400
	d_DEFAULT_LOG_CNT_PER_FILE    = 2 * 1000 * 1000
	d_DEFAULT_USE_C_POOL_IO_SCHED = true
	d_DEFAULT_DATA_DIR            = "/home/stephen/devcloud/DATADIR"
)

type GlobalConfig struct {
	Port             string
	UserProtoMsg     bool
	PageSize         int
	MaxOpenFiles     int
	LogCountPerFile  int
	DataDir          string
	UserCPoolIoSched bool
	IoThreads        int
	PriorIoThreads   int
}

var (
	globalConfigIns  *GlobalConfig
	globalConfigOnce sync.Once
)

func getDefaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Port:            d_DEFAULT_FILE_SERVER_PORT,
		IoThreads:       d_DEFAULT_IO_THREADS,
		PriorIoThreads:  d_DEFAULT_PRIOR_IO_THREADS,
		UserProtoMsg:    d_DEFAULT_USE_PROTO_MSG,
		PageSize:        d_DEFAULT_PAGE_SIZE,
		MaxOpenFiles:    d_DEFAULT_MAX_OPEN_FILES,
		LogCountPerFile: d_DEFAULT_LOG_CNT_PER_FILE,
		DataDir:         d_DEFAULT_DATA_DIR,
	}
}

func GetGlobalConfigIns() *GlobalConfig {
	globalConfigOnce.Do(func() {
		for {
			var conf GlobalConfig
			data, err := io.ReadFile("./config.json") //read config file
			if err != nil {
				LOG_STD("Read json file error: ", err, " , set the defaut config.")
				conf = getDefaultGlobalConfig()
				globalConfigIns = &conf
				break
			}

			datajson := []byte(data)
			err = json.Unmarshal(datajson, &conf)
			if err != nil {
				LOG_STD("Unmarshal json file error: ", err, " , set the defaut config.")
				conf = getDefaultGlobalConfig()
				globalConfigIns = &conf
				break
			}
			globalConfigIns = &conf
			break
		}

		data, err := json.Marshal(globalConfigIns)
		if err != nil {
			LOG_STD("Config maybe dmaged: ", globalConfigIns, ", error: ", err)
			os.Exit(-1)
		} else {
			LOG_STD(string(data))
		}
	})

	return globalConfigIns
}

func InitMaxOpenFiles() bool {
	return SetMaxOpenFiles(uint64(getDefaultGlobalConfig().MaxOpenFiles), uint64(getDefaultGlobalConfig().MaxOpenFiles))
}

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
