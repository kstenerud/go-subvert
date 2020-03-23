package subvert

import (
	"reflect"
	"unsafe"
)

func getFunctionAddress(function interface{}) (address uintptr, err error) {
	rv := reflect.ValueOf(function)
	if err = MakeAddressable(&rv); err != nil {
		return
	}
	pFunc := (*unsafe.Pointer)(unsafe.Pointer(rv.UnsafeAddr()))
	address = uintptr(*pFunc)
	return
}

func newFunctionWithImplementation(template interface{}, implementationPtr uintptr) (function interface{}, err error) {
	rFunc := reflect.MakeFunc(reflect.TypeOf(template), nil)
	if err = MakeAddressable(&rFunc); err != nil {
		return
	}
	pFunc := (*unsafe.Pointer)(unsafe.Pointer(rFunc.UnsafeAddr()))
	*pFunc = unsafe.Pointer(uintptr(implementationPtr))
	function = rFunc.Interface()
	return
}
