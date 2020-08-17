package sizeof

import (
	"bytes"
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
	Tab = "\t"
)
var (
	_chan chan bool
	Chan = int64(unsafe.Sizeof(_chan))
)

type Size struct {
	prefix string
	buffer bytes.Buffer
	result int64
}

func New() *Size {
	return &Size{}
}

func (s *Size) Inner() *Size {
	return &Size{
		prefix: s.prefix+Tab,
	}
}

func SizeOf(v interface{}) int64 {
	val := reflect.ValueOf(v)
	s := New()

	result := s.sizeOf(val)
	fmt.Println(s.buffer.String())
	return result
}

func (s *Size) sizeOf(val reflect.Value) int64 {
	total := s.sizeOfObject(val)
	return total
}

func (s *Size) sizeOfObject(val reflect.Value) int64 {
	switch val.Kind() {
	case reflect.Ptr:
		return s.sizeOf(val.Elem())
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
		return s.Inner().sizeOfMap(val)
	case reflect.Slice:
		return s.Inner().sizeOfSlice(val)
	case reflect.Chan:
		return Chan
	case reflect.Interface:
		return s.sizeOf(val.Elem())
	case reflect.Struct:
		return s.sizeOfStruct(val)
	case reflect.Func:
		return int64(unsafe.Sizeof(func(){}))
	default:
		s.buffer.WriteString(fmt.Sprint("Skipping:", val.Kind(), "\n"))
		return 0
	}
}

func (s *Size) sizeOfMap(val reflect.Value) int64 {
	total := int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		total += s.Inner().sizeOf(key) + s.sizeOf(val.MapIndex(key))
	}
	return total
}

func (s *Size) sizeOfStruct(val reflect.Value) int64 {
	s.buffer.Write([]byte(fmt.Sprintf("%s(%s::%s):\n", s.prefix, pkgName(val), val.Type().Name())))
	total := int64(unsafe.Sizeof(val.Interface()))
	npref := s.prefix + Tab
	for i := 0; i < val.NumField(); i++ {
		s.buffer.WriteString(fmt.Sprintf("%s%s: %s ", npref, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind()))
		inner := s.Inner()
		result := inner.sizeOf(val.Field(i))
		s.buffer.WriteString(fmt.Sprintf("[%d]\n", result))
		total += result
		s.buffer.Write(inner.buffer.Bytes())
	}
	return total
}

func (s *Size) sizeOfSlice(val reflect.Value) int64 {
	total := int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		total += s.Inner().sizeOf(val.Index(i))
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