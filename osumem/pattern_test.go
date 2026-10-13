package osumem

import (
	"bytes"
	"errors"
	"testing"
)

// sliceReader returns a readFunc backed by data whose first byte is at address
// 0, plus a counter of how many reads were performed.
func sliceReader(data []byte) (readFunc, *int) {
	calls := new(int)
	return func(addr uintptr, size int) ([]byte, error) {
		*calls++
		off := int(addr)
		if size <= 0 || off < 0 || off+size > len(data) {
			return nil, errors.New("test: out of range")
		}
		out := make([]byte, size)
		copy(out, data[off:off+size])
		return out, nil
	}, calls
}

// withChunkSize shrinks scanChunkSize for the duration of a test so chunk
// boundaries are exercised without large allocations.
func withChunkSize(t *testing.T, n int) {
	t.Helper()
	old := scanChunkSize
	scanChunkSize = n
	t.Cleanup(func() { scanChunkSize = old })
}

func TestScanRegionFuncStopsAfterAcceptedMatch(t *testing.T) {
	withChunkSize(t, 16)

	data := bytes.Repeat([]byte{0xAA}, 64)
	sig := mustSignature("DE AD BE EF")
	copy(data[4:], []byte{0xDE, 0xAD, 0xBE, 0xEF})
	copy(data[40:], []byte{0xDE, 0xAD, 0xBE, 0xEF})

	read, calls := sliceReader(data)
	var accepted []uintptr
	got, ok := scanRegionFunc(read, 0, uintptr(len(data)), sig, func(addr uintptr) bool {
		accepted = append(accepted, addr)
		return addr == 40
	})

	if !ok || got != 40 {
		t.Fatalf("scanRegionFunc = (%d, %v), want (40, true)", got, ok)
	}
	if len(accepted) != 2 || accepted[0] != 4 || accepted[1] != 40 {
		t.Fatalf("candidates = %v, want [4 40]", accepted)
	}
	// Chunks start at 0,13,26,39 before the match at 40 is found; the final
	// chunk (52) must never be read.
	if *calls != 4 {
		t.Fatalf("reads = %d, want 4 (early exit)", *calls)
	}
}

func TestScanRegionFuncRejectsAll(t *testing.T) {
	withChunkSize(t, 16)

	data := bytes.Repeat([]byte{0xAA}, 64)
	sig := mustSignature("DE AD BE EF")

	read, calls := sliceReader(data)
	got, ok := scanRegionFunc(read, 0, uintptr(len(data)), sig, func(uintptr) bool { return false })

	if ok || got != 0 {
		t.Fatalf("scanRegionFunc = (%d, %v), want (0, false)", got, ok)
	}
	if *calls != 5 {
		t.Fatalf("reads = %d, want 5 (whole region)", *calls)
	}
}

func TestScanRegionFuncNoDuplicateAcrossOverlap(t *testing.T) {
	withChunkSize(t, 16)

	data := bytes.Repeat([]byte{0xAA}, 64)
	sig := mustSignature("DE AD BE EF")
	// Byte 14 lands in the first chunk's 3-byte overlap tail.
	copy(data[14:], []byte{0xDE, 0xAD, 0xBE, 0xEF})

	read, _ := sliceReader(data)
	var accepted []uintptr
	scanRegionFunc(read, 0, uintptr(len(data)), sig, func(addr uintptr) bool {
		accepted = append(accepted, addr)
		return false
	})

	if len(accepted) != 1 || accepted[0] != 14 {
		t.Fatalf("candidates = %v, want [14]", accepted)
	}
}

func TestScanRegionFuncWildcard(t *testing.T) {
	data := []byte{0x00, 0x11, 0x22, 0x33, 0x44, 0x55}
	sig := mustSignature("11 ?? 33")

	read, _ := sliceReader(data)
	got, ok := scanRegionFunc(read, 0, uintptr(len(data)), sig, func(uintptr) bool { return true })
	if !ok || got != 1 {
		t.Fatalf("scanRegionFunc = (%d, %v), want (1, true)", got, ok)
	}
}
