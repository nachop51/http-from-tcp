package response

import (
	"fmt"
	"io"

	"httpfromtcp/internal/headers"
)

type writerStatus int

const (
	writerInit writerStatus = iota
	writerHeaders
	writerBody
	writerDone
)

type Writer struct {
	io.Writer
	writerStatus writerStatus
}

type StatusCode int

const (
	StatusOK         StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternal   StatusCode = 500
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		Writer:       w,
		writerStatus: writerInit,
	}
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.writerStatus != writerInit {
		return fmt.Errorf("invalid writer status: %v", w.writerStatus)
	}

	var reason string
	switch statusCode {
	case StatusOK:
		reason = "OK"
	case StatusBadRequest:
		reason = "Bad Request"
	case StatusInternal:
		reason = "Internal Server Error"
	}
	_, err := io.WriteString(w, "HTTP/1.1 "+fmt.Sprint(int(statusCode))+" "+reason+"\r\n")

	if err == nil {
		w.writerStatus = writerHeaders
	}

	return err
}

func GetDefaultHeaders(contentLength int) headers.Headers {
	h := headers.NewHeaders()
	h["connection"] = "close"
	h["content-type"] = "text/html"
	h["content-length"] = fmt.Sprint(contentLength)

	return h
}

func (w *Writer) WriteHeaders(h headers.Headers) error {
	if w.writerStatus != writerHeaders {
		return fmt.Errorf("invalid writer status: %v", w.writerStatus)
	}

	for k, v := range h {
		_, err := io.WriteString(w, fmt.Sprintf("%s: %s\r\n", k, v))
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\r\n")

	if err == nil {
		w.writerStatus = writerBody
	}

	return err
}

func (w *Writer) WriteBody(body []byte) (int, error) {
	if w.writerStatus != writerBody {
		return 0, fmt.Errorf("invalid writer status: %v", w.writerStatus)
	}

	n, err := w.Write(body)

	if err == nil {
		w.writerStatus = writerDone
	}

	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	hexStr := fmt.Sprintf("%x\r\n", len(p))
	hex, err := w.Write([]byte(hexStr))
	if err != nil {
		return hex, err
	}
	body, err := w.Write(p)
	if err != nil {
		return body, err
	}
	_, err = fmt.Fprint(w, "\r\n")
	return hex + body + 2, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	writed, err := fmt.Fprint(w, "0\r\n")
	return writed, err
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	for k, v := range h {
		_, err := io.WriteString(w, fmt.Sprintf("%s: %s\r\n", k, v))
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\r\n")

	return err
}
