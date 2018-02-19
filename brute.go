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
	"github.com/rjeczalik/notify"

	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"sync"
	"time"
	"github.com/elazarl/go-bindata-assetfs"
	"github.com/rrborja/brute/assets"
	"html/template"
	"strings"
	"net/url"
	. "github.com/rrborja/brute/log"
)

var cwd = ""
var gotool = filepath.Join(runtime.GOROOT(), "bin", "go")

var projectName string

var magicNumber = []byte{0x62, 0x72, 0x75, 0x74, 0x65}

var r *mux.Router

var endpoints sync.Map

var requestSession RequestSession

var template404Page *template.Template
var template700Page *template.Template

type ConnWriter struct {
	io.Reader
	net.Conn
	*sync.Mutex
}

func (connWriter *ConnWriter) Write(data []byte) (int, error) {
	connWriter.Lock()
	defer connWriter.Unlock()

	return connWriter.Conn.Write(data)
}

type Config struct {
	Name       string  `yaml:"name"`
	Remote     string  `yaml:"remote"`
	Authorizer string  `yaml:authorizer`
	Routes     []Route `yaml:,flow`
}

type Route struct {
	Path      string `yaml:"path"`
	Directory string `yaml:"directory"`
}

type ControllerEndpoint struct {
	ProjectName string
	Route
	runtimeFile string
}

type RequestSession struct {
	store map[[32]byte]*ContextHolder
	mutex sync.RWMutex
}
type ContextHolder struct {
	RpcArguments map[string]string
	Stream       chan *EchoPacket
	End          chan bool
	Message		 url.Values
	Method		 string
	Route
}

type EchoPacket struct {
	SessionId [32]byte
	Body      []byte
	Code	  int
}

type DefaultHeadersHandler struct {
	http.Handler
}

func defaultNotFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	template404Page.Execute(w, &struct{
		ProjectName string
		Path string
		RandomNoun string
	}{projectName, r.URL.Path, "MyEndpoint"})
}

func (sessions *RequestSession) AcceptRpc(id [32]byte, ack *struct{Method string; Message url.Values; Arguments map[string]string}) error {
	sessions.mutex.RLock()
	defer sessions.mutex.RUnlock()

	session := sessions.store[id]
	ack.Method = session.Method
	ack.Message = session.Message
	ack.Arguments = session.RpcArguments
	return nil
}

func (sessions *RequestSession) Write(packet *EchoPacket, ack *bool) error {
	sessions.mutex.RLock()
	defer sessions.mutex.RUnlock()

	session := sessions.store[packet.SessionId]
	session.Stream <- packet

	*ack = true
	return nil
}

func (sessions *RequestSession) Close(packet *EchoPacket, ack *bool) error {
	sessions.mutex.RLock()
	defer sessions.mutex.RUnlock()

	session := sessions.store[packet.SessionId]
	close(session.Stream)
	session.End <- true

	*ack = true
	return nil
}

func Delegate(w http.ResponseWriter, stream <-chan *EchoPacket) {
	for buf := range stream {
		switch buf.Code {
		case 700:
			template700Page.Execute(w, struct{
				ProjectName string
				Message string
			}{projectName, string(buf.Body)})
		default:
			if len(buf.Body) >= 3 && string(buf.Body[:3]) == "~ct" {
				w.Header().Set("Content-Type", string(buf.Body[3:]))
			} else {
				w.Write(buf.Body)
			}
		}
	}
}

func init() {
	_cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cwd = _cwd
	requestSession.store = make(map[[32]byte]*ContextHolder)

	loadTemplates()
}

func loadTemplates() {
	var data []byte
	var err error

	/* Parse 404 page template */
	data, err = assets.Asset("static/default-pages/404/404.html"); check(err)
	template404Page, err = template.New("404 Page Template").Parse(string(data)); check(err)

	/* Parse 700 page template */
	data, err = assets.Asset("static/default-pages/700/700.html"); check(err)
	template700Page, err = template.New("700 Page Template").Parse(string(data)); check(err)
}

func New(config *Config) {
	os.Mkdir("bin/endpoints", 0700)
	os.Mkdir("bin/hosted", 0700)
	os.Mkdir("bin/hosted/static", 0700)
	os.Mkdir("bin/hosted/assets", 0700)
	os.Mkdir("bin/temp", 0700)
	os.Mkdir("bin/temp/db", 0700)
	os.Mkdir("bin/build", 0700)

	for _, route := range config.Routes {
		buildEndpoint(route)
	}

	r = mux.NewRouter()

	addy, err := net.ResolveTCPAddr("tcp", "localhost:12000")
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	rpc.Register(&requestSession)
	go rpc.Accept(inbound)
}

func HostStaticFiles() {
	assets := http.FileServer(&assetfs.AssetFS{Asset: assets.Asset, AssetDir: assets.AssetDir, AssetInfo: assets.AssetInfo, Prefix: "static"})
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", assets))
}

func SetProjectName(name string) {
	projectName = name
}

func CleanUp() {
	os.Remove("bin")
}

func rebuildRootEndpoint(route Route) (string, error) {
	Log("Building " + route.Directory)

	tmpBuilds := filepath.Join("bin", "build")
	endpointBuilds := filepath.Join("bin", "endpoints")

	var out, routeDirectory string
	if len(route.Path) > 0 {
		out = filepath.Join(cwd, tmpBuilds, route.Directory)
		routeDirectory = filepath.Join(cwd, "src", route.Directory)
	} else {
		out = filepath.Join(cwd, tmpBuilds, "root")
		routeDirectory = filepath.Join(cwd, "src")
	}

	sourceFile := filepath.Join(routeDirectory, "main.go")

	cmd := exec.Command(gotool, "build", "-o", out, sourceFile)

	stdout, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	reason, _ := ioutil.ReadAll(stdout)
	if len(reason) > 0 {
		Log(fmt.Sprintf("%s", reason))
	}

	if err := cmd.Wait(); err != nil {
		return "", err
	} else {
		Log("Done!\n")
		err := os.Rename(out, filepath.Join(cwd, endpointBuilds, route.Directory))
		if err != nil {
			panic(err)
		}
		return routeDirectory, nil
	}

}

func rebuildEndpoint(route Route) (string, error) {
	return rebuildRootEndpoint(route)
}

func buildEndpoint(route Route) {
	sourceEndpointDirectory, err := rebuildEndpoint(route)
	if err != nil {
		panic(err)
	}

	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(sourceEndpointDirectory, c, notify.All); err != nil {
		log.Fatal(err)
	}
	go func(c <-chan notify.EventInfo) {
		for range c {
			Log(fmt.Sprintf("Attempting to restart %s due to code changes...\n", route.Directory))
			if endpoint, ok := endpoints.Load(route.Directory); ok {
				_, err := rebuildEndpoint(route)
				if err == nil {
					endpoints.Delete(route.Directory)
					StartEndpoint(route)
				} else {
					log.Println(err)
				}

				err = endpoint.(*ConnWriter).Close()
				if err != nil {
					LogError(ErrorLog{err, err.Error()})
					debug.PrintStack()
					// TODO: Fix by implementing a feature that will auto restart the endpoint's RPC connection
				}
			}
		}
	}(c)
}

func HostRootEndpoint() {
	root := Route{"", "root"}

	buildEndpoint(root)
	StartRootEndpoint(root)

	source := filepath.Join(cwd, "bin", "endpoints", "root")
	endpoint := &ControllerEndpoint{projectName, root, source}

	r.Handle("/", endpoint).Name("root")
}

func Deploy(config *Config) {
	CleanUp()

	for _, route := range config.Routes {
		build := filepath.Join(cwd, "bin", "endpoints", route.Directory)

		endpoint := &ControllerEndpoint{config.Name, route, build}
		r.Handle(route.Path, endpoint).Name(route.Directory)
	}

	r.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Accept-Ranges", "bytes")
		w.Header().Set("Server", "brute.io")
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("ETag", `"1"`)
		favicon, _ := assets.Asset("static/default-pages/favicon.ico")
		w.Write(favicon)
		w.WriteHeader(200)
	})

	r.NotFoundHandler = http.HandlerFunc(defaultNotFoundHandler)
	HostStaticFiles()

	HostRootEndpoint()

	srv := &http.Server{Addr: ":8080", Handler: http.HandlerFunc(func (w http.ResponseWriter, req *http.Request) {
		target := "https://" + req.Host + req.URL.Path
		if len(req.URL.RawQuery) > 0 {
			target += "?" + req.URL.RawQuery
		}
		w.Header().Set("server", "brute.io")
		w.Header().Add("X-comment", "You must use HTTPS next time.")
		http.Redirect(w, req, target,
			http.StatusTemporaryRedirect)
	})}

	srv.SetKeepAlivesEnabled(true)

	go srv.ListenAndServe()

	secureSrv := &http.Server{Addr: ":8443", Handler: r}
	secureSrv.SetKeepAlivesEnabled(true)
	secureSrv.ListenAndServeTLS("cert.pem", "tls.key")
}

func AddEndpoint(route *Route) {

}

func RandomSessionId(ip string, unixSeconds int64) [32]byte {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(unixSeconds))

	r := make([]byte, 4)
	binary.LittleEndian.PutUint16(r, RandomNumber())

	hash := sha256.New()
	hash.Write(append([]byte(ip), append(b, r...)...))

	var finalHash [32]byte
	copy(finalHash[:], hash.Sum(nil))

	return finalHash
}

func (controller *ControllerEndpoint) RedirectEndpointOnLoading(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Endpoint " + controller.Directory + " is still loading. Try again for a few seconds"))
}

func (controller *ControllerEndpoint) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	sid := RandomSessionId(r.RemoteAddr, time.Now().Unix())

	w.Header().Set("X-Brute-Session-ID", hex.EncodeToString(sid[:]))
	w.Header().Set("Server", "brute.io")

	if _, ok := endpoints.Load(controller.Route.Directory); !ok {
		controller.RedirectEndpointOnLoading(w, r)
		return
	}

	context := &ContextHolder{Stream: make(chan *EchoPacket, 100), End: make(chan bool, 1)}
	defer close(context.End)

	requestSession.mutex.Lock()
	requestSession.store[sid] = context
	requestSession.mutex.Unlock()

	context.Route = controller.Route

	pathArgs := mux.Vars(r)
	for k, v := range r.URL.Query() {
		if existing, ok := pathArgs[k]; ok {
			log.Printf("Path through key %s already exists. [existing: %v, this: %v]", k, existing, v)
		} else {
			pathArgs[k] = strings.Join(v, "~")
		}
	}

	context.Method = r.Method
	context.RpcArguments = pathArgs

	r.ParseForm()
	context.Message = r.Form

	if endpoint, ok := endpoints.Load(controller.Route.Directory); ok {
		endpoint.(*ConnWriter).Write(sid[:])
	}

	Delegate(w, context.Stream)

	<-context.End
}

func StartEndpoints(config *Config) {
	for _, route := range config.Routes {
		StartEndpoint(route)
	}
}

func StartRootEndpoint(route Route) {
	Log(fmt.Sprintf("Starting endpoint %s", route.Directory))

	var out string
	var env string
	if len(route.Path) > 0 {
		out = filepath.Join(cwd, "bin", "endpoints", route.Directory)
		env = fmt.Sprintf("ROUTE=%s", route.Directory)
	} else {
		out = filepath.Join(cwd, "bin", "endpoints", "root")
		env = fmt.Sprintf("ROUTE=%s", "root")
	}

	cmd := exec.Command(out)
	cmd.Env = []string{env}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//cmdOut, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		LogError(ErrorLog{err, fmt.Sprintf("Could not run endpoint daemon %s", route.Directory)})
	}
}

func StartEndpoint(route Route) {
	StartRootEndpoint(route)
}

func RunEndpointService() net.Listener {
	Log("Starting Endpoint Service...")

	l, err := net.Listen("tcp", ":11000")
	if err != nil {
		LogError(ErrorLog{err, fmt.Sprintf("Can't start listening for endpoints: %v", err)})
		os.Exit(1)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()

			Log("Incoming endpoint connection: " + conn.RemoteAddr().String())

			if err != nil {
				LogError(ErrorLog{err, fmt.Sprintf("Error accepting: %v", err.Error())})
				continue
			}

			bin := make([]byte, 5)
			conn.Read(bin)

			if !HandshakeFormat(bin) {
				Log("Cannot accept an incoming connection")
				continue
			}

			size := make([]byte, 4)
			conn.Read(size)

			s, _ := strconv.Atoi(string(size))
			block := make([]byte, s)
			conn.Read(block)

			routeDirectory := string(block)

			Log(fmt.Sprintf("Connection accepted from %s\n", routeDirectory))

			endpoints.Store(routeDirectory, &ConnWriter{Mutex: new(sync.Mutex), Conn: conn})
		}
	}()

	return l
}

func HandshakeFormat(initial []byte) bool {
	for i, m := range initial {
		if uint8(m) != uint8(magicNumber[i]) {
			return false
		}
	}
	return true
}

func RandomNumber() uint16 {
	var number uint16
	binary.Read(rand.Reader, binary.LittleEndian, &number)
	return number
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
