package osumem

// memory is the narrow read surface the client decoders depend on. It is
// satisfied by the platform-specific *process on Windows and by the in-memory
// fake used in tests, which is what lets the beatmap decoding be tested without
// a live osu! process.
type memory interface {
	// readPtr reads one pointer-sized value (4 or 8 bytes depending on the
	// target's bitness).
	readPtr(addr uintptr) (uintptr, error)
	// readPtrPtr dereferences twice (tsprocess's readPointer).
	readPtrPtr(addr uintptr) (uintptr, error)
	readUint32(addr uintptr) (uint32, error)
	readInt32(addr uintptr) (int32, error)
	readUint64(addr uintptr) (uint64, error)
	readFloat(addr uintptr) (float32, error)
	// readSharpString decodes a .NET System.String object at addr.
	readSharpString(addr uintptr) (string, error)
	// readSharpStringPtr reads a pointer at addr then decodes the string there.
	readSharpStringPtr(addr uintptr) (string, error)
}
