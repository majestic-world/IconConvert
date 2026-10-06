// Command IconConvert turns Source Images into multi-size Favicons.
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
