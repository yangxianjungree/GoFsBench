package common

import (
	"sync"
)

type TaskElem struct {
	done chan bool
	task func(v interface{})
	args interface{}
}

type TaskConfig struct {
	buckets     int
	maxTaskSize int64
}

type TaskPool struct {
	*TaskConfig
	control []chan bool
	tasks   chan *TaskElem
	wg      sync.WaitGroup
}

func NewTaskConfig() *TaskConfig {
	return &TaskConfig{
		buckets:     16,
		maxTaskSize: 5000,
	}
}

func NewTaskPoolBuckets(count uint32, tasks int64) *TaskConfig {
	if count == 0 || !(count&(^count+1) == count) {
		count = 16
	}

	if tasks <= 0 {
		tasks = 5000
	}

	c := &TaskConfig{
		buckets:     int(count),
		maxTaskSize: tasks,
	}
	return c
}

func NewTaskPool(c *TaskConfig) *TaskPool {
	t := &TaskPool{
		TaskConfig: c,
		tasks:      make(chan *TaskElem, c.maxTaskSize),
		wg:         sync.WaitGroup{},
	}
	for i := 0; i < c.buckets; i++ {
		b := make(chan bool, 10)
		t.control = append(t.control, b)
		t.wg.Add(1)
		go t.taskWorkerFunc(i)
	}
	return t
}

func (t *TaskPool) PushTask(task *TaskElem) {
	t.tasks <- task
}

func (t *TaskPool) taskWorkerFunc(index int) {
	defer func() {
		close(t.control[index])
		t.wg.Done()
		LOG_STD("TaskPool worker ", index, " quit.")
	}()

	for {
		select {
		case quit := <-t.control[index]:
			if quit {
				return
			}
		case task := <-t.tasks:
			// LOG_STD("TaskPool worker ", index, " get task.")
			task.task(task.args)
			task.done <- true
		}
	}
}

func (t *TaskPool) WaitAll() {
	for i := 0; i < t.buckets; i++ {
		t.control[i] <- true
	}
	t.wg.Wait()
}
