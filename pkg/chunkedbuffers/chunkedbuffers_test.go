package chunkedbuffers

import (
	"io"
	"testing"
)

func TestChunkedBuffer1(t *testing.T) {
	InitChunkedAllocator(10)

	s := "Hello World!"
	cbuf := NewChunkedBuffer(int64(len(s)))
	n, err := cbuf.Write([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected: %v\n", err)
	}
	if n != len(s) {
		t.Fatalf("Bad write - %d\n", n)
	}

	buf := make([]byte, len(s))
	n, err = cbuf.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected: %v\n", err)
	}
	if n != 12 {
		t.Fatalf("Bad write - %d\n", n)
	}

	if string(buf) != s {
		t.Fatalf("Got unexpected data: %s", string(buf))
	}
}

func TestChunkedBufferExtraWrite1(t *testing.T) {
	InitChunkedAllocator(10)

	s := "Hello World!"
	cbuf := NewChunkedBuffer(int64(len(s)))
	n, err := cbuf.Write([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected: %v\n", err)
	}
	if n != len(s) {
		t.Fatalf("Bad write - %d\n", n)
	}

	_, err = cbuf.Write([]byte{1})
	if err != io.EOF {
		t.Fatalf("Unexpected: %v\n", err)
	}
}

func TestChunkedBufferExtraRead1(t *testing.T) {
	InitChunkedAllocator(10)

	s := "Hello World!"
	cbuf := NewChunkedBuffer(int64(len(s)))
	n, err := cbuf.Write([]byte(s))
	if err != nil {
		t.Fatalf("Unexpected: %v\n", err)
	}
	if n != len(s) {
		t.Fatalf("Bad write - %d\n", n)
	}

	buf := make([]byte, len(s))
	n, err = cbuf.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected: %v\n", err)
	}
	if n != 12 {
		t.Fatalf("Bad write - %d\n", n)
	}
	if string(buf) != s {
		t.Fatalf("Got unexpected data: %s", string(buf))
	}

	_, err = cbuf.Read(buf[0:1])
	if err != io.EOF {
		t.Fatalf("Unexpected: %v\n", err)
	}
}
