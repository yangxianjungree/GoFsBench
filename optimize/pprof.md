# 使用 pprof 工具调试流程

## 1. 命令

```s
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=60
```

## 2. 记录

### 2.1 使用 gosched 主动调度

common.BockingUtilDoneChannel 之前是使用以下代码实现的。这会导致大量的CPU上下文切换。

```go
for {
    select {
    case <-done:
        return
    default:
        runtime.gosched()
    }
}
```

- go 线程池

```s
(pprof) top 20
Showing nodes accounting for 96.33s, 95.48% of 100.89s total
Dropped 250 nodes (cum <= 0.50s)
Showing top 20 nodes out of 88
      flat  flat%   sum%        cum   cum%
    68.30s 67.70% 67.70%     68.30s 67.70%  runtime.futex
     7.08s  7.02% 74.72%      7.18s  7.12%  syscall.Syscall
     5.04s  5.00% 79.71%      5.06s  5.02%  syscall.Syscall6
     3.10s  3.07% 82.78%      4.93s  4.89%  common.BockingUtilDoneChannel
     1.94s  1.92% 84.71%      2.31s  2.29%  runtime.casgstatus
     1.91s  1.89% 86.60%      4.21s  4.17%  runtime.lock2
     1.54s  1.53% 88.13%      1.54s  1.53%  runtime.empty (inline)
     1.37s  1.36% 89.48%      1.37s  1.36%  runtime.(*gQueue).pop (inline)
     1.28s  1.27% 90.75%     68.56s 67.96%  runtime.unlock2
     1.03s  1.02% 91.77%      1.03s  1.02%  runtime.procyield
     0.90s  0.89% 92.67%     39.36s 39.01%  runtime.schedule
     0.60s  0.59% 93.26%      0.61s   0.6%  runtime.runqput
     0.56s  0.56% 93.82%      0.56s  0.56%  runtime.osyield
     0.51s  0.51% 94.32%      0.51s  0.51%  runtime.nanotime (inline)
     0.49s  0.49% 94.81%      1.71s  1.69%  runtime.execute
     0.22s  0.22% 95.02%     79.36s 78.66%  runtime.goschedImpl
     0.16s  0.16% 95.18%     80.24s 79.53%  runtime.mcall
     0.12s  0.12% 95.30%      1.73s  1.71%  runtime.chanrecv
     0.09s 0.089% 95.39%     35.57s 35.26%  runtime.findrunnable
     0.09s 0.089% 95.48%      2.07s  2.05%  runtime.globrunqget
```

- c 线程池

```s
(pprof) top 20
Showing nodes accounting for 104.56s, 97.63% of 107.10s total
Dropped 163 nodes (cum <= 0.54s)
Showing top 20 nodes out of 61
      flat  flat%   sum%        cum   cum%
    52.64s 49.15% 49.15%     52.64s 49.15%  runtime.futex
    32.73s 30.56% 79.71%     32.74s 30.57%  runtime.cgocall
     6.29s  5.87% 85.58%      6.31s  5.89%  syscall.Syscall
     5.35s  5.00% 90.58%      5.35s  5.00%  [libpthread-2.31.so]
     1.71s  1.60% 92.18%      2.81s  2.62%  common.BockingUtilDoneChannel
     1.14s  1.06% 93.24%      1.14s  1.06%  runtime.rtsigprocmask
     1.06s  0.99% 94.23%      1.23s  1.15%  runtime.globrunqget
     1.02s  0.95% 95.18%      1.05s  0.98%  runtime.chanrecv
     0.79s  0.74% 95.92%      0.79s  0.74%  runtime.casgstatus
     0.47s  0.44% 96.36%      1.31s  1.22%  runtime.lock2
     0.42s  0.39% 96.75%     43.64s 40.75%  runtime.schedule
     0.28s  0.26% 97.01%     27.40s 25.58%  runtime.unlock2
     0.27s  0.25% 97.26%      0.70s  0.65%  runtime.execute
     0.20s  0.19% 97.45%     38.67s 36.11%  runtime.goschedImpl
     0.06s 0.056% 97.51%      0.56s  0.52%  runtime.notesleep
     0.05s 0.047% 97.55%      1.09s  1.02%  runtime.selectnbrecv
     0.03s 0.028% 97.58%     42.34s 39.53%  main.upload_file
     0.03s 0.028% 97.61%      0.93s  0.87%  runtime.futexsleep
     0.01s 0.0093% 97.62%     14.30s 13.35%  common.(*FileWrapper).Write
     0.01s 0.0093% 97.63%     14.57s 13.60%  common.NetToDisk
```

### 2.2 go 直接调用 cgo 接口，将任务插入线程池队列

- c 线程池

```s
(pprof) top 20
Showing nodes accounting for 107.25s, 97.78% of 109.69s total
Dropped 152 nodes (cum <= 0.55s)
Showing top 20 nodes out of 60
      flat  flat%   sum%        cum   cum%
    54.65s 49.82% 49.82%     54.65s 49.82%  runtime.futex
    35.17s 32.06% 81.89%     35.18s 32.07%  runtime.cgocall
     6.39s  5.83% 87.71%      6.42s  5.85%  syscall.Syscall
     3.47s  3.16% 90.87%      3.47s  3.16%  [libpthread-2.31.so]
     1.85s  1.69% 92.56%      3.15s  2.87%  common.BockingUtilDoneChannel
     1.29s  1.18% 93.74%      1.29s  1.18%  runtime.chanrecv
     1.18s  1.08% 94.81%      1.18s  1.08%  runtime.rtsigprocmask
     0.91s  0.83% 95.64%      1.08s  0.98%  runtime.globrunqget
     0.65s  0.59% 96.23%      0.66s   0.6%  runtime.casgstatus
     0.49s  0.45% 96.68%     46.44s 42.34%  runtime.schedule
     0.38s  0.35% 97.03%      1.04s  0.95%  runtime.lock2
     0.28s  0.26% 97.28%      0.73s  0.67%  runtime.execute
     0.24s  0.22% 97.50%     25.21s 22.98%  runtime.unlock2
     0.12s  0.11% 97.61%     36.81s 33.56%  runtime.goschedImpl
     0.07s 0.064% 97.68%      0.63s  0.57%  runtime.notesleep
     0.03s 0.027% 97.70%      0.71s  0.65%  runtime.cgocallback
     0.02s 0.018% 97.72%      0.59s  0.54%  common.waitOpenCallBack
     0.02s 0.018% 97.74%     45.46s 41.44%  main.handle_conn
     0.02s 0.018% 97.76%      6.06s  5.52%  net.(*netFD).Write
     0.02s 0.018% 97.78%     58.60s 53.42%  runtime.mcall
```

### 2.3 go 将封装好的文件操作任务放入带缓冲的 channel 中，然后分别分发给线程池、协程池

- go 协程池

```s
(pprof) top 20
Showing nodes accounting for 23.43s, 88.22% of 26.56s total
Dropped 240 nodes (cum <= 0.13s)
Showing top 20 nodes out of 122
      flat  flat%   sum%        cum   cum%
    12.47s 46.95% 46.95%     12.66s 47.67%  syscall.Syscall
     8.62s 32.45% 79.41%      8.72s 32.83%  syscall.Syscall6
     0.52s  1.96% 81.36%      0.52s  1.96%  runtime.futex
     0.21s  0.79% 82.15%      0.22s  0.83%  runtime.removespecial
     0.20s  0.75% 82.91%      0.20s  0.75%  runtime.epollwait
     0.20s  0.75% 83.66%      0.35s  1.32%  runtime.scanobject
     0.16s   0.6% 84.26%      0.17s  0.64%  runtime.findObject
     0.14s  0.53% 84.79%      0.14s  0.53%  runtime.madvise
     0.14s  0.53% 85.32%      0.14s  0.53%  runtime.nextFreeFast (inline)
     0.13s  0.49% 85.81%      0.25s  0.94%  runtime.pcvalue
     0.11s  0.41% 86.22%      0.65s  2.45%  runtime.gentraceback
     0.10s  0.38% 86.60%      0.18s  0.68%  runtime.exitsyscall
     0.09s  0.34% 86.94%      0.70s  2.64%  runtime.schedule
     0.08s   0.3% 87.24%      1.08s  4.07%  runtime.mallocgc
     0.06s  0.23% 87.46%      0.21s  0.79%  runtime.scanblock
     0.05s  0.19% 87.65%      1.60s  6.02%  internal/poll.(*FD).Read
     0.05s  0.19% 87.84%      0.57s  2.15%  log.(*Logger).Output
     0.04s  0.15% 87.99%      0.65s  2.45%  runtime.makeslice
     0.03s  0.11% 88.10%      5.27s 19.84%  common.send_message_proto
     0.03s  0.11% 88.22%      0.73s  2.75%  encoding/binary.Read
```

```s
(pprof) top 20
Showing nodes accounting for 13.79s, 87.95% of 15.68s total
Dropped 217 nodes (cum <= 0.08s)
Showing top 20 nodes out of 143
      flat  flat%   sum%        cum   cum%
     6.14s 39.16% 39.16%      6.27s 39.99%  syscall.Syscall
     5.34s 34.06% 73.21%      5.35s 34.12%  syscall.Syscall6
     0.77s  4.91% 78.12%      0.77s  4.91%  runtime.futex
     0.16s  1.02% 79.15%      0.32s  2.04%  runtime.scanobject
     0.15s  0.96% 80.10%      0.15s  0.96%  runtime.epollwait
     0.12s  0.77% 80.87%      0.12s  0.77%  runtime.procyield
     0.10s  0.64% 81.51%      0.41s  2.61%  runtime.gentraceback
     0.10s  0.64% 82.14%      0.11s   0.7%  runtime.lock2
     0.09s  0.57% 82.72%      0.10s  0.64%  runtime.casgstatus
     0.09s  0.57% 83.29%      0.09s  0.57%  runtime.gopark
     0.09s  0.57% 83.86%      0.09s  0.57%  runtime.madvise
     0.09s  0.57% 84.44%      0.09s  0.57%  runtime.memclrNoHeapPointers
     0.09s  0.57% 85.01%      0.41s  2.61%  runtime.send
     0.08s  0.51% 85.52%      0.19s  1.21%  runtime.chanrecv
     0.08s  0.51% 86.03%      0.08s  0.51%  runtime.nanotime
     0.08s  0.51% 86.54%      0.19s  1.21%  runtime.pcvalue
     0.08s  0.51% 87.05%      0.10s  0.64%  runtime.step
     0.06s  0.38% 87.44%      0.79s  5.04%  runtime.mallocgc
     0.04s  0.26% 87.69%      1.96s 12.50%  net.(*netFD).Write
     0.04s  0.26% 87.95%      0.08s  0.51%  runtime.netpollready
```

- c 线程池

```s
(pprof) top 20
Showing nodes accounting for 31.14s, 97.92% of 31.80s total
Dropped 120 nodes (cum <= 0.16s)
Showing top 20 nodes out of 62
      flat  flat%   sum%        cum   cum%
    24.43s 76.82% 76.82%     24.46s 76.92%  runtime.cgocall
     2.39s  7.52% 84.34%      2.39s  7.52%  runtime.futex
     1.71s  5.38% 89.72%      1.71s  5.38%  [libpthread-2.31.so]
     1.71s  5.38% 95.09%      1.72s  5.41%  syscall.Syscall
     0.45s  1.42% 96.51%      0.45s  1.42%  runtime.rtsigprocmask
     0.38s  1.19% 97.70%      0.38s  1.19%  [libc-2.31.so]
     0.01s 0.031% 97.74%      1.32s  4.15%  runtime.chansend
     0.01s 0.031% 97.77%      0.23s  0.72%  runtime.gcDrain
     0.01s 0.031% 97.80%      0.18s  0.57%  runtime.markroot
     0.01s 0.031% 97.83%      0.29s  0.91%  runtime.needm
     0.01s 0.031% 97.86%      1.30s  4.09%  runtime.send
     0.01s 0.031% 97.89%      1.67s  5.25%  runtime.systemstack
     0.01s 0.031% 97.92%      1.67s  5.25%  syscall.write
         0     0% 97.92%      0.53s  1.67%  _cgoexp_febed96ba46c_go_done_callback
         0     0% 97.92%      0.16s   0.5%  _cgoexp_febed96ba46c_go_done_close_callback
         0     0% 97.92%      0.35s  1.10%  _cgoexp_febed96ba46c_go_done_open_callback
         0     0% 97.92%      0.28s  0.88%  _cgoexp_febed96ba46c_go_done_rename_callback
         0     0% 97.92%      4.99s 15.69%  common.(*FileWrapper).Close
         0     0% 97.92%      9.72s 30.57%  common.(*FileWrapper).Write
         0     0% 97.92%      1.67s  5.25%  common.(*FsProtocol).SendMsg

```
