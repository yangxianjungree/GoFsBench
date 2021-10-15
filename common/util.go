package common

import (
	"math"
	"math/rand"
	"runtime"
	"time"
)

type BenchBoard struct {
	success_tasks uint64
	start         time.Time
}

func NewBenchBoard(tasks uint64) *BenchBoard {
	return &BenchBoard{
		start:         time.Now(),
		success_tasks: tasks,
	}
}

func (b *BenchBoard) SetSuccessTasks(s uint64) {
	b.success_tasks = s
}

func (b *BenchBoard) ShowBenchBoard() {
	gap := float64(time.Since(b.start).Milliseconds())
	LOG_STD("Total success tasks: ", b.success_tasks, ", cost: ", gap, "ms.")
	if gap != 0.0 {
		LOG_STD("qps: ", (float64(b.success_tasks)/gap)*1000)
	}
}

func PanicStackInfo() []byte {
	buf := make([]byte, PANIC_STACK_SIZE)
	for {
		n := runtime.Stack(buf, true)
		if n == len(buf) {
			buf = make([]byte, PAGE_SIZE<<1)
			continue
		}
		break
	}
	return buf
}

//获取一个n位随机数
func GetRand(n int) int {
	//设置随机数动态种子
	rand.Seed(time.Now().UnixNano())
	//求出随机数的位数上限
	pow10 := math.Pow10(n)
	//获取随机数
	return rand.Intn(int(pow10))
}

func BockingUtilDoneChannel(done chan bool) {
	<-done

	// for {
	// 	select {
	// 	case <-done:
	// 		return
	// 	default:
	// 		runtime.Gosched()
	// 		// time.Sleep(time.Duration(GetRand(4)) * time.Nanosecond)
	// 	}
	// }
}
