package aqua

import (
	"bytes"
	"encoding/json"
	// "github.com/mgutz/logxi/v1"
	"reflect"
	"strings"

	"github.com/rightjoin/aero/panik"
	"github.com/rightjoin/aero/refl"
)

// var logEnc = log.New("enc")

func encode(r []reflect.Value, typ []string) []byte {

	// logEnc.Info("entering encode()", "vals", r, "types", typ)

	buf := new(bytes.Buffer)
	encd := json.NewEncoder(buf)

	for i, _ := range r {
		encodeItem(encd, r[i], typ[i])
	}

	// logEnc.Info("leaving encode()", "data", string(buf.Bytes()))

	return buf.Bytes()
}

func encodeItem(j *json.Encoder, r reflect.Value, t string) {

	// logEnc.Info("entering encodeItem()", "val", r, "type", t)

	switch {
	case t == "int":
		//		logEnc.Info("encode.int")
		panik.On(j.Encode(r.Int()))
	case t == "map":
		// logEnc.Info("encode.map")
		panik.On(j.Encode(r.Interface().(map[string]interface{})))
	case t == "string":
		// logEnc.Info("encode.string")
		panik.On(j.Encode(r.String()))
	case t == "i:.":
		// logEnc.Info("encode.i{}/")
		s := refl.ObjSignature(r.Interface())
		panik.On(j.Encode(s))
		encodeItem(j, r, s)
	case strings.HasPrefix(t, "st:"):
		// logEnc.Info("encode.struct")
		panik.On(j.Encode(r.Interface()))
	case strings.HasPrefix(t, "sl:"):
		// logEnc.Info("encode.slice")
		panik.On(j.Encode(r.Interface()))
	default:
		panik.Do("Can't encode '%s' for endpoint cache", t)
	}
}

func decode(data []byte, typ []string) []reflect.Value {
	buf := bytes.NewBuffer(data)
	decd := json.NewDecoder(buf)
	r := make([]reflect.Value, len(typ))
	for i, _ := range typ {
		r[i] = decodeItem(decd, typ[i])
	}
	return r
}

func decodeItem(j *json.Decoder, t string) reflect.Value {
	var r reflect.Value
	switch {
	case t == "int":
		var i int
		panik.On(j.Decode(&i))
		r = reflect.ValueOf(i)
	case t == "map":
		var m map[string]interface{}
		panik.On(j.Decode(&m))
		r = reflect.ValueOf(m)
	case t == "string":
		var s string
		panik.On(j.Decode(&s))
		r = reflect.ValueOf(s)
	case t == "i:.":
		var s string
		panik.On(j.Decode(&s))
		r = decodeItem(j, s)
	case strings.HasPrefix(t, "st:"):
		var m map[string]interface{}
		panik.On(j.Decode(&m))
		r = reflect.ValueOf(m)
	case strings.HasPrefix(t, "sl:"):
		var a []interface{}
		panik.On(j.Decode(&a))
		r = reflect.ValueOf(a)
	default:
		panik.Do("Can't decdoe '%s' for endpoint cache", t)
	}

	return r
}
