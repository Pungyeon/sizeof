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
	Char = int64(unsafe.Sizeof('c'))
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

func New(size int64) *Size {
	return &Size{
		result: size,
	}
}

func (s *Size) Inner() *Size {
	return &Size{
		prefix: s.prefix+Tab,
	}
}

func SizeOf(v interface{}) *Size {
	val := reflect.ValueOf(v)
	s := New(0)

	result := s.sizeOf(val)
	fmt.Println(s.buffer.String())
	return result
}

func (s *Size) sizeOf(val reflect.Value) *Size {
	total := s.sizeOfObject(val)
	return total
}

func (s *Size) sizeOfObject(val reflect.Value) *Size {
	switch val.Kind() {
	case reflect.Ptr:
		return s.sizeOf(val.Elem())
	case reflect.Int64:
		return New(Int64)
	case reflect.Int32:
		return New(Int32)
	case reflect.Uint32:
		return New(Uint32)
	case reflect.Int:
		return New(Int)
	case reflect.Bool:
		return New(Bool)
	case reflect.String:
		return New(Char * int64(val.Len()))
	case reflect.Map:
		return s.Inner().sizeOfMap(val)
	case reflect.Slice:
		return s.Inner().sizeOfSlice(val)
	case reflect.Chan:
		return New(Chan)
	case reflect.Interface:
		return s.sizeOf(val.Elem())
	case reflect.Struct:
		return s.sizeOfStruct(val)
	case reflect.Func:
		return New(int64(unsafe.Sizeof(func(){})))
	default:
		s.buffer.WriteString(fmt.Sprint("Skipping:", val.Kind(), "\n"))
		return New(0)
	}
}

func (s *Size) sizeOfMap(val reflect.Value) *Size {
	s.result += int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		s.result += s.Inner().sizeOf(key).result + s.sizeOf(val.MapIndex(key)).result
	}
	return s
}

func (s *Size) sizeOfStruct(val reflect.Value) *Size {
	s.buffer.Write([]byte(fmt.Sprintf("%s(%s::%s):\n", s.prefix, pkgName(val), val.Type().Name())))
	s.result += int64(unsafe.Sizeof(val.Interface()))
	npref := s.prefix + Tab
	for i := 0; i < val.NumField(); i++ {
		s.buffer.WriteString(fmt.Sprintf("%s%s: %s ", npref, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind()))
		inner := s.Inner()
		result := inner.sizeOf(val.Field(i))
		s.buffer.WriteString(fmt.Sprintf("[%d]\n", result))
		s.result += result.result
		s.buffer.Write(inner.buffer.Bytes())
	}
	return s
}

func (s *Size) sizeOfSlice(val reflect.Value) *Size {
	s.result += int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		s.result += s.Inner().sizeOf(val.Index(i)).result
	}
	return s
}

func pkgName(a reflect.Value) string {
	paths := strings.Split(a.Type().PkgPath(), "/")
	if len(paths) == 1 {
		return paths[0]
	}
	return paths[len(paths)-1]
}