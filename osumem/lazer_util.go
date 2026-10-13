package osumem

// lazerToStableStatus maps osu!lazer's BeatmapOnlineStatus enum onto the
// legacy stable ranked-status numbers, matching tosu's lazerToStableStatus.
var lazerToStableStatus = map[int32]int32{
	-4: 1, // LocallyModified
	-3: 1, // None
	-2: 2, // Graveyard
	-1: 2, // WIP
	0:  2, // Pending
	1:  4, // Ranked
	2:  5, // Approved
	3:  6, // Qualified
	4:  7, // Loved
}

// toLazerPath converts a content hash into the relative path osu!lazer uses on
// disk under its files/ directory: <first char>/<first two chars>/<hash>.
func toLazerPath(hash string) string {
	if len(hash) < 2 {
		return ""
	}
	return hash[0:1] + "/" + hash[0:2] + "/" + hash
}
