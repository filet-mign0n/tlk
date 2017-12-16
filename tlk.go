package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

const (
	connAttempts = 3
)

var (
	ch    = make(chan net.Conn)
	offCh = make(chan int)
	wg    = &sync.WaitGroup{}

	key     = flag.String("k", "mKHlhb797Yp9olUi", "aes key")
	host    = flag.String("h", "localhost", "host")
	port    = flag.String("p", "7777", "port")
	debug   = flag.Bool("d", false, "debug")
	verbose = flag.Bool("v", false, "verbose")
)

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
		conn, err := net.Dial("tcp", *host+":"+*port)
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
    // add close logic here
	listener, _ := net.Listen("tcp", ":"+*port)
	convo.log(fmt.Sprint("listening on", listener.Addr()))
	for {
		conn, _ := listener.Accept()
		convo.log(fmt.Sprint(conn.RemoteAddr().String(), " connected"))
		return conn
	}
}

func main() {
	wg.Add(1)
	go runTermui(ch, wg)

	c, ok := clt()
	if !ok {
		c = srv()
	}
	f := handleConn(c)
	f.Listen()
	convo.f = f

	wg.Wait()
	fmt.Println("wait done")
	os.Exit(0)
}
