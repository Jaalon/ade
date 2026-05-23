---
name: specification-en
description: Iterative project specification creation (English)
---

# Skill: Specification (EN)

This guide describes how to generate a project specification iteratively with questions and answers until complete resolution.

## What I Do

I guide the creation of a complete project specification through an iterative process of questions, answers, and clarification until all inconsistencies are resolved.

## When to Use Me

Use this skill when you want to create a structured project specification from a rough idea, iterating until everything is clear and consistent.

## Process Steps

### Step 1: Initiation
- Create a documentation working directory (ex: `docs/specification/`) if it doesn't exist
- Generate the `questions.md` file if it doesn't exist
- If a specification file already exists, use its content as the specification base
- Otherwise, generate this empty specification.md file and ask the user to describe what the project will do in questions.md
- Ask open-ended questions in questions.md to better understand the project

### Step 2: Specification Creation
- Read user responses in `questions.md`
- Update `specification.md` with:
  - Structured project description
  - Clear placeholders (ex: `[TO_CLARIFY: ...]`) to identify fuzzy points
  - Inconsistency markers (ex: `[INCONSISTENCY: ...]`)
- Generate a new `questions.md` file containing:
  - Questions needed to resolve placeholders
  - Relevant context for each question
  - Detected inconsistencies to resolve

### Step 3: Update
- Read responses in `questions.md`
- Update `specification.md` by replacing placeholders with answers
- Resolve identified inconsistencies

### Step 4: Iteration
- Relaunch Step 2 with a new agent
- Check if there are any remaining placeholders or inconsistencies
- If yes, continue iteration
- If no, delete the questions.md file and consider the specification complete

## File Formats

### questions.md
```markdown
# Specification Questions

## Context
[Brief summary of current project]

## Questions
1. [Clear question with context]
2. [Clear question with context]
...

## Inconsistencies to Resolve
- [Description of inconsistency]
...
```

### specification.md
```markdown
# Specification - [Project Name]

## Overview
[General description]

## Features
[List of features]

## [Section with placeholder]
[TO_CLARIFY: Necessary details about...]

## [Section with inconsistencies]
[INCONSISTENCY: Identified conflict between...]
```

## Usage Example

**Launching the skill:**
```
/ specification-en
```

**Step 1 - Agent creates `docs/specification/questions.md`:**
```markdown
# Specification Questions

## Context
Explain what the project will do

## Questions
1. Describe briefly what your project will do (objective, main functionality)
2. Who is the target audience?
3. Are there any specific technical constraints?
```

**User responds in `questions.md`:**
```markdown
1. A web application to manage collaborative tasks with Kanban boards
2. Agile development teams
3. Must be mobile compatible, no tech stack constraints
```

**Step 2 - Agent generates `specification.md`:**
```markdown
# Specification - Collaborative Kanban Application

## Overview
Web application allowing agile development teams to manage their tasks via collaborative Kanban boards.

## Features
- Creation of customizable boards and columns
- Task management with drag-and-drop
- [TO_CLARIFY: Authentication system (SSO, email, OAuth?)]
- [TO_CLARIFY: Role and permission management]

## Technical Constraints
- Mobile compatible (responsive design)
- [INCONSISTENCY: No tech stack imposed but need real-time for collaboration]
```

**And generates a new `questions.md`:**
```markdown
# Specification Questions

## Context
Kanban application for agile teams, mobile compatible, need real-time for collaboration

## Questions
1. [TO_CLARIFY: Authentication system] Which authentication system do you want? (SSO, email/password, OAuth with Google/GitHub?)
2. [TO_CLARIFY: Roles] What roles do you plan? (Admin, member, guest?)
3. [INCONSISTENCY: Real-time vs Tech stack] You have no tech stack preference but need real-time. Do you prefer a WebSocket solution, Server-Sent Events, or an approach with BaaS like Supabase/Firebase?
```

**Transition Step 1 → Step 2:**
The user must inform the agent that the questions are answered by simply typing:
```
I've answered the questions, continue
```
The agent does not automatically detect file modifications. The user must explicitly give the instruction to continue.

**Step 3 - User responds in `questions.md`, then types:**
```
Questions answered, update the specification
```
Agent updates `specification.md` by replacing placeholders with answers.

**Transition Step 3 → Step 4:**
Agent automatically checks if there are any placeholders left. If there are, it relaunches Step 2 with a new cycle of questions.

**Step 4 - Iteration:**
Agent relaunches Step 2, verifies there are no more placeholders, the specification is complete!

## How to Move from One Step to Another

| Transition | User Action | Agent Action |
|------------|-------------|--------------|
| Step 1 → 2 | Type `continue` or `I've answered` | Reads `questions.md` and generates `specification.md` + new `questions.md` |
| Step 2 → 3 | Type `continue` or `questions answered` | Reads answers and updates `specification.md` |
| Step 3 → 4 | Automatic | Checks if placeholders remain, relaunches Step 2 if necessary |
| End | Automatic | Announces that the specification is complete |

**Important:** The agent does not monitor file changes in real-time. The user must always give an explicit instruction for the agent to move to the next step, except for the automatic transition Step 3 → 4.

## Completion Criteria

The specification is finished when `specification.md` no longer contains any `[TO_CLARIFY: ...]` or `[INCONSISTENCY: ...]` placeholders.
