package main

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"sync/atomic"

	. "common"
)

type BenchParms struct {
	Operation int
	Corotines int
	Loop      int
	File_size int
}

var (
	wg         *sync.WaitGroup = &sync.WaitGroup{}
	Bench_Loop int64           = 0
)

func client_bench_mark(conn net.Conn, params *BenchParms, checksum string) error {
	msg_send := GenRequestProtocol(MSG_BENCH_MARK, checksum, "")

	err := msg_send.SendMsg(conn)
	if err != nil {
		ERR("Bench mark send msg failed, error: ", err)
		return err
	}

	msg_rsp, err := GetProtoMsg(conn)
	if msg_rsp == nil {
		ERR("Bench mark recv msg failed.")
		return err
	}

	return nil
}

func client_upload_file(conn net.Conn, params *BenchParms, checksum string) error {
	if params.File_size <= 0 {
		return errors.New("file size is less than 0")
	}

	msg := GenRequestProtocol(MSG_UPLOAD, checksum, "")
	msg.SetSize(int64(params.File_size))
	err := msg.SendMsg(conn)
	if err != nil {
		ERR("Upload send msg failed, error: ", err)
		return err
	}

	rsp, err := GetProtoMsg(conn)
	if rsp == nil {
		ERR("Upload recv msg failed.")
		return err
	}

	if rsp.GetOperation() != MSG_UPLOAD_RSP {
		return errors.New("Upload resoponse not match, op: " + strconv.Itoa(int(rsp.GetOperation())))
	}

	if rsp.GetMessage() != OP_START {
		return errors.New("upload peer is not ready, remote msg: " + rsp.GetMessage())
	}

	buf := make([]byte, PAGE_SIZE)
	total_sent := 0
	for total_sent < params.File_size {
		next := params.File_size - total_sent
		if next >= PAGE_SIZE {
			next = PAGE_SIZE
		}
		n, err := conn.Write(buf[:next])
		if err != nil {
			ERR("Upload send file content failed, error: ", err)
			return err
		}
		total_sent += n
	}

	end, err := GetProtoMsg(conn)
	if err != nil {
		ERR("Upload recv file end protocol failed, error: ", err)
		return err
	}
	if end.GetMessage() != OP_END {
		return errors.New("Upload recv file end failed, remote msg: " + end.GetMessage())
	}

	ack := GenRequestProtocol(MSG_UPLOAD, checksum, OP_ACK)
	return ack.SendMsg(conn)
}

func client_download_file(conn net.Conn, params *BenchParms, checksum string) error {
	msg := GenRequestProtocol(MSG_DOWNLOAD, checksum, "")

	err := msg.SendMsg(conn)
	if err != nil {
		ERR("Download send msg failed, error: ", err)
		return err
	}

	rsp, err := GetProtoMsg(conn)
	if rsp == nil {
		ERR("Download recv msg failed.")
		return err
	}

	if rsp.GetOperation() != MSG_DOWNLOAD_RSP {
		return errors.New("Download resoponse not match, op: " + strconv.Itoa(int(rsp.GetOperation())))
	}

	if rsp.GetMessage() != OP_START {
		return errors.New("download peer is not ready, remote msg: " + rsp.GetMessage())
	}

	if rsp.GetSize() <= 0 {
		return errors.New("File size is not normal: " + strconv.Itoa(int(rsp.GetSize())))
	}

	buf := make([]byte, PAGE_SIZE)
	total_rcv := 0
	for total_rcv < int(rsp.GetSize()) {
		next := int(rsp.GetSize()) - total_rcv
		if next >= PAGE_SIZE {
			next = PAGE_SIZE
		}
		n, err := conn.Read(buf[:next])
		if err != nil {
			ERR("Upload send file content failed, error: ", err)
			return err
		}
		total_rcv += n
	}

	end, err := GetProtoMsg(conn)
	if err != nil {
		ERR("Download recv file end protocol failed, error: ", err)
		return err
	}
	if end.GetMessage() != OP_END {
		return errors.New("download recv file end failed")
	}

	ack := GenRequestProtocol(MSG_DOWNLOAD, checksum, OP_ACK)
	return ack.SendMsg(conn)
}

func client_delete_file(conn net.Conn, params *BenchParms, checksum string) error {
	msg := GenRequestProtocol(MSG_DELETE, checksum, "")

	err := msg.SendMsg(conn)
	if err != nil {
		ERR("Delete send msg failed, error: ", err)
		return err
	}

	rsp, err := GetProtoMsg(conn)
	if rsp == nil {
		ERR("Delete recv msg failed.")
		return err
	}

	if rsp.GetOperation() != MSG_DELETE_RSP {
		return errors.New("Delete resoponse not match, op: " + strconv.Itoa(int(rsp.GetOperation())))
	}

	DBG("Delete file from remote: ", rsp.GetMessage())

	return nil
}

func client_exist_bench(conn net.Conn, params *BenchParms, checksum string) error {

	send_exist := GenRequestProtocol(MSG_EXIST, checksum, "")

	err := send_exist.SendMsg(conn)
	if err != nil {
		ERR("Client exist send msg failed, error: ", err)
		return err
	}

	msg_rcv, err := GetProtoMsg(conn)
	if msg_rcv == nil {
		ERR("Client exist recv reponse failed.")
		return err
	}

	if msg_rcv.GetOperation() != MSG_EXIST_RSP {
		return errors.New("Exist resoponse not match, op: " + strconv.Itoa(int(msg_rcv.GetOperation())))
	}

	DBG("Msg about existence: ", msg_rcv.GetMessage())

	return nil
}

func handle_conn(conn net.Conn, params *BenchParms) {
	defer func() {
		if err := recover(); err != nil {
			ERR("recover error: ", err)
			DBG("Panic stack: ", string(PanicStackInfo()))
		}
	}()

	defer conn.Close()

	DBG("client: ", conn.LocalAddr().String())

	defer wg.Done()

	var err error
	for {
		index := atomic.AddInt64(&Bench_Loop, 1)
		if index > int64(params.Loop) {
			break
		}

		DBG("current loop index: ", index)

		checksum := FILE_PREFIX + strconv.Itoa(int(index))
		switch params.Operation {
		case MSG_DOWNLOAD:
			err = client_download_file(conn, params, checksum)
		case MSG_BENCH_MARK:
			err = client_bench_mark(conn, params, checksum)
		case MSG_UPLOAD:
			err = client_upload_file(conn, params, checksum)
		case MSG_EXIST:
			err = client_exist_bench(conn, params, checksum)
		case MSG_DELETE:
			err = client_delete_file(conn, params, checksum)
		default:
			err = errors.New("Operation: " + strconv.Itoa(params.Operation) + " is not support.")
		}
		if err != nil {
			ERR("Handle msg failed, error: ", err)
			break
		}
		atomic.AddInt64(&SUCCESS_TASKS, 1)
	}
}

func file_client(server_addr string, params *BenchParms) {
	for i := 0; i < params.Corotines; i++ {
		cli, err := net.Dial("tcp4", server_addr)
		if err != nil {
			DBG("Tcp connect error: ", err.Error())
			continue
		}

		wg.Add(1)
		go handle_conn(cli, params)
	}

	wg.Wait()
}
