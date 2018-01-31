package client

import (
	"sync"
	"io"
)

type HandlerSessions map[int64]Handler
type WriterSessions map[int64]io.Writer

var (
	handlerSessions HandlerSessions
	handlerSessionsMutex sync.RWMutex
)

func (handlerSession HandlerSessions) Get(sessionId int64) (handler Handler, ok bool) {
	handlerSessionsMutex.RLock()
	defer handlerSessionsMutex.RUnlock()

	handler, ok = handlerSession[sessionId]
	return
}
func (handlerSession HandlerSessions) Set(sessionId int64, handler Handler) {
	handlerSessionsMutex.Lock()
	defer handlerSessionsMutex.Unlock()

	handlerSession[sessionId] = handler
}

var (
	writerSessions WriterSessions
	writerSessionsMutex sync.RWMutex
)

func (writerSession WriterSessions) Get(sessionId int64) (writer io.Writer, ok bool) {
	writerSessionsMutex.RLock()
	defer writerSessionsMutex.RUnlock()

	writer, ok = writerSession[sessionId]
	return
}
func (writerSession WriterSessions) Set(sessionId int64, writer io.Writer) {
	writerSessionsMutex.Lock()
	defer writerSessionsMutex.Unlock()

	writerSession[sessionId] = writer
	return
}

func init() {
	handlerSessions = make(HandlerSessions)
	writerSessions = make(WriterSessions)
}