---
name: process-calls
description: Manually process every new or changed root-level expert-call DOCX, create house-style summaries, rebuild the consolidated master, and refresh the claim-based case index. Use for the nightly run.
argument-hint: "[optional: a note such as process all new calls]"
disable-model-invocation: true
---

# Process expert calls

Run the complete nightly workflow against `${CLAUDE_PROJECT_DIR}`. The user has approved direct Anthropic processing. This workflow does not require an Anthropic API key; it runs through the user's authenticated Claude Code session.

## Non-negotiable boundaries

- Treat root-level `.docx` files as immutable source transcripts
- Process one document as one call
- Ignore temporary files beginning with `~$`
- Put every generated artifact under `_Expert_Call_System/`
- Never count the same source hash twice
- Preserve numbers, ambiguity, caveats, source anchors, and competing claims
- Do not load all raw transcripts into the parent context
- Do not use web search unless the user explicitly asks for external verification

## Deterministic runner

Use this command form for every utility call:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" <command> --root "${CLAUDE_PROJECT_DIR}" ...
```

## Phase 1: initialize and discover

1. Run `init`
2. Run `discover` and save its JSON output to:
   `_Expert_Call_System/_System/logs/latest_discovery.json`
3. Read that JSON
4. Process only entries in `pending`
5. Report but do not process entries in `duplicates` or `missing`

A call is pending when it is new, changed, extracted but incomplete, rendered but incomplete, or previously failed. A changed source keeps its stable call ID and is reprocessed

## Phase 2: process each pending call sequentially

For each pending call:

1. Run:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" extract \
  --root "${CLAUDE_PROJECT_DIR}" \
  --call-id "<CALL_ID>"
```

2. Read the extraction JSON returned by the command. It gives the exact values for:
   - `source_file`
   - `source_sha256`
   - `extracted_path`
   - `summary_spec_path`
   - `call_record_path`

3. Delegate exactly one call to the `expert-calls:call-summarizer` subagent. Pass all six values above plus `PROJECT_ROOT=${CLAUDE_PROJECT_DIR}`. Do not paste the full transcript into the parent conversation

4. Validate the two JSON outputs:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" validate \
  --root "${CLAUDE_PROJECT_DIR}" \
  --call-id "<CALL_ID>" \
  --summary "<SUMMARY_SPEC_PATH>" \
  --record "<CALL_RECORD_PATH>"
```

5. If validation fails, give the same subagent one focused repair attempt containing only the validation errors and the file paths. Validate again

6. Render the standalone Word summary:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" render-summary \
  --root "${CLAUDE_PROJECT_DIR}" \
  --call-id "<CALL_ID>"
```

7. Mark the call complete:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" mark-complete \
  --root "${CLAUDE_PROJECT_DIR}" \
  --call-id "<CALL_ID>"
```

8. If extraction, summarization, repair, validation, or rendering still fails, run `mark-failed` with a precise error message, continue to the next call, and do not invent a summary

Do not parallelize call processing. Sequential processing avoids registry and output conflicts

## Phase 3: rebuild the compilation and case index

After all pending calls have been attempted:

1. Run `render-master`. This rebuilds the master from all complete summaries in stable first-seen order. It is append-like to the user, but avoids duplicate or stale sections when a source changes
2. Follow the full workflow in the `compile-case` skill, beginning with `build-ledger`
3. Run `status`

## Completion response

State:

- Number processed successfully
- Number unchanged
- Number failed and their call IDs
- Duplicate and missing-source warnings
- Paths to the consolidated master, `CASE_MAP.md`, and review queue
- Whether the claim index completed fully

Do not reproduce the summaries in chat. The Word files and case pack are the deliverables
