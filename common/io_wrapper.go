package common

import (
	"os"
)

type FileWrapper struct {
	real *os.File
	fd   int
	done chan bool
}

func NewFileWrapper(f *os.File, fd int, done chan bool) *FileWrapper {
	return &FileWrapper{
		real: f,
		fd:   fd,
		done: done,
	}
}

func InitIoPool() {
	initCIoPool()
	initGoIoPool()
}

func OpenFileWrapper(name string, flag int, perm os.FileMode) (*FileWrapper, error) {
	done := make(chan bool)
	if GetGlobalConfigIns().UseGoIoPool() {
		fi, err := pushOpenTask(name, flag, perm, done)
		if err != nil {
			return nil, err
		}
		return NewFileWrapper(fi, -1, done), nil
	} else if GetGlobalConfigIns().UseCIoPool() {

		fd, err := cPoolOpen(name, flag, perm, done)
		if err != nil {
			return nil, err
		}
		return NewFileWrapper(nil, fd, done), nil
	}

	close(done)
	fi, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewFileWrapper(fi, -1, nil), nil
}

func (f *FileWrapper) Write(b []byte) (n int, err error) {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushWriteTask(f.real, b, f.done)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolWrite(int(f.fd), b, f.done)
	}
	return f.real.Write(b)
}

func (f *FileWrapper) Read(b []byte) (n int, err error) {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushReadTask(f.real, b, f.done)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolRead(int(f.real.Fd()), b, f.done)
	}
	return f.real.Read(b)
}

func (f *FileWrapper) Close() error {
	if GetGlobalConfigIns().UseGoIoPool() {
		if f.real != nil {
			return pushCloseTask(f.real, f.done)
		}
	} else if GetGlobalConfigIns().UseCIoPool() {
		if f.fd != -1 {
			return cPoolClose(f.fd, f.done)
		}
	}
	if f.done != nil {
		close(f.done)
	}
	if f.real != nil {
		return f.real.Close()
	}
	return nil
}

func RenameWrapper(oldname, newname string) error {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushRenameTask(oldname, newname)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolRename(oldname, newname)
	}
	return os.Rename(oldname, newname)
}
