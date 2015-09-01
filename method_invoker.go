package aqua

import (
	"fmt"
	"github.com/thejackrabbit/aero/db/orm"
	"github.com/thejackrabbit/aero/panik"
	"reflect"
	"strings"
)

// Responsibility:
// Used to store reference to either a method (of an object) or a global function.
// So it acts as a wrapper in this manner.
// This is then used at runtime to call/execute/invoke this reference by providing
// a simpler interface (Do).

type Invoker struct {
	addr   interface{}
	name   string
	exists bool

	outCount  int
	outParams []string

	inpCount  int
	inpParams []string

	db    bool
	model interface{}
}

func NewMethodInvoker(addr interface{}, method string) Invoker {
	out := &Invoker{
		addr:   addr,
		name:   method,
		exists: false,
	}

	// validation
	symb := getSignOfObject(addr)
	panik.If(!strings.HasPrefix(symb, "*st:"), "Invoker expects address of a struct")

	var m reflect.Method
	m, out.exists = reflect.TypeOf(out.addr).MethodByName(out.name)
	if out.exists {
		out.decipherOutputs(m.Type)
		out.decipherInputs(m.Type)
	}

	return *out
}

func (me *Invoker) decipherOutputs(mt reflect.Type) {

	me.outCount = mt.NumOut()
	me.outParams = make([]string, mt.NumOut())

	for i := 0; i < mt.NumOut(); i++ {
		pt := mt.Out(i)
		me.outParams[i] = getSignOfType(pt)
	}
}

func (me *Invoker) decipherInputs(mt reflect.Type) {

	me.inpCount = mt.NumIn() - 1 // skip the first param (me)
	me.inpParams = make([]string, mt.NumIn()-1)

	for i := 1; i < mt.NumIn(); i++ {
		pt := mt.In(i)
		me.inpParams[i-1] = getSignOfType(pt)
	}
}

func (me *Invoker) Do(v []reflect.Value) []reflect.Value {
	if me.db {
		ReadModel(me.model, v[0].String())
		return []reflect.Value{reflect.ValueOf(me.model)}
	}
	return reflect.ValueOf(me.addr).MethodByName(me.name).Call(v)
}

func (me *Invoker) Pr() {
	fmt.Printf("%s.%s has %d inputs and %d outParamsputs\n", me.addr, me.name, me.inpCount, me.outCount)
	for i, s := range me.outParams {
		fmt.Printf(" outParams -> %s && %s\n", s, me.outParams[i])
	}
	for i, s := range me.inpParams {
		fmt.Printf(" inpParams -> %s && %s\n", s, me.inpParams[i])
	}
}

func ReadModel(modl interface{}, pk string) {
	fmt.Println(pk)
	my := orm.Get(false)
	my.SingularTable(true)
	my.Debug().First(modl, pk)
	fmt.Println(modl)
}
