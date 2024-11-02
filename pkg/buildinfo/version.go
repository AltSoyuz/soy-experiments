package buildinfo

import "fmt"

// Version must be set via -ldflags '-X'
var Version string

func Init() {
	fmt.Printf("Version: %s\n", Version)
}
