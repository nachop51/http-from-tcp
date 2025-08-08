package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("    Host: localhost:42069    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 31, n)
	assert.False(t, done)

	// Valid 2 headers with existing headers
	headers = NewHeaders()
	data = []byte("    Host: localhost:42069    \r\n    User-Agent: test-agent    \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 31, n)
	assert.False(t, done)

	n, done, err = headers.Parse(data[n:])
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "test-agent", headers["user-agent"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Empty line (end of headers)
	headers = NewHeaders()
	data = []byte("\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)

	// Test: No CRLF found (incomplete data)
	headers = NewHeaders()
	data = []byte("Host: localhost:42069")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid header without colon
	headers = NewHeaders()
	data = []byte("InvalidHeader\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with space in name
	headers = NewHeaders()
	data = []byte("Host Name: localhost:42069\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, "space not allowed in header", err.Error())
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Header with multiple colons in value
	headers = NewHeaders()
	data = []byte("Authorization: Bearer: token:123\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "Bearer: token:123", headers["authorization"])
	assert.Equal(t, 34, n)
	assert.False(t, done)

	// Test: Header with only one space after colon
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["content-type"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test: Header with no space after colon
	headers = NewHeaders()
	data = []byte("Accept:text/html\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "text/html", headers["accept"])
	assert.Equal(t, 18, n)
	assert.False(t, done)

	// Test: Invalid headers
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, "invalid header", err.Error())
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: multiple same name header
	headers = NewHeaders()
	headers["host"] = "localhost:42069"
	data = []byte("Host: localhost:42070\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069, localhost:42070", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: multiple same name header parsing
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42070\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	prevN, done, err := headers.Parse(data[n:])
	require.NoError(t, err)
	assert.Equal(t, "localhost:42069, localhost:42070", headers["host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)
	// test done after it
	n, done, err = headers.Parse(data[prevN+n:])
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.True(t, done)
}
