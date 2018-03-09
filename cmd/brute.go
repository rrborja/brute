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

package cmd

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/rrborja/brute"
	"github.com/rrborja/brute/cmd/templates"
	"gopkg.in/yaml.v2"

	. "github.com/rrborja/brute/log"
)

const (
	ServiceHost = "localhost"
	ServicePort = "11792"
	ServiceType = "tcp"
)

func RunService() net.Listener {
	l, err := net.Listen(ServiceType, ":"+ServicePort)
	if err != nil {
		LogError(ErrorLog{err, fmt.Sprintf("Error listening: %v", err)})
		os.Exit(1)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()
			if err != nil {
				LogError(ErrorLog{err, fmt.Sprintf("Error accepting: %v", err.Error())})
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
		LogError(ErrorLog{err, err.Error()})
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
	case "legal":
		return ProcessLegalMenu(args[1:]...)
	case "live":
		return DeployAsLive(args[1:]...)
	default:
		return fmt.Errorf("unknown command: %v", args[0])
	}
}

func DeployAsLive(args ...string) error {
	if len(args) > 0 {
		errors.New("unknown additional args: " + strings.Join(args, ", "))
	}
	brute.SetHttpPort(80)
	brute.SetSecureHttpPort(443)
	return errors.New("deploy")
}

func ProcessLegalMenu(args ...string) error {
	fmt.Println(`
Few innovative works of Free and Open Source software have really helped
designing, producing, and getting the Brute Web Engine to where it is
today. Where it is a legal doctrine and a requirement that Brute must
comply their license agreements, the command argument "legal" you have
supplied will list all copyright notices of the libraries that Brute
uses. On behalf of the Brute development community, thank you so much
for providing these fantastic masterpieces!

			                    - Ritchie Borja`)
	NewLines(2)

	prompt := promptui.Select{
		Label: "Libraries Brute uses",
		Items: Libraries(),
	}

	_, name, err := prompt.Run()
	check(err)

	fmt.Println(Summary(name))

	return nil
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

		Log(fmt.Sprintf("Endpoint %v successfully added\n", *name))
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
	defer Log("Setup complete!")

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