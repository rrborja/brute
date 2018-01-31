package client

import (
	"fmt"
	"github.com/rrborja/brute"
	"github.com/silentred/gid"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"runtime"
)

var magicNumber = []byte{0x62, 0x72, 0x75, 0x74, 0x65}

var customIn = os.Stdin
var customOut = os.Stdout

var client *rpc.Client

type Handler interface {
	Lock()
	Unlock()
	Write([]byte) (int, error)
	Call(string, *brute.EchoPacket, *bool) error
}

type Context struct {
	Name string
	SessionId [32]byte
	Arguments map[string]string
	Rpc       func(string, interface{}, interface{}) error
	*sync.Mutex
}

func (context *Context) Write(buf []byte) (n int, err error) {
	var ack bool
	err = context.Rpc("RequestSession.Write", &brute.EchoPacket{context.SessionId, buf, 200}, &ack)
	n = len(buf)
	return
}

func (context *Context) Lock() {
	context.Mutex.Lock()
}

func (context *Context) Unlock() {
	context.Mutex.Unlock()
}

func (context *Context) Call(functionName string, packet *brute.EchoPacket, ack *bool) error {
	return context.Rpc(functionName, packet, ack)
}



func Out(data []byte, args ...interface{}) {
	// use runtime.Caller to restrict calling this method only the endpoint's handler source code
	context := handlerSessions[gid.Get()]

	context.Lock()
	defer context.Unlock()

	if len(args) > 0 {
		data = []byte(fmt.Sprintf(string(data), args))
	}

	context.Write(data)
}

func Echo(message string, args ...interface{}) {
	Out([]byte(message), args...)
}

func Handle(handler func(args map[string]string), callEvents <- chan Context) {
	for callEvent := range callEvents {
		go func(handlerSessions HandlerSessions, handler func(args map[string]string), callEvent Context) {
			defer func(callEvent Context) {
				var ack bool
				if err := callEvent.Rpc("RequestSession.Close",
					&brute.EchoPacket{SessionId: callEvent.SessionId},
					&ack); err != nil {
					panic(err)
				}
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					buf = buf[:runtime.Stack(buf, false)]
					fmt.Fprintf(os.Stderr, "Endpoint %s encountered an error: %v\n%s", callEvent.Name, r, buf)
				}
			}(callEvent)

			sessionId := gid.Get()
			writer := &callEvent

			handlerSessions.Set(sessionId, writer)
			writerSessions.Set(sessionId, writer)

			SetWriter(sessionId, &RenderStackHolder{root: new(RenderStack), writer: writer})

			// Start processing the endpoint while listening for writes to pass packets to the connected Client
			handler(callEvent.Arguments)
		}(handlerSessions, handler, callEvent)
	}
}

func SystemMessage(message string) {
	context := handlerSessions[gid.Get()]

	context.Lock()
	defer context.Unlock()

	var ack bool
	context.Call("RequestSession.Write", &brute.EchoPacket{context.(*Context).SessionId, []byte(message), 700}, &ack)
}

func Run(handler func(args map[string]string)) {

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
		if _, err := conn.Read(ack); err != nil {
			break
		}

		var sid [32]byte
		copy(sid[:], ack)

		var arguments map[string]string
		if err := client.Call("RequestSession.AcceptRpc", sid, &arguments); err == nil {
			callEvent <- Context{
				Name: source,
				SessionId: sid,
				Arguments: arguments,
				Rpc:       client.Call,
				Mutex:     new(sync.Mutex),
			}
		} else {
			panic(err)
		}
	}

	client.Close()
	os.Exit(0)
}
