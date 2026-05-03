# Test Quality Gold Standard

**Load this reference when:** writing tests from a spec, or reviewing tests before implementation begins.

This document is pasted into test writer and test reviewer subagent prompts. Every rule must be actionable by an LLM agent working from a spec in a specific programming language. If a rule can't be checked mechanically against the spec and the test code, it doesn't belong here.

---

## The Core Rule

**Every assertion must verify something the spec says.** If you cannot point to the sentence in the spec that justifies an assertion, do not write it.

This prevents the most dangerous failure mode: the agent writes an assertion that matches what the code *does* rather than what the code *should do*. Example: the spec says "button fires a command event." The implementation calls `event.Clear()` (setting `What = EvNothing`). The test asserts `event.IsCleared()`. The test passes. The button is broken. The assertion came from the code, not the spec.

An assertion derived from the implementation is called an *implementation-derived oracle.* It is always wrong, even when it passes, because it encodes the implementation's bugs as correct behavior.

---

## Hard Rules

These always apply. No exceptions.

### 1. Assertions trace to spec

For each assertion, you must be able to answer: "which sentence in the spec does this verify?" If the answer is "it just seems right" or "that's what the code does," delete the assertion and start over.

### 2. Test the spec's API, not workarounds

Tests must call the same functions a user would call, as described in the spec. If the spec says `MessageBox()` opens a modal dialog, the test calls `MessageBox()`. If `MessageBox()` doesn't work, the test fails — it does not substitute a plain `Window` inserted manually.

**Check:** if you removed all workarounds from your test, would it still pass? If not, you've routed around a bug. Report BLOCKED instead of shipping the workaround.

### 3. Don't manually set framework-provided state

If you manually set state that the framework normally provides (focus flags, owner references, visibility, event routing), your test proves the component works with hand-crafted state, not that the framework delivers that state correctly. You need both:
- A unit test that sets state directly (proves the component's logic)
- An integration test where the framework sets state naturally (proves the connection)

The unit test alone is insufficient. The integration test is where the real bugs live.

### 4. Write falsifying tests

For each happy-path test, also write a test that catches the most plausible shortcut. Ask: "What's the laziest implementation that would pass my happy-path test?"

- Return the input unchanged → test with different input/output
- Only handle the first item → test with multiple items
- Hardcode the expected value → test with a second set of values
- Skip the hard part → test the behavior that requires the hard part
- Treat zero/nil/empty as valid → test that zero/nil/empty is rejected

If you can't think of a shortcut, the requirement may be too vague — flag it.

### 5. No test weakening

When a test fails, exactly three responses are valid:
1. **Fix the code** — the test caught a real bug
2. **Move the test** — it belongs at a different scope
3. **Delete the test** — the requirement changed and the test is obsolete

Never weaken an assertion to match unexpected behavior. Never add a workaround so the test passes without testing the real API. Never change a test because "it's too strict" without a spec-based justification. Changing a test to match the code is confirmation bias by definition.

### 6. One behavior per test

Each test verifies one thing. If the test name has "and" in it, split it. If a test fails, the developer should know exactly which behavior is broken from the test name alone.

### 7. Tests are independent and deterministic

No order dependency between tests. No shared mutable state. Each test creates its own fixtures. No `sleep()` — if you need to wait, wait for a condition with a timeout that fails with a diagnostic message. Seed RNGs. Inject clocks. Stub external systems you don't own.

---

## Techniques

Use these when they apply. Not every test needs all of them.

### Boundary and edge-case testing

For any quantity or value in a test, consider: zero, one, typical, maximum, maximum+1. For collections: empty, single item, many items. For strings: empty, whitespace, very long. Don't test all of these for every value — focus on values where the spec implies a boundary or where a wrong implementation would handle the edge differently.

### Context questioning

For each requirement, ask: "Is there another context where this event produces a different outcome?" This generates edge cases. Example: the spec says "pressing Enter activates the focused button." Context question: "What if no button is focused? What if the focused widget is a text input, not a button?"

### Outcome questioning

For each requirement, ask: "Is there another important outcome besides the one I'm testing?" Example: testing that a dialog returns the correct command — but also, does the dialog close? Does focus return to the previous window?

### Integration tests with real dependencies

When your component connects to a component from a previous task, write at least one test using the *real* dependency, not a mock. If the dependency doesn't exist yet, mark the test as pending with a comment naming the dependency.

---

## Anti-Patterns to Catch

### Tests that validate bugs as correct

The implementation does X. The test asserts X. But the spec says Y. This is the most dangerous anti-pattern because the test passes, the code review finds nothing, and the bug ships.

**How to catch it:** for each assertion, re-read the spec requirement it claims to verify. Does the assertion test what the spec says, or what the code does?

### Tests that route around broken behavior

The spec's API doesn't work, so the test uses a different approach that happens to exercise similar behavior. The test passes, but the spec's API is completely untested.

**How to catch it:** does the test call the same functions described in the spec? If the test builds a custom handler, inserts a plain component, or uses a helper that bypasses the standard API — that's a workaround, not a test.

### Tests with incidental assertions

Assertions that verify implementation details rather than spec requirements. "The internal list has 3 elements" when the spec only says "the user sees 3 items." These break on refactoring and don't verify user-visible behavior.

**How to catch it:** would a different correct implementation fail this assertion? If yes, the assertion is too coupled to implementation details.

---

## Test Levels

Three levels of testing serve different purposes:

| Level | What It Tests | How |
|-------|--------------|-----|
| Unit | One component in isolation | Call functions directly, mock external dependencies |
| Integration | Multiple real components wired together | Call functions directly, use real dependencies (no mocks for things you built) |
| End-to-end | The actual application as a user experiences it | Build the binary, run it, interact through its real interface (tmux, HTTP, subprocess) |

Unit and integration tests are well-handled by the test writer workflow. The gap is e2e: tests that go through the real runtime — the real event loop, the real server, the real CLI parser. An "integration test" that calls methods programmatically in a test function is not an e2e test, even if it uses real components.

E2e tests are specified in the phase plan's final task and grow incrementally with each phase.

---

## Writer's Checklist

Before submitting tests:

- [ ] Every assertion verifies a specific spec requirement (can you quote it?)
- [ ] Falsifying test exists for each happy-path test
- [ ] No workarounds — tests use the spec's API
- [ ] No manually-set framework state without a corresponding framework-path test
- [ ] One behavior per test, named as a sentence describing that behavior

## Reviewer's Checklist

Before approving tests:

- [ ] I derived my own expected test list from the spec BEFORE reading the tests
- [ ] Every assertion traces to a spec requirement (checked, not assumed)
- [ ] No spec requirements are untested
- [ ] Plausible implementation shortcuts would be caught by the test suite
- [ ] No tests that validate likely buggy behavior or route around broken APIs
