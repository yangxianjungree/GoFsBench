# 机器基本信息

## 1. 机器规格

### 1.1 操作系统

```shell
# uname -a

Linux xxx 5.11.0-34-generic #36~20.04.1-Ubuntu SMP Fri Aug 27 08:06:32 UTC 2021 x86_64 x86_64 x86_64 GNU/Linux
```

### 1.2 硬件

- CPU
  英特尔 Core i7-8700 @ 3.20GHz 六核 @ 3.20GHz

- 内存
  8G

- 磁盘
  机械硬盘

- 网卡
  英特尔 Ethernet Connection  I219-V / 华擎 (千兆网卡)

  ```s
  iperf3 -s
  ...
  iperf3 -c 127.0.0.1 -t 60
  ...
  [ ID] Interval           Transfer     Bitrate         Retr
  [  5]   0.00-60.00  sec   331 GBytes  47.3 Gbits/sec    0             sender
  [  5]   0.00-60.00  sec   331 GBytes  47.3 Gbits/sec                  receiver
  ```

### 1.3 网络带宽测试

手头只有一台虚拟机，就测本地网络带宽吧。

服务端：

```s
$ iperf3 -s
...
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate
[  5]   0.00-61.03  sec   367 MBytes  50.5 Mbits/sec                  receiver
-----------------------------------------------------------
```

客户端：

```s
$ iperf3 -c 192.168.1.5 -t 60
...
```

## 2. 日志性能数据

```s
# go test -bench=. -cpu=16 -benchtime="5s" -benchmem

goos: linux
goarch: amd64
pkg: common
cpu: Intel(R) Core(TM) i7-8700 CPU @ 3.20GHz
BenchmarkLogging-16      2915629          2065 ns/op          56 B/op          2 allocs/op
PASS
ok      common  8.108s
```
