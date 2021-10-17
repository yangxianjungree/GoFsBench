# 磁盘性能指标分析

## 1. 规格

- 测试对象：机械硬盘

- 测试指标：IOPS和MBPS(吞吐率)

- 测试工具：Fio、dd

- 测试参数:  IO大小，寻址空间，队列深度，读写模式，随机/顺序模式

- 测试方法： 先测试 dd， 然后用 fio 测试

### 1.1 磁盘支持队列深度

使用 lsscsi 工具查看磁盘支持队列深度

```s
$ lsscsi -l
[3:0:0:0]    cd/dvd  NECVMWar VMware SATA CD01 1.00  /dev/sr0 
  state=running queue_depth=1 scsi_level=6 type=5 device_blocked=0 timeout=30
[32:0:0:0]   disk    VMware,  VMware Virtual S 1.0   /dev/sda 
  state=running queue_depth=32 scsi_level=3 type=0 device_blocked=0 timeout=180
```

- 但是经过实际测试，使用IO深度为 64 才能将磁盘使用率提升接近至 100%。

iostat 查看磁盘sda利用率

```s
$ iostat -x sda 3
...
```

## 2. 数据记录

### 2.1 使用 dd 工具测试

```s
dd if=/dev/zero of=test.img bs=4k count=1M
```

|40G test.img|MBPS(吞吐率)|耗时|
|--|--|--|
|bs=4k|145 MB/s|295.238 s|
|bs=40k|118 MB/s|365.417 s|
|bs=400k|109 MB/s|384.626 s|
|bs=4m|123 MB/s|347.802 s|
|bs=40m|133 MB/s|322.477 s|
|bs=400m|135 MB/s|311.3 s|
|bs=4g(只写入20G)|162 MB/s|132.488 s|

### 2.2 使用 fio 工具测试

FIO的测试参数：

- ioengine: 负载引擎，我们一般使用libaio，发起异步IO请求。

- bs: IO大小

- direct: 直写，绕过操作系统Cache。因为我们测试的是硬盘，而不是操作系统的Cache，所以设置为1。

- rw: 读写模式，有顺序写write、顺序读read、随机写randwrite、随机读randread等。

- size: 寻址空间，IO会落在 [0, size)这个区间的硬盘空间上。这是一个可以影响IOPS的参数。一般设置为硬盘的大小。

- filename: 测试对象

- iodepth: 队列深度，只有使用libaio时才有意义。这是一个可以影响IOPS的参数。

- runtime: 测试时长

```s
写：
$ fio -ioengine=libaio -bs=4k -direct=1 -thread -rw=randwrite -size=20G -filename=./test.img -name="EBS 4K randwrite test" -iodepth=64 -runtime=60 -numjobs=1

读：
$ fio -ioengine=libaio -bs=4k -direct=1 -thread -rw=randread -size=20G -filename=/dev/sda -name="EBS 4K randwrite test" -iodepth=64 -runtime=60 -numjobs=1
```

|20G test.img bs=4k|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|7842KiB/s|96.38%|2251.32|
|write|38.4MiB/s|96.39%|10070.98|
|randread|17.8MiB/s|99.98%|4552.33|
|read|69.7MiB/s|100.00%|17775.17|

|20G test.img bs=40k|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|19.8MiB/s|99.75%|509.78|
|write|30.8MiB/s|97.60%|789.85|
|randread|88.2MiB/s|100.00%|2261.43|
|read|365MiB/s|100.00%|9311.67|

|20G test.img bs=400k|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|38.2MiB/s|98.04%|103.11|
|write|83.1MiB/s|98.75%|215.85|
|randread|210MiB/s|99.91%|539.57|
|read|535MiB/s|99.94%|1305.39|

|20G test.img bs=4m|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|50.3MiB/s|98.87%|21.06|
|write|57.4MiB/s|98.49%|24.84|
|randread|488MiB/s|99.84%|122.76|
|read|517MiB/s|99.80%|127.38|

|20G test.img bs=40m|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|33.9MiB/s|95.09%||
|write|96.5MiB/s|97.54%|4.03|
|randread|603MiB/s|98.39%|14.37|
|read|488MiB/s|94.09%|14.55|

|20G test.img bs=400m|MBPS(吞吐率)|util|iops|
|--|--|--|--|
|randwrite|core dumped|||
