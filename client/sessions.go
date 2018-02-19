package client

import (
	"sync"
	"io"
)

type SessionId int64

func (sessionId SessionId) Persist() bool {
	if _, ok := sessionSignals[sessionId]; ok {
		return false
	}
	sessionSignals[sessionId] = make(chan interface{})
	return true
}

func (sessionId SessionId) Cleanup() (done chan interface{}, ok bool) {
	done, ok = sessionSignals[sessionId]
	delete(sessionSignals, sessionId)
	return
}

type SessionSignals map[SessionId]chan interface{}
var sessionSignals SessionSignals

type HandlerSessions map[SessionId]Handler
type WriterSessions map[SessionId]io.Writer

var (
	handlerSessions HandlerSessions
	handlerSessionsMutex sync.RWMutex
)

func (handlerSession HandlerSessions) Get(sessionId SessionId) (handler Handler, ok bool) {
	handlerSessionsMutex.RLock()
	defer handlerSessionsMutex.RUnlock()

	handler, ok = handlerSession[sessionId]
	return
}
func (handlerSession HandlerSessions) Set(sessionId SessionId, handler Handler) {
	handlerSessionsMutex.Lock()
	defer handlerSessionsMutex.Unlock()

	handlerSession[sessionId] = handler
}

var (
	writerSessions WriterSessions
	writerSessionsMutex sync.RWMutex
)

func (writerSession WriterSessions) Get(sessionId SessionId) (writer io.Writer, ok bool) {
	writerSessionsMutex.RLock()
	defer writerSessionsMutex.RUnlock()

	writer, ok = writerSession[sessionId]
	return
}
func (writerSession WriterSessions) Set(sessionId SessionId, writer io.Writer) {
	writerSessionsMutex.Lock()
	defer writerSessionsMutex.Unlock()

	writerSession[sessionId] = writer
	return
}

func init() {
	handlerSessions = make(HandlerSessions)
	writerSessions = make(WriterSessions)
	sessionSignals = make(SessionSignals)
}