//go:build windows

package osumem

import (
	"errors"
	"fmt"
	"strings"
	"syscall"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Win32 bindings (stdlib only; no golang.org/x/sys dependency)
// ---------------------------------------------------------------------------

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	advapi32 = syscall.NewLazyDLL("advapi32.dll")

	procReadProcessMemory = kernel32.NewProc("ReadProcessMemory")
	procVirtualQueryEx    = kernel32.NewProc("VirtualQueryEx")
	procOpenProcess       = kernel32.NewProc("OpenProcess")
	procCloseHandle       = kernel32.NewProc("CloseHandle")
	procGetCurrentProcess = kernel32.NewProc("GetCurrentProcess")
	procIsWow64Process    = kernel32.NewProc("IsWow64Process")
	procSnapshot          = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32First    = kernel32.NewProc("Process32FirstW")
	procProcess32Next     = kernel32.NewProc("Process32NextW")
	procFullImageName     = kernel32.NewProc("QueryFullProcessImageNameW")

	procOpenProcessToken      = advapi32.NewProc("OpenProcessToken")
	procLookupPrivilegeValue  = advapi32.NewProc("LookupPrivilegeValueW")
	procAdjustTokenPrivileges = advapi32.NewProc("AdjustTokenPrivileges")
)

const (
	processVMRead           = 0x0010
	processQueryInformation = 0x0400
	th32csSnapProcess       = 0x00000002
	invalidHandleValue      = ^uintptr(0)

	memCommit    = 0x1000
	pageNoAccess = 0x01
	pageGuard    = 0x100

	tokenAdjustPrivileges = 0x0020
	tokenQuery            = 0x0008
	sePrivilegeEnabled    = 0x0002
)

// processEntry32 mirrors the Win32 PROCESSENTRY32W structure. Go inserts the
// same padding as the C compiler (notably before DefaultHeapID on 64-bit), so
// unsafe.Sizeof yields the value Windows expects for dwSize.
type processEntry32 struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	Threads         uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

// memoryBasicInformation mirrors Win32 MEMORY_BASIC_INFORMATION. The trailing
// padding Go adds makes unsafe.Sizeof match the OS structure for this process's
// bitness, which is what VirtualQueryEx expects.
type memoryBasicInformation struct {
	BaseAddress       uintptr
	AllocationBase    uintptr
	AllocationProtect uint32
	RegionSize        uintptr
	State             uint32
	Protect           uint32
	Type              uint32
}

type luid struct {
	LowPart  uint32
	HighPart int32
}

type tokenPrivileges struct {
	PrivilegeCount uint32
	Luid           luid
	Attributes     uint32
}

// findProcesses returns the PIDs of every running process whose executable
// name matches one of names (case-insensitive).
func findProcesses(names ...string) []uint32 {
	var pids []uint32

	snap, _, _ := procSnapshot.Call(th32csSnapProcess, 0)
	if snap == invalidHandleValue || snap == 0 {
		return nil
	}
	defer procCloseHandle.Call(snap)

	var entry processEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	ok, _, _ := procProcess32First.Call(snap, uintptr(unsafe.Pointer(&entry)))
	for ok != 0 {
		exe := syscall.UTF16ToString(entry.ExeFile[:])
		for _, name := range names {
			if strings.EqualFold(exe, name) {
				pids = append(pids, entry.ProcessID)
				break
			}
		}
		ok, _, _ = procProcess32Next.Call(snap, uintptr(unsafe.Pointer(&entry)))
	}
	return pids
}

func openHandle(pid uint32) (uintptr, error) {
	h, _, err := procOpenProcess.Call(
		processVMRead|processQueryInformation,
		0,
		uintptr(pid),
	)
	if h == 0 {
		return 0, fmt.Errorf("osumem: OpenProcess(%d) failed: %w", pid, err)
	}
	return h, nil
}

// enableDebugPrivilege grants SeDebugPrivilege so elevated osu! instances can
// be read. Failures are silent: reading a same-privilege process still works.
func enableDebugPrivilege() {
	cur, _, _ := procGetCurrentProcess.Call()

	var token syscall.Handle
	r, _, _ := procOpenProcessToken.Call(
		cur,
		tokenAdjustPrivileges|tokenQuery,
		uintptr(unsafe.Pointer(&token)),
	)
	if r == 0 {
		return
	}
	defer procCloseHandle.Call(uintptr(token))

	var priv luid
	name, _ := syscall.UTF16PtrFromString("SeDebugPrivilege")
	if r, _, _ = procLookupPrivilegeValue.Call(0, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(&priv))); r == 0 {
		return
	}

	tp := tokenPrivileges{PrivilegeCount: 1, Luid: priv, Attributes: sePrivilegeEnabled}
	procAdjustTokenPrivileges.Call(uintptr(token), 0, uintptr(unsafe.Pointer(&tp)), 0, 0, 0)
}

// isProcess64Bit reports whether pid is a 64-bit process. On 64-bit Windows,
// IsWow64Process returns TRUE (1) only for 32-bit processes running under
// WOW64, so "not WOW64" means native 64-bit. This is how tosu distinguishes
// lazer (64-bit) from stable (32-bit).
func isProcess64Bit(pid uint32) bool {
	h, err := openHandle(pid)
	if err != nil {
		return false
	}
	defer procCloseHandle.Call(h)

	var wow64 int32
	r, _, _ := procIsWow64Process.Call(h, uintptr(unsafe.Pointer(&wow64)))
	if r == 0 {
		return false
	}
	return wow64 == 0
}

// ---------------------------------------------------------------------------
// process methods
// ---------------------------------------------------------------------------

func (p *process) close() {
	if p.handle != 0 {
		procCloseHandle.Call(p.handle)
		p.handle = 0
	}
}

// read copies size bytes from the target at addr.
func (p *process) read(addr uintptr, size int) ([]byte, error) {
	if size <= 0 {
		return nil, errors.New("osumem: invalid read size")
	}
	buf := make([]byte, size)
	var read uintptr
	r, _, err := procReadProcessMemory.Call(
		p.handle,
		addr,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(size),
		uintptr(unsafe.Pointer(&read)),
	)
	if r == 0 {
		return nil, fmt.Errorf("osumem: ReadProcessMemory(0x%X) failed: %w", addr, err)
	}
	return buf, nil
}

// imagePath returns the full path to the attached executable.
func (p *process) imagePath() (string, error) {
	buf := make([]uint16, 32768)
	size := uint32(len(buf))
	r, _, err := procFullImageName.Call(
		p.handle,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if r == 0 {
		return "", fmt.Errorf("osumem: QueryFullProcessImageNameW failed: %w", err)
	}
	return syscall.UTF16ToString(buf[:size]), nil
}

// forEachRegion walks the target's virtual address space, invoking fn for each
// committed, non-guarded, readable region. Iteration stops early if fn returns
// true.
func (p *process) forEachRegion(fn func(base, size uintptr) bool) {
	var mbi memoryBasicInformation
	addr := uintptr(0)

	for {
		r, _, _ := procVirtualQueryEx.Call(
			p.handle,
			addr,
			uintptr(unsafe.Pointer(&mbi)),
			unsafe.Sizeof(mbi),
		)
		if r == 0 {
			return
		}

		base := mbi.BaseAddress
		size := mbi.RegionSize

		if mbi.State == memCommit && mbi.Protect&(pageNoAccess|pageGuard) == 0 && size > 0 {
			if fn(base, size) {
				return
			}
		}

		next := base + size
		if next <= addr {
			return // guard against non-advancing regions
		}
		addr = next
	}
}
