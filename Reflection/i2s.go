package convert

import (
	"fmt"
	"reflect"
)

func recursiveCopy(json reflect.Value, struc reflect.Value) error {
	isObject := (json.Kind() == reflect.Map && struc.Kind() == reflect.Struct)
	isArray := (json.Kind() == reflect.Slice && struc.Kind() == reflect.Slice)
	if !(isObject || isArray) {
		return fmt.Errorf(
			"(in-out) Only object-struct or array-slice\n"+
				"in: %v\nout:%#v", json, struc,
		)
	}

	switch json.Kind() {

	case reflect.Map:
		iter := json.MapRange()
		for iter.Next() {
			jsonKey := iter.Key().String()
			jsonField := iter.Value().Elem()
			strucField := struc.FieldByName(jsonKey)

			switch jsonField.Kind() {
			case reflect.Bool, reflect.String, reflect.Float64:
				if strucField.Kind() == reflect.Int && jsonField.Kind() == reflect.Float64 {
					jsonField = jsonField.Convert(strucField.Type())
				}
				if jsonField.Kind() != strucField.Kind() {
					return fmt.Errorf(
						"Fields are not assigned\n"+
							"%v<->%#v", jsonField, strucField,
					)
				}

				strucField.Set(jsonField)
			case reflect.Map, reflect.Slice:
				err := recursiveCopy(jsonField, strucField)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf(
					"Field type <%s> not implemented\n%v",
					jsonField.Type(), json,
				)
			}
		}

	case reflect.Slice:
		struc.Set(reflect.MakeSlice(
			struc.Type(), json.Len(), json.Cap(),
		))

		for i := 0; i < json.Len(); i++ {
			err := recursiveCopy(
				json.Index(i).Elem(),
				struc.Index(i),
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func i2s(data interface{}, out interface{}) error {
	json, strucPtr := reflect.ValueOf(data), reflect.ValueOf(out)
	if strucPtr.Kind() != reflect.Ptr {
		return fmt.Errorf("Should be pointer in out value")
	}
	return recursiveCopy(json, strucPtr.Elem())
}
