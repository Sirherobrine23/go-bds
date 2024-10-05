//go:build windows

package projfs

import (
	"encoding/binary"
	"syscall"
	"unsafe"
)

type IProvider interface {
	CancelCommand(commandID int32)
	StartDirectoryEnumeration()
}

func SetGUID(b []byte, guid *syscall.GUID) {
	guid.Data1 = binary.LittleEndian.Uint32(b[0:4])
	guid.Data2 = binary.LittleEndian.Uint16(b[4:6])
	guid.Data3 = binary.LittleEndian.Uint16(b[6:8])
	guid.Data4 = ([8]byte)(b[8:16])
}

func GetPointer(str string) uintptr {
	ptr, err := syscall.UTF16PtrFromString(str)
	if err != nil {
		return 0
	}
	return uintptr(unsafe.Pointer(ptr))
}

func GetString(str uintptr) string {
	if p := (*uint16)(unsafe.Pointer(str)); p != nil {
		n, end := 0, unsafe.Pointer(p)
		for *(*uint16)(end) != 0 {
			end = unsafe.Pointer(uintptr(end) + unsafe.Sizeof(*p))
			n++
		}
		return syscall.UTF16ToString(unsafe.Slice(p, n))
	}
	return ""
}
