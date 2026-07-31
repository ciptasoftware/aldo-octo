# AI Agent System Instructions (Cursor, Copilot, ChatGPT, etc.)

**CRITICAL INSTRUCTION FOR ALL LLMs AND AI AGENTS:**
If you are reading this file as part of your context window, you are assisting a human developer in building features for the `go-ai` framework. 

You **MUST** strictly adhere to the architectural rules below and in [`architecture_guidelines.md`](architecture_guidelines.md). Failure to do so will break the production environment.

---

## 🛑 THE GOLDEN RULE: CORE VS APPS

The repository is strictly divided into two zones. You are ONLY allowed to write, edit, or delete files in the **APPS ZONE**.

### ❌ CORE ZONE (DO NOT TOUCH)
All folders and files in the root directory (except `/apps`, `/docs`, and `.agents`) belong to the Core Framework. This includes:
- `main.go`
- `render/`, `router/`, `middleware/`, `database/`, `security/`, `static/`, `logger/`, `cache/`, `websocket/`
- `go.mod`, `go.sum`

**DO NOT** modify any files in the Core Zone. **DO NOT** add new imports to `main.go`. **DO NOT** suggest changes to the core architecture unless explicitly requested by the user with the exact phrase: *"I am the Lead Engineer, bypass core protection"*.

### ✅ APPS ZONE (YOUR PLAYGROUND)
All new features, domains, UI, and business logic MUST be created inside the `/apps` directory. 

When asked to "create a new feature", "add a route", or "make a UI":
1. Create a new folder in `/apps/` (e.g., `/apps/users/`).
2. Write your `models.go` (for structs and SQLite migrations) and `routes.go` (for HTTP handlers) inside that new folder.
3. If it requires HTML/CSS, put them in `/apps/ui/templates/` and `/apps/ui/public/`.

---

## 🔌 HOW TO REGISTER A NEW APP

Since you cannot touch `main.go`, how do you link your new app to the server?
You use the Bootstrap file!

**File:** `apps/bootstrap/bootstrap.go`
This is the ONLY file outside your feature folder that you should edit when adding a new feature.

1. Import your new app: `import "go-ai/apps/yourfeature"`
2. Inside the `Register` function, call your migration: `yourfeature.MigrateTable(db)`
3. Inside the `Register` function, call your router: `yourfeature.RegisterRoutes(r, tmplEngine, log, db)`

---

## 🛠️ TECHNICAL CONSTRAINTS & STACK

When writing code for this project, you must follow these constraints:
1. **100% Standard Library:** Do NOT suggest installing third-party frameworks like Gin, Echo, Fiber, or GORM. Use `net/http` (`http.ServeMux` from Go 1.22+).
2. **Database:** Use raw SQL with `database/sql`. The primary driver is SQLite (`modernc.org/sqlite`). Always use `Context` (`QueryContext`, `ExecContext`).
3. **Frontend:** Use HTMX and standard Go HTML templates (`html/template`). Avoid heavy JS frameworks (React/Vue/Svelte) unless specifically asked.
4. **Error Handling:** Log errors using the provided `logger` instance (`log.Error("msg", "error", err)`). Do NOT use `fmt.Println`. Return JSON errors using `render.JSONError`.
5. **No Global State:** Do not create global `db` or `log` variables in your app. Always pass them via dependency injection in your `RegisterRoutes` function.

---

## 🚀 TESTING & RUNNING

If you need to test the code using terminal commands:
- **DO NOT** use `go run main.go`. (It triggers Windows Firewall prompts).
- **ALWAYS** build and run the binary:
  ```powershell
  # 1. Build
  go build -ldflags="-s -w" -trimpath -o .build\win\go-ai.exe main.go
  
  # 2. Run
  .\.build\win\go-ai.exe
  ```

---

*By acknowledging these rules, you ensure the `go-ai` framework remains clean, modular, and production-grade. Now, go ahead and assist the user with their `/apps`!*
