# Data dictionary

## Stable identifiers

- `EC-0001`: one registered source call
- `EC-0001-E001`: one atomic evidence record from that call
- `JOB-...`: one semantic compile group derived from module, topic, and primary entity
- `CLM-...`: one normalized claim cluster

## Registries

- `processing_registry.json`: source hashes, stable sequence, status, and generated paths
- `source_registry.json`: call metadata and provenance paths
- `evidence_ledger.jsonl`: one source-specific assertion per line
- `claim_registry.jsonl`: one normalized synthesized claim per line
- `entity_registry.json`: canonical entities, aliases, roles, and coverage
- `topic_index.json`: compact coverage counts

## Maps

- `CASE_MAP.md`: compact routing layer and contested areas
- `SUPPLIER_MAP.md`: supplier topics and claims
- `CUSTOMER_MAP.md`: customer topics and claims
- `COMPETITOR_MAP.md`: competitor topics and claims

## Source hierarchy

- Extracted transcript anchors are the highest-fidelity source available to the plugin
- Call summaries are human-readable derivatives
- Evidence records are structured source assertions
- Claim clusters are analyst synthesis across records
- Maps are routing indexes, not evidence
