package main

import (
	"errors"
	"net"
	"os"
	"sync/atomic"
	"time"

	. "common"
)

func bench_mark(np *FsProtocol, conn net.Conn) error {
	atomic.AddInt64(&Bench_Loop, 1)

	msg_rsp := GenRequestProtocol(MSG_BENCH_MARK_RSP, np.GetChecksum(), "")

	return msg_rsp.SendMsg(conn)
}

func download_file(np *FsProtocol, conn net.Conn) error {
	rsp := GenRequestProtocol(MSG_DOWNLOAD_RSP, np.GetChecksum(), OP_START)

	file_path := GetFilePath(np.GetChecksum())
	fi, err := os.OpenFile(file_path, os.O_RDONLY, 0)
	if err != nil {
		ERR("Download open file: ", file_path, " failed, error: ", err)
		rsp.SetMessage(err.Error())
		return rsp.SendMsg(conn)
	}
	defer fi.Close()

	st, err := fi.Stat()
	if err != nil {
		ERR("Download get file: ", file_path, " stat failed, error: ", err)
		rsp.SetMessage(err.Error())
		return rsp.SendMsg(conn)
	}

	rsp.SetSize(st.Size())
	rsp.SetType(int32(GetFileType(st)))

	err = rsp.SendMsg(conn)
	if err != nil {
		ERR("Download send begin msg failed, error: ", err)
		return err
	}

	err = DiskToNet(conn, fi, int(st.Size()))
	if err != nil {
		ERR("Download file ", np.GetChecksum(), " failed, error: ", err)
		return err
	}

	end := GenRequestProtocol(MSG_DOWNLOAD, np.GetChecksum(), OP_END)

	err = end.SendMsg(conn)
	if err != nil {
		ERR("Download send end msg failed, error: ", err)
		return err
	}

	ack, err := GetProtoMsg(conn)
	if err != nil {
		ERR("Download recv file ack protocol failed, error: ", err)
		return err
	}
	if ack.GetMessage() != OP_ACK {
		return errors.New("Download didnt recv ack, remote msg: " + ack.GetMessage())
	}

	return nil
}

func upload_file(np *FsProtocol, conn net.Conn) error {
	rsp := GenRequestProtocol(MSG_UPLOAD_RSP, np.GetChecksum(), OP_START)

	if np.GetChecksum() == "" || np.GetSize() <= 0 {
		rsp.SetMessage("Checksum is empty or file size unnormal.")
		return rsp.SendMsg(conn)
	}

	tmp_path := GetFileTmpPath(np.GetChecksum())
	file_path := GetFilePath(np.GetChecksum())
	fi, err := OpenFileWrapper(tmp_path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		rsp.SetMessage(err.Error())
		return rsp.SendMsg(conn)
	}
	defer fi.Close()

	err = rsp.SendMsg(conn)
	if err != nil {
		ERR("Upload send begin msg failed, error: ", err)
		return err
	}

	err = NetToDisk(conn, fi, int(np.GetSize()))
	if err != nil {
		ERR("Upload file ", np.GetChecksum(), " failed, error: ", err)
		return err
	}

	end := GenRequestProtocol(MSG_UPLOAD, np.GetChecksum(), OP_END)

	rename_start := time.Now()
	err = RenameWrapper(tmp_path, file_path)
	if err != nil {
		ERR("Upload rename ", tmp_path, " to ", file_path, " failed, error: ", err)
		end.SetMessage(err.Error())
		return end.SendMsg(conn)
	}
	DBG("Rname costing time: ", time.Since(rename_start).Milliseconds())

	err = end.SendMsg(conn)
	if err != nil {
		ERR("Upload send end msg failed, error: ", err)
		return err
	}

	ack, err := GetProtoMsg(conn)
	if err != nil {
		ERR("Upload recv file ack protocol failed, error: ", err)
		return err
	}
	if ack.GetMessage() != OP_ACK {
		return errors.New("Upload didnt recv ack, remote msg: " + ack.GetMessage())
	}

	return nil
}

func delete_file(np *FsProtocol, conn net.Conn) error {
	rsp := GenRequestProtocol(MSG_DELETE_RSP, np.GetChecksum(), "")

	if np.GetChecksum() == "" {
		rsp.SetMessage("Checksum is empty")
		return rsp.SendMsg(conn)
	}

	dst := GetFilePath(np.GetChecksum())
	err := os.Remove(dst)

	if err != nil {
		rsp.SetMessage(err.Error())
		return rsp.SendMsg(conn)
	}

	rsp.SetMessage(OP_SUCCESS)

	return rsp.SendMsg(conn)
}

func exist_file(np *FsProtocol, conn net.Conn) error {
	rsp := GenRequestProtocol(MSG_EXIST_RSP, np.GetChecksum(), "")

	if np.GetChecksum() == "" {
		rsp.SetMessage("Checksum is empty")
		return rsp.SendMsg(conn)
	}

	dst := GetFilePath(np.GetChecksum())
	st, err := os.Stat(dst)

	if err != nil {
		rsp.SetMessage(err.Error())
		return rsp.SendMsg(conn)
	}

	rsp.SetSize(st.Size())
	rsp.SetType(int32(GetFileType(st)))
	rsp.SetMessage(OP_SUCCESS)

	return rsp.SendMsg(conn)
}

var handler_map map[int32]func(np *FsProtocol, conn net.Conn) error

func register_func() {
	handler_map = make(map[int32]func(np *FsProtocol, conn net.Conn) error)
	handler_map[MSG_BENCH_MARK] = bench_mark
	handler_map[MSG_DOWNLOAD] = download_file
	handler_map[MSG_UPLOAD] = upload_file
	handler_map[MSG_EXIST] = exist_file
	handler_map[MSG_DELETE] = delete_file
}

func get_handler(msg int32) func(np *FsProtocol, conn net.Conn) error {
	hand, ok := handler_map[msg]
	if !ok {
		return nil
	}
	return hand
}

func handle_conn(conn net.Conn) {
	defer func() {
		if err := recover(); err != nil {
			ERR("recover error: ", err)
			DBG("Panic stack: ", string(PanicStackInfo()))
		}
	}()

	defer conn.Close()

	DBG("Accept tcp client: ", conn.RemoteAddr().String())

	atomic.AddInt64(&Connections, 1)

	for {
		msg, err := GetProtoMsg(conn)
		if msg == nil || err != nil {
			ERR("Decode Net protocol failed, error: ", err)
			return
		}

		handle := get_handler(msg.GetOperation())
		if handle == nil {
			ERR("Operation ", msg.GetOperation(), " is not supported.")
			return
		}

		err = handle(msg, conn)
		if err != nil {
			ERR("Handle operation: ", msg.GetOperation(), ", checksum: ", msg.GetChecksum(), " failed, error: ", err)
			break
		}
	}

}

func file_server(ser string) {
	resSer, err := net.ResolveTCPAddr("tcp", ser)
	if err != nil {
		ERR("Resolve tcp addr error: ", err.Error())
		return
	}

	register_func()

	listen, err := net.ListenTCP("tcp4", resSer)
	if err != nil {
		ERR("Tcp listen error: ", err.Error())
		return
	}
	defer listen.Close()

	DBG("start server successful......")

	go show_statistic()

	for {
		connection, err := listen.Accept()
		if err != nil {
			ERR("Accept error: ", err.Error())
		} else {
			go handle_conn(connection)
		}
	}
}
