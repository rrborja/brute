package ui

import (
	"github.com/jroimartin/gocui"
	"fmt"
	"log"
)

var (
    version string
	globalView chan *gocui.Gui
)

func init() {
	globalView = make(chan *gocui.Gui)
}

func mainTerminal() {
	g, err := gocui.NewGui(gocui.OutputNormal)

	if err != nil {
		panic(err)
	}

	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, cleanTerminal); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if view, err := g.SetView("header", 0, -1, maxX, 7); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Autoscroll = true
		view.Wrap = true
		view.Frame = false

		fmt.Fprint(view, Logo(version, false))
	}
	if view, err := g.SetView("logs", 0, 8, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		view.Autoscroll = true
		view.Wrap = true
		view.Frame = false

		globalView <- g
	}
	return nil
}

func cleanTerminal(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func RunTerminal() {
	mainTerminal()
}