module server

replace common v0.0.0 => ../common

go 1.17

require common v0.0.0

require (
	github.com/golang/snappy v0.0.0-20180518054509-2e65f85255db // indirect
	github.com/syndtr/goleveldb v1.0.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
)
