package evaluator

import (
	"base/object"
	"sync"
)

func RegisterChannelBuiltins() {
	builtins["chan"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			ch := &ChannelObject{
				items: []object.Object{},
				mu:    &sync.Mutex{},
			}
			sendFn := &object.Builtin{
				Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
					if len(innerArgs) != 1 {
						return newError("chan.send needs exactly 1 argument")
					}
					ch.mu.Lock()
					ch.items = append(ch.items, innerArgs[0])
					ch.mu.Unlock()
					return NULL
				},
			}
			readAllFn := &object.Builtin{
				Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
					ch.mu.Lock()
					defer ch.mu.Unlock()
					elements := make([]object.Object, len(ch.items))
					copy(elements, ch.items)
					return &object.Array{Elements: elements}
				},
			}
			return &object.Hash{
				Pairs: map[string]object.Object{
					"send":     sendFn,
					"read_all": readAllFn,
				},
			}
		},
	}
}

type ChannelObject struct {
	items []object.Object
	mu    *sync.Mutex
}
