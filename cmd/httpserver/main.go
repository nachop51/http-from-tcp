package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handleVideoFunc)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

// func handlerFunc(w *response.Writer, req *request.Request) {
// 	res := []byte(`<html>
//   <head>
//     <title>200 OK</title>
//   </head>
//   <body>
//     <h1>Success!</h1>
//     <p>Your request was an absolute banger.</p>
//   </body>
// </html>`)
// 	statusCode := response.StatusOK

// 	if req.RequestLine.RequestTarget == "/yourproblem" {
// 		res = []byte(`<html>
//   <head>
//     <title>400 Bad Request</title>
//   </head>
//   <body>
//     <h1>Bad Request</h1>
//     <p>Your request honestly kinda sucked.</p>
//   </body>
// </html>`)
// 		statusCode = response.StatusBadRequest
// 	}

// 	if req.RequestLine.RequestTarget == "/myproblem" {
// 		res = []byte(`<html>
//   <head>
//     <title>500 Internal Server Error</title>
//   </head>
//   <body>
//     <h1>Internal Server Error</h1>
//     <p>Okay, you know what? This one is on me.</p>
//   </body>
// </html>`)
// 		statusCode = response.StatusInternal
// 	}

// 	w.WriteStatusLine(statusCode)
// 	w.WriteHeaders(response.GetDefaultHeaders(len(res)))
// 	w.WriteBody(res)
// }

func HandlerFunc(w *response.Writer, req *request.Request) {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		w.WriteStatusLine(response.StatusOK)
		w.WriteHeaders(response.GetDefaultHeaders(0))
		w.WriteBody([]byte{})
		return
	}

	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin/")

	res, err := http.Get("https://httpbin.org/" + path)
	if err != nil {
		w.WriteStatusLine(response.StatusInternal)
		res := []byte("Error fetching httpbin")
		w.WriteHeaders(response.GetDefaultHeaders(len(res)))
		w.WriteBody(res)
		return
	}

	w.WriteStatusLine(response.StatusOK)
	h := response.GetDefaultHeaders(0)
	delete(h, "content-length")
	h["host"] = "httpbin.org"
	h["transfer-encoding"] = "chunked"
	h["Trailer"] = "X-Content-Sha256, X-Content-Length"
	w.WriteHeaders(h)

	defer res.Body.Close()

	buf := make([]byte, 1024)

	resBody := make([]byte, 0)

	for {
		n, err := res.Body.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				w.WriteChunkedBodyDone()
				break
			}
			break
		}
		resBody = append(resBody, buf[:n]...)
		w.WriteChunkedBody(buf[:n])
	}

	shaValue := sha256.Sum256(resBody)
	trailers := headers.Headers{
		"X-Content-Sha256": fmt.Sprintf("%x", shaValue),
		"X-Content-Length": fmt.Sprintf("%d", len(resBody)),
	}
	w.WriteTrailers(trailers)
}

func handleVideoFunc(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget != "/video" {
		w.WriteStatusLine(response.StatusBadRequest)
		res := []byte("Not Found")
		w.WriteHeaders(response.GetDefaultHeaders(len(res)))
		w.WriteBody(res)
		return
	}

	videoBuff, err := os.ReadFile("assets/vim.mp4")
	if err != nil {
		w.WriteStatusLine(response.StatusInternal)
		res := []byte("file not found")
		w.WriteHeaders(response.GetDefaultHeaders(len(res)))
		w.WriteBody(res)
		return
	}

	w.WriteStatusLine(response.StatusOK)
	headers := response.GetDefaultHeaders(len(videoBuff))
	headers["content-type"] = "video/mp4"
	w.WriteHeaders(headers)
	w.WriteBody(videoBuff)
}
