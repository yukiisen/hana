//go:build linux

package osumem

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// This file ports packages/tsprocess/lib/memory/memory_linux.cc: process
// discovery via /proc, memory reads via /proc/<pid>/mem and region enumeration
// via /proc/<pid>/maps.

// findProcesses returns the PIDs of every running process whose executable
// name (from /proc/<pid>/comm) contains one of names. This mirrors tsprocess,
// which substring-matches the comm file; on Linux osu!lazer's comm is "osu!".
func findProcesses(names ...string) []uint32 {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	var pids []uint32
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.ParseUint(entry.Name(), 10, 32)
		if err != nil {
			continue
		}
		comm, err := os.ReadFile("/proc/" + entry.Name() + "/comm")
		if err != nil {
			continue
		}
		name := strings.TrimSpace(string(comm))
		for _, want := range names {
			if strings.Contains(name, want) {
				pids = append(pids, uint32(pid))
				break
			}
		}
	}
	return pids
}

// openHandle opens /proc/<pid>/mem. Reading another process's memory requires
// the same ptrace-level permission as process_vm_readv: run as root, grant
// cap_sys_ptrace, or allow it via /proc/sys/kernel/yama/ptrace_scope.
func openHandle(pid uint32) (uintptr, error) {
	path := fmt.Sprintf("/proc/%d/mem", pid)
	fd, err := syscall.Open(path, syscall.O_RDONLY, 0)
	if err != nil {
		return 0, fmt.Errorf("osumem: open %s: %w (run as root or setcap cap_sys_ptrace=eip)", path, err)
	}
	return uintptr(fd), nil
}

// enableDebugPrivilege is a no-op on Linux; privileges are governed by ptrace.
func enableDebugPrivilege() {}

// isProcess64Bit reads the ELF class byte from /proc/<pid>/exe (2 = 64-bit).
func isProcess64Bit(pid uint32) bool {
	f, err := os.Open(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return false
	}
	defer f.Close()

	var header [5]byte
	if _, err := io.ReadFull(f, header[:]); err != nil {
		return false
	}
	if header[0] != 0x7f || header[1] != 'E' || header[2] != 'L' || header[3] != 'F' {
		return false
	}
	return header[4] == 2
}

// ---------------------------------------------------------------------------
// process methods
// ---------------------------------------------------------------------------

func (p *process) close() {
	if p.handle != 0 {
		syscall.Close(int(p.handle))
		p.handle = 0
	}
}

// alive reports whether the target process is still running. A closed handle is
// treated as dead; otherwise kill(pid, 0) distinguishes a missing process
// (ESRCH) from one we may not signal (EPERM, still alive). A process left as a
// zombie counts as gone because its memory can no longer be read.
func (p *process) alive() bool {
	if p.handle == 0 {
		return false
	}
	if err := syscall.Kill(int(p.pid), 0); err != nil {
		return errors.Is(err, syscall.EPERM)
	}
	return !isZombie(p.pid)
}

// isZombie reports whether pid has exited but not yet been reaped. The state
// field follows the parenthesised comm name in /proc/<pid>/stat; finding the
// last ')' skips any spaces inside the name.
func isZombie(pid uint32) bool {
	b, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return true
	}
	i := bytes.LastIndexByte(b, ')')
	if i < 0 || i+2 >= len(b) {
		return false
	}
	return b[i+2] == 'Z'
}

// read copies size bytes from the target at addr through /proc/<pid>/mem.
func (p *process) read(addr uintptr, size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("osumem: invalid read size")
	}
	buf := make([]byte, size)
	n, err := syscall.Pread(int(p.handle), buf, int64(addr))
	if err != nil {
		return nil, fmt.Errorf("osumem: read /proc/%d/mem at 0x%X failed: %w", p.pid, addr, err)
	}
	if n != size {
		return nil, fmt.Errorf("osumem: short read at 0x%X: got %d of %d bytes", addr, n, size)
	}
	return buf, nil
}

// imagePath returns the target's executable path.
func (p *process) imagePath() (string, error) {
	path, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", p.pid))
	if err != nil {
		return "", fmt.Errorf("osumem: readlink /proc/%d/exe: %w", p.pid, err)
	}
	return path, nil
}

// forEachRegion walks the target's mapped, readable and writable regions,
// invoking fn for each. Matching tsprocess's query_regions, only rw mappings
// are scanned. Iteration stops early if fn returns true.
func (p *process) forEachRegion(fn func(base, size uintptr) bool) {
	f, err := os.Open(fmt.Sprintf("/proc/%d/maps", p.pid))
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		perms := fields[1]
		if len(perms) < 2 || perms[0] != 'r' || perms[1] != 'w' {
			continue
		}

		dash := strings.IndexByte(fields[0], '-')
		if dash < 0 {
			continue
		}
		start, err1 := strconv.ParseUint(fields[0][:dash], 16, 64)
		end, err2 := strconv.ParseUint(fields[0][dash+1:], 16, 64)
		if err1 != nil || err2 != nil || end <= start {
			continue
		}

		if fn(uintptr(start), uintptr(end-start)) {
			return
		}
	}
}
