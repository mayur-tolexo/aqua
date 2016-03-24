package aqua

// import (
// 	. "github.com/smartystreets/goconvey/convey"
// 	"testing"
// )

// type sacStruct struct {
// 	Field string
// }

// func TestSacSet(t *testing.T) {

// 	Convey("Given a Sac, Then Set() method", t, func() {
// 		s := NewSac()

// 		Convey("should accept literals", func() {
// 			s.Set("a-string", "bingo")
// 			So(s.Data["a-string"], ShouldEqual, "bingo")

// 			s.Set("an-int", 123)
// 			So(s.Data["an-int"], ShouldEqual, 123)
// 		})

// 		Convey("should accept a Sac", func() {
// 			b := NewSac().Set("sac-b", "value")
// 			s.Set("sac-a", b)

// 			m, ok := s.Data["sac-a"].(map[string]interface{})
// 			So(ok, ShouldBeTrue)
// 			So(m["sac-b"], ShouldEqual, "value")
// 		})

// 		Convey("should accept a map", func() {
// 			m := make(map[string]interface{})
// 			m["map"] = 123
// 			s.Set("a-map", m)

// 			m, ok := s.Data["a-map"].(map[string]interface{})
// 			So(ok, ShouldBeTrue)
// 			So(m["map"], ShouldEqual, 123)
// 		})

// 		Convey("should accept a struct", func() {
// 			st := sacStruct{Field: "oh"}

// 			s.Set("a-struct", st)

// 			m, ok := s.Data["a-struct"].(map[string]interface{})
// 			So(ok, ShouldBeTrue)
// 			So(m["Field"], ShouldEqual, "oh")
// 		})
// 	})
// }

// func TestSacMerge(t *testing.T) {
// 	Convey("Given a Sac, Then Merge() method", t, func() {
// 		s := NewSac()

// 		m1 := make(map[string]interface{})
// 		m1["a"] = "A"
// 		m2 := make(map[string]interface{})
// 		m2["b"] = "B"

// 		s.Merge(m1).Merge(m2).Merge(sacStruct{Field: "C"})

// 		val, ok := s.Data["a"]
// 		So(ok, ShouldBeTrue)
// 		So(val, ShouldEqual, "A")

// 		val, ok = s.Data["b"]
// 		So(ok, ShouldBeTrue)
// 		So(val, ShouldEqual, "B")

// 		val, ok = s.Data["Field"]
// 		So(ok, ShouldBeTrue)
// 		So(val, ShouldEqual, "C")
// 	})
// }
