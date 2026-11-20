package main

import (
	"os"
	"os/signal"
	"syscall"
)

func main() {
	signal.Ignore(syscall.SIGPIPE)
	rc := run(os.Args[1:], osRuntime{})
	os.Exit(int(rc))
}
