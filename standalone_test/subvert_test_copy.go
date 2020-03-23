package main

import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/kstenerud/go-subvert"
)

type SubvertTester struct {
	A int
	a int
	int
}

func getPanic(function func()) (result interface{}) {
	defer func() {
		if e := recover(); e != nil {
			result = e
		}
	}()

	function()
	return
}

func assertPanics(function func()) (err error) {
	if getPanic(function) == nil {
		return fmt.Errorf("Expected function to panic")
	}
	return
}

func assertDoesNotPanic(function func()) (err error) {
	if e := getPanic(function); e != nil {
		return fmt.Errorf("%v", e)
	}
	return
}

func TestAddressable() (err error) {
	rv := reflect.ValueOf(1)

	if err = assertPanics(func() { rv.Addr() }); err != nil {
		return
	}
	if err = subvert.MakeAddressable(&rv); err != nil {
		return
	}
	rv.Addr()
	return
}

func TestWritable() (err error) {
	v := SubvertTester{}

	rv_A := reflect.ValueOf(v).FieldByName("A")
	rv_a := reflect.ValueOf(v).FieldByName("a")
	rv_int := reflect.ValueOf(v).FieldByName("int")

	rv_A.Interface()

	if err = assertPanics(func() { rv_a.Interface() }); err != nil {
		return
	}
	if err = subvert.MakeWritable(&rv_a); err != nil {
		return
	}
	rv_a.Interface()

	if err = assertPanics(func() { rv_int.Interface() }); err != nil {
		return
	}
	if err = subvert.MakeWritable(&rv_int); err != nil {
		return
	}
	rv_int.Interface()
	return
}

func TestExposeFunction() (err error) {
	return assertDoesNotPanic(func() {
		var exposed interface{}
		exposed, err = subvert.ExposeFunction("reflect.methodName", (func() string)(nil))
		if err != nil {
			return
		}
		if exposed == nil {
			err = fmt.Errorf("exposed should not be nil")
			return
		}
		f := exposed.(func() string)
		expected := "github.com/kstenerud/go-subvert.getPanic"
		actual := f()
		if actual != expected {
			err = fmt.Errorf("Expected [%v] but got [%v]", expected, actual)
			return
		}
	})
}

func TestPatchMemory() (err error) {
	// Note: If two const strings have the same value, they will occupy the same
	// location in memory! Thus myStr was made with a different value from
	// constString
	const myStr = "some test"
	rv := reflect.ValueOf(myStr)
	if err = subvert.MakeAddressable(&rv); err != nil {
		return
	}
	strAddr := rv.Addr().Pointer()
	strBytes := *((*unsafe.Pointer)(unsafe.Pointer(strAddr)))
	oldMem, err := subvert.PatchMemory(uintptr(strBytes)+5, []byte("XXXX"))
	if err != nil {
		return
	}

	expectedOldMem := "test"
	if string(oldMem) != expectedOldMem {
		return fmt.Errorf("Expected oldMem to be %v but got %v", expectedOldMem, string(oldMem))
	}

	expected := "some XXXX"
	// Note: Comparing myStr will fail due to cached static data. You must copy it first.
	actual := myStr
	if actual != expected {
		return fmt.Errorf("Expected %v but got %v", expected, actual)
	}
	return
}

func TestSliceAddr() (err error) {
	expected := "abcd"
	addr := subvert.GetSliceAddr([]byte(expected))
	actual := string(subvert.SliceAtAddress(addr, 4))
	if actual != expected {
		return fmt.Errorf("Expected %v but got %v", expected, actual)
	}
	return
}

//go:noinline
func zFunc() string {
	return "z"
}

func TestAliasFunction() (err error) {
	fIntf, err := subvert.AliasFunction(zFunc)
	if err != nil {
		return
	}
	f := fIntf.(func() string)

	expected := "z"
	actual := f()
	if actual != expected {
		return fmt.Errorf("Expected %v, but got %v", expected, actual)
	}
	return
}
