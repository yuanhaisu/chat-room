package main

import (
	"chat-room/client"
	"flag"
	"fmt"
	"os"
)

var (
	buildstamp = ""
	githash    = ""
)

func main() {
	args := os.Args
	if len(args) == 2 && (args[1] == "--version") {
		fmt.Printf("Git Commit Hash: %s\n", githash)
		fmt.Printf("UTC Build Time : %s\n", buildstamp)
		return
	}
	flag.Parse()
	client.Execute()
}
