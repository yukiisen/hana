// Package osumem reads osu! process memory and exposes the currently selected
// beatmap's metadata.
//
// It is a Go port of the memory-reading core of tosu
// (https://github.com/tosuapp/tosu). tosu supports two very different clients:
//
//   - osu!stable: a 32-bit .NET Framework / XNA application. Its globals are
//     located with hard-coded AOB (array-of-bytes) signature scans, then read
//     through a graph of fixed pointer offsets. See stable.go.
//   - osu!lazer:  a 64-bit .NET 8 application. Its layout changes with every
//     release, so instead of fixed signatures it is located by scanning a
//     single stable anchor pattern and validating a vtable, then read through
//     versioned offsets shipped in offsets.json. See lazer.go.
//
// The two clients share all of the decoding logic (signature parsing, pointer
// walking and C# string decoding), which lives in platform-independent files.
// Only the operating-system primitives (process discovery, raw reads and
// region enumeration) differ, and those live in process_windows.go and
// process_linux.go.
package osumem

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"unicode/utf16"
)

// ---------------------------------------------------------------------------
// Public errors
// ---------------------------------------------------------------------------

var (
	// ErrProcessNotFound is returned when no osu! process is running.
	ErrProcessNotFound = errors.New("osumem: osu! process not found")
	// ErrPatternNotFound is returned when an AOB signature cannot be located.
	ErrPatternNotFound = errors.New("osumem: memory pattern not found")
	// ErrBeatmapNotLoaded is returned when no beatmap is currently selected.
	ErrBeatmapNotLoaded = errors.New("osumem: no beatmap loaded")
)

// Client is implemented by StableReader and LazerReader. It is the
// client-agnostic surface returned by Open.
type Client interface {
	// Kind returns "stable" or "lazer".
	Kind() string
	// PID returns the attached process id.
	PID() uint32
	// Beatmap reads and returns metadata for the selected beatmap.
	Beatmap() (*BeatmapInfo, error)
	// BeatmapFilePath resolves the on-disk .osu file for info, or "" when the
	// path cannot be determined.
	BeatmapFilePath(info *BeatmapInfo) string
	// Close releases the process handle.
	Close()
}

// Client kind identifiers.
const (
	KindStable = "stable"
	KindLazer  = "lazer"
)

// BeatmapInfo is the client-agnostic beatmap metadata returned by Beatmap.
// Some fields are only meaningful for one client (e.g. Folder is empty for
// lazer, and Plays is always zero there); see the per-client readers for notes.
type BeatmapInfo struct {
	Gamemode           int32   `json:"gamemode"`
	Checksum           string  `json:"checksum"`
	Filename           string  `json:"filename"`
	Plays              int32   `json:"plays"`
	Artist             string  `json:"artist"`
	ArtistOriginal     string  `json:"artistOriginal"`
	Title              string  `json:"title"`
	TitleOriginal      string  `json:"titleOriginal"`
	AR                 float32 `json:"ar"`
	CS                 float32 `json:"cs"`
	HP                 float32 `json:"hp"`
	OD                 float32 `json:"od"`
	AudioFilename      string  `json:"audioFilename"`
	BackgroundFilename string  `json:"backgroundFilename"`
	Folder             string  `json:"folder"`
	Creator            string  `json:"creator"`
	Difficulty         string  `json:"difficulty"`
	MapID              int32   `json:"mapID"`
	SetID              int32   `json:"setID"`
	RankedStatus       int32   `json:"rankedStatus"`
	ObjectCount        int32   `json:"objectCount"`
}

// ---------------------------------------------------------------------------
// Attached process
// ---------------------------------------------------------------------------

// process is an attached target process plus its pointer width. bitness is 32
// for osu!stable and 64 for osu!lazer; it drives pointer and C#-string layouts.
// The handle is an OS object (a Win32 HANDLE on Windows, an open
// /proc/<pid>/mem file descriptor on Linux) and is interpreted by the
// platform-specific methods.
type process struct {
	handle  uintptr
	pid     uint32
	bitness int
}

// ---------------------------------------------------------------------------
// Raw memory reads (platform-independent wrappers over process.read)
// ---------------------------------------------------------------------------

func (p *process) readUint32(addr uintptr) (uint32, error) {
	b, err := p.read(addr, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(b), nil
}

func (p *process) readInt32(addr uintptr) (int32, error) {
	b, err := p.read(addr, 4)
	if err != nil {
		return 0, err
	}
	return int32(binary.LittleEndian.Uint32(b)), nil
}

func (p *process) readUint64(addr uintptr) (uint64, error) {
	b, err := p.read(addr, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(b), nil
}

func (p *process) readFloat(addr uintptr) (float32, error) {
	v, err := p.readUint32(addr)
	if err != nil {
		return 0, err
	}
	return math.Float32frombits(v), nil
}

// readPtr reads one pointer-sized value from addr (4 bytes for stable, 8 for
// lazer). This corresponds to tsprocess's readIntPtr.
func (p *process) readPtr(addr uintptr) (uintptr, error) {
	if p.bitness == 64 {
		v, err := p.readUint64(addr)
		return uintptr(v), err
	}
	v, err := p.readUint32(addr)
	return uintptr(v), err
}

// readPtrPtr dereferences twice, matching tsprocess's readPointer which is
// defined as readIntPtr(readIntPtr(address)).
func (p *process) readPtrPtr(addr uintptr) (uintptr, error) {
	first, err := p.readPtr(addr)
	if err != nil || first == 0 {
		return 0, err
	}
	return p.readPtr(first)
}

// decodeUTF16 converts UTF-16 code units to a Go string. This is a
// platform-independent replacement for syscall.UTF16ToString, which only
// exists on Windows.
func decodeUTF16(u []uint16) string {
	return string(utf16.Decode(u))
}

// readSharpString decodes a .NET System.String, matching tsprocess's
// readCSharpString. The object layout is [length:int32] followed by UTF-16
// code units, with the length field placed at pointerSize/2 * 4 bytes past the
// object header (offset 4 for 32-bit, 8 for 64-bit).
func (p *process) readSharpString(addr uintptr) (string, error) {
	if addr == 0 {
		return "", nil
	}

	lengthOffset := uintptr(4)
	if p.bitness == 64 {
		lengthOffset = 8
	}

	lenBuf, err := p.read(addr+lengthOffset, 4)
	if err != nil {
		return "", err
	}
	length := int(int32(binary.LittleEndian.Uint32(lenBuf)))
	// Anything outside this range is not a plausible osu! string; tosu treats
	// such values as an empty string rather than failing.
	if length <= 0 || length >= 4096 {
		return "", nil
	}

	dataBuf, err := p.read(addr+lengthOffset+4, length*2)
	if err != nil {
		return "", err
	}

	u16 := make([]uint16, length)
	for i := range u16 {
		u16[i] = binary.LittleEndian.Uint16(dataBuf[i*2:])
	}
	return decodeUTF16(u16), nil
}

// readSharpStringPtr reads a pointer at addr and then decodes the C# string it
// points to (tsprocess readSharpStringPtr).
func (p *process) readSharpStringPtr(addr uintptr) (string, error) {
	ptr, err := p.readPtr(addr)
	if err != nil || ptr == 0 {
		return "", err
	}
	return p.readSharpString(ptr)
}

// ---------------------------------------------------------------------------
// AOB pattern scanning
// ---------------------------------------------------------------------------

// scan returns the address of the first signature match.
func (p *process) scan(sig *signature) (uintptr, error) {
	var found uintptr
	p.forEachRegion(func(base, size uintptr) bool {
		if a, ok := scanRegionFirst(p.read, base, size, sig); ok {
			found = a
			return true
		}
		return false
	})
	if found == 0 {
		return 0, ErrPatternNotFound
	}
	return found, nil
}

// scanFunc returns the first signature match accepted by fn, walking the rw
// regions in order and stopping as soon as one is found. Unlike scan it can
// consider several candidates, and unlike a collect-all scan it validates each
// candidate as it is discovered, so it never reads regions past the one holding
// the accepted match. lazer's anchor resolution uses it this way.
func (p *process) scanFunc(sig *signature, fn func(addr uintptr) bool) uintptr {
	var found uintptr
	p.forEachRegion(func(base, size uintptr) bool {
		if a, ok := scanRegionFunc(p.read, base, size, sig, fn); ok {
			found = a
			return true
		}
		return false
	})
	return found
}

// mustSignature parses a constant pattern. Patterns are compile-time literals,
// so a parse failure is a programming error and panics.
func mustSignature(pattern string) *signature {
	sig, err := newSignature(pattern)
	if err != nil {
		panic(err)
	}
	return sig
}

// ---------------------------------------------------------------------------
// Entry points
// ---------------------------------------------------------------------------

// Open finds a running osu! process (stable or lazer), attaches to it and
// returns the appropriate reader. The client is chosen by pointer width, the
// same way tosu does: lazer is 64-bit, stable is 32-bit.
func Open() (Client, error) {
	pids := findProcesses("osu!.exe", "osulazer.exe", "osu!")
	if len(pids) == 0 {
		return nil, ErrProcessNotFound
	}
	return OpenPID(pids[0])
}

// OpenPID attaches to a specific process id and selects the reader by bitness.
func OpenPID(pid uint32) (Client, error) {
	enableDebugPrivilege()

	handle, err := openHandle(pid)
	if err != nil {
		return nil, err
	}

	p := &process{handle: handle, pid: pid, bitness: 32}
	if isProcess64Bit(pid) {
		p.bitness = 64
	}

	if p.bitness == 64 {
		return newLazerReader(p)
	}
	return newStableReader(p)
}

// newStableReader resolves osu!stable's base signatures and builds a reader.
func newStableReader(p *process) (*StableReader, error) {
	baseAddr, err := p.scan(mustSignature(sigBaseAddr))
	if err != nil {
		p.close()
		return nil, fmt.Errorf("osumem: locate stable base: %w", err)
	}
	// The settings signature is only needed to resolve the Songs folder; it is
	// optional, so a miss just leaves the folder resolution to the default.
	settingsClassAddr, _ := p.scan(mustSignature(sigSettingsClass))

	var osuPath string
	if path, err := p.imagePath(); err == nil {
		osuPath = filepath.Dir(path)
	}

	return &StableReader{
		mem:               p,
		pid:               p.pid,
		closeFn:           p.close,
		baseAddr:          baseAddr,
		settingsClassAddr: settingsClassAddr,
		osuPath:           osuPath,
	}, nil
}

// newLazerReader loads the versioned offsets and builds a reader whose anchor
// scan may re-resolve GameBase on demand.
func newLazerReader(p *process) (*LazerReader, error) {
	offsets, err := loadLazerOffsets()
	if err != nil {
		p.close()
		return nil, err
	}

	l := &LazerReader{
		mem:     p,
		pid:     p.pid,
		closeFn: p.close,
		offsets: offsets,
	}
	l.resolve = func() uintptr {
		var gameBase uintptr
		p.scanFunc(mustSignature(sigLazerAnchor), func(anchor uintptr) bool {
			gameBase = l.resolveGameBaseFromAnchor(anchor)
			return gameBase != 0
		})
		return gameBase
	}
	return l, nil
}
