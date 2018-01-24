package client

import (
	"testing"
	"fmt"
)

func TestCorrectEvaluatedHtml(t *testing.T) {
	h :=
		Div().Id("navigator").Class("nav-blue").Value(func() {
		Div().Class("class-1").Class("Green").Value("test s")
	})

	fmt.Println(h.Render())

	t.Fail()
}