# Ship as a single executable without cgo

IconConvert is used by dropping images onto the `.exe` or calling it from `PATH`, so it must run as one self-contained Windows binary with nothing else installed. Third-party Go libraries are welcome, but only those that build with `CGO_ENABLED=0` and need no companion DLLs or C toolchain; formats beyond the standard library (AVIF, HEIC via WASM-backed decoders, SVG via a pure-Go rasterizer) are covered within that limit.

## Considered Options

- **Native libraries via cgo (libvips, ImageMagick, libheif, librsvg)**: broader and more faithful format coverage, rejected because they require shipping DLLs or a C toolchain and break the single-file drop-onto-exe workflow.
- **Standard library and `golang.org/x/image` only**: rejected as too narrow; it leaves out phone photos (HEIC), AVIF and SVG logos.

## Consequences

- SVG support is partial: some SVG features render incorrectly or not at all.
- WASM-backed decoders make the executable noticeably larger.
