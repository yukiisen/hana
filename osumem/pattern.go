package osumem

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

// This file holds the OS-independent half of memory scanning: signature
// parsing and byte matching. The actual process access (ReadProcessMemory /
// process_vm_readv) is injected as a readFunc so this logic can be unit tested
// on any platform without a live osu! process.

// readFunc reads size bytes from the target address space.
type readFunc func(addr uintptr, size int) ([]byte, error)

// signature is a parsed AOB pattern: literal bytes plus a per-byte mask where 1
// means "must match" and 0 means wildcard.
type signature struct {
	bytes []byte
	mask  []byte
}

// newSignature parses a space-separated pattern such as "F8 01 ?? 04 83 65".
func newSignature(pattern string) (*signature, error) {
	tokens := strings.Fields(pattern)
	if len(tokens) == 0 {
		return nil, errors.New("osumem: empty signature")
	}

	s := &signature{
		bytes: make([]byte, len(tokens)),
		mask:  make([]byte, len(tokens)),
	}
	for i, tok := range tokens {
		if tok == "??" || tok == "?" {
			continue
		}
		var v uint64
		if _, err := fmt.Sscanf(tok, "%x", &v); err != nil {
			return nil, fmt.Errorf("osumem: bad signature byte %q: %w", tok, err)
		}
		s.bytes[i] = byte(v)
		s.mask[i] = 1
	}
	return s, nil
}

// literalRun returns the offset and length of the longest run of literal
// (non-wildcard) bytes in the signature. findFrom uses it as a needle for
// bytes.Index, whose SIMD implementation is dramatically faster than comparing
// the full pattern at every offset.
func (s *signature) literalRun() (off, length int) {
	off = -1
	for i := 0; i < len(s.mask); {
		if s.mask[i] == 0 {
			i++
			continue
		}
		j := i
		for j < len(s.mask) && s.mask[j] == 1 {
			j++
		}
		if j-i > length {
			off, length = i, j-i
		}
		i = j
	}
	return off, length
}

// findFrom returns the index of the first match at or after start, or -1.
func (s *signature) findFrom(data []byte, start int) int {
	n := len(s.bytes)
	if n == 0 || start < 0 {
		return -1
	}

	off, runLen := s.literalRun()
	if off < 0 {
		// All-wildcard signature: any position (within bounds) matches.
		if start <= len(data)-n {
			return start
		}
		return -1
	}

	needle := s.bytes[off : off+runLen]
	from := start
	for {
		pos := bytes.Index(data[from:], needle)
		if pos < 0 {
			return -1
		}
		pos += from

		// Candidate pattern start; verify only if it fits and is >= start.
		i := pos - off
		if i >= start && i+n <= len(data) {
			match := true
			for j := 0; j < n; j++ {
				if s.mask[j] == 1 && data[i+j] != s.bytes[j] {
					match = false
					break
				}
			}
			if match {
				return i
			}
		}

		from = pos + 1
		if from > len(data)-runLen {
			return -1
		}
	}
}

// scanChunkSize bounds how much of a region is read at once so huge mappings do
// not exhaust memory. It is a var (not const) so tests can shrink it to
// exercise the chunk-boundary overlap logic cheaply.
var scanChunkSize = 8 << 20 // 8 MiB

// scanRegionFirst returns the address of the first signature match in
// [base, base+size). Chunks overlap by signature-length-1 bytes so a match that
// straddles a chunk boundary is still found.
func scanRegionFirst(read readFunc, base, size uintptr, sig *signature) (uintptr, bool) {
	overlap := uintptr(len(sig.bytes) - 1)

	var offset uintptr
	for offset < size {
		readLen := size - offset
		if readLen > uintptr(scanChunkSize) {
			readLen = uintptr(scanChunkSize)
		}

		if data, err := read(base+offset, int(readLen)); err == nil {
			if idx := sig.findFrom(data, 0); idx >= 0 {
				return base + offset + uintptr(idx), true
			}
		}

		if readLen < uintptr(scanChunkSize) || readLen <= overlap {
			break
		}
		offset += readLen - overlap
	}
	return 0, false
}

// scanRegionFunc walks the signature matches in [base, base+size) in address
// order, calling fn for each and stopping as soon as fn accepts one. It returns
// that address and true, or 0, false when no candidate is accepted. Matches
// starting in a non-final chunk's overlap tail are deferred to the next chunk so
// no match is visited twice.
func scanRegionFunc(read readFunc, base, size uintptr, sig *signature, fn func(addr uintptr) bool) (uintptr, bool) {
	overlap := uintptr(len(sig.bytes) - 1)

	var offset uintptr
	for offset < size {
		readLen := size - offset
		last := readLen <= uintptr(scanChunkSize)
		if readLen > uintptr(scanChunkSize) {
			readLen = uintptr(scanChunkSize)
		}

		if data, err := read(base+offset, int(readLen)); err == nil {
			limit := len(data)
			if !last {
				limit = int(readLen - overlap)
			}

			for i := 0; ; {
				idx := sig.findFrom(data, i)
				if idx < 0 || idx >= limit {
					break
				}
				addr := base + offset + uintptr(idx)
				if fn(addr) {
					return addr, true
				}
				i = idx + 1
			}
		}

		if last || readLen <= overlap {
			break
		}
		offset += readLen - overlap
	}
	return 0, false
}
