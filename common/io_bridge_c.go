package common

// #include "io_bridge_c.h"
// #include <unistd.h>
// #include <pthread.h>
// #include <ctype.h>
// #include <errno.h>
// #include <stdlib.h>
// #include <stdio.h>
// #include <string.h>
// #include <sys/types.h>
// #include <sys/stat.h>
// #include <fcntl.h>
import "C"
import (
	"errors"
	"os"
	"strconv"
	"sync/atomic"
	"unsafe"
)

var (
	isCIoPoolInit int32 = 0
)

type CPoolArgs struct {
	fd    C.int
	n     C.int
	errno C.int
	cap   C.int
	buff  *C.char
	done  chan bool
}

func initCIoPool() {
	if atomic.LoadInt32(&isCIoPoolInit) == 0 && GetGlobalConfigIns().UseCIoPool() {
		atomic.StoreInt32(&isCIoPoolInit, 1)
		C.init_thread_pool(C.int(GetGlobalConfigIns().IoThreads), C.int(GetGlobalConfigIns().PriorIoThreads))
		LOG_STD("Use C io thread pool.......")
	}
}

func DestroyCPool() {
	C.destroy_thread_pool()
	atomic.StoreInt32(&isCIoPoolInit, 0)
}

//export go_done_callback
func go_done_callback(args *C.int) {
	// LOG_STD("go_done_callback one done...............")
	t := (*CPoolArgs)(unsafe.Pointer(args))
	t.done <- true
}

//export go_debug_log
func go_debug_log(msg *C.char) {
	LOG_STD("C function, ", C.GoString(msg))
}

func cPoolRead(fd int, buf []byte, done chan bool) (int, error) {
	args := &CPoolArgs{
		fd:    C.int(fd),
		n:     0,
		errno: 0,
		cap:   C.int(len(buf)),
		buff:  (*C.char)(unsafe.Pointer(&buf[0])),
		done:  done,
	}

	C.bridge_pool_read((*C.int)(unsafe.Pointer(args)))

	return waitCallBack(args)
}

func cPoolWrite(fd int, buf []byte, done chan bool) (int, error) {
	args := &CPoolArgs{
		fd:    C.int(fd),
		n:     0,
		errno: 0,
		cap:   C.int(len(buf)),
		buff:  (*C.char)(unsafe.Pointer(&buf[0])),
		done:  done,
	}

	C.bridge_pool_write((*C.int)(unsafe.Pointer(args)))

	return waitCallBack(args)
}

type cPoolOpenArgs struct {
	fd    C.int
	flag  C.int
	mode  C.int
	path  *C.char
	errno int
	done  chan bool
}

//export go_done_open_callback
func go_done_open_callback(args *C.int) {
	// LOG_STD("go_done_open_callback one done...............")
	t := (*cPoolOpenArgs)(unsafe.Pointer(args))
	t.done <- true
}

type cPoolCloseArgs struct {
	ret   C.int
	fd    C.int
	errno int
	done  chan bool
}

//export go_done_close_callback
func go_done_close_callback(args *C.int) {
	// LOG_STD("go_done_close_callback one done...............")
	t := (*cPoolCloseArgs)(unsafe.Pointer(args))
	t.done <- true
}

type cPoolRenameArgs struct {
	ret     C.int
	oldpath *C.char
	newpath *C.char
	errno   int
	done    chan bool
}

//export go_done_rename_callback
func go_done_rename_callback(args *C.int) {
	// LOG_STD("go_done_rename_callback one done...............")
	t := (*cPoolRenameArgs)(unsafe.Pointer(args))
	t.done <- true
}

func waitCallBack(args *CPoolArgs) (int, error) {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return int(args.n), err
}

func waitOpenCallBack(args *cPoolOpenArgs) (int, error) {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done open msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return int(args.fd), err
}

func waitCloseCallBack(args *cPoolCloseArgs) error {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done close msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return err
}

func waitRenameCallBack(args *cPoolRenameArgs) error {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done rename msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return err
}

func cPoolOpen(name string, flag int, perm os.FileMode, done chan bool) (int, error) {
	buf := []byte(name)
	args := &cPoolOpenArgs{
		fd:    0,
		flag:  C.int(flag),
		mode:  C.int(perm),
		path:  (*C.char)(unsafe.Pointer(&buf[0])),
		errno: 0,
		done:  done,
	}

	C.bridge_pool_open((*C.int)(unsafe.Pointer(args)))

	return waitOpenCallBack(args)
}

func cRead(fd int, buf []byte) int {
	l := len(buf)
	return int(C.bridge_read(C.int(fd), (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(l)))
}

func cWrite(fd int, buf []byte) int {
	l := len(buf)
	return int(C.bridge_write(C.int(fd), (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(l)))
}

func cPoolClose(fd int, done chan bool) error {
	args := &cPoolCloseArgs{
		fd:    C.int(fd),
		ret:   0,
		errno: 0,
		done:  done,
	}
	defer close(done)

	C.bridge_pool_close((*C.int)(unsafe.Pointer(args)))

	return waitCloseCallBack(args)
}

func cPoolRename(oldname, newname string) error {
	ol := []byte(oldname)
	nw := []byte(newname)
	args := &cPoolRenameArgs{
		ret:     0,
		oldpath: (*C.char)(unsafe.Pointer(&ol[0])),
		newpath: (*C.char)(unsafe.Pointer(&nw[0])),
		errno:   0,
		done:    make(chan bool),
	}

	C.bridge_pool_rename((*C.int)(unsafe.Pointer(args)))

	return waitRenameCallBack(args)
}
