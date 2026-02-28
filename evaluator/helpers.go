package evaluator

import (
	"base/object"
	"fmt"
	"time"
)

func baseObjectToGoType(obj object.Object) interface{} {
	switch o := obj.(type) {
	case *object.Integer:
		return o.Value
	case *object.Float:
		return o.Value
	case *object.String:
		return o.Value
	case *object.Boolean:
		return o.Value
	case *object.Null:
		return nil
	case *object.Array:
		arr := make([]interface{}, len(o.Elements))
		for i, el := range o.Elements {
			arr[i] = baseObjectToGoType(el)
		}
		return arr
	case *object.Hash:
		hashMap := make(map[string]interface{})
		for k, v := range o.Pairs {
			hashMap[k] = baseObjectToGoType(v)
		}
		return hashMap
	default:
		return fmt.Sprintf("<unserializable_type:%s>", o.Type())
	}
}

func goTypeToBaseObject(val interface{}) object.Object {
	switch v := val.(type) {
	case time.Time:
		return &object.String{Value: v.Format(time.RFC3339)}
	case float64:
		if v == float64(int64(v)) {
			return &object.Integer{Value: int64(v)}
		}
		return &object.Float{Value: v}
	case float32:
		return &object.Float{Value: float64(v)}
	case int64:
		return &object.Integer{Value: v}
	case int:
		return &object.Integer{Value: int64(v)}
	case string:
		return &object.String{Value: v}
	case []byte:
		return &object.String{Value: string(v)}
	case bool:
		if v {
			return TRUE
		}
		return FALSE
	case nil:
		return NULL
	case []string:
		elements := make([]object.Object, len(v))
		for i, el := range v {
			elements[i] = &object.String{Value: el}
		}
		return &object.Array{Elements: elements}
	case []interface{}:
		elements := make([]object.Object, len(v))
		for i, el := range v {
			elements[i] = goTypeToBaseObject(el)
		}
		return &object.Array{Elements: elements}
	case map[string]interface{}:
		pairs := make(map[string]object.Object)
		for k, el := range v {
			pairs[k] = goTypeToBaseObject(el)
		}
		return &object.Hash{Pairs: pairs}
	default:
		return newError("unsupported type in conversion: %T", v)
	}
}
