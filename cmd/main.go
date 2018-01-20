// Copyright 2018 Ritchie Borja
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"github.com/rrborja/brute"
	"github.com/rrborja/brute/cmd/templates"
	"gopkg.in/yaml.v2"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const (
	ServiceHost = "localhost"
	ServicePort = "11792"
	ServiceType = "tcp"
)

// brute set remote -url="192.168.1.152"
// brute unset remote
// brute check remote
// brute add endpoint -name=Ritchie -path=borja
// brute remove endpoint -name=Ritchie
// brute update endpoint -name=Ritchie
func main() {
	if len(os.Args) > 1 {
		if err := ProcessArgument(os.Args[1:]...); err != nil {
			check(err)
		} else {
			os.Exit(0)
		}
	}

	fmt.Println("Checking contents...")
	if config, err := CheckCurrentProjectFolder(); err != nil {
		log.Fatal(err)
	} else {
		brute.New(config)

		l := RunService()
		defer l.Close()

		e := brute.RunEndpointService()
		defer e.Close()

		brute.StartEndpoints(config)

		brute.Deploy(config)
	}
}

func RunService() net.Listener {
	l, err := net.Listen(ServiceType, ":"+ServicePort)
	if err != nil {
		fmt.Println("Error listening: ", err)
		os.Exit(1)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				continue
			}

			handleInternalCommand(conn)
		}
	}()

	return l
}

type ServiceMessage struct {
	Command string
}

func (msg *ServiceMessage) Execute() {
	switch msg.Command {
	case "add-endpoint":
	case "update-endpoint":
	case "remove-endpoint":
	}
}

func handleInternalCommand(c net.Conn) {
	d := json.NewDecoder(c)

	var msg ServiceMessage

	err := d.Decode(&msg)
	if err != nil {
		log.Printf(err.Error())
	}

	//TODO: handle master server commands and signals

	c.Close()
}

func ProcessArgument(args ...string) error {
	switch strings.ToLower(args[0]) {
	case "add":
		return ProcessTypeForAdd(args[1:]...)
	case "remove":
		return ProcessTypeForRemove(args[1:]...)
	case "update":
		return ProcessTypeForUpdate(args[1:]...)
	default:
		return fmt.Errorf("unknown command: %v", args[0])
	}
}

func ProcessTypeForAdd(args ...string) error {
	if len(args) == 0 {
		return errors.New("expected additional arguments for add")
	}

	switch strings.ToLower(args[0]) {
	case "endpoint":
		name := flag.String("name", "", "the name of the endpoint")
		path := flag.String("path", "", "the URI path of the endpoint")

		flag.CommandLine.Parse(args[1:])

		if len(*name) == 0 {
			*name = SetNameOfRoute()
		}
		if len(*path) == 0 {
			*path = SetPathOfRoute()
		}

		config, err := CheckCurrentProjectFolder()
		check(err)

		for _, route := range config.Routes {
			if route.Directory == *name {
				return fmt.Errorf("cannot add an existing endpoint %v", *name)
			}
		}

		config.Routes = append(config.Routes, brute.Route{Path: *path, Directory: *name})

		CreateProjectFiles(config)

		fmt.Printf("Endpoint %v successfully added\n", *name)
	default:
		return fmt.Errorf("unknown feature %v", args[0])
	}

	return nil
}

func ProcessTypeForRemove(args ...string) error {
	return nil
}

func ProcessTypeForUpdate(args ...string) error {
	return nil
}

func CheckCurrentProjectFolder() (*brute.Config, error) {
	cwd, err := os.Getwd()
	check(err)

	files, err := ioutil.ReadDir(cwd)
	check(err)

	if len(files) > 0 {
		config, err := CheckExistingValidProject(files)
		if err != nil {
			return nil, err
		}
		return config, nil
	}

	return CreateNewProject()
}

func CheckExistingValidProject(files []os.FileInfo) (*brute.Config, error) {
	var validProject bool
	for _, file := range files {
		if !file.IsDir() && file.Name() == ".brute.yml" {
			validProject = true
			break
		}
	}

	if !validProject {
		return nil, errors.New("to be able to setup the current directory, it has to contain no files")
	}

	data, err := ioutil.ReadFile(".brute.yml")
	check(err)

	var config *brute.Config
	err = yaml.Unmarshal([]byte(data), &config)
	check(err)

	return config, nil
}

func CreateNewProject() (*brute.Config, error) {
	defer fmt.Println("Setup complete!")

	config := &brute.Config{}
	defer CreateProjectFiles(config)

	{
		prompt := promptui.Prompt{
			Label:   "Project name [untitled]",
			Default: "untitled",
		}

		name, err := prompt.Run()
		check(err)

		config.Name = name
	}

	ConfigRoutes(config)

	return config, nil
}

func SetNameOfRoute() string {
	routePrompt := promptui.Prompt{
		Label: "Name",
	}

	name, err := routePrompt.Run()
	check(err)

	return name
}

func SetPathOfRoute() string {
	routePrompt := promptui.Prompt{
		Label: "Path",
	}

	path, err := routePrompt.Run()
	check(err)

	return path
}

func ConfigRoutes(config *brute.Config) {
	prompt := promptui.Prompt{
		Label:     "Create new route",
		IsConfirm: true,
		Default:   "Y",
	}

	_, err := prompt.Run()

	if err == nil {
		routes := make([]brute.Route, 1)
		for {
			var route brute.Route

			route.Directory = SetNameOfRoute()
			route.Path = SetPathOfRoute()

			routes[len(routes)-1] = route

			routePrompt := promptui.Prompt{
				Label:     "Create more routes",
				IsConfirm: true,
				Default:   "Y",
			}

			_, err := routePrompt.Run()

			if err != nil {
				break
			}

			routes = append(routes, make([]brute.Route, 1)...)
		}

		config.Routes = routes
	}
}

type EmptyGoTemplate struct {
	Path string
}

func CreateProjectFiles(config *brute.Config) {
	os.Mkdir("src", 0700)

	for _, route := range config.Routes {
		routeDirectory := filepath.Join("src", route.Directory)
		os.Mkdir(routeDirectory, 0700)

		mainFile := filepath.Join(routeDirectory, "main.go")
		//if that endpoint logic exists
		if _, err := os.Stat(mainFile); err == nil {
			continue
		}

		goFile, err := os.Create(mainFile)
		check(err)

		w := bufio.NewWriter(goFile)

		emptyGoTemplate := EmptyGoTemplate{route.Path}
		tmpl, err := template.New("controller").Parse(templates.EmptyController)
		check(err)
		check(tmpl.Execute(w, emptyGoTemplate))

		w.Flush()
		goFile.Close()
	}

	ModifyProjectConfig(config)
}

func ModifyProjectConfig(config *brute.Config) {
	configData, err := yaml.Marshal(config)
	check(err)

	check(ioutil.WriteFile(".brute.yml", configData, 0700))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
