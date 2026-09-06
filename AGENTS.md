- [PLAN.md](PLAN.md) is the product scope and architectural baseline. Beads is the executable work graph and status source; do not infer current status from PLAN milestone lists.
- [README.md](README.md) is documentation for humans and agents, don't duplicate content here that is useful to write there also but link to it
- Individual module level `README.md` files in `src/` contain more detailed architectural documentation

## Starting or resuming work

1. Run `bd prime`, then use `bd ready`, `bd show <id>`, and `bd update <id> --claim`.
2. Read the issue's referenced PLAN milestone plus the README files for every module it changes. A task's blockers may establish contracts that the task relies on, so inspect their repository outputs and `bd show` records rather than assuming an implementation.
3. Inspect current APIs in the local reference checkouts before porting behavior:
   - Go behavior reference: `../gtt/README.md`
   - Jolt architecture reference: `../jtt/README.md`
   - let-go runtime, docs, and examples: `../tmp/let-go/README.md`
4. Keep stable architecture and module contracts in README documentation. Keep work state, discoveries, and follow-up tasks in Beads.
5. Do not assume APIs from an older revision, and never use live ChatGPT calls in automated tests or consume real reset credits.

# let-go

Using let-go Clojure dialect, docs:

- Repo <https://github.com/nooga/let-go> (local checkout `../tmp/let-go/README.md`)
- Examples <https://github.com/nooga/let-go/tree/main/examples>

The project pins and builds let-go from the local source checkout; do not mix an ambient `lg` binary with a newer `scripts/lg-compile`:

```bash
scripts/build-toolchain.sh
scripts/run-vm.sh health
scripts/test.sh
scripts/test-native.sh
scripts/build-aot.sh
scripts/check-m0.sh
```

A global install is optional for unrelated experiments (`go install github.com/nooga/let-go@latest`). Project commands must use `build/toolchain/lg`, whose revision is defined in `scripts/toolchain.env`.

For command-line experiments, quote let-go forms or use shell command
substitution with heredoc for multi-line expressions:

```bash
build/toolchain/lg -e '(+ 1 1)'

# multi-line
build/toolchain/lg -e "$(cat <<'EOF'
  (let [a 1]
    (+ 1 a))
EOF
)"
```

Use `brepl balance <file>` to attempt fix invalid parentheses.

## REPL-driven development

Start a persistent project nREPL with the pinned binary and keep its stdin open, then evaluate through `brepl`:

```bash
build/toolchain/lg -source-paths src:test -n -p 7888
brepl -p 7888 '(require (quote llm-proxy.core))'
brepl -p 7888 '(llm-proxy.core/health-data)'
```

let-go `require` does not accept Clojure's `:reload` option. Re-evaluate an already loaded namespace with `(load-file "src/llm_proxy/core.lg")`. Prefer iteration against the live process over unnecessary restarts.

## Reporting upstream issues

File issues with the upstream `let-go` project or its documentation using reproducible examples recorded in `docs/LET-GO-ISSUES.md`.

Document local downstream workarounds well for submitting upstream as issues. When a fix is relevant for submitting upstream, create a pull request ready branch for it from latest `let-go:main` in the `../tmp/let-go/` repository, noticing it in `docs/LET-GO-ISSUES.md` and pushing it to.

# Workflow

Use Beads, not markdown task lists, for durable work state.
Do atomic commits.

## Documenting with D2 diagrams

Architectural patterns not easily representable via ASCII should be outputted as D2 diagrams. Use D2 diagram fenced code blocks in documentation and during progress output them as temporary png's which can be rendered in Pi via `show_image` on a Kitty image protocol capable terminals:

```
d2 - - <<'EOF' | magick - /tmp/example.png
  direction: down
  x -> y
  y -> z
EOF
```

If output is simple enough to fit horizontally on half screen size terminal window, use `direction:right` image from taking too much vertical space.

<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:6cd5cc61 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See <https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md> for details and anti-patterns.

## Agent Context Profiles

The managed Beads block is task-tracking guidance, not permission to override repository, user, or orchestrator instructions.

- **Conservative (default)**: Use `bd` for task tracking. Do not run git commits, git pushes, or Dolt remote sync unless explicitly asked. At handoff, report changed files, validation, and suggested next commands.
- **Minimal**: Keep tool instruction files as pointers to `bd prime`; use the same conservative git policy unless active instructions say otherwise.
- **Team-maintainer**: Only when the repository explicitly opts in, agents may close beads, run quality gates, commit, and push as part of session close. A current "do not commit" or "do not push" instruction still wins.

## Session Completion

This protocol applies when ending a Beads implementation workflow. It is subordinate to explicit user, repository, and orchestrator instructions.

1. **File issues for remaining work** - Create beads for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **Handle git/sync by active profile**:

   ```bash
   # Conservative/minimal/default: report status and proposed commands; wait for approval.
   git status

   # Team-maintainer opt-in only, unless current instructions forbid it:
   git pull --rebase
   git push
   git status
   ```

5. **Hand off** - Summarize changes, validation, issue status, and any blocked sync/commit/push step

**Critical rules:**

- Explicit user or orchestrator instructions override this Beads block.
- Do not commit or push without clear authority from the active profile or the current user request.
- If a required sync or push is blocked, stop and report the exact command and error.
<!-- END BEADS INTEGRATION -->

<!-- BEGIN BEADS CODEX SETUP: generated by bd setup codex -->
## Beads Issue Tracker

Use Beads (`bd`) for durable task tracking in repositories that include it. Use the `beads` skill at `.agents/skills/beads/SKILL.md` (project install) or `~/.agents/skills/beads/SKILL.md` (global install) for Beads workflow guidance, then use the `bd` CLI for issue operations.

### Quick Reference

```bash
bd ready                # Find available work
bd show <id>            # View issue details
bd update <id> --claim  # Claim work
bd close <id>           # Complete work
bd prime                # Refresh Beads context
```

### Rules

- Use `bd` for all task tracking; do not create markdown TODO lists.
- Run `bd prime` when Beads context is missing or stale. Codex 0.129.0+ can load Beads context automatically through native hooks; use `/hooks` to inspect or toggle them.
- Keep persistent project memory in Beads via `bd remember`; do not create ad hoc memory files.

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.
<!-- END BEADS CODEX SETUP -->
