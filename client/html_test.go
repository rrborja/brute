package client

import (
	"testing"
	"fmt"
)

func TestCorrectEvaluatedHtml(t *testing.T) {
	items := []string{"hello", ",", " ", "world"}

	h :=
	Div().Value(func() {
		Div().Value("test s1")
		Div().Value(func() {
			Br()
			UnorderedList().Value(func() {
				for _, item := range items {
					Value(item)
				}
			})
		})
		Br()
		A("https://brute.io").Value("Click Here")
		Div().Class("class-1").Class("Green").Value("rat")
	})

	/*
	<div>
	   <div>test s1</div>
	   <div>
		  <br>
		  <ul>
			 <li>Hello 1</li>
			 <li>&lt;html&gt;</li>
			 <li>3</li>
		  </ul>
	   </div>
	   <br><a href="https://brute.io">Click Here</a>
	   <div class="class-1 Green">rat</div>
	</div>
	*/

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