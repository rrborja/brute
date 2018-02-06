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
	"reflect"
	"net/http"
	"net/url"
	"github.com/rrborja/brute/client/html/meta/mime"
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
	Call(string, *brute.EchoPacket, *bool) error
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

func (context *Context) SetContentType(mime mime.Mime) {
	var ack bool
	context.Rpc("RequestSession.SetContentType", &brute.EchoPacket{context.SessionId, []byte(mime), 200}, &ack)
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
	context, _ := handlerSessions.Get(gid.Get())

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

func Handle(handler map[string]interface{}, callEvents <- chan Context) {
	for callEvent := range callEvents {
		go func(handlerSessions HandlerSessions, handlers map[string]interface{}, callEvent Context) {
			sessionId := gid.Get()

			defer func(callEvent Context) {
				renderStackHolder, ok := Writer(sessionId)
				if ok {
					if len(renderStackHolder.headElements) > 0 || renderStackHolder.body != nil {
						renderStackHolder.writer.Write([]byte("</body></html>"))
					}
				}

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

			writer := &callEvent

			handlerSessions.Set(sessionId, writer)
			writerSessions.Set(sessionId, writer)

			SetWriter(sessionId, &RenderStackHolder{root: new(RenderStack), writer: writer})

			// Start processing the endpoint while listening for writes to pass packets to the connected Client
			if handler, ok := handlers[callEvent.Method]; ok {
				funcValue := reflect.ValueOf(handler)
				funcType := funcValue.Type()

				if funcType.Kind() != reflect.Func {
					panic(fmt.Errorf("did I just tell you not to modify the main() for this endpoint %s? %v\n", callEvent.Name, handler))
				}

				switch callEvent.Method {
				case http.MethodGet, http.MethodDelete:
					if funcType.NumIn() == 0 {
						handler.(func())()
					} else if funcType.NumIn() == 1 && funcType.In(0).AssignableTo(reflect.TypeOf(map[string]string{})) {
						handler.(func(map[string]string))(callEvent.Arguments)
					} else {
						panic(fmt.Errorf("invalid function signature for this endpoint %s: %v\n", callEvent.Name, handler))
					}
				case http.MethodPost, http.MethodPut, http.MethodPatch:
					if funcType.NumIn() == 0 {
						handler.(func())()
					} else if funcType.NumIn() == 1 {
						argType := funcType.In(0)
						if argType.AssignableTo(reflect.TypeOf(Message{})) {
							handler.(func(Message))(callEvent.Message)
						} else if argType.AssignableTo(reflect.TypeOf(map[string]string{})) {
							handler.(func(map[string]string))(callEvent.Arguments)
						} else {
							panic(fmt.Errorf("invalid function signature for this endpoint %s: %v\n", callEvent.Name, handler))
						}
					} else if funcType.NumIn() == 2 {
						argType1st := funcType.In(0)
						argType2nd := funcType.In(1)
						if argType1st.AssignableTo(reflect.TypeOf(Message{})) && argType2nd.AssignableTo(reflect.TypeOf(map[string]string{})) {
							handler.(func(Message, map[string]string))(callEvent.Message, callEvent.Arguments)
						} else if argType2nd.AssignableTo(reflect.TypeOf(Message{})) && argType1st.AssignableTo(reflect.TypeOf(map[string]string{})) {
							handler.(func(map[string]string, Message))(callEvent.Arguments, callEvent.Message)
						} else {
							panic(fmt.Errorf("invalid function signature for this endpoint %s: %v\n", callEvent.Name, handler))
						}
					} else {
						panic(fmt.Errorf("invalid function signature for this endpoint %s: %v\n", callEvent.Name, handler))
					}
				default:
					log.Printf("unsupported method %s for this endpoint %s\n", callEvent.Method, callEvent.Name)
				}
			} else {
				log.Printf("Method %v is incompatible. Running a Get method for endpoint %s instead\n", callEvent.Method, callEvent.Name)
				handlers[http.MethodGet].(func(map[string]string))(callEvent.Arguments)
			}
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
