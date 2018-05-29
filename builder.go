package simplehttp

import (
	"fmt"
	"log"
	"reflect"
	"time"
)

//Use the schema tags to build the url
func (sh *SimpleHttp) BuildUrl(url string, values interface{}) string {

	u := url
	first := true

	s := reflect.TypeOf(values)
	v := reflect.ValueOf(values)
	for i := 0; i < s.NumField(); i++ {
		sf := s.Field(i)
		vf := v.FieldByName(sf.Name)
		val := vf.Interface()
		tag := sf.Tag.Get("schema")

		if len(tag) > 0 {
			insert := false
			var override string

			switch t := val.(type) {
			case uint64:
				if t > 0 {
					insert = true
				}
			case string:
				if len(t) > 0 {
					insert = true
				}
			case time.Time:
				insert = !t.IsZero()
				if insert {
					override = fmt.Sprintf("%s", t.UTC().Format(time.RFC3339))
				}
			default:
				log.Printf("Unhandled type for %s: %v\n", sf.Name, vf.Type())
			}

			if insert {
				if first {
					u = u + "?"
					first = false
				} else {
					u = u + "&"
				}

				if len(override) == 0 {
					u = u + fmt.Sprintf("%s=%v", tag, val)
				} else {
					u = u + fmt.Sprintf("%s=%s", tag, override)
				}
			}
		}

	}

	return u
}
