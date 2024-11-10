package screenshot

import (
	"fmt"
	"image"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/gonutz/w32"
)

func CaptureRect(handle syscall.Handle, rect w32.RECT) (*image.RGBA, error) {
	hDC := GetDC(0)
	getWindowRect.Call(uintptr(0), uintptr(hDC), uintptr(w32.PW_CLIENTONLY))
	if hDC == 0 {
		return nil, fmt.Errorf("Could not Get primary display err:%d.\n", GetLastError())
	}
	defer ReleaseDC(0, hDC)

	m_hDC := CreateCompatibleDC(hDC)
	if m_hDC == 0 {
		return nil, fmt.Errorf("Could not Create Compatible DC err:%d.\n", GetLastError())
	}
	defer DeleteDC(m_hDC)

	x, y := int(rect.Left), int(rect.Top)
	width := rect.Right - rect.Left
	height := rect.Bottom - rect.Top
	expectedSize := width * height * 4

	bt := BITMAPINFO{}
	bt.BmiHeader.BiSize = uint32(reflect.TypeOf(bt.BmiHeader).Size())
	bt.BmiHeader.BiWidth = int32(width)
	bt.BmiHeader.BiHeight = -int32(height)
	bt.BmiHeader.BiPlanes = 1
	bt.BmiHeader.BiBitCount = 32
	bt.BmiHeader.BiCompression = BI_RGB
	bt.BmiHeader.BiSizeImage = 0

	var ptr unsafe.Pointer

	m_hBmp := CreateDIBSection(m_hDC, &bt, DIB_RGB_COLORS, &ptr, 0, 0)

	if m_hBmp == 0 {
		return nil, fmt.Errorf("Could not Create DIB Section err:%d.\n", GetLastError())
	}
	if m_hBmp == InvalidParameter {
		return nil, fmt.Errorf("One or more of the input parameters is invalid while calling CreateDIBSection.\n")
	}
	defer DeleteObject(HGDIOBJ(m_hBmp))

	obj := SelectObject(m_hDC, HGDIOBJ(m_hBmp))
	if obj == 0 {
		return nil, fmt.Errorf("error occurred and the selected object is not a region err:%d.\n", GetLastError())
	}
	if obj == 0xffffffff { //GDI_ERROR
		return nil, fmt.Errorf("GDI_ERROR while calling SelectObject err:%d.\n", GetLastError())
	}
	defer DeleteObject(obj)

	if !BitBlt(m_hDC, 0, 0, int(width), int(height), hDC, x, y, SRCCOPY) {
		return nil, fmt.Errorf("BitBlt failed err:%d.\n", GetLastError())
	}

	var slice []byte = make([]byte, width*height*4)
	hdrp := (*reflect.SliceHeader)(unsafe.Pointer(&slice))
	hdrp.Data = uintptr(ptr)
	hdrp.Len = x * y * 4
	hdrp.Cap = x * y * 4

	imageBytes := make([]byte, expectedSize)

	copy(imageBytes, (*[1 << 32]byte)(ptr)[:width*height*4:width*height*4])

	for i := 0; i < len(imageBytes); i += 4 {
		imageBytes[i], imageBytes[i+2] = imageBytes[i+2], imageBytes[i]
		imageBytes[i+3] = 255
	}

	img := &image.RGBA{Pix: imageBytes, Stride: 4 * int(width), Rect: image.Rect(0, 0, int(width), int(height))}
	return img, nil
}

func GetDeviceCaps(hdc HDC, index int) int {
	ret, _, _ := procGetDeviceCaps.Call(
		uintptr(hdc),
		uintptr(index))

	return int(ret)
}

func GetDC(hwnd syscall.Handle) HDC {
	ret, _, _ := procGetDC.Call(
		uintptr(hwnd))

	return HDC(ret)
}

func ReleaseDC(hwnd syscall.Handle, hDC HDC) bool {
	ret, _, _ := procReleaseDC.Call(
		uintptr(hwnd),
		uintptr(hDC))

	return ret != 0
}

func DeleteDC(hdc HDC) bool {
	ret, _, _ := procDeleteDC.Call(
		uintptr(hdc))

	return ret != 0
}

func GetLastError() uint32 {
	ret, _, _ := procGetLastError.Call()
	return uint32(ret)
}

func BitBlt(hdcDest HDC, nXDest, nYDest, nWidth, nHeight int, hdcSrc HDC, nXSrc, nYSrc int, dwRop uint) bool {
	ret, _, _ := procBitBlt.Call(
		uintptr(hdcDest),
		uintptr(nXDest),
		uintptr(nYDest),
		uintptr(nWidth),
		uintptr(nHeight),
		uintptr(hdcSrc),
		uintptr(nXSrc),
		uintptr(nYSrc),
		uintptr(dwRop))

	return ret != 0
}

func SelectObject(hdc HDC, hgdiobj HGDIOBJ) HGDIOBJ {
	ret, _, _ := procSelectObject.Call(
		uintptr(hdc),
		uintptr(hgdiobj))

	if ret == 0 {
		panic("SelectObject failed")
	}

	return HGDIOBJ(ret)
}

func DeleteObject(hObject HGDIOBJ) bool {
	ret, _, _ := procDeleteObject.Call(
		uintptr(hObject))

	return ret != 0
}

func CreateDIBSection(hdc HDC, pbmi *BITMAPINFO, iUsage uint, ppvBits *unsafe.Pointer, hSection HANDLE, dwOffset uint) HBITMAP {
	ret, _, _ := procCreateDIBSection.Call(
		uintptr(hdc),
		uintptr(unsafe.Pointer(pbmi)),
		uintptr(iUsage),
		uintptr(unsafe.Pointer(ppvBits)),
		uintptr(hSection),
		uintptr(dwOffset))

	return HBITMAP(ret)
}

func CreateCompatibleDC(hdc HDC) HDC {
	ret, _, _ := procCreateCompatibleDC.Call(
		uintptr(hdc))

	if ret == 0 {
		panic("Create compatible DC failed")
	}

	return HDC(ret)
}

type (
	HANDLE  uintptr
	HGDIOBJ HANDLE
	HDC     HANDLE
	HBITMAP HANDLE
)

type BITMAPINFO struct {
	BmiHeader BITMAPINFOHEADER
	BmiColors *RGBQUAD
}

type BITMAPINFOHEADER struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

type RGBQUAD struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

const (
	HORZRES          = 8
	VERTRES          = 10
	DESKTOPHORZRES   = 118
	DESKTOPVERTRES   = 117
	BI_RGB           = 0
	InvalidParameter = 2
	DIB_RGB_COLORS   = 0
	SRCCOPY          = 0x00CC0020
)

var (
	modgdi32               = syscall.NewLazyDLL("gdi32.dll")
	moduser32              = syscall.NewLazyDLL("user32.dll")
	modkernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetDC              = moduser32.NewProc("GetDC")
	procReleaseDC          = moduser32.NewProc("ReleaseDC")
	procDeleteDC           = modgdi32.NewProc("DeleteDC")
	procBitBlt             = modgdi32.NewProc("BitBlt")
	procDeleteObject       = modgdi32.NewProc("DeleteObject")
	procSelectObject       = modgdi32.NewProc("SelectObject")
	procCreateDIBSection   = modgdi32.NewProc("CreateDIBSection")
	procCreateCompatibleDC = modgdi32.NewProc("CreateCompatibleDC")
	procGetDeviceCaps      = modgdi32.NewProc("GetDeviceCaps")
	procGetLastError       = modkernel32.NewProc("GetLastError")
	getWindowRect          = moduser32.NewProc("GetWindowRect")
)
