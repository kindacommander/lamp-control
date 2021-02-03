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
		enableCors(&w)
		if r.Method == "OPTIONS" {
			return
		}

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

		fmt.Println(cmd.String())

		commands <- cmd
	})
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
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
