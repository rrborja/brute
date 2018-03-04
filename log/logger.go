package log

import (
	"sync"
	"strings"
)

var lock sync.Mutex

var _Log chan string

var _LogError chan ErrorLog

var _NewLines chan int

type ErrorLog struct {
	Error error
	String string
}

func CloseLogger() {
	close(_Log)
	close(_LogError)
	close(_NewLines)
}

func init() {
	_Log = make(chan string)
	_LogError = make(chan ErrorLog)
	_NewLines = make(chan int)

	go newLines(_NewLines)
	go _log(_Log)
	go logError(_LogError)
}

func newLines(countChan <- chan int) {
	for count := range countChan {
		func(count int) {
			lock.Lock()
			defer lock.Unlock()
			print(strings.Repeat("\n", count))
		}(count)
	}
}

func _log(messageChan <- chan string) {
	for message := range messageChan {
		func(message string) {

			lock.Lock()
			defer lock.Unlock()
			println("[LOG] " + message)

		}(message)
	}
}

func logError(messageChan <- chan ErrorLog) {
	for message := range messageChan {
		Log(message.String)
	}
}

func Log(log string) {
	_Log <- log
}

func LogError(log ErrorLog) {
	_LogError <- log
}

func NewLines(count int) {
	_NewLines <- count
}
