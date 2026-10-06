package convert

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/png"
)

var errCorruptIco = errors.New("corrupt ICO file")

// decodeIco decodes the largest image embedded in an ICO Source Image. Entries may
// be PNG-compressed or classic DIBs (1, 4, 8, 24 or 32 bits per pixel).
func decodeIco(data []byte) (image.Image, error) {
	if len(data) < icoHeaderLen {
		return nil, errCorruptIco
	}
	count := int(binary.LittleEndian.Uint16(data[4:]))
	if count == 0 || len(data) < icoHeaderLen+icoDirEntryLen*count {
		return nil, errCorruptIco
	}

	best, bestArea, bestBpp := -1, 0, 0
	for i := range count {
		e := data[icoHeaderLen+icoDirEntryLen*i:]
		w, h := int(e[0]), int(e[1])
		if w == 0 {
			w = 256
		}
		if h == 0 {
			h = 256
		}
		bpp := int(binary.LittleEndian.Uint16(e[6:]))
		if area := w * h; area > bestArea || (area == bestArea && bpp > bestBpp) {
			best, bestArea, bestBpp = i, area, bpp
		}
	}

	e := data[icoHeaderLen+icoDirEntryLen*best:]
	length := int(binary.LittleEndian.Uint32(e[8:]))
	offset := int(binary.LittleEndian.Uint32(e[12:]))
	if offset < 0 || length <= 0 || offset+length > len(data) || offset+length < offset {
		return nil, errCorruptIco
	}
	payload := data[offset : offset+length]
	if bytes.HasPrefix(payload, []byte("\x89PNG\r\n\x1a\n")) {
		return png.Decode(bytes.NewReader(payload))
	}
	return decodeDIB(payload)
}

// decodeDIB decodes a BITMAPINFOHEADER-based icon bitmap: XOR colour data
// (bottom-up, height doubled) followed by a 1-bit AND transparency mask.
func decodeDIB(p []byte) (image.Image, error) {
	if len(p) < 40 {
		return nil, errCorruptIco
	}
	headerLen := int(binary.LittleEndian.Uint32(p[0:]))
	w := int(int32(binary.LittleEndian.Uint32(p[4:])))
	h := int(int32(binary.LittleEndian.Uint32(p[8:]))) / 2
	bpp := int(binary.LittleEndian.Uint16(p[14:]))
	compression := binary.LittleEndian.Uint32(p[16:])
	if headerLen < 40 || headerLen > len(p) || w <= 0 || h <= 0 || w > 1024 || h > 1024 || compression != 0 {
		return nil, errCorruptIco
	}

	var palette []color.NRGBA
	if bpp <= 8 {
		switch bpp {
		case 1, 4, 8:
		default:
			return nil, errCorruptIco
		}
		n := int(binary.LittleEndian.Uint32(p[32:]))
		if n == 0 {
			n = 1 << bpp
		}
		start := headerLen
		if start+4*n > len(p) {
			return nil, errCorruptIco
		}
		palette = make([]color.NRGBA, n)
		for i := range palette {
			q := p[start+4*i:]
			palette[i] = color.NRGBA{R: q[2], G: q[1], B: q[0], A: 255}
		}
	} else if bpp != 24 && bpp != 32 {
		return nil, errCorruptIco
	}

	xorStart := headerLen + 4*len(palette)
	xorStride := ((w*bpp + 31) / 32) * 4
	andStart := xorStart + xorStride*h
	andStride := ((w + 31) / 32) * 4
	if xorStart+xorStride*h > len(p) {
		return nil, errCorruptIco
	}
	hasMask := andStart+andStride*h <= len(p)

	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	anyAlpha := false
	for y := range h {
		row := p[xorStart+xorStride*(h-1-y):]
		for x := range w {
			var c color.NRGBA
			switch bpp {
			case 32:
				c = color.NRGBA{R: row[4*x+2], G: row[4*x+1], B: row[4*x], A: row[4*x+3]}
				anyAlpha = anyAlpha || c.A != 0
			case 24:
				c = color.NRGBA{R: row[3*x+2], G: row[3*x+1], B: row[3*x], A: 255}
			default:
				bitPos := x * bpp
				idx := int(row[bitPos/8]>>(8-bpp-bitPos%8)) & (1<<bpp - 1)
				if idx >= len(palette) {
					return nil, errCorruptIco
				}
				c = palette[idx]
			}
			img.SetNRGBA(x, y, c)
		}
	}

	// 32-bit entries carry their own alpha; older ones rely on the AND mask.
	if bpp == 32 && anyAlpha {
		return img, nil
	}
	for y := range h {
		mask := []byte(nil)
		if hasMask {
			mask = p[andStart+andStride*(h-1-y):]
		}
		for x := range w {
			c := img.NRGBAAt(x, y)
			c.A = 255
			if mask != nil && mask[x/8]&(0x80>>(x%8)) != 0 {
				c.A = 0
			}
			img.SetNRGBA(x, y, c)
		}
	}
	return img, nil
}
