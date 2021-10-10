package common

import (
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

var (
	DATA_DIR         = GetGlobalConfigIns().DataDir
	STROGES_FILE_DIR = "stroges"
	TMP_FILE_DIR     = "tmp"
)

func GetFilePath(checksum string) string {
	return filepath.Join(DATA_DIR, STROGES_FILE_DIR, checksum)
}

func GetFileTmpPath(checksum string) string {
	return filepath.Join(DATA_DIR, TMP_FILE_DIR, checksum+strconv.Itoa(rand.Intn(100000)))
}

func GetFileType(fileinfo os.FileInfo) int {
	if fileinfo.IsDir() {
		return FILE_TYPE_DIR
	} else {
		return FILE_TYPE_FILE
	}
}

func DiskToNet(conn net.Conn, fi *os.File, filesize int) error {
	buf := make([]byte, PAGE_SIZE)
	total_sent := 0
	for total_sent < filesize {
		next := filesize - total_sent
		if next >= PAGE_SIZE {
			next = PAGE_SIZE
		}

		n, err := fi.Read(buf[:next])
		if err != nil {
			ERR("DiskToNet read file content from disk failed, error: ", err)
			return err
		}

		_, err = conn.Write(buf[:n])
		if err != nil {
			ERR("DiskToNet write file content to net failed, error: ", err)
			return err
		}

		total_sent += n
	}
	return nil
}

func NetToDisk(conn net.Conn, fi *FileWrapper, filesize int) error {
	buf := make([]byte, PAGE_SIZE)
	total_sent := 0
	for total_sent < filesize {
		next := filesize - total_sent
		if next >= PAGE_SIZE {
			next = PAGE_SIZE
		}
		n, err := conn.Read(buf[:next])
		if err != nil {
			ERR("NetToDisk read file content failed, error: ", err)
			return err
		}

		_, err = fi.Write(buf[:n])
		if err != nil {
			ERR("NetToDisk write file content to disk failed, error: ", err)
			return err
		}

		total_sent += n
	}

	return nil
}
