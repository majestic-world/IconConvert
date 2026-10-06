BINARY    := IconConvert.exe
RESOURCES := rsrc_windows_amd64.syso

# Release build: no cgo (ADR-0001), no local paths, no debug symbols or DWARF.
export CGO_ENABLED := 0
export GOOS        := windows
export GOARCH      := amd64
LDFLAGS := -s -w -buildid=

.PHONY: build test clean

build: $(RESOURCES)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) .

# Embedded icon and version information; regenerated when their sources change.
$(RESOURCES): winres/winres.json assets/icon.ico
	go generate

test:
	go test ./...

clean:
	go clean
	-rm -f $(BINARY)
