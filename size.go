package sizeof

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

const (
	Int32 int32 = 0
)

func SizeOf(v interface{}) int64 {
	val := reflect.ValueOf(v)
	return sizeOf(val, "")
}

func sizeOf(val reflect.Value, prefix string) int64 {
	total := sizeOfObject(val, prefix)
	fmt.Printf("(%s) [%d]\n", val.Kind(), total)
	return total
}

func sizeOfObject(val reflect.Value, prefix string) int64 {
	switch val.Kind() {
	case reflect.Ptr:
		return sizeOf(val.Elem(), prefix)
	case reflect.Int32:
		return int64(unsafe.Sizeof(Int32))
	case reflect.String:
		return int64(unsafe.Sizeof('c')) * int64(val.Len())
	case reflect.Map:
		return sizeOfMap(val, prefix+"\t")
	case reflect.Slice:
		return sizeOfSlice(val, prefix+"\t")
	case reflect.Chan:
		var d chan bool
		return int64(unsafe.Sizeof(d))
	case reflect.Struct:
		return sizeOfStruct(val, prefix+"\t")
	default:
		fmt.Println("Skipping:", val.Kind())
		return 0
	}
}

func sizeOfMap(val reflect.Value, prefix string) int64 {
	fmt.Println("(Map):")
	total := int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		total += sizeOf(key, prefix+"\t") + sizeOf(val.MapIndex(key), "")
	}
	fmt.Println(prefix, "Total:", total)
	return total
}

func sizeOfStruct(val reflect.Value, prefix string) int64 {
	fmt.Printf("(%s::%s):\n", pkgName(val), val.Type().Name())
	total := int64(unsafe.Sizeof(val.Interface()))
	for i := 0; i < val.NumField(); i++ {
		fmt.Printf("\t%s: ", val.Type().Field(i).Name)
		total += sizeOf(val.Field(i), prefix)
	}
	fmt.Println("(Struct):", total)
	return total
}

func sizeOfSlice(val reflect.Value, prefix string) int64 {
	fmt.Println("(Slice):")
	total := int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		total += sizeOf(val.Index(i), prefix+"\t")
	}
	fmt.Println(prefix, "Total:", total)
	return total
}

func pkgName(a reflect.Value) string {
	paths := strings.Split(a.Type().PkgPath(), "/")
	if len(paths) == 1 {
		return paths[0]
	}
	return paths[len(paths)-1]
}