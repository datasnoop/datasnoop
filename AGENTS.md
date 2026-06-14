# DataSnoop Agent Guide

This repository uses OpenSpec and a Ralph-inspired working loop. All state required to continue work must live in versionable files; conversation history is not a canonical source.

## Repository language

All persisted repository content MUST be written in English. This includes documentation, OpenSpec artifacts, source code, identifiers, package and module names, APIs, database schemas, variables, functions, types, source comments, test names and descriptions, user-facing copy, error messages, logs, commit messages, configuration guidance, and progress records.

Do not introduce mixed-language identifiers into code, even when the agent-user conversation is in another language. Established external names and protocol terms remain unchanged when interoperability requires their exact spelling.

Agents may interact with users in the user's preferred language. Conversation language does not change the required language of repository output.

## Reading order

1. `docs/product/vision.md`
2. `docs/product/principles.md`
3. `docs/product/capability-map.md`
4. `docs/roadmap/current.md`
5. The active change returned by `openspec list --json`
6. Related ADRs in `docs/decisions/`
7. `docs/roadmap/opportunities.md` when evaluating scope expansion

## Ralph loop

For each iteration:

1. reread the sources above and inspect the repository's actual state;
2. select the first incomplete and unblocked item in `docs/roadmap/current.md`;
3. execute only that increment or a small, verifiable unit of it;
4. run the relevant validations;
5. update persistent state with results, evidence, decisions, and the next step;
6. stop when there is material ambiguity, destructive risk, or a need for authorization.

Do not mark an item complete without verifiable evidence. Do not use repeated iterations to expand scope.

## Canonical sources

- Vision and audience: `docs/product/vision.md`
- Decision criteria: `docs/product/principles.md`
- Product territory: `docs/product/capability-map.md`
- Current work and loop state: `docs/roadmap/current.md`
- Deliberately deferred follow-ups: `docs/roadmap/opportunities.md`
- Architectural rationale: `docs/decisions/`
- Durable behavior: `openspec/specs/`
- Executable planning: `openspec/changes/`

Notion and conversations are discovery sources. Once a decision is consolidated, capture it in the appropriate canonical document.

## Planning rules

- Name specs after durable capabilities, never phases such as D0, D1, D2, MVP, or Phase 1.
- Name changes after user-observable outcomes, not internal components.
- Record structural decisions in ADRs and normative behavior in specs.
- Record future work in `opportunities.md` with motivation, reason for deferral, dependencies, and a resumption trigger.
- OTLP is an external contract; it must not directly dictate the persistence model or interface language.
- The default experience does not require knowledge of OpenTelemetry, OTLP, or Collector.
- The Go SDK is a reference and dogfooding implementation, not a definition of the target market.
- Preserve existing changes that do not belong to the current increment.

## OpenSpec validation

Use the CLI to create changes and artifacts; never manually create a directory under `openspec/changes/`.

Before delivering planning work:

```bash
openspec status --change <change>
openspec validate <change>
```
