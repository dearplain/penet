package penet

import (
	"fmt"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

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
		fmt.Println(err)
	}
}

func main() {

	go setupPProf()

	// SetRate(1200 * 5000)

	listener, err := Listen("", ":8070")
	if err != nil {
		fmt.Println(err)
		return
	}

	var sendSize = 10000
	var sendNum = 10000
	var sendTotal = sendSize * sendNum
	go func() {
		conn, err := Dial("", "127.0.0.1:8070")
		if err != nil {
			fmt.Println(err)
			return
		}

		var start = time.Now()
		var total = 0
		// data := []byte("ok-------------------")
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
		log.Info("write close", "total:", total, time.Since(start))
		// conn.Write([]byte("hello--------------"))
		// var b = make([]byte, 128)
		// n, _ := conn.Read(b)
		// fmt.Println(string(b[:n]))
	}()

	// go func() {
	// 	for i := 0; i < 50000; i++ {
	// 		Dial("", "127.0.0.1:8070")
	// 		// conn.Write([]byte(`hello`))
	// 	}
	// }()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		log.Info("accept conn", conn)
		var preTotal = 0
		go func(conn net.Conn) {
			total := 0
			j := 0
			now := time.Now()
			start := now
			var b = make([]byte, 2048)
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
				// fmt.Println("read:", n, string(b[:n]))
				if total-preTotal > 5000000 {
					preTotal = total
					curNow := time.Now()
					// elsap := ((curNow.UnixNano() - now.UnixNano())
					// rate := 10000 / elsap / int64(time.Millisecond)) * 1000
					log.Info("read:", total, time.Since(now))
					now = curNow
				}
				if total >= sendTotal {
					log.Info("read end:", total, time.Since(start))
				}
			}
			log.Info("read total:", total)
			// conn.Write([]byte("ok-------------------"))
			// var b = make([]byte, 128)
			// n, _ := conn.Read(b)
			// fmt.Println(string(b[:n]))
		}(conn)
	}

}
