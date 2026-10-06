package convert

import (
	"bytes"
	"encoding/binary"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"slices"

	"github.com/gen2brain/avif"
	"github.com/gen2brain/heic"
	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

var errUnsupportedFormat = errors.New("not a supported image format (PNG, JPEG, GIF, BMP, TIFF, WebP, ICO, AVIF, HEIC, SVG)")

// decodeRaster decodes a raster Source Image by sniffing its signature bytes,
// whatever its name suffix.
func decodeRaster(data []byte) (image.Image, error) {
	var decode func(io.Reader) (image.Image, error)
	switch {
	case bytes.HasPrefix(data, []byte("\x89PNG\r\n\x1a\n")):
		decode = png.Decode
	case bytes.HasPrefix(data, []byte("\xff\xd8\xff")):
		decode = jpeg.Decode
	case bytes.HasPrefix(data, []byte("GIF87a")), bytes.HasPrefix(data, []byte("GIF89a")):
		decode = gif.Decode // first frame
	case bytes.HasPrefix(data, []byte("BM")):
		decode = bmp.Decode
	case bytes.HasPrefix(data, []byte("II*\x00")), bytes.HasPrefix(data, []byte("MM\x00*")):
		decode = tiff.Decode
	case len(data) >= 12 && string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		decode = webp.Decode
	case bytes.HasPrefix(data, []byte("\x00\x00\x01\x00")):
		return decodeIco(data)
	default:
		switch isoBrand(data) {
		case "avif":
			decode = func(r io.Reader) (image.Image, error) { return avif.Decode(r) }
		case "heic":
			decode = heic.Decode
		default:
			return nil, errUnsupportedFormat
		}
	}
	return decode(bytes.NewReader(data))
}

var (
	avifBrands = []string{"avif", "avis"}
	heicBrands = []string{"heic", "heix", "heim", "heis", "hevc", "hevx", "hevm", "hevs"}
)

// isoBrand classifies an ISO-BMFF file ("ftyp" box first) as "avif", "heic"
// or "" from its major and compatible brands.
func isoBrand(data []byte) string {
	if len(data) < 16 || string(data[4:8]) != "ftyp" {
		return ""
	}
	boxLen := min(int(binary.BigEndian.Uint32(data[0:4])), len(data))
	brands := []string{string(data[8:12])}
	for i := 16; i+4 <= boxLen; i += 4 {
		brands = append(brands, string(data[i:i+4]))
	}
	for _, b := range brands {
		if slices.Contains(avifBrands, b) {
			return "avif"
		}
	}
	for _, b := range brands {
		if slices.Contains(heicBrands, b) {
			return "heic"
		}
	}
	return ""
}
