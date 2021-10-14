package common

import (
	"encoding/json"
	io "io/ioutil"
	"os"
	"sync"
)

var (
	PAGE_SIZE        = GetGlobalConfigIns().PageSize
	LOG_CNT_PER_FILE = int64(getDefaultGlobalConfig().LogCountPerFile)
)

const (
	d_DEFAULT_FILE_SERVER_PORT  = "9999"                           // 服务端口号
	d_DEFAULT_USE_PROTO_MSG     = true                             // 是否使用 protobuf 协议，否则使用 json 协议
	d_DEFAULT_PAGE_SIZE         = 4096                             // 每条连接每次最多读取数据量
	d_DEFAULT_MAX_OPEN_FILES    = 102400                           // 设置服务进程最大可使用的文件句柄数
	d_DEFAULT_LOG_CNT_PER_FILE  = 2 * 1000 * 1000                  // 单个日志文件最多记录日志行数
	d_DEFAULT_DATA_DIR          = "/home/stephen/devcloud/DATADIR" // 存放文件的上层目录
	d_DEFAULT_USE_POOL_IO_SCHED = "G"                              // C: 启用 c 线程池来处理文件; G: 启用 go 线程池; 其他：不使用 IO 线程池
	d_DEFAULT_IO_THREADS        = 16                               // 启用 c/go IO 线程数量
	d_DEFAULT_PRIOR_IO_THREADS  = 3                                // 如果使用 cgo IO 线程池，启用高优先级的 cgo IO 线程数量
	d_DEFAULT_WAITING_QUEUE_LEN = 1000000                          // 启用高优先级 cgo IO 线程数量
)

type GlobalConfig struct {
	Port            string
	UserProtoMsg    bool
	PageSize        int
	MaxOpenFiles    int
	LogCountPerFile int
	DataDir         string
	UserPoolIoSched string
	IoThreads       uint32
	PriorIoThreads  uint32
	WaitingQueueLen int64
}

var (
	globalConfigIns  *GlobalConfig
	globalConfigOnce sync.Once
)

func getDefaultGlobalConfig() GlobalConfig {
	return GlobalConfig{
		Port:            d_DEFAULT_FILE_SERVER_PORT,
		UserProtoMsg:    d_DEFAULT_USE_PROTO_MSG,
		PageSize:        d_DEFAULT_PAGE_SIZE,
		MaxOpenFiles:    d_DEFAULT_MAX_OPEN_FILES,
		LogCountPerFile: d_DEFAULT_LOG_CNT_PER_FILE,
		DataDir:         d_DEFAULT_DATA_DIR,
		UserPoolIoSched: d_DEFAULT_USE_POOL_IO_SCHED,
		IoThreads:       d_DEFAULT_IO_THREADS,
		PriorIoThreads:  d_DEFAULT_PRIOR_IO_THREADS,
		WaitingQueueLen: d_DEFAULT_WAITING_QUEUE_LEN,
	}
}

func GetGlobalConfigIns() *GlobalConfig {
	globalConfigOnce.Do(func() {
		for {
			var conf GlobalConfig
			// todo:
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

func (c *GlobalConfig) UseCIoPool() bool {
	return c.UserPoolIoSched == "C"
}

func (c *GlobalConfig) UseGoIoPool() bool {
	return c.UserPoolIoSched == "G"
}
