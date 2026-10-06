package convert_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
)

var (
	red  = color.NRGBA{R: 255, A: 255}
	blue = color.NRGBA{B: 255, A: 255}
)

func TestRasterFormatsConvert(t *testing.T) {
	img := solidImage(64, 64, red)
	cases := map[string]func() []byte{
		"logo.jpg": func() []byte {
			var buf bytes.Buffer
			must(t, jpeg.Encode(&buf, img, nil))
			return buf.Bytes()
		},
		"logo.gif": func() []byte {
			var buf bytes.Buffer
			must(t, gif.Encode(&buf, img, nil))
			return buf.Bytes()
		},
		"logo.bmp": func() []byte {
			var buf bytes.Buffer
			must(t, bmp.Encode(&buf, img))
			return buf.Bytes()
		},
		"logo.tiff": func() []byte {
			var buf bytes.Buffer
			must(t, tiff.Encode(&buf, img, nil))
			return buf.Bytes()
		},
	}
	for name, encode := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			src := writeFile(t, dir, name, encode())

			outcome, _, stderr := run(t, src)

			if outcome.ExitCode != 0 {
				t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
			}
			entries := readIco(t, filepath.Join(dir, "favicon.ico"))
			if got, want := entrySizes(entries), []int{16, 32, 48, 64}; !reflect.DeepEqual(got, want) {
				t.Fatalf("icon sizes = %v, want %v", got, want)
			}
			assertColour(t, entries[0].img, red)
		})
	}
}

func TestCommittedFixtureFormatsConvert(t *testing.T) {
	cases := []struct {
		fixture string
		want    []int
	}{
		{"sample.webp", []int{16, 32, 48, 64, 128}}, // 150x100
		{"sample.avif", allIconSizes},               // 512x512
		{"sample.heic", allIconSizes},               // 512x512
	}
	for _, c := range cases {
		t.Run(c.fixture, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", c.fixture))
			must(t, err)
			dir := t.TempDir()
			src := writeFile(t, dir, c.fixture, data)

			outcome, _, stderr := run(t, src)

			if outcome.ExitCode != 0 {
				t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
			}
			if got := entrySizes(readIco(t, filepath.Join(dir, "favicon.ico"))); !reflect.DeepEqual(got, c.want) {
				t.Fatalf("icon sizes = %v, want %v", got, c.want)
			}
		})
	}
}

func TestAnimatedGIFConvertsFromFirstFrame(t *testing.T) {
	palette := color.Palette{red, blue}
	frame := func(index uint8) *image.Paletted {
		p := image.NewPaletted(image.Rect(0, 0, 32, 32), palette)
		for i := range p.Pix {
			p.Pix[i] = index
		}
		return p
	}
	var buf bytes.Buffer
	must(t, gif.EncodeAll(&buf, &gif.GIF{Image: []*image.Paletted{frame(0), frame(1)}, Delay: []int{10, 10}}))
	dir := t.TempDir()
	src := writeFile(t, dir, "anim.gif", buf.Bytes())

	run(t, src)

	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	assertColour(t, entries[0].img, red)
}

func TestSourceImageIsRecognisedByContentNotExtension(t *testing.T) {
	dir := t.TempDir()
	png := writePNG(t, dir, "logo.png", 64, 64)
	data, err := os.ReadFile(png)
	must(t, err)
	mislabelled := writeFile(t, dir, "photo.jpg", data)

	outcome, _, stderr := run(t, mislabelled)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
	}
	readIco(t, filepath.Join(dir, "favicon.ico"))
}

func TestBatchSourceImagesSharingABaseNameGetDistinctFavicons(t *testing.T) {
	dir := t.TempDir()
	logoPNG := writePNG(t, dir, "logo.png", 64, 64)
	var buf bytes.Buffer
	must(t, jpeg.Encode(&buf, solidImage(64, 64, blue), nil))
	logoJPG := writeFile(t, dir, "logo.jpg", buf.Bytes())

	run(t, logoPNG, logoJPG)

	readIco(t, filepath.Join(dir, "logo.ico"))
	readIco(t, filepath.Join(dir, "logo2.ico"))
}

func TestIcoSourceConvertsFromItsLargestEntry(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "app.ico", buildIco(t,
		bmpIcoEntry(solidImage(32, 32, red)),
		pngIcoEntry(t, solidImage(128, 128, blue)),
	))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
	}
	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got, want := entrySizes(entries), []int{16, 32, 48, 64, 128}; !reflect.DeepEqual(got, want) {
		t.Fatalf("icon sizes = %v, want %v", got, want)
	}
	assertColour(t, entries[0].img, blue)
}

func TestIcoSourceWithBitmapEntryConverts(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "app.ico", buildIco(t, bmpIcoEntry(solidImage(48, 48, red))))

	run(t, src)

	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got, want := entrySizes(entries), []int{16, 32, 48}; !reflect.DeepEqual(got, want) {
		t.Fatalf("icon sizes = %v, want %v", got, want)
	}
	assertColour(t, entries[0].img, red)
}

func TestIcoSourceInBatchIsLeftUntouched(t *testing.T) {
	dir := t.TempDir()
	original := buildIco(t, pngIcoEntry(t, solidImage(64, 64, red)))
	app := writeFile(t, dir, "app.ico", original)
	logo := writePNG(t, dir, "logo.png", 64, 64)

	run(t, app, logo)

	assertFileContent(t, app, string(original))
	readIco(t, filepath.Join(dir, "app2.ico"))
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func assertColour(t *testing.T, img image.Image, want color.NRGBA) {
	t.Helper()
	b := img.Bounds()
	got := color.NRGBAModel.Convert(img.At(b.Dx()/2, b.Dy()/2)).(color.NRGBA)
	const tolerance = 8 // JPEG is lossy
	diff := func(a, b uint8) int { return max(int(a)-int(b), int(b)-int(a)) }
	if diff(got.R, want.R) > tolerance || diff(got.G, want.G) > tolerance || diff(got.B, want.B) > tolerance || got.A != want.A {
		t.Fatalf("centre colour = %v, want %v", got, want)
	}
}

type testIcoEntry struct {
	side    int
	bpp     uint16
	payload []byte
}

func pngIcoEntry(t *testing.T, img image.Image) testIcoEntry {
	var buf bytes.Buffer
	must(t, png.Encode(&buf, img))
	return testIcoEntry{side: img.Bounds().Dx(), bpp: 32, payload: buf.Bytes()}
}

// bmpIcoEntry encodes img as a 32-bit BI_RGB DIB, the classic ICO entry form:
// doubled height, bottom-up BGRA rows, followed by a 1-bit AND mask.
func bmpIcoEntry(img *image.NRGBA) testIcoEntry {
	side := img.Bounds().Dx()
	maskStride := ((side + 31) / 32) * 4
	payload := make([]byte, 40, 40+side*side*4+maskStride*side)
	binary.LittleEndian.PutUint32(payload[0:], 40)
	binary.LittleEndian.PutUint32(payload[4:], uint32(side))
	binary.LittleEndian.PutUint32(payload[8:], uint32(side*2))
	binary.LittleEndian.PutUint16(payload[12:], 1)
	binary.LittleEndian.PutUint16(payload[14:], 32)
	for y := side - 1; y >= 0; y-- {
		for x := range side {
			c := img.NRGBAAt(x, y)
			payload = append(payload, c.B, c.G, c.R, c.A)
		}
	}
	payload = append(payload, make([]byte, maskStride*side)...)
	return testIcoEntry{side: side, bpp: 32, payload: payload}
}

func buildIco(t *testing.T, entries ...testIcoEntry) []byte {
	t.Helper()
	header := make([]byte, 6+16*len(entries))
	binary.LittleEndian.PutUint16(header[2:], 1)
	binary.LittleEndian.PutUint16(header[4:], uint16(len(entries)))
	offset := len(header)
	var body []byte
	for i, e := range entries {
		d := header[6+16*i:]
		d[0], d[1] = byte(e.side%256), byte(e.side%256)
		binary.LittleEndian.PutUint16(d[4:], 1)
		binary.LittleEndian.PutUint16(d[6:], e.bpp)
		binary.LittleEndian.PutUint32(d[8:], uint32(len(e.payload)))
		binary.LittleEndian.PutUint32(d[12:], uint32(offset+len(body)))
		body = append(body, e.payload...)
	}
	return append(header, body...)
}
