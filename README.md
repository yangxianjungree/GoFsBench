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

### 4. 性能测试记录

[磁盘测试](bench_doc/disk.md)

[rename优化](optimize/rename.md)

[本地网络测试](bench_doc/peer.md)

[使用原始go IO接口测试](bench_doc/bench_origin.md)

[使用原始glibc IO接口测试](bench_doc/bench_cpool.md)

### 5. 解决Go的IO线程问题

使用过Go调度机制的伙伴应该知道，对于没有实现poll方法的文件IO系统调用的库封装函数，是没法使用epoll、poll等多路复用机制的，所以Go的做法是，当IO函数阻塞一段时间影响到同一线程（M）中其他协程（G）的运行了，就创建一个新的线程单独服务这个阻塞IO。对应远端文件服务系统而言，瓶颈就在于磁盘性能，也就是文件IO操作，最终导致系统创建大量线程，基本上就是一条连接一个线程。而Go默认支持创建10000个线程，也就是要超过1w的并发，就要修改runtime里面的MaxProcs。但是想要支持更多，系统可能就 ulimit -u 又要有限制了。

其中的难点就是Go调度机制对于大量文件IO的场景不适合。

此项目结合线程池、cgo、系统性能分析工具、Go的pprof大杀器，解决了这个问题。

主要的解决思路：

- 封装文件IO操作任务，将Go中会导致创建线程的接口独立封装起来；
- 使用channel等待文件IO操作完成， 使用gosched让出协程，避免chanel等待时间过长而创建线程；
- 结合cgo将文件IO操作任务放入到glibc下的线程池中去完成；
- 调用glibc的接口完成磁盘IO；

### 6. TODO

安装上述方式基本实现了上传、下载文件的接口，服务创建的线程确实没有随并发线性增长。但是使用 vmstat pidstat 等工具发现服务导致线程的上下文切换太多。我试过使用优先级线程，效果不是很明显。

希望我的这个项目能给大家带来一些帮助。也非常期待大家的指教和意见。
