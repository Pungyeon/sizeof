package sizeof

import (
	"fmt"
	"reflect"
	"testing"
	"unsafe"
)

type Flatter interface {}

type Flat struct {
	name string
	contacts map[string]Flat
	tags []string
}

func TestSizeOf(t *testing.T) {
	var char rune = 'd'
	var str string = "dddd"
	var i interface{}
	flatSize := int64(unsafe.Sizeof(i))
	mapSize := int64(unsafe.Sizeof(map[string]Flat{}))
	sliceSize := int64(unsafe.Sizeof([]string{}))
	fullFlatSize := flatSize + mapSize + sliceSize

	check(t, char, int64(unsafe.Sizeof(char)))
	check(t, str, int64(unsafe.Sizeof(char))*int64(len(str)))

	check(t, Flat{}, fullFlatSize)

	t.Run("struct size", func(t *testing.T) {
		fmt.Println("--------------------")
		check(t,
			Flat{
				name: "dingeling",
			},
			fullFlatSize+int64(unsafe.Sizeof('c')*9))
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

func check(t *testing.T, a interface{}, b int64) {
	size := SizeOf(a)
	if size != b {
		t.Fatalf("Not equal size (%s): %d != %d",
			reflect.ValueOf(a).Kind(), size, b)
	}
}