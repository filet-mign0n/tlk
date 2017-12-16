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
	in     chan string
	out    chan string
	off    chan int
	status string
	name   string
	key    []byte
}

func (f *Friend) Read() {
	defer f.wg.Done()
	for {
		select {
		case line := <-f.in:
			decryptMsg, err := decrypt(f.key, line[:len(line)-1])
			if err != nil {
				convo.log(err.Error())
                continue
			}
			convo.chat(decryptMsg[:len(decryptMsg)-1])
		case _ = <-f.off:
			return
		}
	}
}

// no need to seperate!
// https://gist.github.com/rcrowley/5474430
func (f *Friend) ReadConn() {
    defer f.wg.Done()
    defer f.conn.Close()
	for {
		line, err := f.rw.ReadString('\n')
        switch {
        // doesn't seem to work like with EOF of net.Conn
        case err == io.EOF:
            convo.log("friend left, closing conn")
            close(f.off)
            return
        case err != nil:
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
		case <-f.off:
			return
		}
	}
}

func (f *Friend) Listen() {
	f.wg.Add(3)
    go f.ReadConn()
	go f.Read()
	go f.Write()
    /*
	go func() {
		defer wg.Done()
		for {
			select {
			case <-f.off:
				f.conn.Close()
			}
		}
	}()
    */
}

func NewFriend(conn net.Conn) *Friend {
	rw := bufio.NewReadWriter(bufio.NewReader(conn),
		bufio.NewWriter(conn),
	)
	f := &Friend{
		conn:   conn,
		rw:     rw,
		wg:     &sync.WaitGroup{},
		in:     make(chan string),
		out:    make(chan string),
		off:    offCh,
		status: "noauth",
		name:   "fox",
		key:    []byte(*key),
	}

	f.Listen()
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
