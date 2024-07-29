package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"strings"
	"syscall"
	"golang.org/x/sys/unix"
)

const BreakPointInstr = 0xCCCCCCCCCCCCCCCC

type DebugError struct {
	Message string
}

func (d DebugError) Error() string {
	return d.Message
}

type Breakpoint struct {
	Enabled  bool
	Address  uintptr
	Backup   uint64
}

func (bp *Breakpoint) Enable(dt *DebugTarget) error {
	data, err := p.ReadUint64(b.Address)
	if err == nil {
		b.Backup = data
		err = p.WriteUint64(b.Address, BreakPointInstr)
		if err != nil {
			errMsg := fmt.Sprintf("Error enable breakpoint at address 0x%X: %s\n", b.Address, err.Error())
			return DebugError{Message: errMsg}
		}
		b.Enabled = true
		return nil
	} else {
		errMsg := fmt.Sprintf("Error reading bytes at address 0x%X: %s\n", b.Address, err.Error())
		return DebugError{Message: errMsg}
	}
}

func (bp *Breakpoint) Disable(dt *DebugTarget, resetpc bool) error {
	if b.Enabled {
		b.Enabled = false
		err := p.WriteUint64(b.Address, b.Backup)
		if err != nil {
			errMsg := fmt.Sprintf("Error restoring saved bytes to address 0x%X: %s", b.Address, err.Error())
			return DebugError{Message: errMsg}
		}
		if setIp {
			err = p.SetRegPC(uint64(b.Address))
			if err != nil {
				errMsg := fmt.Sprintf("Error set PC to 0x%X: %s", b.Address, err.Error())
				return DebugError{Message: errMsg}
			}
		}
		return nil
	}
	errMsg := "Breakpoint not set!"
	return DebugError{Message: errMsg}
}

func singalfilter(sig unix.Signal) unix.Signal {
	if sig == syscall.SIGTRAP {
		return 0
	} else {
		return sig
	}
}

type DebugTarget struct {
	Pid         int
	IsAttached  bool
	IsThread    bool
	IsLittle    bool
	Breakpoints map[uintptr]Breakpoint
}

func CreateDebugTarget(pid int, isAttached bool, isThread bool, isLittleEdian bool) (DebugTarget, error) {
	var err error = nil
	if !isAttached {
		attachErr := unix.PtraceAttach(pid)
		if attachErr != nil {
			errMsg := fmt.Sprintf("Could not attach to PID %d: %s", pid, attachErr.Error())
			err = DebugError{Message: errMsg}
		} else {
			err = nil
		}
	}
	process := DebugTarget{
		Pid:         pid,
		IsAttached:  true,
		IsThread:    isThread,
		IsLittle:    isLittleEdian,
		Breakpoints: make(map[uintptr]Breakpoint),
	}
	return process, err
}

func (dt *DebugTarget) Detach() error {
	if !dt.IsAttached {
		return nil
	}
	err := unix.PtraceDetach(dt.Pid)
	dt.IsAttached = false
	return err
}

func (dt *DebugTarget) SingleStep() error {
	return unix.PtraceSingleStep(dt.Pid);
}

func (dt *DebugTarget) ContToSignal(sig unix.Signal) error {
	return unix.PtraceCont(dt.Pid, int(singalfilter(sig)))
}

func (dt *DebugTarget) Cont() error {
	return unix.PtraceCont(dt.Pid, 0)
}


func (dt *DebugTarget) SyscallToSignal(sig unix.Signal) error {
	return unix.PtraceSyscall(dt.Pid, int(singalfilter(sig)))
}

func (dt *DebugTarget) Syscall() error {
	return unix.PtraceSyscall(dt.Pid, 0)
}

func (dt *DebugTarget) SendSig(sig unix.Signal) error {
	return unix.Kill(dt.Pid, sig)
}

func (dt *DebugTarget) SendSigKill() error {
	return unix.Kill(dt.Pid, syscall.SIGKILL)
}

func (dt *DebugTarget) GetRegs() (unix.PtraceRegs, error) {
	var regs unix.PtraceRegs
	err := unix.PtraceGetRegs(dt.Pid, &regs)
	return regs, err
}

func (dt *PtraceProcess) SetRegs(regs *unix.PtraceRegs) error {
	return unix.PtraceSetRegs(dt.Pid, regs)
}

func (dt *DebugTarget) GetReg(name string) (uint64, error) {
	name = strings.Title(name)
	regs, err := dt.GetRegs()
	if err != nil {
		return 0, err
	}
	v := reflect.ValueOf(&regs).Elem().FieldByName(name)
	if v.IsValid() {
		return v.Uint(), nil
	}
	return 0, nil
}

func (dt *DebugTarget) SetReg(name string, value uint64) error {
	name = strings.Title(name)
	regs, err := dt.GetRegs()
	if err != nil {
		return err
	}
	v := reflect.ValueOf(&regs).Elem().FieldByName(name)
	if v.IsValid() {
		v.SetUint(value)
		dt.SetRegs(&regs)
		return nil
	}
	return nil
}

func (dt *DebugTarget) Wait(block bool) (unix.WaitStatus, error) {
	options := 0
	if !block {
		options |= unix.WNOHANG
	}
	var wstatus unix.WaitStatus
	var rusage unix.Rusage
	_, err := unix.Wait4(p.Pid, &wstatus, options, &rusage)
	return wstatus, err
}

func (dt *DebugTarget) ReadUint64(address uintptr) (uint64, error) {
	bytes := make([]byte, 8)
	_, err := unix.PtracePeekData(dt.Pid, address, bytes)
	if dt.IsLittle {
		return binary.LittleEndian.Uint64(bytes), err
	}
	return binary.BigEndian.Uint64(bytes), err
}

func (dt *DebugTarget) WriteUint64(address uintptr, value uint64) error {
	bytes := make([]byte, 8)
	if dt.IsLittle {
		binary.LittleEndian.PutUint64(bytes, value)
	} else {
		binary.BigEndian.PutUint64(bytes, value)
	}
	_, err := unix.PtracePokeData(dt.Pid, address, bytes)
	return err
}

func (dt *DebugTarget) ReadBytes(address uintptr, size int) ([]byte, error) {
	bytes := make([]byte, size)
	_, err := unix.PtracePeekData(dt.Pid, address, bytes)
	return bytes, err
}

func (dt *DebugTarget) WriteBytes(address uintptr, bytes []byte) error {
	_, err := unix.PtracePokeData(dt.Pid, address, bytes)
	return err
}

func (dt *DebugTarget) ReadMemory(address uintptr, size int) ([]byte, error) {
	buf := make([]byte, size)
	localVec := []unix.Iovec{{
		Base: &buf[0],
		Len:  uint64(size),
	}}
	remoteVec := []unix.RemoteIovec{{
		Base: address,
		Len:  size,
	}}
	_, err := unix.ProcessVMReadv(dt.Pid, localVec, remoteVec, 0)
	return buf, err
}

func (dt *DebugTarget) WriteMemory(address uintptr, data []byte) error {
	size := len(data)
	localVec := []unix.Iovec{{
		Base: &data[0],
		Len:  uint64(size),
	}}
	remoteVec := []unix.RemoteIovec{{
		Base: address,
		Len:  size,
	}}
	_, err := unix.ProcessVMWritev(dt.Pid, localVec, remoteVec, 0)
	return err
}

func (dt *DebugTarget) GetBreakpoint(address uintptr) (Breakpoint, bool) {
	b, ok := dt.Breakpoints[address]
	return b, ok
}

func (dt *PtraceProcess) SetBreakpoint(address uintptr) error {
	_, ok := dt.GetBreakpoint(address)
	if !ok {
		b := Breakpoint{
			Enabled:  false,
			Address:  address,
			Backup:   0,
		}
		err := b.Enable(dt)
		if err != nil {
			return DebugError{Message: err.Error()}
		}
		dt.Breakpoints[address] = b
		return nil
	} else {
		return DebugError{Message: "Breakpoint already exists!"}
	}
}

func (dt *PtraceProcess) DelBreakpoint(address uintptr, resetpc bool) error {
	b, ok := dt.GetBreakpoint(address)
	if ok {
		err := b.Disable(dt, resetpc)
		if err != nil {
			return DebugError{Message: err.Error()}
		}
		delete(dt.Breakpoints, b.Address)
		return nil
	} else {
		return DebugError{Message: "Breakpoint does not exist!"}
	}
}

func (dt *PtraceProcess) SetOptions(options int) error {
	return unix.PtraceSetOptions(dt.Pid, options)
}


func (dt *PtraceProcess) SetRegPC(pc uint64) error {
	regs, err := dt.GetRegs()
	if err != nil {
		return err
	}
	regs.Rip = ip
	return dt.SetRegs(&regs)
}

func (dt *PtraceProcess) GetRegPC() (uint64, error) {
	regs, err := dt.GetRegs()
	return regs.Rip, err
}

func (dt *PtraceProcess) SetRegSP(sp uint64) error {
	regs, err := dt.GetRegs()
	if err != nil {
		return err
	}
	regs.Rsp = sp
	return dt.SetRegs(&regs)
}

func (dt *PtraceProcess) GetRegSP() (uint64, error) {
	regs, err := dt.GetRegs()
	return regs.Rsp, err
}

func (dt *PtraceProcess) SetRegFP(fp uint64) error {
	regs, err := dt.GetRegs()
	if err != nil {
		return err
	}
	regs.Rbp = fp
	return p.SetRegs(&regs)
}

func (dt *PtraceProcess) GetRegFP() (uint64, error) {
	regs, err := dt.GetRegs()
	return regs.Rbp, err
}