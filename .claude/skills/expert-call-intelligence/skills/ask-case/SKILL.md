---
name: ask-case
description: Answer a supplier, customer, competitor, market, or evidence-tracing question from the compact case index without loading the full expert-call corpus.
argument-hint: "<question about the case>"
---

# Ask the expert-call case

Question: `$ARGUMENTS`

Use `${CLAUDE_PROJECT_DIR}` as the case root

## Retrieval-first protocol

1. Confirm `_Expert_Call_System/02_Knowledge_Pack/CASE_MAP.md` exists. If it does not, tell the user to run `/expert-calls:process-calls` or `/expert-calls:compile-case`
2. Read `CASE_MAP.md`, but do not scan all transcripts or all summaries
3. Use the Write tool to save the exact question, with no extra text, to `_Expert_Call_System/_System/query_packets/latest_question.txt`
4. Generate a compact query packet:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" search \
  --root "${CLAUDE_PROJECT_DIR}" \
  --query-file "_Expert_Call_System/_System/query_packets/latest_question.txt" \
  --out "_Expert_Call_System/_System/query_packets/latest.md"
```

5. Delegate the question to the `expert-calls:case-analyst` subagent. Pass the question, project root, query-packet path, and runner path
6. Return the subagent's answer in chat

## Answer quality standard

The answer must:

- Lead with the strongest evidence-backed conclusion
- Distinguish consensus, majority view, minority view, true contradiction, and mere scope differences
- Separate expert evidence from analyst inference
- Cite `CLM-...`, `EC-...`, and `EC-...-E...` identifiers inline
- State the relevant period, geography, product, customer, scenario, and denominator when material
- State confidence and evidence gaps
- Use exact source tracing only when needed

Never claim that absence from the retrieved packet proves absence from the corpus. Say that the indexed evidence did not surface it
