package common

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"strconv"

	"google.golang.org/protobuf/proto"
)

func new_protocol_proto(op int32, checksum string, t int32, size int64, msg string) *NetFsProtocolProto {
	np := &NetFsProtocolProto{
		Version:   NetVersion,
		Magic:     NetMagic,
		Appid:     NetAppid,
		Operation: op,
		Checksum:  checksum,
		Type:      t,
		Size:      size,
		Message:   msg,
	}
	return np
}

func gen_request_protocol_proto(op int32, checksum, msg string) *NetFsProtocolProto {
	np := &NetFsProtocolProto{
		Version:   NetVersion,
		Magic:     NetMagic,
		Appid:     NetAppid,
		Operation: op,
		Checksum:  checksum,
		Type:      FILE_TYPE_INIT,
		Size:      0,
		Message:   msg,
	}
	return np
}

func check_protocol_proto(np *NetFsProtocolProto) bool {
	if np.Version != NetVersion || np.Magic != NetMagic || np.Appid != NetAppid {
		return false
	}

	return true
}

func encode_proto(np *NetFsProtocolProto) ([]byte, error) {
	pkg := new(bytes.Buffer)

	bytes_proto, err := proto.Marshal(np)
	if err != nil {
		ERR("Can't serislize NetFsProtocolProto, file: ", np.Checksum)
	}

	len := int32(len(bytes_proto))
	err = binary.Write(pkg, binary.LittleEndian, len)
	if err != nil {
		return nil, err
	}

	err = binary.Write(pkg, binary.LittleEndian, bytes_proto)
	if err != nil {
		return nil, err
	}

	return pkg.Bytes(), nil
}

func send_message_proto(np *NetFsProtocolProto, conn net.Conn) error {
	send_data, err := encode_proto(np)
	if err != nil {
		return err
	}

	total := len(send_data)
	var len_write int = 0
	// LOG_STD("Send data length: ", total)

	for {
		len, err := conn.Write(send_data[len_write:])
		if err != nil {
			DBG("Net write error: ", err)
			return err
		}

		len_write += len
		if len_write < total {
			continue
		}
		return nil
	}
}

func get_proto_msg_proto(conn net.Conn) (*NetFsProtocolProto, error) {
	var length int32
	err := binary.Read(conn, binary.LittleEndian, &length)
	if err != nil {
		ERR("Read length failed, error: ", err)
		return nil, err
	}

	rcv := 0
	data := make([]byte, length)
	for {
		le, err := conn.Read(data[rcv:])
		if err != nil {
			ERR("Read protocal data failed. error: ", err)
			return nil, err
		}

		rcv += le
		if rcv < int(length) {
			continue
		}

		// DBG("Recv data: ", data)

		var r NetFsProtocolProto
		err = proto.Unmarshal(data, &r)
		if err != nil {
			ERR("Can't deserislize ", data, err)
			return nil, err
		}

		// DBG("Recv data: ", r)

		if !check_protocol_proto(&r) {
			return nil, errors.New("version: " + r.Version + "magic: " + strconv.Itoa(int(r.Magic)) + " and appid: " + strconv.Itoa(int(r.Appid)) + " is not matched.")
		}
		return &r, nil
	}
}
