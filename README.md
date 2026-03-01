# Simple Go Web Example

This is a minimal Go web application meant to demonstrate MCP workflows:

- create a GitHub repository via MCP
- push source code using `mcp_github-mcp_push_files`
- run a basic HTTP server locally

## Running locally

```bash
cd copilot_go_website/simple-go-web
# initialize modules if needed
# go mod init github.com/yourusername/simple-go-web

# optionally change the video to fetch
export VIDEO_URL="https://www.youtube.com/watch?v=IYfvmAbwRvs&t=1898s"

go run main.go
```

Then open http://localhost:8080 in your browser; the page will display the title, URL, and English transcript for the configured video.

You can update `VIDEO_URL` to point at any other YouTube video and restart the server.

## Tests

A basic test suite exercises parsing helpers and HTTP logic. Run:

```bash
go test ./...
```

The tests use local HTTP servers and do not contact external YouTube services.
