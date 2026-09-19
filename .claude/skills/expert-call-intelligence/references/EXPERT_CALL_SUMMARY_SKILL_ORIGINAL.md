# Expert call summary

Converts a messy, verbatim expert-call transcript into a clean, analyst-written interview summary matching the established house style, ready to be filed standalone or appended into a master expert-call notes compilation

## Required workflow

1. Read `STYLE_GUIDE.md` in full
2. Extract transcript text and identify role, company, interviewer, expert, and scope
3. Read the whole transcript before drafting
4. De-duplicate transcription artifacts and identify the clean version of each answer
5. Regroup by analytical theme rather than chronology
6. Draft stakeholder type, focus, and note where needed
7. Draft theme leads and precise attributed sub-bullets, preserving every important number and ambiguity
8. Render a Word document with real nested bullets and US Letter page size
9. Visually verify the rendered Word output
10. For a compilation, preserve stable interview order and numbering

## Common pitfalls

- Do not write Q&A
- Do not lose numbers
- Do not over-quote
- Do not silently resolve ambiguity
- Do not invent facts absent from the transcript

The packaged implementation uses a dependency-free deterministic renderer rather than the original Node script, while retaining the same JSON summary shape and Word layout conventions
