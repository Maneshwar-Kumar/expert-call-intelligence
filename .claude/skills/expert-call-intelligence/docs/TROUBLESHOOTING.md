# Troubleshooting

## Commands do not appear

- Start Claude Code from the folder where this plugin was unzipped
- Accept workspace trust when prompted
- Confirm `.claude/skills/expert-call-intelligence/.claude-plugin/plugin.json` exists
- Restart Claude Code after first unzip

## A transcript is not discovered

- It must be a root-level `.docx`
- It must not start with `~$`
- Generated files under `_Expert_Call_System` are intentionally ignored

## A call is repeatedly pending

Run:

```bash
bash ".claude/skills/expert-call-intelligence/bin/expertcall.sh" status --root "."
```

Then inspect `_Expert_Call_System/_System/processing_registry.json` and the review queue

## Word summary fails to open

Run the packaged self-test:

```bash
bash ".claude/skills/expert-call-intelligence/bin/expertcall.sh" self-test
```

If it passes, reprocess the affected call. If it fails, use a binary matching the current operating system and architecture

## SharePoint sync conflict

Do not run the nightly command concurrently from two machines. Wait until SharePoint/OneDrive sync is complete before starting and before closing the session

## Index finalization stops

This is deliberate when a claim shard fails validation. Inspect the failed job and rerun `/expert-calls:compile-case`; the last valid case index remains available until replacement completes

## Corporate security blocks the packaged helper

The helper binaries are unsigned release artifacts. Some organizations use AppLocker, Defender application control, or endpoint policies that block executables launched from a synced SharePoint folder. In that case, IT must approve/hash-allow the binary, rebuild it from the included Go source, or deploy an internally signed build. Do not disable endpoint protection to work around the block
