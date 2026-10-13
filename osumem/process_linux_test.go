//go:build linux

package osumem

import (
	"encoding/binary"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode/utf16"
	"unsafe"
)

// selfProcess attaches to the test binary itself via /proc/self/mem so the
// Linux backend can be exercised without a live osu! process.
func selfProcess(t *testing.T) *process {
	t.Helper()
	pid := uint32(os.Getpid())
	handle, err := openHandle(pid)
	if err != nil {
		t.Fatalf("openHandle(self): %v", err)
	}
	p := &process{handle: handle, pid: pid, bitness: 64}
	t.Cleanup(p.close)
	return p
}

func TestReadSelf(t *testing.T) {
	p := selfProcess(t)

	probe := int32(0x12345678)
	got, err := p.readInt32(uintptr(unsafe.Pointer(&probe)))
	if err != nil {
		t.Fatalf("readInt32: %v", err)
	}
	if got != probe {
		t.Fatalf("readInt32 = %#x, want %#x", got, probe)
	}
	runtime.KeepAlive(probe)
}

func TestReadSharpString(t *testing.T) {
	p := selfProcess(t)

	want := "osu! beatmap ビートマップ"
	units := utf16.Encode([]rune(want))

	buf := make([]byte, 12+len(units)*2)
	binary.LittleEndian.PutUint32(buf[8:], uint32(len(units)))
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[12+i*2:], u)
	}

	got, err := p.readSharpString(uintptr(unsafe.Pointer(&buf[0])))
	if err != nil {
		t.Fatalf("readSharpString: %v", err)
	}
	if got != want {
		t.Fatalf("readSharpString = %q, want %q", got, want)
	}
}

func TestFindProcesses(t *testing.T) {
	comm, err := os.ReadFile("/proc/self/comm")
	if err != nil {
		t.Fatalf("read comm: %v", err)
	}
	name := strings.TrimSpace(string(comm))

	found := false
	for _, pid := range findProcesses(name) {
		if pid == uint32(os.Getpid()) {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("findProcesses(%q) did not include self pid %d", name, os.Getpid())
	}
}

func TestIsProcess64Bit(t *testing.T) {
	got := isProcess64Bit(uint32(os.Getpid()))
	want := strconv.IntSize == 64
	if got != want {
		t.Fatalf("isProcess64Bit(self) = %v, want %v", got, want)
	}
}

func TestForEachRegion(t *testing.T) {
	p := selfProcess(t)

	probe := int32(1)
	target := uintptr(unsafe.Pointer(&probe))

	count := 0
	covered := false
	p.forEachRegion(func(base, size uintptr) bool {
		count++
		if target >= base && target < base+size {
			covered = true
		}
		return false
	})
	runtime.KeepAlive(probe)

	if count == 0 {
		t.Fatal("forEachRegion yielded no regions")
	}
	if !covered {
		t.Fatalf("forEachRegion did not cover probe address %#x", target)
	}
}
