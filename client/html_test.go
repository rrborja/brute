package client

import (
	"testing"
	"fmt"
)

func TestCorrectEvaluatedHtml(t *testing.T) {
	h := Div().Id("navigator").Class("nav-blue").Value(func() {
		Div().Class("class-1").Class("Green").Value("test s1")
		Div().Class("list").Class("Green").Value(func() {
			UnorderedList().Value("Hello 1", "2", "3")
			//UnorderedList().Value("Hello 2")
			//UnorderedList().Value("Hello 3")
		})
		Div().Class("class-1").Class("Green").Value("rat")

	})

	//h := Div().Id("navigator").Class("nav-blue").Value(func() {
	//	Div().Id("navigator2").Class("nav-blue2").Value(func() {
	//		Div().Id("navigator3").Class("nav-blue3").Value(func() {
	//
	//		})
	//	})
	//})

	//h := Div().Id("").Value("test")

	fmt.Println(h)

	t.Fail()
}