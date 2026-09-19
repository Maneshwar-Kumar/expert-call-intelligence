# Expert Call Intelligence case

This folder contains one immutable root-level Word transcript per expert call and an installed project plugin

## Commands

- Nightly manual run: `/expert-calls:process-calls`
- Rebuild only the evidence index: `/expert-calls:compile-case`
- Ask the case: `/expert-calls:ask-case <question>`

## Permanent rules

- Never edit, rename, move, or overwrite root-level transcript DOCX files
- Put all generated content under `_Expert_Call_System/`
- Treat one root-level DOCX as one call
- Preserve exact numbers, definitions, denominators, scope, uncertainty, and provenance
- Do not resolve competing expert claims by averaging or guessing
- Do not count duplicate or derivative sources as independent corroboration
- For case questions, retrieve from the compact case pack before reading source passages
- Read original transcript passages only when source tracing is needed
- Do not use web search unless explicitly asked for external verification
