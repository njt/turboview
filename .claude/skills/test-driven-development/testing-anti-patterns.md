# Testing Anti-Patterns

**Load this reference when:** writing or changing tests, adding mocks, or tempted to add test-only methods to production code.

## Overview

Tests must verify real behavior, not mock behavior. Mocks are a means to isolate, not the thing being tested.

**Core principle:** Test what the code does, not what the mocks do.

**Following strict TDD prevents these anti-patterns.**

## The Iron Laws

```
1. NEVER test mock behavior
2. NEVER add test-only methods to production classes
3. NEVER mock without understanding dependencies
```

## Anti-Pattern 1: Testing Mock Behavior

**The violation:**
```typescript
// ❌ BAD: Testing that the mock exists
test('renders sidebar', () => {
  render(<Page />);
  expect(screen.getByTestId('sidebar-mock')).toBeInTheDocument();
});
```

**Why this is wrong:**
- You're verifying the mock works, not that the component works
- Test passes when mock is present, fails when it's not
- Tells you nothing about real behavior

**your human partner's correction:** "Are we testing the behavior of a mock?"

**The fix:**
```typescript
// ✅ GOOD: Test real component or don't mock it
test('renders sidebar', () => {
  render(<Page />);  // Don't mock sidebar
  expect(screen.getByRole('navigation')).toBeInTheDocument();
});

// OR if sidebar must be mocked for isolation:
// Don't assert on the mock - test Page's behavior with sidebar present
```

### Gate Function

```
BEFORE asserting on any mock element:
  Ask: "Am I testing real component behavior or just mock existence?"

  IF testing mock existence:
    STOP - Delete the assertion or unmock the component

  Test real behavior instead
```

## Anti-Pattern 2: Test-Only Methods in Production

**The violation:**
```typescript
// ❌ BAD: destroy() only used in tests
class Session {
  async destroy() {  // Looks like production API!
    await this._workspaceManager?.destroyWorkspace(this.id);
    // ... cleanup
  }
}

// In tests
afterEach(() => session.destroy());
```

**Why this is wrong:**
- Production class polluted with test-only code
- Dangerous if accidentally called in production
- Violates YAGNI and separation of concerns
- Confuses object lifecycle with entity lifecycle

**The fix:**
```typescript
// ✅ GOOD: Test utilities handle test cleanup
// Session has no destroy() - it's stateless in production

// In test-utils/
export async function cleanupSession(session: Session) {
  const workspace = session.getWorkspaceInfo();
  if (workspace) {
    await workspaceManager.destroyWorkspace(workspace.id);
  }
}

// In tests
afterEach(() => cleanupSession(session));
```

### Gate Function

```
BEFORE adding any method to production class:
  Ask: "Is this only used by tests?"

  IF yes:
    STOP - Don't add it
    Put it in test utilities instead

  Ask: "Does this class own this resource's lifecycle?"

  IF no:
    STOP - Wrong class for this method
```

## Anti-Pattern 3: Mocking Without Understanding

**The violation:**
```typescript
// ❌ BAD: Mock breaks test logic
test('detects duplicate server', () => {
  // Mock prevents config write that test depends on!
  vi.mock('ToolCatalog', () => ({
    discoverAndCacheTools: vi.fn().mockResolvedValue(undefined)
  }));

  await addServer(config);
  await addServer(config);  // Should throw - but won't!
});
```

**Why this is wrong:**
- Mocked method had side effect test depended on (writing config)
- Over-mocking to "be safe" breaks actual behavior
- Test passes for wrong reason or fails mysteriously

**The fix:**
```typescript
// ✅ GOOD: Mock at correct level
test('detects duplicate server', () => {
  // Mock the slow part, preserve behavior test needs
  vi.mock('MCPServerManager'); // Just mock slow server startup

  await addServer(config);  // Config written
  await addServer(config);  // Duplicate detected ✓
});
```

### Gate Function

```
BEFORE mocking any method:
  STOP - Don't mock yet

  1. Ask: "What side effects does the real method have?"
  2. Ask: "Does this test depend on any of those side effects?"
  3. Ask: "Do I fully understand what this test needs?"

  IF depends on side effects:
    Mock at lower level (the actual slow/external operation)
    OR use test doubles that preserve necessary behavior
    NOT the high-level method the test depends on

  IF unsure what test depends on:
    Run test with real implementation FIRST
    Observe what actually needs to happen
    THEN add minimal mocking at the right level

  Red flags:
    - "I'll mock this to be safe"
    - "This might be slow, better mock it"
    - Mocking without understanding the dependency chain
```

## Anti-Pattern 4: Incomplete Mocks

**The violation:**
```typescript
// ❌ BAD: Partial mock - only fields you think you need
const mockResponse = {
  status: 'success',
  data: { userId: '123', name: 'Alice' }
  // Missing: metadata that downstream code uses
};

// Later: breaks when code accesses response.metadata.requestId
```

**Why this is wrong:**
- **Partial mocks hide structural assumptions** - You only mocked fields you know about
- **Downstream code may depend on fields you didn't include** - Silent failures
- **Tests pass but integration fails** - Mock incomplete, real API complete
- **False confidence** - Test proves nothing about real behavior

**The Iron Rule:** Mock the COMPLETE data structure as it exists in reality, not just fields your immediate test uses.

**The fix:**
```typescript
// ✅ GOOD: Mirror real API completeness
const mockResponse = {
  status: 'success',
  data: { userId: '123', name: 'Alice' },
  metadata: { requestId: 'req-789', timestamp: 1234567890 }
  // All fields real API returns
};
```

### Gate Function

```
BEFORE creating mock responses:
  Check: "What fields does the real API response contain?"

  Actions:
    1. Examine actual API response from docs/examples
    2. Include ALL fields system might consume downstream
    3. Verify mock matches real response schema completely

  Critical:
    If you're creating a mock, you must understand the ENTIRE structure
    Partial mocks fail silently when code depends on omitted fields

  If uncertain: Include all documented fields
```

## Anti-Pattern 5: All Unit Tests, No Integration Tests

**The violation:**
```
✅ 400 unit tests passing
✅ Every component tested in isolation
❌ Components don't actually work together
❌ Modal dialogs invisible (ExecView never redraws)
❌ Focus never propagates (Application never sets SfFocused)
❌ Menu popups never open (MenuBar never calls ExecView)
```

**Why this is wrong:**
- Unit tests with mocked dependencies prove each piece works alone
- They say nothing about whether the pieces connect
- State that the framework normally sets (focus flags, owner references) gets manually set in unit tests, hiding the fact that the framework doesn't set it
- Each subagent/developer tests their own work, nobody tests the seams
- 400 passing tests create false confidence — "it's well tested" when the core user flows are broken

**Real-world example:** A TUI framework had 493 passing tests. Dialogs were invisible, menus never opened, and keyboard events never reached widgets. Every component worked perfectly in isolation. The connections between them were completely untested.

**The fix:**
```
For every component that connects to another:
1. Write a unit test (component in isolation) — proves the mechanism
2. Write an integration test (real component chain) — proves the connection
3. Both must pass before claiming the task is complete

The integration test for a Button isn't "does fire() set the command?"
It's "does pressing Enter in a real Application reach a Button in a Window
and produce the expected command?"
```

**Warning signs:**
- Tests that manually set state the framework should set (e.g., `btn.SetState(SfFocused, true)` instead of letting the Application propagate focus)
- Tests that mock the owner/parent when a real owner exists from a previous task
- Tests that verify method return values but never test a real user scenario
- High test count with no test that creates more than one real component
- E2e/integration tests that work around broken behavior instead of testing it

### Gate Function

```
BEFORE claiming a component is tested:
  Ask: "Do I have a test where this component is used inside its REAL parent?"
  Ask: "Do I have a test where the framework (not my test setup) provides the
        state this component needs (focus, owner, events)?"

  IF no to either:
    Write the missing integration test
    If it fails, the component has a connection bug — fix it

  Unit tests + integration tests = tested
  Unit tests alone = half tested
```

## When Mocks Become Too Complex

**Warning signs:**
- Mock setup longer than test logic
- Mocking everything to make test pass
- Mocks missing methods real components have
- Test breaks when mock changes

**your human partner's question:** "Do we need to be using a mock here?"

**Consider:** Integration tests with real components often simpler than complex mocks

## TDD Prevents These Anti-Patterns

**Why TDD helps:**
1. **Write test first** → Forces you to think about what you're actually testing
2. **Watch it fail** → Confirms test tests real behavior, not mocks
3. **Minimal implementation** → No test-only methods creep in
4. **Real dependencies** → You see what the test actually needs before mocking

**If you're testing mock behavior, you violated TDD** - you added mocks without watching test fail against real code first.

## Anti-Pattern 6: Tests That Validate Bugs as Correct

**The violation:**
```go
// Button.fire() calls event.Clear(), setting What = EvNothing
// Test asserts the event is cleared — validating the bug
func TestButton_Click_FiresCommand(t *testing.T) {
    btn.HandleEvent(clickEvent)
    if !clickEvent.IsCleared() {           // ← testing the bug
        t.Error("click should be consumed")
    }
}
```

**Why this is wrong:**
- The spec says "button fires a command event" — the test should check for a command event
- `IsCleared()` means `What == EvNothing`, which is the opposite of "fires a command"
- The test passes because it tests what the code does, not what it should do
- TDD was supposed to prevent this, but the test and implementation were written by the same agent in the same session — the agent encoded its misunderstanding in both

**The fix:**
```go
func TestButton_Click_FiresCommand(t *testing.T) {
    btn.HandleEvent(clickEvent)
    if clickEvent.What != EvCommand {      // ← testing the spec
        t.Error("click should produce EvCommand")
    }
}
```

### Gate Function

```
BEFORE writing a test assertion:
  Ask: "Am I testing what the SPEC says should happen, or what the CODE does?"

  IF you wrote the implementation and the test in the same session:
    You are biased. Re-read the spec requirement, then write the assertion
    from the spec, not from your implementation.

  IF something didn't work as expected and you changed the test:
    STOP. That's test weakening. Change the code, not the test.

  IF something didn't work as expected and you built a workaround:
    STOP. Report BLOCKED. The workaround will pass tests but hide a real bug.
```

## Anti-Pattern 7: E2e Tests That Route Around Broken Behavior

**The violation:**
```go
// Spec says: MessageBox() opens a modal dialog
// Reality: ExecView doesn't redraw, so MessageBox is invisible
// Test app: creates a plain Window instead of calling MessageBox()

func openMessageDialog() {
    w := NewWindow(...)     // workaround — not testing MessageBox
    w.Insert(okBtn)
    app.Desktop().Insert(w) // manual insert, not ExecView
}
```

**Why this is wrong:**
- The test "passes" but doesn't test the feature
- The actual API (`MessageBox`, `ExecView`) is completely untested
- The workaround hides a critical bug (invisible modal dialogs)
- 15 e2e tests pass with 100% workaround coverage and 0% API coverage

**The fix:**
- Tests must exercise the API described in the spec, not alternatives
- If the spec API doesn't work, the test should FAIL, not work around it
- If you discover a bug in another component while writing tests, report it — don't route around it

### Gate Function

```
BEFORE writing a test app or e2e test:
  Ask: "Am I calling the same functions a user would call?"
  Ask: "If I removed my workarounds, would the test still pass?"

  IF you built a custom handler because the standard API didn't work:
    That's a bug report, not a test. Report BLOCKED.

  IF you used a simpler component because the spec'd one was broken:
    That's a bug report, not a test. Report BLOCKED.
```

## Quick Reference

| Anti-Pattern | Fix |
|--------------|-----|
| Assert on mock elements | Test real component or unmock it |
| Test-only methods in production | Move to test utilities |
| Mock without understanding | Understand dependencies first, mock minimally |
| Incomplete mocks | Mirror real API completely |
| All unit tests, no integration | Write integration tests with real components |
| Tests validate bugs as correct | Test from the spec, not from your implementation |
| E2e routes around broken behavior | Test the real API; report BLOCKED if it's broken |

## Red Flags

- Assertion checks for `*-mock` test IDs
- Methods only called in test files
- Mock setup is >50% of test
- Test fails when you remove mock
- Can't explain why mock is needed
- Mocking "just to be safe"

## The Bottom Line

**Mocks are tools to isolate, not things to test.**

If TDD reveals you're testing mock behavior, you've gone wrong.

Fix: Test real behavior or question why you're mocking at all.
