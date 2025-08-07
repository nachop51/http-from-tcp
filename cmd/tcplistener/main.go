package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	// for line := range lines {
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		fmt.Println("Connection accepted")

		lines := getLinesChannel(conn)

		for line := range lines {
			fmt.Println(line)
		}
		conn.Close()
		fmt.Println("Connection closed")
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		var buffer []byte
		for {
			data := make([]byte, 8)

			read, err := f.Read(data)
			if err == io.EOF {
				if len(buffer) > 0 {
					lines <- string(buffer)
				}
				close(lines)
				return
			}

			buffer = append(buffer, data[:read]...)

			if i := bytes.IndexByte(buffer, '\n'); i != -1 {
				lines <- string(buffer[:i])
				buffer = buffer[i+1:]
			}
		}
	}()

	return lines
}
