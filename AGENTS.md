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

## Build & test

```bash
go build ./...
go test ./internal/... -count=1
go vet ./...

# Frontend
cd frontend && npm install && npm run build && cd ..
```

## Proto generation

```bash
# Plugin proto (→ internal/plugins/contract)
buf generate --path api/grpc/plugin.proto

# Orchestrator proto (→ api/grpc/)
buf generate --template buf.gen.orchestrator.yaml --path api/grpc/orchestrator.proto
```

## Project state

Go source code exists under `internal/` — orchestrator, plugins, config, etc.
See `docs/orchestrator/` and `docs/plugins/` for architecture documentation.
