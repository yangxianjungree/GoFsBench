# Go Filesystem bench

此仓是使用Go实现的一个远端文件系统服务。用于记录调优过程，并解决Go处理阻塞的文件IO而创建过多线程的问题。

## 1. 配置

在客户端和服务端分别有一个 config.json 配置文件，说明如下：

```json
{
 "Port":"9999", // 服务端口号
 "UserProtoMsg":true, // 是否使用protobuf协议，否则使用json协议
 "PageSize":4096, // 每条连接每次最多读取数据量
 "MaxOpenFiles":102400, // 设置服务进程最大可使用的文件句柄数
 "LogCountPerFile":2000000, // 单个日志文件最多记录日志行数
 "DataDir":"xx", // Server存放文件的上层目录
 "UserPoolIoSched":"G", // C: 启用 c 线程池来处理文件; G: 启用 go 协程池; 其他：不使用 IO 池
 "IoThreads":16, // 启用 IO 池中的线程/协程数量
 "PriorIoThreads":3, // 启用高优先级cgo IO线程数量
 "SetCpuAffinity":true, // 是否设置cgo中IO线程的CPU黏合性
 "WaitingQueueLen":1000000 // 启用IO池的带缓冲任务队列长度
}
```

## 2. 服务端

代码在 fs_server 目录。服务端提供文件的上传、下载、查询、协议性能测试等服务。

使用方式：

```s
$ make clean
$ make
$ ./FsServer 
{"Port":"9999","UserProtoMsg":true,"PageSize":4096,"MaxOpenFiles":102400,"LogCountPerFile":2000000,"DataDir":"/home/stephen/devcloud/DATADIR","UserPoolIoSched":"C","IoThreads":32,"PriorIoThreads":0,"SetCpuAffinity":true,"WaitingQueueLen":1000000}
current thread id: 140694714423040, CPU 0
current thread id: 140694689244928, CPU 3
current thread id: 140694706030336, CPU 1
current thread id: 140694337353472, CPU 4
current thread id: 140694320568064, CPU 1
current thread id: 140694697637632, CPU 2
...
Use C io thread pool.......
raise pprof http server....
```

## 3. 客户端

代码在 fs_client 目录。客户端主要是用于测试各个接口的性能，并未提供正常的客户端功能。客户端上传、下载只使用内存，并未实际从本地磁盘读写文件。这点主要参考了fastdfs的性能测试方法。如果客户端和服务端的网络带宽不如本地磁盘速度，建议在同同一台机器上运行客户端和服务端，这样可以减少网络带来的影响。

使用方式：

```s
$ make clean
$ make
$ ./FsClient
{"Port":"9999","UserProtoMsg":true,"PageSize":4096,"MaxOpenFiles":102400,"LogCountPerFile":2000000,"DataDir":"","UserPoolIoSched":"","IoThreads":0,"PriorIoThreads":0,"SetCpuAffinity":false,"WaitingQueueLen":0}
[clean/bench/upload/download/exist/delete] corotine_num loop_num [file_size]
```

运行示例：

```s
$ ./FsClient upload 500 2000000 5000
{"Port":"9999","UserProtoMsg":true,"PageSize":4096,"MaxOpenFiles":102400,"LogCountPerFile":2000000,"DataDir":"","UserPoolIoSched":"","IoThreads":0,"PriorIoThreads":0,"SetCpuAffinity":false,"WaitingQueueLen":0}
Parmas:  &{3 500 2000000 5000}
Total success tasks:  2000000 , cost:  252746 ms.
qps:  7913.0826996272945
```

- 测试接口：
  - clean: 用来清理性能测试中可能留下的文件
  - bench: 测试协议性能
  - upload: 测试上传功能性能
  - download: 测试下载性能
  - exist: 测试查询性能
  - delete: 测试删除性能
- corotine_num： 并发数，即使用的协程数
- loop_num： 总共请求次数
- file_size： 请求文件大小（只用于upload接口，其他可以不填）

## 4. 解决Go的IO线程问题

使用过Go调度机制的伙伴应该知道，对于 Linux 中的任何文件，如果在文件的 ```f_op``` 中没有实现 ```poll```函数，是没法使用 epoll、poll 等多路复用机制的。所以 Go 的做法是，对于那些没有实现 ```poll```方法的磁盘文件，如果IO操作阻塞一段时间，影响到同一线程（M）中其他协程（G）的运行了，就创建一个新的线程单独服务这个阻塞 IO。对应远端文件服务系统而言，瓶颈就在于磁盘性能，也就是文件 IO 操作，最终导致系统创建大量线程，基本上就是一条连接一个线程。而Go默认支持创建10000个线程，也就是要超过1w的并发，就要修改runtime里面的MaxProcs。但是想要支持更多，系统可能就 ```ulimit -u``` 又要有限制了。

其中的难点就是Go调度机制对于大量文件IO的场景不适合。

此项目结合线程池、cgo、系统性能分析工具、Go的pprof大杀器，解决了这个问题。

主要的解决思路：

- 封装文件IO操作任务，将Go中会导致创建线程的接口独立封装起来；
- 使用channel等待文件IO操作完成，将线程的执行权交给其他协程，避免IO操作等待时间过长而创建线程；
- 执行文件IO任务：
  - （1）结合cgo将文件IO操作任务放入到glibc下的线程池中去完成；调用glibc的接口完成磁盘IO；
  - （2）将任务放到协程池中执行。

### 4.1 数据示例

- **PAGE_SIZE = 4k， 请求文件50k，请求数200,000，500并发**

|实现方式/数据|耗时ms|qps|创建的系统线程数|
|--|--|--|--|
|50K不使用IO池化|169934|1176|512|
|Go协程池：32|173167|1154|48|
|C线程池：32，使用CPU黏合，不使用优先级协程|191358|1045|50|
|C线程池：32，不使用CPU黏合，不使用优先级协程|181389|1102|50|

## 5. 性能测试记录

[虚拟机规格/日志性能](bench_doc/machine_info.md)

[fio磁盘性能测试](bench_doc/disk.md)

[rename优化](optimize/rename.md)

[协议解析性能](bench_doc/bench_protocol.md)

[原始 Go IO 接口性能](bench_doc/bench_origin.md)

[IO 池化性能](bench_doc/bench_io_pool.md)

## 6. 结论

使用IO池的思想是源于fastdfs项目，如果对fastdfs（使用C实现的）感兴趣，可以去看一下。github中还有一个go实现的fastdfs，但我没有看到代码中解决线程创建过多的实现方式，可能我阅读的不够仔细。

基于Go协程的实现，基本上等待这些协程在执行IO操作时，让调度器将它们升级为线程，这样就可以达到目的了。这样实现的优点是：实现简单，思路清晰；

而基于cgo实现的IO线程池，需要考虑在并发场景下，将IO操作任务从Go插入到C中任务队列的调用开销、锁竞争问题、线程池实现、任务完成后通知调用者等问题。优点是：让自己有一定的成就感，可以在C代码中控制线程的属性（优先级、CPU黏合性）、生命周期等。

 希望我的这个项目能给大家带来一些帮助。也非常期待大家的指教和意见。
