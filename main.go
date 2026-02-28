package main

import (
	"base/evaluator"
	"base/lexer"
	"base/object"
	"base/parser"
	"base/repl"
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
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
		checkVersion(false)
	case "update", "--update":
		updateBase()
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
	fmt.Printf("%sHello %s!%s This is the %sB.A.S.E.%s programming language (%sv%s%s)!\n", Cyan, user.Username, Reset, Purple, Reset, Gray, object.VERSION, Reset)
	fmt.Printf("Type %sexit%s to quit. Use %sbase help%s for commands.\n", Red, Reset, Yellow, Reset)
	checkVersion(true)
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
	fmt.Printf("\n%sB.A.S.E.%s - %sBackend Automation & Scripting Environment%s (%sv%s%s)\n\n", Purple, Reset, Gray, Reset, Cyan, object.VERSION, Reset)

	fmt.Printf("%sUSAGE:%s\n", Yellow, Reset)
	fmt.Printf("  base                          Start interactive REPL\n")
	fmt.Printf("  base <script.base>            Run a script file\n")
	fmt.Printf("  base -e \"code\"                Evaluate a one-liner\n")
	fmt.Printf("  base -v, --version            Print version\n")
	fmt.Printf("  base update                   Update B.A.S.E. to latest version\n")
	fmt.Printf("  base help                     Show this help menu\n")
	fmt.Printf("  base check <file.base>        Check syntax without executing\n")
	fmt.Printf("  base new <name>               Scaffold a new project\n")
	fmt.Printf("  base run                      Run project from base.json\n")
	fmt.Printf("  base uninstall                Remove base from system\n\n")

	fmt.Printf("%sCORE MODULES:%s\n", Yellow, Reset)
	fmt.Printf("  %shttp%s     get, post, put, patch, delete, ping\n", Cyan, Reset)
	fmt.Printf("  %sdb%s       connect, query, insert, update, delete, exec, aggregate\n", Cyan, Reset)
	fmt.Printf("  %sserver%s   listen, static\n", Cyan, Reset)
	fmt.Printf("  %sfile%s     read, write, append, exists, delete, list, mkdir, replace\n", Cyan, Reset)
	fmt.Printf("  %scrypto%s   uuid, hash, encrypt_file, decrypt_file\n", Cyan, Reset)
	fmt.Printf("  %ssys%s      exec, timestamp, version\n", Cyan, Reset)
	fmt.Printf("  %smath%s     abs, sqrt, pow, round, sin, cos, log\n", Cyan, Reset)
	fmt.Printf("  %sstring%s   upper, lower, replace, slice, pad_left\n", Cyan, Reset)
	fmt.Printf("  %slist%s     map, filter, sort, contains, length\n", Cyan, Reset)
	fmt.Printf("  %sencode%s   base64\n", Cyan, Reset)
	fmt.Printf("  %sdecode%s   base64\n", Cyan, Reset)
	fmt.Printf("  %scsv%s       read\n", Cyan, Reset)
	fmt.Printf("  %syaml%s      write\n", Cyan, Reset)
	fmt.Printf("  %sssh%s       exec remote commands\n", Cyan, Reset)
	fmt.Printf("  %snotify%s    discord, email\n", Cyan, Reset)
	fmt.Printf("  %sschedule%s  recurring jobs\n", Cyan, Reset)
	fmt.Printf("  %schan%s      thread-safe channels\n\n", Cyan, Reset)

	fmt.Printf("%sUTILITIES:%s\n", Yellow, Reset)
	fmt.Printf("  log(msg, lvl?)   Wait(sec)      Type(v)  \n")
	fmt.Printf("  print(args..)    wait_all()     env.get(n)\n\n")

	fmt.Printf("%sEXAMPLES:%s\n", Yellow, Reset)
	fmt.Printf("  base script.base              Run a script\n")
	fmt.Printf("  base -e \"print(1 + 2)\"        Quick math\n\n")

	fmt.Printf("For full documentation visit: %shttps://github.com/igorkalen/base%s\n\n", Blue, Reset)
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

	fmt.Printf("\n%s✓%s Created project '%s%s%s'\n", Green, Reset, Cyan, name, Reset)
	fmt.Printf("  %sFiles:%s\n", Yellow, Reset)
	fmt.Printf("    %sbase.json%s       — project config\n", Green, Reset)
	fmt.Printf("    %smain.base%s       — entry point\n", Green, Reset)
	fmt.Printf("    %sserver.base%s     — web server\n", Green, Reset)
	fmt.Printf("    %spublic/%s         — static files (HTML/CSS/JS)\n", Green, Reset)
	fmt.Printf("\n  %sGet started:%s\n", Yellow, Reset)
	fmt.Printf("    cd %s && %sbase main.base%s\n", name, Cyan, Reset)
	fmt.Printf("    cd %s && %sbase server.base%s\n\n", name, Cyan, Reset)
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

func checkVersion(quiet bool) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/igorkalen/base/releases/latest", nil)
	req.Header.Set("User-Agent", "B.A.S.E.-CLI")

	resp, err := client.Do(req)
	if err != nil {
		if !quiet {
			fmt.Printf("%sError checking for updates: %s%s\n", Red, err, Reset)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(object.VERSION, "v")

	if latest != current {
		fmt.Printf("\n%s🚀 Update available: %s%s -> %s%s%s\n", Yellow, Reset, current, Green, latest, Reset)
		fmt.Printf("Run %sbase update%s to upgrade to the latest version.\n\n", Cyan, Reset)
	} else if !quiet {
		fmt.Printf("%sB.A.S.E. is already up to date (v%s).%s\n", Green, object.VERSION, Reset)
	}
}

func updateBase() {
	fmt.Printf("%sChecking for updates...%s\n", Blue, Reset)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/igorkalen/base/releases/latest", nil)
	req.Header.Set("User-Agent", "B.A.S.E.-CLI")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%sError checking for updates: %s%s\n", Red, err, Reset)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("%sError: GitHub API returned status %d%s\n", Red, resp.StatusCode, Reset)
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("%sError parsing update info.%s\n", Red, Reset)
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(object.VERSION, "v")

	if latest == current {
		fmt.Printf("%sB.A.S.E. is already at the latest version (v%s).%s\n", Green, object.VERSION, Reset)
		return
	}

	fmt.Printf("A new version is available: %s%s%s\n", Green, latest, Reset)
	fmt.Printf("Do you want to update? (Y/n): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(strings.ToLower(input))

	if input != "" && input != "y" && input != "yes" {
		fmt.Println("Update cancelled.")
		return
	}

	fmt.Printf("%sUpdating B.A.S.E....%s\n", Blue, Reset)
	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/igorkalen/base/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("%sUpdate failed: %s%s\n", Red, err, Reset)
		return
	}
}
