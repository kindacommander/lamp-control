package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
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

// State contain current lamp config
type State struct {
	PoweredOn  bool `json:"power"`
	Effect     int  `json:"effect"`
	Brightness int  `json:"brightness"`
	Speed      int  `json:"speed"`
	Scale      int  `json:"scale"`
}

func setupRoutes() {
	go connectToLamp()

	http.HandleFunc("/cmd", func(w http.ResponseWriter, r *http.Request) {
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
	http.HandleFunc("/curr-state", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Current state request")
		enableCors(&w)
		if r.Method == "OPTIONS" {
			return
		}

		conn, err := net.Dial("udp", addr)
		if err != nil {
			panic(err)
		}

		cmd := Command{"GET", 0}
		conn.Write([]byte(cmd.String()))

		res := make([]byte, 32)
		conn.Read(res)
		conn.Close()

		res = bytes.Trim(res, "\x00")

		fmt.Println(res)
		valuesStr := strings.Split(strings.Trim(string(res), "CURR "), " ")
		fmt.Println(valuesStr)

		vals := make([]int, 5)
		for i, val := range valuesStr {
			v, err := strconv.Atoi(val)
			if err != nil {
				panic(err)
			}
			vals[i] = v
		}

		state := State{PoweredOn: itob(vals[4]), Effect: vals[0], Brightness: vals[1], Speed: vals[2], Scale: vals[3]}
		json, err := json.Marshal(state)
		if err != nil {
			panic(err)
		}
		w.Write(json)
	})
}

func itob(i int) bool {
	if i == 0 {
		return false
	}
	return true
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

func main() {
	fmt.Println("Golang HTTP to UDP proxy")
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)
	setupRoutes()
	log.Fatal(http.ListenAndServe(":8080", nil))
}
