# AGENTS.md

## Conventions

- Write code and comments in English.
- Reply to the user in Brazilian Portuguese (pt-BR).
- Commit only when the user asks.
- Commits follow Conventional Commits (`feat:`, `fix:`, …) with the description in pt-BR, authored solely by the user: omit `Co-authored-by` trailers.
- Inflect each counted noun to its exact count (`1 classe`, `2 classes`); write the concrete form in place of `(s)`.

## Agent skills

### Issue tracker

Issues live as local markdown files under `.scratch/<feature>/`. See `docs/agents/issue-tracker.md`.

### Triage labels

Default five canonical labels (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`), recorded as `Status:` lines in issue files. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: root `GLOSSARY.md` + `docs/adr/`. See `docs/agents/domain.md`.
