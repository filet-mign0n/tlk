package main

import (
    "io"
//    "time"
	"bufio"
	"net"
	"sync"
)

type Friend struct {
	conn   net.Conn
	rw     *bufio.ReadWriter
	wg     *sync.WaitGroup
	out    chan string
	off    chan int
	status string
	name   string
	key    []byte
}

func (f *Friend) Read() {
	defer f.wg.Done()
    OFF:
        for {
            select {
            case <-off:
                convo.log("f.Read got off")
                //time.Sleep(time.Second*4)
                return
            default:
                goto rwRead
            }
        }
    rwRead:
        for {
            line, err := f.rw.ReadString('\n')
            //f.conn.SetReadDeadline(time.Now().Add(1e9))
            switch {
            // doesn't seem to work like with EOF of net.Conn
            case err == io.EOF:
                convo.log("friend left, closing conn")
                close(off)
                return
            case err != nil:
                convo.log("friend ReadConn")
                convo.log(err.Error())
                return
            case len(line) > 0:
                decryptMsg, err := decrypt(f.key, line[:len(line)-1])
                if err != nil {
                    convo.log(err.Error())
                }
                convo.chat(decryptMsg[:len(decryptMsg)-1])
                goto OFF
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
	f.wg.Add(2)
	go f.Read()
	go f.Write()
    f.wg.Wait()
    convo.log("f.wg.Wait()")
    f.conn.Close()
    f.conn = nil
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
		off:    off,
		status: "noauth",
		name:   "fox",
		key:    []byte(*key),
	}
	return f
}

/*
type HandleFunc func(*bufio.ReadWriter)

type ChatRoom struct {
    f   *Friend
    joins    chan net.Conn
    in chan string
    out chan string
    handler  map[string]HandleFunc
}

func (chatRoom *ChatRoom) Broadcast(data string) {
    for _, f := range chatRoom.f {
        f.out <- data
    }
}

func (chatRoom *ChatRoom) Join(connection net.Conn) {
    f := NewFriend(connection)
    chatRoom.f = f
    fmt.Println("f connected")
    go func() {
        for {
            chatRoom.in <- <-f.in
        }
    }()
}

func (chatRoom *ChatRoom) Listen() {
    go func() {
        for {
            select {
            case data := <-chatRoom.in:
                chatRoom.Broadcast(data)
            case conn := <-chatRoom.joins:
                chatRoom.Join(conn)
            }
        }
    }()
}

func NewChatRoom() *ChatRoom {
    chatRoom := &ChatRoom{
        f:   *Friend,
        joins:    make(chan net.Conn),
        in: make(chan string),
        out: make(chan string),
    }

    chatRoom.Listen()

    return chatRoom
}

func NewClient() *Client {
    conn, err := net.Dial("tcp", addr)
    if err != nil {
        return nil, fmt.Println(err, "Dialing "+addr+" failed")
    }
    return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil

}

func erv() {
    //chatRoom := NewChatRoom()

    listener, _ := net.Listen("tcp", ":7777")

    //go func() {
        for {
            conn, _ := listener.Accept()
            fmt.Println("got conn")
            conn.Close()
            //chatRoom.joins <- conn
        }
    //}()
}
*/
