package json

import (
	"github.com/silentred/gid"
	"io"
	"fmt"
)

func Test() {
	Map("_href", List(
			Map("name", "ritchie"),
		),
	)
	Map("id", 2)
}

type session struct {


	endBuf chan interface{}
	waitBuf chan interface{}

	Type
	io.Writer
}

type Type int
const (
	LIST = 1 << iota
	MAP
)

func (session *session) List(values ...interface{}) (element JsonString) {
	if session.Type == 0 {
		session.Type = LIST
		session.Write([]byte("["))
	}

	//for _, value := range values {
	//
	//}

	panic("")
}

func (session *session) Map(key string, value interface{}) (entry JsonString) {
	if session.Type == 0 {
		session.Type = MAP
		session.Write([]byte("{"))
	} else {
		session.Write([]byte(","))
	}

	switch v := value.(type) {
	case int: entry = JsonString(fmt.Sprintf(`"%s":%d`, key, v))
	case float32, float64: entry = JsonString(fmt.Sprintf(`"%s":%f`, key, value))
	case string: entry = JsonString(fmt.Sprintf(`"%s":"%s"`, key, v))
	case bool, JsonString: entry = JsonString(fmt.Sprintf(`"%s":%v`, key, v))
	default: panic(fmt.Errorf("unknown type: %v", v))
	}

	session.Write([]byte(entry))

	return
}

type JsonString string

type ResponseWriter interface {
	List(value ...interface{}) JsonString
	Map(key string, value interface{}) JsonString
}

var jsonWriterSession map[int64]ResponseWriter

func init() {
	jsonWriterSession = make(map[int64]ResponseWriter)
}

func AddSession(sessionId int64, writer io.Writer) (chan interface{}, chan interface{}) {
	endBuf := make(chan interface{})
	waitBuf := make(chan interface{})

	s := &session{endBuf: endBuf, waitBuf: waitBuf, Writer: writer}

	go func(s *session) {
		<- endBuf
		switch s.Type {
		case MAP: s.Write([]byte("}"))
		case LIST: s.Write([]byte("]"))
		}
		close(waitBuf)
	}(s)

	jsonWriterSession[sessionId] = s

	return endBuf, waitBuf
}

func CloseSession(sessionId int64) {
	s := jsonWriterSession[sessionId].(*session)
	close(s.endBuf)
	<- s.waitBuf
}

func JsonWriterSession(sessionId int64) ResponseWriter {
	return jsonWriterSession[sessionId]
}

func List(value ...interface{}) interface{} {
	sessionId := gid.Get()
	return JsonWriterSession(sessionId).List(value)
}

func Map(key string, value interface{}) interface{} {
	sessionId := gid.Get()
	return JsonWriterSession(sessionId).Map(key, value)
}