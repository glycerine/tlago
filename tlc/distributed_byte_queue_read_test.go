package tlc

import (
	"bytes"
	"encoding/binary"
	"os"
	"testing"
)

// DiskByteArrayQueue uses BufferedDataInputStream.read, not readFully. It
// publishes the allocated slot before reading and ignores a short EOF count.
// No original queue test covers these raw byte-array recovery boundaries.
func TestDistributedBytePoolReadRetainsSourcePartialSlots(t *testing.T) {
	t.Run("negative_length", func(t *testing.T) {
		name := t.TempDir() + string(os.PathSeparator) + "0"
		if err := os.WriteFile(name, []byte{255, 255, 255, 255}, 0600); err != nil {
			t.Fatal(err)
		}
		destination := [][]byte{{9}}
		defer func() {
			if _, ok := recover().(*NegativeArraySizeException); !ok || !bytes.Equal(destination[0], []byte{9}) {
				t.Fatal("negative length lost allocation failure or replaced the old slot")
			}
		}()
		_ = readByteArrayPoolFile(name, destination)
	})
	for _, count := range []int{1, 2} {
		t.Run(intToDecimal(count), func(t *testing.T) {
			name := t.TempDir() + string(os.PathSeparator) + "0"
			data := binary.BigEndian.AppendUint32(nil, 4)
			data = append(data, 1, 2)
			if err := os.WriteFile(name, data, 0600); err != nil {
				t.Fatal(err)
			}
			original := []byte{9}
			destination := [][]byte{original}
			if count == 2 {
				destination = append(destination, original)
			}
			reader := NewByteArrayPoolReader(count, name)
			result, err := reader.DoWork(destination, name+"next")
			if !bytes.Equal(destination[0], []byte{1, 2, 0, 0}) {
				t.Fatalf("read lost allocated/partially filled slot: %v", destination)
			}
			if count == 1 {
				if err != nil || len(result) != 1 || &result[0] != &destination[0] || reader.poolFile != name+"next" || !reader.canRead {
					t.Fatalf("source short final read did not complete: %v/%v", result, err)
				}
			} else if !isJavaIOException(err) || result != nil || &destination[1][0] != &original[0] || reader.poolFile != name || reader.canRead || reader.isFull {
				t.Fatalf("missing next length changed untouched slots or pending work: %v/%v", result, err)
			}
		})
	}
}

func TestDistributedByteQueueRecoveryRetainsInactiveAndPartialSlots(t *testing.T) {
	for _, phase := range []string{"short_final_entry", "missing_final_length"} {
		t.Run(phase, func(t *testing.T) {
			q := byteQueuePathFixture(t.TempDir())
			original := []byte{9}
			q.enqBuf = [][]byte{original, original}
			q.deqBuf = [][]byte{original, original}
			oldFile, oldPending := q.loFile, q.reader.poolFile
			data := binary.BigEndian.AppendUint64(nil, 2)
			for _, value := range []uint32{5, 0, 1, 1} {
				data = binary.BigEndian.AppendUint32(data, value)
			}
			data = binary.BigEndian.AppendUint32(data, 2)
			data = append(data, 7, 8)
			if phase == "short_final_entry" {
				data = binary.BigEndian.AppendUint32(data, 4)
				data = append(data, 1, 2)
			} else {
				data = append(data, 0, 0, 0)
			}
			if err := os.WriteFile(q.queuePath("queue.chkpt"), data, 0600); err != nil {
				t.Fatal(err)
			}
			err := q.Recover()
			if q.len != 2 || q.loPool != 5 || q.hiPool != 0 || q.enqIndex != 1 || q.deqIndex != 1 || q.lastLoPool != 4 || !bytes.Equal(q.enqBuf[0], []byte{7, 8}) {
				t.Fatal("recovery discarded completed header/prefix mutations")
			}
			if !bytes.Equal(q.enqBuf[1], original) || &q.enqBuf[1][0] != &original[0] || &q.deqBuf[0][0] != &original[0] {
				t.Fatal("recovery cleared inactive source buffer slots")
			}
			if phase == "short_final_entry" {
				if err != nil || !bytes.Equal(q.deqBuf[1], []byte{1, 2, 0, 0}) || q.loFile != q.poolName(5) || q.reader.poolFile != q.poolName(4) || q.reader.canRead {
					t.Fatalf("short final entry did not complete source recovery: %v", err)
				}
			} else if !isJavaIOException(err) || &q.deqBuf[1][0] != &original[0] || q.loFile != oldFile || q.reader.poolFile != oldPending || q.reader.canRead {
				t.Fatalf("failed length read changed unread slot or restarted reader: %v", err)
			}
		})
	}
}
