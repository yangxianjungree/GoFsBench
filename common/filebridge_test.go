package common

import (
	"testing"
)

func TestBridgeWrite(t *testing.T) {
	LOG_STD("Unit test........TestBridgeWrite.........")

	str := "mmmmmmmmmm\n"
	CWrite(1, []byte(str))
}

func TestBridgeRead(t *testing.T) {
	LOG_STD("Unit test........TestBridgeRead.........")

	str := make([]byte, 10)
	CRead(1, str)
}

func TestBridgePoolWrite(t *testing.T) {
	LOG_STD("Unit test........TestBridgePoolWrite.........")
	InitCPool()

	str := "mmmmmmmmmTestBridgePoolWritem\n"
	done := make(chan bool)
	CPoolWrite(1, []byte(str), done)
	LOG_STD("TestBridgePoolWrite done..........")
}

func TestBridgePoolRead(t *testing.T) {
	LOG_STD("Unit test........TestBridgePoolRead.........")
	InitCPool()

	str := make([]byte, 10)
	done := make(chan bool)
	CPoolRead(1, str, done)
	LOG_STD("Have read data: ", str, " from stdio 1.......")
}
