package common

// import (
// 	"testing"
// 	"time"
// )

// func doTask(v interface{}) {
// 	LOG_STD("Begin task........")
// 	time.Sleep(1 * time.Second)
// 	LOG_STD("Do task over........")
// }

// func TestTaskPush(t *testing.T) {
// 	LOG_STD("Unit test.................")

// 	task := &TaskElem{
// 		done: make(chan bool),
// 		task: doTask,
// 	}
// 	defer close(task.done)

// 	pool := NewTaskPool(NewTaskPoolBuckets(16, 200))
// 	pool.PushTask(task)

// 	go func() {
// 		<-task.done
// 		LOG_STD("Recv done msg....")
// 	}()

// 	pool.WaitAll()
// }
