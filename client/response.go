package client

import "fmt"

func SetHeader(key string, value interface{}) error {
	context, _ := handlerSessions.Get(Gid())
	var ack bool
	return context.Call("RequestSession.Write",
		&EchoPacket{context.(*Context).SessionId,
		[]byte(fmt.Sprintf("%s=%v", key, value)), 40,
		}, &ack)
}

func SetHttpCode(statusCode int) {
	context, _ := handlerSessions.Get(Gid())
	context.WriteHeader(statusCode)
}