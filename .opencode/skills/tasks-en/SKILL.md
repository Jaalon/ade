---
name: tasks-en
description: Breaks a story into executable "prompt" tasks-fr for a code agent with tests, documentation, and API examples
---

## What I do

I take an existing story and break it down into elementary tasks. Each task contains a maximally detailed prompt ready to be executed by a code agent, along with automated tests, documentation, and API examples (`.http` files). I verify the coherence of the generated tasks and clarify with the user via a `questions.md` file until everything is resolved.

## When to use me

Use this skill after creating your user stories with the `stories` skill, when you want to move to implementation.

## Workflow

### Step 1 - Read the story
- Take the story number provided as argument (e.g. `/ tasks-en 001`)
- Read the file `docs/stories/story-XXX.md`
- Extract the description, acceptance criteria, tests, and expected documentation
- Read the specification and other stories for full context

### Step 2 - Decomposition & generation with detailed prompts
- Analyze the story to identify atomic work units
- For each unit, generate an extremely detailed prompt including:
  - **Precise objective** of the task
  - **Functional description**: expected behavior, nominal and edge cases
  - **Files involved**: exact paths, nature of change (create, modify, delete)
  - **Technical constraints**: framework, patterns, project conventions, error handling, performance, security
  - **Function/class signatures**: names, parameters, types, return values
  - **Data structures**: schemas, models, DTOs, interfaces
  - **Tests to implement**: concrete unit, integration, and E2E scenarios
  - **Documentation**: what to document and where
  - **Usage examples**: code snippets, HTTP API requests
- Create a directory `docs/tasks/story-XXX/`
- Generate files `docs/tasks/story-XXX/task-YYY.md`

### Step 3 - Coherence verification
- Analyze the generated tasks against each other:
  - Check for circular dependencies or problematic execution order
  - Verify each task has clear inputs and outputs
  - Verify the sum of tasks covers 100% of the story acceptance criteria
  - Identify duplicates or overlaps between tasks
  - Check naming, types, and signature consistency across tasks
- Analyze coherence with the existing project:
  - Check compatibility with existing conventions and architecture
  - Verify referenced files exist or are correctly marked for creation

### Step 4 - Create questions.md
- Create a file `docs/tasks/story-XXX/questions.md`
- Add clarification questions for:
  - Undecided implementation choices
  - Inconsistencies between tasks
  - Ambiguities in the story
  - Dependencies and execution order
  - Alternative technical proposals
- Each question must include:
  - Context (task involved, story section)
  - Proposals to guide the user
  - Implications of each possible choice

### Step 5 - Update
- The user answers the questions in `questions.md`
- The user types: `I've answered, update the tasks`
- Update the tasks based on the answers
- Return to Step 3

### Step 6 - Workflow completion
- When nothing remains to clarify
- Delete the `questions.md` file
- Present the final plan to the user

## Detailed prompt format

### Task file structure (task-YYY.md)

```markdown
# Task #YYY - Story #XXX : [Task title]

## Objective
[Concise description of what this task should deliver]

## Context
- Story #XXX : [Link to story]
- Depends on : Task #ZZZ (if applicable)
- Required for : Task #WWW (if applicable)

## Prompt

As a code agent, you must implement the following.

### Functional description
[Detailed description of expected behavior]

**Nominal cases:**
- [Case 1]
- [Case 2]

**Edge cases:**
- [Edge case 1]
- [Edge case 2]

**Error handling:**
- [Error 1] → [Expected behavior]
- [Error 2] → [Expected behavior]

### Files involved

| File | Action | Description |
|------|--------|-------------|
| `src/feature/example.ts` | Create | New service |
| `src/feature/types.ts` | Modify | Add type X |
| `src/feature/__tests__/example.test.ts` | Create | Unit tests |

### Signatures

```typescript
// Function to implement
function example(
  param1: string,
  param2: number
): Promise<Result>
```

### Technical constraints
- **Framework** : Use [framework] (version X)
- **Pattern** : Follow [XXX] pattern already used in [reference file]
- **Style** : Respect the project's ESLint and Prettier config
- **Performance** : [constraint if applicable]
- **Security** : [constraint if applicable]
- **Tests** : Minimum 80% coverage for this task

### Tests to implement

#### Unit tests
- **File** : `src/feature/__tests__/example.test.ts`
- Scenario 1 : [Description]
  - Input : [Input data]
  - Expected result : [Output]
- Scenario 2 : [Description]
  - Input : [Input data]
  - Expected result : [Output]

#### Integration tests
- **File** : `src/feature/__tests__/integration.test.ts`
- Scenario : [Description]

### Documentation

#### Documentation to create
- `docs/feature/README.md` : New feature documentation
- `docs/feature/examples.md` : Usage examples

#### Documentation to update
- `docs/api/README.md` : Add section on the new endpoint

### API examples (if applicable)
- File : `docs/tasks/story-XXX/task-YYY-examples.http`

See HTTP request examples in the associated .http file.
```

### questions.md format

```markdown
# Clarification questions - Story #XXX

## Context
Story #XXX : [Story title]
Tasks generated : [task list]

## Questions

### 1. [Task #YYY] - [Topic]
**Context :** [Precise context description]
**Question :** [Multiple choice or open question]
**Proposals :**
- a) [Proposal A] → [Implication]
- b) [Proposal B] → [Implication]
- c) Other : [specify]

### 2. [Inconsistency between tasks #YYY and #ZZZ] - [Topic]
**Context :** Task #YYY defines X while task #ZZZ expects Y.
**Question :** Which definition is correct ?
**Proposals :**
- a) Align with task #YYY : [consequence]
- b) Align with task #ZZZ : [consequence]
- c) Define a new standard : [proposal]

### 3. [Story ambiguity] - [Topic]
**Context :** The story mentions [ambiguous point] without precision.
**Question :** [Clarification question]
**Proposals :**
- a) [Proposal A]
- b) [Proposal B]
```

## Step transitions

| Transition | User action | Agent action |
|------------|-------------|--------------|
| Step 1 | Provide story number (e.g. `/ tasks-en 001`) | Read story and project context |
| Step 1 → 2 | Automatic | Decompose and generate detailed prompts |
| Step 2 → 3 | Automatic | Verify task coherence |
| Step 3 → 4 | Automatic | Create `questions.md` with points to clarify |
| Step 4 → 5 | Answer in `questions.md` then type `I've answered, update the tasks` | Update tasks according to answers |
| Step 5 → 3 | Automatic | Re-verify coherence (iteration) |
| Step 5 → 6 | Automatic (no more questions) | Delete `questions.md`, present final plan |

**Important :** The agent does not monitor file modifications in real time. The user must always give an explicit instruction to move to the next step.

## Usage example

**Launch :**
```
/ tasks-en 001
```

**Reads `docs/stories/story-001.md` :**
Story describing the creation of a REST API for task management (CRUD).

**Decomposition into 3 tasks with detailed prompts :**
- `docs/tasks/story-001/task-001.md` : Data model, database, and validations
- `docs/tasks/story-001/task-002.md` : Full CRUD endpoints with signatures and error handling
- `docs/tasks/story-001/task-003.md` : Automated tests, documentation, and `.http` files

**Coherence verification :**
- Task #002 references a type defined in task #001 → OK
- Task #003 covers all test cases from acceptance criteria → OK
- Error handling : task #002 and specification aligned → OK

**Creates `docs/tasks/story-001/questions.md` :**
- Question about error response format (JSON API vs custom standard)
- Question about task execution order (task #001 must precede #002)

**Update :**
The user answers, the agent updates the tasks, iteration continues until everything is resolved.
