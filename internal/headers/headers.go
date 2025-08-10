package headers

import (
	"bytes"
	"errors"
	"strings"
)

const allowedCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789~#$%^'*+-.^_`|~"

type Headers map[string]string

func NewHeaders() Headers {
	return Headers{}
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false
	err = nil

	newLineIdx := bytes.Index(data, []byte("\r\n"))

	switch newLineIdx {
	case -1:
		return
	case 0:
		n = 2
		done = true
		return
	}

	trimmed := bytes.Trim(data[:newLineIdx], " ")
	parts := strings.Split(string(trimmed), ":")

	if len(parts) < 2 {
		return 0, false, errors.New("malformed header: correct way: field-name: field-line-value")
	}

	if strings.Contains(parts[0], " ") {
		return 0, false, errors.New("space not allowed in header")
	} else if len(parts[0]) == 0 {
		return 0, false, errors.New("empty header name")
	}

	for _, b := range parts[0] {
		if !strings.Contains(allowedCharset, string(b)) {
			return 0, false, errors.New("invalid header")
		}
	}

	value := strings.Join(parts[1:], ":")
	value = strings.TrimPrefix(value, " ")

	if value[0] == ' ' {
		return 0, false, errors.New("only 1 OWS allowed")
	}

	key := strings.ToLower(string(parts[0]))
	value = strings.Trim(value, " ")

	if v, ok := h[key]; ok {
		h[key] = v + ", " + value
	} else {
		h[key] = value
	}

	return len(data[:newLineIdx]) + 2, false, nil
}

func (h Headers) Get(key string) (string, bool) {
	value, ok := h[strings.ToLower(key)]
	return value, ok
}
