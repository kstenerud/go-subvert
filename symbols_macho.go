// +build darwin

package subvert

import (
	"debug/gosym"
	"debug/macho"
	"fmt"
	"io"
)

const canLoadSymbolsFromMemory = false
const processStartAddress = uintptr(0x400000)

func readSymbols(reader io.ReaderAt) (symTable *gosym.Table, err error) {
	exe, err := macho.NewFile(reader)
	if err != nil {
		return
	}
	defer exe.Close()

	var sect *macho.Section
	if sect = exe.Section("__text"); sect == nil {
		err = fmt.Errorf("Unable to find Mach-O __text section")
		return
	}
	textStart := sect.Addr

	if sect = exe.Section("__gopclntab"); sect == nil {
		err = fmt.Errorf("Unable to find Mach-O __gopclntab section")
		return
	}
	lineTableData, err := sect.Data()
	if err != nil {
		return
	}

	lineTable := gosym.NewLineTable(lineTableData, textStart)
	return gosym.NewTable([]byte{}, lineTable)
}
