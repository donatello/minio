/*
 * Minio Cloud Storage, (C) 2016 Minio, Inc.
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

package cmd

import (
	"io"

	cb "github.com/minio/minio/pkg/chunkedbuffers"
	"github.com/minio/minio/pkg/errors"
)

type errIdx struct {
	idx int
	err error
}

func (s ErasureStorage) readConcurrent(volume, path string, offset, length int64,
	verifiers []*BitrotVerifier) (buffers []*cb.ChunkedBuffer, err error) {

	requiredBlocks := s.dataBlocks
	errChan := make(chan errIdx)
	errs := make([]error, len(s.disks))
	finished := make([]bool, len(s.disks))

	drainErrors := func(count int, buffers []*cb.ChunkedBuffer) {
		for i := 0; i < count; i++ {
			_ = <-errChan
		}
		for _, buffer := range buffers {
			cb.FreeChunkedBuffer(buffer)
		}
	}

	stageBuffers := make([]*cb.ChunkedBuffer, len(s.disks))
	for i := range stageBuffers {
		stageBuffers[i] = cb.NewChunkedBuffer(length)
	}
	buffers = make([]*cb.ChunkedBuffer, len(s.disks))

	var launchIndex int
	var finishedCount, successCount int

outerLoop:
	for launchIndex == 0 || finishedCount < launchIndex {
		numLaunched := 0
		for ; launchIndex < requiredBlocks; launchIndex++ {
			go readFromFile(s.disks[launchIndex], volume, path,
				offset, stageBuffers[launchIndex],
				verifiers[launchIndex], launchIndex, errChan)
			numLaunched++
		}
		select {
		case errVal := <-errChan:
			finishedCount++
			finished[errVal.idx] = true
			errs[errVal.idx] = errVal.err
			if errVal.err != nil {
				// at least one disk failed to return
				// data, so we request all remaining
				// disks, in the next iteration. Also,
				// note that entering this block a
				// second time does not launch new go
				// routines.
				requiredBlocks = s.dataBlocks + s.parityBlocks
			} else {
				successCount++
				buffers[errVal.idx] = stageBuffers[errVal.idx]
				stageBuffers[errVal.idx] = nil
				if successCount == s.dataBlocks {
					// got enough data, so drain
					// the errChan and return the
					// success buffers
					go drainErrors(numLaunched-finishedCount, stageBuffers)
					break outerLoop
				}
			}
		}
	}
	return
}

func readFromFile(disk StorageAPI, volume, path string, offset int64,
	buffer *cb.ChunkedBuffer, verifier *BitrotVerifier, i int, errChan chan<- errIdx) {

	if disk == OfflineDisk {
		errChan <- errIdx{i, errors.Trace(errDiskNotFound)}
		return
	}
	_, err := disk.ReadFile(volume, path, offset, buffer, verifier)
	errChan <- errIdx{i, err}
}

// ReadFile reads as much data as requested from the file under the given volume and path and writes the data to the provided writer.
// The algorithm and the keys/checksums are used to verify the integrity of the given file. ReadFile will read data from the given offset
// up to the given length. If parts of the file are corrupted ReadFile tries to reconstruct the data.
func (s ErasureStorage) ReadFile(writer io.Writer, volume, path string, offset, length int64, totalLength int64, checksums [][]byte, algorithm BitrotAlgorithm, blocksize int64) (f ErasureFileInfo, err error) {
	if offset < 0 || length < 0 {
		return f, errors.Trace(errUnexpected)
	}
	if offset+length > totalLength {
		return f, errors.Trace(errUnexpected)
	}
	if !algorithm.Available() {
		return f, errors.Trace(errBitrotHashAlgoInvalid)
	}

	f.Checksums = make([][]byte, len(s.disks))
	verifiers := make([]*BitrotVerifier, len(s.disks))
	for i, disk := range s.disks {
		if disk == OfflineDisk {
			continue
		}
		verifiers[i] = NewBitrotVerifier(algorithm, checksums[i])
	}

	chunksize := getChunkSize(blocksize, s.dataBlocks)

	// We read all whole-blocks of erasure coded data containing
	// the requested data range.
	//
	// The start index of the erasure coded block containing the
	// `offset` byte of data is:
	partDataStartIndex := (offset / blocksize) * chunksize
	// The start index of the erasure coded block containing the
	// (last) byte of data at the index `offset + length - 1` is:
	blockStartIndex := ((offset + length - 1) / blocksize) * chunksize
	// However, we need the end index of the e.c. block containing
	// the last byte - we need to check if that block is the last
	// block in the part (in that case, it may be have a different
	// chunk size)
	isLastBlock := (totalLength-1)/blocksize == (offset+length-1)/blocksize
	var partDataEndIndex int64
	if isLastBlock {
		lastBlockChunkSize := getChunkSize(totalLength%blocksize, s.dataBlocks)
		partDataEndIndex = blockStartIndex + lastBlockChunkSize - 1
	} else {
		partDataEndIndex = blockStartIndex + chunksize - 1
	}

	// Thus, the length of data to be read from the part file(s) is:
	partDataLength := partDataEndIndex - partDataStartIndex + 1

	var buffers []*cb.ChunkedBuffer
	buffers, err = s.readConcurrent(volume, path, partDataStartIndex, partDataLength, verifiers)
	if err != nil {
		return
	}

	blocks := make([][]byte, len(s.disks))
	needsReconstruction := false
	availableCount := 0
	for i := range blocks {
		if buffers[i] != nil {
			blocks[i] = make([]byte, chunksize)
			availableCount++
		} else {
			blocks[i] = make([]byte, 0, chunksize)
			if i < s.dataBlocks {
				needsReconstruction = true
			}
		}
	}
	if availableCount < s.dataBlocks {
		return f, errors.Trace(errXLReadQuorum)
	}

	numChunks := ceilFrac(partDataLength, chunksize)
	copyLen := chunksize
	for chunkNumber := int64(0); chunkNumber < numChunks; chunkNumber++ {
		if chunkNumber == numChunks-1 {
			copyLen = partDataLength % chunksize
			for i := range blocks {
				if buffers[i] != nil {
					blocks[i] = make([]byte, copyLen)
				} else {
					blocks[i] = make([]byte, 0, copyLen)
				}
			}
		} else {
			for i := range blocks {
				if buffers[i] == nil {
					blocks[i] = blocks[i][0:0]
				}
			}
		}

		for i := range blocks {
			if buffers[i] != nil {
				_, err = io.ReadFull(buffers[i], blocks[i])
				if err != nil {
					return f, errors.Trace(err)
				}
			}
		}

		if needsReconstruction {
			if err = s.ErasureDecodeDataBlocks(blocks); err != nil {
				return f, errors.Trace(err)
			}
		}

		var writeStart, writeLength int64
		if chunkNumber == 0 {
			writeStart = offset % blocksize
		}
		writeLength = blocksize - writeStart
		if chunkNumber == numChunks-1 {
			writeLength = copyLen - writeStart
		}
		n, err := writeDataBlocks(writer, blocks, s.dataBlocks, writeStart, writeLength)
		if err != nil {
			return f, err
		}

		f.Size += n
	}

	f.Algorithm = algorithm
	for i, disk := range s.disks {
		if disk == OfflineDisk {
			continue
		}
		f.Checksums[i] = verifiers[i].Sum(nil)
	}
	return f, nil
}

func erasureCountMissingBlocks(blocks [][]byte, limit int) int {
	missing := 0
	for i := range blocks[:limit] {
		if len(blocks[i]) == 0 {
			missing++
		}
	}
	return missing
}
