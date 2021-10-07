package main

// . "common"

var (
	Connections int64 = 0
	Bench_Loop  int64 = 0
)

func show_statistic() {
	// old := atomic.LoadInt64(&Connections)
	// for {
	// 	time.Sleep(1 * time.Second)
	// 	// LOG_STD("current thread count: ", runtime.GOMAXPROCS(-1))
	// 	// inc := atomic.LoadInt64(&Connections)
	// 	// ERR("Connections increase: ", inc-old, ", sum: ", inc)
	// 	// old = inc
	// }
}
