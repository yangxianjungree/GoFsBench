package common

import (
	"errors"
	"net"
)

const (
	USER_PROTOCOL_TYPE = "proto"
)

type FsProtocol struct {
	proto_proto *NetFsProtocolProto
	json_proto  *NetFsProtocolJson
}

func NewProtocol(op int32, checksum string, t int32, size int64, msg string) *FsProtocol {
	if USER_PROTOCOL_TYPE == "json" {
		return &FsProtocol{
			json_proto:  new_protocol_json(op, checksum, t, size, msg),
			proto_proto: nil,
		}
	}
	return &FsProtocol{
		json_proto:  nil,
		proto_proto: new_protocol_proto(op, checksum, t, size, msg),
	}
}

func GenRequestProtocol(op int32, checksum, msg string) *FsProtocol {
	if USER_PROTOCOL_TYPE == "json" {
		return &FsProtocol{
			json_proto:  gen_request_protocol_json(op, checksum, msg),
			proto_proto: nil,
		}
	}

	return &FsProtocol{
		json_proto:  nil,
		proto_proto: gen_request_protocol_proto(op, checksum, msg),
	}
}

func (np *FsProtocol) CheckProtocol() bool {
	if np.json_proto == nil && np.proto_proto == nil {
		return false
	}

	if np.json_proto != nil {
		return np.json_proto.check_protocol_json()
	} else {
		return check_protocol_proto(np.proto_proto)
	}
}

func (np *FsProtocol) DebugProtocol() {
	DBG("json: ", np.json_proto, ", proto: ", np.proto_proto)
}

func (np *FsProtocol) SendMsg(conn net.Conn) error {
	if np.json_proto == nil && np.proto_proto == nil {
		return errors.New("fs protocol is empty")
	}

	if np.json_proto != nil {
		return np.json_proto.send_message_json(conn)
	} else {
		return send_message_proto(np.proto_proto, conn)
	}
}

func GetProtoMsg(conn net.Conn) (*FsProtocol, error) {
	np := &FsProtocol{
		json_proto:  nil,
		proto_proto: nil,
	}
	var err error

	if USER_PROTOCOL_TYPE == "json" {
		np.json_proto, err = get_proto_msg_json(conn)
	} else {
		np.proto_proto, err = get_proto_msg_proto(conn)
	}

	return np, err
}

func (x *FsProtocol) GetVersion() string {
	if x.json_proto != nil {
		return x.json_proto.Version
	} else if x.proto_proto != nil {
		return x.proto_proto.Version
	}
	return ""
}

func (x *FsProtocol) GetMagic() int32 {
	if x.json_proto != nil {
		return x.json_proto.Magic
	} else if x.proto_proto != nil {
		return x.proto_proto.Magic
	}
	return 0
}

func (x *FsProtocol) GetAppid() int32 {
	if x.json_proto != nil {
		return x.json_proto.Appid
	} else if x.proto_proto != nil {
		return x.proto_proto.Appid
	}
	return 0
}

func (x *FsProtocol) GetOperation() int32 {
	if x.json_proto != nil {
		return x.json_proto.Operation
	} else if x.proto_proto != nil {
		return x.proto_proto.GetOperation()
	}
	return 0
}

func (x *FsProtocol) GetChecksum() string {
	if x.json_proto != nil {
		return x.json_proto.Checksum
	} else if x.proto_proto != nil {
		return x.proto_proto.GetChecksum()
	}
	return ""
}

func (x *FsProtocol) GetType() int32 {
	if x.json_proto != nil {
		return x.json_proto.Type
	} else if x.proto_proto != nil {
		return x.proto_proto.GetType()
	}
	return 0
}

func (x *FsProtocol) GetSize() int64 {
	if x.json_proto != nil {
		return x.json_proto.Size
	} else if x.proto_proto != nil {
		return x.proto_proto.GetSize()
	}
	return 0
}

func (x *FsProtocol) GetMessage() string {
	if x.json_proto != nil {
		return x.json_proto.Message
	} else if x.proto_proto != nil {
		return x.proto_proto.GetMessage()
	}
	return ""
}

func (x *FsProtocol) SetType(ty int32) {
	if x.json_proto != nil {
		x.json_proto.Type = ty
	} else if x.proto_proto != nil {
		x.proto_proto.Type = ty
	}
}

func (x *FsProtocol) SetSize(size int64) {
	if x.json_proto != nil {
		x.json_proto.Size = size
	} else if x.proto_proto != nil {
		x.proto_proto.Size = size
	}
}

func (x *FsProtocol) SetMessage(msg string) {
	if x.json_proto != nil {
		x.json_proto.Message = msg
	} else if x.proto_proto != nil {
		x.proto_proto.Message = msg
	}
}
