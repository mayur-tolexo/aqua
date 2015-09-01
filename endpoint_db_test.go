package aqua

//
//import (
//	"fmt"
//	. "github.com/smartystreets/goconvey/convey"
//	"testing"
//)
//
//type dbBindService struct {
//	RestService
//
//	autoGet CrudApi
//}
//
//type User struct {
//}
//
//func (me *dbBindService) AutoGet() CrudApi {
//	return CrudApi{
//		Model: User{},
//	}
//}
//
//func TestDBEndpoint(t *testing.T) {
//
//	s := NewRestServer()
//	s.AddService(&dbBindService{})
//	s.Port = getUniquePortForTestCase()
//	s.RunAsync()
//
//	Convey("Given a DB bound endpoint", t, func() {
//		Convey("The rest server should be bound to CRUD methods", func() {
//			_, ok := s.apis[cleanUrl("GET", "db-bind", "auto-get")]
//			fmt.Println(s.apis)
//			So(ok, ShouldBeTrue)
//		})
//	})
//}
//
//// TODO: any function with dbApi return should not take any parameter
//
//// TODO: db api's should not have only 1 parameterized api
