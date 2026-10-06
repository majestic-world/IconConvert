# IconConvert

A Windows command-line tool that turns images into multi-size `.ico` files, invoked either by dropping images onto the executable or by calling it by name from the images' folder.

## Language

### Conversion

**Source Image**:
An image file in a Supported Format that the user hands to IconConvert to be turned into a Favicon; a raster one must be at least 16 pixels on its longer side.
_Avoid_: Input, original, file

**Supported Format**:
One of the image formats IconConvert reads: PNG, JPEG, GIF (first frame only), BMP, TIFF, WebP, ICO, AVIF, HEIC and SVG.
_Avoid_: Extension, file type

**Vector Image**:
An SVG Source Image, drawn afresh at every Icon Size, so it has no pixel dimensions and never yields a Skipped Size.
_Avoid_: Scalable image

**Favicon**:
The `.ico` file IconConvert produces from one Source Image, written to the Source Image's folder and never over an existing file.
_Avoid_: Output, icon file, ico

**Icon Size**:
One of the standard square pixel dimensions a Favicon may carry: 16, 32, 48, 64, 128 and 256.
_Avoid_: Resolution, layer, frame

**Skipped Size**:
An Icon Size larger than the longer side of its raster Source Image, left out of the Favicon and reported to the user as a warning.
_Avoid_: Missing size, dropped size

**Icon Name**:
The file name a Favicon receives: `favicon` for a lone Source Image, the Source Image's base name within a Batch, followed by the lowest free number starting at 2 when that name is taken.
_Avoid_: Output name, target path

### Invocation

**Drop Invocation**:
Running IconConvert by dragging one or more Source Images onto the executable in Explorer.
_Avoid_: Drag mode, double-click

**Command Invocation**:
Running IconConvert by name from a terminal, with the executable reachable through `PATH`.
_Avoid_: CLI mode, environment-variable mode

**Batch**:
A single invocation carrying more than one Source Image, each converted independently so one failure leaves the others intact.
_Avoid_: Multiple files, bulk
