package sizeof

import (
	"fmt"
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

func SizeOf(v interface{}) int64 {
	val := reflect.ValueOf(v)
	return sizeOf(val, "")
}

func sizeOf(val reflect.Value, prefix string) int64 {
	total := sizeOfObject(val, prefix)
	return total
}

func sizeOfObject(val reflect.Value, prefix string) int64 {
	switch val.Kind() {
	case reflect.Ptr:
		return sizeOf(val.Elem(), prefix)
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
		return sizeOfMap(val, prefix+"\t")
	case reflect.Slice:
		return sizeOfSlice(val, prefix+"\t")
	case reflect.Chan:
		var d chan bool
		return int64(unsafe.Sizeof(d))
	case reflect.Interface:
		var i interface{}
		return int64(unsafe.Sizeof(i))
	case reflect.Struct:
		return sizeOfStruct(val, prefix)
	case reflect.Func:
		return int64(unsafe.Sizeof(func(){}))
	default:
		fmt.Println("Skipping:", val.Kind())
		return 0
	}
}

func sizeOfMap(val reflect.Value, prefix string) int64 {
	total := int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		total += sizeOf(key, prefix+"\t") + sizeOf(val.MapIndex(key), "")
	}
	return total
}

func sizeOfStruct(val reflect.Value, prefix string) int64 {
	fmt.Printf("%s(%s::%s):\n", prefix, pkgName(val), val.Type().Name())
	total := int64(unsafe.Sizeof(val.Interface()))
	npref := prefix + "\t"
	for i := 0; i < val.NumField(); i++ {
		result := sizeOf(val.Field(i), npref)
		fmt.Printf("%s%s: %s [%d]\n", npref, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind(), result)
		total += result
	}
	return total
}

func sizeOfSlice(val reflect.Value, prefix string) int64 {
	total := int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		total += sizeOf(val.Index(i), prefix+"\t")
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