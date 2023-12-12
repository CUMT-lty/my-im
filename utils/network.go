package utils

import (
	"fmt"
	"strings"
)

const (
	networkSplit = "@"
)

// 检查网络连接的格式
func ParseNetwork(str string) (network, addr string, err error) { // TODO: 这种写法在有 err 的时候返回更方便
	if idx := strings.Index(str, networkSplit); idx == -1 {
		err = fmt.Errorf("addr: \"%s\" error, must be network@tcp:port or network@unixsocket", str)
		return
	} else {
		network = str[:idx]
		addr = str[idx+1:]
		fmt.Println("utils.network ==> ", "network=", network, "addr=", addr)
		return
	}
}
