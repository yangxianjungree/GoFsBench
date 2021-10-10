# Go Filesystem bench

此仓是使用Go实现的一个远端文件系统服务。用于记录系统调优过程，并解决Go处理阻塞的文件IO而创建过多线程的问题。

## 1. 配置

在客户端和服务端分别有一个 config.json 配置文件，说明如下：

```json
{
 "Port":"9999", // 服务端口号
 "UserProtoMsg":true, // 是否使用protobuf协议，否则使用json协议
 "PageSize":4096, // 每条连接每次最多读取数据量
 "MaxOpenFiles":102400, // 设置服务进程最大可使用的文件句柄数
 "LogCountPerFile":2000000, // 单个日志文件最多记录日志行数
 "DataDir":"xx", // 存放文件的上层目录
 "UserCPoolIoSched":true, // 是否使用cgo IO线程来处理文件IO
 "IoThreads":16, // 启用cgo IO线程数量
 "PriorIoThreads":3 // 启用高优先级cgo IO线程数量
}
```

## 2. 服务端

代码在 fs_server 目录。服务端提供文件的上传、下载、查询、协议性能测试等服务。

使用方式：

```s
$ make clean
$ make
$ ./FsServer // 如果开启了具有高优先级cgo IO线程，请使用 sudo ./FsServer
{"Port":"9999","UserProtoMsg":true,"PageSize":4096,"MaxOpenFiles":102400,"LogCountPerFile":2000000,"UserCPoolIoSched":true,"IoThreads":64,"PriorIoThreads":3}
raise pprof http server.....
```

### 3. 客户端

代码在 fs_client 目录。客户端主要是用测试各个接口的性能，并未提供正常的客户端功能。客户端上传、下载只使用内存，并未实际从本地磁盘读写文件。这点主要参考了fastdfs的性能测试方法。如果客户端和服务端的网络带宽不如本地磁盘速度，建议在同同一台机器上运行客户端和服务端，这样可以减少网络带来的影响。

使用方式：

```s
$ make clean
$ make
$ ./FsClient [clean/bench/upload/download/exist/delete] corotine_num loop_num [file_size]

```

运行示例：

```s
$ ./FsClient upload 500 2000000 5000
{"Port":"9999","UserProtoMsg":true,"PageSize":4096,"MaxOpenFiles":102400,"LogCountPerFile":2000000,"UserCPoolIoSched":true,"IoThreads":16,"PriorIoThreads":3}
Parmas:  &{3 500 2000000 5000}
Total success tasks:  2000000 , cost:  252746 ms.
qps:  7913.0826996272945
```

- clean: 用来清理性能测试中可能留下的文件
- bench: 测试协议性能
- upload: 测试上传功能性能
- download: 测试下载性能
- exist: 测试查询性能
- delete: 测试删除性能
