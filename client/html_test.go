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
			Break()
			UnorderedList().Value(func() {
				for _, item := range items {
					Value(item)
				}
			})
		})
		Break()
		A("https://brute.io").Value("Click Here")
		Div().Class("class-1").Class("Green").Value("rat")
	})

	fmt.Println(h)

	t.Fail()
}

func TestFormEvaluatedHtml(t *testing.T) {
	h := Form().Action("/action_page.php").Get().Value(func() {
		Value("Username: "); Textfield("username")
		Value("Password: "); PasswordField("password")
	}, Submit("submit"))

	fmt.Println(h)

	t.Fail()
}