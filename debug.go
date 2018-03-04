package brute

import (
	"errors"
	"fmt"
	"bytes"
	"strings"
)

type Source []string

type EndpointFunc func() []byte

func (endpointFunc EndpointFunc) Close() error {
	return errors.New("this is a 500 endpoint logic")
}

func (endpointFunc EndpointFunc) Write(data []byte) (int, error) {
	panic("not implemented")
}

func Compile(source []byte) Source {
	var compiledSource []string

	var start int
	for end, s := range source {
		if s == '\n' {
			compiledSource = append(compiledSource, string(source[start:end]))
			start = end+1
		}
	}

	compiledSource = append(compiledSource, string(source[start:]))

	return compiledSource
}

func (source Source) snippetRange(col int) (int, int) {
	max := 15
	if len(source) < max {
		return 1, max
	}

	start := col - max/2
	end := col + max/2

	if start < 0 {
		end += -start
	}

	if end > len(source) {
		end = len(source) - 1
		start = len(source) - max
	}

	return start, end
}

func sliceToRealInt(chars string) (i int) {
	for j, ch := range chars {
		factor := 1
		for k:=0; k<len(chars) - j - 1; k++ {
			factor *= 10
		}
		i += (int(ch) - '0') * factor
	}
	return
}

func (source Source) buildDebugPage(route Route, reason []byte) (EndpointFunc, error) {
	if reason[0] != '#' || len(reason) == 0 {
		return nil, errors.New("not a compiler error")
	}

	var start int
	for ; reason[start] != '\n' && start<len(reason)-1; start++ {}

	reason = reason[start+1:]

	var paths, descs string
	var rows, cols int


	reasons := strings.Split(string(reason), ":")

	paths = string(reasons[0])
	rows = sliceToRealInt(reasons[1])
	cols = sliceToRealInt(reasons[2])
	descs = string(reasons[3])

	_,_ = descs, cols


	return func() []byte {

		var line bytes.Buffer
		start, end := source.snippetRange(rows)
		line.WriteString(fmt.Sprintf("<pre><code>%s</code></pre>", strings.Join(source[start:end+1], "\n")))
		return []byte(fmt.Sprintf(`<!DOCTYPE html><html><head><meta charset="UTF-8"><title>%s</title></head><body><div id="code-1" class="code-block"><h3>%s</h3>%s</div></body></html>`, descs, paths, line.String()))
	}, nil
}

func BuildDebugEndpoint(route Route, sourceCode, reasons []byte) {
	//TODO: Generate source code for broken endpoint
	endpointFunc, err := Compile(sourceCode).buildDebugPage(route, reasons)
	check(err)

	endpoints.Delete(route.Directory)
	endpoints.Store(route.Directory, endpointFunc)

}