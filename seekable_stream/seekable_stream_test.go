package seekable_stream

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type DummyReader int

func (dr *DummyReader) Read(b []byte) (n int, err error) {
	err = errors.New("Throwing error: Expected")
	return
}

func TestWrapBytes(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Try to wrap around []byte
	ss := new(SeekableStream)
	buf := new(bytes.Buffer)
	buf.WriteString("Hello, World!")
	ss.WrapBytes(buf.Bytes())
	output := ss.String()
	assert.Equal(buf.String(), output, "Strings don't match")
	// Do a rewind
	ss.Rewind()
	output = ss.String()
	assert.Equal(buf.String(), output, "Strings don't match")

}

func TestWrapBuffer(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Try to wrap around bytes.Buffer
	buf := new(bytes.Buffer)
	str := "Hello, World!-1"
	buf.WriteString(str)
	ss := new(SeekableStream)
	ss.WrapReader(buf)
	output := ss.String()
	assert.Equal(str, output, "Strings don't match")
	ss.Rewind()
	output = ss.String()
	assert.Equal(str, output, "Strings don't match")
}

func TestLen(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	// Try to wrap around bytes.Buffer
	buf := new(bytes.Buffer)
	str := "Hello, World!-len"
	buf.WriteString(str)
	ss := new(SeekableStream)
	ss.WrapReader(buf)
	assert.Equal(len(str), ss.Len(), "Length of strings don't match")
	ss.Rewind()
	assert.Equal(len(str), ss.Len(), "Length of Strings don't match")
}

func TestError(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	ss := new(SeekableStream)
	dr := new(DummyReader)
	err := ss.WrapReader(dr)
	assert.NotNil(err, "Expected error when using a reader that throws error while reading")
}
