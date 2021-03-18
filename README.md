# penet

## 协议头

发送包

| package type(1) | data type(1) | id(8) | body len(2) | seq(4) | timestamp(4) | wnd(4) | ... (real data) |
|---|---|---|---|---|---|---|---|

回复包

| package type(1) | data type(1) | id(8) | body len(2) | timestamp(4) | wnd(4) | acked_seq(4) | ack seq1(4) | ack seq2 (4) | ... ack seq n (4) |
|---|---|---|---|---|---|---|---|---|---|

## 特点

- 简单
    - 协议简洁，代码易懂，方便用户修改。
- 使用方便
    - 提供golang net包类似的tcp接口，写、读接口是同步的。
    - 阻塞式接口，不需要用户限制发送数据的速度。
- 完整。比较完整模拟了tcp接口的行为
    - 完整的生命周期管理，使用时候无需担心资源泄漏。网络异常等情况读写操作会返回错误。
    - write完可以直接close，网络正常情况下，另一端可以收到所有数据。
    - 对方关闭链接，也可以读完对方在关闭之前发送的所有数据，读完之后才会收到关闭消息。
    - 支持无限长时间链接，不需要发送心跳包。
    - 支持读超时。
- cpu、内存消耗低
    - 精心设计的代码，极大减少管道、协程、锁、timer、select的滥用。
    - 没有链接时所有线程都处于阻塞状态，没有cpu消耗。
    - 1W空转链接消耗2%的cpu(笔记本i5)。
- 延迟低
    - 没有握手包的延时。
- 流量小
    - 协议包头小，12byte，只占一般包大小的1/100。
    - 已经ack的包不需要重发。
    - ack包批量返回。
    - 有基本的丢包限速功能。
    - 对端不接受数据，不会频繁发送。

## 注意事项

- 默认的参数不一定适合所有人。penet.SetRate可以设置每秒发送的最大的字节。

## 例子

例子1：

加速&兼容各种手机端梯子：

手机端科学上网软件(shadowsocks，比如Outline/SuperWingy/小火箭，tcp) --> 国内转换服务器(tcp转penet) --> 国外fast-shadowsocks服务器(penet)

国内转换服务器代码：
```go
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
```

## 测试

test/ben/ben.go 可测试&验证顺序性、正确性和benmark速度。penet.SetRate调整速度，penet.SetDropRate调整丢包率。

## TODO

- [ ] 慢启动
- [ ] 速度探测

