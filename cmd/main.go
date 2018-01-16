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
	"log"
	"os"
	"errors"
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"github.com/manifoldco/promptui"
	"github.com/rrborja/brute"
	"html/template"
	"bufio"
	"github.com/rrborja/brute/cmd/templates"
)

const pathSeparator = string(os.PathSeparator)

func main() {
	log.Println("Checking contents...")
	if config, err := CheckCurrentProjectFolder(); err != nil {
		log.Fatal(err)
	} else {
		brute.Deploy(config)
	}
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
	defer log.Println("Setup complete!")

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

	{
		prompt := promptui.Prompt{
			Label:     "Create new route",
			IsConfirm: true,
			Default: "Y",
		}

		_, err := prompt.Run()

		if err == nil {
			routes := make([]brute.Route, 1)
			for {
				var route brute.Route
				{
					routePrompt := promptui.Prompt{
						Label: "Name",
					}

					name, err := routePrompt.Run()
					check(err)

					route.Directory = name
				}
				{
					routePrompt := promptui.Prompt{
						Label: "Path",
					}

					path, err := routePrompt.Run()
					check(err)

					route.Path = path
				}
				routes[len(routes) - 1] = route

				routePrompt := promptui.Prompt{
					Label: "Create more routes",
					IsConfirm: true,
					Default: "Y",
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

	return config, nil
}

type EmptyGoTemplate struct {
	Path string
}

func CreateProjectFiles(config *brute.Config) {
	check(os.Mkdir(config.Name, 0700))

	for _, route := range config.Routes {
		check(os.Mkdir(config.Name + pathSeparator + route.Directory, 0700))

		goFile, err := os.Create(config.Name + pathSeparator + route.Directory + pathSeparator + "main.go")
		check(err)

		w := bufio.NewWriter(goFile)

		emptyGoTemplate := EmptyGoTemplate{route.Path}
		tmpl, err := template.New("controller").Parse(templates.EmptyController)
		check(err)
		check(tmpl.Execute(w, emptyGoTemplate))

		w.Flush()
		goFile.Close()
	}

	configData, err := yaml.Marshal(config)
	check(err)

	check(ioutil.WriteFile(".brute.yml", configData, 0700))
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}