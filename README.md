# Rook

**Rook** is a standalone, autonomous security agent for vulnerability research,
bug hunting and source-code auditing. It is a single Go executable built on the
[ChatBotKit Go SDK](https://github.com/chatbotkit/go-sdk), with a library of
security skills embedded directly into the binary — no external files, no setup
beyond an API key.

Give Rook a target and a scope, and it works through the problem the way a
researcher would: reconnaissance, analysis, hypothesis, verification, and a
written report.

> ⚠️ **Authorized use only.** Rook is an offensive-security tool. Only run it
> against systems, code and services you own or are explicitly authorized to
> test. Always pass an explicit `--scope`.

## Features

- **Single self-contained binary.** The skill library is compiled into the
  executable via Go's `embed`, so it ships and runs as one file.
- **Autonomous agent loop.** Built on the Go SDK's `agent.ExecuteWithTools` —
  the agent plans, acts, tracks progress and exits on its own, bounded by
  `--max-iterations`.
- **Built-in tools.** File read/write/edit and sandboxed shell execution via
  the SDK's `DefaultTools`.
- **Embedded skill library.** Phase-by-phase security playbooks (see below)
  surfaced to the model through the SDK skills feature.
- **Cross-platform releases.** GitHub Actions builds binaries for Linux, macOS
  and Windows (amd64/arm64) on every tag.

## Install

### From a release

Download the archive for your platform from the
[releases page](https://github.com/chatbotkit/rook/releases), extract it, and
put `rook` on your `PATH`.

### From source

```bash
go install github.com/chatbotkit/rook/cmd/rook@latest
```

Or clone and build with the provided `Makefile`:

```bash
make build      # → ./rook
```

## Usage

```bash
export CHATBOTKIT_API_SECRET="your-api-key"

# Audit a local codebase
rook --scope "repo: ./server, no network access" \
     "Audit the HTTP handlers in ./server for injection and auth bypass bugs"

# Hunt with reasoning streamed to the terminal
rook -v --scope-file scope.txt "Find SSRF in the URL-fetching service"

# Version
rook version
```

Rook loads a `.env` file automatically if present (see `.env.example`).

### Flags

| Flag | Default | Description |
| ---- | ------- | ----------- |
| `--model` | `qwen-3.6-plus` | Model the agent reasons with |
| `--max-iterations` | `40` | Maximum agent iterations before a forced stop |
| `--scope` | — | Authorization boundary (hosts, repos, paths) |
| `--scope-file` | — | Read the authorization scope from a file |
| `-v`, `--verbose` | `false` | Stream the agent's reasoning tokens to stdout |
| `-V`, `--version` | — | Print version and exit |

The agent's findings stream to **stderr**; with `--verbose`, reasoning tokens
stream to **stdout**. The final report is delivered as the agent's response —
Rook does not write files on its own. If you want the report (or any other
artifact) saved to disk, ask for it in the task and the agent will use its
`write` tool.

## Embedded Skills

Rook ships with **51 security skills** — each a `SKILL.md` playbook under
[`skills/`](skills/), embedded into the binary at build time and offered to the
agent as it works. They cover, roughly:

- **Methodology & mindset** — `bug-bounty`, `bb-methodology`, `redteam-mindset`,
  `bb-local-toolkit`, `hunt-dispatch`.
- **Web/API vulnerability hunting** (24 `hunt-*` classes + `security-arsenal`) —
  IDOR, SQLi, XSS, SSRF, RCE, SSTI, XXE, CSRF, OAuth, SAML, GraphQL, auth/MFA
  bypass, ATO, business logic, cache poisoning, HTTP smuggling, file upload,
  API misconfig, race conditions, and more.
- **Enterprise & infrastructure attack chains** — `m365-entra-attack`,
  `okta-attack`, `cloud-iam-deep`, `vmware-vcenter-attack`,
  `enterprise-vpn-attack`, `hunt-sharepoint`, `hunt-aspnet`, `hunt-ntlm-info`,
  `apk-redteam-pipeline`, `supply-chain-attack-recon`.
- **Recon & OSINT** — `web2-recon`, `offensive-osint`, `osint-methodology`,
  `hunt-subdomain`.
- **Web3** — `web3-audit`, `meme-coin-audit`.
- **Triage, reporting & hygiene** — `triage-validation`, `bugcrowd-reporting`,
  `report-writing`, `redteam-report-template`, `evidence-hygiene`,
  `mid-engagement-ir-detection`.

These skills are sourced from the **claude-bughunter** project — see
[Credits](#credits).

### Adding a skill

Create `skills/<name>/SKILL.md` with YAML front matter:

```markdown
---
name: My Skill
description: One sentence the model uses to decide when to apply this skill.
---

# My Skill

Step-by-step guidance...
```

Rebuild the binary — the new skill is picked up automatically by the `embed`
directive. No registration code required.

## How it works

```
cmd/rook          CLI: flags, .env, signal handling, version
internal/config   Central config: default model, max iterations, system prompt
internal/agent    Loads embedded skills, registers tools, drives the agent loop
internal/version  Build-time version + GitHub release update check
embed.go          //go:embed skills  →  the embedded skill library
skills/           SKILL.md playbooks compiled into the binary
```

The default model and the agent's system prompt (backstory) live in one place —
[`internal/config/config.go`](internal/config/config.go) — so they can be tuned
without touching the CLI or the agent loop.

At startup Rook loads the embedded skills with `agent.LoadSkillsFromFS`,
registers `agent.DefaultTools()`, builds a security-focused backstory that
pins the agent to your authorized scope, and runs `agent.ExecuteWithTools`
until the agent calls `exit`.

## Development

The committed `go.mod` pins a published version of the Go SDK, so the
standalone repository builds from a clean clone with no extra steps:

```bash
git clone https://github.com/chatbotkit/rook
cd rook
go build ./...        # or: make build
```

```bash
make build    # build ./rook
make test     # run tests
make vet      # go vet
make dist     # cross-platform release archives under dist/
```

### Developing against a local go-sdk

To build against a local checkout of the Go SDK instead of the published
module, place it at `../go-sdk` (or anywhere) and create a Go workspace:

```bash
make workspace        # writes a gitignored go.work
```

`go.work` is **gitignored**, so it only affects your local builds. See
[RELEASES.md](RELEASES.md) for the release flow.

## Credits

Rook's embedded skill library is sourced from the **claude-bughunter** project
by **[Sachin Sharma](https://www.linkedin.com/in/sachinsharma8080/)**:

> https://github.com/elementalsouls/Claude-BugHunter

The skills are used under the MIT License (Copyright © 2026 Sachin Sharma). The
full upstream license is preserved in [NOTICE.md](NOTICE.md). Our thanks to the
author and the bug-bounty community whose disclosed reports informed them.

## License

Rook itself is MIT licensed — see [LICENSE](LICENSE). Bundled third-party
content retains its original license; see [NOTICE.md](NOTICE.md).
