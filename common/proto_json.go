package common

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"net"
	"strconv"
)

type NetFsProtocolJson struct {
	Version   string `json:"version"`
	Magic     int32  `json:"magic"`
	Appid     int32  `json:"appid"`
	Operation int32  `json:"operation"`
	Checksum  string `json:"checksum"`
	Type      int32  `json:"type"`
	Size      int64  `json:"size"`
	Message   string `json:"message"`
}

func new_protocol_json(op int32, checksum string, t int32, size int64, msg string) *NetFsProtocolJson {
	np := &NetFsProtocolJson{
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

func gen_request_protocol_json(op int32, checksum, msg string) *NetFsProtocolJson {
	np := &NetFsProtocolJson{
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

func (np *NetFsProtocolJson) check_protocol_json() bool {
	if np.Version != NetVersion || np.Magic != NetMagic || np.Appid != NetAppid {
		return false
	}

	return true
}

func (np *NetFsProtocolJson) encode_json() ([]byte, error) {
	pkg := new(bytes.Buffer)

	bytes_json, err := json.Marshal(np)
	if err != nil {
		ERR("Can't serislize NetFsProtocolJson, file: ", np.Checksum)
	}

	len := int32(len(bytes_json))
	err = binary.Write(pkg, binary.LittleEndian, len)
	if err != nil {
		return nil, err
	}

	err = binary.Write(pkg, binary.LittleEndian, bytes_json)
	if err != nil {
		return nil, err
	}

	return pkg.Bytes(), nil
}

func (np *NetFsProtocolJson) send_message_json(conn net.Conn) error {
	send_data, err := np.encode_json()
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

func get_proto_msg_json(conn net.Conn) (*NetFsProtocolJson, error) {
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

		var r NetFsProtocolJson
		err = json.Unmarshal(data, &r)
		if err != nil {
			ERR("Can't deserislize ", data, err)
			return nil, err
		}

		// DBG("Recv data: ", r)

		if !r.check_protocol_json() {

			return nil, errors.New("version: " + r.Version + "magic: " + strconv.Itoa(int(r.Magic)) + " and appid: " + strconv.Itoa(int(r.Appid)) + " is not matched.")
		}
		return &r, nil
	}
}
