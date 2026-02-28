# B.A.S.E.

B.A.S.E. (Backend Automation & Scripting Environment) is a simple, standalone programming language for writing backend logic and server scripts.

It's built in Go and compiles to a single 22MB binary. There are no dependencies to install, no `npm`, and no `pip`. Just download the binary and you can run HTTP servers, query databases, send Discord webhooks, and automate your machine.

## Installation

Run this to download the binary and install it to `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/igorkalen/base/main/install.sh | bash
```

Check if it works:
```bash
base --version
```

## How to use it

Start the interactive REPL to test things out:
```bash
base
```

Run a file:
```bash
base my_script.base
```

Create a new web server project (creates a folder with a `server.base` file):
```bash
base new my-api
cd my-api
base run
```

Run a quick one-liner:
```bash
base -e 'print("Hello world")'
```

---

## What can it do?

B.A.S.E. comes with a big standard library. Everything below works out of the box.

### Web Server
You can spin up an HTTP server or serve static HTML files easily.

```base
// Serve a folder of HTML/CSS files on port 8080
server.static(8080, "./public")

// Create a custom API route
server.listen(3000, "/api", function(req, res) {
    if req.method == "GET" {
        res.send(200, {"message": "Server is running"})
    }
})

// CRITICAL: Servers and cron jobs run natively in the background. 
// You must keep the main script alive so it doesn't immediately exit!
wait(999999)
```

### HTTP Client
Send requests to other APIs. It supports custom headers, timeouts, and retries.

```base
let data = http.get("https://api.example.com/data", {
    "headers": {"Authorization": "Bearer token123"},
    "timeout": 5,
    "retries": 3
})
```

### Databases
Connect to SQLite, MySQL, PostgreSQL, or MongoDB natively.

```base
db.connect("main", "sqlite", "./data.db")
db.exec("main", "CREATE TABLE users (name TEXT)")
db.insert("main", "users", {"name": "Igor"})

let users = db.query("main", "SELECT * FROM users")
```

### Notifications (Discord & Email)
Send messages directly to a Discord channel or via an SMTP email server.

```base
// Send a message to a Discord webhook
notify.discord("https://discord.com/api/webhooks/...", "Server deployment finished!")

// Send an email (host, port, username, password, to_email, subject, body)
notify.email("smtp.gmail.com", "587", "you@gmail.com", "password", "target@email.com", "Alert", "Server is down!")
```

### Background Tasks (Concurrency)
Spawn tasks to run in the background. `chan()` creates a thread-safe channel to collect the results.

```base
let results = chan()

spawn function() {
    let ping = http.ping("https://google.com")
    results.send(ping)
}()

wait_all()
print(results.read_all())
```

### Files & Encryption
Read/write files, generate UUIDs, and encrypt files using AES-256-GCM.

```base
let id = crypto.uuid()

// Encrypt a file
crypto.encrypt_file("aes-256-gcm", "secrets.json", "my-password")

// Read a file
let content = file.read("config.json")
```

### Cron Jobs
Run functions automatically on a schedule.

```base
schedule("0 * * * *", function() {
    log("Running hourly task", "INFO")
})

wait(999999) // Keep the script alive
```

---

## Syntax Basics

The syntax looks like JavaScript or C, but uses `let` for variables.

```base
// Variables
let username = "Igor"
let active = true
let config = {"debug": true, "retries": 3}
let tags = ["dev", "admin"]

// Functions
function calculate(x, y) {
    return x + y
}

// If statements
if active {
    print("Welcome " + username)
}

// Loops
foreach t in tags {
    print(t)
}

for let i = 0; i < 5; i = i + 1 {
    print(i)
}
```

## Full list of built-ins
To see everything the language can do, type `base help` in your terminal. It lists all modules for `http`, `db`, `file`, `math`, `list`, `crypto`, `string`, `csv`, `yaml`, `ws`, `ssh`, and more.

## Contributing

1. Clone the repo: `git clone https://github.com/igorkalen/base.git`
2. Install dependencies: `go mod download`
3. Make changes (built-ins are in `/evaluator`, language rules are in `/parser` and `/lexer`)
4. Test it locally: `go run main.go tests/my_test.base`
5. Open a Pull Request!
