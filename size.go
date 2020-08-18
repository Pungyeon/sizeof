package sizeof

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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

type Option func(*Size)



func WithWriter(w io.ReadWriter) Option {
	return func(s *Size) {
		s.buffer = w
	}
}

type Size struct {
	prefix string
	buffer io.ReadWriter
	result int64
	opts []Option
}

func New(size int64, opts ...Option) *Size {
	s := &Size{
		result: size,
		buffer: &bytes.Buffer{},
		opts: opts,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Size) String() string {
	data, err := ioutil.ReadAll(s.buffer)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

func (s *Size) Result() int64 {
	return s.result
}

func (s *Size) inner() *Size {
	inner := &Size{
		prefix: s.prefix+Tab,
		buffer: &bytes.Buffer{},
	}

	for _, opt := range s.opts {
		opt(inner)
	}

	return inner
}

func SizeOf(v interface{}, opts ...Option) *Size {
	return New(0, opts...).sizeOf(reflect.ValueOf(v))
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
		return s.inner().sizeOfMap(val)
	case reflect.Slice:
		return s.inner().sizeOfSlice(val)
	case reflect.Chan:
		return Chan
	case reflect.Interface:
		return s.sizeOf(val.Elem())
	case reflect.Struct:
		return s.sizeOfStruct(val)
	case reflect.Func:
		return Func
	default:
		s.buffer.Write([]byte(fmt.Sprint("Skipping:", val.Kind(), "\n")))
		return New(0)
	}
}

func (s *Size) sizeOfMap(val reflect.Value) *Size {
	s.result += int64(unsafe.Sizeof(map[int]int{}))
	for _, key := range val.MapKeys() {
		s.result += s.inner().sizeOf(key).result + s.sizeOf(val.MapIndex(key)).result
	}
	return s
}

func (s *Size) sizeOfStruct(val reflect.Value) *Size {
	s.writeStructHeader(val)
	s.result += int64(unsafe.Sizeof(val.Interface()))
	for i := 0; i < val.NumField(); i++ {
		s.writeProperty(val, i)
		s.writeResult(s.inner().sizeOf(val.Field(i)))
	}
	return s
}

func (s *Size) writeStructHeader(val reflect.Value) {
	fmt.Printf("%s(%s::%s):\n", s.prefix, pkgName(val), val.Type().Name())
	s.buffer.Write([]byte(fmt.Sprintf("%s(%s::%s):\n", s.prefix, pkgName(val), val.Type().Name())))
}

func (s *Size) writeProperty(val reflect.Value, i int) {
	s.buffer.Write([]byte(
		fmt.Sprintf("%s%s: %s ",
			s.prefix + Tab, val.Type().Field(i).Name, val.Type().Field(i).Type.Kind())))
}

func (s *Size) writeResult(inner *Size) {
	s.buffer.Write([]byte(fmt.Sprintf("[%d]\n", inner.result)))
	s.result += inner.result
	data, err := ioutil.ReadAll(inner.buffer)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	s.buffer.Write(data)
}

func (s *Size) sizeOfSlice(val reflect.Value) *Size {
	s.result += int64(unsafe.Sizeof([]int{}))
	for i := 0; i < val.Len(); i++ {
		s.result += s.inner().sizeOf(val.Index(i)).result
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