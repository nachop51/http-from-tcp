package request

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

const bufferSize int = 8
const (
	requestInit = iota
	requestParsingHeaders
	requestParsingBody
	requestStateDone
)

type Request struct {
	RequestLine RequestLine
	Headers     headers.Headers
	Body        []byte
	state       int
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
		state: requestInit,
		Body:  nil,
	}

	for req.state != requestStateDone {
		read, err := reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if req.state == requestParsingBody {
					if strconv.Itoa(len(req.Body)) != req.Headers["content-length"] {
						return nil, errors.New("unexpected EOF while reading body")
					}
				}
				req.state = requestStateDone
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

			if req.state == requestStateDone {
				break
			}
		} else {
			newBuf := make([]byte, len(buf)+bufferSize)
			copy(newBuf, buf)
			buf = newBuf
		}
	}

	return &req, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.state != requestStateDone {
		n, err := r.parseNext(data[totalBytesParsed:])
		totalBytesParsed += n

		if err != nil {
			return 0, err
		}
		if n == 0 {
			if totalBytesParsed != 0 {
				return totalBytesParsed, nil
			}
			return 0, nil
		}
	}
	return totalBytesParsed, nil
}

func (r *Request) parseNext(data []byte) (int, error) {
	switch r.state {
	case requestInit:
		read, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, nil
		}
		r.state = requestParsingHeaders

		return read, nil
	case requestParsingHeaders:
		read, done, err := r.parseHeaders(data)
		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, nil
		}
		if done {
			r.state = requestParsingBody
		}
		return read, nil
	case requestParsingBody:
		read, done, err := r.parseBody(data)
		if done {
			r.state = requestStateDone
			return read, nil
		}

		if err != nil {
			return 0, err
		}
		if read == 0 {
			return 0, nil
		}

		return read, nil
	case requestStateDone:
		return 0, errors.New("error: trying to read data in a done state")
	}

	return 0, errors.New("error: unknown state")
}

func (r *Request) parseBody(data []byte) (int, bool, error) {
	contentLengthStr, ok := r.Headers.Get("Content-Length")
	if !ok {
		return 0, true, nil
	}

	contentLength, err := strconv.Atoi(contentLengthStr)
	if err != nil || contentLength < 0 {
		return 0, false, errors.New("invalid content-length header NaN or negative")
	}

	if r.Body == nil {
		r.Body = make([]byte, 0)
	}

	r.Body = append(r.Body, data...)

	if len(r.Body) > contentLength {
		return 0, false, errors.New("body is greater than the content-length specified")
	}

	if len(r.Body) == contentLength {
		return len(data), true, nil
	}

	return len(data), false, nil
}

func (r *Request) parseHeaders(data []byte) (int, bool, error) {
	if r.Headers == nil {
		r.Headers = headers.NewHeaders()
	}

	n, done, err := r.Headers.Parse(data)
	if err != nil {
		return 0, false, err
	}
	if n == 0 {
		return 0, false, nil
	}

	if done {
		return n, true, nil
	}

	return n, false, nil
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	idx := bytes.Index(data, []byte("\r\n"))
	if idx == -1 {
		return 0, nil
	}

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

	return idx + 2, nil
}
