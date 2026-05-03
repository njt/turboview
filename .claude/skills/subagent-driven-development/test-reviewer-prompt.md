# Test Reviewer Subagent Prompt Template

Use this template when dispatching a test reviewer subagent.

**Purpose:** Review tests as tests — verify they would catch real bugs and that their assertions come from the spec, not from implementation assumptions.

**Important:** Provide the spec text inline in the prompt. Provide the test files as paths the agent must read (not inline). This forces the agent to derive expectations from the spec before it can see the tests.

```
Task tool (general-purpose):
  description: "Review tests for Task N: [task name]"
  prompt: |
    You are reviewing tests BEFORE any implementation exists. Your job is
    to verify these tests would actually catch bugs — not to check code
    style or conventions.

    ## CRITICAL: Derive expectations BEFORE reading tests

    You will receive the spec below and the test file paths at the end.

    **Before you open any test file**, you must:
    1. Read the spec carefully
    2. Write down YOUR list of what you would test — each spec requirement
       and the specific behavior you'd verify
    3. Only THEN read the test files and compare

    This is not a suggestion. Reading the tests first anchors you to the
    writer's assumptions and blind spots. If you skip the independent
    derivation, you will approve tests that miss requirements because
    you never noticed the requirements were missing.

    ## The Specification

    [FULL TEXT of task/feature spec — paste it here]

    ## Test Quality Rules

    [FULL TEXT of test-quality-gold-standard.md — paste it here]

    ## Test Writer's Report

    [From test writer: what they wrote, any assumptions they flagged]

    ## Your Review

    After completing your independent derivation, read the test files
    and assess:

    ### 1. Missing requirements

    Compare your independently-derived list against the writer's tests.
    - What spec requirements aren't tested?
    - What edge cases did you identify that the writer missed?
    - If the writer made assumptions (DONE_WITH_CONCERNS), are they
      reasonable?

    This is the most important check. A test suite that covers all
    requirements is more valuable than one that covers half of them
    with beautiful assertions.

    ### 2. Spec traceability

    For each assertion: can you trace it back to a concrete sentence
    in the spec?

    Flag any assertion where the justification is "it seems right" or
    "the code would probably do this." Those are implementation-derived
    oracles — they encode assumptions, not requirements.

    You don't need to categorize every assertion into an oracle taxonomy.
    The question is binary: spec says it, or spec doesn't say it.

    ### 3. Falsification gaps

    For each spec requirement, ask: what's the most plausible shortcut
    an implementer could take? Would the test suite catch it?

    List specific shortcuts that would pass all current tests. These
    are the tests the writer needs to add.

    ### 4. Anti-pattern scan

    Check for:
    - **Tests that validate likely buggy behavior**: assertion matches
      a plausible bug rather than the spec requirement
    - **Tests that route around the spec's API**: test uses a workaround
      instead of calling the function the spec describes
    - **Tests that manually set framework state**: focus, ownership,
      visibility set in test setup instead of provided by the framework
      (acceptable for unit tests, but there should also be a framework-
      path test)
    - **Multiple behaviors in one test**: test name has "and" or test
      has assertions about unrelated behaviors

    ### 5. Integration coverage (if applicable)

    If this component connects to existing components: is there at
    least one test using the real dependency? If not, flag it.

    ## Report Format

    **Assessment:** ✅ Approved | ⚠️ Approved with notes | ❌ Revisions needed

    **My independent test list:**
    [Your derived list — show your work so the controller can assess
    the quality of the review]

    **Missing requirements:** [spec requirements with no test — CRITICAL]

    **Falsification gaps:** [shortcuts that would pass all tests]

    **Issues:** [anything else, with file:line and how to fix]

    **If ❌ Revisions needed:**
    List exactly what the test writer must change. Be specific enough
    that a fresh test writer subagent can make the fixes without
    additional context.

    ## Test Files to Review

    [File paths — the agent must Read these after completing step 2]
```
