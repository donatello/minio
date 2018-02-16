/*
 * Minio Cloud Storage, (C) 2018 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package chunkedbuffers implements buffers on top of fixed length
// (smaller) buffers to improve GC performance. It uses sync.Pool to
// improve reuse of allocated buffers.
package chunkedbuffers

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sync"
)

// 1 GiB is maximum chunk size
const maxChunkSize = 1024 * 1024 * 1024

// A ChunkedAllocator allocates ChunkedBuffers using multiple fixed
// length buffers.
type ChunkedAllocator struct {
	pool      sync.Pool
	chunkSize int64
}

// Singleton for the package
var chunkedAllocator *ChunkedAllocator

// InitChunkedAllocator initializes the ChunkedAllocator that creates
// ChunkedBuffers using fixed-length buffers of length
// `chunkSizeBytes`.
func InitChunkedAllocator(chunkSizeBytes int64) error {
	if chunkSizeBytes <= 0 || chunkSizeBytes > maxChunkSize {
		return fmt.Errorf("Invalid chunk size provided")
	}

	chunkedAllocator = &ChunkedAllocator{
		sync.Pool{New: func() interface{} {
			return make([]byte, chunkSizeBytes)
		}},
		chunkSizeBytes,
	}
	return nil
}

// Unexported function: pass b == 0 and crash!
func ceiling(a, b int64) (n int64) {
	n = a / b
	if a%b != 0 {
		n++
	}
	return
}

// NewChunkedBuffer initializes a ChunkedBuffer with the given
// capacity.
func NewChunkedBuffer(cap int64) *ChunkedBuffer {
	num := ceiling(cap, chunkedAllocator.chunkSize)

	bufs := make([][]byte, num)
	return &ChunkedBuffer{buffer: bufs, cap: cap}
}

// FreeChunkedBuffer frees the chunked buffer and returns the
// underlying buffers to the pool. It is not necessary to call this
// function if the chunked buffer was written to capacity and then
// read completely.
func FreeChunkedBuffer(b *ChunkedBuffer) {
	if b == nil {
		return
	}
	for i := range b.buffer {
		if len(b.buffer[i]) > 0 {
			chunkedAllocator.pool.Put(b.buffer[i])
		}
	}
}

// A ChunkedBuffer implements the Reader and Writer interfaces. The
// capacity of a ChunkedBuffer is fixed at initialization. Data may be
// written to the ChunkedBuffer upto its capacity, and may be read
// upto the number of bytes written to it. Writing to a ChunkedBuffer
// past its capacity or reading past the number of bytes written to
// the ChunkedBuffer returns `io.EOF`.
//
// When writing to a ChunkedBuffer, allocation is performed on demand
// from a sync.Pool. When reading from a ChunkedBuffer, when
// underlying buffer chunks have been read, they are eagerly returned
// to the sync.Pool.
type ChunkedBuffer struct {
	buffer [][]byte

	// byte counts: read, written and capacity
	rc, wc, cap int64
}

// GetCapacity returns the number of bytes that can be written to the
// chunked buffer.
func (b *ChunkedBuffer) GetCapacity() (n int64) {
	return b.cap - b.wc
}

func (b *ChunkedBuffer) Read(p []byte) (n int, err error) {
	// Read returns an error if we read past the last byte
	// previously written. In addition, it returns chunks to the
	// pool, immediately after reading them.

	chunkIx := b.rc / chunkedAllocator.chunkSize
	numChunks := ceiling(b.wc, chunkedAllocator.chunkSize)

	isCapFilled := b.wc == b.cap
	start := b.rc % chunkedAllocator.chunkSize
	end := chunkedAllocator.chunkSize
	for n < len(p) && chunkIx < numChunks && b.rc+int64(n) < b.wc {
		isLastChunk := chunkIx+1 == numChunks

		// Adjust end offset for last chunk to be read
		if isLastChunk {
			end = b.wc % chunkedAllocator.chunkSize
		}

		// Copy into provided buffer
		copiedLen := copy(p[n:], b.buffer[chunkIx][start:end])

		// If we finished reading the current chunk, return it
		// to the pool.
		fullChunkRead := end-start == int64(copiedLen)
		if fullChunkRead && (!isLastChunk || isCapFilled) {
			chunkedAllocator.pool.Put(b.buffer[chunkIx])
			b.buffer[chunkIx] = nil
		}

		start = 0
		n += copiedLen
		chunkIx++
	}

	// Update read count
	b.rc += int64(n)

	if n < len(p) {
		err = io.EOF
	}
	return
}

func (b *ChunkedBuffer) Write(p []byte) (n int, err error) {
	// Write returns an error if we write more than the declared
	// capacity. In addition, it allocates chunks just before
	// writing to them.

	chunkIx := b.wc / chunkedAllocator.chunkSize
	numChunks := ceiling(b.cap, chunkedAllocator.chunkSize)

	start := b.wc % chunkedAllocator.chunkSize
	end := chunkedAllocator.chunkSize
	for n < len(p) && chunkIx < numChunks {
		isLastChunk := chunkIx+1 == numChunks
		// Adjust end offset for last chunk to be written
		if isLastChunk {
			end = b.cap % chunkedAllocator.chunkSize
		}

		// Allocate current chunk if it is not allocated
		if len(b.buffer[chunkIx]) == 0 {
			b.buffer[chunkIx], _ = chunkedAllocator.pool.Get().([]byte)
		}

		copiedLen := copy(b.buffer[chunkIx][start:end], p[n:])

		start = 0
		n += copiedLen
		chunkIx++
	}

	// Update write count
	b.wc += int64(n)

	if n < len(p) {
		err = io.EOF
	}
	return
}

// GobEncode implements gob.GobEncoder
func (b *ChunkedBuffer) GobEncode() (buf []byte, err error) {
	numBuffer := make([]byte, 2*binary.MaxVarintLen64)
	fCount := binary.PutVarint(numBuffer, b.cap)
	sCount := binary.PutVarint(numBuffer[fCount:], b.wc)
	s := int64(fCount) + int64(sCount)

	buf = make([]byte, s+b.wc)
	copy(buf, numBuffer[:s])

	// Copy buffers into buf, without updating ChunkedBuffer
	// internal state.
	for i, v := range b.buffer {
		if i < len(b.buffer)-1 {
			_ = copy(buf[s:], v)
			s += chunkedAllocator.chunkSize
		} else {
			_ = copy(buf[s:], v[:b.wc%chunkedAllocator.chunkSize])
		}
	}
	return
}

// GobDecode implements gob.GobDecoder
func (b *ChunkedBuffer) GobDecode(buf []byte) (err error) {
	// Read requested data-length from `buf`
	r := bytes.NewReader(buf)
	var capacity, writtenCount int64
	capacity, err = binary.ReadVarint(r)
	if err != nil {
		return
	}
	writtenCount, err = binary.ReadVarint(r)
	if err != nil {
		return
	}

	// Drop all existing buffers and initialize receiver
	FreeChunkedBuffer(b)
	num := ceiling(capacity, chunkedAllocator.chunkSize)
	b.buffer = make([][]byte, num)
	b.cap = capacity
	b.wc = 0

	// Copy data from `buf` into the receiver.
	_, err = io.CopyN(b, r, writtenCount)
	return
}
