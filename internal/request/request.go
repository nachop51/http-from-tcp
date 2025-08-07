package request

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
)

const bufferSize int = 8
const (
	Initialized = iota
	Done
)

type Request struct {
	RequestLine RequestLine
	State       int
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize)
	readToIndex := 0

	var req Request = Request{
		State: Initialized,
	}

	for req.State != Done {
		read, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				req.State = Done
				break
			}
			return nil, err
		}
		readToIndex += read

		parsed, err := req.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsed > 0 {
			newBuf := make([]byte, readToIndex-parsed)
			copy(newBuf, buf[parsed:])
			buf = newBuf
			readToIndex -= parsed

			if req.State == Done {
				break
			}
		} else {
			newBuf := make([]byte, len(buf)+bufferSize)
			copy(newBuf, buf)
			buf = newBuf
		}
	}

	return &req, nil

	// data, err := io.ReadAll(reader)
	// if err != nil {
	// 	return nil, err
	// }

	// tokens := strings.Split(string(data), "\r\n")

	// requestLine, err := parseRequestLine(tokens[0])
	// if err != nil {
	// 	return nil, err
	// }

	// return &Request{
	// 	RequestLine: *requestLine,
	// }, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case Initialized:
		read, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, nil
		}
		r.State = Done
		return read, nil
	case Done:
		return 0, errors.New("error: trying to read data in a done state")
	}

	return 0, errors.New("error: unknown state")
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, nil
	}
	fmt.Println(string(data))

	parts := strings.Split(string(data[:idx]), " ")

	if len(parts) != 3 {
		return 0, errors.New("malformed request-line")
	}

	method := parts[0]

	for i := range method {
		if method[i] < 65 || method[i] > 90 {
			return 0, errors.New("not alphabetic capital only characters")
		}
	}

	version := strings.Split(parts[2], "/")

	if len(version) != 2 || version[0] != "HTTP" {
		return 0, errors.New("bad http version")
	}

	if strings.Compare(version[1], "1.1") != 0 {
		return 0, errors.New("unsupported http version")
	}

	r.RequestLine = RequestLine{
		Method:        string(method),
		RequestTarget: parts[1],
		HttpVersion:   version[1],
	}

	return idx, nil
}
