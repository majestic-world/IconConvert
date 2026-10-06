// Command IconConvert turns Source Images into multi-size Favicons.
//
// The executable's icon and version information come from winres/winres.json;
// regenerate the embedded resources with `go generate` after changing it.
//
//go:generate go run github.com/tc-hib/go-winres@v0.3.3 make --arch amd64
package main

import (
	"bufio"
	"fmt"
	"os"

	"IconConvert/internal/convert"
)

func main() {
	outcome := convert.Run(os.Args[1:], os.Stdout, os.Stderr)
	if isDropInvocation() && (outcome.ExitCode != 0 || outcome.Warned) {
		fmt.Print("\nPress Enter to close...")
		bufio.NewReader(os.Stdin).ReadString('\n')
	}
	os.Exit(outcome.ExitCode)
}
