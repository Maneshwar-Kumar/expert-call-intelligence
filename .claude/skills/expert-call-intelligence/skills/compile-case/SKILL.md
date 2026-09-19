---
name: compile-case
description: Rebuild the evidence ledger, competing-claim registry, supplier/customer/competitor maps, entity cards, source cards, and compact case map from all completed expert-call records.
argument-hint: "[optional: rebuild the case index]"
disable-model-invocation: true
---

# Compile the case evidence system

Use `${CLAUDE_PROJECT_DIR}` as the case root and this deterministic command form:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" <command> --root "${CLAUDE_PROJECT_DIR}" ...
```

## 1. Build the evidence ledger and semantic jobs

Run:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" build-ledger \
  --root "${CLAUDE_PROJECT_DIR}"
```

Then run `discover-jobs` and save its JSON to `_Expert_Call_System/_System/logs/latest_compile_jobs.json`

## 2. Compile pending claim groups

For each pending job, sequentially:

1. Delegate to the `expert-calls:claim-compiler` subagent
2. Pass:
   - `PROJECT_ROOT=${CLAUDE_PROJECT_DIR}`
   - the job's `job_path`
   - the job's `output_path`
   - the job's `job_id`
3. Validate:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" validate-shard \
  --root "${CLAUDE_PROJECT_DIR}" \
  --job "<JOB_PATH>" \
  --shard "<OUTPUT_PATH>"
```

4. On failure, give the agent one focused repair attempt with the validation errors, then validate again
5. If a shard remains invalid, stop finalization and report the failed job. Do not silently build an incomplete index

Do not parallelize shard writes

## 3. Finalize the compact case pack

When all shards validate, run:

```bash
bash "${CLAUDE_PLUGIN_ROOT}/bin/expertcall.sh" finalize-index \
  --root "${CLAUDE_PROJECT_DIR}"
```

This generates or refreshes:

- `CASE_MAP.md`
- `SUPPLIER_MAP.md`
- `CUSTOMER_MAP.md`
- `COMPETITOR_MAP.md`
- `claim_registry.jsonl`
- `evidence_ledger.jsonl`
- `entity_registry.json`
- `source_registry.json`
- `topic_index.json`
- `OPEN_QUESTIONS.md`
- entity and source cards

Then run `status`

## Compilation principles

- Evidence records are source-specific and immutable until their underlying transcript changes
- Claim clusters are synthesized views, not replacements for evidence
- A different period, geography, product, customer, scenario, or denominator is usually a scope difference, not automatically a contradiction
- Genuine disagreements remain visible as supporting, contradicting, and qualifying evidence
- No source gets extra weight merely because its summary is longer
- A derivative or duplicate source is not independent corroboration
- The case map must remain compact and act as a routing index, not a prose dump
