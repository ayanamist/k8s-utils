package protobuf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

type StreamBuffer struct {
	idx int
	len int
	r   io.Reader
}

func NewStreamBuffer(r io.Reader, len int) *StreamBuffer {
	return &StreamBuffer{
		r:   r,
		len: len,
	}
}

func (s *StreamBuffer) Len() int {
	return s.len
}

func (s *StreamBuffer) Get(i int) (byte, error) {
	if i < s.idx {
		return 0, fmt.Errorf("invalid index %d < %d", i, s.idx)
	}
	needRead := i - s.idx + 1
	buf := make([]byte, needRead)
	if _, err := io.ReadFull(s.r, buf); err != nil {
		return 0, err
	}
	s.idx = i + 1
	return buf[needRead-1], nil
}

func (s *StreamBuffer) Slice(start, end int) ([]byte, error) {
	if start < s.idx {
		return nil, fmt.Errorf("invalid index %d < %d", start, s.idx)
	}
	if start == end {
		return []byte{}, nil
	}
	newIdx := end
	start -= s.idx
	end -= s.idx
	buf := make([]byte, end)
	if _, err := io.ReadFull(s.r, buf); err != nil {
		return nil, err
	}
	s.idx = newIdx
	return buf[start:end], nil
}

func (s *StreamBuffer) SubStream(start, end int) (*StreamBuffer, error) {
	if start < s.idx {
		return nil, fmt.Errorf("invalid index %d < %d", start, s.idx)
	}
	if start == end {
		return NewStreamBuffer(bytes.NewReader(nil), 0), nil
	}
	newIdx := end
	start -= s.idx
	end -= s.idx
	s.idx = newIdx
	if start > 0 {
		if _, err := io.Copy(ioutil.Discard, io.LimitReader(s.r, int64(start))); err != nil {
			return nil, err
		}
	}
	return NewStreamBuffer(io.LimitReader(s.r, int64(end-start)), end-start), nil
}

func (s *StreamBuffer) Discard() error {
	_, err := io.Copy(ioutil.Discard, s.r)
	return err
}
