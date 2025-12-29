# Go Debugging Guide for ActiveLog

## Table of Contents
1. [Prerequisites](#prerequisites)
2. [VS Code Setup](#vs-code-setup)
3. [Using the Debugger](#using-the-debugger)
4. [Common Debugging Scenarios](#common-debugging-scenarios)
5. [Troubleshooting](#troubleshooting)
6. [Best Practices](#best-practices)

---

## Prerequisites

### 1. Install Delve (Go Debugger)

Delve is the standard debugger for Go. Install it globally:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Verify installation:

```bash
dlv version
```

### 2. Install VS Code Go Extension

1. Open VS Code
2. Press `Cmd+Shift+X` (Mac) or `Ctrl+Shift+X` (Windows/Linux)
3. Search for "Go" (by Go Team at Google)
4. Click Install

### 3. Install Go Tools

When you first open a Go file, VS Code will prompt you to install additional Go tools. Click "Install All" or run:

```
Cmd+Shift+P → Go: Install/Update Tools → Select All → OK
```

---

## VS Code Setup

The `.vscode` folder contains two configuration files:

### `launch.json` - Debug Configurations

Six pre-configured debugging scenarios:

1. **Launch API Server** - Debug your main API application
2. **Debug Current File** - Debug the currently open Go file
3. **Debug Current Test** - Debug a specific test function
4. **Debug All Tests** - Debug all tests in the workspace
5. **Debug Package Tests** - Debug all tests in the current package
6. **Attach to Process** - Attach debugger to a running Go process

### `settings.json` - Go Development Settings

Configured for:
- Auto-formatting on save
- Automatic import organization
- Linting with golangci-lint
- Race detection in tests
- Optimized IntelliSense

---

## Using the Debugger

### Method 1: Using Debug Configurations (Recommended)

1. **Set Breakpoints**
   - Click in the gutter (left of line numbers) to add a red dot
   - Or press `F9` on the desired line

2. **Start Debugging**
   - Press `F5` or click the green play button in the Debug panel
   - Select a configuration from the dropdown:
     - Choose "Launch API Server" for main application
     - Choose "Debug Current Test" when in a test file

3. **Debug Controls**
   - `F5` - Continue
   - `F10` - Step Over
   - `F11` - Step Into
   - `Shift+F11` - Step Out
   - `Shift+F5` - Stop

### Method 2: Quick Debug

**Debug Main Application:**
```bash
# From terminal
dlv debug ./cmd/api
```

**Debug Specific Test:**
```bash
# Run Delve for a specific test
dlv test ./internal/handlers -- -test.run TestCreateActivity
```

### Method 3: Debug Test Functions

When in a test file, VS Code shows "debug test" above each test function:

```go
func TestCreateUser(t *testing.T) {  // <- Click "debug test" link above
    // test code
}
```

---

## Common Debugging Scenarios

### 1. Debug the API Server

```go
// cmd/api/main.go
func main() {
    // Set breakpoint here
    router := setupRouter()

    // Or here to debug server startup
    log.Fatal(http.ListenAndServe(":8080", router))
}
```

**Steps:**
1. Set breakpoint in `main()` or handler function
2. Press `F5` → Select "Launch API Server"
3. Use Postman/curl to trigger the endpoint
4. Debugger will pause at breakpoint

### 2. Debug HTTP Handlers

```go
// internal/handlers/activity.go
func (h *ActivityHandler) CreateActivity(w http.ResponseWriter, r *http.Request) {
    var req CreateActivityRequest

    // Set breakpoint here to inspect request
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        // Or here to debug error handling
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    // Breakpoint here to see decoded data
    activity, err := h.service.Create(r.Context(), req)
    // ...
}
```

### 3. Debug Tests

```go
// internal/handlers/activity_test.go
func TestCreateActivity(t *testing.T) {
    // Set breakpoint here
    handler := NewActivityHandler(mockService)

    req := httptest.NewRequest("POST", "/activities", body)
    rr := httptest.NewRecorder()

    // Or here to debug handler execution
    handler.CreateActivity(rr, req)

    // Or here to inspect response
    assert.Equal(t, http.StatusCreated, rr.Code)
}
```

**Steps:**
1. Open test file
2. Set breakpoint
3. Click "debug test" link above function OR press `F5` → "Debug Current Test"

### 4. Debug Database Queries

```go
// internal/repository/activity.go
func (r *ActivityRepository) Create(ctx context.Context, activity *Activity) error {
    query := `INSERT INTO activities (user_id, type, duration) VALUES ($1, $2, $3) RETURNING id`

    // Set breakpoint to inspect query and params
    err := r.db.QueryRowContext(ctx, query, activity.UserID, activity.Type, activity.Duration).Scan(&activity.ID)

    // Breakpoint here to see error or success
    return err
}
```

### 5. Debug Concurrent Code (Goroutines)

```go
func ProcessActivities(activities []Activity) {
    var wg sync.WaitGroup

    for _, activity := range activities {
        wg.Add(1)
        go func(a Activity) {
            defer wg.Done()

            // Set breakpoint here
            // Note: With multiple goroutines, debugger will pause each one
            processActivity(a)
        }(activity)
    }

    wg.Wait()
}
```

**Tips for debugging goroutines:**
- Set breakpoints in goroutine functions
- Use the Debug Console to inspect goroutine states
- Check the Call Stack panel to see which goroutine you're in

---

## Debugging Tips & Tricks

### Inspect Variables

- **Hover** over variables to see values
- **Variables Panel** (left sidebar) shows all local variables
- **Watch Panel** - Add expressions to monitor: `len(users)`, `user.Email`, etc.
- **Debug Console** - Evaluate expressions: type `user.ID` and press Enter

### Conditional Breakpoints

Right-click on a breakpoint → Edit Breakpoint → Add condition:
```
user.ID == 123
len(activities) > 0
err != nil
```

### Logpoints (Print without stopping)

Right-click in gutter → Add Logpoint:
```
User ID: {user.ID}, Email: {user.Email}
```

### Debug Console Commands

```go
// Evaluate expressions
user.Email
len(activities)
activity.Duration.Minutes()

// Call functions
fmt.Println(user)
```

---

## Troubleshooting

### Issue: "could not launch process: decoding dwarf section info"

**Solution:** Rebuild with debug symbols:
```bash
go build -gcflags="all=-N -l" ./cmd/api
```

### Issue: Breakpoints not hitting

**Possible causes:**
1. Code not executed yet (add logging to verify)
2. Optimized build (use debug build flags)
3. Wrong configuration selected
4. Stale binary (rebuild: `go build`)

**Solution:**
```bash
# Clean and rebuild
go clean -cache
go build ./cmd/api
```

### Issue: "Failed to launch: could not find dlv"

**Solution:**
```bash
# Ensure Delve is in PATH
which dlv

# If not found, install:
go install github.com/go-delve/delve/cmd/dlv@latest

# Add Go bin to PATH in ~/.zshrc or ~/.bashrc:
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: Variables show "<optimized out>"

**Solution:** Disable optimizations:
```json
// In launch.json, add to configuration:
"buildFlags": "-gcflags='all=-N -l'"
```

### Issue: Debugger slow or timing out

**Solution:** Increase timeout in `.vscode/settings.json`:
```json
"go.delveConfig": {
    "dlvLoadConfig": {
        "maxStringLen": 50,      // Reduce from 120
        "maxArrayValues": 20     // Reduce from 64
    }
}
```

---

## Best Practices

### 1. Strategic Breakpoint Placement

✅ **Good:**
```go
func CreateUser(req *CreateUserRequest) (*User, error) {
    // Breakpoint at function entry to see inputs
    if err := req.Validate(); err != nil {
        // Breakpoint in error path
        return nil, err
    }

    user := &User{Email: req.Email}
    // Breakpoint before DB operation
    if err := db.Create(user); err != nil {
        return nil, err
    }

    // Breakpoint to verify successful creation
    return user, nil
}
```

❌ **Avoid:**
- Breakpoints in tight loops (use conditional breakpoints instead)
- Breakpoints in frequently-called functions without conditions

### 2. Use Logging + Debugging Together

```go
func ProcessActivity(a *Activity) error {
    log.Printf("Processing activity: %+v", a)  // Log for production

    // Breakpoint here for debugging
    if err := validate(a); err != nil {
        log.Printf("Validation failed: %v", err)
        return err
    }

    return nil
}
```

### 3. Test-Driven Debugging

1. Write a failing test
2. Debug the test to understand why it fails
3. Fix the code
4. Verify test passes

### 4. Debug Configuration Templates

Create task-specific configurations:

```json
{
    "name": "Debug with Production DB",
    "type": "go",
    "request": "launch",
    "mode": "debug",
    "program": "${workspaceFolder}/cmd/api",
    "env": {
        "DB_HOST": "prod-db.example.com",
        "ENV": "staging"
    }
}
```

---

## Quick Reference

### Keyboard Shortcuts (Mac)

| Action | Shortcut |
|--------|----------|
| Toggle Breakpoint | `F9` |
| Start Debugging | `F5` |
| Stop Debugging | `Shift+F5` |
| Step Over | `F10` |
| Step Into | `F11` |
| Step Out | `Shift+F11` |
| Continue | `F5` |
| Restart | `Cmd+Shift+F5` |
| Show Debug Console | `Cmd+Shift+Y` |

### Common Debug Console Commands

```go
// Print variable
p user

// Print formatted
p fmt.Sprintf("%+v", user)

// Check type
p reflect.TypeOf(user)

// Length of slice
p len(activities)

// Call method
p user.IsActive()
```

---

## Next Steps

1. **Practice:** Set breakpoints in your existing code and step through execution
2. **Watch:** Observe how data flows through your application
3. **Experiment:** Try conditional breakpoints and logpoints
4. **Profile:** Learn to use Go's profiling tools alongside debugging

---

## Resources

- [Delve Documentation](https://github.com/go-delve/delve/tree/master/Documentation)
- [VS Code Go Debugging](https://github.com/golang/vscode-go/wiki/debugging)
- [Go Debugging Best Practices](https://go.dev/doc/diagnostics)

---

**Remember:** Debugging is a skill. The more you practice, the faster you'll identify and fix issues. Happy debugging!
