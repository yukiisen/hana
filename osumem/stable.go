package osumem

import (
	"fmt"
	"path/filepath"
)

// osu!stable is a 32-bit .NET/XNA application. Its addresses are not exported
// in any symbol table, so tosu locates the interesting globals by scanning for
// short, version-stable stretches of machine code (AOB signatures) and then
// reading through fixed offsets from those addresses.
//
// This reader ports the exact signatures and offsets from
// packages/tosu/src/memory/stable.ts. The struct itself is platform independent
// (it only needs a memory); the Windows constructor that resolves the
// signatures lives in osumem.go.
const (
	// sigBaseAddr anchors the "base address" global that points into the core
	// gameplay/beatmap structures. In stable.ts this is `baseAddr`.
	sigBaseAddr = "F8 01 74 04 83 65"
	// sigSettingsClass anchors the settings/config singleton, used only to
	// resolve the Songs folder path.
	sigSettingsClass = "83 E0 20 85 C0 7E 2F"
)

// StableReader reads an attached 32-bit osu!stable process.
type StableReader struct {
	mem     memory
	pid     uint32
	closeFn func()

	// baseAddr and settingsClassAddr are the resolved locations of the two
	// signatures above.
	baseAddr          uintptr
	settingsClassAddr uintptr

	// osuPath is the directory containing osu!.exe.
	osuPath string
}

// Kind implements Client.
func (r *StableReader) Kind() string { return KindStable }

// PID implements Client.
func (r *StableReader) PID() uint32 { return r.pid }

// Close implements Client.
func (r *StableReader) Close() {
	if r.closeFn != nil {
		r.closeFn()
	}
}

// SongsFolder reads the configured Songs directory from the settings singleton.
// The path may be absolute or relative to the osu! install directory.
//
// Pointer chain (stable.ts global()):
//
//	[[[SettingsClass + 0x8] + 0xB8] + 0x4] -> C# string
func (r *StableReader) SongsFolder() string {
	if r.settingsClassAddr == 0 {
		return ""
	}
	m := r.mem

	// [SettingsClass + 0x8]
	p1, err := m.readUint32(r.settingsClassAddr + 0x8)
	if err != nil || p1 == 0 {
		return ""
	}
	// [[SettingsClass + 0x8] + 0xB8]
	p2, err := m.readUint32(uintptr(p1) + 0xb8)
	if err != nil || p2 == 0 {
		return ""
	}
	// [[[SettingsClass + 0x8] + 0xB8] + 0x4]
	p3, err := m.readUint32(uintptr(p2) + 0x4)
	if err != nil || p3 == 0 {
		return ""
	}
	folder, err := m.readSharpString(uintptr(p3))
	if err != nil {
		return ""
	}
	return folder
}

// Beatmap reads the currently selected beatmap.
//
// Offsets mirror stable.ts menu(). The beatmap object is reached from the base
// global:
//
//	beatmapAddr = readPointer(baseAddr - 0xC)   // double dereference
//
// All subsequent fields are read from beatmapAddr using fixed offsets.
func (r *StableReader) Beatmap() (*BeatmapInfo, error) {
	m := r.mem
	base := r.baseAddr

	// beatmapAddr = readPointer(baseAddr - 0xC) (tsprocess readPointer = two derefs)
	beatmapAddr, err := m.readPtrPtr(base - 0xc)
	if err != nil {
		return nil, fmt.Errorf("osumem: read beatmap pointer: %w", err)
	}
	if beatmapAddr == 0 {
		return nil, ErrBeatmapNotLoaded
	}

	// The ruleset base is single-dereferenced for the play count.
	rulesetBase, _ := m.readUint32(base - 0x33)
	// Gamemode, matching tosu, is double-dereferenced from the same slot.
	gamemode, _ := m.readPtrPtr(base - 0x33)

	// Helpers over the beatmap struct. Any failed field read degrades to the
	// zero value so a single bad offset cannot abort the whole read.
	i32 := func(off uintptr) int32 {
		if v, err := m.readInt32(beatmapAddr + off); err == nil {
			return v
		}
		return 0
	}
	f32 := func(off uintptr) float32 {
		if v, err := m.readFloat(beatmapAddr + off); err == nil {
			return v
		}
		return 0
	}
	str := func(off uintptr) string {
		ptr, err := m.readUint32(beatmapAddr + off)
		if err != nil || ptr == 0 {
			return ""
		}
		s, err := m.readSharpString(uintptr(ptr))
		if err != nil {
			return ""
		}
		return s
	}

	var plays int32
	if rulesetBase != 0 {
		plays, _ = m.readInt32(uintptr(rulesetBase) + 0xc)
	}

	info := &BeatmapInfo{
		Gamemode:           int32(gamemode),
		Checksum:           str(0x6c),
		Filename:           str(0x90),
		Plays:              plays,
		Artist:             str(0x18),
		ArtistOriginal:     str(0x1c),
		Title:              str(0x24),
		TitleOriginal:      str(0x28),
		AR:                 f32(0x2c),
		CS:                 f32(0x30),
		HP:                 f32(0x34),
		OD:                 f32(0x38),
		AudioFilename:      str(0x64),
		BackgroundFilename: str(0x68),
		Folder:             str(0x78),
		Creator:            str(0x7c),
		Difficulty:         str(0xac),
		MapID:              i32(0xc8),
		SetID:              i32(0xcc),
		RankedStatus:       i32(0x12c),
		ObjectCount:        i32(0xf8),
	}

	// A real beatmap always has a .osu filename; its absence means the user is
	// sitting in the main menu or a song isn't selected yet.
	if info.Filename == "" {
		return nil, ErrBeatmapNotLoaded
	}
	return info, nil
}

// BeatmapFilePath resolves <osuPath>/<SongsFolder>/<folder>/<filename>.
func (r *StableReader) BeatmapFilePath(info *BeatmapInfo) string {
	if info == nil || info.Folder == "" || info.Filename == "" {
		return ""
	}

	root := r.osuPath
	if root == "" {
		root = "."
	}

	if songs := r.SongsFolder(); songs != "" {
		if filepath.IsAbs(songs) {
			root = songs
		} else {
			root = filepath.Join(root, songs)
		}
	} else {
		root = filepath.Join(root, "Songs")
	}

	return filepath.Join(root, info.Folder, info.Filename)
}
