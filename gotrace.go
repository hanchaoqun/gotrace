package main

import (
	"fmt"
	"time"
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

	pc, err := target.GetRegPC()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("PC: 0x%X\n", pc)

	bpc := (pc / 0x100) * 0x100 + 0x100
	fmt.Printf("Breakpoint Address: 0x%X\n", bpc)

	err = target.SetBreakpoint(uintptr(bpc))
	if err != nil {
		fmt.Println(err.Error())
	}

	target.Continue()
	target.Wait(true)

	pc, err = target.GetRegPC()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Printf("PC: 0x%X\n", pc)

	if pc == bpc + 1 {
		fmt.Println("Breakpoint hit successfully!")
	}

	err = target.DelBreakpoint(uintptr(bpc), true)
	if err != nil {
		fmt.Println(err.Error())
	}

	pc, _ = target.GetRegPC()
	if pc == bpc {
		fmt.Println("Breakpoint removed successfully!")
	}

	target.Continue()
	target.Wait(true)

	n := 5
	for {
		time.Sleep(time.Second * 1)
		n = n - 1
		if n <= 0 {
			break
		}
	}
}