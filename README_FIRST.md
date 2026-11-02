# Expert Call Intelligence for Claude Code

This package turns a SharePoint-synced case folder containing one raw expert-call transcript per Word file into:

- one polished house-style Word summary per call
- one continuously rebuilt consolidated Word compilation
- an atomic evidence ledger
- a registry of supporting, competing, and qualifying claims
- supplier, customer, and competitor maps
- a compact case map for low-context querying
- provenance-aware answers through Claude Code

## Install

1. Open the SharePoint case folder in File Explorer through OneDrive/SharePoint sync
2. Unzip this package **directly into that folder**, not into a new wrapper folder
3. Confirm you can see `README_FIRST.md`, `CLAUDE.md`, `.claude`, `launchers`, and `_Expert_Call_System`
4. Put one raw transcript `.docx` per call in the root of the folder
5. Start Claude Code from the root by double-clicking `launchers/START_CLAUDE_CODE.cmd`, or open a terminal in the folder and run `claude`
6. Accept workspace trust on the first launch

## Nightly operation

In Claude Code, run:

```text
/expert-calls:process-calls
```

The command processes only new or changed root-level transcripts, then rebuilds the master and case index

## Query the case

```text
/expert-calls:ask-case What competing views exist on supplier qualification time?
```

```text
/expert-calls:ask-case Why is Supplier A gaining share, and what evidence cuts against that view?
```

```text
/expert-calls:ask-case How firm is the customer order book, by funnel stage?
```

## Output locations

- Summaries: `_Expert_Call_System/01_Processed_Summaries/`
- Case pack: `_Expert_Call_System/02_Knowledge_Pack/`
- Consolidated Word file: `_Expert_Call_System/03_Compiled_Output/Expert_Call_Notes_Master.docx`
- Failures requiring attention: `_Expert_Call_System/04_Review_Queue/`

## Operational reality

No API key, local server, database service, or background scheduler is required. Claude inference runs through the authenticated Claude Code session. The packaged deterministic helper runs only while the command is active and makes no network calls

The folder must be accessible to the Claude Code session. In the common setup, this means a SharePoint/OneDrive-synced local folder. Merely uploading the ZIP to SharePoint without running Claude Code against the synced folder will not execute it

Do not run the workflow concurrently from two machines. Allow SharePoint sync to finish before and after each nightly run

## Health check

Inside the case root:

```bash
bash ".claude/skills/expert-call-intelligence/bin/expertcall.sh" self-test
```

See `.claude/skills/expert-call-intelligence/docs/TROUBLESHOOTING.md` for common issues
