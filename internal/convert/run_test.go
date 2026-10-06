package convert_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"IconConvert/internal/convert"
)

var allIconSizes = []int{16, 32, 48, 64, 128, 256}

func TestLonePNGProducesFaviconWithEveryIconSize(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "logo.png", 512, 512)

	outcome, _, _ := run(t, src)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", outcome.ExitCode)
	}
	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got := entrySizes(entries); !reflect.DeepEqual(got, allIconSizes) {
		t.Fatalf("icon sizes = %v, want %v", got, allIconSizes)
	}
}

func TestExistingFaviconIsKeptAndNextFreeNumberIsUsed(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "logo.png", 64, 64)
	writeFile(t, dir, "favicon.ico", []byte("previous"))
	writeFile(t, dir, "favicon3.ico", []byte("previous3"))

	run(t, src)

	assertFileContent(t, filepath.Join(dir, "favicon.ico"), "previous")
	assertFileContent(t, filepath.Join(dir, "favicon3.ico"), "previous3")
	readIco(t, filepath.Join(dir, "favicon2.ico"))
	assertMissing(t, filepath.Join(dir, "favicon4.ico"))
}

func TestBatchNamesEachFaviconAfterItsSourceImage(t *testing.T) {
	dir := t.TempDir()
	logo := writePNG(t, dir, "logo.png", 64, 64)
	banner := writePNG(t, dir, "banner.png", 64, 64)
	writeFile(t, dir, "banner.ico", []byte("previous"))

	outcome, _, _ := run(t, logo, banner)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0", outcome.ExitCode)
	}
	readIco(t, filepath.Join(dir, "logo.ico"))
	assertFileContent(t, filepath.Join(dir, "banner.ico"), "previous")
	readIco(t, filepath.Join(dir, "banner2.ico"))
	assertMissing(t, filepath.Join(dir, "favicon.ico"))
}

func TestFaviconIsWrittenInTheSourceImageFolder(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "assets")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	src := writePNG(t, sub, "logo.png", 64, 64)
	t.Chdir(dir)

	run(t, src)

	readIco(t, filepath.Join(sub, "favicon.ico"))
	assertMissing(t, filepath.Join(dir, "favicon.ico"))
}

func TestSmallRasterImageSkipsLargerIconSizesWithWarning(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "logo.png", 64, 64)

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 || !outcome.Warned {
		t.Fatalf("outcome = %+v, want exit 0 with warning", outcome)
	}
	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got, want := entrySizes(entries), []int{16, 32, 48, 64}; !reflect.DeepEqual(got, want) {
		t.Fatalf("icon sizes = %v, want %v", got, want)
	}
	if !strings.Contains(stderr, "128") || !strings.Contains(stderr, "256") {
		t.Fatalf("warning should name the Skipped Sizes 128 and 256, got:\n%s", stderr)
	}
}

func TestFullSizedImageRaisesNoWarning(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "logo.png", 256, 256)

	outcome, _, _ := run(t, src)

	if outcome.Warned {
		t.Fatal("a 256 px Source Image should not warn")
	}
}

func TestWideImageIsCentredOnTransparentSquareWithoutSkippingSizes(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "banner.png", 300, 100)

	run(t, src)

	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got := entrySizes(entries); !reflect.DeepEqual(got, allIconSizes) {
		t.Fatalf("icon sizes = %v, want %v", got, allIconSizes)
	}
	largest := entries[len(entries)-1].img
	assertAlpha(t, largest, 128, 4, 0)     // padding above the content
	assertAlpha(t, largest, 128, 251, 0)   // padding below the content
	assertAlpha(t, largest, 128, 128, 255) // content centre
	assertAlpha(t, largest, 2, 128, 255)   // content reaches the left edge: not cropped
}

func TestRasterImageBelowSixteenPixelsIsRejected(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "tiny.png", 10, 12)

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", outcome.ExitCode)
	}
	if !strings.Contains(stderr, "tiny.png") {
		t.Fatalf("error should name the Source Image, got:\n%s", stderr)
	}
	assertMissing(t, filepath.Join(dir, "favicon.ico"))
}

func TestSixteenPixelImageGetsOnlyTheSmallestIconSize(t *testing.T) {
	dir := t.TempDir()
	src := writePNG(t, dir, "strip.png", 16, 8)

	run(t, src)

	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got, want := entrySizes(entries), []int{16}; !reflect.DeepEqual(got, want) {
		t.Fatalf("icon sizes = %v, want %v", got, want)
	}
}

// run invokes the runner and captures its output.
func run(t *testing.T, paths ...string) (convert.Outcome, string, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	outcome := convert.Run(paths, &stdout, &stderr)
	return outcome, stdout.String(), stderr.String()
}

func solidImage(w, h int, c color.Color) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.Set(x, y, c)
		}
	}
	return img
}

func writeFile(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func writePNG(t *testing.T, dir, name string, w, h int) string {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, solidImage(w, h, color.NRGBA{R: 200, G: 30, B: 30, A: 255})); err != nil {
		t.Fatal(err)
	}
	return writeFile(t, dir, name, buf.Bytes())
}

type icoEntry struct {
	declaredSize int
	img          image.Image
}

// readIco parses an .ico file independently of the production encoder.
func readIco(t *testing.T, path string) []icoEntry {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading favicon: %v", err)
	}
	if len(data) < 6 || binary.LittleEndian.Uint16(data[0:]) != 0 || binary.LittleEndian.Uint16(data[2:]) != 1 {
		t.Fatalf("%s: not an icon file", path)
	}
	count := int(binary.LittleEndian.Uint16(data[4:]))
	entries := make([]icoEntry, 0, count)
	for i := range count {
		dirEntry := data[6+16*i : 6+16*(i+1)]
		size := int(dirEntry[0])
		if size == 0 {
			size = 256
		}
		if int(dirEntry[1]) != size%256 {
			t.Fatalf("entry %d: width %d and height %d differ", i, dirEntry[0], dirEntry[1])
		}
		length := binary.LittleEndian.Uint32(dirEntry[8:])
		offset := binary.LittleEndian.Uint32(dirEntry[12:])
		img, err := png.Decode(bytes.NewReader(data[offset : offset+length]))
		if err != nil {
			t.Fatalf("entry %d: decoding PNG payload: %v", i, err)
		}
		if b := img.Bounds(); b.Dx() != size || b.Dy() != size {
			t.Fatalf("entry %d: declared %d px but image is %dx%d", i, size, b.Dx(), b.Dy())
		}
		entries = append(entries, icoEntry{declaredSize: size, img: img})
	}
	return entries
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s was modified", filepath.Base(path))
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("%s should not exist (stat err: %v)", filepath.Base(path), err)
	}
}

func assertAlpha(t *testing.T, img image.Image, x, y int, want uint8) {
	t.Helper()
	got := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA).A
	if got != want {
		t.Fatalf("alpha at (%d,%d) = %d, want %d", x, y, got, want)
	}
}

func entrySizes(entries []icoEntry) []int {
	sizes := make([]int, len(entries))
	for i, e := range entries {
		sizes[i] = e.declaredSize
	}
	return sizes
}
