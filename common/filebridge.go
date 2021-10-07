package common

// #include <unistd.h>
// #include "filebridge.h"
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
	"runtime"
	"strconv"
	"unsafe"
)

type CPoolArgs struct {
	fd    C.int
	n     C.int
	errno C.int
	cap   C.int
	buff  *C.char
	done  chan bool
}

func initCPool(threads int) {
	C.init_thread_pool(C.int(threads))
}

func destroyCPool() {
	C.destroy_thread_pool()
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

func CPoolRead(fd int, buf []byte) (int, error) {
	args := &CPoolArgs{
		fd:    C.int(fd),
		n:     0,
		errno: 0,
		cap:   C.int(len(buf)),
		buff:  (*C.char)(unsafe.Pointer(&buf[0])),
		done:  make(chan bool),
	}

	C.bridge_pool_read((*C.int)(unsafe.Pointer(args)))

	return waitCallBack(args)
}

func CPoolWrite(fd int, buf []byte) (int, error) {
	args := &CPoolArgs{
		fd:    C.int(fd),
		n:     0,
		errno: 0,
		cap:   C.int(len(buf)),
		buff:  (*C.char)(unsafe.Pointer(&buf[0])),
		done:  make(chan bool),
	}

	C.bridge_pool_write((*C.int)(unsafe.Pointer(args)))

	return waitCallBack(args)
}

type CPoolOpenArgs struct {
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
	t := (*CPoolOpenArgs)(unsafe.Pointer(args))
	t.done <- true
}

func CPoolOpen(name string, flag int, perm os.FileMode) (int, error) {
	buf := []byte(name)
	args := &CPoolOpenArgs{
		fd:    0,
		flag:  C.int(flag),
		mode:  C.int(perm),
		path:  (*C.char)(unsafe.Pointer(&buf[0])),
		errno: 0,
		done:  make(chan bool),
	}

	C.bridge_pool_open((*C.int)(unsafe.Pointer(args)))

	return waitOpenCallBack(args)
}

type CPoolCloseArgs struct {
	ret   C.int
	fd    C.int
	errno int
	done  chan bool
}

//export go_done_close_callback
func go_done_close_callback(args *C.int) {
	// LOG_STD("go_done_close_callback one done...............")
	t := (*CPoolCloseArgs)(unsafe.Pointer(args))
	t.done <- true
}

func CPoolClose(fd int) error {
	args := &CPoolCloseArgs{
		fd:    C.int(fd),
		ret:   0,
		errno: 0,
		done:  make(chan bool),
	}

	C.bridge_pool_close((*C.int)(unsafe.Pointer(args)))

	return waitCloseCallBack(args)
}

type CPoolRenameArgs struct {
	ret     C.int
	oldpath *C.char
	newpath *C.char
	errno   int
	done    chan bool
}

//export go_done_rename_callback
func go_done_rename_callback(args *C.int) {
	// LOG_STD("go_done_rename_callback one done...............")
	t := (*CPoolRenameArgs)(unsafe.Pointer(args))
	t.done <- true
}

func CPoolRename(oldname, newname string) error {
	if TYPE_FILE_WRAPPER_ORIGIN {
		return os.Rename(oldname, newname)
	}

	ol := []byte(oldname)
	nw := []byte(newname)
	args := &CPoolRenameArgs{
		ret:     0,
		oldpath: (*C.char)(unsafe.Pointer(&ol[0])),
		newpath: (*C.char)(unsafe.Pointer(&nw[0])),
		errno:   0,
		done:    make(chan bool),
	}

	C.bridge_pool_rename((*C.int)(unsafe.Pointer(args)))

	return waitRenameCallBack(args)
}

func BockingUtilDoneChannel(done chan bool) {
	for {
		select {
		case <-done:
			return
		default:
			runtime.Gosched()
			continue
		}
	}
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

func waitOpenCallBack(args *CPoolOpenArgs) (int, error) {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done open msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return int(args.fd), err
}

func waitCloseCallBack(args *CPoolCloseArgs) error {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done close msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return err
}

func waitRenameCallBack(args *CPoolRenameArgs) error {
	BockingUtilDoneChannel(args.done)

	// LOG_STD("Get done rename msg...........")
	var err error = nil
	if int(args.errno) != 0 {
		err = errors.New("errno is: " + strconv.Itoa(int(args.errno)))
	}
	return err
}

func CRead(fd int, buf []byte) int {
	l := len(buf)
	return int(C.bridge_read(C.int(fd), (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(l)))
}

func CWrite(fd int, buf []byte) int {
	l := len(buf)
	return int(C.bridge_write(C.int(fd), (*C.char)(unsafe.Pointer(&buf[0])), C.ulong(l)))
}
