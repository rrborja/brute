package client

import (
	"reflect"
	"fmt"
	"log"
	"runtime"
	"os"
	"net/http"
)

func Handle(handler map[string]interface{}, callEvents <- chan Context) {
	for callEvent := range callEvents {
		go func(handlerSessions HandlerSessions, handlers map[string]interface{}, callEvent Context) {
			sessionId := Gid()
			sessionId.Persist()

			defer func(callEvent Context) {
				if done, ok := sessionId.Cleanup(); ok {
					(<- <-done)()
					close(done)
					sessionId.Purge()
				}

				var ack bool
				if err := callEvent.Rpc("RequestSession.Close",
					&EchoPacket{SessionId: callEvent.SessionId},
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

			//clientHtml.SetWriter(sessionId, clientHtml.CreateRenderStackHolder(new(clientHtml.RenderStack), writer))

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
