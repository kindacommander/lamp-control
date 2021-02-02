package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
)

const addr = "192.168.31.68:8888"

var commands = make(chan Command)

// Command contains instructions for a lamp to change it's features
type Command struct {
	Feature string `json:"feature"`
	Value   int    `json:"value"`
}

func (c *Command) String() string {
	return c.Feature + fmt.Sprint(c.Value)
}

func main() {
	fmt.Println("Golang HTTP to UDP proxy")
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func setupRoutes() {
	go connectToLamp()

	http.HandleFunc("/udp", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		var cmd Command
		err = json.Unmarshal(body, &cmd)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		commands <- cmd
	})
}

func connectToLamp() {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	for {
		cmd := <-commands
		conn.Write([]byte(cmd.String()))
	}
}
