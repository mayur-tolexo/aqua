package aqua

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"testing"

	"github.com/mayur-tolexo/aero/db/cstr"
	"github.com/mayur-tolexo/aero/ds"
	. "github.com/smartystreets/goconvey/convey"
)

type epMock struct{}

func (me *epMock) Handler1(w http.ResponseWriter, r *http.Request) {}

func TestStandardHttpHandlerIsIdentifiedCorrectly(t *testing.T) {
	Convey("Given an endpoint and a Service Controller", t, func() {
		Convey("The standard http handler should be identified correctly", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Handler1"), Fixture{}, "GET", nil, nil, nil)
			So(ep.stdHandler, ShouldBeTrue)
		})
	})
}

func (me *epMock) Aide1(w http.ResponseWriter, r *http.Request) {}
func (me *epMock) Aide2(j Aide) string                          { return "" }
func (me *epMock) Aide3(j Aide, s string) string                { return "" }
func (me *epMock) Aide4(s string, j Aide) string                { return "" }
func (me *epMock) Aide5(j Aide, k Aide) string                  { return "" }

func TestAideInputIsIdentifiedCorrectly(t *testing.T) {
	Convey("Given an endpoint and a Service Controller", t, func() {
		Convey("A standard http function should NOT be marked for Aide", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Aide1"), Fixture{Url: "/abc/{d}/{e}"}, "GET", nil, nil, nil) //"/abc/{d}/{e}",
			So(ep.needsAide, ShouldBeFalse)
		})
		Convey("A function with one Aide input should be marked for Aide", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Aide2"), Fixture{Url: "/abc"}, "GET", nil, nil, nil) // "/abc",
			So(ep.needsAide, ShouldBeTrue)
		})
		Convey("Aide input in the begining should not work", func() {
			So(func() {
				NewEndPoint(NewMethodInvoker(&epMock{}, "Aide3"), Fixture{Url: "/abc/{d}"}, "GET", nil, nil, nil) // "/abc/{d}",
			}, ShouldPanic)
			So(func() {
				NewEndPoint(NewMethodInvoker(&epMock{}, "Aide5"), Fixture{Url: "/abc/{e}"}, "GET", nil, nil, nil) // "/abc/{e}",
			}, ShouldPanic)
		})
		Convey("Aide input at the end should be ok", func() {
			ep := NewEndPoint(NewMethodInvoker(&epMock{}, "Aide4"), Fixture{Url: "/abc/{d}"}, "GET", nil, nil, nil) // "/abc/{d}",
			So(ep.needsAide, ShouldBeTrue)
		})
	})
}

type verService struct {
	RestService   `root:"versioning"`
	api_version_1 GET `version:"1" url:"api"`
	api_version_2 GET `version:"2" url:"api"`
}

func (me *verService) Api_version_1() string { return "one" }
func (me *verService) Api_version_2() string { return "two" }

type newVerService struct {
	RestService   `root:"versioning"`
	api_version_3 GET `version:"3" url:"api"`
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
	getapi      GET `version:"1.0" url:"api"`
	noversion   GET `url:"noversion-here"`
}

func (me *namingServ) Getapi() string { return "whoa" }

func (me *namingServ) Noversion() string { return "cool" }

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

	Convey("Given a GET endpoint specified with prefix, folder, url but no version", t, func() {
		Convey("Then the complete url should be combination of above all", func() {
			url := fmt.Sprintf("http://localhost:%d/day/any/noversion-here", s.Port)
			code, _, _ := getUrl(url, nil)
			So(code, ShouldEqual, 200)
		})
	})

}

type dataService struct {
	RestService
	getStruct  GET
	getStructI GET
	getString  GET
	getStringI GET
	getMap     GET
	getMapI    GET
	getSlice   GET
	getSliceI  GET
}

func (me *dataService) GetStruct() Fixture {
	return Fixture{
		Version: "1.2.3",
	}
}

func (me *dataService) GetStructI() interface{} {
	return Fixture{
		Version: "1.2.3.4",
	}
}

func (me *dataService) GetString() string {
	return "5"
}

func (me *dataService) GetStringI() interface{} {
	return "5.5"
}

func (me *dataService) GetMap() map[string]interface{} {
	m := map[string]interface{}{"whats": "up", "num": 1234}
	return m
}

func (me *dataService) GetMapI() interface{} {
	m := map[string]interface{}{"whats": "up", "num": 12345}
	return m
}

func (me *dataService) GetSlice() []string {
	return []string{"one", "two"}
}

func (me *dataService) GetSliceI() interface{} {
	return []string{"three", "four"}
}

func TestAllOutputDataFormats(t *testing.T) {
	s := NewRestServer()
	s.AddService(&dataService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a service that provides all data formats", t, func() {

		Convey("Then the struct output should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-struct", s.Port)
			_, _, content := getUrl(url, nil)
			var f Fixture
			json.Unmarshal([]byte(content), &f)
			So(f.Version, ShouldEqual, "1.2.3")
		})

		Convey("Then the struct output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-struct-i", s.Port)
			_, _, content := getUrl(url, nil)
			var f Fixture
			json.Unmarshal([]byte(content), &f)
			So(f.Version, ShouldEqual, "1.2.3.4")
		})

		Convey("Then the string output should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-string", s.Port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "5")
		})

		Convey("Then the string output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-string-i", s.Port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "5.5")
		})

		Convey("Then the map output should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-map", s.Port)
			_, _, content := getUrl(url, nil)
			var m map[string]interface{}
			json.Unmarshal([]byte(content), &m)
			So(m["whats"], ShouldEqual, "up")
			So(m["num"], ShouldEqual, 1234)
		})

		Convey("Then the map output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-map-i", s.Port)
			_, _, content := getUrl(url, nil)
			var m map[string]interface{}
			json.Unmarshal([]byte(content), &m)
			So(m["whats"], ShouldEqual, "up")
			So(m["num"], ShouldEqual, 12345)
		})

		Convey("Then the [slice] output should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-slice", s.Port)
			_, _, content := getUrl(url, nil)
			var m []interface{}
			json.Unmarshal([]byte(content), &m)
			So(m[0], ShouldEqual, "one")
			So(m[1], ShouldEqual, "two")
		})

		Convey("Then the [slice] output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/data/get-slice-i", s.Port)
			_, _, content := getUrl(url, nil)
			var m []interface{}
			json.Unmarshal([]byte(content), &m)
			So(m[0], ShouldEqual, "three")
			So(m[1], ShouldEqual, "four")
		})

	})
}

type errService struct {
	RestService
	getErrorI  GET
	getFaultI  GET
	postErrorI POST
}

func (me *errService) GetErrorI() interface{} {
	return errors.New("bingo-error")
}

func (me *errService) GetFaultI() interface{} {
	return Fault{
		Message: "there was an error",
		Issue:   errors.New("bingo-error"),
	}
}

func (me *errService) PostErrorI() interface{} {
	return Fault{
		Message: "there was an error",
		Issue:   errors.New("shingo-error"),
	}
}

func TestErrorFormats(t *testing.T) {
	s := NewRestServer()
	s.AddService(&errService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a service that provides all data formats", t, func() {

		Convey("Then the error output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/err/get-error-i", s.Port)
			code, _, content := getUrl(url, nil)
			So(code, ShouldEqual, 404)
			var m map[string]interface{}
			json.Unmarshal([]byte(content), &m)
			val, _ := m["message"]
			So(val, ShouldEqual, "Oops! An error occurred")
			val2, _ := m["issue"]
			So(val2, ShouldEqual, "bingo-error")
		})

		Convey("Then the Fault output for interface{} should work", func() {
			url := fmt.Sprintf("http://localhost:%d/err/get-fault-i", s.Port)
			code, _, content := getUrl(url, nil)
			So(code, ShouldEqual, 404)
			var m map[string]interface{}
			json.Unmarshal([]byte(content), &m)
			val, _ := m["message"]
			So(val, ShouldEqual, "there was an error")
			val2, _ := m["issue"]
			So(val2, ShouldEqual, "bingo-error")
		})

		Convey("Then the Fault output for interface{} should work for a Post request", func() {
			url := fmt.Sprintf("http://localhost:%d/err/post-error-i", s.Port)
			code, _, _ := postUrl(url, nil, nil)
			So(code, ShouldEqual, 417)
		})
	})
}

type TwoParams struct {
	RestService
	getStruct GET
	getString GET
	getMap    GET
	getSlice  GET
	getI      GET

	getStruct2 GET
	getString2 GET
	getMap2    GET
	getSlice2  GET
	getI2      GET

	getString3 GET
	getFault3  GET
}

func (s *TwoParams) GetStruct() (int, Fixture) {
	return 200, Fixture{}
}

func (s *TwoParams) GetString() (int, string) {
	return 200, "abc"
}

func (s *TwoParams) GetMap() (int, map[string]interface{}) {
	var m map[string]interface{} = make(map[string]interface{})
	return 200, m
}

func (s *TwoParams) GetSlice() (int, []Fixture) {
	return 200, []Fixture{
		{},
	}
}

func (s *TwoParams) GetI() (int, interface{}) {
	return 200, 12345
}

func (s *TwoParams) GetStruct2() (Fixture, error) {
	return Fixture{Allow: "+-"}, nil
}

func (s *TwoParams) GetString2() (string, error) {
	return "abc2", nil
}

func (s *TwoParams) GetMap2() (map[string]interface{}, error) {
	var m map[string]interface{} = make(map[string]interface{})
	m["is-it"] = "really-true"
	return m, nil
}

func (s *TwoParams) GetSlice2() ([]Fixture, error) {
	return []Fixture{
		{Allow: "+"}, {Deny: "-"},
	}, nil
}

func (s *TwoParams) GetI2() (interface{}, error) {
	return "12345", nil
}

func (s *TwoParams) GetString3() (string, error) {
	return "", errors.New("whatsup")
}

func (s *TwoParams) GetFault3() (string, error) {
	return "", Fault{
		HTTPCode: 410,
		Message:  "there it is",
		Issue:    errors.New("whatsup"),
	}
}

func TestServicesReturning2Params(t *testing.T) {

	Convey("Given a service that has services returning 2 parameters", t, func() {
		Convey("Then returning int (status code) followed by is map/struct/interface/string/slice is acceptable", func() {
			So(func() {
				s := NewRestServer()
				s.AddService(&TwoParams{})
				s.Port = getUniquePortForTestCase()
				s.RunAsync()
			}, ShouldNotPanic)
		})
	})
}

func TestServiceReturnsXPlusError(t *testing.T) {

	Convey("Given a RestServer", t, func() {

		Convey("Endpoint implementations of with 2 return types <something>, error should be supported", func() {

			s := NewRestServer()
			s.AddService(&TwoParams{})
			s.Port = getUniquePortForTestCase()
			s.RunAsync()

			Convey("And when error is nil", func() {

				Convey("Then string <something> must be sent to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-string2", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 200)
					So(data, ShouldEqual, "abc2")
				})

				Convey("Then struct <something> must be sent to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-struct2", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 200)
					var js map[string]interface{}
					ds.Load(&js, []byte(data))
					So(js["Allow"], ShouldEqual, "+-")
				})

				Convey("Then slice <something> must be sent to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-slice2", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 200)
					var js []interface{}
					ds.Load(&js, []byte(data))
					So(len(js), ShouldEqual, 2)
					So(js[0].(map[string]interface{})["Allow"], ShouldEqual, "+")
				})

				Convey("Then map <something> must be sent to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-map2", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 200)
					var js map[string]interface{}
					ds.Load(&js, []byte(data))
					So(js["is-it"], ShouldEqual, "really-true")
				})

				Convey("Then interface <something> must be sent to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-i2", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 200)
					inty, err := strconv.Atoi(data)
					fmt.Println(inty, err)
					So(inty, ShouldEqual, 12345)
					So(err, ShouldBeNil)
				})
			})

			Convey("And when error is not nil", func() {

				Convey("Then error must be returned to caller", func() {
					url := fmt.Sprintf("http://localhost:%d/two-params/get-string3", s.Port)
					code, _, data := getUrl(url, nil)
					So(code, ShouldEqual, 404)
					var js map[string]interface{}
					ds.Load(&js, []byte(data))
					So(js["message"], ShouldEqual, "Oops! An error occurred")
					So(js["issue"], ShouldEqual, "whatsup")
				})

				Convey("And if error is a Fault", func() {

					Convey("Then Fault, along with any custom http status must be returned", func() {
						url := fmt.Sprintf("http://localhost:%d/two-params/get-fault3", s.Port)
						code, _, data := getUrl(url, nil)
						So(code, ShouldEqual, 410)
						var js map[string]interface{}
						ds.Load(&js, []byte(data))
						So(js["message"], ShouldEqual, "there it is")
						So(js["issue"], ShouldEqual, "whatsup")
					})

				})
			})

		})

	})
}

type someModel struct {
}

type crudOut1Service struct {
	RestService
	outMethod CRUD
}

func (s *crudOut1Service) OutMethod() CRUD {
	return CRUD{
		//Storage: cstr.Storage{Engine: "mysql", Conn: "blah"},
		Storage: cstr.Storage{
			Engine: "mysql",
			Conn:   "blah",
		},
		Model: func() (interface{}, interface{}) {
			return &someModel{}, nil
		},
	}
}

type crudOut2Service struct {
	RestService
	outMethod CRUD
}

func (s *crudOut2Service) OutMethod() (int, string) {
	return 200, "blah"
}

type crudOut3Service struct {
	RestService
	outMethod CRUD
}

func (s *crudOut3Service) OutMethod() string {
	return "something"
}

func TestCrudMethodOutput(t *testing.T) {

	Convey("Given a CRUD api endpoint", t, func() {
		Convey("Then its method output must return 1 item only", func() {
			Convey("And its return type must be CrudApi", func() {
				s := NewRestServer()
				s.AddService(&crudOut1Service{})
				s.Port = getUniquePortForTestCase()
				So(func() {
					s.RunAsync()
				}, ShouldNotPanic)
			})

		})

		Convey("Then it must not return 0 or more than 1 outputs", func() {
			s := NewRestServer()
			s.AddService(&crudOut2Service{})
			s.Port = getUniquePortForTestCase()
			So(func() {
				s.RunAsync()
			}, ShouldPanic)

		})

		Convey("Then return string or int would panic", func() {
			s := NewRestServer()
			s.AddService(&crudOut3Service{})
			s.Port = getUniquePortForTestCase()
			So(func() {
				s.RunAsync()
			}, ShouldPanic)
		})

	})
}
