package client

import (
	"fmt"
	"github.com/rrborja/brute"
	"github.com/silentred/gid"
	"log"
	"net"
	"net/rpc"
	"os"
	"strings"
	"sync"
)

var magicNumber = []byte{0x62, 0x72, 0x75, 0x74, 0x65}

var customIn = os.Stdin
var customOut = os.Stdout

var client *rpc.Client

var handlerSessions HandlerSessions

type HandlerSessions map[int64]*Context

type Context struct {
	SessionId [32]byte
	Arguments []string
	Rpc       func(string, interface{}, interface{}) error
	*sync.Mutex
}

func init() {
	//os.Stdin = nil
	//os.Stdout = nil

	handlerSessions = make(HandlerSessions)
}

func Out(data []byte) {
	// use runtime.Caller to restrict calling this method only the endpoint's handler source code
	context := handlerSessions[gid.Get()]

	var ack bool
	context.Rpc("RequestSession.Write", &brute.EchoPacket{context.SessionId, data}, &ack)

}

func Echo(message string) {
	Out([]byte(message))
}

func Handle(handler func(args ...string), callEvents <-chan Context) {
	for callEvent := range callEvents {
		handlerSessions[gid.Get()] = &callEvent
		handler(callEvent.Arguments...)
		var ack bool
		if err := callEvent.Rpc("RequestSession.Close", &brute.EchoPacket{SessionId: callEvent.SessionId}, &ack); err != nil {
			panic(err)
		}
	}
}

func Run(handler func(args ...string)) {

	conn, err := net.Dial("tcp", "localhost:11000")
	if err != nil {
		panic(err)
	}

	source := os.Getenv("ROUTE")
	size := fmt.Sprintf("%04d", len(source))
	message := append(magicNumber, append([]byte(size), []byte(source)...)...)

	conn.Write(message)

	// At this point, This endpoint has been plugged to the master brute server
	// Succeeding conn.Reads are always be session ID until connection closes

	client, err = rpc.Dial("tcp", "localhost:12000")
	if err != nil {
		log.Fatal(err)
	}

	callEvent := make(chan Context, 100)

	go Handle(handler, callEvent)

	ack := make([]byte, 32)
	for {
		conn.Read(ack)

		var sid [32]byte
		copy(sid[:], ack)

		var arguments string
		if err := client.Call("RequestSession.AcceptRpc", sid, &arguments); err == nil {
			callEvent <- Context{
				SessionId: sid,
				Arguments: strings.Split(arguments, " "),
				Rpc:       client.Call,
				Mutex:     new(sync.Mutex),
			}
		} else {
			panic(err)
		}
	}
}
