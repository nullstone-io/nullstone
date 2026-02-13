# How to test

Running `nullstone mcp-server` starts the Nullstone MCP server that is packaged in this repo.
If running using local code, you can run `go run nullstone/main.go mcp-server` to start the MCP server with the current source code.

## Inspector

There is a published MCP inspector tool that helps interact with the MCP server through a browser window.

Install the tool, then start it.
```shell
npm install -g @modelcontextprotocol/inspector
mcp-inspector go run nullstone/main.go mcp-server
```
