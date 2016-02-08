package aqua

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

type AStruct struct {
	I         string
	Therefore string
}

func TestEncodingDecoding(t *testing.T) {

	Convey("When you encode a reflect array of supported items", t, func() {

		r := make([]reflect.Value, 0)
		tp := make([]string, 0)

		// one: int
		r = append(r, reflect.ValueOf(12345))
		tp = append(tp, "int")

		// two: string
		r = append(r, reflect.ValueOf("54321"))
		tp = append(tp, "string")

		// three: map
		m := make(map[string]interface{})
		m["key1"] = "value1"
		m["key2"] = "value2"
		r = append(r, reflect.ValueOf(m))
		tp = append(tp, "map")

		// four: struct as interface
		var im interface{}
		im = AStruct{I: "i think", Therefore: "i am"}
		r = append(r, reflect.ValueOf(im))
		tp = append(tp, "i:.")

		// five: slice as interface
		slc := make([]AStruct, 2)
		slc[0] = AStruct{I: "one", Therefore: "eleven"}
		slc[1] = AStruct{I: "two", Therefore: "twelve"}
		var is interface{}
		is = slc
		r = append(r, reflect.ValueOf(is))
		tp = append(tp, "i:.")

		d := encode(r, tp)

		Convey("Then decoding it should return the same values", func() {

			r := decode(d, tp)

			// one: int
			So(r[0].Int(), ShouldEqual, 12345)

			// two: string
			So(r[1].String(), ShouldEqual, "54321")

			// three: map
			So(r[2].Interface().(map[string]interface{})["key1"], ShouldEqual, "value1")
			So(r[2].Interface().(map[string]interface{})["key2"], ShouldEqual, "value2")

			// four: struct (map)
			So(r[3].Interface().(map[string]interface{})["I"], ShouldEqual, "i think")

			// five: slice ([]interface{})
			So(r[4].Interface().([]interface{})[0].(map[string]interface{})["I"], ShouldEqual, "one")
		})
	})
}
