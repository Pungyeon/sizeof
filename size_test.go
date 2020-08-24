package sizeof

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

type Flatter interface {}

type SliceInterface struct {
	things []Abstract
}

type Abstract struct {
	flatter Flatter
}

type Inner struct {
	dinner string
}

type Flat struct {
	name string
	contacts map[string]Flat
	tags []string
	channel chan string
	inner Inner
}

func TestSizeOf(t *testing.T) {
	var char rune = 'd'
	var str string = "dddd"
	var i interface{}
	var c chan int
	interfaceSize := int64(unsafe.Sizeof(i))
	mapSize := int64(unsafe.Sizeof(map[string]Flat{}))
	sliceSize := int64(unsafe.Sizeof([]string{}))
	chanSize := int64(unsafe.Sizeof(c))
	fullFlatSize := (interfaceSize*2) + mapSize + sliceSize + chanSize

	t.Run("check primitives", func(t *testing.T) {
		check(t, char, int64(unsafe.Sizeof(char)))
		check(t, str, int64(unsafe.Sizeof(char))*int64(len(str)))
		check(t, Flat{}, fullFlatSize)
	})

	t.Run("struct size", func(t *testing.T) {
		fmt.Println("--------------------")
		check(t,
			Flat{
				name: "dingeling",
				inner: Inner{
					dinner: "chicken winner",
				},
			},
			fullFlatSize+int64(unsafe.Sizeof('c')*23))
	})

	var flatter Flatter
	t.Run("interface size", func(t *testing.T) {
		flatter = &Flat{
			name: "dingeling",
		}
		check(t,
			flatter,
			fullFlatSize+int64(unsafe.Sizeof('c')*9))
	})
}

func BenchmarkLargeSlice(t *testing.B) {
	slice := make([]Abstract, 1_000_000)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		size := SizeOf(&SliceInterface{ things: slice})
		if size.Result() > 16000000 {
			t.Fatal(size.Result())
		}
	}
}

var nestedObjectString = "{\n    \"Abstract\": {\n        \"flatter\": {\n            \"Flat\": {\n                \"channel\": 8,\n                \"contacts\": 8,\n                \"inner\": {\n                    \"Inner\": {\n                        \"dinner\": 0\n                    }\n                },\n                \"name\": 40,\n                \"tags\": 24\n            }\n        }\n    }\n}"

var simpleString = `{
    "Flat": {
        "channel": 8,
        "contacts": 8,
        "inner": {
            "Inner": {
                "dinner": 0
            }
        },
        "name": 0,
        "tags": 72
    }
}`

func TestInterfaceString(t *testing.T) {
	t.Run("simple string", func(t *testing.T) {
		s := SizeOf(
			&Flat{tags: []string{
			"ding", "dong", "dyno",
		}},
		WithVerbose())

		if s.String() != simpleString {
			t.Fatal(s.String())
		}
	})

	t.Run("slice interface string", func(t *testing.T) {
		size := SizeOf(
			&SliceInterface{things: make([]Abstract, 100)},
			WithVerbose())

		if size.String() != "{\n    \"SliceInterface\": {\n        \"things\": 1624\n    }\n}" {
			t.Fatalf("wrong string result returned:\n%#v\n", size.String())

		}
	})

	t.Run("nested object string", func(t *testing.T) {
		s := SizeOf(
			&Abstract{flatter: &Flat{ name: "dingeling "}},
			WithVerbose())
		if s.String() != nestedObjectString {
			t.Fatal(s.String())
		}
	})
}

func check(t *testing.T, a interface{}, b int64) {
	size := SizeOf(a)
	if size.Result() != b {
		t.Fatalf("Not equal size (%s): %d != %d",
			reflect.ValueOf(a).Kind(), size.Result(), b)
	}
}