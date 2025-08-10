package main

import (
	"fmt"
	"log"
	"net"

	"httpfromtcp/internal/request"
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

		// lines := getLinesChannel(conn)

		// for line := range lines {
		// 	fmt.Println(line)
		// }

		req, err := request.RequestFromReader(conn)
		if err != nil {
			fmt.Println("Error reading request:", err)
			continue
		}

		fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
			req.RequestLine.Method, req.RequestLine.RequestTarget, req.RequestLine.HttpVersion)
		fmt.Println("Headers:")
		for k, v := range req.Headers {
			fmt.Printf("- %s: %s\n", k, v)
		}
		fmt.Printf("Body:\n%s\n", string(req.Body))

		conn.Close()
		fmt.Println("Connection closed")
	}
}

// func getLinesChannel(f io.ReadCloser) <-chan string {
// 	lines := make(chan string)

// 	go func() {(f io.ReadCloser) <-chan string {
// 	lines := make(chan string)

// 	go func() {
// 		var buffer []byte
// 		for {
// 			data := make([]byte, 8)

// 			read, err := f.Read(data)
// 			if err == io.EOF {
// 				if len(buffer) > 0 {
// 					lines <- string(buffer)
// 				}
// 				close(lines)
// 				return
// 			}

// 			buffer = append(buffer, data[:read]...)

// 			if i := bytes.IndexByte(buffer, '\n'); i != -1 {
// 				lines <- string(buffer[:i])
// 				buffer = buffer[i+1:]
// 			}
// 		}
// 	}()

// 	return lines
// }
// 		var buffer []byte
// 		for {
// 			data := make([]byte, 8)

// 			read, err := f.Read(data)
// 			if err == io.EOF {
// 				if len(buffer) > 0 {
// 					lines <- string(buffer)
// 				}
// 				close(lines)
// 				return
// 			}

// 			buffer = append(buffer, data[:read]...)

// 			if i := bytes.IndexByte(buffer, '\n'); i != -1 {
// 				lines <- string(buffer[:i])
// 				buffer = buffer[i+1:]
// 			}
// 		}
// 	}()

// 	return lines
// }
