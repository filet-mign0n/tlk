package main

import (
	"bufio"
	"net"
)

type Friend struct {
	rw       *bufio.ReadWriter
	incoming chan string
	outgoing chan string
	status   string
	name     string
    key      []byte
}

func (friend *Friend) Read() {
	for {
		line, _ := friend.rw.ReadString('\n')
		if len(line) > 0 {
            decryptMsg, _ := decrypt(friend.key, line[:len(line)-1])
			convo.chat(decryptMsg[:len(decryptMsg)-1])
		}
	}
}

func (friend *Friend) Write() {
	for data := range friend.outgoing {
        data = data + "\n"
        data, _ := encrypt(friend.key, data)
        //convo.log("crypto: "+data)
        data = data + "\n"
		friend.rw.WriteString(data)
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
        key:      []byte(*key),
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
