// +build !windows,!darwin

package subvert

import (
	"debug/elf"
	"debug/gosym"
	"fmt"
	"io"
)

const canLoadSymbolsFromMemory = true
const processStartAddress = uintptr(0x400000)

func readSymbols(reader io.ReaderAt) (symTable *gosym.Table, err error) {
	exe, err := elf.NewFile(reader)
	if err != nil {
		return
	}
	defer exe.Close()

	sect := exe.Section(".text")
	if sect == nil {
		err = fmt.Errorf("Unable to find ELF .text section")
		return
	}
	textStart := sect.Addr

	sect = exe.Section(".gopclntab")
	if sect == nil {
		err = fmt.Errorf("Unable to find ELF .gopclntab section")
		return
	}
	lineTableData, err := sect.Data()
	if err != nil {
		return
	}

	lineTable := gosym.NewLineTable(lineTableData, textStart)
	return gosym.NewTable([]byte{}, lineTable)
}
