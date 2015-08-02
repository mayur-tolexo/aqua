package aqua

import (
	"encoding/json"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"net/http"
	"testing"
)

type epMock struct{}

func (me *epMock) Handler1(w http.ResponseWriter, r *http.Request) {}

func TestStandardHttpHandlerIsIdentifiedCorrectly(t *testing.T) {
	Convey("Given an endpoint and a Service Controller", t, func() {
		Convey("The standard http handler should be identified correctly", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Handler1"), Fixture{}, "GET", nil, nil)
			So(ep.stdHandler, ShouldBeTrue)
		})
	})
}

func (me *epMock) Jar1(w http.ResponseWriter, r *http.Request) {}
func (me *epMock) Jar2(j Jar) string                           { return "" }
func (me *epMock) Jar3(j Jar, s string) string                 { return "" }
func (me *epMock) Jar4(s string, j Jar) string                 { return "" }
func (me *epMock) Jar5(j Jar, k Jar) string                    { return "" }

func TestJarInputIsIdentifiedCorrectly(t *testing.T) {
	Convey("Given an endpoint and a Service Controller", t, func() {
		Convey("A standard http function should NOT be marked for Jar", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Jar1"), Fixture{Url: "/abc/{d}/{e}"}, "GET", nil, nil) //"/abc/{d}/{e}",
			So(ep.jarInput, ShouldBeFalse)
		})
		Convey("A function with one Jar input should be marked for Jar", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Jar2"), Fixture{Url: "/abc"}, "GET", nil, nil) // "/abc",
			So(ep.jarInput, ShouldBeTrue)
		})
		Convey("Jar input in the begining should not work", func() {
			So(func() {
				NewEndPoint(NewMethodInvoker(&epMock{}, "Jar3"), Fixture{Url: "/abc/{d}"}, "GET", nil, nil) // "/abc/{d}",
			}, ShouldPanic)
			So(func() {
				NewEndPoint(NewMethodInvoker(&epMock{}, "Jar5"), Fixture{Url: "/abc/{e}"}, "GET", nil, nil) // "/abc/{e}",
			}, ShouldPanic)
		})
		Convey("Jar input at the end should be ok", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Jar4"), Fixture{Url: "/abc/{d}"}, "GET", nil, nil) // "/abc/{d}",
			So(ep.jarInput, ShouldBeTrue)
		})
	})
}

type verService struct {
	RestService   `root:"versioning"`
	api_version_1 GetApi `version:"1" url:"api"`
	api_version_2 GetApi `version:"2" url:"api"`
}

func (me *verService) Api_version_1() string { return "one" }
func (me *verService) Api_version_2() string { return "two" }

type newVerService struct {
	RestService   `root:"versioning"`
	api_version_3 GetApi `version:"3" url:"api"`
}

func (me *newVerService) Api_version_3() string { return "three" }

func TestVersionCapability(t *testing.T) {

	s := NewRestServer()
	s.AddService(&verService{})
	s.AddService(&newVerService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a GET endpoint specified as version 1", t, func() {
		Convey("Then the servers should return 404 for direct calls", func() {
			url := fmt.Sprintf("http://localhost:%d/versioning/api", s.Port)
			code, _, _ := getUrl(url, nil)
			So(code, ShouldEqual, 404)
		})
		Convey("Then the servers should honour urls with version prefix", func() {
			url := fmt.Sprintf("http://localhost:%d/v1/versioning/api", s.Port)
			code, _, content := getUrl(url, nil)
			So(code, ShouldEqual, 200)
			So(content, ShouldEqual, "one")
		})
		Convey("Then the servers should honour urls with accept headers of style1", func() {
			url := fmt.Sprintf("http://localhost:%d/versioning/api", s.Port)
			head := make(map[string]string)
			head["Accept"] = "application/" + defaults.Vendor + "-v1+json"
			code, _, content := getUrl(url, head)
			So(code, ShouldEqual, 200)
			So(content, ShouldEqual, "one")
		})
		Convey("Then the servers should honour urls with accept headers of style2", func() {
			url := fmt.Sprintf("http://localhost:%d/versioning/api", s.Port)
			head := make(map[string]string)
			head["Accept"] = "application/" + defaults.Vendor + "+json;version=1"
			code, _, content := getUrl(url, head)
			So(code, ShouldEqual, 200)
			So(content, ShouldEqual, "one")
		})
		Convey("Then an endpoint in the same service with the same url but different version should be independant", func() {
			url := fmt.Sprintf("http://localhost:%d/versioning/api", s.Port)
			head := make(map[string]string)
			head["Accept"] = "application/" + defaults.Vendor + "-v2+json"
			code, _, content := getUrl(url, head)
			So(code, ShouldEqual, 200)
			So(content, ShouldEqual, "two")
		})
		Convey("Then an endpoint in a different service with the same url but different version should be independant", func() {
			url := fmt.Sprintf("http://localhost:%d/versioning/api", s.Port)
			head := make(map[string]string)
			head["Accept"] = "application/" + defaults.Vendor + "-v3+json"
			code, _, content := getUrl(url, head)
			So(code, ShouldEqual, 200)
			So(content, ShouldEqual, "three")
		})
	})
}

type namingServ struct {
	RestService `root:"any" prefix:"day"`
	getapi      GetApi `version:"1.0" url:"api"`
}

func (me *namingServ) Getapi() string { return "whoa" }

func TestUrlNameConstruction(t *testing.T) {

	s := NewRestServer()
	s.AddService(&namingServ{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a GET endpoint specified with prefix, folder, version and url", t, func() {
		Convey("Then the complete url should be combination of above all", func() {
			url := fmt.Sprintf("http://localhost:%d/day/v1.0/any/api", s.Port)
			code, _, _ := getUrl(url, nil)
			So(code, ShouldEqual, 200)
		})
	})
}

type structService struct {
	RestService
	getStruct GetApi
}

func (me *structService) GetStruct() Fixture {
	return Fixture{
		Version: "1.2.3",
	}
}

func TestStructOutputIsAllowed(t *testing.T) {
	s := NewRestServer()
	s.AddService(&structService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a service endpoint that retuns a struct", t, func() {
		url := fmt.Sprintf("http://localhost:%d/struct/get-struct", s.Port)
		_, _, content := getUrl(url, nil)

		Convey("Then the field(s) of the struct should have the same value as passed", func() {
			var f Fixture
			json.Unmarshal([]byte(content), &f)
			So(f.Version, ShouldEqual, "1.2.3")
		})
	})
}
