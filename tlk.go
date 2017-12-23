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
	off    = make(chan int)
    connCh = make(chan net.Conn)
	wg     = &sync.WaitGroup{}

	key     = flag.String("k", "mKHlhb797Yp9olUi", "aes key")
	host    = flag.String("h", "localhost", "host")
	port    = flag.String("p", "7777", "port")
	debug   = flag.Bool("d", false, "debug")
	verbose = flag.Bool("v", false, "verbose")
)

func hdl() {
    defer wg.Done()
    for {
        select {
        case <-off:
            convo.log("hdl got off")
            //time.Sleep(time.Second*4)
            return
        case conn := <-connCh:
            if convo.f == nil {
                convo.log("handling conn")
                friend := NewFriend(conn)
                convo.f = friend
                wg.Add(1)
                friend.Listen()
            } else {
                convo.log("hdl got another conn")
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
                //time.Sleep(time.Second*4)
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
            connCh <- conn
            return true
        }
    return false
}

// pass wg as arg
func srv() bool {
    defer wg.Done()
	listener, _ := net.Listen("tcp", ":"+*port)
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
            connCh <- conn
            return
        }
    }()

    for {
        select {
        case <-off:
            convo.log("srv got off")
            //time.Sleep(time.Second*4)
            return false
        }
    }
}

func logOff() {
    defer wg.Done()
    for {
        select {
        case <-off:
            convo.log("anon got off")
            return
        }
    }
}

func seek() {
    defer wg.Done()
    wg.Add(1)
    go hdl()
	wg.Add(1)
    go logOff()
    gotConn := clt()
    if !gotConn {
        wg.Add(1)
        go srv()
    }
}

func main() {
	wg.Add(1)
	go tui(wg)
	wg.Add(1)
    go seek()
	wg.Wait()
	fmt.Println("wait done")
	os.Exit(0)
}
