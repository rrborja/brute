package ui

import (
	"sync"
	"fmt"
	"strings"
	"github.com/jroimartin/gocui"
)

var lock sync.Mutex

var gv *gocui.Gui

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

func EnsureGui() {
	if gv == nil {
		gv = <- globalView
	}
}

func newLines(countChan <- chan int) {
	for count := range countChan {
		func(count int) {
			EnsureGui()

			gv.Update(func(g *gocui.Gui) error {
				lock.Lock()
				defer lock.Unlock()

				loggerGui, err := g.View("logs")
				if err != nil {
					panic(err)
				}

				fmt.Fprint(loggerGui, strings.Repeat("\n", count))
				return nil
				})
			lock.Lock()
			lock.Unlock()
		}(count)
	}
}

func _log(messageChan <- chan string) {
	for message := range messageChan {
		func(message string) {
			EnsureGui()

			gv.Update(func(g *gocui.Gui) error {
				lock.Lock()
				defer lock.Unlock()

				loggerGui, err := g.View("logs")
				if err != nil {
					panic(err)
				}

				fmt.Fprint(loggerGui, "[LOG] ")
				if message[len(message)-1] == '\n' {
					fmt.Fprint(loggerGui, message)
				} else {
					fmt.Fprintln(loggerGui, message)
				}

				return nil
			})
			lock.Lock()
			lock.Unlock()
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
