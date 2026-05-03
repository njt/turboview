---
name: writing-plans
description: Use when you have a spec or requirements for a multi-step task, before touching code
---

# Writing Plans

## Overview

Write comprehensive implementation plans assuming the engineer has zero context for our codebase. Document everything they need to know: which files to touch for each task, implementation code, and clear requirements that describe what the code must do. DRY. YAGNI. Frequent commits.

Assume they are a skilled developer, but know almost nothing about our toolset or problem domain.

**Important:** Plans describe **what to build** and **what it must do**, not **what the tests look like.** Test code is written by a dedicated test writer subagent during execution (see subagent-driven-development). The plan's job is to give the test writer clear, testable requirements and the implementer clear implementation guidance.

**Announce at start:** "I'm using the writing-plans skill to create the implementation plan."

**Context:** This should be run in a dedicated worktree (created by brainstorming skill).

**Save plans to:** `docs/superpowers/plans/YYYY-MM-DD-<feature-name>.md`
- (User preferences for plan location override this default)

## Scope Check

If the spec covers multiple independent subsystems, it should have been broken into sub-project specs during brainstorming. If it wasn't, suggest breaking this into separate plans — one per subsystem. Each plan should produce working, testable software on its own.

## Phase Sequencing

When decomposing a multi-phase project, Phase 1 must produce a **runnable application**. Each subsequent phase adds a **capability** — a coherent group of related features that together make the demo app do something meaningfully new.

Foundation types (data structures, helper functions, internal abstractions) are not their own phase. They get built as implementation details inside the phase whose capability requires them.

**Runnable application:** Builds to a binary. Launches. Draws or outputs something (even if minimal). Responds to at least one input. Exits cleanly. Not "compiles" — "you can interact with it."

**Capability:** A cluster of related features that produce a visible change when you run the demo app. If you wouldn't demo one feature without the others, they belong in the same phase. The test for whether something is a capability: can you write a new e2e test for it?

**Phase size:** A phase is too small if it's a single function or trivial feature that doesn't warrant its own e2e test. A phase is too big if it takes more than ~10 tasks. A phase should be a meaningful increment — something you could show someone and they'd see a difference in the running app.

### Anti-Pattern: Layer Phases

A phase that only builds types and data structures, with no change to what the demo app does when you run it, is a **layer** — not a capability slice. Don't do this.

Wrong: "Phase 1: geometry types, flag types, draw buffer." Nothing runs.

Right: "Phase 1: app launches, shows an empty screen, quits on Escape." The geometry types, flags, and draw buffer get built because the app needs them, not as standalone deliverables.

### Planning Question

Ask **"what should the demo app do after this phase?"** not "what types do we need?" The demo app capability defines the phase. The types are implementation details pulled in to support it.

### Worked Example: CLI Build Tool

Layer decomposition (wrong):
```
Phase 1: Argument parsing types, config file loading, output formatting
Phase 2: Subcommand registry, help text generation
Phase 3: Build logic, dependency resolution
Phase 4: Deploy logic, remote client
Phase 5: Main entrypoint wiring everything together
```

Capability decomposition (right):
```
Phase 1: `tool --version` prints version and exits
         (pulls in: main, minimal arg parsing, version constant)
Phase 2: Project scaffolding — `tool init myproject` creates dir with config
         (pulls in: init subcommand, config types, file writing)
Phase 3: Build — `tool build` compiles the project, reports errors
         (pulls in: build subcommand, config reading, dependency resolution)
Phase 4: Deploy — `tool deploy` pushes to remote with progress output
         (pulls in: deploy subcommand, remote client, progress reporting)
Phase 5: Polish — `tool help`, error messages, config validation
         (pulls in: help system, subcommand registry, validation)
```

Phase 1's app is minimal — just `--version`. But it's a real binary that runs. Each subsequent phase adds functionality you can see by running the tool. Foundation types (config structs, arg parsing) land in the phase that needs them.

## File Structure

Before defining tasks, map out which files will be created or modified and what each one is responsible for. This is where decomposition decisions get locked in.

- Design units with clear boundaries and well-defined interfaces. Each file should have one clear responsibility.
- You reason best about code you can hold in context at once, and your edits are more reliable when files are focused. Prefer smaller, focused files over large ones that do too much.
- Files that change together should live together. Split by responsibility, not by technical layer.
- In existing codebases, follow established patterns. If the codebase uses large files, don't unilaterally restructure - but if a file you're modifying has grown unwieldy, including a split in the plan is reasonable.

This structure informs the task decomposition. Each task should produce self-contained changes that make sense independently.

## Task Granularity

Each task should be a coherent unit of work: one component, one feature, or one interface. A task is too big if its requirements span multiple unrelated behaviors. A task is too small if it can't be tested independently.

Tasks contain two things:
1. **Requirements** — what the code must do, stated clearly enough that someone reading only the requirements (not the implementation) can write tests for it
2. **Implementation** — code and guidance for building it

## Plan Document Header

**Every plan MUST start with this header:**

```markdown
# [Feature Name] Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** [One sentence describing what this builds]

**Architecture:** [2-3 sentences about approach]

**Tech Stack:** [Key technologies/libraries]

---
```

## Task Structure

````markdown
### Task N: [Component Name]

**Files:**
- Create: `exact/path/to/file.py`
- Modify: `exact/path/to/existing.py:123-145`
- Test: `tests/exact/path/to/test.py`

**Requirements:**
- `function(input)` returns `expected` for valid input
- `function(None)` raises `ValueError` with message "input required"
- When called with a list, processes items in order and returns results

**Implementation:**

```python
def function(input):
    if input is None:
        raise ValueError("input required")
    return expected
```

**Run tests:** `pytest tests/path/test.py -v`

**Commit:** `git commit -m "feat: add specific feature"`
````

**Requirements must be testable.** Each requirement should describe observable behavior — what goes in, what comes out, what changes. Avoid vague requirements like "handles errors properly" — say which errors and what happens for each.

**Do NOT include test code in tasks.** Test code is written by a dedicated test writer subagent during execution. The test writer receives the requirements from the plan and the spec, and writes tests independently. Including test code in the plan defeats the information asymmetry that prevents confirmation bias.

**Do include implementation code.** The implementer needs concrete guidance — types, function signatures, data structures, algorithms. The more specific the implementation guidance, the better.

## Integration Checkpoints

After every group of tasks that build on each other, add an **Integration Checkpoint** task. This is a task whose sole purpose is to write integration tests that verify the components connect correctly.

**When to add a checkpoint:**
- After a batch of tasks that produce components which call each other
- After the first task that consumes an interface built by a previous task
- Before moving to a new subsystem that depends on the one just built

**What the checkpoint task contains:**
- Requirements describing what must work across the components just built
- Which components to wire together (real components, no mocks)
- What user-visible behaviors to verify through the real component chain
- What state flows to verify between components (e.g., focus propagation, event dispatch, data passing)

**Example checkpoint task:**

````markdown
### Task N: Integration Checkpoint — Event Dispatch Chain

**Purpose:** Verify that components from Tasks 1-4 work together.

**Requirements (for test writer):**
- A keystroke injected at the Application level reaches a Button inside a Window through the full dispatch chain
- Focus propagation works: Application → Desktop → Window → Button without manually setting focus flags
- A Button inside a real Window fires its command when activated via keyboard

**Components to wire up:** Application, Desktop, Window, Button (all real, no mocks)

**Run:** `go test ./tv/... -run TestIntegration`
````

**The integration checkpoint is not optional.** Unit tests prove components work in isolation. Integration tests prove they work together. Skipping the checkpoint means the seams between tasks are untested — this is where the bugs live.

## End-to-End Test Task

Every phase plan must include a **final task** that updates (or creates) the project's e2e test suite. The e2e test builds the actual binary and interacts with it through its real interface — not method calls in a test function, not mock event sources, not manual single-stepping through an event loop.

**What the e2e task contains:**
- Update the demo app to exercise the new functionality from this phase
- Build the binary
- Run it and interact through its real interface
- Assert on observable output

**The suite grows incrementally.** Phase 1's e2e test: app starts, responds to one input, exits cleanly. Phase 2 adds the new capability. Phase N: all previous tests still pass (regression), plus new functionality. The suite is cumulative — it never shrinks.

**Project-type guidance:**

The e2e test must interact through the same interface a real user would:

| Project Type | E2E Approach |
|-------------|-------------|
| TUI | build binary → launch in tmux → `tmux send-keys` → `tmux capture-pane` → assert screen contents |
| Web app | build → start server → HTTP requests or browser automation → assert responses |
| API/service | build → start server → real HTTP/gRPC client → assert responses |
| CLI tool | build → run as subprocess → pipe stdin / pass args → assert stdout, stderr, exit code |
| Library | build and run an example program that uses the public API → assert its output |

"Integration tests" that call methods programmatically in a test function are valuable but are **not** e2e tests. The e2e test must go through the real runtime: the real event loop, the real server, the real CLI argument parser.

**Runnable entry point.** The e2e task must produce a single entry point the phase gate can invoke — a shell script (`e2e/run.sh`), a Makefile target (`make e2e`), a test file (`e2e_test.go`), or equivalent. The phase gate needs to run the full suite without guessing what to call. Phase 1 creates this entry point; subsequent phases extend it.

## Remember
- Exact file paths always
- Complete implementation code in plan (not "add validation")
- No test code in plan — test writer writes tests during execution
- Requirements must be specific and testable (observable behavior, not vague goals)
- Exact run/commit commands
- DRY, YAGNI, frequent commits
- Integration checkpoints between related task groups
- Every phase plan ends with an e2e test task that builds and runs the actual app

## Plan Review Loop

After writing the complete plan:

1. Dispatch a single plan-document-reviewer subagent (see plan-document-reviewer-prompt.md) with precisely crafted review context — never your session history. This keeps the reviewer focused on the plan, not your thought process.
   - Provide: path to the plan document, path to spec document
2. If ❌ Issues Found: fix the issues, re-dispatch reviewer for the whole plan
3. If ✅ Approved: proceed to execution handoff

**Review loop guidance:**
- Same agent that wrote the plan fixes it (preserves context)
- If loop exceeds 3 iterations, surface to human for guidance
- Reviewers are advisory — explain disagreements if you believe feedback is incorrect

## Execution Handoff

After saving the plan, offer execution choice:

**"Plan complete and saved to `docs/superpowers/plans/<filename>.md`. Execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration. I'll pause after this phase for your go-ahead.

**2. Subagent-Driven Auto-Chain** - Same as #1, but after this phase completes I automatically plan and execute the next phase. Choose this for overnight/unattended multi-phase execution.

**3. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints

**Which approach?"**

**If Subagent-Driven chosen (option 1 or 2):**
- **REQUIRED SUB-SKILL:** Use superpowers:subagent-driven-development
- Fresh subagent per task + test writing + two-stage review
- If option 2: create execution state file with `Auto-chain: yes` and proceed through all phases without pausing

**If Inline Execution chosen:**
- **REQUIRED SUB-SKILL:** Use superpowers:executing-plans
- Batch execution with checkpoints for review
