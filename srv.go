package main

import (
    "io"
//  "time"
	"bufio"
	"net"
	"sync"
)

type Friend struct {
	conn   net.Conn
	rw     *bufio.ReadWriter
	wg     *sync.WaitGroup
	out    chan string
	in     chan string
	off    chan int
	status string
	name   string
	key    []byte
}

func (f *Friend) Read() {
	defer f.wg.Done()
    for {
        select {
        case <-off:
            convo.log("f.Read got off")
            //time.Sleep(time.Second*4)
            return
        case line:= <-f.in:
            decryptMsg, err := decrypt(f.key, line[:len(line)-1])
            if err != nil {
                convo.log(err.Error())
            }
            convo.chat(decryptMsg[:len(decryptMsg)-1])
        }
    }
}

func (f *Friend) ReadConn() {
    f.wg.Done()
    for {
        line, err := f.rw.ReadString('\n')
        //f.conn.SetReadDeadline(time.Now().Add(1e9))
        // doesn't seem to work like with EOF of net.Conn
        if err == io.EOF {
            convo.log("friend left, closing conn")
            close(off)
            return
        }
        if err != nil {
            convo.log("friend ReadConn")
            convo.log(err.Error())
            return
        }
        if len(line) > 0 {
            f.in <- line
        }
    }
}

func (f *Friend) Write() {
	defer f.wg.Done()
    for {
        //f.conn.SetWriteDeadline(time.Now().Add(1e9))
        select {
        case data := <-f.out:
            data = data + "\n"
            data, err := encrypt(f.key, data)
            if err != nil {
                convo.log(err.Error())
                continue
            }
            if *debug {
                convo.log("crypto: " + data)
            }
            data = data + "\n"
            f.rw.WriteString(data)
            f.rw.Flush()
        case <-off:
            convo.log("Write got off")
            return
        }
    }
}

func (f *Friend) Listen() {
    defer wg.Done()
	f.wg.Add(3)
    go f.ReadConn()
	go f.Read()
	go f.Write()
    f.wg.Wait()
    convo.log("Listen() f.wg.Wait()")
    f.conn.Close()
    f.conn = nil
    wg.Add(1)
    go seek()
}

func NewFriend(conn net.Conn) *Friend {
	rw := bufio.NewReadWriter(bufio.NewReader(conn),
		bufio.NewWriter(conn),
	)
	f := &Friend{
		conn:   conn,
		rw:     rw,
		wg:     &sync.WaitGroup{},
		out:    make(chan string),
		in:     make(chan string),
		off:    off,
		status: "noauth",
		name:   "fox",
		key:    []byte(*key),
	}
	return f
}
