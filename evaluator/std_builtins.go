package evaluator

import (
	"base/object"
	"math"
	"strings"
)

func RegisterStdBuiltins() {

	builtins["math.abs"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			switch arg := args[0].(type) {
			case *object.Integer:
				if arg.Value < 0 {
					return &object.Integer{Value: -arg.Value}
				}
				return arg
			case *object.Float:
				return &object.Float{Value: math.Abs(arg.Value)}
			default:
				return newError("argument to `math.abs` must be INTEGER or FLOAT, got %s", args[0].Type())
			}
		},
	}

	builtins["math.sqrt"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var val float64
			switch arg := args[0].(type) {
			case *object.Integer:
				val = float64(arg.Value)
			case *object.Float:
				val = arg.Value
			default:
				return newError("argument to `math.sqrt` must be INTEGER or FLOAT, got %s", args[0].Type())
			}
			return &object.Float{Value: math.Sqrt(val)}
		},
	}

	builtins["math.pow"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			var base, exponent float64

			getFloat := func(obj object.Object) (float64, bool) {
				if i, ok := obj.(*object.Integer); ok {
					return float64(i.Value), true
				}
				if f, ok := obj.(*object.Float); ok {
					return f.Value, true
				}
				return 0, false
			}

			var ok bool
			base, ok = getFloat(args[0])
			if !ok {
				return newError("first argument to `math.pow` must be numeric")
			}
			exponent, ok = getFloat(args[1])
			if !ok {
				return newError("second argument to `math.pow` must be numeric")
			}

			return &object.Float{Value: math.Pow(base, exponent)}
		},
	}

	builtins["math.round"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			if f, ok := args[0].(*object.Float); ok {
				return &object.Integer{Value: int64(math.Round(f.Value))}
			}
			if i, ok := args[0].(*object.Integer); ok {
				return i
			}
			return newError("argument to `math.round` must be numeric")
		},
	}

	builtins["math.sin"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var val float64
			switch arg := args[0].(type) {
			case *object.Integer:
				val = float64(arg.Value)
			case *object.Float:
				val = arg.Value
			default:
				return newError("argument to `math.sin` must be numeric")
			}
			return &object.Float{Value: math.Sin(val)}
		},
	}

	builtins["math.cos"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var val float64
			switch arg := args[0].(type) {
			case *object.Integer:
				val = float64(arg.Value)
			case *object.Float:
				val = arg.Value
			default:
				return newError("argument to `math.cos` must be numeric")
			}
			return &object.Float{Value: math.Cos(val)}
		},
	}

	builtins["math.log"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var val float64
			switch arg := args[0].(type) {
			case *object.Integer:
				val = float64(arg.Value)
			case *object.Float:
				val = arg.Value
			default:
				return newError("argument to `math.log` must be numeric")
			}
			return &object.Float{Value: math.Log10(val)}
		},
	}

	builtins["type"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	}

	builtins["string.upper"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			inputStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				inputStr = s.Value
			}
			return &object.String{Value: strings.ToUpper(inputStr)}
		},
	}

	builtins["string.lower"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			inputStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				inputStr = s.Value
			}
			return &object.String{Value: strings.ToLower(inputStr)}
		},
	}

	builtins["string.replace"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			s, ok1 := args[0].(*object.String)
			old, ok2 := args[1].(*object.String)
			new, ok3 := args[2].(*object.String)

			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `string.replace` must be STRING")
			}

			return &object.String{Value: strings.ReplaceAll(s.Value, old.Value, new.Value)}
		},
	}

	builtins["string.slice"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}

			// Auto-convert first argument to string
			inputStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				inputStr = s.Value
			}

			start, ok2 := args[1].(*object.Integer)
			if !ok2 {
				return newError("second argument to `string.slice` must be INTEGER")
			}

			startVal := int(start.Value)
			endVal := len(inputStr)

			if len(args) == 3 {
				if end, ok := args[2].(*object.Integer); ok {
					endVal = int(end.Value)
				}
			}

			// Bounds safety
			if startVal < 0 {
				startVal = 0
			}
			if endVal > len(inputStr) {
				endVal = len(inputStr)
			}
			if startVal > endVal {
				return &object.String{Value: ""}
			}

			return &object.String{Value: inputStr[startVal:endVal]}
		},
	}

	builtins["string.pad_left"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}

			// Auto-convert first argument to string
			inputStr := args[0].Inspect()
			if s, ok := args[0].(*object.String); ok {
				inputStr = s.Value
			}

			targetLen, ok2 := args[1].(*object.Integer)
			padChar, ok3 := args[2].(*object.String)

			if !ok2 || !ok3 {
				return newError("arguments to `string.pad_left` must be (ANY, INTEGER, STRING)")
			}

			str := inputStr
			pad := padChar.Value
			if len(pad) == 0 {
				return &object.String{Value: str}
			}

			for len(str) < int(targetLen.Value) {
				str = pad + str
			}

			return &object.String{Value: str}
		},
	}

	builtins["wait_all"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			env.Root().Wait()
			return NULL
		},
	}
}
