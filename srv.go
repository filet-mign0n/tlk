package main

import (
    "io"
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
    mode   string
    key    []byte
}

func (f *Friend) Read(wg *sync.WaitGroup) {
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

func (f *Friend) ReadConn(wg *sync.WaitGroup) {
    f.wg.Done()
    for {
        line, err := f.rw.ReadString('\n')
        //f.conn.SetReadDeadline(time.Now().Add(1e9))
        // doesn't seem to work like with EOF of net.Conn
        if err == io.EOF {
            convo.log("friend left, closing conn")
            close(disct)
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

func (f *Friend) Write(wg *sync.WaitGroup) {
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

func (f *Friend) Listen(wg *sync.WaitGroup) {
    defer wg.Done()
    f.wg.Add(3)
    go f.ReadConn(wg)
    go f.Read(wg)
    go f.Write(wg)
    f.wg.Wait()
    convo.log("friend.Listen -> f.wg.Wait() over")
}

func NewFriend(tc *tlkConn) *Friend {
    rw := bufio.NewReadWriter(bufio.NewReader(tc.conn),
        bufio.NewWriter(tc.conn),
    )
    f := &Friend{
        conn:   tc.conn,
        rw:     rw,
        wg:     &sync.WaitGroup{},
        out:    make(chan string),
        in:     make(chan string),
        off:    off,
        status: "noauth",
        name:   "fox",
        mode:   tc.mode,
        key:    []byte(*key),
    }
    return f
}
