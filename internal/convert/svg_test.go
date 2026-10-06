package convert_test

import (
	"image/color"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const wideRedSVG = `<?xml version="1.0" encoding="UTF-8"?>
<!-- a 2:1 red banner -->
<svg xmlns="http://www.w3.org/2000/svg" width="10" height="5" viewBox="0 0 100 50">
  <rect x="0" y="0" width="100" height="50" fill="#ff0000"/>
</svg>
`

func TestVectorImageGetsEveryIconSizeEvenWhenTiny(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "logo.svg", []byte(wideRedSVG))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 || outcome.Warned {
		t.Fatalf("outcome = %+v, want clean success; stderr:\n%s", outcome, stderr)
	}
	entries := readIco(t, filepath.Join(dir, "favicon.ico"))
	if got := entrySizes(entries); !reflect.DeepEqual(got, allIconSizes) {
		t.Fatalf("icon sizes = %v, want %v", got, allIconSizes)
	}
	largest := entries[len(entries)-1].img
	assertAlpha(t, largest, 128, 4, 0)   // padding above the 2:1 content
	assertAlpha(t, largest, 128, 251, 0) // padding below
	assertColour(t, largest, red)
	assertAlpha(t, largest, 2, 128, 255) // content spans the full width
}

func TestVectorImageIsRecognisedByContentNotExtension(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "logo.txt", []byte(wideRedSVG))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
	}
	readIco(t, filepath.Join(dir, "favicon.ico"))
}

// Minified SVGs (such as Simple Icons) pack arc flags without separators:
// "a12 12 0 1012 12" means rx=12 ry=12 rotation=0 large-arc=1 sweep=0 x=12 y=12.
func TestVectorImageWithPackedArcFlagsIsDrawn(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "disc.svg", []byte(`<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12 0a12 12 0 1012 12A12 12 0 0012 0z"/></svg>`))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
	}
	largest := readIco(t, filepath.Join(dir, "favicon.ico"))[5].img
	assertAlpha(t, largest, 128, 128, 255) // disc centre
	assertAlpha(t, largest, 2, 2, 0)       // outside the disc
}

// The hole is a circle drawn as two arcs, the second one an implicit repeat of
// the arc command, as in minified Simple Icons SVGs.
func TestVectorImageWithImplicitlyRepeatedArcsKeepsItsHole(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "ring.svg", []byte(`<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12 0a12 12 0 1012 12A12 12 0 0012 0zm0 20A8 8 0 1120 12 8 8 0 0112 20z"/></svg>`))

	run(t, src)

	largest := readIco(t, filepath.Join(dir, "favicon.ico"))[5].img
	assertAlpha(t, largest, 128, 128, 0)  // centre of the hole
	assertAlpha(t, largest, 181, 181, 0)  // hole, between the second arc and its chord
	assertAlpha(t, largest, 128, 10, 255) // ring
}

// Icon sets such as Lucide paint with currentColor, which a standalone SVG
// resolves to its initial value: black.
func TestVectorImagePaintedWithCurrentColorIsDrawnBlack(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "bar.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><path d="M2 12H22"/></svg>`))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 0 {
		t.Fatalf("exit code = %d, stderr:\n%s", outcome.ExitCode, stderr)
	}
	largest := readIco(t, filepath.Join(dir, "favicon.ico"))[5].img
	assertColour(t, largest, color.NRGBA{A: 255}) // the stroke crosses the centre
	assertAlpha(t, largest, 128, 20, 0)           // fill="none": nothing above the stroke
}

// A stroke-width is in the SVG's own units, so it grows with the Icon Size:
// 2 units of a 24-unit viewBox are about 21 px at 256 px.
func TestVectorImageStrokeWidthScalesWithIconSize(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "bar.svg", []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="#000" stroke-width="2"><path d="M2 12H22"/></svg>`))

	run(t, src)

	largest := readIco(t, filepath.Join(dir, "favicon.ico"))[5].img
	assertAlpha(t, largest, 128, 128+8, 255) // inside the 21 px stroke
	assertAlpha(t, largest, 128, 128+14, 0)  // beyond it
}

func TestVectorImageThatDrawsNothingIsRejected(t *testing.T) {
	dir := t.TempDir()
	src := writeFile(t, dir, "label.svg", []byte(`<svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><text x="2" y="12">hi</text></svg>`))

	outcome, _, stderr := run(t, src)

	if outcome.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", outcome.ExitCode)
	}
	if !strings.Contains(stderr, "label.svg") {
		t.Fatalf("error should name the Source Image, got:\n%s", stderr)
	}
	assertMissing(t, filepath.Join(dir, "favicon.ico"))
}
