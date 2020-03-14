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

	f := ExposeFunction("reflect.methodName", (func() string)(nil)).(func() string)
	if f != nil {
		fmt.Printf("Result of reflect.methodName: %v\n", f())
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

func TestExposeFunction(t *testing.T) {
	assertDoesNotPanic(t, func() {
		f := ExposeFunction("reflect.methodName", (func() string)(nil)).(func() string)
		if f == nil {
			t.Errorf("Cannot find reflect.methodName. This test is no longer valid.")
			return
		}
		expected := "github.com/kstenerud/go-subvert.getPanic"
		actual := f()
		if actual != expected {
			t.Errorf("Expected [%v] but got [%v]", expected, actual)
		}
	})
}
