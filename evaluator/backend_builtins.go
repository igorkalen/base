package evaluator

import (
	"base/object"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RegisterBackendBuiltins() {
	builtins["http.get"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `http.get` must be STRING")
			}
			var opts *object.Hash
			if len(args) > 1 {
				opts, _ = args[1].(*object.Hash)
			}
			return doHTTPRequest("GET", urlStr.Value, nil, opts)
		},
	}

	builtins["http.post"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `http.post` must be STRING")
			}
			var opts *object.Hash
			if len(args) > 2 {
				opts, _ = args[2].(*object.Hash)
			}
			return doHTTPRequest("POST", urlStr.Value, args[1], opts)
		},
	}

	builtins["http.put"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `http.put` must be STRING")
			}
			var opts *object.Hash
			if len(args) > 2 {
				opts, _ = args[2].(*object.Hash)
			}
			return doHTTPRequest("PUT", urlStr.Value, args[1], opts)
		},
	}

	builtins["http.patch"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("first argument to `http.patch` must be STRING")
			}
			var opts *object.Hash
			if len(args) > 2 {
				opts, _ = args[2].(*object.Hash)
			}
			return doHTTPRequest("PATCH", urlStr.Value, args[1], opts)
		},
	}

	builtins["http.delete"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `http.delete` must be STRING")
			}
			var opts *object.Hash
			if len(args) > 1 {
				opts, _ = args[1].(*object.Hash)
			}
			return doHTTPRequest("DELETE", urlStr.Value, nil, opts)
		},
	}

	builtins["http.ping"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			urlStr, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `http.ping` must be STRING")
			}
			timeout := 5 * time.Second
			if len(args) > 1 {
				if opts, ok := args[1].(*object.Hash); ok {
					if t, ok := opts.Pairs["timeout"]; ok {
						if tInt, ok := t.(*object.Integer); ok {
							timeout = time.Duration(tInt.Value) * time.Second
						}
					}
				}
			}
			client := &http.Client{Timeout: timeout}
			start := time.Now()
			resp, err := client.Get(urlStr.Value)
			elapsed := time.Since(start).Milliseconds()
			if err != nil {
				return &object.Hash{
					Pairs: map[string]object.Object{
						"ok":      FALSE,
						"latency": &object.Integer{Value: elapsed},
						"error":   &object.String{Value: err.Error()},
					},
				}
			}
			defer resp.Body.Close()
			return &object.Hash{
				Pairs: map[string]object.Object{
					"ok":      TRUE,
					"status":  &object.Integer{Value: int64(resp.StatusCode)},
					"latency": &object.Integer{Value: elapsed},
				},
			}
		},
	}

	builtins["file.read"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			filePath, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.read` must be STRING")
			}
			content, err := ioutil.ReadFile(filePath.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}
			return &object.String{Value: string(content)}
		},
	}

	builtins["file.write"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2+", len(args))
			}
			filePath, ok1 := args[0].(*object.String)
			if !ok1 {
				return newError("first argument to `file.write` must be STRING")
			}

			var content []byte
			switch arg := args[1].(type) {
			case *object.String:
				content = []byte(arg.Value)
			case *object.Hash, *object.Array:
				content, _ = json.MarshalIndent(baseObjectToGoType(arg), "", "  ")
			default:
				content = []byte(arg.Inspect())
			}

			err := ioutil.WriteFile(filePath.Value, content, 0644)
			if err != nil {
				return newError("could not write file: %s", err.Error())
			}
			return TRUE
		},
	}

	builtins["file.append"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			content, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `file.append` must be STRING")
			}
			f, err := os.OpenFile(path.Value, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return newError("could not open file: %s", err.Error())
			}
			defer f.Close()
			if _, err := f.WriteString(content.Value); err != nil {
				return newError("could not write to file: %s", err.Error())
			}
			return TRUE
		},
	}

	builtins["file.replace"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			path, ok1 := args[0].(*object.String)
			oldText, ok2 := args[1].(*object.String)
			newText, ok3 := args[2].(*object.String)
			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `file.replace` must be STRING")
			}
			content, err := ioutil.ReadFile(path.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}
			replaced := strings.ReplaceAll(string(content), oldText.Value, newText.Value)
			err = ioutil.WriteFile(path.Value, []byte(replaced), 0644)
			if err != nil {
				return newError("could not write file: %s", err.Error())
			}
			return TRUE
		},
	}

	builtins["file.json_update"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			update, ok2 := args[1].(*object.Hash)
			if !ok1 || !ok2 {
				return newError("arguments to `file.json_update` must be (STRING, HASH)")
			}

			content, _ := ioutil.ReadFile(path.Value)
			var data map[string]interface{}
			json.Unmarshal(content, &data)
			if data == nil {
				data = make(map[string]interface{})
			}

			updateData := baseObjectToGoType(update).(map[string]interface{})
			for k, v := range updateData {
				data[k] = v
			}

			newContent, _ := json.MarshalIndent(data, "", "  ")
			ioutil.WriteFile(path.Value, newContent, 0644)
			return TRUE
		},
	}

	builtins["file.exists"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.exists` must be STRING")
			}
			_, err := os.Stat(path.Value)
			return &object.Boolean{Value: err == nil}
		},
	}

	builtins["file.mkdir"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.mkdir` must be STRING")
			}
			os.MkdirAll(path.Value, 0755)
			return TRUE
		},
	}

	builtins["file.delete"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.delete` must be STRING")
			}
			os.RemoveAll(path.Value)
			return TRUE
		},
	}

	builtins["file.list"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `file.list` must be STRING")
			}
			files, _ := ioutil.ReadDir(path.Value)
			elements := make([]object.Object, len(files))
			for i, f := range files {
				elements[i] = &object.String{Value: f.Name()}
			}
			return &object.Array{Elements: elements}
		},
	}

	builtins["sys.exec"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1+", len(args))
			}
			cmdString, _ := args[0].(*object.String)
			var cmdArgs []string
			for _, arg := range args[1:] {
				cmdArgs = append(cmdArgs, arg.Inspect())
			}
			cmd := exec.Command(cmdString.Value, cmdArgs...)
			out, _ := cmd.CombinedOutput()
			return &object.String{Value: string(out)}
		},
	}

	builtins["sys.timestamp"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) == 0 {
				return &object.Integer{Value: time.Now().Unix()}
			}
			format, _ := args[0].(*object.String)
			goFormat := time.RFC3339
			switch format.Value {
			case "YYYY-MM-DD":
				goFormat = "2006-01-02"
			}
			return &object.String{Value: time.Now().Format(goFormat)}
		},
	}

	builtins["sys.version"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			return &object.String{Value: object.VERSION}
		},
	}

	builtins["env.get"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			name, _ := args[0].(*object.String)
			return &object.String{Value: os.Getenv(name.Value)}
		},
	}

	builtins["wait"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			var seconds float64
			switch arg := args[0].(type) {
			case *object.Integer:
				seconds = float64(arg.Value)
			case *object.Float:
				seconds = arg.Value
			}
			time.Sleep(time.Duration(seconds * float64(time.Second)))
			return NULL
		},
	}

	builtins["log"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 1 {
				return newError("wrong number of arguments. got=%d, want=1+", len(args))
			}
			msg := args[0].Inspect()
			level := "INFO"
			if len(args) > 1 {
				level = args[1].Inspect()
			}
			fmt.Printf("[%s] %s %s\n", time.Now().Format("2006-01-02 15:04:05"), level, msg)
			return NULL
		},
	}
}

func parseHeaders(h http.Header) map[string]interface{} {
	res := make(map[string]interface{})
	for k, v := range h {
		if len(v) == 1 {
			res[k] = v[0]
		} else {
			res[k] = v
		}
	}
	return res
}

func doHTTPRequest(method string, urlStr string, body object.Object, opts *object.Hash) object.Object {
	timeout := 30 * time.Second
	retries := 0
	customHeaders := map[string]string{}

	if opts != nil {
		if t, ok := opts.Pairs["timeout"]; ok {
			if tInt, ok := t.(*object.Integer); ok {
				timeout = time.Duration(tInt.Value) * time.Second
			}
		}
		if r, ok := opts.Pairs["retries"]; ok {
			if rInt, ok := r.(*object.Integer); ok {
				retries = int(rInt.Value)
			}
		}
		if h, ok := opts.Pairs["headers"]; ok {
			if hHash, ok := h.(*object.Hash); ok {
				for k, v := range hHash.Pairs {
					customHeaders[k] = v.Inspect()
				}
			}
		}
	}

	client := &http.Client{Timeout: timeout}
	var lastErr error

	for attempt := 0; attempt <= retries; attempt++ {
		var bodyReader io.Reader
		if body != nil {
			if hash, ok := body.(*object.Hash); ok {
				jsonBytes, _ := json.Marshal(baseObjectToGoType(hash))
				bodyReader = bytes.NewBuffer(jsonBytes)
			} else if str, ok := body.(*object.String); ok {
				bodyReader = bytes.NewBufferString(str.Value)
			} else {
				bodyReader = bytes.NewBufferString(body.Inspect())
			}
		}

		req, err := http.NewRequest(method, urlStr, bodyReader)
		if err != nil {
			return newError("http.%s error: %s", strings.ToLower(method), err.Error())
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		for k, v := range customHeaders {
			req.Header.Set(k, v)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			if attempt < retries {
				time.Sleep(time.Duration(attempt+1) * time.Second)
				continue
			}
			return newError("http.%s error after %d retries: %s", strings.ToLower(method), retries, err.Error())
		}
		defer resp.Body.Close()
		resBody, _ := ioutil.ReadAll(resp.Body)

		return &object.Hash{
			Pairs: map[string]object.Object{
				"status":  &object.Integer{Value: int64(resp.StatusCode)},
				"body":    &object.String{Value: string(resBody)},
				"headers": goTypeToBaseObject(parseHeaders(resp.Header)),
			},
		}
	}

	return newError("http.%s error: %s", strings.ToLower(method), lastErr.Error())
}
