package main

import (
	"fmt"
	"os"

	"github.com/kstenerud/go-subvert"
)

func runTests(tests []func() error) bool {
	fmt.Printf("Running standalone tests...\n")
	success := true
	for _, test := range tests {
		symbol, err := subvert.GetFunctionSymbol(test)
		if err != nil {
			fmt.Println(err)
			return false
		}
		if err := test(); err != nil {
			fmt.Printf("Test %v failed: %v\n", symbol.Name, err)
			success = false
		}
	}
	return success
}

func main() {
	if !runTests([]func() error{
		TestAddressable,
		TestAliasFunction,
		TestExposeFunction,
		TestPatchMemory,
		TestSliceAddr,
		TestWritable,
	}) {
		fmt.Printf("Tests failed\n")
		os.Exit(1)
	}
}
