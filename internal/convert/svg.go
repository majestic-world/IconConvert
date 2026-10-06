package convert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// isSVG reports whether data is an XML document whose root element is <svg>.
func isSVG(data []byte) bool {
	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			return false
		}
		switch t := tok.(type) {
		case xml.StartElement:
			return t.Name.Local == "svg"
		case xml.CharData:
			if len(bytes.TrimSpace(t)) > 0 {
				return false
			}
		}
	}
}

// vectorIcons draws a Vector Image afresh at every Icon Size, centred on a
// transparent square with its aspect ratio preserved.
func vectorIcons(data []byte) ([]image.Image, error) {
	icon, err := oksvg.ReadIconStream(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	w, h := icon.ViewBox.W, icon.ViewBox.H
	if w <= 0 || h <= 0 {
		return nil, errors.New("SVG has no usable size (missing viewBox, width or height)")
	}

	images := make([]image.Image, 0, len(iconSizes))
	for _, size := range iconSizes {
		side := float64(size)
		scale := side / max(w, h)
		tw, th := w*scale, h*scale
		icon.SetTarget((side-tw)/2, (side-th)/2, tw, th)

		dst := image.NewNRGBA(image.Rect(0, 0, size, size))
		scanner := rasterx.NewScannerGV(size, size, dst, dst.Bounds())
		icon.Draw(rasterx.NewDasher(size, size, scanner), 1)
		images = append(images, dst)
	}
	return images, nil
}
