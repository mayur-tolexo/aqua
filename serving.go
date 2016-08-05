package aqua

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/rightjoin/aero/ds"
	"github.com/rightjoin/aero/refl"
)

func writeOutput(w http.ResponseWriter, r *http.Request, signs []string, vals []reflect.Value, pretty string) {

	if len(signs) == 1 {
		if signs[0] == "int" {
			w.WriteHeader(int(vals[0].Int()))
		} else {
			writeItem(w, r, signs[0], vals[0], pretty)
		}
	} else if len(signs) == 2 {
		if signs[0] == "int" {
			// first thing would be an integer (http status code)
			w.WriteHeader(int(vals[0].Int()))
			// second be the payload
			writeItem(w, r, signs[1], vals[1], pretty)
		} else if signs[1] == "i:.error" {
			if vals[1].IsNil() {
				writeItem(w, r, signs[0], vals[0], pretty)
			} else {
				writeItem(w, r, signs[1], vals[1], pretty)
			}
		} else {
			panic("unsupported format")
		}
	}
}

func writeItem(w http.ResponseWriter, r *http.Request, sign string, val reflect.Value, pretty string) {

	// Dereference a pointer to a struct or slice
	if strings.HasPrefix(sign, "*st:") || strings.HasPrefix(sign, "*sl:") {
		o := val.Elem()
		writeItem(w, r, refl.TypeSignature(o.Type()), o, pretty)
		return
	}

	switch {
	case sign == "string":
		//fmt.Printf("Sign:%s, SignDynamic:%s, Val:%s, Val.I{}:%s, Val.String():%s", sign, getSignOfObject(val.Interface()), val, val.Interface(), val.String())
		v := val.Interface().(string)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", strconv.Itoa(len(v)))
		fmt.Fprintf(w, "%s", v)
	case sign == "st:github.com/rightjoin/aqua.Fault":
		f := val.Interface().(Fault)
		j, err := ds.ToBytes(f, pretty == "true" || pretty == "1")
		if err != nil {
			panic(err)
		}
		if f.HTTPCode != 0 {
			w.WriteHeader(f.HTTPCode)
		} else {
			// 417: Expectation failed
			switch r.Method {
			case "GET":
				w.WriteHeader(404)
			case "POST":
				w.WriteHeader(417)
			case "DELETE":
				w.WriteHeader(417)
			case "PUT":
				w.WriteHeader(444) // TODO: change
			default:
				panic(fmt.Sprintf("Status code missing for method: %", r.Method))
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.Write(j)
	case isError(val.Interface()) || sign == "i:.error":
		if val.IsNil() {
			m := map[string]interface{}{"success": 1}
			writeItem(w, r, refl.ObjSignature(m), reflect.ValueOf(m), pretty)
		} else {
			f, ok := val.Interface().(Fault)
			if !ok {
				f = Fault{
					Message: "Oops! An error occurred",
					Issue:   val.Interface().(error),
				}
			}
			writeItem(w, r, refl.ObjSignature(f), reflect.ValueOf(f), pretty)
		}
	case sign == "map":
		j, _ := ds.ToBytes(val.Interface(), pretty == "true" || pretty == "1")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.Write(j)
	case strings.HasPrefix(sign, "st:"):
		j, _ := ds.ToBytes(val.Interface(), pretty == "true" || pretty == "1")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.Write(j)
	case strings.HasPrefix(sign, "sl:"), strings.HasPrefix(sign, "ar:"):
		j, _ := ds.ToBytes(val.Interface(), pretty == "true" || pretty == "1")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(j)))
		w.Write(j)
	case sign == "i:.":
		writeItem(w, r, refl.ObjSignature(val.Interface()), val, pretty)
		// fmt.Println("interface{} resolves to:", getSignOfObject(val.Interface()))
		//TODO: error handling in case the returned object is an error
		//TODO: along with int, xx, also support xx, error as a function

	// case sign == "bool":
	// 	i := 0
	// 	if val.Bool() == true {
	// 		i = 1
	// 	}
	// 	j, _ := ds.ToBytes(map[string]interface{}{"success": i}, pretty == "true" || pretty == "1")
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.Header().Set("Content-Length", strconv.Itoa(len(j)))
	// 	w.Write(j)
	default:
		fmt.Printf("Don't know how to  %s?\n", sign)
	}
}
