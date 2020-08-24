package sizeof

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unsafe"
)

var (
	_chan  chan bool
	Tab          = "\t"
	Bool   *Size = New(int64(unsafe.Sizeof(false)))
	Int64  *Size = New(int64(unsafe.Sizeof(int64(0))))
	Int32  *Size = New(int64(unsafe.Sizeof(int32(0))))
	Uint32 *Size = New(int64(unsafe.Sizeof(uint32(0))))
	Uint16 *Size = New(int64(unsafe.Sizeof(uint16(0))))
	Uint8  *Size = New(int64(unsafe.Sizeof(uint8(0))))
	Int    *Size = New(int64(unsafe.Sizeof(int(0))))
	Chan   *Size = New(int64(unsafe.Sizeof(_chan)))
	Func   *Size = New(int64(unsafe.Sizeof(func() {})))
	Char         = int64(unsafe.Sizeof('c'))
)

type Option func(*Size)

func WithVerbose() Option {
	return func(size *Size) {
		size.verbose = true
	}
}

func WithSliceLimit(limit int) Option {
	return func(size *Size) {
		size.sliceLimit = limit
	}
}

type Size struct {
	prefix     string
	result     int64
	stats      map[string]interface{}
	verbose    bool
	sliceLimit int
}

func New(size int64, opts ...Option) *Size {
	s := &Size{
		result:     size,
		stats:      map[string]interface{}{},
		sliceLimit: 10_000,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

func (s *Size) SizeOf(v interface{}) *Size {
	return s.sizeOf(reflect.ValueOf(v))
}

func (s *Size) String() string {
	data, err := json.MarshalIndent(s.stats, "", "    ")
	if err != nil {
		panic(err)
	}
	return string(data)
}

func (s *Size) Result() int64 {
	return s.result
}

func (s *Size) inner() *Size {
	return &Size{
		prefix:     s.prefix + Tab,
		stats:      map[string]interface{}{},
		verbose:    s.verbose,
		sliceLimit: s.sliceLimit,
	}
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
	case reflect.Uint16:
		return Uint16
	case reflect.Uint8:
		return Uint8
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
	case reflect.Invalid:
		return New(0)
	default:
		fmt.Print("Skipping:", val.Kind(), "\n")
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
	s.stats[val.Type().Name()] = map[string]interface{}{}
	s.result += int64(unsafe.Sizeof(val.Interface()))
	for i := 0; i < val.NumField(); i++ {
		s.writeResult(val, i)
	}
	return s
}

func (s *Size) writeResult(val reflect.Value, i int) {
	inner := s.inner().sizeOf(val.Field(i))
	s.result += inner.result
	if s.verbose {
		m := s.stats[val.Type().Name()].(map[string]interface{})
		if len(inner.stats) == 0 {
			m[val.Type().Field(i).Name] = inner.result
		} else {
			m[val.Type().Field(i).Name] = inner.stats
		}
	}
}

func (s *Size) sizeOfSlice(val reflect.Value) *Size {
	s.result += int64(unsafe.Sizeof([]int{}))
	for i := 0; i < s.getLenWithLimit(val); i++ {
		s.result += s.inner().sizeOf(val.Index(i)).result
	}
	return s
}

func (s *Size) getLenWithLimit(val reflect.Value) int {
	if val.Len() > s.sliceLimit && s.sliceLimit != 0 {
		return s.sliceLimit
	}
	return val.Len()
}
