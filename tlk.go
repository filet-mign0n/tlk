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
	off     chan int // closed when user exits or certain errs
    disct   chan int // closed when friend disconnects
    connCh  chan *tlkConn // pass around the conn to goroutines
	wg      *sync.WaitGroup

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
    off    = make(chan int)
    disct  = make(chan int)
    connCh = make(chan *tlkConn)
	wg     = &sync.WaitGroup{}
}

func hdl(wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-off:
            convo.log("hdl got off")
            return
        case <-disct:
            convo.log("hdl got disct")
            if (convo.f != nil) {
                mode := convo.f.mode
                convo.f.conn.Close()
                convo.f.conn = nil
                convo.f = nil
                if (mode == "clt") {
                    wg.Add(1)
                    go seek(wg)
                }
            }
            disct = make(chan int)
        case tlkc := <-connCh:
            if convo.f == nil {
                convo.log("handling conn")
                friend := NewFriend(tlkc)
                convo.f = friend
                wg.Add(1)
                go friend.Listen(wg)
            } else {
                convo.log("hdl already has open conn")
            }
        }
    }
}

func clt() (bool) {
	ticker := time.NewTicker(time.Second)
    defer ticker.Stop()
	i := 0
    OFF:
        for {
            select {
            case <-off:
                convo.log("clt got off")
                return false
            default:
                goto POLL
            }
        }
    POLL:
        for t := range ticker.C {
            i++
            tStr := t.Format("2006-01-02 15:04:05")
            convo.log(fmt.Sprint("dial attempt #", i, " ", tStr))
            if i >= connAttempts {
                convo.log(fmt.Sprint("exceeded ", connAttempts, " attempts"))
                return false
            }
            conn, err := net.Dial("tcp", *host+":"+*port)
            if err != nil {
                goto OFF
            }
            convo.log("Dial successful")
            connCh <- &tlkConn{conn: conn, mode: "clt"}
            return true
        }
    return false
}

// pass wg as arg
func srv(wg *sync.WaitGroup) {
    defer wg.Done()
	listener, e := net.Listen("tcp", ":"+*port)
    if e != nil {
        convo.log("srv error")
        convo.log(e.Error())
    }
	convo.log(fmt.Sprint("listening on ", listener.Addr()))
    defer listener.Close()

    wg.Add(1)
    go func() {
      defer wg.Done()
        for {
            conn, err := listener.Accept()
            convo.log("go routine inside srv() started")
            convo.log(time.Now().String())
            if nil != err {
                if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
                    convo.log("OpError.Timeout")
                    continue
                }
                convo.log("Accept err")
                convo.log(err.Error())
                return
            }
            convo.log(fmt.Sprint(conn.RemoteAddr().String(), " connected"))
            connCh <- &tlkConn{conn: conn, mode: "srv"}
            //return
        }
    }()

    for {
        select {
        case <-off:
            convo.log("srv got off")
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
	fmt.Println("wait done")
	os.Exit(0)
}
