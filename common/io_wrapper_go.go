package common

import (
	"os"
	"sync"
)

var (
	ioPool     *TaskPool = nil
	ioPoolOnce sync.Once
)

func initGoIoPool() {
	ioPoolOnce.Do(func() {
		if GetGlobalConfigIns().UseGoIoPool() {
			threads := GetGlobalConfigIns().IoThreads
			queueLen := GetGlobalConfigIns().WaitingQueueLen
			ioPool = NewTaskPool(NewTaskPoolBuckets(threads, queueLen))
			LOG_STD("Use go io thread pool.......")
		}
	})
}

type openArgs struct {
	name string
	flag int
	perm os.FileMode
	f    *os.File
	err  error
}

func doOpen(v interface{}) {
	task := v.(*openArgs)
	task.f, task.err = os.OpenFile(task.name, task.flag, task.perm)
}

type readArgs struct {
	f   *os.File
	buf []byte
	n   int
	err error
}

func doRead(v interface{}) {
	task := v.(*readArgs)
	task.n, task.err = task.f.Read(task.buf)
}

type writeArgs struct {
	f   *os.File
	buf []byte
	n   int
	err error
}

func doWrite(v interface{}) {
	task := v.(*writeArgs)
	task.n, task.err = task.f.Write(task.buf)
}

type closeArgs struct {
	f   *os.File
	err error
}

func doClose(v interface{}) {
	task := v.(*closeArgs)
	task.err = task.f.Close()
}

type renameArgs struct {
	old string
	nw  string
	err error
}

func doRename(v interface{}) {
	task := v.(*renameArgs)
	task.err = os.Rename(task.old, task.nw)
}

func pushOpenTask(name string, flag int, perm os.FileMode, done chan bool) (*os.File, error) {
	openArgs := &openArgs{
		name: name,
		flag: flag,
		perm: perm,
		f:    nil,
		err:  nil,
	}

	task := &TaskElem{
		done: done,
		task: doOpen,
		args: openArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return openArgs.f, openArgs.err
}

func pushReadTask(f *os.File, b []byte, done chan bool) (int, error) {
	readArgs := &readArgs{
		f:   f,
		buf: b,
		n:   0,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: doRead,
		args: readArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return readArgs.n, readArgs.err
}

func pushWriteTask(f *os.File, b []byte, done chan bool) (int, error) {
	writeArgs := &writeArgs{
		f:   f,
		buf: b,
		n:   0,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: doWrite,
		args: writeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return writeArgs.n, writeArgs.err
}

func pushCloseTask(f *os.File, done chan bool) error {
	closeArgs := &closeArgs{
		f:   f,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: doClose,
		args: closeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return closeArgs.err
}

func pushRenameTask(oldname, newname string) error {
	writeArgs := &renameArgs{
		old: oldname,
		nw:  newname,
		err: nil,
	}

	done := make(chan bool)
	task := &TaskElem{
		done: done,
		task: doRename,
		args: writeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return writeArgs.err
}
