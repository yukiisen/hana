package osumem

import (
	"path/filepath"
)

// osu!lazer is a 64-bit .NET 8 application built on the osu!framework. Unlike
// stable, its globals cannot be found with short code signatures because the
// JIT-compiled/managed layout changes constantly. tosu instead:
//
//  1. Scans for one very stable anchor pattern (a Vector2 constant stored by
//     ScalingContainer) to get a foothold inside the OsuGame object.
//  2. Sweeps a few candidate byte deltas around that anchor to find an
//     ExternalLinkOpener reference, then walks ExternalLinkOpener.<api>.game to
//     reach the singleton APIAccess and the GameBase object.
//  3. Validates the candidate against the expected GameBase vtable entry (from
//     offsets.json) so only the real object is accepted.
//  4. Reads every managed field through versioned offsets from offsets.json.
//
// This file ports packages/tosu/src/memory/lazer.ts. The Windows constructor
// that owns the process handle and provides the anchor scanner lives in
// osumem.go; this file is platform independent so it can be unit tested with a
// fake memory.

const (
	// sigLazerAnchor is the ScalingContainerTargetDrawSize Vector2 constant.
	// This is the single signature lazer scanning relies on.
	sigLazerAnchor = "01 01 00 00 00 00 80 44 00 00 40 44"

	// fallbackGameBaseVtable is tosu's hard-coded vtable used only when
	// offsets.json does not provide GameBaseVtable.
	fallbackGameBaseVtable uint64 = 7730957910016
)

// Fully-qualified .NET type names used to index offsets.json.
const (
	classOsuGameBase    = "osu.Game.OsuGameBase"
	classOsuGameDesktop = "osu.Desktop.OsuGameDesktop"

	classExternalLinkOpener = "osu.Game.Online.Chat.ExternalLinkOpener"
	classAPIAccess          = "osu.Game.Online.API.APIAccess"

	classWorkingBeatmap = "osu.Game.Beatmaps.WorkingBeatmapCache+BeatmapManagerWorkingBeatmap"
	classBeatmapInfo    = "osu.Game.Beatmaps.BeatmapInfo"
	classBeatmapSetInfo = "osu.Game.Beatmaps.BeatmapSetInfo"
	classBeatmapMeta    = "osu.Game.Beatmaps.BeatmapMetadata"
	classBeatmapDiff    = "osu.Game.Beatmaps.BeatmapDifficulty"
	classRulesetInfo    = "osu.Game.Rulesets.RulesetInfo"
	classRealmUser      = "osu.Game.Models.RealmUser"

	classStorage        = "osu.Framework.Platform.Storage"
	classWrappedStorage = "osu.Game.IO.WrappedStorage"
)

// Field names as they appear in offsets.json.
const (
	fieldBeatmap = "<Beatmap>k__BackingField"
	fieldStorage = "<Storage>k__BackingField"

	fieldRuleset       = "Ruleset"
	fieldExternalAPI   = "<api>k__BackingField"
	fieldAPIGame       = "game"
	fieldBeatmapInfo   = "BeatmapInfo"
	fieldBeatmapSet    = "BeatmapSetInfo"
	fieldMD5Hash       = "<MD5Hash>k__BackingField"
	fieldStatusInt     = "<StatusInt>k__BackingField"
	fieldMetadata      = "<Metadata>k__BackingField"
	fieldDifficulty    = "<Difficulty>k__BackingField"
	fieldHash          = "<Hash>k__BackingField"
	fieldDiffName      = "<DifficultyName>k__BackingField"
	fieldOnlineID      = "<OnlineID>k__BackingField"
	fieldObjectCount   = "<TotalObjectCount>k__BackingField"
	fieldAuthor        = "<Author>k__BackingField"
	fieldAudioFile     = "<AudioFile>k__BackingField"
	fieldBackground    = "<BackgroundFile>k__BackingField"
	fieldTitle         = "<Title>k__BackingField"
	fieldTitleUnicode  = "<TitleUnicode>k__BackingField"
	fieldArtist        = "<Artist>k__BackingField"
	fieldArtistUnicode = "<ArtistUnicode>k__BackingField"
	fieldApproachRate  = "<ApproachRate>k__BackingField"
	fieldCircleSize    = "<CircleSize>k__BackingField"
	fieldDrainRate     = "<DrainRate>k__BackingField"
	fieldOverallDiff   = "<OverallDifficulty>k__BackingField"
	fieldUsername      = "<Username>k__BackingField"
	fieldBasePath      = "<BasePath>k__BackingField"
	fieldUnderlying    = "<UnderlyingStorage>k__BackingField"
)

// LazerReader reads an attached 64-bit osu!lazer process.
type LazerReader struct {
	mem     memory
	pid     uint32
	closeFn func()

	offsets *lazerOffsets

	// gameBaseAddr is the resolved GameBase object.
	gameBaseAddr uintptr
	// resolve re-scans for GameBase; set by the Windows constructor. It is nil
	// for fake readers in tests, in which case the cached address is trusted.
	resolve func() uintptr
}

// Kind implements Client.
func (l *LazerReader) Kind() string { return KindLazer }

// PID implements Client.
func (l *LazerReader) PID() uint32 { return l.pid }

// Close implements Client.
func (l *LazerReader) Close() {
	if l.closeFn != nil {
		l.closeFn()
	}
}

// OsuVersion returns the osu!lazer build the offsets were generated for.
func (l *LazerReader) OsuVersion() string {
	if l.offsets == nil {
		return ""
	}
	return l.offsets.version
}

// field returns the byte offset for a class/field pair, or 0 if unknown.
func (l *LazerReader) field(class, name string) uintptr {
	return uintptr(l.offsets.off(class, name))
}

// Small read helpers; failed reads degrade to zero values.
func (l *LazerReader) ptr(base, off uintptr) uintptr {
	v, _ := l.mem.readPtr(base + off)
	return v
}

func (l *LazerReader) str(base, off uintptr) string {
	s, _ := l.mem.readSharpStringPtr(base + off)
	return s
}

func (l *LazerReader) i32(base, off uintptr) int32 {
	v, _ := l.mem.readInt32(base + off)
	return v
}

func (l *LazerReader) f32(base, off uintptr) float32 {
	v, _ := l.mem.readFloat(base + off)
	return v
}

// checkIfGameBase validates a candidate pointer by comparing the first entry of
// its vtable against the expected GameBase vtable value.
func (l *LazerReader) checkIfGameBase(addr uintptr) bool {
	if addr == 0 {
		return false
	}
	vtable, err := l.mem.readPtr(addr)
	if err != nil || vtable == 0 {
		return false
	}

	expected := l.offsets.gameBaseVtable
	if expected == 0 {
		expected = fallbackGameBaseVtable
	}

	got, err := l.mem.readUint64(vtable)
	if err != nil {
		return false
	}
	return got == expected
}

// resolveGameBaseFromAnchor tries a handful of byte deltas behind the anchor to
// locate ExternalLinkOpener. The distance shifts between osu! versions and the
// field is not in offsets.json, so several candidates are swept and only a
// candidate whose .game pointer passes checkIfGameBase is accepted.
func (l *LazerReader) resolveGameBaseFromAnchor(anchor uintptr) uintptr {
	for _, delta := range []uintptr{0x24, 0x28, 0x2c, 0x20, 0x30, 0x1c, 0x34} {
		externalLinkOpener, err := l.mem.readPtr(anchor - delta)
		if err != nil || externalLinkOpener == 0 {
			continue
		}

		api := l.ptr(externalLinkOpener, l.field(classExternalLinkOpener, fieldExternalAPI))
		if api == 0 {
			continue
		}

		gameBase := l.ptr(api, l.field(classAPIAccess, fieldAPIGame))
		if l.checkIfGameBase(gameBase) {
			return gameBase
		}
	}
	return 0
}

// gameBase returns the cached GameBase, re-resolving it if it no longer
// validates.
func (l *LazerReader) gameBase() uintptr {
	if l.gameBaseAddr == 0 || !l.checkIfGameBase(l.gameBaseAddr) {
		if l.resolve != nil {
			l.gameBaseAddr = l.resolve()
		}
	}
	return l.gameBaseAddr
}

// readGamemode reads the currently selected ruleset's OnlineID (0=osu,
// 1=taiko, 2=catch, 3=mania).
func (l *LazerReader) readGamemode() int32 {
	bindable := l.ptr(l.gameBase(), l.field(classOsuGameDesktop, fieldRuleset))
	if bindable == 0 {
		return 0
	}
	// The Ruleset bindable stores the ruleset info at +0x20.
	rulesetInfo := l.ptr(bindable, 0x20)
	if rulesetInfo == 0 {
		return 0
	}
	return l.i32(rulesetInfo, l.field(classRulesetInfo, fieldOnlineID))
}

// currentBeatmap returns the BeatmapInfo and BeatmapSetInfo object pointers for
// the beatmap currently held by the game's Beatmap bindable:
//
//	GameBase.<Beatmap> -> bindable+0x20 -> working beatmap
//	  working + BeatmapInfo offset -> BeatmapInfo
//	  working + BeatmapSetInfo offset -> BeatmapSetInfo
func (l *LazerReader) currentBeatmap() (info, set uintptr) {
	bindable := l.ptr(l.gameBase(), l.field(classOsuGameBase, fieldBeatmap))
	if bindable == 0 {
		return 0, 0
	}
	working := l.ptr(bindable, 0x20)
	if working == 0 {
		return 0, 0
	}
	info = l.ptr(working, l.field(classWorkingBeatmap, fieldBeatmapInfo))
	set = l.ptr(working, l.field(classWorkingBeatmap, fieldBeatmapSet))
	return info, set
}

// getBeatmapFiles returns a map of file name -> content hash for every file in
// the beatmap set (each file is stored on disk under its hash). This is how the
// audio and background file hashes are resolved.
func (l *LazerReader) getBeatmapFiles(setInfo uintptr) map[string]string {
	result := make(map[string]string)
	if setInfo == 0 {
		return result
	}

	// The files collection is an osu.Framework BindableList: [items ptr] at
	// +0x8 and [count] at +0x10.
	files := l.ptr(setInfo, 0x20)
	if files == 0 {
		return result
	}
	size, _ := l.mem.readInt32(files + 0x10)
	items := l.ptr(files, 0x8)

	for i := int32(0); i < size; i++ {
		current := l.ptr(items, uintptr(0x10+0x8*int(i)))
		if current == 0 {
			continue
		}
		realmFile := l.ptr(current, 0x18)
		if realmFile == 0 {
			continue
		}
		hash := l.str(realmFile, 0x18)
		fileName := l.str(current, 0x20)
		if fileName != "" && hash != "" {
			result[fileName] = hash
		}
	}
	return result
}

// dataPath reads the osu!lazer install/data directory from the game's Storage
// object: GameBase.<Storage> -> underlying storage -> <BasePath>.
func (l *LazerReader) dataPath() string {
	storage := l.ptr(l.gameBase(), l.field(classOsuGameBase, fieldStorage))
	if storage == 0 {
		return ""
	}
	underlying := l.ptr(storage, l.field(classWrappedStorage, fieldUnderlying))
	if underlying == 0 {
		return ""
	}
	return l.str(underlying, l.field(classStorage, fieldBasePath))
}

// SongsFolder returns the lazer files directory (<dataPath>/files), which is
// the analogue of stable's Songs folder.
func (l *LazerReader) SongsFolder() string {
	base := l.dataPath()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "files")
}

// Beatmap reads the currently selected beatmap.
func (l *LazerReader) Beatmap() (*BeatmapInfo, error) {
	if processGone(l.mem) {
		return nil, ErrProcessGone
	}

	infoPtr, setPtr := l.currentBeatmap()
	if infoPtr == 0 {
		// Lazer degrades every failed read to zero, so a closed game looks
		// exactly like an empty menu; distinguish the two here.
		if processGone(l.mem) {
			return nil, ErrProcessGone
		}
		return nil, ErrBeatmapNotLoaded
	}

	// Checksum is the beatmap's MD5 (matching stable's Checksum field).
	checksum := l.str(infoPtr, l.field(classBeatmapInfo, fieldMD5Hash))
	gamemode := l.readGamemode()

	statusInt := l.i32(infoPtr, l.field(classBeatmapInfo, fieldStatusInt))
	rankedStatus, ok := lazerToStableStatus[statusInt]
	if !ok {
		rankedStatus = 0
	}

	metadata := l.ptr(infoPtr, l.field(classBeatmapInfo, fieldMetadata))
	difficulty := l.ptr(infoPtr, l.field(classBeatmapInfo, fieldDifficulty))
	hash := l.str(infoPtr, l.field(classBeatmapInfo, fieldHash))

	author := l.ptr(metadata, l.field(classBeatmapMeta, fieldAuthor))

	// Resolve the audio/background file hashes via the set's file table.
	files := l.getBeatmapFiles(setPtr)
	audioHash := files[l.str(metadata, l.field(classBeatmapMeta, fieldAudioFile))]
	backgroundHash := files[l.str(metadata, l.field(classBeatmapMeta, fieldBackground))]

	info := &BeatmapInfo{
		Gamemode:           gamemode,
		Checksum:           checksum,
		Filename:           toLazerPath(hash),
		Plays:              0, // lazer does not expose a local play count here
		Artist:             l.str(metadata, l.field(classBeatmapMeta, fieldArtist)),
		ArtistOriginal:     l.str(metadata, l.field(classBeatmapMeta, fieldArtistUnicode)),
		Title:              l.str(metadata, l.field(classBeatmapMeta, fieldTitle)),
		TitleOriginal:      l.str(metadata, l.field(classBeatmapMeta, fieldTitleUnicode)),
		AR:                 l.f32(difficulty, l.field(classBeatmapDiff, fieldApproachRate)),
		CS:                 l.f32(difficulty, l.field(classBeatmapDiff, fieldCircleSize)),
		HP:                 l.f32(difficulty, l.field(classBeatmapDiff, fieldDrainRate)),
		OD:                 l.f32(difficulty, l.field(classBeatmapDiff, fieldOverallDiff)),
		AudioFilename:      toLazerPath(audioHash),
		BackgroundFilename: toLazerPath(backgroundHash),
		Folder:             "", // lazer has no per-beatmap folder; files are hash-addressed
		Creator:            l.str(author, l.field(classRealmUser, fieldUsername)),
		Difficulty:         l.str(infoPtr, l.field(classBeatmapInfo, fieldDiffName)),
		MapID:              l.i32(infoPtr, l.field(classBeatmapInfo, fieldOnlineID)),
		SetID:              l.i32(setPtr, l.field(classBeatmapSetInfo, fieldOnlineID)),
		RankedStatus:       rankedStatus,
		ObjectCount:        l.i32(infoPtr, l.field(classBeatmapInfo, fieldObjectCount)),
	}

	// A loaded beatmap always exposes a content hash; without it the game is
	// likely still loading or sitting on the main menu. Check for a vanished
	// process first so the two cases stay distinguishable.
	if info.Filename == "" {
		if processGone(l.mem) {
			return nil, ErrProcessGone
		}
		return nil, ErrBeatmapNotLoaded
	}
	return info, nil
}

// BeatmapFilePath resolves the on-disk .osu file:
// <dataPath>/files/<first char>/<first two chars>/<hash>.
func (l *LazerReader) BeatmapFilePath(info *BeatmapInfo) string {
	if info == nil || info.Filename == "" {
		return ""
	}
	songs := l.SongsFolder()
	if songs == "" {
		return ""
	}
	return filepath.Join(songs, filepath.FromSlash(info.Filename))
}
