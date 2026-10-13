package osumem

import (
	"errors"
	"testing"
)

// fakeMemory is an in-memory memory whose liveness is controllable. Its reads
// always fail with errFakeRead, so any error a decoder returns other than
// ErrProcessGone must come from the decoding path rather than the process.
type fakeMemory struct{ live bool }

var errFakeRead = errors.New("fake: read")

func (f *fakeMemory) alive() bool { return f.live }

func (f *fakeMemory) readPtr(uintptr) (uintptr, error)           { return 0, errFakeRead }
func (f *fakeMemory) readPtrPtr(uintptr) (uintptr, error)        { return 0, errFakeRead }
func (f *fakeMemory) readUint32(uintptr) (uint32, error)         { return 0, errFakeRead }
func (f *fakeMemory) readInt32(uintptr) (int32, error)           { return 0, errFakeRead }
func (f *fakeMemory) readUint64(uintptr) (uint64, error)         { return 0, errFakeRead }
func (f *fakeMemory) readFloat(uintptr) (float32, error)         { return 0, errFakeRead }
func (f *fakeMemory) readSharpString(uintptr) (string, error)    { return "", errFakeRead }
func (f *fakeMemory) readSharpStringPtr(uintptr) (string, error) { return "", errFakeRead }

// A closed game must yield the same error from every client so callers can
// reattach uniformly.
func TestBeatmapReportsProcessGone(t *testing.T) {
	dead := &fakeMemory{live: false}

	stable := &StableReader{mem: dead}
	if _, err := stable.Beatmap(); !errors.Is(err, ErrProcessGone) {
		t.Fatalf("StableReader.Beatmap err = %v, want ErrProcessGone", err)
	}

	lazer := &LazerReader{mem: dead}
	if _, err := lazer.Beatmap(); !errors.Is(err, ErrProcessGone) {
		t.Fatalf("LazerReader.Beatmap err = %v, want ErrProcessGone", err)
	}
}

// A live process must not be misreported as gone: stable surfaces its read
// failure and lazer (which swallows read errors) falls back to the empty-menu
// error.
func TestBeatmapLiveProcessNotReportedGone(t *testing.T) {
	live := &fakeMemory{live: true}

	stable := &StableReader{mem: live}
	if _, err := stable.Beatmap(); errors.Is(err, ErrProcessGone) {
		t.Fatalf("StableReader.Beatmap err = %v, must not be ErrProcessGone", err)
	}

	lazer := &LazerReader{mem: live, offsets: &lazerOffsets{classes: map[string]map[string]int64{}}}
	if _, err := lazer.Beatmap(); !errors.Is(err, ErrBeatmapNotLoaded) {
		t.Fatalf("LazerReader.Beatmap err = %v, want ErrBeatmapNotLoaded", err)
	}
}
