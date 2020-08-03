package util

import "unsafe"

type deepSli struct {
	ptr uintptr
	len uint64
	cap uint64
}

type deepString struct {
	ptr uintptr
	cap uint64
}

func Str2Bytes(src string) (dst []byte) {
	dst = *((*[]byte)((unsafe.Pointer)(&(deepSli{
		(*deepString)(unsafe.Pointer(&src)).ptr,
		(*deepString)(unsafe.Pointer(&src)).cap,
		(*deepString)(unsafe.Pointer(&src)).cap}))))
	return
}

func Bytes2Str(src []byte) (dst string) {
	dst = *(*string)((unsafe.Pointer)(&(deepString{
		(*deepSli)(unsafe.Pointer(&src)).ptr,
		(*deepSli)(unsafe.Pointer(&src)).len,
	})))
	return
}

func Sli2Array(src []interface{}) (dst interface{}) {
	dst = *(*string)((unsafe.Pointer)(&(deepString{
		(*deepSli)(unsafe.Pointer(&src)).ptr,
		(*deepSli)(unsafe.Pointer(&src)).len,
	})))
	return
}