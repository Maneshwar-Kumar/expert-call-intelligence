# Expert Call Intelligence plugin

A project-scoped Claude Code plugin for processing one expert-call transcript per root-level Word document

## Packaged skills

- `/expert-calls:process-calls` — process new or changed calls, render house-style summaries, rebuild the master compilation, and refresh the case index
- `/expert-calls:compile-case` — rebuild the evidence and claim layer without re-summarizing unchanged calls
- `/expert-calls:ask-case <question>` — answer from a compact retrieval packet with competing evidence and source traceability

## Packaged agents

- `expert-calls:call-summarizer` — one-call summary and evidence extraction
- `expert-calls:claim-compiler` — cross-call semantic claim clustering
- `expert-calls:case-analyst` — retrieval-first case questioning

## Runtime model

Claude performs semantic work through the authenticated Claude Code session. A dependency-free, cross-platform helper handles source hashing, DOCX extraction, schema checks, Word rendering, registries, index generation, local search, and source tracing

All case state is written under `_Expert_Call_System/` in the project root. The plugin directory itself is treated as read-only
