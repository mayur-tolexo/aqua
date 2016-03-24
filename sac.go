package aqua

import (
	"reflect"

	"github.com/fatih/structs"
)

func init() {
	structs.DefaultTagName = "json"
}

type Sac1 struct {
	Data map[string]interface{}
}

func NewSac1() *Sac1 {
	return &Sac1{Data: make(map[string]interface{})}
}

func (me *Sac1) Set(key string, i interface{}) *Sac1 {

	if i == nil {
		me.Data[key] = nil
	} else {
		switch reflect.TypeOf(i).Kind() {
		case reflect.Struct:
			if s, ok := i.(Sac1); ok {
				me.Data[key] = s.Data
			} else {
				me.Data[key] = structs.Map(i)
			}
		case reflect.Map:
			me.Data[key] = i
		case reflect.Ptr:
			item := reflect.ValueOf(i).Elem().Interface()
			me.Set(key, item)
		default:
			me.Data[key] = i
		}
	}

	return me
}

// Item being merged must be a struct or a map
func (me *Sac1) Merge(i interface{}) *Sac1 {

	switch reflect.TypeOf(i).Kind() {
	case reflect.Struct:
		if s, ok := i.(Sac1); ok {
			me.Merge(s.Data)
		} else {
			me.Merge(structs.Map(i))
		}
	case reflect.Map:
		m := i.(map[string]interface{})
		for key, val := range m {
			if _, exists := me.Data[key]; exists {
				panic("Merge field already exists:" + key)
			} else {
				me.Data[key] = val
			}
		}
	default:
		panic("Can't merge something that is not struct or map")
	}

	return me
}
