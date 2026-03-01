# Simple Go Web Example

This is a minimal Go web application meant to demonstrate MCP workflows:

- create a GitHub repository via MCP
- push source code using `mcp_github-mcp_push_files`
- run a basic HTTP server locally

## Running locally

```bash
cd copilot_game/simple-go-web
# initialize modules if needed
# go mod init github.com/yourusername/simple-go-web

go run main.go
```

Then open http://localhost:8080 in your browser.
