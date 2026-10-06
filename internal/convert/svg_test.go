package convert_test

import (
	"path/filepath"
	"reflect"
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
