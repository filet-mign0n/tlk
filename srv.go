package main

import (
	"bufio"
	"io"
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

func (f *Friend) Verify() {
	remAddr := f.conn.RemoteAddr().String()
	if addr != remAddr {
		convo.log("e", "Verify: addr != remAddr ", addr, " ", remAddr)
		close(disct)
	}
}

func (f *Friend) Read(wg *sync.WaitGroup) {
	defer f.wg.Done()
	for {
		select {
		case <-off:
			convo.log("d", "f.Read got off")
			return
		case line := <-f.in:
			decryptMsg, err := decrypt(f.key, line[:len(line)-1])
			if err != nil {
				convo.log("e", "f.Read", err.Error())
			}
			if len(decryptMsg) > 0 {
				decryptMsg = decryptMsg[:len(decryptMsg)-1]
			}
			convo.chat(decryptMsg)
		}
	}
}

func (f *Friend) ReadConn(wg *sync.WaitGroup) {
	f.wg.Done()
	for {
		line, err := f.rw.ReadString('\n')
		//f.conn.SetReadDeadline(time.Now().Add(1e9))
		if err == io.EOF {
			convo.log("e", f.conn.RemoteAddr(), " disconnected")
			close(disct)
			return
		}
		if err != nil {
			convo.log("e", "f.ReadConn error:", err.Error())
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
				convo.log("e", "f.Write error:", err.Error())
				continue
			}
			data = data + "\n"
			f.rw.WriteString(data)
			f.rw.Flush()
		case <-off:
			convo.log("d", "f.Write got off")
			return
		}
	}
}

func (f *Friend) Listen(wg *sync.WaitGroup) {
	defer wg.Done()
	f.Verify()
	f.wg.Add(3)
	go f.ReadConn(wg)
	go f.Read(wg)
	go f.Write(wg)
	f.wg.Wait()
	convo.log("d", "<friend.Listen> WaitGroup done")
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
