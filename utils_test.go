package aqua

import (
	. "github.com/smartystreets/goconvey/convey"
	"reflect"
	"testing"
)

func TestUpFirstChar(t *testing.T) {
	Convey("Given the function: upFirstChar()", t, func() {
		Convey("Then it should convert first char to uppercase if it is lowercase", func() {
			So(upFirstChar("golang"), ShouldEqual, "Golang")
		})
		Convey("Then it should not make any change if the first char is already uppercase", func() {
			So(upFirstChar("Golang"), ShouldEqual, "Golang")
		})
		Convey("Then it should make no change to an empty string", func() {
			So(upFirstChar(""), ShouldEqual, "")
		})
	})
}

func TestRemoveMultSlashes(t *testing.T) {
	Convey("Given the function: removeMultSlashes()", t, func() {
		Convey("Then it should replace all multiple-slashes to a single slash", func() {
			So(removeMultSlashes("////"), ShouldEqual, "/")
			So(removeMultSlashes("////a/b//c///"), ShouldEqual, "/a/b/c/")
		})
	})
}

func TestCleanUrl(t *testing.T) {
	Convey("Given the function: cleanUrl()", t, func() {
		Convey("Then it should form the proper urls", func() {
			So(cleanUrl("a", "b", "c"), ShouldEqual, "/a/b/c")
			So(cleanUrl("/a/", "/b/", "/c"), ShouldEqual, "/a/b/c")
			So(cleanUrl("/a/", "/b/", "/c/"), ShouldEqual, "/a/b/c/")
		})
	})
}

func TestGetTypeInfo(t *testing.T) {
	Convey("Given the function: getSymbolFromType()", t, func() {
		Convey("Then it should revert st:<package-name>.<struct-name> for a struct type", func() {
			a := Sac{}
			So(getSignOfType(reflect.TypeOf(a)), ShouldEqual, "st:github.com/thejackrabbit/aqua.Sac")
			So(getSignOfType(reflect.TypeOf(&a)), ShouldEqual, "*st:github.com/thejackrabbit/aqua.Sac")
		})
		Convey("Then it should revert map for a map-type", func() {
			a := make(map[string]interface{})
			So(getSignOfType(reflect.TypeOf(a)), ShouldEqual, "map")
		})
	})
}

func TestToUrlCase(t *testing.T) {
	Convey("Given the function: toUrlCase()", t, func() {
		Convey("Then it should handle camelcase strings accurately", func() {
			So(toUrlCase("AbraKaDabra"), ShouldEqual, "abra-ka-dabra")
			So(toUrlCase("NCR"), ShouldEqual, "n-c-r")
		})
		Convey("Then it should leave numbers as such", func() {
			So(toUrlCase("word1with2num"), ShouldEqual, "word1with2num")
		})
	})
}

func TestExtractUrlPatterns(t *testing.T) {
	Convey("The function: extractRouteVars()", t, func() {
		Convey("Should be able to fetch Url patterns correctly", func() {
			url := "http://www.abc.com/{product}/{category}/{id:[0-9]+}"
			patt := extractRouteVars(url)
			So(patt[0], ShouldEqual, "product")
			So(patt[1], ShouldEqual, "category")
			So(patt[2], ShouldEqual, "id")
		})
	})

}

func TestConvertToType(t *testing.T) {
	Convey("The function: convertToType()", t, func() {
		Convey("Should work for string inputs", func() {
			vars := []string{"abc"}
			vals := convertToType(vars, []string{"string"})
			So(vals[0].Kind().String(), ShouldEqual, "string")
			So(vals[0].String(), ShouldEqual, "abc")
		})
		Convey("Should work for int inputs", func() {
			vars := []string{"abc", "12345"}
			vals := convertToType(vars, []string{"string", "int"})
			So(vals[1].Kind().String(), ShouldEqual, "int")
			So(vals[1].Int(), ShouldEqual, 12345)
		})
	})
}
