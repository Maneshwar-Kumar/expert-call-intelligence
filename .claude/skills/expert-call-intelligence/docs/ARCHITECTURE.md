# Architecture

```text
Root-level call transcript DOCX files
        |
        v
Deterministic discovery, hashing, and DOCX extraction
        |
        +--> one anchored Markdown transcript per call
        |
        v
Claude single-call summarizer
        |
        +--> house-style summary specification
        +--> structured call record and atomic evidence
        |
        v
Deterministic validation and Word rendering
        |
        +--> standalone summary DOCX
        +--> rebuilt consolidated master DOCX
        |
        v
Evidence ledger and semantic compile jobs
        |
        v
Claude claim compiler
        |
        +--> supporting, contradicting, and qualifying evidence
        |
        v
Compact case maps and local retrieval
        |
        v
Claude case analyst answers from a small query packet
```

## Why it is designed this way

- The source transcript remains untouched
- Hashing and stable call IDs make repeated runs safe
- Deterministic code handles files, IDs, Word output, validation, and retrieval
- Claude handles semantic work: synthesis, evidence extraction, and claim comparison
- The master is rebuilt rather than byte-appended, preventing stale duplicate sections
- Case questions retrieve a compact packet instead of absorbing the whole corpus into context

## Processing boundary

No daemon or scheduler is included. The user manually starts Claude Code from the case root and runs `/expert-calls:process-calls`. Anthropic performs model inference. The packaged helper binary runs in the same filesystem environment as Claude Code and does not make network calls
