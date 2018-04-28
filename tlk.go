package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const connAttempts = 3

var (
	off    chan int      // closed when user exits or certain errs
	disct  chan int      // closed when friend disconnects
	connCh chan *tlkConn // pass around the conn to goroutines
	wg     *sync.WaitGroup
	addr   string

	key     = flag.String("k", "mKHlhb797Yp9olUi", "AES key")
	host    = flag.String("h", "localhost", "host")
	port    = flag.String("p", "7777", "port")
	debug   = flag.Bool("d", false, "debug")
	verbose = flag.Bool("v", false, "verbose")
)

type tlkConn struct {
	conn net.Conn
	mode string // clt or srv
}

func init() {
	flag.Parse()

	off = make(chan int)
	disct = make(chan int)
	connCh = make(chan *tlkConn)
	wg = &sync.WaitGroup{}
	addr = *host + ":" + *port
}

func hdl(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-off:
			convo.log("d", "hdl <-off")
			return
		case <-disct:
			convo.log("d", "hdl <-disct")
			if convo.f != nil {
				convo.f.conn.Close()
				convo.f.conn = nil
				convo.f = nil
				wg.Add(1)
				go seek(wg)
			}
			disct = make(chan int)
		case tlkc := <-connCh:
			if convo.f == nil {
				convo.log(
					"d",
					"handling conn from: ",
					tlkc.conn.RemoteAddr(),
				)
				friend := NewFriend(tlkc)
				convo.f = friend
				wg.Add(1)
				go friend.Listen(wg)
			} else {
				convo.log("e", "hdl already has open conn")
				convo.log(
					"e",
					"attempt from: ",
					tlkc.conn.RemoteAddr(),
					" mode: ",
					tlkc.mode,
				)
				tlkc.conn.Close()
			}
		}
	}
}

func clt() bool {
	ticker := time.NewTicker(time.Second)
	time.Sleep(time.Second)
	convo.log("l", "dialing "+addr)

	defer ticker.Stop()
	i := 0
OFF:
	for {
		select {
		case <-off:
			convo.log("d", "clt got off")
			return false
		default:
			goto POLL
		}
	}
POLL:
	for t := range ticker.C {
		i++
		tStr := t.Format("2006-01-02 15:04:05")
		convo.log("d", "dial attempt #", i, " ", tStr)
		if i >= connAttempts {
			convo.log(
                "d",
                "exceeded limit of ",
                connAttempts,
                " dial attempts",
            )
			return false
		}
		conn, err := net.DialTimeout(
			"tcp4",
			addr,
			5*time.Second,
		)
		if err != nil {
			convo.log("e", err.Error())
			goto OFF
		}
		convo.log("s", "dial successful")
		connCh <- &tlkConn{conn: conn, mode: "clt"}
		return true
	}
	return false
}

func srv(wg *sync.WaitGroup) {
	defer wg.Done()
	listener, e := net.Listen("tcp4", ":"+*port)
	if e != nil {
		convo.log("e", "srv error: ", e.Error())
		return
	}
	convo.log("l", "listening on ", listener.Addr())

	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			conn, err := listener.Accept()
			if nil != err {
				if opErr, ok := err.(*net.OpError); ok &&
					opErr.Timeout() {
					convo.log("e", "net.OpError.Timeout")
					continue
				}
				convo.log("e", "accept err:", err.Error())
				listener.Close()
				return
			}
			convo.log("s", conn.RemoteAddr(), " connected")
			connCh <- &tlkConn{conn: conn, mode: "srv"}
			listener.Close()
			return
		}
	}(wg)

	for {
		select {
		case <-off:
			listener.Close()
			convo.log("d", "srv got off")
			return
		}
	}
}

func seek(wg *sync.WaitGroup) {
	defer wg.Done()
	gotConn := clt()
	if !gotConn {
		wg.Add(1)
		go srv(wg)
	}
}

func main() {
	wg.Add(3)
	go tui(wg)
	go hdl(wg)
	go seek(wg)
	wg.Wait()
	if *debug {
		fmt.Println("<main> WaitGroup done")
	}
	os.Exit(0)
}
