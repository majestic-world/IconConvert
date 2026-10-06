package convert_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestUnreadableSourceImagesAreReportedAndProduceNothing(t *testing.T) {
	cases := map[string]func(t *testing.T, dir string) string{
		"unsupported file": func(t *testing.T, dir string) string {
			return writeFile(t, dir, "notes.txt", []byte("just some text"))
		},
		"corrupt image": func(t *testing.T, dir string) string {
			return writeFile(t, dir, "broken.png", []byte("\x89PNG\r\n\x1a\ngarbage"))
		},
		"corrupt ico": func(t *testing.T, dir string) string {
			return writeFile(t, dir, "broken.ico", []byte("\x00\x00\x01\x00\x01\x00garbage"))
		},
		"missing path": func(t *testing.T, dir string) string {
			return filepath.Join(dir, "nowhere.png")
		},
		"folder": func(t *testing.T, dir string) string {
			sub := filepath.Join(dir, "images")
			must(t, os.Mkdir(sub, 0o755))
			return sub
		},
	}
	for name, setup := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			src := setup(t, dir)

			outcome, _, stderr := run(t, src)

			if outcome.ExitCode != 1 {
				t.Fatalf("exit code = %d, want 1", outcome.ExitCode)
			}
			if !strings.Contains(stderr, filepath.Base(src)) {
				t.Fatalf("error should name %s, got:\n%s", filepath.Base(src), stderr)
			}
			assertNoFavicons(t, dir)
		})
	}
}

func TestBatchKeepsConvertingPastAFailure(t *testing.T) {
	dir := t.TempDir()
	bad := writeFile(t, dir, "notes.txt", []byte("just some text"))
	good := writePNG(t, dir, "logo.png", 256, 256)

	outcome, stdout, stderr := run(t, bad, good)

	if outcome.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", outcome.ExitCode)
	}
	readIco(t, filepath.Join(dir, "logo.ico"))
	if !strings.Contains(stdout, "logo.ico") {
		t.Fatalf("output should report the created Favicon, got:\n%s", stdout)
	}
	if !strings.Contains(stderr, "notes.txt") {
		t.Fatalf("errors should name the failed Source Image, got:\n%s", stderr)
	}
}

func TestSummaryListsEachSourceImageOnceWithItsResult(t *testing.T) {
	dir := t.TempDir()
	good := writePNG(t, dir, "logo.png", 256, 256)
	small := writePNG(t, dir, "small.png", 64, 64)
	bad := writeFile(t, dir, "notes.txt", []byte("just some text"))

	_, stdout, stderr := run(t, good, small, bad)

	lines := strings.Split(stdout+stderr, "\n")
	assertSingleLine(t, lines, "logo.png", "logo.ico")
	assertSingleLine(t, lines, "small.png", "small.ico", "128", "256")
	assertSingleLine(t, lines, "notes.txt", "error")
}

// assertSingleLine checks that exactly one line mentions source and that this
// line also carries every detail.
func assertSingleLine(t *testing.T, lines []string, source string, details ...string) {
	t.Helper()
	var matches []string
	for _, l := range lines {
		if strings.Contains(l, source) {
			matches = append(matches, l)
		}
	}
	if len(matches) != 1 {
		t.Fatalf("want exactly one line for %s, got %q", source, matches)
	}
	for _, d := range details {
		if !strings.Contains(matches[0], d) {
			t.Fatalf("line for %s should mention %q, got %q", source, d, matches[0])
		}
	}
}

func TestNoSourceImagesPrintsUsageAndFails(t *testing.T) {
	outcome, _, stderr := run(t)

	if outcome.ExitCode != 1 {
		t.Fatalf("exit code = %d, want 1", outcome.ExitCode)
	}
	if !strings.Contains(stderr, "IconConvert") {
		t.Fatalf("usage should name the command, got:\n%s", stderr)
	}
}

func assertNoFavicons(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, "*.ico"))
	must(t, err)
	for _, m := range matches {
		if filepath.Base(m) != "broken.ico" {
			t.Fatalf("unexpected Favicon %s", filepath.Base(m))
		}
	}
}
