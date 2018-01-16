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

package brute

import (
	"github.com/gorilla/mux"
	"net/http"
	"os/exec"
	"log"
	"os"
	"path/filepath"
	"runtime"
)

var r *mux.Router

type Config struct {
	Name string `yaml:"name"`
	Routes []Route `yaml:,flow`
}

type Route struct {
	Path string `yaml:"path"`
	Directory string `yaml:"directory"`
}

type ControllerEndpoint struct {
	ProjectName string
	Route string
	runtimeFile string
}

func New() {
	r = mux.NewRouter()
}

func Deploy(config *Config) {
	for _, route := range config.Routes {
		endpoint := &ControllerEndpoint{config.Name, route.Path, route.Directory}
		r.Handle(route.Path, endpoint)
	}
	http.ListenAndServe(":8080", r)
}

func AddEndpoint(route *Route) {

}

func (controller *ControllerEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//reader, writer := controller.Execute(arguments...)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	gotool := filepath.Join(runtime.GOROOT(), "bin", "go")

	args := []string{"run", cwd + string(os.PathSeparator) + controller.ProjectName + string(os.PathSeparator) + controller.runtimeFile + string(os.PathSeparator) + "main.go"}

	pathArgs := mux.Vars(r)
	if len(pathArgs) > 0 {
		args = append(args, "--")
		for k, v := range pathArgs {
			args = append(args, "-" + k + "=" + v)
		}
	}

	cmd := exec.Command(gotool, args...)

	cmd.Stdout = w

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}



}

