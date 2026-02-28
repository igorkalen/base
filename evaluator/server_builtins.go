package evaluator

import (
	"base/object"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func RegisterServerBuiltins() {
	builtins["server.listen"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			port, ok1 := args[0].(*object.Integer)
			path, ok2 := args[1].(*object.String)
			fn, ok3 := args[2].(*object.Function)

			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `server.listen` must be (INTEGER, STRING, FUNCTION)")
			}

			mux := http.NewServeMux()

			handler := func(w http.ResponseWriter, r *http.Request) {
				body, _ := ioutil.ReadAll(r.Body)
				var bodyObj interface{}
				json.Unmarshal(body, &bodyObj)

				queryParams := &object.Hash{Pairs: map[string]object.Object{}}
				for k, v := range r.URL.Query() {
					if len(v) == 1 {
						queryParams.Pairs[k] = &object.String{Value: v[0]}
					} else {
						elements := make([]object.Object, len(v))
						for i, val := range v {
							elements[i] = &object.String{Value: val}
						}
						queryParams.Pairs[k] = &object.Array{Elements: elements}
					}
				}

				reqHash := &object.Hash{
					Pairs: map[string]object.Object{
						"method":  &object.String{Value: r.Method},
						"path":    &object.String{Value: r.URL.Path},
						"query":   queryParams,
						"headers": goTypeToBaseObject(parseHeaders(r.Header)),
						"body":    goTypeToBaseObject(bodyObj),
					},
				}

				customHeaders := map[string]string{}

				resSend := &object.Builtin{
					Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
						if len(innerArgs) < 2 {
							return newError("res.send needs (status, body)")
						}
						status, _ := innerArgs[0].(*object.Integer)
						for k, v := range customHeaders {
							w.Header().Set(k, v)
						}
						if w.Header().Get("Content-Type") == "" {
							w.Header().Set("Content-Type", "application/json")
						}
						w.WriteHeader(int(status.Value))
						jsonBytes, _ := json.Marshal(baseObjectToGoType(innerArgs[1]))
						w.Write(jsonBytes)
						return NULL
					},
				}

				resHtml := &object.Builtin{
					Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
						if len(innerArgs) < 2 {
							return newError("res.html needs (status, content)")
						}
						status, _ := innerArgs[0].(*object.Integer)
						for k, v := range customHeaders {
							w.Header().Set(k, v)
						}
						w.Header().Set("Content-Type", "text/html; charset=utf-8")
						w.WriteHeader(int(status.Value))
						w.Write([]byte(innerArgs[1].Inspect()))
						return NULL
					},
				}

				resHeader := &object.Builtin{
					Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
						if len(innerArgs) != 2 {
							return newError("res.header needs (key, value)")
						}
						customHeaders[innerArgs[0].Inspect()] = innerArgs[1].Inspect()
						return NULL
					},
				}

				resFile := &object.Builtin{
					Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
						if len(innerArgs) < 2 {
							return newError("res.file needs (status, filepath)")
						}
						status, _ := innerArgs[0].(*object.Integer)
						filePath, _ := innerArgs[1].(*object.String)
						content, err := ioutil.ReadFile(filePath.Value)
						if err != nil {
							w.WriteHeader(404)
							w.Write([]byte("File not found"))
							return NULL
						}
						ext := filepath.Ext(filePath.Value)
						mimeType := mime.TypeByExtension(ext)
						if mimeType == "" {
							mimeType = "application/octet-stream"
						}
						for k, v := range customHeaders {
							w.Header().Set(k, v)
						}
						w.Header().Set("Content-Type", mimeType)
						w.WriteHeader(int(status.Value))
						w.Write(content)
						return NULL
					},
				}

				resObj := &object.Hash{
					Pairs: map[string]object.Object{
						"send":   resSend,
						"html":   resHtml,
						"header": resHeader,
						"file":   resFile,
					},
				}

				applyFunction(env, fn, []object.Object{reqHash, resObj})
			}

			mux.HandleFunc(path.Value, handler)

			go func() {
				KeepAlive = true
				fmt.Printf("B.A.S.E. Server listening on :%d%s\n", port.Value, path.Value)
				http.ListenAndServe(fmt.Sprintf(":%d", port.Value), mux)
			}()

			return TRUE
		},
	}

	builtins["server.static"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			port, ok1 := args[0].(*object.Integer)
			dir, ok2 := args[1].(*object.String)

			if !ok1 || !ok2 {
				return newError("arguments to `server.static` must be (INTEGER, STRING)")
			}

			mux := http.NewServeMux()
			fs := http.FileServer(http.Dir(dir.Value))

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fullPath := filepath.Join(dir.Value, filepath.Clean(r.URL.Path))

				info, err := os.Stat(fullPath)
				if err != nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(404)
					w.Write([]byte(default404Page()))
					return
				}

				if info.IsDir() {
					indexPath := filepath.Join(fullPath, "index.html")
					if _, err := os.Stat(indexPath); err == nil {
						r.URL.Path = path.Join(r.URL.Path, "index.html")
					}
				}

				fs.ServeHTTP(w, r)
			})

			go func() {
				KeepAlive = true
				fmt.Printf("B.A.S.E. Static Server on :%d serving %s\n", port.Value, dir.Value)
				http.ListenAndServe(fmt.Sprintf(":%d", port.Value), mux)
			}()

			return TRUE
		},
	}
}

func default404Page() string {
	return `<!DOCTYPE html>
<html>
<head><title>404 - Not Found</title>
<style>
body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#0a0a0a;color:#fff}
.box{text-align:center}
h1{font-size:72px;margin:0;background:linear-gradient(135deg,#667eea,#764ba2);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
p{color:#888;font-size:18px}
a{color:#667eea;text-decoration:none}
</style>
</head>
<body><div class="box"><h1>404</h1><p>This page doesn't exist.</p><p>Powered by <a href="#">B.A.S.E.</a></p></div></body>
</html>`
}

func parseQueryString(raw string) map[string]string {
	result := map[string]string{}
	parsed, _ := url.ParseQuery(raw)
	for k, v := range parsed {
		result[k] = strings.Join(v, ",")
	}
	return result
}
