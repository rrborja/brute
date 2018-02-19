package client

import "io"

type ContentWriter interface {
	Writer() io.Writer
}

func SetWriter(sessionId int64, writer ContentWriter) {

}

