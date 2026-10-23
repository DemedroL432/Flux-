package main

import (
	"bytes"
	"encoding/binary"
	_ "embed"
)

//go:embed build/appicon.png
var embeddedAppiconPNG []byte

// pngIHDRSize returns width/height from PNG IHDR chunk (must be square for tray).
func pngIHDRSize(png []byte) (w, h uint32, ok bool) {
	if len(png) < 24 {
		return 0, 0, false
	}
	if string(png[0:8]) != "\x89PNG\r\n\x1a\n" {
		return 0, 0, false
	}
	w = binary.BigEndian.Uint32(png[16:20])
	h = binary.BigEndian.Uint32(png[20:24])
	return w, h, true
}

// icoDimByte: ICO directory uses 0..255; 0 means 256. For larger PNG-in-ICO (Vista+), 0 still used for >=256.
func icoDimByte(dim uint32) byte {
	if dim == 256 || dim > 256 {
		return 0
	}
	return byte(dim)
}

// pngBytesToICO упаковывает PNG в ICO (формат с PNG-пейлоадом, Vista+).
func pngBytesToICO(png []byte) []byte {
	if len(png) < 8 {
		return nil
	}
	w, h, ok := pngIHDRSize(png)
	if !ok || w != h {
		return nil
	}
	n := len(png)
	off := 6 + 16
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.LittleEndian, uint16(0)) // reserved
	_ = binary.Write(buf, binary.LittleEndian, uint16(1)) // type: icon
	_ = binary.Write(buf, binary.LittleEndian, uint16(1)) // count
	// ICONDIRENTRY
	buf.WriteByte(icoDimByte(w))
	buf.WriteByte(icoDimByte(h))
	buf.WriteByte(0) // color count
	buf.WriteByte(0) // reserved
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))   // planes
	_ = binary.Write(buf, binary.LittleEndian, uint16(32))  // bit count
	_ = binary.Write(buf, binary.LittleEndian, uint32(n))   // size
	_ = binary.Write(buf, binary.LittleEndian, uint32(off)) // offset
	_, _ = buf.Write(png)
	return buf.Bytes()
}

func trayIconICOBytes() []byte {
	if len(embeddedAppiconPNG) == 0 {
		return nil
	}
	return pngBytesToICO(embeddedAppiconPNG)
}
