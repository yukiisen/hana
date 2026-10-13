package osumem

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

// offsets.json is a verbatim copy of tosu's
// packages/tosu/src/assets/offsets.json. osu!lazer is rewritten frequently and
// its managed object layout changes with almost every release, so the field
// offsets cannot be hard-coded in the binary. Instead the osu!lazer client
// ships them as data keyed by fully-qualified .NET type name, and this package
// embeds the same table.
//
// The file contains two scalar keys (OsuVersion, GameBaseVtable) and a set of
// object keys, each mapping a C# field name to its byte offset, for example:
//
//	"osu.Game.Beatmaps.BeatmapInfo": { "<OnlineID>k__BackingField": 140, ... }
//
//go:embed offsets.json
var offsetsJSON []byte

// lazerOffsets is the parsed form of offsets.json. The versioned, per-type
// offsets are kept in a generic map (class -> field -> offset) so no Go code
// needs to change when osu! adds or renames a field.
type lazerOffsets struct {
	version string
	// gameBaseVtable is the expected first vtable entry used to validate that a
	// candidate pointer really is the osu! GameBase object.
	gameBaseVtable uint64
	classes        map[string]map[string]int64
}

// loadLazerOffsets parses the embedded offsets table.
func loadLazerOffsets() (*lazerOffsets, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(offsetsJSON, &raw); err != nil {
		return nil, fmt.Errorf("osumem: parse offsets.json: %w", err)
	}

	o := &lazerOffsets{classes: make(map[string]map[string]int64)}
	for key, value := range raw {
		switch key {
		case "OsuVersion":
			_ = json.Unmarshal(value, &o.version)
		case "GameBaseVtable":
			_ = json.Unmarshal(value, &o.gameBaseVtable)
		default:
			// Every other top-level key is a class name mapping to its fields.
			var fields map[string]int64
			if err := json.Unmarshal(value, &fields); err == nil {
				o.classes[key] = fields
			}
		}
	}
	return o, nil
}

// off returns the byte offset of field within class, or 0 when unknown. A
// missing offset resolves to 0 (the same as tosu, where a `undefined` offset
// would produce NaN/LookupError); callers that read a required field should
// validate the result themselves.
func (o *lazerOffsets) off(class, field string) uint64 {
	if fields, ok := o.classes[class]; ok {
		if v, ok := fields[field]; ok && v >= 0 {
			return uint64(v)
		}
	}
	return 0
}
