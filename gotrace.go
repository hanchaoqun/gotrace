package main

import (
	"os"
	"os/exec"
)

func main() {
	programName := os.Args[1]
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid
}