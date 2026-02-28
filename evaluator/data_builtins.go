package evaluator

import (
	"base/object"
	"bytes"
	"encoding/csv"
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

func RegisterDataBuiltins() {
	builtins["csv.read"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `csv.read` must be STRING")
			}
			content, err := ioutil.ReadFile(path.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}
			r := csv.NewReader(bytes.NewReader(content))
			records, err := r.ReadAll()
			if err != nil {
				return newError("csv parse error: %s", err.Error())
			}
			return goTypeToBaseObject(records)
		},
	}

	builtins["yaml.write"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			data := args[1]
			if !ok1 {
				return newError("first argument to `yaml.write` must be STRING")
			}
			goData := baseObjectToGoType(data)
			yamlBytes, err := yaml.Marshal(goData)
			if err != nil {
				return newError("yaml marshal error: %s", err.Error())
			}
			err = ioutil.WriteFile(path.Value, yamlBytes, 0644)
			if err != nil {
				return newError("could not write file: %s", err.Error())
			}
			return TRUE
		},
	}
}
