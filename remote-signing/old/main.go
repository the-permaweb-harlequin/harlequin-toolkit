package main

import (
	"os"
)

func main() {
	HandleRemoteSigningCommand(os.Args[1:])
}
