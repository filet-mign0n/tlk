package main

import (
    "errors"
    "bufio"
    "net"
    "fmt"
    "time"

    //tui "github.com/gizak/termui"
)

const (
    connAttempts = 2
    listenTimeout   = 4 * time.Second
)

var connCh = make(chan net.Conn)

func pollAndListen(ch chan<- net.Conn, attempts int) {
    // first try to poll for an open socket on the other end
    timer := time.NewTimer(listenTimeout)
    go func() {
        ticker := time.NewTicker(time.Second)
        i := 0
        for {
            select {
            case t := <- ticker.C:
                i++
                tStr := t.Format("2006-01-02 15:04:05")
                convo.log(fmt.Sprint("dial attempt #", i, " ", tStr))
                if i >= attempts {
                    convo.log(fmt.Sprint("exceeded ",
                        attempts, " attempts",
                    ))
                    ticker.Stop()
                    break
                }
                conn, err := net.Dial("tcp", "localhost:7777")
                if err != nil {
                    continue
                }
                ch <- conn
                ticker.Stop()
            case <- timer.C:
                ticket.Stop()
                return
            }
        }
    }()
/*
    // else listen until timeout
    convo.log(fmt.Sprint("now listening"))
    listener, _ := net.Listen("tcp", ":7777")
    for {
        conn, _ := listener.Accept()
        convo.log(fmt.Sprint("some entity dialed in ",
            conn.RemoteAddr(),
        ))
        ch <- conn
        timer.Stop()
        return
        <-timer.C:
    }
    convo.log(fmt.Sprint("listen timeout"))
    close(ch)
*/
}

func handleConn(c net.Conn, mode string) error {
    switch mode {
    case "clt":
        convo.log("clt")

        rw := bufio.NewReadWriter(
            bufio.NewReader(c),
            bufio.NewWriter(c),
        )
        rw.WriteString("test")
        convo.log(fmt.Sprint("got clt conn: ", c.RemoteAddr()))
        convo.log("entering client mode")

        c.Close()
        return nil
    case "srv":
        convo.log("srv")
        return nil
    default:
        return errors.New("unknown conn mode")
    }
}

func main() {
    go runTermui()
    go pollAndListen(connCh, connAttempts)

    for {
        c, ok := <- connCh
        if !ok {
            convo.log("polled and listened for too long, bye.")
            return
        }
        handleConn(c, "clt")
    }
}

