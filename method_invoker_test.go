package aqua

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

type invokerMock struct{}

func (me *invokerMock) InputZero()                                                             {}
func (me *invokerMock) InputInt(a int)                                                         {}
func (me *invokerMock) InputMulti(a int, b *string, c iInvoker, d invokerMock, e *invokerMock) {}

func (me *invokerMock) OutputZero() {}
func (me *invokerMock) OutputInt() int {
	return 0
}
func (me *invokerMock) OutputMulti() (int, *string, iInvoker, invokerMock, *invokerMock) {
	s := "string"
	return 0, &s, invokerMock{}, invokerMock{}, &invokerMock{}
}

type iInvoker interface{}

func TestInvokerCtor(t *testing.T) {

	Convey("Given that Invocker::ctor expects address of struct", t, func() {
		Convey("Then it should throw panic on being passed a literal", func() {
			So(func() { NewMethodInvoker("abc", "any") }, ShouldPanic)
		})

		Convey("Then it should throw panic on being passed a struct", func() {
			So(func() { NewMethodInvoker(invokerMock{}, "any") }, ShouldPanic)
		})

		Convey("Then it should accept struct address safely", func() {
			So(func() { NewMethodInvoker(&invokerMock{}, "InputZero") }, ShouldNotPanic)
		})
	})
}

func TestInputParameters(t *testing.T) {

	a := &invokerMock{}

	Convey("The Invoker", t, func() {
		Convey("Should correctly identify no/zero inpParamsuts", func() {
			i := NewMethodInvoker(a, "InputZero")
			So(i.inpCount, ShouldEqual, 0)
		})
		Convey("Should correctly identify int inpParamsuts", func() {
			i := NewMethodInvoker(a, "InputInt")
			So(i.inpCount, ShouldEqual, 1)
			So(i.inpParams[0], ShouldEqual, "int")
		})
		Convey("Should correctly identify multi inpParamsut parameters", func() {
			i := NewMethodInvoker(a, "InputMulti")
			So(i.inpCount, ShouldEqual, 5)
			So(i.inpParams[0], ShouldEqual, "int")
			So(i.inpParams[1], ShouldEqual, "*string")
			So(i.inpParams[2], ShouldEqual, "i:github.com/rightjoin/aqua.iInvoker")
			So(i.inpParams[3], ShouldEqual, "st:github.com/rightjoin/aqua.invokerMock")
			So(i.inpParams[4], ShouldEqual, "*st:github.com/rightjoin/aqua.invokerMock")
		})
	})
}

func TestOutputParameters(t *testing.T) {

	a := &invokerMock{}

	Convey("The Invoker", t, func() {
		Convey("Should correctly identify no/zero outParamsput", func() {
			i := NewMethodInvoker(a, "OutputZero")
			So(i.outCount, ShouldEqual, 0)
		})
		Convey("Should correctly identify int outParamsput", func() {
			i := NewMethodInvoker(a, "OutputInt")
			So(i.outCount, ShouldEqual, 1)
			So(i.outParams[0], ShouldEqual, "int")
		})
		Convey("Should correctly identify multi outParamsput parameters", func() {
			i := NewMethodInvoker(a, "OutputMulti")
			So(i.outCount, ShouldEqual, 5)
			So(i.outParams[0], ShouldEqual, "int")
			So(i.outParams[1], ShouldEqual, "*string")
			So(i.outParams[2], ShouldEqual, "i:github.com/rightjoin/aqua.iInvoker")
			So(i.outParams[3], ShouldEqual, "st:github.com/rightjoin/aqua.invokerMock")
			So(i.outParams[4], ShouldEqual, "*st:github.com/rightjoin/aqua.invokerMock")
		})
	})
}
