package aqua

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type aideService struct {
	RestService
	echo  GET
	echo2 GET
}

func (u *aideService) Echo(j Aide) string {
	j.LoadVars()
	return j.QueryVars["abc"]
}

func (u *aideService) Echo2(j Aide) string {
	return j.QueryVars["def"]
}

func TestJarForHttpGETMethod(t *testing.T) {

	s := NewRestServer()
	s.AddService(&aideService{})
	s.Port = getUniquePortForTestCase()
	s.RunAsync()

	Convey("Given a RestServer and a service", t, func() {
		Convey("Echo service should return Query String assigned to key: abc", func() {
			url := fmt.Sprintf("http://localhost:%d/aide/echo?abc=whatsUp", s.Port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "whatsUp")
		})
		Convey("Echo2 service should fail since LoadVars is not invoked", func() {
			url := fmt.Sprintf("http://localhost:%d/aide/echo2?def=hello", s.Port)
			_, _, content := getUrl(url, nil)
			So(content, ShouldEqual, "")
		})

	})
}
