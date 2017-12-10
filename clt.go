package main

import (
	"io"
	"net"
	//"log"
	"bufio"
	"encoding/gob"
	"strconv"
	"strings"
	"sync"

	t "github.com/gizak/termui"
	//"flag"

	"github.com/pkg/errors"
)

const (
	Port = ":7777"
	Key  = "lol"
)

type Friend struct {
	name     string
	incoming chan string
	outgoing chan string
	reader   *bufio.Reader
	writer   *bufio.Writer
}

type complexData struct {
	N int
	S string
	M map[string]int
	P []byte
	C *complexData
}

func Open(addr string) (*bufio.ReadWriter, error) {

	//log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

type HandleFunc func(*bufio.ReadWriter)

type Endpoint struct {
	listener net.Listener
	handler  map[string]HandleFunc

	m sync.RWMutex
}

func NewEndpoint() *Endpoint {
	return &Endpoint{
		handler: map[string]HandleFunc{},
	}
}

func (e *Endpoint) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

func (e *Endpoint) Listen() error {
	var err error
	e.listener, err = net.Listen("tcp", Port)
	if err != nil {
		return errors.Wrap(err, "Unable to listen on "+e.listener.Addr().String()+"\n")
	}
	//log.Println("Listen on", e.listener.Addr().String())
	for {
		//log.Println("Accept a connection request.")
		conn, err := e.listener.Accept()
		if err != nil {
			//log.Println("Failed accepting a connection request:", err)
			continue
		}
		//log.Println("Handle incoming messages.")
		go e.handleMessages(conn)
	}
}

func (e *Endpoint) handleMessages(conn net.Conn) {
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	for {
		//log.Print("Received command")
		cmd, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			//log.Println("Reached EOF - close this connection.\n  ---")
			//log.Println("EOF")
			return
		case err != nil:
			//log.Println("\nError reading command. Got: '"+cmd+"'\n", err)
			return
		}

		cmd = strings.Trim(cmd, "\n ")
		//log.Println(cmd)

		e.m.RLock()
		handleCommand, ok := e.handler[cmd]
		e.m.RUnlock()
		if !ok {
			//log.Println("Command '"+cmd+"' is not registered.")
			return
		}
		handleCommand(rw)
	}
}

func handleKey(rw *bufio.ReadWriter) {
	//log.Print("Received KEY:")
	s, err := rw.ReadString('\n')
	if err != nil {
		//log.Println("Cannot read from connection.\n")
	}
	s = strings.Trim(s, "\n ")
	//log.Println("got key:", s)
	*convo.Chats = s
	t.Render(t.Body)
	if s == Key {
		_, err = rw.WriteString("ACK_KEY")
	} else {
		_, err = rw.WriteString("BAD_KEY")
	}
	if err != nil {
		//log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		//log.Println("Flush failed.", err)
	}

}

func handleStrings(rw *bufio.ReadWriter) {
	//log.Print("Receive STRING message:")
	s, err := rw.ReadString('\n')
	if err != nil {
		//log.Println("Cannot read from connection.\n")
	}
	s = strings.Trim(s, "\n ")
	//log.Println(s)
	_, err = rw.WriteString("Thank you.\n")
	if err != nil {
		//log.Println("Cannot write to connection.\n", err)
	}
	err = rw.Flush()
	if err != nil {
		//log.Println("Flush failed.", err)
	}
}

func handleGob(rw *bufio.ReadWriter) {
	//log.Print("Receive GOB data:")
	var data complexData

	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		//log.Println("error decoding GOB dta:", err)
		return
	}
	//log.Printf("Outer complexData struct: \n%#v\n", data)
	//log.Printf("Inner complexData struct: \n%#v\n", data.C)
}

func client(ip string) error {
	testStruct := complexData{
		N: 23,
		S: "string data",
		M: map[string]int{"one": 1, "two": 2, "three": 3},
		P: []byte("lol"),
		C: &complexData{
			N: 256,
			S: "Recursive structs? Piece of cake!",
			M: map[string]int{"01": 1, "10": 2, "11": 3},
		},
	}

	rw, err := Open(ip + Port)
	if err != nil {
		return errors.Wrap(err, "Client: Failed to open conn to "+ip+Port)
	}
	//log.Println("Send the string request.")
	n, err := rw.WriteString("STRING\n")
	if err != nil {
		return errors.Wrap(err, "Could not send the STRING request ("+strconv.Itoa(n)+" bytes written)")
	}
	n, err = rw.WriteString("Additional data.\n")
	if err != nil {
		return errors.Wrap(err, "Could not send additional String data ("+strconv.Itoa(n)+" bytes written)")
	}
	//log.Println("Flush the buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}
	//log.Println("Read the reply.")
	response, err := rw.ReadString('\n')
	if err != nil {
		return errors.Wrap(err, "Client: Failed to read the reply: '"+response+"'")
	}
	//log.Println("STRING request: got a response:", response)
	//log.Println("Send a struct as GOB:")
	//log.Printf("Outer complexData struct: \n%#v\n", testStruct)
	//log.Printf("Inner complexData struct: \n%#v\n", testStruct.C)
	enc := gob.NewEncoder(rw)
	n, err = rw.WriteString("GOB\n")
	if err != nil {
		return errors.Wrap(err, "Could not write GOB data("+strconv.Itoa(n)+" bytes written)")
	}
	err = enc.Encode(testStruct)
	if err != nil {
		return errors.Wrapf(err, "Encode failed for struct: %#v", testStruct)
	}
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}
	for {
	}
	return nil
}

func server() error {
	endpoint := NewEndpoint()
	endpoint.AddHandleFunc("KEY", handleKey)
	endpoint.AddHandleFunc("STRING", handleStrings)
	endpoint.AddHandleFunc("GOB", handleGob)

	return endpoint.Listen()
}

func main() {
	go runTermui()
	err := server()
	if err != nil {
		//log.Println("Error:", errors.WithStack(err))
	}
	//log.Println("Server done")

	/*
	   connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")
	   flag.Parse()

	   if *connect != "" {
	       err := client(*connect)
	       if err != nil {
	           //log.Println("Error:", errors.WithStack(err))
	       }
	       //log.Println("Client done.")
	   }
	*/
}

func init() {
	//log.SetFlags(//log.Lshortfile)
}
