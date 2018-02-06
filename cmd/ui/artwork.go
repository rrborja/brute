package ui

import (
	"fmt"
	"strings"
	"bytes"
)

const icon = `
  _\/ .\/_
  . \  / .
_\___\/___/_
 /.  /\  .\
  __/. \__
   /\  /\`

const logo = `
 _                _         _
| |__  _ __ _   _| |_ ___  (_) ___
| '_ \| '__| | | | __/ _ \ | |/ _ \
| |_) | |  | |_| | ||  __/_| | (_) |
|_.__/|_|   \__,_|\__\___(_)_|\___/
`

const magicNumbers = "62:72:75:74:65:2E 69 6F "
const magicNumberColorOrder = "\x1B[91m,\x1B[31m,\x1B[95m,\x1B[34m,\x1B[96m,\x1B[0m"

func Logo(version string, allowClear bool) string {
	output := bytes.NewBufferString("")

	iconLines := strings.Split(icon, "\n")
	logoLines := strings.Split(logo, "\n")

	for i, iconLine := range iconLines {
		if len(iconLine) < 13 && i < len(iconLines) {
			iconLine += strings.Repeat(" ", 13 - len(iconLine))
		}
		fmt.Fprint(output, iconLine)
		if len(logoLines) - 1 > i {
			fmt.Fprintln(output, logoLines[i])
		}
	}

	fmt.Fprint(output, "  ")

	colors := strings.Split(magicNumberColorOrder, ",")
	for i, magicNumber := range strings.Split(magicNumbers, ":") {
		if len(colors) > i {
			fmt.Fprint(output, colors[i] + magicNumber + " ")
		} else {
			fmt.Fprint(output, magicNumber)
		}
	}

	fmt.Fprint(output, "\x1B[1;37m" + version + "\x1B[0m\n")

	return output.String()
}