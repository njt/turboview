# Spec Compliance Reviewer Prompt Template

Use this template when dispatching a spec compliance reviewer subagent.

**Purpose:** Verify implementer built what was requested (nothing more, nothing less)

```
Task tool (general-purpose):
  description: "Review spec compliance for Task N"
  prompt: |
    You are reviewing whether an implementation matches its specification.

    ## What Was Requested

    [FULL TEXT of task requirements]

    ## What Implementer Claims They Built

    [From implementer's report]

    ## CRITICAL: Do Not Trust the Report

    The implementer finished suspiciously quickly. Their report may be incomplete,
    inaccurate, or optimistic. You MUST verify everything independently.

    **DO NOT:**
    - Take their word for what they implemented
    - Trust their claims about completeness
    - Accept their interpretation of requirements

    **DO:**
    - Read the actual code they wrote
    - Compare actual implementation to requirements line by line
    - Check for missing pieces they claimed to implement
    - Look for extra features they didn't mention

    ## Your Job

    Read the implementation code and verify:

    **Missing requirements:**
    - Did they implement everything that was requested?
    - Are there requirements they skipped or missed?
    - Did they claim something works but didn't actually implement it?

    **Extra/unneeded work:**
    - Did they build things that weren't requested?
    - Did they over-engineer or add unnecessary features?
    - Did they add "nice to haves" that weren't in spec?

    **Misunderstandings:**
    - Did they interpret requirements differently than intended?
    - Did they solve the wrong problem?
    - Did they implement the right feature but wrong way?

    **Test integrity (CRITICAL):**
    - Do tests exercise the ACTUAL API described in the spec, or a workaround?
      (e.g., if the spec says "MessageBox() opens a modal dialog," does the test
      call MessageBox(), or does it create a plain Window and insert it manually?)
    - Do tests validate CORRECT behavior, or do they validate a bug as correct?
      (e.g., if the spec says "button fires a command event," does the test check
      that the event IS a command, or that the event was cleared/consumed?)
    - If a test manually sets up state that the framework should provide (focus,
      ownership, visibility), flag it — the test may pass while the framework path
      is broken.
    - If the implementer built workarounds for broken behavior in other components,
      that is a spec compliance failure: the test should test the spec's API, not
      route around it. Report it as ❌ even if tests pass.

    **Verify by reading code AND tests, not by trusting report.**

    Report:
    - ✅ Spec compliant (if everything matches after code inspection)
    - ❌ Issues found: [list specifically what's missing or extra, with file:line references]
```
