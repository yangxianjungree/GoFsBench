package common

import (
	"os"
	"sync"
)

var (
	ioPool     *TaskPool = nil
	ioPoolOnce sync.Once
)

func InitIoPool(gos int) {
	ioPoolOnce.Do(func() {
		ioPool = NewTaskPool(NewTaskPoolBuckets(10, 100000))
	})
}

type OpenArgs struct {
	name string
	flag int
	perm os.FileMode
	f    *os.File
	err  error
}

func DoOpen(v interface{}) {
	task := v.(*OpenArgs)
	task.f, task.err = os.OpenFile(task.name, task.flag, task.perm)
}

func PushOpenTask(name string, flag int, perm os.FileMode, done chan bool) (*os.File, error) {
	openArgs := &OpenArgs{
		name: name,
		flag: flag,
		perm: perm,
		f:    nil,
		err:  nil,
	}

	task := &TaskElem{
		done: done,
		task: DoOpen,
		args: openArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return openArgs.f, openArgs.err
}

type ReadArgs struct {
	f   *os.File
	buf []byte
	n   int
	err error
}

func DoRead(v interface{}) {
	task := v.(*ReadArgs)
	task.n, task.err = task.f.Read(task.buf)
}

func PushReadTask(f *os.File, b []byte, done chan bool) (int, error) {
	readArgs := &ReadArgs{
		f:   f,
		buf: b,
		n:   0,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: DoRead,
		args: readArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return readArgs.n, readArgs.err
}

type WriteArgs struct {
	f   *os.File
	buf []byte
	n   int
	err error
}

func DoWrite(v interface{}) {
	task := v.(*WriteArgs)
	task.n, task.err = task.f.Write(task.buf)
}

func PushWriteTask(f *os.File, b []byte, done chan bool) (int, error) {
	writeArgs := &WriteArgs{
		f:   f,
		buf: b,
		n:   0,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: DoWrite,
		args: writeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return writeArgs.n, writeArgs.err
}

type CloseArgs struct {
	f   *os.File
	err error
}

func DoClose(v interface{}) {
	task := v.(*CloseArgs)
	task.err = task.f.Close()
}

func PushCloseTask(f *os.File, done chan bool) error {
	closeArgs := &CloseArgs{
		f:   f,
		err: nil,
	}
	task := &TaskElem{
		done: done,
		task: DoClose,
		args: closeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return closeArgs.err
}

type RenameArgs struct {
	old string
	nw  string
	err error
}

func DoRename(v interface{}) {
	task := v.(*RenameArgs)
	task.err = os.Rename(task.old, task.nw)
}

func PushRenameTask(oldname, newname string) error {
	writeArgs := &RenameArgs{
		old: oldname,
		nw:  newname,
		err: nil,
	}

	done := make(chan bool)
	task := &TaskElem{
		done: done,
		task: DoRename,
		args: writeArgs,
	}
	ioPool.PushTask(task)

	BockingUtilDoneChannel(done)
	return writeArgs.err
}
