package streamlister

import (
	"bytes"
	"fmt"
	"io"
)

type streamBuffer struct {
	idx int
	len int
	r   io.Reader
}

func newStreamBuffer(r io.Reader, len int) *streamBuffer {
	return &streamBuffer{
		r:   r,
		len: len,
	}
}

func (s *streamBuffer) Len() int {
	return s.len
}

func (s *streamBuffer) Get(i int) (byte, error) {
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

func (s *streamBuffer) Slice(start, end int) ([]byte, error) {
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

func (s *streamBuffer) SubStream(start, end int) (*streamBuffer, error) {
	if start < s.idx {
		return nil, fmt.Errorf("invalid index %d < %d", start, s.idx)
	}
	if start == end {
		return newStreamBuffer(bytes.NewReader(nil), 0), nil
	}
	newIdx := end
	start -= s.idx
	end -= s.idx
	s.idx = newIdx
	if start > 0 {
		if _, err := io.Copy(io.Discard, io.LimitReader(s.r, int64(start))); err != nil {
			return nil, err
		}
	}
	return newStreamBuffer(io.LimitReader(s.r, int64(end-start)), end-start), nil
}

func (s *streamBuffer) Discard() error {
	_, err := io.Copy(io.Discard, s.r)
	return err
}
