package main

import (
	"bufio"
	"net"
	//    "fmt"
)

/*
type Line struct {
    impl interface{}
}

func New (l LineImpl) *Line {
    return &Line{l}
}

func NewLine(connection net.Conn) *Line {
    writer := bufio.NewWriter(connection)
    reader := bufio.NewReader(connection)

    impl := interface{
        incoming: make(chan string),
        outgoing: make(chan string),
        reader:   reader,
        writer:   writer,
    }

    friend.Listen()

    return friend
}

func (l Line) Read() {
    for {
        msg, _ := l.impl.reader.ReadString("\n")
        if len(msg) > 0 { fmt.Println(msg) }
        friend.incoming <- line
    }
}

func (l Line) Write() {
    for data := range l.impl.outgoing {
        l.impl.writer.WriteString(data)
        l.impl.writer.Flush()
    }
}

func (l Line) Listen() {
    go l.impl.Read()
    go l.impl.Write()
}
*/

type Friend struct {
	rw       *bufio.ReadWriter
	incoming chan string
	outgoing chan string
	status   string
	name     string
}

func (friend *Friend) Read() {
	for {
		line, _ := friend.rw.ReadString('\n')
		if len(line) > 0 {
			convo.chat(line[:len(line)-1])
		}
	}
}

func (friend *Friend) Write() {
	for data := range friend.outgoing {
		friend.rw.WriteString(data + "\n")
		friend.rw.Flush()
	}
}

func (friend *Friend) Listen() {
	go friend.Read()
	go friend.Write()
}

func NewFriend(connection net.Conn) *Friend {
	rw := bufio.NewReadWriter(bufio.NewReader(connection),
		bufio.NewWriter(connection),
	)
	friend := &Friend{
		rw:       rw,
		incoming: make(chan string),
		outgoing: make(chan string),
		status:   "begin",
		name:     "fox",
	}

	friend.Listen()
	return friend
}

/*
type HandleFunc func(*bufio.ReadWriter)

type ChatRoom struct {
    friend   *Friend
    joins    chan net.Conn
    incoming chan string
    outgoing chan string
    handler  map[string]HandleFunc
}

func (chatRoom *ChatRoom) Broadcast(data string) {
    for _, friend := range chatRoom.friend {
        friend.outgoing <- data
    }
}

func (chatRoom *ChatRoom) Join(connection net.Conn) {
    friend := NewFriend(connection)
    chatRoom.friend = friend
    fmt.Println("friend connected")
    go func() {
        for {
            chatRoom.incoming <- <-friend.incoming
        }
    }()
}

func (chatRoom *ChatRoom) Listen() {
    go func() {
        for {
            select {
            case data := <-chatRoom.incoming:
                chatRoom.Broadcast(data)
            case conn := <-chatRoom.joins:
                chatRoom.Join(conn)
            }
        }
    }()
}

func NewChatRoom() *ChatRoom {
    chatRoom := &ChatRoom{
        friend:   *Friend,
        joins:    make(chan net.Conn),
        incoming: make(chan string),
        outgoing: make(chan string),
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
