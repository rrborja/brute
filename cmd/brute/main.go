package main

//go:generate echo "hello world"

import (
	"os"
	"log"
	. "github.com/rrborja/brute"
	. "github.com/rrborja/brute/cmd"
	. "github.com/rrborja/brute/cmd/ui"
	. "github.com/rrborja/brute/log"
)

var Version string

// brute set remote -url="192.168.1.152"
// brute unset remote
// brute check remote
// brute add endpoint -name=Ritchie -path=borja
// brute remove endpoint -name=Ritchie
// brute update endpoint -name=Ritchie
func main() {
	Logo(Version, true)

	if len(os.Args) > 1 {
		if err := ProcessArgument(os.Args[1:]...); err != nil {
			if err != nil {
				log.Fatal(err)
			}
		} else {
			os.Exit(0)
		}
	}

	Log("Checking contents...")

	if config, err := CheckCurrentProjectFolder(); err != nil {
		log.Fatal(err)
	} else {
		New(config)

		SetProjectName(config.Name)

		l := RunService()
		defer l.Close()

		e := RunEndpointService()
		defer e.Close()

		StartAuthorizer(config)

		StartEndpoints(config)
		Deploy(config)
	}
}
