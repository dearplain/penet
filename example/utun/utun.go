package main

import (
	"flag"
	"fmt"
	"io"
	"net"

	"github.com/dearplain/penet"
)

// ./utun -l :9086 -r xx.xx.xx.xx:xxx(远端fast-shadowsocks服务地址)
func main() {
	var local = flag.String("l", ":8070", "local")
	var remote = flag.String("r", "", "remote")
	flag.Parse()
	listener, err := net.Listen("tcp", *local)
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go func(localConn net.Conn) {
			remoteConn, err := penet.Dial("", *remote)
			if err != nil {
				return
			}
			go func() {
				io.Copy(localConn, remoteConn)
				remoteConn.Close()
				localConn.Close()
			}()
			io.Copy(remoteConn, localConn)
			remoteConn.Close()
			localConn.Close()
		}(conn)
	}
}
