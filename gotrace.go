package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	programName := os.Args[1]
	cmd := exec.Command(programName)
	cmd.Start()
	pid := cmd.Process.Pid

	target, _ := CreateDebugTarget(pid, false, false)
	target.Wait(true)

	pc, _ := target.GetRegPC()
	fmt.Printf("PC: 0x%X\n", PC)

	bpc := (pc / 0x100) * 0x100 + 0x100
	fmt.Printf("Breakpoint Address: 0x%X\n", bpc)

	err := target.SetBreakpoint(uintptr(bpc))
	if err != nil {
		fmt.Println(err.Error())
	}

	target.Continue()
	target.Wait(true)

	pc, _ = target.GetRegPC()
	fmt.Printf("PC: 0x%X\n", PC)

	if pc == bpc + 1 {
		fmt.Println("Breakpoint hit successfully!")
	}

	err := target.DelBreakpoint(uintptr(bpc), true)
	if err != nil {
		fmt.Println(err.Error())
	}

	pc, _ = target.GetRegPC()
	if pc == bpc {
		fmt.Println("Breakpoint removed successfully!")
	}
	
	target.Continue()
	target.Wait(true)
}