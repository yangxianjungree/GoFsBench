# Rename

rename系统调用会调用lockname，而lockname会在参数中的两个路径的共同父目录或者文件系统中加锁。因此，在Go频繁调用rename，会创建大量的线程，故而每个线程都会竞争stroges和tmp目录（两个目录都在DATADIR目录下）的父目录中的锁。

```s
$ ~/devcloud/go_root/src/fs_server$ ps aux | grep Client
stephen   212821  137  0.0 7953668 4800 pts/2    Dl+  20:11   1:42 ./FsClient rename 500 2000000 5000
stephen   213914  0.0  0.0  12116   732 pts/8    S+   20:12   0:00 grep --color=auto Client
$ ~/devcloud/go_root/src/fs_server$ pstree -aps 212821 | wc -l
513
```

可以将这些文件按照按照一定的规则分配二级目录，比如我测试用的index，然后对10取余。然后分别在不同子目录中调用rename。下面是2000000个文件的测试数据：

|2000000个文件|rename(./stroges/file_x, ./tmp/file_x.tmp)|rename(./x%10/file_x, ./x%10/file_x.tmp)|
|--|--|--|
|第一次|183174 ms|33383 ms|
|第二次|320616 ms|45914 ms|
|第三次|298331 ms|45187 ms|
|从总目录到各个子目录|106332 ms|-|
|从子目录到总目录|284959 ms|-|
