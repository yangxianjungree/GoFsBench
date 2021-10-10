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
}

func NewFileWrapper(f *os.File, fd int) *FileWrapper {
	return &FileWrapper{
		real: f,
		fd:   fd,
	}
}

func OpenFileWrapper(name string, flag int, perm os.FileMode) (*FileWrapper, error) {
	if TYPE_FILE_WRAPPER_ORIGIN {
		fi, err := os.OpenFile(name, flag, perm)
		if err != nil {
			return nil, err
		}
		return NewFileWrapper(fi, -1), nil
	}

	fd, err := CPoolOpen(name, flag, perm)
	if err != nil {
		return nil, err
	}
	return NewFileWrapper(nil, fd), nil
}

func (f *FileWrapper) Close() error {
	if f.real != nil {
		return f.real.Close()
	}
	if f.fd != -1 {
		return CPoolClose(f.fd)
	}
	return nil
}

func (f *FileWrapper) Write(b []byte) (n int, err error) {
	if TYPE_FILE_WRAPPER_ORIGIN {
		return f.real.Write(b)
	}

	return CPoolWrite(int(f.fd), b)
}

func (f *FileWrapper) Read(b []byte) (n int, err error) {
	if TYPE_FILE_WRAPPER_ORIGIN {
		return f.real.Read(b)
	}

	return CPoolRead(int(f.real.Fd()), b)
}
