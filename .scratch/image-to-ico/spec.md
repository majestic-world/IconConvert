Status: ready-for-agent

# Spec: Convert images to multi-size `.ico`

## Problem Statement

I regularly need `.ico` files (favicons, app icons) made from images I already have: PNG logos, JPEG screenshots, phone photos in HEIC, SVG logos. Today that means uploading each image to a website such as icoconverter.com, picking sizes, downloading the result and moving it into the right folder. I want to do it locally, instantly, from wherever the image already sits, without a browser and without installing anything besides one executable.

## Solution

IconConvert is a single Windows executable that turns each Source Image into a Favicon carrying the standard Icon Sizes (16, 32, 48, 64, 128, 256). It is used two ways:

- **Drop Invocation**: drag one or more Source Images onto the `.exe` in Explorer.
- **Command Invocation**: with the executable on `PATH`, run `IconConvert <image> [<image>…]` from a terminal.

Each Favicon is written next to its Source Image and never over an existing file. A lone Source Image produces `favicon.ico`; in a Batch each Favicon is named after its Source Image (`logo.png` → `logo.ico`). When the name is taken, the lowest free number starting at 2 is appended (`favicon2.ico`, `favicon3.ico`, …). Raster Source Images smaller than an Icon Size skip that size with a warning; Vector Images (SVG) are drawn afresh at every Icon Size. Failures are reported clearly, the rest of a Batch still converts, and during a Drop Invocation the console window stays open whenever something went wrong.

## User Stories

1. As a user, I want to drag a PNG onto the executable, so that I get a `.ico` without opening a terminal or a browser.
2. As a user, I want to run `IconConvert logo.png` from the image's folder, so that I can convert from a terminal I already have open.
3. As a user, I want the executable to be reachable through `PATH`, so that I can call it from any folder.
4. As a user, I want a lone Source Image to produce a file named `favicon.ico`, so that the result is ready to drop into a website.
5. As a user, I want the Favicon written into the same folder as its Source Image, so that I find it where I expect regardless of how I launched the tool.
6. As a user, I want an existing `favicon.ico` to stay untouched and the new Favicon to be named `favicon2.ico`, so that I never lose a previous icon.
7. As a user, I want the lowest free number used when names are taken (`favicon.ico` and `favicon3.ico` exist → `favicon2.ico`), so that numbering fills gaps predictably.
8. As a user, I want to drag several images at once, so that I can convert a whole set in one go.
9. As a user, I want each Favicon in a Batch named after its Source Image (`logo.png` → `logo.ico`), so that I can tell which icon came from which image.
10. As a user, I want `logo.png` and `logo.jpg` in the same Batch to produce `logo.ico` and `logo2.ico`, so that neither result overwrites the other.
11. As a user, I want a Batch Source Image whose Icon Name already exists on disk to get the next free number, so that existing files are never replaced.
12. As a user, I want an `.ico` Source Image in a Batch to leave the original untouched (`app.ico` → `app2.ico`), so that re-converting an icon is safe.
13. As a user, I want every Favicon to contain the 16, 32, 48, 64, 128 and 256 pixel Icon Sizes, so that it looks sharp in tabs, taskbars, desktops and Explorer views.
14. As a user, I want a non-square Source Image centred on a transparent square, so that my logo is never stretched or cropped.
15. As a user, I want transparency in my Source Image preserved in the Favicon, so that icons blend with any background.
16. As a user, I want Icon Sizes larger than my raster Source Image's longer side left out of the Favicon, so that I never get blurry upscaled images.
17. As a user, I want a warning listing the Skipped Sizes, so that I know the Favicon is incomplete and can supply a larger image.
18. As a user, I want a 300×100 image to still receive the 256 Icon Size, so that wide logos are not penalised by their short side.
19. As a user, I want a raster Source Image smaller than 16 pixels on its longer side rejected with an error, so that I never get an empty Favicon.
20. As a user, I want an SVG drawn directly at every Icon Size, so that vector logos are crisp at all sizes and never produce Skipped Sizes.
21. As a user, I want PNG, JPEG, GIF, BMP, TIFF, WebP, ICO, AVIF, HEIC and SVG accepted, so that I can convert whatever image I happen to have.
22. As a user, I want an animated GIF converted from its first frame, so that I still get a usable icon.
23. As a user, I want an ICO Source Image converted from its largest embedded image, so that re-sizing an existing icon keeps the best quality available.
24. As a user, I want a file renamed with the wrong extension still recognised by its content, so that mislabelled images convert correctly.
25. As a user, I want an unsupported file (such as a `.txt`) reported as an error, so that I understand why nothing was produced.
26. As a user, I want a corrupt image reported as an error, so that I know which file is broken.
27. As a user, I want a path that does not exist or is a folder reported as an error, so that typos are obvious.
28. As a user, I want a failure in one Batch Source Image to leave the others converting, so that one bad file does not waste the whole run.
29. As a user, I want a summary at the end listing each Source Image with its Favicon path, warnings or error, so that I can review a Batch at a glance.
30. As a user, I want the console window to stay open waiting for a key after a Drop Invocation that had an error, so that I can read what went wrong.
31. As a user, I want the console window to close immediately after a fully successful Drop Invocation, so that a successful conversion needs no extra click.
32. As a user, I want Skipped Size warnings to keep the window open waiting for a key during a Drop Invocation, so that I notice an incomplete Favicon.
33. As a user running a Command Invocation, I want no key prompt, so that the tool never blocks my terminal or scripts.
34. As a user running a Command Invocation, I want a non-zero exit code when any Source Image failed, so that scripts can detect failure.
35. As a user running a Command Invocation, I want exit code 0 when every Source Image converted, even with warnings, so that scripts treat incomplete-but-valid Favicons as success.
36. As a user running IconConvert with no arguments, I want a short usage message, so that I learn how to use it.
37. As a user double-clicking the executable with no files, I want the usage message to stay visible until I press a key, so that I can read it.
38. As a user, I want to copy one `.exe` to any Windows machine and run it, so that there is nothing to install.

## Implementation Decisions

- **Single executable, no cgo** (ADR-0001): build with `CGO_ENABLED=0`; third-party Go libraries are allowed only if they need no DLLs or C toolchain.
- **Decoders by Supported Format**:
  - PNG, JPEG, GIF: Go standard library (GIF yields its first frame).
  - BMP, TIFF, WebP: `golang.org/x/image`.
  - AVIF: `github.com/gen2brain/avif` (WASM-backed, cgo-free).
  - HEIC: `github.com/gen2brain/heic` (WASM-backed, cgo-free).
  - SVG: `github.com/srwiley/oksvg` + `github.com/srwiley/rasterx` (partial SVG 2.0 support accepted).
  - ICO: a small in-project reader that selects the largest embedded entry (PNG- or BMP-encoded entries).
- **Format detection by content**: raster formats are recognised by their signature bytes, ignoring the extension. SVG, being text, is recognised by an XML/`<svg` root.
- **Raster pipeline**: decode → centre on a transparent square canvas whose side is the image's longer side → for each Icon Size not larger than that side, downscale with a high-quality filter (Catmull-Rom from `golang.org/x/image/draw`) → collect the rest as Skipped Sizes.
- **Vector pipeline**: for each Icon Size, rasterise the SVG directly into a transparent square of that size, preserving aspect ratio and centring. Vector Images never yield Skipped Sizes and are exempt from the 16-pixel minimum.
- **ICO encoding**: every Icon Size stored as a PNG-compressed entry (supported since Windows Vista), 32-bit RGBA, entries ordered smallest to largest.
- **Icon Name resolution**: base name is `favicon` when exactly one Source Image is given, otherwise the Source Image's file name without extension. Candidates are `<base>.ico`, then `<base>2.ico`, `<base>3.ico`, …; the first that does not exist in the Source Image's folder wins. The file is created with exclusive-create semantics so an existing file is never overwritten, even under a race; on collision the next candidate is tried.
- **Batch processing**: Source Images are converted sequentially in argument order, each independently; a failure records an error and moves on. Sequential order makes collisions within a Batch (`logo.png` + `logo.jpg`) resolve deterministically.
- **The seam: one runner function.** All behaviour sits behind a single entry point that takes the list of Source Image paths plus stdout/stderr writers and returns an outcome holding the exit code and whether any warning was printed. It writes Favicons to disk and prints per-image results and the summary. Exit code is `0` when every Source Image produced a Favicon (warnings allowed), `1` when any failed or when no paths were given (usage printed to stderr).
- **Thin `main`**: reads `os.Args`, calls the runner, then decides whether to pause. It detects a Drop Invocation by checking whether the process owns its console alone (Windows `GetConsoleProcessList` via `golang.org/x/sys/windows` returns 1). On a Drop Invocation it waits for Enter when the outcome has a non-zero exit code or a warning; a clean run exits immediately. A Command Invocation never pauses.
- **User-facing messages in English**, consistent with the project rule that code is written in English.
- **Module shape**: one deep conversion module (decode, square/resize, ICO encode, Icon Name resolution) behind the runner; `main` holds only process concerns. Internal pieces are not exported beyond what the runner needs.

## Testing Decisions

- **Good tests exercise external behaviour only**: given Source Image files on disk, assert the Favicon files that appear (names, contents), the warnings and errors printed, and the exit code. No assertions on internal functions, intermediate images or call order.
- **One seam: the runner function.** Every test creates Source Images in `t.TempDir()`, calls the runner with their paths and captured writers, then inspects the folder. A test helper parses the produced `.ico` (header + directory entries, decoding each PNG entry) to read back the actual Icon Sizes and verify pixel dimensions.
- **Fixtures**: PNG, JPEG, GIF, BMP and TIFF Source Images are generated in-test with encoders from the standard library and `golang.org/x/image`; SVG fixtures are written as inline text. WebP, AVIF and HEIC fixtures are small committed files under `testdata`, since no cgo-free encoder is available for all of them.
- **Behaviours to cover**: Icon Name for a lone image and for a Batch; first-free numbering with gaps; `.ico` Source Image not overwritten; Batch name collisions; each Supported Format converts; extension-mismatch detection; non-square centring with transparent padding; Skipped Sizes and their warning for a small raster image; 300×100 receiving 256; rejection below 16 px; SVG producing all six Icon Sizes; unsupported, corrupt, missing and directory paths; Batch continuing past a failure; exit codes; usage with no arguments.
- **`main` (Drop Invocation detection and pause) is verified manually**: drop a large valid image (window closes), drop a 64×64 image (Skipped Size warning, window waits for Enter), drop an invalid file (window waits for Enter), double-click with no files (usage stays visible). The runner's warning flag in the outcome is covered by the automated tests.
- **Prior art**: none; the repository has no code yet. These tests set the convention.

## Out of Scope

- Configurable Icon Sizes or custom output names/paths.
- Upscaling raster images to fill Skipped Sizes.
- Applying EXIF orientation to JPEG/HEIC photos; images convert as stored.
- Formats outside the Supported Format list (JPEG XL, PSD, RAW, PDF, …).
- Full SVG 2.0 fidelity; SVG features unsupported by the rasterizer may render incorrectly.
- Output formats other than `.ico` (`.icns`, PNG sets, web manifests).
- Installer, Explorer context-menu integration, automatic `PATH` setup.
- Non-Windows platforms.
- Recursing into folders passed as arguments.

## Further Notes

- WASM-backed decoders increase executable size noticeably (ADR-0001 consequence).
- Vocabulary follows `GLOSSARY.md`: Source Image, Supported Format, Vector Image, Favicon, Icon Size, Skipped Size, Icon Name, Drop Invocation, Command Invocation, Batch.
