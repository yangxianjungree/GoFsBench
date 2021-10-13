package common

import (
	"os"
)

var (
	TYPE_FILE_WRAPPER_ORIGIN = !GetGlobalConfigIns().UserCPoolIoSched
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

func OpenFileWrapper(name string, flag int, perm os.FileMode) (*FileWrapper, error) {
	// if TYPE_FILE_WRAPPER_ORIGIN {
	// 	fi, err := os.OpenFile(name, flag, perm)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return NewFileWrapper(fi, -1, nil), nil
	// }
	done := make(chan bool)

	if TYPE_FILE_WRAPPER_ORIGIN {
		fi, err := PushOpenTask(name, flag, perm, done)
		if err != nil {
			return nil, err
		}
		return NewFileWrapper(fi, -1, done), nil
	}

	fd, err := CPoolOpen(name, flag, perm, done)
	if err != nil {
		return nil, err
	}
	return NewFileWrapper(nil, fd, done), nil
}

func (f *FileWrapper) Close() error {
	if f.real != nil {
		// return f.real.Close()
		return PushCloseTask(f.real, f.done)
	}
	if f.fd != -1 {
		return CPoolClose(f.fd, f.done)
	}
	if f.done != nil {
		close(f.done)
	}
	return nil
}

func (f *FileWrapper) Write(b []byte) (n int, err error) {
	if TYPE_FILE_WRAPPER_ORIGIN {
		// return f.real.Write(b)
		return PushWriteTask(f.real, b, f.done)
	}

	return CPoolWrite(int(f.fd), b, f.done)
}

func (f *FileWrapper) Read(b []byte) (n int, err error) {
	if TYPE_FILE_WRAPPER_ORIGIN {
		// return f.real.Read(b)
		return PushReadTask(f.real, b, f.done)
	}

	return CPoolRead(int(f.real.Fd()), b, f.done)
}
