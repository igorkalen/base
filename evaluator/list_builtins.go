package evaluator

import (
	"base/object"
	"sort"
)

func RegisterListBuiltins() {
	builtins["list.length"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.length` must be ARRAY, got %s", args[0].Type())
			}
			return &object.Integer{Value: int64(len(arr.Elements))}
		},
	}

	builtins["list.map"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `list.map` must be (ARRAY, FUNCTION)")
			}

			newElements := make([]object.Object, len(arr.Elements))
			for i, el := range arr.Elements {
				newElements[i] = applyFunction(env, fn, []object.Object{el})
			}
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.filter"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `list.filter` must be (ARRAY, FUNCTION)")
			}

			newElements := []object.Object{}
			for _, el := range arr.Elements {
				res := applyFunction(env, fn, []object.Object{el})
				if res == TRUE {
					newElements = append(newElements, el)
				}
			}
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.contains"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `list.contains` must be ARRAY")
			}
			target := args[1]
			for _, el := range arr.Elements {
				if el.Inspect() == target.Inspect() { 
					return TRUE
				}
			}
			return FALSE
		},
	}

	builtins["list.sort"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.sort` must be ARRAY")
			}
			
			newElements := make([]object.Object, len(arr.Elements))
			copy(newElements, arr.Elements)
			sort.Slice(newElements, func(i, j int) bool {
				return newElements[i].Inspect() < newElements[j].Inspect()
			})
			return &object.Array{Elements: newElements}
		},
	}
}
