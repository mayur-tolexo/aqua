package aqua

import (
	"testing"

	"github.com/rightjoin/aero/str"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUpFirstChar(t *testing.T) {
	Convey("Given the function: upFirstChar()", t, func() {
		Convey("Then it should convert first char to uppercase if it is lowercase", func() {
			So(str.SentenceCase("golang"), ShouldEqual, "Golang")
		})
		Convey("Then it should not make any change if the first char is already uppercase", func() {
			So(str.SentenceCase("Golang"), ShouldEqual, "Golang")
		})
		Convey("Then it should make no change to an empty string", func() {
			So(str.SentenceCase(""), ShouldEqual, "")
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

func TestToUrlCase(t *testing.T) {
	Convey("Given the function: toUrlCase()", t, func() {
		Convey("Then it should handle camelcase strings accurately", func() {
			So(str.UrlCase("AbraKaDabra"), ShouldEqual, "abra-ka-dabra")
			So(str.UrlCase("NCR"), ShouldEqual, "n-c-r")
		})
		Convey("Then it should leave numbers as such", func() {
			So(str.UrlCase("word1with2num"), ShouldEqual, "word1with2num")
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
