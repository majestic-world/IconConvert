// Package convert turns Source Images into multi-size Favicons.
package convert

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	xdraw "golang.org/x/image/draw"
)

// iconSizes lists the standard Icon Sizes, smallest first.
var iconSizes = []int{16, 32, 48, 64, 128, 256}

// Outcome reports how a run ended.
type Outcome struct {
	// ExitCode is 0 when every Source Image produced a Favicon, 1 otherwise.
	ExitCode int
	// Warned is true when any warning was printed.
	Warned bool
}

const usage = `Usage: IconConvert <image> [<image>...]

Converts each image into a multi-size .ico next to it:
  one image   -> favicon.ico
  many images -> <image name>.ico
An existing file is never overwritten; a number is appended instead (favicon2.ico).
Supported formats: PNG, JPEG, GIF, BMP, TIFF, WebP, ICO, AVIF, HEIC, SVG.
`

// Run converts every Source Image in paths into a Favicon beside it. It
// reports one line per Source Image (warnings and errors on stderr) followed
// by the totals.
func Run(paths []string, stdout, stderr io.Writer) Outcome {
	if len(paths) == 0 {
		fmt.Fprint(stderr, usage)
		return Outcome{ExitCode: 1}
	}

	var outcome Outcome
	created, failed, warned := 0, 0, 0
	for _, path := range paths {
		base := "favicon"
		if len(paths) > 1 {
			base = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}
		favicon, skipped, err := convertOne(path, base)
		switch {
		case err != nil:
			fmt.Fprintf(stderr, "error: %s: %v\n", path, err)
			failed++
		case len(skipped) > 0:
			fmt.Fprintf(stderr, "warning: %s -> %s: image too small for Icon Sizes %s px; skipped\n", path, favicon, joinInts(skipped))
			created++
			warned++
		default:
			fmt.Fprintf(stdout, "%s -> %s\n", path, favicon)
			created++
		}
	}

	fmt.Fprintf(stdout, "\n%s created", count(created, "Favicon", "Favicons"))
	if warned > 0 {
		fmt.Fprintf(stdout, ", %d with warnings", warned)
	}
	if failed > 0 {
		fmt.Fprintf(stdout, ", %s failed", count(failed, "image", "images"))
		outcome.ExitCode = 1
	}
	fmt.Fprintln(stdout, ".")
	outcome.Warned = warned > 0
	return outcome
}

// convertOne writes the Favicon for one Source Image and returns its file
// name along with the Skipped Sizes.
func convertOne(path, base string) (string, []int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", nil, unwrapPath(err)
	}
	if info.IsDir() {
		return "", nil, errors.New("is a folder, not an image")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", nil, unwrapPath(err)
	}
	images, skipped, err := renderIconSizes(data)
	if err != nil {
		return "", nil, err
	}

	out, err := createFavicon(filepath.Dir(path), base)
	if err != nil {
		return "", nil, err
	}
	if err := encodeIco(out, images); err != nil {
		out.Close()
		os.Remove(out.Name())
		return "", nil, err
	}
	if err := out.Close(); err != nil {
		os.Remove(out.Name())
		return "", nil, err
	}
	return filepath.Base(out.Name()), skipped, nil
}

// renderIconSizes renders a Source Image at every Icon Size it supports and
// returns the Skipped Sizes.
func renderIconSizes(data []byte) ([]image.Image, []int, error) {
	if isSVG(data) {
		images, err := vectorIcons(data)
		return images, nil, err
	}
	src, err := decodeRaster(data)
	if err != nil {
		return nil, nil, err
	}
	return rasterIcons(src)
}

// rasterIcons centres src on a transparent square and downscales it to every
// Icon Size that fits; larger Icon Sizes are returned as skipped.
func rasterIcons(src image.Image) ([]image.Image, []int, error) {
	b := src.Bounds()
	side := max(b.Dx(), b.Dy())
	if side < iconSizes[0] {
		return nil, nil, fmt.Errorf("image is %dx%d px; at least %d px on the longer side is required", b.Dx(), b.Dy(), iconSizes[0])
	}

	canvas := image.NewNRGBA(image.Rect(0, 0, side, side))
	offset := image.Pt((side-b.Dx())/2, (side-b.Dy())/2)
	draw.Draw(canvas, b.Sub(b.Min).Add(offset), src, b.Min, draw.Src)

	var images []image.Image
	var skipped []int
	for _, size := range iconSizes {
		if size > side {
			skipped = append(skipped, size)
			continue
		}
		dst := image.NewNRGBA(image.Rect(0, 0, size, size))
		xdraw.CatmullRom.Scale(dst, dst.Bounds(), canvas, canvas.Bounds(), draw.Src, nil)
		images = append(images, dst)
	}
	return images, skipped, nil
}

// createFavicon exclusively creates the first free Icon Name in dir:
// <base>.ico, then <base>2.ico, <base>3.ico, ….
func createFavicon(dir, base string) (*os.File, error) {
	for n := 1; ; n++ {
		name := base + ".ico"
		if n > 1 {
			name = base + strconv.Itoa(n) + ".ico"
		}
		f, err := os.OpenFile(filepath.Join(dir, name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if errors.Is(err, fs.ErrExist) {
			continue
		}
		return f, err
	}
}

func joinInts(values []int) string {
	parts := make([]string, len(values))
	for i, v := range values {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ", ")
}

func count(n int, singular, plural string) string {
	if n == 1 {
		return "1 " + singular
	}
	return strconv.Itoa(n) + " " + plural
}

// unwrapPath drops the path from a *fs.PathError, since messages already
// name the Source Image.
func unwrapPath(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}
