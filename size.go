package sizeof

import (
	"fmt"
	"io"
	"reflect"
	"strings"
	"unsafe"
)

const (
	Bool  = int64(unsafe.Sizeof(false))
	Int64 = int64(unsafe.Sizeof(int64(0)))
	Int32 = int64(unsafe.Sizeof(int32(0)))
	Uint32 = int64(unsafe.Sizeof(uint32(0)))
	Int = int64(unsafe.Sizeof(int(0)))
)

func SizeOf(v interface{}, w io.Writer) int64 {
	val := reflect.ValueOf(v)
	return sizeOf(val, "", w)
}

func sizeOf(val reflect.Value, prefix string, w io.Writer) int64 {
	total := sizeOfObject(val, prefix, w)
	return total
}

func sizeOfObject(val reflect.Value, prefix string, w io.Writer) int64 {
	switch val.Kind() {
	case reflect.Ptr:
		return sizeOf(val.Elem(), prefix, w)
	case reflect.Int64:
		return Int64
	case reflect.Int32:
		return Int32
	case reflect.Uint32:
		return Uint32
	case reflect.Int:
		return Int
	case reflect.Bool:
		return Bool
	case reflect.String:
		return int64(unsafe.Sizeof('c')) * int64(val.Len())
	case reflect.Map:
		return sizeOfMap(val, prefix+"\t", w)
	case reflect.Slice:
		return sizeOfSlice(val, prefix+"\t", w)
	case reflect.Chan:
		var d chan bool
		return int64(unsafe.Sizeof(d))
	case reflect.Interface:
		fmt.Println("interface:", val.Elem())
		return sizeOf(val.Elem(), prefix, w)
	case reflect.Struct:
		return sizeOfStruct(val, prefix, w)
	case reflect.Func:
		return int64(unsafe.Sizeof(func(){}))
	default:
		w.Write([]byte(fmt.Sprint("Skipping:", val.Kind(), "\n")))
		return 0
	}
}

func sizeOfMap(val reflect.Value, prefix string, w io.Writer) int64 {
	total := int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		total += sizeOf(key, prefix+"\t", w) + sizeOf(val.MapIndex(key), "", w)
	}
	return total
}

func sizeOfStruct(val reflect.Value, prefix string, w io.Writer) int64 {
	w.Write([]byte(fmt.Sprintf("%s(%s::%s):\n", prefix, pkgName(val), val.Type().Name())))
	total := int64(unsafe.Sizeof(val.Interface()))
	npref := prefix + "\t"
	for i := 0; i < val.NumField(); i++ {
		result := sizeOf(val.Field(i), npref, w)
		w.Write([]byte(fmt.Sprintf("%s%s: %s [%d]\n", npref, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind(), result)))
		total += result
	}
	return total
}

func sizeOfSlice(val reflect.Value, prefix string, w io.Writer) int64 {
	total := int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		total += sizeOf(val.Index(i), prefix+"\t", w)
	}
	return total
}

func pkgName(a reflect.Value) string {
	paths := strings.Split(a.Type().PkgPath(), "/")
	if len(paths) == 1 {
		return paths[0]
	}
	return paths[len(paths)-1]
}