package evaluator

import (
	"base/object"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

func RegisterWSBuiltins() {
	builtins["ws.connect"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2 (url, callback)", len(args))
			}
			urlStr, ok1 := args[0].(*object.String)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `ws.connect` must be (STRING, FUNCTION)")
			}

			header := http.Header{}
			dialer := websocket.Dialer{
				HandshakeTimeout: 10 * time.Second,
			}

			conn, _, err := dialer.Dial(urlStr.Value, header)
			if err != nil {
				return newError("ws.connect error: %s", err.Error())
			}

			go func() {
				defer conn.Close()
				for {
					_, message, err := conn.ReadMessage()
					if err != nil {
						fmt.Printf("WebSocket closed: %s\n", err.Error())
						return
					}
					msgObj := &object.String{Value: string(message)}
					applyFunction(env, fn, []object.Object{msgObj})
				}
			}()

			return TRUE
		},
	}
}
