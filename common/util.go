package common

import (
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
