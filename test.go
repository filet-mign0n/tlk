package main

import (
	//    "errors"
	//"bufio"
	"fmt"
	"net"
	"time"
	//tui "github.com/gizak/termui"
)

const (
	connAttempts = 3
	host         = "localhost"
	port         = ":7777"
)

var ch = make(chan net.Conn)

func handleConn(c net.Conn) *Friend {
	convo.log("handling conn")
	friend := NewFriend(c)
	return friend
}

func clt() (net.Conn, bool) {
	ticker := time.NewTicker(time.Second)
	i := 0
	for t := range ticker.C {
		i++
		tStr := t.Format("2006-01-02 15:04:05")
		convo.log(fmt.Sprint("dial attempt #", i, " ", tStr))
		if i >= connAttempts {
			convo.log(fmt.Sprint("exceeded ", connAttempts, " attempts"))
			ticker.Stop()
			return nil, false
		}
		conn, err := net.Dial("tcp", host+port)
		if err != nil {
			continue
		}
		convo.log("Dial successful")
		ticker.Stop()
		return conn, true
	}
	return nil, false
}

func srv() net.Conn {
	convo.log("entering server mode")
	listener, _ := net.Listen("tcp", port)
	for {
		conn, _ := listener.Accept()
		convo.log(fmt.Sprint("server got conn from: ", conn.RemoteAddr))
		return conn
	}
}

func main() {
	go runTermui(ch)

	c, ok := clt()
	if !ok {
		c = srv()
	}
	f := handleConn(c)
	convo.friend = f
	if ok {
		// time.Sleep(2*time.Second)
		// f.rw.WriteString("Why Hello\n")
		// f.rw.Flush()
	}

	for _ = range ch {
	}
	convo.log("closing shop")
}
