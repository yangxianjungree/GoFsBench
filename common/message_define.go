package common

const (
	MSG_NONE_ZERO = iota
	MSG_BENCH_MARK
	MSG_BENCH_MARK_RSP
	MSG_UPLOAD
	MSG_UPLOAD_RSP
	MSG_DOWNLOAD
	MSG_DOWNLOAD_RSP
	MSG_EXIST
	MSG_EXIST_RSP
	MSG_DELETE
	MSG_DELETE_RSP
	MSG_NONE_END
)

const (
	FILE_TYPE_INIT = iota
	FILE_TYPE_FILE
	FILE_TYPE_LINK
	FILE_TYPE_DIR
	FILE_TYPE_END
)

const (
	Kib                     = 1024
	Mib                     = Kib * Kib
	Gib                     = Mib * Kib
	NetVersion       string = "0001"
	NetMagic         int32  = 0x12345
	NetAppid         int32  = 0x12345
	OP_SUCCESS       string = "success"
	OP_FAILED        string = "failed"
	OP_START         string = "START"
	OP_END           string = "END"
	OP_ACK           string = "ACK"
	PANIC_STACK_SIZE        = 1 << 20
)
