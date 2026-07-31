//go:build windows

package eco

import (
	"errors"
	"syscall"
	"unsafe"
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

var (
	crypt32                = syscall.NewLazyDLL("crypt32.dll")
	kernel32               = syscall.NewLazyDLL("kernel32.dll")
	procCryptProtectData   = crypt32.NewProc("CryptProtectData")
	procCryptUnprotectData = crypt32.NewProc("CryptUnprotectData")
	procLocalFree          = kernel32.NewProc("LocalFree")
)

func protectLocalKey(in []byte) ([]byte, error)   { return dpapi(in, true) }
func unprotectLocalKey(in []byte) ([]byte, error) { return dpapi(in, false) }

func dpapi(in []byte, protect bool) ([]byte, error) {
	if len(in) == 0 {
		return nil, errors.New("empty key material")
	}
	ib := dataBlob{cbData: uint32(len(in)), pbData: &in[0]}
	var ob dataBlob
	var r uintptr
	if protect {
		r, _, _ = procCryptProtectData.Call(uintptr(unsafe.Pointer(&ib)), 0, 0, 0, 0, 0x01, uintptr(unsafe.Pointer(&ob)))
	} else {
		r, _, _ = procCryptUnprotectData.Call(uintptr(unsafe.Pointer(&ib)), 0, 0, 0, 0, 0x01, uintptr(unsafe.Pointer(&ob)))
	}
	if r == 0 {
		return nil, syscall.GetLastError()
	}
	defer procLocalFree.Call(uintptr(unsafe.Pointer(ob.pbData)))
	out := make([]byte, ob.cbData)
	copy(out, unsafe.Slice(ob.pbData, ob.cbData))
	return out, nil
}
