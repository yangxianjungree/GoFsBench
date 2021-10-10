package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	. "common"
)

var (
	SUCCESS_TASKS int64 = 0
	server              = ":" + GetGlobalConfigIns().Port
	FILE_PREFIX         = "bench_mark_file_"
	USAGE               = "Usage: ./client [clean/bench/upload/download/exist/delete] corotine_num loop_num [file_size]"
)

func clearance(num, loop int) {
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				index := atomic.AddInt64(&Bench_Loop, 1)
				if index > int64(loop) {
					break
				}

				f_path := GetFilePath(FILE_PREFIX + strconv.FormatInt(index, 10))
				err := os.Remove(f_path)
				if err == nil {
					atomic.AddInt64(&SUCCESS_TASKS, 1)
				} else {
					LOG_STD("Romve file ", f_path, " error: ", err)
				}
			}
		}()
	}

	wg.Wait()
	LOG_STD("Total task: ", loop, ", success: ", SUCCESS_TASKS)
}

func get_tmp_path(checksum string) string {
	return filepath.Join(DATA_DIR, TMP_FILE_DIR, checksum+".tmp")
}

func get_tmp_path1(checksum string, index string) string {
	return filepath.Join(DATA_DIR, TMP_FILE_DIR, index, checksum+".tmp")
}

func get_file_path1(checksum string, index string) string {
	return filepath.Join(DATA_DIR, TMP_FILE_DIR, index, checksum)
}

func rename(num, loop int) {
	for i := 0; i < num; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				index := atomic.AddInt64(&Bench_Loop, 1)
				if index > int64(loop) {
					break
				}

				file_name := FILE_PREFIX + strconv.FormatInt(index, 10)
				f_path := GetFilePath(file_name)
				// tmp_path_origin := get_file_path1(file_name, strconv.FormatInt(index%10, 10))
				tmp_path := get_tmp_path1(file_name, strconv.FormatInt(index%10, 10))
				err := os.Rename(tmp_path, f_path)
				// err := os.Rename(tmp_path, tmp_path_origin)
				// err := os.Rename(tmp_path, f_path)
				if err == nil {
					atomic.AddInt64(&SUCCESS_TASKS, 1)
				} else {
					LOG_STD("Rename file ", f_path, " error: ", err)
				}
			}
		}()
	}

	wg.Wait()
	LOG_STD("Total task: ", loop, ", success: ", SUCCESS_TASKS)
}

func main() {
	args := os.Args
	if len(args) < 4 {
		LOG_STD(USAGE)
		return
	}

	num, err := strconv.Atoi(args[2])
	if err != nil {
		LOG_STD("Cant convert corotines num, error: ", err)
		LOG_STD(USAGE)
		return
	}

	loop, err := strconv.Atoi(args[3])
	if err != nil {
		LOG_STD("Cant convert loop num, error: ", err)
		LOG_STD(USAGE)
		return
	}

	if loop < 1 || num < 1 {
		loop = 1
		num = 1
	}

	if !Init() {
		ERR("Init failed.")
		return
	}

	bench := NewBenchBoard(uint64(loop))
	defer bench.ShowBenchBoard()

	op := -1
	file_size := 0
	if args[1] == "clean" {
		clearance(num, loop)
		bench.SetSuccessTasks(uint64(atomic.LoadInt64(&SUCCESS_TASKS)))
		return
	} else if args[1] == "rename" {
		rename(num, loop)
		bench.SetSuccessTasks(uint64(atomic.LoadInt64(&SUCCESS_TASKS)))
		return
	} else {
		if len(args) < 5 {
			file_size = 5 * Kib
		} else {
			file_size, err = strconv.Atoi(args[4])
			if err != nil {
				LOG_STD("Cant convert file size, error: ", err)
				LOG_STD(USAGE)
				return
			}
		}
	}

	if args[1] == "bench" {
		op = MSG_BENCH_MARK
	} else if args[1] == "upload" {
		op = MSG_UPLOAD
	} else if args[1] == "download" {
		op = MSG_DOWNLOAD
	} else if args[1] == "exist" {
		op = MSG_EXIST
	} else if args[1] == "delete" {
		op = MSG_DELETE
	}

	params := &BenchParms{
		Operation: op,
		Corotines: num,
		Loop:      loop,
		File_size: file_size,
	}

	data, err := json.Marshal(params)
	if err != nil {
		LOG_STD("Parmas: ", params)
	} else {
		LOG_STD("Parmas: ", string(data))
	}

	file_client(server, params)
	bench.SetSuccessTasks(uint64(atomic.LoadInt64(&SUCCESS_TASKS)))
}
