package sizeof

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"testing"
	"unsafe"
)

type Flatter interface {}

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

	check(t, char, int64(unsafe.Sizeof(char)))
	check(t, str, int64(unsafe.Sizeof(char))*int64(len(str)))


	b := &bytes.Buffer{}
	func (w io.Writer) {
		SizeOf(b)
	}(b)

	check(t, Flat{}, fullFlatSize)

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

	t.Run("interface size", func(t *testing.T) {
		var flatter Flatter
		flatter = &Flat{
			name: "dingeling",
		}
		check(t,
			flatter,
			fullFlatSize+int64(unsafe.Sizeof('c')*9))
	})


}

func TestInterface(t *testing.T) {
	t.Error(SizeOf(&Abstract{flatter: &Flat{ name: "dingeling "}}).buffer.String())
	t.Error(SizeOf(&Flat{ contacts: map[string]Flat{
		"lasse": Flat{
			name: "jakobsen",
		},
	}, name: "dingeling "}).buffer.String())
}

func check(t *testing.T, a interface{}, b int64) {
	size := SizeOf(a)
	if size.result != b {
		t.Fatalf("Not equal size (%s): %d != %d",
			reflect.ValueOf(a).Kind(), size.result, b)
	}
}