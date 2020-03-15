package subvert

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

type SubvertTester struct {
	A int
	a int
	int
}

const constString = "testing"

func Demonstrate() {
	v := SubvertTester{1, 2, 3}

	rv := reflect.ValueOf(v)
	rv_A := rv.FieldByName("A")
	rv_a := rv.FieldByName("a")
	rv_int := rv.FieldByName("int")

	fmt.Printf("Interface of A: %v\n", rv_A.Interface())

	// MakeWritable

	// rv_a.Interface() // This would panic
	if err := MakeWritable(&rv_a); err != nil {
		// TODO: Handle this
	} else {
		fmt.Printf("Interface of a: %v\n", rv_a.Interface())
	}

	// rv_int.Interface() // This would panic
	if err := MakeWritable(&rv_int); err != nil {
		// TODO: Handle this
	} else {
		fmt.Printf("Interface of int: %v\n", rv_int.Interface())
	}

	// MakeAddressable

	// rv.Addr() // This would panic
	if err := MakeAddressable(&rv); err != nil {
		// TODO: Handle this
	} else {
		fmt.Printf("Pointer to v: %v\n", rv.Addr())
	}

	// ExposeFunction

	exposed, err := ExposeFunction("reflect.methodName", (func() string)(nil))
	if err != nil {
		// TODO: Handle this
	} else {
		f := exposed.(func() string)
		fmt.Printf("Result of reflect.methodName: %v\n", f())
	}

	// PatchMemory

	rv = reflect.ValueOf(constString)
	if err := MakeAddressable(&rv); err != nil {
		// TODO: Handle this
	} else {
		strAddr := rv.Addr().Pointer()
		strBytes := *((*unsafe.Pointer)(unsafe.Pointer(strAddr)))
		if oldMem, err := PatchMemory(uintptr(strBytes), []byte("XX")); err != nil {
			// TODO: Handle this
		} else {
			fmt.Printf("constString is now: %v, Oldmem = %v\n", constString, string(oldMem))
		}
	}
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

func assertPanics(t *testing.T, function func()) {
	if getPanic(function) == nil {
		t.Errorf("Expected function to panic")
	}
}

func assertDoesNotPanic(t *testing.T, function func()) {
	if err := getPanic(function); err != nil {
		t.Errorf("Unexpected panic: %v", err)
	}
}

func TestDemonstrate(t *testing.T) {
	Demonstrate()
}

func TestAddressable(t *testing.T) {
	rv := reflect.ValueOf(1)

	assertPanics(t, func() { rv.Addr() })
	if err := MakeAddressable(&rv); err != nil {
		t.Error(err)
	}
	rv.Addr()
}

func TestWritable(t *testing.T) {
	v := SubvertTester{}

	rv_A := reflect.ValueOf(v).FieldByName("A")
	rv_a := reflect.ValueOf(v).FieldByName("a")
	rv_int := reflect.ValueOf(v).FieldByName("int")

	rv_A.Interface()

	assertPanics(t, func() { rv_a.Interface() })
	if err := MakeWritable(&rv_a); err != nil {
		t.Error(err)
		return
	}
	rv_a.Interface()

	assertPanics(t, func() { rv_int.Interface() })
	if err := MakeWritable(&rv_int); err != nil {
		t.Error(err)
		return
	}
	rv_int.Interface()
}

func TestExposeFunction(t *testing.T) {
	if runtime.GOOS == "windows" {
		fmt.Printf("Skipping TestExposeFunction because it doesn't work in test binaries on this platform\n")
		return
	}

	assertDoesNotPanic(t, func() {
		exposed, err := ExposeFunction("reflect.methodName", (func() string)(nil))
		if err != nil {
			t.Error(err)
			return
		}
		if exposed == nil {
			t.Errorf("exposed should not be nil")
			return
		}
		f := exposed.(func() string)
		expected := "github.com/kstenerud/go-subvert.getPanic"
		actual := f()
		if actual != expected {
			t.Errorf("Expected [%v] but got [%v]", expected, actual)
		}
	})
}

func TestPatchMemory(t *testing.T) {
	// Note: If two const strings have the same value, they will occupy the same
	// location in memory! Thus myStr was made with a different value from
	// constString
	const myStr = "some test"
	rv := reflect.ValueOf(myStr)
	if err := MakeAddressable(&rv); err != nil {
		t.Error(err)
		return
	}
	strAddr := rv.Addr().Pointer()
	strBytes := *((*unsafe.Pointer)(unsafe.Pointer(strAddr)))
	oldMem, err := PatchMemory(uintptr(strBytes)+5, []byte("XXXX"))
	if err != nil {
		t.Error(err)
		return
	}

	expectedOldMem := "test"
	if string(oldMem) != expectedOldMem {
		t.Errorf("Expected oldMem to be %v but got %v", expectedOldMem, string(oldMem))
		return
	}

	expected := "some XXXX"
	// Note: Comparing myStr will fail due to cached static data. You must copy it first.
	actual := myStr
	if actual != expected {
		t.Errorf("Expected %v but got %v", expected, actual)
	}
}
