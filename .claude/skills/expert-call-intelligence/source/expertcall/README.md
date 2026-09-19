# Helper source

The deterministic helper is written in Go and uses only the standard library. It performs local DOCX extraction and rendering, SHA-256 registry management, JSON validation, case-pack assembly, lexical retrieval, and source tracing. It makes no network calls

Build example:

```bash
go build -trimpath -ldflags "-s -w" -o expertcall .
```
