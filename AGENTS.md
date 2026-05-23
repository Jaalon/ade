# AGENTS.md

## Overview

- Go 1.26 module (`automated_dev_environment`) — no source code yet.
- OpenCode agent skills in `.opencode/skills/` for specification/story/task workflows (EN + FR).
- MCP server configured in `.opencode/config.json`.

## Available workflows (skills)

Use these via OpenCode's skill-loading mechanism or by prompting the agent to read the SKILL.md:

| Skill dir | Purpose |
|---|---|
| `specification-en` / `specification-fr` | Iterative spec creation |
| `story-en` / `story-fr` | Break spec into user stories |
| `tasks-fr` | Break stories into code tasks |
| `feedback-fr` | Generate stories from user feedback |

Skills write to `docs/` directories (`docs/specification/`, `docs/stories/`, etc.) and use `questions.md` for async Q&A.

## Project state

No Go source files exist yet. After generating specs/stories/tasks, agents should create Go packages under the module root following standard Go conventions.
