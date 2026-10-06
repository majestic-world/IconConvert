BINARY    := bin/IconConvert.exe
CMD       := ./cmd/iconconvert
WINRES    := cmd/iconconvert/winres
RESOURCES := cmd/iconconvert/rsrc_windows_amd64.syso

# Release build: no cgo (ADR-0001), no local paths, no debug symbols or DWARF.
export CGO_ENABLED := 0
export GOOS        := windows
export GOARCH      := amd64
LDFLAGS := -s -w -buildid=

.PHONY: build test clean

build: $(RESOURCES)
	go build -trimpath -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD)

# Embedded icon and version information; regenerated when their sources change.
$(RESOURCES): $(WINRES)/winres.json $(WINRES)/icon.ico
	go generate $(CMD)

test:
	go test ./...

clean:
	go clean
	-rm -rf bin
