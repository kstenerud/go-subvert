package subvert

import (
	"fmt"
	"reflect"
	"testing"
)

type SubvertTester struct {
	A int
	a int
	int
}

func Demonstrate() {
	v := SubvertTester{1, 2, 3}

	rv := reflect.ValueOf(v)
	rv_A := rv.FieldByName("A")
	rv_a := rv.FieldByName("a")
	rv_int := rv.FieldByName("int")

	fmt.Printf("Interface of A: %v\n", rv_A.Interface())

	// rv_a.Interface() // This would panic
	MakeWritable(&rv_a)
	fmt.Printf("Interface of a: %v\n", rv_a.Interface())

	// rv_int.Interface() // This would panic
	MakeWritable(&rv_int)
	fmt.Printf("Interface of int: %v\n", rv_int.Interface())

	// rv.Addr() // This would panic
	MakeAddressable(&rv)
	fmt.Printf("Pointer to v: %v\n", rv.Addr())
}

func doesFunctionPanic(function func()) (didPanic bool) {
	defer func() {
		if e := recover(); e != nil {
			didPanic = true
		}
	}()

	function()
	return
}

func assertPanics(t *testing.T, function func()) {
	if !doesFunctionPanic(function) {
		t.Errorf("Expected function to panic")
	}
}

func TestEnabled(t *testing.T) {
	if !IsEnabled() {
		t.Error("IsEnabled() returned false. Check the logs and send an " +
			"error report to https://github.com/kstenerud/go-subvert/issues")
	}
}

func TestDemonstrate(t *testing.T) {
	Demonstrate()
}

func TestAddressable(t *testing.T) {
	rv := reflect.ValueOf(1)

	assertPanics(t, func() { rv.Addr() })
	MakeAddressable(&rv)
	rv.Addr()
}

func TestWritable(t *testing.T) {
	v := SubvertTester{}

	rv_A := reflect.ValueOf(v).FieldByName("A")
	rv_a := reflect.ValueOf(v).FieldByName("a")
	rv_int := reflect.ValueOf(v).FieldByName("int")

	rv_A.Interface()

	assertPanics(t, func() { rv_a.Interface() })
	MakeWritable(&rv_a)
	rv_a.Interface()

	assertPanics(t, func() { rv_int.Interface() })
	MakeWritable(&rv_int)
	rv_int.Interface()
}
