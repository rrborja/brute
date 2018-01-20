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
)

var cwd = ""
var gotool = filepath.Join(runtime.GOROOT(), "bin", "go")

var magicNumber = []byte{0x62, 0x72, 0x75, 0x74, 0x65}

var r *mux.Router

var endpoints sync.Map

var requestSession RequestSession

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

type RequestSession map[[32]byte]*ContextHolder
type ContextHolder struct {
	RpcArguments map[string]string
	Stream       chan []byte
	End          chan bool
	Route
}

type EchoPacket struct {
	SessionId [32]byte
	Body      []byte
}

func (session RequestSession) AcceptRpc(id [32]byte, ack *map[string]string) error {
	*ack = session[id].RpcArguments
	return nil
}

func (session RequestSession) Write(packet *EchoPacket, ack *bool) error {
	session[packet.SessionId].Stream <- packet.Body
	*ack = true
	return nil
}

func (session RequestSession) Close(packet *EchoPacket, ack *bool) error {
	close(session[packet.SessionId].Stream)
	session[packet.SessionId].End <- true
	*ack = true
	return nil
}

func Delegate(w io.Writer, stream <-chan []byte) {
	for buf := range stream {
		w.Write(buf)
	}
}

func init() {
	_cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cwd = _cwd
	requestSession = make(RequestSession)
}

func New(config *Config) {
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

	rpc.Register(requestSession)
	go rpc.Accept(inbound)
}

func CleanUp() {
	os.Remove("endpoints")
}

func rebuildEndpoint(route Route) string {
	out := filepath.Join(cwd, "endpoints", route.Directory)
	routeDirectory := filepath.Join(cwd, "src", route.Directory)
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
		fmt.Printf("%s\n", reason)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}

	return routeDirectory
}

func buildEndpoint(route Route) {
	sourceEndpointDirectory := rebuildEndpoint(route)

	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(sourceEndpointDirectory, c, notify.All); err != nil {
		log.Fatal(err)
	}
	go func(c <-chan notify.EventInfo) {
		for range c {
			fmt.Printf("Attempting to restart %s due to code changes...\n", route.Directory)
			if endpoint, ok := endpoints.Load(route.Directory); ok {
				err := endpoint.(*ConnWriter).Close()
				if err != nil {
					fmt.Println(err)
					debug.PrintStack()
				} else {
					endpoints.Delete(route.Directory)
					rebuildEndpoint(route)
					StartEndpoint(route)
				}
			}
		}
	}(c)
}

func Deploy(config *Config) {
	CleanUp()
	os.Mkdir("endpoints", 0700)

	for _, route := range config.Routes {
		build := filepath.Join(cwd, "endpoints", route.Directory)

		endpoint := &ControllerEndpoint{config.Name, route, build}
		r.Handle(route.Path, endpoint).Name(route.Directory)
	}

	http.ListenAndServe(":8080", r)
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

	context := &ContextHolder{Stream: make(chan []byte, 100), End: make(chan bool, 1)}
	defer close(context.End)

	requestSession[sid] = context

	context.Route = controller.Route

	pathArgs := mux.Vars(r)
	context.RpcArguments = pathArgs

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

func StartEndpoint(route Route) {
	fmt.Printf("Starting endpoint %s\n", route.Directory)

	out := filepath.Join(cwd, "endpoints", route.Directory)
	cmd := exec.Command(out, route.Directory)
	cmd.Env = []string{fmt.Sprintf("ROUTE=%s", route.Directory)}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	//cmdOut, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		fmt.Printf("Could not run endpoint daemon %s", route.Directory)
	}
}

func RunEndpointService() net.Listener {
	fmt.Println("Starting Endpoint Service...")

	l, err := net.Listen("tcp", ":11000")
	if err != nil {
		fmt.Println("Can't start listening for endpoints: ", err)
		os.Exit(1)
	}

	go func() {
		for {
			// Listen for an incoming connection.
			conn, err := l.Accept()

			fmt.Println("Incoming connection:")

			if err != nil {
				fmt.Println("Error accepting: ", err.Error())
				continue
			}

			bin := make([]byte, 5)
			conn.Read(bin)

			if !HandshakeFormat(bin) {
				fmt.Println("Cannot accept an incoming connection")
				continue
			}

			size := make([]byte, 4)
			conn.Read(size)

			s, _ := strconv.Atoi(string(size))
			block := make([]byte, s)
			conn.Read(block)

			routeDirectory := string(block)

			fmt.Printf("Connection accepted from %s\n", routeDirectory)

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
