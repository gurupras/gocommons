package seekable_stream

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type SeekableStream struct {
	bytes []byte
	*bytes.Reader
}

func (ss *SeekableStream) WrapBytes(b []byte) {
	ss.bytes = b
	ss.Reader = bytes.NewReader(ss.bytes)
}

func (ss *SeekableStream) WrapReader(reader io.Reader) (err error) {
	buf := new(bytes.Buffer)
	if _, err = io.Copy(buf, reader); err != nil {
		err = errors.New(fmt.Sprintf("Failed to wrap reader: %v", err))
		return
	}
	ss.bytes = make([]byte, buf.Len())
	copy(ss.bytes, buf.Bytes())
	ss.Reader = bytes.NewReader(ss.bytes)
	return
}

func (ss *SeekableStream) Rewind() {
	ss.Reader = bytes.NewReader(ss.bytes)
}

func (ss *SeekableStream) String() string {
	return string(ss.bytes)
}

func (ss *SeekableStream) Len() int {
	return len(ss.bytes)
}
