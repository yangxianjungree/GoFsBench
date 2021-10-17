package common

import (
	"io/fs"
	"os"
)

type FileWrapper struct {
	real *os.File
	fd   int
	name string
	done chan bool
}

func NewFileWrapper(f *os.File, fd int, name string, done chan bool) *FileWrapper {
	return &FileWrapper{
		real: f,
		fd:   fd,
		name: name,
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
		return NewFileWrapper(fi, -1, fi.Name(), done), nil
	} else if GetGlobalConfigIns().UseCIoPool() {
		fd, err := cPoolOpen(name, flag, perm, done)
		if err != nil {
			return nil, err
		}
		return NewFileWrapper(nil, fd, name, done), nil
	}

	close(done)
	fi, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewFileWrapper(fi, -1, fi.Name(), nil), nil
}

func (f *FileWrapper) Stat() (fs.FileInfo, error) {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushStatTask(f.real, f.done)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolStat(int(f.fd), f.name, f.done)
	}
	return f.real.Stat()
}

func (f *FileWrapper) Write(b []byte) (int, error) {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushWriteTask(f.real, b, f.done)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolWrite(int(f.fd), b, f.done)
	}
	return f.real.Write(b)
}

func (f *FileWrapper) Read(b []byte) (int, error) {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushReadTask(f.real, b, f.done)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolRead(int(f.fd), b, f.done)
	}
	return f.real.Read(b)
}

func (f *FileWrapper) Close() error {
	var err error = nil
	if GetGlobalConfigIns().UseGoIoPool() {
		if f.real != nil {
			err = pushCloseTask(f.real, f.done)
		}
	} else if GetGlobalConfigIns().UseCIoPool() {
		if f.fd != -1 {
			err = cPoolClose(f.fd, f.done)
		}
	}
	if f.done != nil {
		close(f.done)
	}
	if f.real != nil {
		err = f.real.Close()
	}
	return err
}

func RenameWrapper(oldname, newname string) error {
	if GetGlobalConfigIns().UseGoIoPool() {
		return pushRenameTask(oldname, newname)
	} else if GetGlobalConfigIns().UseCIoPool() {
		return cPoolRename(oldname, newname)
	}
	return os.Rename(oldname, newname)
}
