package client

import (
	"fmt"
	"github.com/silentred/gid"
	"log"
	"net"
	"net/rpc"
	"os"
	"sync"
	"runtime"
	"reflect"
	"net/http"
	"net/url"
)

var magicNumber = []byte{0x62, 0x72, 0x75, 0x74, 0x65}

var customIn = os.Stdin
var customOut = os.Stdout

var client *rpc.Client

type Message url.Values

type Handler interface {
	Lock()
	Unlock()
	Write([]byte) (int, error)
	Call(string, *EchoPacket, *bool) error
}

type EchoPacket struct {
	SessionId [32]byte
	Body      []byte
	Code	  int
}

type Context struct {
	Name 		string
	SessionId 	[32]byte
	Method 		string
	Message
	Arguments 	map[string]string
	Rpc       	func(string, interface{}, interface{}) error

	*sync.Mutex
}

func (context *Context) SetContentType(mime string) {
	var ack bool
	context.Rpc("RequestSession.SetContentType", &EchoPacket{context.SessionId, []byte(mime), 200}, &ack)
}

func (context *Context) Write(buf []byte) (n int, err error) {
	var ack bool
	err = context.Rpc("RequestSession.Write", &EchoPacket{context.SessionId, buf, 200}, &ack)
	n = len(buf)
	return
}

func (context *Context) Lock() {
	context.Mutex.Lock()
}

func (context *Context) Unlock() {
	context.Mutex.Unlock()
}

func (context *Context) Call(functionName string, packet *EchoPacket, ack *bool) error {
	return context.Rpc(functionName, packet, ack)
}

func Gid() SessionId {
	return SessionId(gid.Get())
}

func Out(data []byte) (n int, err error) {
	// use runtime.Caller to restrict calling this method only the endpoint's handler source code
	context, _ := handlerSessions.Get(Gid())

	context.Lock()
	defer context.Unlock()

	return context.Write(data)
}

func Echo(message string, args ...interface{}) {
	Out([]byte(fmt.Sprintf(message, args)))
}

func SystemMessage(message string) {
	context := handlerSessions[Gid()]

	context.Lock()
	defer context.Unlock()

	var ack bool
	context.Call("RequestSession.Write", &EchoPacket{context.(*Context).SessionId, []byte(message), 700}, &ack)
}

func With(handlers ...interface{}) []interface{} {
	return handlers
}

func Run(handler func(args map[string]string), handlers ...interface{}) {

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

	nonGetHandlers := map[string]interface{}{http.MethodGet : handler}
	for _, nonGetHandler := range handlers {
		handler := nonGetHandler
		standardHandlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
		switch standardHandlerName {
		case "main.Create", "main.create":
			nonGetHandlers[http.MethodPost] = handler
		case "main.Update", "main.update":
			nonGetHandlers[http.MethodPut] = handler
		case "main.PartialUpdate", "main.partialUpdate":
			nonGetHandlers[http.MethodPatch] = handler
		case "main.Delete", "main.delete":
			nonGetHandlers[http.MethodDelete] = handler
		default:
			log.Print("Didn't I tell you not to modify the main() function of your endpoint code?")
		}
	}

	callEvent := make(chan Context, 100)

	go Handle(nonGetHandlers, callEvent)

	ack := make([]byte, 32)
	for {
		if _, err := conn.Read(ack); err != nil {
			break
		}

		var sid [32]byte
		copy(sid[:], ack)

		var rpcResponse struct{Method string; Message url.Values; Arguments map[string]string}
		if err := client.Call("RequestSession.AcceptRpc", sid, &rpcResponse); err == nil {
			callEvent <- Context{
				Name: 		source,
				SessionId: 	sid,
				Method: 	rpcResponse.Method,
				Message: 	Message(rpcResponse.Message),
				Arguments: 	rpcResponse.Arguments,
				Rpc:       	client.Call,
				Mutex:     	new(sync.Mutex),
			}
		} else {
			panic(err)
		}
	}

	client.Close()
	os.Exit(0)
}
