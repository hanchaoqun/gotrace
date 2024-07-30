//go:build arm64

package main

import (
	"encoding/binary"
)

func debugerBytes2Uint64(bytes []byte) uint64 {
	return binary.LittleEndian.Uint64(bytes)
}

func debugerUint642Bytes(value uint64) []byte{
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, value)
	return bytes
}

func debugerGetRegPC(regs *unix.PtraceRegs) uint64 {
	return regs.Pc
}

func debugerSetRegPC(regs *unix.PtraceRegs, value uint64) {
	regs.Pc = value
}

func debugerGetRegSP(regs *unix.PtraceRegs) uint64 {
	return regs.Sp
}

func debugerSetRegSP(regs *unix.PtraceRegs, value uint64) {
	regs.Sp = value
}
