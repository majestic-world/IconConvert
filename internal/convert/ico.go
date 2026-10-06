package convert

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/png"
	"io"
)

const (
	icoHeaderLen   = 6
	icoDirEntryLen = 16
)

// encodeIco writes one square image per Icon Size as a PNG-compressed Favicon, in the given order.
func encodeIco(w io.Writer, images []image.Image) error {
	payloads := make([][]byte, len(images))
	for i, img := range images {
		var buf bytes.Buffer
		if err := png.Encode(&buf, img); err != nil {
			return err
		}
		payloads[i] = buf.Bytes()
	}

	header := make([]byte, icoHeaderLen+icoDirEntryLen*len(images))
	binary.LittleEndian.PutUint16(header[2:], 1) // type: icon
	binary.LittleEndian.PutUint16(header[4:], uint16(len(images)))
	offset := len(header)
	for i, img := range images {
		entry := header[icoHeaderLen+icoDirEntryLen*i:]
		side := img.Bounds().Dx()
		entry[0] = byte(side % 256) // 256 is stored as 0
		entry[1] = byte(side % 256)
		binary.LittleEndian.PutUint16(entry[4:], 1)  // colour planes
		binary.LittleEndian.PutUint16(entry[6:], 32) // bits per pixel
		binary.LittleEndian.PutUint32(entry[8:], uint32(len(payloads[i])))
		binary.LittleEndian.PutUint32(entry[12:], uint32(offset))
		offset += len(payloads[i])
	}

	if _, err := w.Write(header); err != nil {
		return err
	}
	for _, p := range payloads {
		if _, err := w.Write(p); err != nil {
			return err
		}
	}
	return nil
}
