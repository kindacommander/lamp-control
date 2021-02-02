package main

import "net"

func main() {
	const addr = "192.168.31.68:8888"

	conn, err := net.Dial("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	conn.Write([]byte("P_ON"))
}
