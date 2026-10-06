package convert

import (
	"bytes"
	"encoding/xml"
	"errors"
	"image"
	"regexp"
	"strings"

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
	data = pathDataAttr.ReplaceAllFunc(data, func(attr []byte) []byte {
		open := bytes.IndexAny(attr, `"'`)
		value := normalizePathData(attr[open+1 : len(attr)-1])
		out := make([]byte, 0, open+len(value)+2)
		out = append(out, attr[:open+1]...)
		out = append(out, value...)
		return append(out, attr[len(attr)-1])
	})
	// A standalone SVG has no CSS context, so currentColor takes the initial
	// CSS colour: black.
	icon, err := oksvg.ReadReplacingCurrentColor(bytes.NewReader(data), "black")
	if err != nil {
		return nil, err
	}
	w, h := icon.ViewBox.W, icon.ViewBox.H
	if w <= 0 || h <= 0 {
		return nil, errors.New("SVG has no usable size (missing viewBox, width or height)")
	}

	// The rasterizer applies stroke widths and dashes in output pixels; they
	// are SVG units, so scale them with each Icon Size.
	strokes := make([]oksvg.PathStyle, len(icon.SVGPaths))
	for i, p := range icon.SVGPaths {
		strokes[i] = p.PathStyle
	}

	images := make([]image.Image, 0, len(iconSizes))
	var largest *image.NRGBA
	for _, size := range iconSizes {
		side := float64(size)
		scale := side / max(w, h)
		tw, th := w*scale, h*scale
		icon.SetTarget((side-tw)/2, (side-th)/2, tw, th)
		for i := range icon.SVGPaths {
			p, orig := &icon.SVGPaths[i], strokes[i]
			p.LineWidth = orig.LineWidth * scale
			p.DashOffset = orig.DashOffset * scale
			p.Dash = make([]float64, len(orig.Dash))
			for j, d := range orig.Dash {
				p.Dash[j] = d * scale
			}
		}

		dst := image.NewNRGBA(image.Rect(0, 0, size, size))
		scanner := rasterx.NewScannerGV(size, size, dst, dst.Bounds())
		icon.Draw(rasterx.NewDasher(size, size, scanner), 1)
		images = append(images, dst)
		largest = dst
	}
	if isBlank(largest) {
		return nil, errors.New("SVG drew nothing; it may rely on features IconConvert cannot render (such as text)")
	}
	return images, nil
}

func isBlank(img *image.NRGBA) bool {
	for i := 3; i < len(img.Pix); i += 4 {
		if img.Pix[i] != 0 {
			return false
		}
	}
	return true
}

// pathDataAttr matches a quoted `d` attribute.
var pathDataAttr = regexp.MustCompile(`\sd\s*=\s*("[^"]*"|'[^']*')`)

// normalizePathData rewrites SVG path data into a form the rasterizer reads
// correctly: every command and argument separated by a space, and every
// implicitly repeated command written out. Minified SVGs pack arc flags
// without separators ("a12 12 0 1012 12" is rx ry rotation large-arc=1
// sweep=0 x=12 y=12) and chain arcs without repeating "a", both of which the
// rasterizer misreads. Unparseable data is returned unchanged.
func normalizePathData(d []byte) []byte {
	out := make([]byte, 0, len(d)+len(d)/2)
	var cmd byte
	arg := 0
	for i := 0; i < len(d); {
		c := d[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == ',':
			i++
		case strings.IndexByte("MmLlHhVvCcSsQqTtAaZz", c) >= 0:
			cmd, arg = c, 0
			out = append(out, ' ', c)
			i++
		default:
			end := i + 1
			isFlag := (cmd == 'a' || cmd == 'A') && (arg%7 == 3 || arg%7 == 4)
			if !isFlag || (c != '0' && c != '1') {
				end = scanNumber(d, i)
				if end == i {
					return d
				}
			}
			if n := pathArgCounts[cmd]; n > 0 && arg > 0 && arg%n == 0 {
				// The rasterizer mishandles implicit repeats; spell the
				// command out. A repeated moveto continues as lineto.
				repeat := cmd
				switch repeat {
				case 'M':
					repeat = 'L'
				case 'm':
					repeat = 'l'
				}
				out = append(out, ' ', repeat)
			}
			out = append(out, ' ')
			out = append(out, d[i:end]...)
			arg++
			i = end
		}
	}
	return bytes.TrimSpace(out)
}

// pathArgCounts is the number of arguments each path command consumes.
var pathArgCounts = map[byte]int{
	'M': 2, 'm': 2, 'L': 2, 'l': 2, 'H': 1, 'h': 1, 'V': 1, 'v': 1,
	'C': 6, 'c': 6, 'S': 4, 's': 4, 'Q': 4, 'q': 4, 'T': 2, 't': 2,
	'A': 7, 'a': 7,
}

// scanNumber returns the end of the SVG number starting at s[i], or i when
// there is none.
func scanNumber(s []byte, i int) int {
	isDigit := func(j int) bool { return j < len(s) && s[j] >= '0' && s[j] <= '9' }
	j := i
	if j < len(s) && (s[j] == '+' || s[j] == '-') {
		j++
	}
	digits := 0
	for ; isDigit(j); j++ {
		digits++
	}
	if j < len(s) && s[j] == '.' && (digits > 0 || isDigit(j+1)) {
		for j++; isDigit(j); j++ {
			digits++
		}
	}
	if digits == 0 {
		return i
	}
	if j < len(s) && (s[j] == 'e' || s[j] == 'E') {
		k := j + 1
		if k < len(s) && (s[k] == '+' || s[k] == '-') {
			k++
		}
		if isDigit(k) {
			for ; isDigit(k); k++ {
			}
			j = k
		}
	}
	return j
}
