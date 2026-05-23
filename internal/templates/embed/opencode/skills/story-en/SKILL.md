---
name: story-en
description: Breaking down specifications into User Stories (English)
---

# Skill: Stories (EN)

This guide describes how to generate user stories from a specification considering documentation and test automation.

## What I Do

I break down a specification into actionable user stories, taking into account software documentation and test automation. Each story must bring proven value to the end user.

## When to Use Me

Use this skill when you have a project specification and want to generate structured, testable, and documented user stories.

## Workflow

### Case 1: No Stories Exist (start from scratch)

**Step 1 - Read the specification:**
- User provides the specification path as argument (ex: specification.md or docs/specification/specification.md)
- Read the provided specification
- Generate user stories that meet the criteria:
  - Consider existing software documentation
  - Integrate test automation
  - Bring proven value to the end user
- Create a working directory (ex: docs/stories/) if it doesn't exist
- Generate story files (ex: docs/stories/story-001.md, story-002.md, etc.)

**Step 2 - Create the questions file:**
- Create a temporary file docs/stories/questions.md
- This file will contain the questions the user needs to answer

**Step 3 - Analysis and clarification:**
- Analyze the generated stories
- Verify consistency with the specification
- Identify what is unclear or missing
- Add clarification questions in questions.md with:
  - Question context
  - Possible proposals so the user can answer directly
  - References to documentation if applicable

**Step 4 - Update:**
- User answers questions in questions.md
- User types: I've answered the questions, update the stories
- Update stories based on given answers
- Return to Step 3

**Step 5 - Generate implementation plan and workflow end:**
- When there's nothing left to clarify (no more questions in questions.md)
- Delete the questions.md file
- Analyze dependencies between stories (from ## Dependencies section of each story)
- Generate the file docs/stories/implementation-plan.md containing:
  - Recommended implementation order (based on dependencies, user value, complexity)
  - Stories that can be developed in parallel (without cross-dependencies)
  - Justifications for the proposed order
  - Dependency diagram in Mermaid format
- Announce the end of the workflow

### Case 2: Stories Already Exist

- Start directly at Step 3 of Case 1
- Read existing stories in the working directory
- Analyze and iterate until complete resolution

## File Formats

### Story (story-XXX.md)
``markdown
# Story #XXX: [Story Title]

## Description
As a [role], I want [action] so that [benefit].

## Acceptance Criteria
- [ ] Criterion 1
- [ ] Criterion 2
- [ ] Criterion 3

## Automated Tests
- Unit test: [Description]
- Integration test: [Description]
- E2E test: [Description]

## Documentation
- [Link to impacted documentation]
- [New documentation to create]

## User Value
[Description of proven value for the end user]

## Dependencies
- [Story #YYY]
- [Existing documentation]
``

### questions.md
``markdown
# Clarification Questions - Stories

## Context
[Specification used: spec.md]
[Stories generated: 5 stories]

## Questions
1. **[Story #002 - Acceptance Criterion]**
   - Context: Story #002 mentions authentication but criteria don't specify role management.
   - Question: What roles should be managed (Admin, User, Guest)?
   - Proposals:
     - a) Admin (full access), User (limited access), Guest (read-only)
     - b) Admin and User only
     - c) Other: [specify]

2. **[Story #004 - Automated Test]**
   - Context: Story #004 requires E2E tests for task drag-and-drop.
   - Question: Do you already use an E2E framework (Cypress, Playwright, Selenium)?
   - Proposals:
     - a) Cypress (already configured in the project)
     - b) Playwright (recommended for modern apps)
     - c) Other: [specify]

## Detected Inconsistencies
- [Description of inconsistency between stories or with the specification]
``

### implementation-plan.md
``markdown
# Story Implementation Plan

## Recommended Implementation Order
1. **Story #001** - [Title] (Priority: High)
   - Justification: [No dependencies, high user value]
2. **Story #002** - [Title] (Priority: High)
   - Justification: [Depends on Story #001, required for subsequent features]
3. **Story #003** - [Title] (Priority: Medium)
   - Justification: [Can be developed in parallel with #004]

## Stories Developable in Parallel
- **Group 1**: Story #003, Story #004 (no cross-dependencies)
- **Group 2**: Story #005, Story #006 (depend only on already implemented stories)

## Dependency Diagram (Mermaid format)
``mermaid
flowchart TD
    %% Ordered stories
    S001[Story #001: Title] --> S002[Story #002: Title]
    S002 --> S003[Story #003: Title]
    S002 --> S004[Story #004: Title]
    
    %% Parallel groups
    subgraph Parallel1
        S003
        S004
    end
    
    S003 --> S005[Story #005: Title]
    S004 --> S005
    S005 --> S006[Story #006: Title]
    
    subgraph Parallel2
        S005
        S006
    end
``

## Notes
- Stories should be implemented in the indicated order to respect dependencies
- Parallel groups can be developed simultaneously by different teams
``

## How to Move from One Step to Another

| Transition | User Action | Agent Action |
|------------|-------------|--------------|
| Step 1 → 2 | Provide the specification path as argument | Generates initial stories and creates questions.md |
| Step 2 → 3 | Automatic after questions.md creation | Analyzes stories and adds questions |
| Step 3 → 4 | Type I've answered the questions, update the stories | Updates stories according to answers |
| Step 4 → 3 | Automatic after update | Re-analyzes updated stories |
| Step 4 → 5 | Automatic (no more questions) | Deletes questions.md, generates implementation-plan.md, announces end |

**Important:** The agent does not monitor file changes in real-time. The user must always give an explicit instruction for the agent to move to the next step.

## Usage Example

**Case 1 - Start from scratch:**
``
/ story-en specification.md
``

**Case 2 - Existing stories:**
``
/ story-en
``
(Agent automatically detects existing stories in docs/stories/ and starts analysis)
