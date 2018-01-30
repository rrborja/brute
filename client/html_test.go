package client

import (
	"testing"
	"fmt"
	"github.com/silentred/gid"
	"os"
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
		Value("Username: "); TextField("username")
		Value("Password: "); PasswordField("password")
	}, Submit("submit"))

	fmt.Println(h)

	t.Fail()
}

type test Context
func (test *test) Write(b []byte) (n int, e error) {
	fmt.Print(string(b))
	return
}

func Test700EvaluatedHtml(t *testing.T) {

	writerSessions[gid.Get()] = os.Stdout

	message := `What you are seeing right now is
the default page generated by Brute Web Engine
designed by Ritchie Borja. Brute Web Engine will
be released on mid-March 2018.

Thank you for visiting. https://github.com/rrborja`


	Title("Welcome &middot; ")
	PreloadStylesheet("/static/default-pages/style.css")
	Author("brute.io")

	Div().Value(func() {
		Header().Value(func() {
			Div().Class("row").Value(func() {
				Div().Class("logo").Value(func() {
					Span().Class("logo-icon").Value("❆")
					Span().Class("logo-text").Value("βrute")
				})
			})
		})
		Main().Class("main-content-particle-js").Value(func() {
			Div().Class("main-content").Value(func() {
				Div().Class("row").Value(func() {
					H1().Class("info-title").Value("200")
					Paragraph().Value(message)
					A("https://github.com/rrborja").Class("box").
						Value("This is a generated 200 page.")
				})
			})
		})
		Footer().Value(func() {
			Div().Class("row").Value(func() {
				Span().Value(`This page including its content and style
will not be displayed on your production sites.
You must specify your custom default pages.`)
			})
		})
	})

	t.Fail()
}