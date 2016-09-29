package nstssh

import (
	"fmt"
	"os"
	"testing"
)

func TestSTDIO(t *testing.T) {
	IDENTITY = os.Getenv("HOME") + "/.ssh/id_rsa"
	output, err := Command("localhost", "ping").Output()
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(output))
}

func TestConnectFail(t *testing.T) {
	IDENTITY = os.Getenv("HOME") + "/.ssh/id_rsa"
	err := Command("192.168.1.248", "uname").Run()
	if err != nil {
		t.Error(err)
	}
}
