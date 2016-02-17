// swaginfo project main.go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/justinas/alice"
)

// NetInterfaces returns a map with the interface name as a key, containing
// the addresses as a slice of strings.
func NetInterfaces() (map[string][]string, error) {

	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	res := make(map[string][]string)
	for _, iface := range ifaces {
		res[iface.Name] = nil
		addrs, err := iface.Addrs()
		if err != nil {
			break
		} else {
			for _, addr := range addrs {
				// Add the address to the iface to the result list.
				res[iface.Name] = append(res[iface.Name], addr.String())
			}
		}
	}
	return res, nil
}

// ContainerInfo contains the information to be returned in the JSON response.
// And it can act as a http.Handler.
type ContainerInfo struct {
	Hostname  string
	Addresses map[string][]string
}

func getContainerInfo() (*ContainerInfo, error) {
	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	// Get the interfaces.
	ni, err := NetInterfaces()
	if err != nil {
		return nil, err
	}
	ci := &ContainerInfo{
		Hostname:  hn,
		Addresses: ni,
	}
	return ci, nil
}

func (ci *ContainerInfo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if ci.Addresses == nil {
		if cit, err := getContainerInfo(); err == nil {
			ci.Hostname = cit.Hostname
			ci.Addresses = cit.Addresses
		}
	}
	if ci.Addresses == nil {
		http.Error(w, "error getting the container info", http.StatusInternalServerError)
		return
	}
	jsb, err := json.Marshal(ci)
	if err != nil {
		http.Error(w, "error json marshal", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsb)
}

// InfoHandler returns the container info in JSON.
func InfoHandler(w http.ResponseWriter, r *http.Request) {
	ci := info
	jsb, err := json.Marshal(ci)
	if err != nil {
		http.Error(w, "error json marshal", http.StatusInternalServerError)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(jsb)
}

// NetInterfacesString returns a string representation of the response from
// NetInterfaces.
func NetInterfacesString(ni map[string][]string) string {
	w := &bytes.Buffer{}

	for k := range ni {
		for _, s := range ni[k] {
			fmt.Fprintf(w, "[%s]%s\n", k, s)
		}
	}
	return w.String()
}

// AddrSiceString returns a string representation of []net.Addr.
func AddrSliceString(addrs []net.Addr) string {
	var w = &bytes.Buffer{}

	for _, a := range addrs {
		fmt.Fprintf(w, "%s : %s\n", a.Network(), a.String())
	}
	return w.String()
}

var info *ContainerInfo

// RunServer starts the server that returns the container info and blocks.
func RunServer() error {
	info, _ = getContainerInfo()
	ci := &ContainerInfo{}
	c := alice.New(LoggingHandler)
	http.Handle("/info", c.Then(ci))
	return http.ListenAndServe(":8080", nil)
}

// LoggingHandler is a middleware wraper that writes out the start and duration
// of the request handler plus the requested URI.
func LoggingHandler(h http.Handler) http.Handler {
	f := func(w http.ResponseWriter, r *http.Request) {
		t0 := time.Now()
		h.ServeHTTP(w, r)
		d := time.Now().Sub(t0)
		fmt.Printf("%s, %s, %s\n", t0.String(), d.String(), r.URL.String())
	}
	return http.HandlerFunc(f)
}

func main() {
	RunServer()
}
