package json

import (
	"io"
	"fmt"
	"bufio"
	"github.com/rrborja/brute/client"
)

type session struct {

	listStream ListFunc

	touched bool

	endBuf  chan interface{}
	waitBuf chan interface{}

	jsonType Type
	*bufio.Writer
}

type Type int

const (
	LIST = 1 << iota
	MAP
)

func (s *session) Write(buf []byte) {
	defer s.Writer.Flush()
	s.Writer.Write(buf)
}

func (s *session) List(values ...interface{}) {
	s.touched = true

	if s.jsonType == 0 {
		s.jsonType = LIST
		s.Write([]byte("["))
	} else {
		s.Write([]byte(","))
	}

	var start bool

	for _, value := range values {
		if start {
			s.Write([]byte(","))
		} else {
			start = true
		}

		switch v := value.(type) {
		case func() interface{}:
			newVal := v()
			s.formatElement(newVal)
		default:
			s.formatElement(v)
		}
	}
}
func (s *session) formatElement(value interface{}) {
	switch v := value.(type) {
	case int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64:
		s.Write([]byte(fmt.Sprintf("%d", v)))
	case float32, float64:
		s.Write([]byte(fmt.Sprintf("%f", v)))
	case string:
		s.Write([]byte(fmt.Sprintf(`"%s"`, v)))
	case bool:
		s.Write([]byte(fmt.Sprintf("%v", v)))
	case func():
		func(v func()) {
			s.jsonType = 0
			defer func(s *session) { s.jsonType = LIST }(s)
			defer s.Write([]byte("}"))
			v()
		}(v)
	case ListFunc: {
		func(v func()) {
			s.jsonType = 0
			defer func(s *session) { s.jsonType = LIST }(s)
			defer s.Write([]byte("]"))
			v()
		}(v)
	}
	default:
		panic(fmt.Errorf("unknown type: %v", v))
	}
}

func isOnSession(s ResponseWriter) bool {
	return s.(*session).touched
}

func (s *session) Map(key string, value interface{}) {
	s.touched = true

	if s.jsonType == 0 {
		s.jsonType = MAP
		s.Write([]byte("{"))
	} else {
		s.Write([]byte(","))
	}

	var entry JsonString

	switch v := value.(type) {
	case int, int8, int16, int32, int64,
	uint, uint8, uint16, uint32, uint64:
		entry = JsonString(fmt.Sprintf(`"%s":%d`, key, v))
	case float32, float64:
		entry = JsonString(fmt.Sprintf(`"%s":%f`, key, value))
	case string:
		entry = JsonString(fmt.Sprintf(`"%s":"%s"`, key, v))
	case bool:
		entry = JsonString(fmt.Sprintf(`"%s":%v`, key, v))
	case func():
		s.Write([]byte(fmt.Sprintf(`"%s":`, key)))
		s.jsonType = 0
		defer func(s *session) { s.jsonType = MAP }(s)
		defer func() { s.Write([]byte("}")) }()

		v()

		return
	case ListFunc:
		s.Write([]byte(fmt.Sprintf(`"%s":`, key)))
		s.jsonType = 0
		defer func(s *session) { s.jsonType = MAP }(s)
		defer func() { s.Write([]byte("]")) }()

		v()

		if s.listStream != nil {
			s.listStream = nil
		}

		return
	default:
		panic(fmt.Errorf("unknown type: %v", v))
	}

	s.Write([]byte(entry))
}

func (s *session) Type() Type {
	return s.jsonType
}

type JsonString string

type ResponseWriter interface {
	List(value ...interface{})
	Map(key string, value interface{})
	Type() Type
}

var jsonWriterSession map[client.SessionId]ResponseWriter

func init() {
	jsonWriterSession = make(map[client.SessionId]ResponseWriter)
}

func AddSession(sessionId client.SessionId, writer io.Writer) (chan interface{}, chan interface{}) {
	endBuf := make(chan interface{})
	waitBuf := make(chan interface{})

	s := &session{endBuf: endBuf, waitBuf: waitBuf, Writer: bufio.NewWriter(writer)}

	go func(s *session) {
		<-endBuf
		switch s.jsonType {
		case MAP:
			s.Write([]byte("}"))
		case LIST:
			s.Write([]byte("]"))
		}
		close(waitBuf)
	}(s)

	jsonWriterSession[sessionId] = s

	return endBuf, waitBuf
}

func CloseSession(sessionId client.SessionId) {
	s := jsonWriterSession[sessionId].(*session)

	if s.listStream != nil {
		s.listStream()
		s.listStream = nil
	}

	close(s.endBuf)
	s.endBuf = nil
	<-s.waitBuf
}

func JsonWriterSession(sessionId client.SessionId) ResponseWriter {
	return jsonWriterSession[sessionId]
}

type ListFunc func()

func List(value ...interface{}) ListFunc {
	sessionId := client.Gid()
	s := JsonWriterSession(sessionId)

	callback := func() {
		s.List(value...)
		s.(*session).listStream = nil
	}

	s.(*session).listStream = callback

	return callback
}

func Map(key string, value interface{}) {
	sessionId := client.Gid()
	JsonWriterSession(sessionId).Map(key, value)
}

func Element(value interface{}) func() interface{} {
	return func() interface{} {
		return value
	}
}