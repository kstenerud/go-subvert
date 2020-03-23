// +build 396 amd64 amd64p32

package subvert

import (
	"encoding/binary"
	"fmt"

	"golang.org/x/arch/x86/x86asm"
)

const callOpFirstByte = byte(0xe8)
const callOpLength = 5
const callOpArgLength = 4

// Maps function location to a list of places where it's called from
var callLocations map[uintptr][]uintptr

func initCallCache() (err error) {
	if callLocations != nil {
		return
	}

	table, err := GetSymbolTable()
	if err != nil {
		return
	}

	addCallLocation := func(callLoc, callDst uintptr) {
		if callLoc == 0 || callDst == 0 {
			panic(fmt.Errorf("callLoc %x, callDst %x", callLoc, callDst))
		}
		locations, ok := callLocations[callDst]
		if !ok {
			locations = make([]uintptr, 0, 4)
		}
		locations = append(locations, callLoc)
		callLocations[callDst] = locations
	}

	callLocations = make(map[uintptr][]uintptr)

	registerSize := 32
	if is64BitUintptr {
		registerSize = 64
	}

	for _, f := range table.Funcs {
		bytes := SliceAtAddress(uintptr(f.Entry), int(f.End-f.Entry))
		pc := uintptr(f.Entry)
		for len(bytes) >= callOpLength {
			inst, _ := x86asm.Decode(bytes, registerSize)
			if bytes[0] == callOpFirstByte {
				argDst := bytes[1:callOpLength]
				callArg := uintptr(int32(binary.LittleEndian.Uint32(argDst)))
				callDst := uintptr(pc + callArg + callOpLength)
				addCallLocation(pc+1, callDst)
			}
			pc += uintptr(int64(inst.Len))
			bytes = bytes[inst.Len:]
		}
	}

	return
}

func osRedirectCalls(src, dst uintptr) (err error) {
	if err = initCallCache(); err != nil {
		return
	}

	callLocations, ok := callLocations[src]
	if !ok {
		err = fmt.Errorf("Function is not referenced in this program")
		return
	}

	for _, loc := range callLocations {
		newCallArg := uint32(dst - loc - callOpArgLength)
		bytes := SliceAtAddress(loc, callOpArgLength)

		err = applyToProtectedMemory(GetSliceAddr(bytes), callOpArgLength, func() {
			binary.LittleEndian.PutUint32(bytes, newCallArg)
		})
		if err != nil {
			return
		}
	}

	return
}
