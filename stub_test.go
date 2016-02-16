package aqua

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type stubService struct {
	RestService
	mock       GET `stub:"mocks/mock.json"`
	mockNoFile GET `stub:"mocks/missing.json"`
}

func TestStubFileMissing(t *testing.T) {

	s := NewRestServer()
	s.AddService(&stubService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a service stub", t, func() {
		Convey("When the corresponding stub file is missing in current AND executable dir", func() {
			Convey("Then the server should return 400 series error", func() {
				url := fmt.Sprintf("http://localhost:%d/stub/mock-no-file", s.Port)
				code, _, content := getUrl(url, nil)
				So(code, ShouldEqual, 400)
				fmt.Println(content)
			})
		})
	})

}

func TestMockStub(t *testing.T) {

	s := NewRestServer()
	s.AddService(&stubService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a service stub", t, func() {
		Convey("When the corresponding stub file is found in current OR executable dir", func() {
			Convey("Then the server should return content of file", func() {
				url := fmt.Sprintf("http://localhost:%d/stub/mock", s.Port)
				_, _, content := getUrl(url, nil)
				So(content, ShouldEqual, "MOCK DATA")
			})
		})
	})

}
