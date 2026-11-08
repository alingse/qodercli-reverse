// Package main qodercli-reverse decompiled entry point
package main

import (
	"github.com/alingse/qodercli-reverse/decompiled/cmd"
)

// Build-time variables injected via ldflags.
var (
	commit    = "unknown"
	buildTime = "unknown"
)

func main() {
	cmd.Execute()
}
