package brute

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestValidSourceValue(t *testing.T) {
	sample := `Div().Value(func() {
		Form().Action("/login").Post().Value(func() {
			Span().Value("Username: "); TextField("username"); Break()
			Span().Value("Password: "); PasswordField("password"); Break()
			Input().Class("hello").Attributes(NewAttr("type", "hidden"), NewAttr("name", "csrf"), NewAttr("value", "h123h123o3")).Value(nil)
		}, Submit("Login"))
	})`

	source := Compile([]byte(sample))

	var actual []string
	for _, s := range source {
		actual = append(actual, string(s))
	}

	assert.Equal(t, sample, strings.Join(actual, "\n"))
}

func TestBlamePartSource(t *testing.T) {
	sample := `Div().Value(func() {
		Form().Action("/login").Post().Value(func() { d
			Span().Value("Username: "); TextField("username"); Break()
			Span().Value("Password: "); PasswordField("password"); Break()
			Input().Class("hello").Attributes(NewAttr("type", "hidden"), NewAttr("name", "csrf"), NewAttr("value", "h123h123o3")).Value(nil)
		}, Submit("Login"))
	})`

	assert.Equal(t, )
}


