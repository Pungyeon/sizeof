package sizeof

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"unsafe"
)

var (
	_chan chan bool
	Tab = "\t"
	Bool *Size  = New(int64(unsafe.Sizeof(false)))
	Int64 *Size = New(int64(unsafe.Sizeof(int64(0))))
	Int32 *Size = New(int64(unsafe.Sizeof(int32(0))))
	Uint32 *Size = New(int64(unsafe.Sizeof(uint32(0))))
	Int *Size = New(int64(unsafe.Sizeof(int(0))))
	Chan *Size = New(int64(unsafe.Sizeof(_chan)))
	Func *Size = New(int64(unsafe.Sizeof(func(){})))
	Char = int64(unsafe.Sizeof('c'))
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
	return New(0).sizeOf(reflect.ValueOf(v))
}

func (s *Size) add(n int64) *Size {
	s.result += n
	return s
}

func (s *Size) sizeOf(val reflect.Value) *Size {
	return s.sizeOfObject(val)
}

func (s *Size) sizeOfObject(val reflect.Value) *Size {
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
		return New(Char * int64(val.Len()))
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
		return Func
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
	s.writeStructHeader(val)
	s.result += int64(unsafe.Sizeof(val.Interface()))
	for i := 0; i < val.NumField(); i++ {
		s.writeProperty(val, i)
		s.writeResult(s.Inner().sizeOf(val.Field(i)))
	}
	return s
}

func (s *Size) writeStructHeader(val reflect.Value) {
	s.buffer.WriteString(fmt.Sprintf("%s(%s::%s):\n", s.prefix, pkgName(val), val.Type().Name()))
}

func (s *Size) writeProperty(val reflect.Value, i int) {
	s.buffer.WriteString(
		fmt.Sprintf("%s%s: %s ",
			s.prefix + Tab, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind()))
}

func (s *Size) writeResult(inner *Size) {
	s.buffer.WriteString(fmt.Sprintf("[%d]\n", inner.result))
	s.result += inner.result
	s.buffer.Write(inner.buffer.Bytes())
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