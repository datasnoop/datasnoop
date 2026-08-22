# Git Workflow

DataSnoop uses a simple mainline workflow. `main` is the integration branch and
all changes reach it through short-lived branches and pull requests.

## Branches

Use one of these prefixes followed by a concise, kebab-case outcome:

- `feat/` for user-visible capabilities;
- `fix/` for defect corrections;
- `chore/` for maintenance;
- `docs/` for documentation-only changes;
- `refactor/` for behavior-preserving structural changes;
- `test/` for test-only changes;
- `ci/` for automation changes; and
- `codex/` for branches created by Codex.

Do not create `develop`, `release/*`, or `hotfix/*` branches. Urgent fixes use
a short-lived `fix/` branch and follow the same pull-request checks.

## Commits

Use Conventional Commits:

```text
<type>(<optional scope>): <imperative summary>
```

Examples:

```text
feat(api): add authenticated OTLP receiver
fix(persistence): preserve transaction rollback outcome
docs(workflow): document branch conventions
ci: verify database migrations in pull requests
```

Use a type that describes the intent, keep the summary concise, and write all
commit messages in English.

## Pull requests

Open a pull request from a short-lived branch to `main`. Before requesting
review, run the relevant checks locally. A pull request may merge only after its
required CI checks pass and its review policy is satisfied.
