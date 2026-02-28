package main

import (
	"base/evaluator"
	"base/lexer"
	"base/object"
	"base/parser"
	"base/repl"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	arg := os.Args[1]

	switch arg {
	case "-v", "--version":
		fmt.Printf("B.A.S.E. version %s\n", object.VERSION)
	case "-e":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base -e \"code\"")
			os.Exit(1)
		}
		evalString(os.Args[2])
	case "help":
		printHelp()
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base check <file.base>")
			os.Exit(1)
		}
		checkFile(os.Args[2])
	case "uninstall":
		uninstallBase()
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base new <project-name>")
			os.Exit(1)
		}
		scaffoldProject(os.Args[2])
	case "run":
		runFromConfig()
	default:
		if strings.HasSuffix(arg, ".base") {
			runFile(arg)
		} else {
			fmt.Printf("Unknown command: %s\n", arg)
			fmt.Println("Run 'base help' for usage information.")
			os.Exit(1)
		}
	}
}

func startREPL() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("Hello %s! This is the B.A.S.E. programming language (v%s)!\n", user.Username, object.VERSION)
	fmt.Println("Type 'exit' to quit. Use 'base help' for commands.")
	registerAllBuiltins()
	registerImportHandler()
	repl.Start(os.Stdin, os.Stdout)
}

func registerAllBuiltins() {
	evaluator.RegisterBackendBuiltins()
	evaluator.RegisterJSONBuiltins()
	evaluator.RegisterStdBuiltins()
	evaluator.RegisterListBuiltins()
	evaluator.RegisterDBBuiltins()
	evaluator.RegisterCryptoBuiltins()
	evaluator.RegisterDataBuiltins()
	evaluator.RegisterSSHBuiltins()
	evaluator.RegisterServerBuiltins()
	evaluator.RegisterSystemBuiltins()
	evaluator.RegisterNotifyBuiltins()
	evaluator.RegisterChannelBuiltins()
	evaluator.RegisterWSBuiltins()
}

func printHelp() {
	help := `
B.A.S.E. - Backend Automation & Scripting Environment (v` + object.VERSION + `)

USAGE:
  base                          Start interactive REPL
  base <script.base>            Run a script file
  base -e "code"                Evaluate a one-liner
  base -v, --version            Print version
  base help                     Show this help menu
  base check <file.base>        Check syntax without executing
  base new <name>               Scaffold a new project
  base run                      Run project from base.json
  base uninstall                Remove base from system

BUILT-IN MODULES:
  http.get(url, opts?)          GET request with optional headers/timeout/retries
  http.post(url, body, opts?)   POST request
  http.put(url, body, opts?)    PUT request
  http.patch(url, body, opts?)  PATCH request
  http.delete(url, opts?)       DELETE request
  http.ping(url, opts?)         Lightweight health check (returns ok, status, latency)

  db.connect(alias, type, dsn)  Connect to SQLite/MySQL/PostgreSQL/MongoDB
  db.query(alias, sql, args..)  Query a database
  db.insert(alias, table, data) Insert a record
  db.update(alias, tbl, w, d)   Update records
  db.delete(alias, tbl, where)  Delete records
  db.exec(alias, sql)           Execute raw SQL
  db.insert_many(a, tbl, arr)   Bulk insert
  db.aggregate(a, coll, pipe)   MongoDB aggregation

  server.listen(port, path, fn) Start HTTP server with handler
  server.static(port, dir)      Serve static files (HTML/CSS/JS)

  file.read(path)               Read file contents
  file.write(path, data)        Write to file
  file.append(path, data)       Append to file
  file.exists(path)             Check if file exists
  file.delete(path)             Delete file or directory
  file.list(path)               List directory contents
  file.mkdir(path)              Create directories
  file.replace(path, old, new)  Find and replace in file
  file.json_update(path, data)  Merge data into JSON file

  crypto.uuid()                 Generate UUID v4
  crypto.hash(data)             SHA256 hash
  crypto.encrypt_file(alg,f,k)  AES-256-GCM file encryption
  crypto.decrypt_file(alg,d,k)  AES-256-GCM decryption
  encode.base64(data)           Base64 encode
  decode.base64(data)           Base64 decode

  math.abs(n)                   Absolute value
  math.sqrt(n)                  Square root
  math.pow(base, exp)           Power
  math.round(n)                 Round to nearest integer
  math.sin(n)                   Sine
  math.cos(n)                   Cosine
  math.log(n)                   Log base 10

  string.upper(s)               Uppercase
  string.lower(s)               Lowercase
  string.replace(s, old, new)   Find and replace

  list.map(arr, fn)             Transform each element
  list.filter(arr, fn)          Filter elements
  list.sort(arr)                Sort elements
  list.contains(arr, val)       Check if contains
  list.length(arr)              Array length

  csv.read(path)                Parse CSV file
  yaml.write(path, data)        Write YAML file

  ws.connect(url, fn)           WebSocket client with message callback
  chan()                         Thread-safe channel (.send, .read_all)

  sys.exec(cmd, args..)         Execute shell command
  sys.timestamp(fmt?)           Current timestamp
  sys.version()                 Language version
  env.get(name)                 Get environment variable
  type(value)                   Get type name of a value

  ssh.exec(host, user, key, cmd) Execute remote SSH command
  notify.discord(webhook, msg)   Send Discord notification
  notify.email(host,port,u,p,..) Send email via SMTP

  schedule(cron, fn)            Schedule a recurring job
  archive.zip(src, dest)        Create ZIP archive
  wait(seconds)                 Pause execution
  wait_all()                    Wait for all spawned tasks
  log(msg, level?)              Structured logging
  print(args..)                 Print to stdout

OPTIONS HASH (for HTTP methods):
  { headers: { "Key": "Value" }, timeout: 10, retries: 3 }

RESPONSE OBJECT (for server.listen):
  res.send(status, body)        Send JSON response
  res.html(status, content)     Send HTML response
  res.header(key, value)        Set custom header
  res.file(status, filepath)    Serve a file with auto MIME type

SYNTAX BASICS:
  let age = 42                  // Variable declaration
  let tags = ["api", "db"]      // Array
  let config = {"port": 8080}   // Dictionary/JSON

  function greet(name) {        // Function
      return "Hello " + name
  }

  foreach t in tags {           // Loop
      print(t)
  }

EXAMPLES:
  base script.base              Run a script
  base -e "print(1 + 2)"        Quick math
  base check app.base           Lint your code
  base new my-api               Start a new project
`
	fmt.Println(help)
}

func checkFile(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Printf("Found %d syntax error(s) in %s:\n", len(p.Errors()), filename)
		for _, msg := range p.Errors() {
			fmt.Printf("  ✗ %s\n", msg)
		}
		os.Exit(1)
	}

	fmt.Printf("✓ %s — no syntax errors found\n", filename)
}

func uninstallBase() {
	paths := []string{"/usr/local/bin/base"}
	fmt.Println("Uninstalling B.A.S.E....")
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			err := os.Remove(p)
			if err != nil {
				fmt.Printf("Error removing %s: %s\n", p, err)
				fmt.Println("Try running with sudo: sudo base uninstall")
				os.Exit(1)
			}
			fmt.Printf("Removed %s\n", p)
		}
	}
	fmt.Println("B.A.S.E. has been uninstalled.")
}

func scaffoldProject(name string) {
	dir := name
	os.MkdirAll(dir, 0755)

	config := map[string]string{
		"name":  name,
		"entry": "main.base",
	}
	configBytes, _ := json.MarshalIndent(config, "", "  ")
	ioutil.WriteFile(filepath.Join(dir, "base.json"), configBytes, 0644)

	mainContent := `log("Hello from " + "` + name + `!");
`
	ioutil.WriteFile(filepath.Join(dir, "main.base"), []byte(mainContent), 0644)

	publicDir := filepath.Join(dir, "public")
	os.MkdirAll(publicDir, 0755)
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>` + name + `</title>
<link rel="stylesheet" href="style.css">
</head>
<body>
<h1>` + name + `</h1>
<p>Powered by B.A.S.E.</p>
<script src="app.js"></script>
</body>
</html>`
	ioutil.WriteFile(filepath.Join(publicDir, "index.html"), []byte(htmlContent), 0644)

	cssContent := `* { margin: 0; padding: 0; box-sizing: border-box; }
body { font-family: system-ui, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; background: #0a0a0a; color: #fff; }
h1 { font-size: 48px; background: linear-gradient(135deg, #667eea, #764ba2); -webkit-background-clip: text; -webkit-text-fill-color: transparent; }
p { color: #888; margin-top: 8px; }
`
	ioutil.WriteFile(filepath.Join(publicDir, "style.css"), []byte(cssContent), 0644)

	jsContent := `console.log("B.A.S.E. app loaded");
`
	ioutil.WriteFile(filepath.Join(publicDir, "app.js"), []byte(jsContent), 0644)

	serverContent := `server.static(3000, "./public");
log("Server running at http://localhost:3000");
`
	ioutil.WriteFile(filepath.Join(dir, "server.base"), []byte(serverContent), 0644)

	fmt.Printf("✓ Created project '%s'\n", name)
	fmt.Println("  Files:")
	fmt.Println("    base.json       — project config")
	fmt.Println("    main.base       — entry point")
	fmt.Println("    server.base     — web server")
	fmt.Println("    public/         — static files (HTML/CSS/JS)")
	fmt.Printf("\n  Get started:\n")
	fmt.Printf("    cd %s && base main.base\n", name)
	fmt.Printf("    cd %s && base server.base\n", name)
}

func runFromConfig() {
	content, err := ioutil.ReadFile("base.json")
	if err != nil {
		fmt.Println("No base.json found in current directory.")
		fmt.Println("Run 'base new <name>' to create a project, or create base.json manually.")
		os.Exit(1)
	}

	var config map[string]string
	if err := json.Unmarshal(content, &config); err != nil {
		fmt.Printf("Error parsing base.json: %s\n", err)
		os.Exit(1)
	}

	entry, ok := config["entry"]
	if !ok {
		fmt.Println("base.json missing 'entry' field.")
		os.Exit(1)
	}

	runFile(entry)
}

func evalString(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Woops! We ran into some B.A.S.E. parse errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	registerAllBuiltins()
	registerImportHandler()
	env := object.NewEnvironment()
	evaluated := evaluator.Eval(program, env)

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		fmt.Println(evaluated.Inspect())
		os.Exit(1)
	}

	if evaluator.KeepAlive {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
	}
}

func runFile(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Woops! We ran into some B.A.S.E. parse errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	registerAllBuiltins()
	registerImportHandler()
	env := object.NewEnvironment()
	evaluated := evaluator.Eval(program, env)

	if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
		fmt.Println(evaluated.Inspect())
		os.Exit(1)
	}

	if evaluator.KeepAlive {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
	}
}

func registerImportHandler() {
	evaluator.ImportHandler = func(path string) (object.Object, error) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, err
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			return nil, fmt.Errorf("parse errors in %s", path)
		}

		env := object.NewEnvironment()
		evaluator.Eval(program, env)

		return env.Export(), nil
	}
}
