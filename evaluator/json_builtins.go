package evaluator

import (
	"base/object"
	"encoding/json"
)

func RegisterJSONBuiltins() {
	builtins["json.parse"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			strObj, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `json.parse` must be STRING, got %s", args[0].Type())
			}

			
			var result interface{}
			err := json.Unmarshal([]byte(strObj.Value), &result)
			if err != nil {
				return newError("failed to parse JSON: %s", err.Error())
			}

			
			return goTypeToBaseObject(result)
		},
	}

	builtins["json.stringify"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			
			goVal := baseObjectToGoType(args[0])

			jsonBytes, err := json.MarshalIndent(goVal, "", "  ")
			if err != nil {
				return newError("failed to stringify object: %s", err.Error())
			}

			return &object.String{Value: string(jsonBytes)}
		},
	}
}
