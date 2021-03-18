package main

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/dearplain/penet"
	"github.com/siddontang/go/log"
)

func setupPProf() {
	r := http.NewServeMux()
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)

	if err := http.ListenAndServe("127.0.0.1:9083", r); err != nil {
		log.Error(err)
	}
}

func main() {

	go setupPProf()

	penet.SetRate(30 * 1024 * 1024) // 设置最大速度为10Mbyte/s
	// penet.SetDropRate(0.1)          // 设置10%丢包率（接收和发送），用于测试

	listener, err := penet.Listen("", ":8070")
	if err != nil {
		log.Error(err)
		return
	}

	var sendSize = 10000
	var sendNum = 10000
	var sendTotal = sendSize * sendNum
	go func() {
		conn, err := penet.Dial("", "127.0.0.1:8070")
		if err != nil {
			log.Error(err)
			return
		}

		var start = time.Now()
		var total = 0
		data := make([]byte, sendSize)
		j := 0
		for i := 0; i < sendNum; i++ {
			for k := range data {
				data[k] = byte(j)
				j++
			}
			n, err := conn.Write(data)
			if n < 10000 || err != nil {
				log.Error("err:", err)
				break
			}
			total += len(data)
			if n != len(data) {
				panic("some err")
			}
		}
		conn.Close()
		log.Info("write close ", "total:", total, time.Since(start))
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("accept conn", conn)
		var preTotal = 0
		go func(conn net.Conn) {
			total := 0
			j := 0
			now := time.Now()
			start := now
			var b = make([]byte, 4086)
			for {
				n, err := conn.Read(b)
				if n <= 0 || err != nil {
					log.Info("read err:", err, n)
					break
				}
				for i := range b[:n] {
					if b[i] == byte(j) {
						j++
					} else {
						panic(fmt.Sprint("total", total, b[i], byte(j), j, i, n, b[:n]))
					}
				}
				total += n
				if total-preTotal > 5000000 {
					curNow := time.Now()
					elsap := curNow.UnixNano() - now.UnixNano()
					rate := float64(total-preTotal) / float64(elsap) * float64(time.Second) / float64(1024*1024)
					log.Infof("read: %d, %v, %.2fMB/s", total, time.Since(now), rate)
					now = curNow
					preTotal = total
				}
				if total >= sendTotal {
					log.Info("read end:", total, time.Since(start))
				}
			}
			log.Info("read total:", total)
		}(conn)
	}

}
