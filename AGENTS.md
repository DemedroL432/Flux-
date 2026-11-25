Flux / PulseShield — Agent Instructions
Environment and scope
- This repository is PulseShield (Flux), a Windows desktop networking application.
- Main stack: Go backend, Wails desktop shell, React frontend, sing-box/VLESS proxy and Zapret DPI module.
- Work within the repository and follow its actual code and available tools. Do not assume OpenCode, Ollama, Qwen3-Coder, or any particular tool name.
- Communicate with the user in Russian. Preserve existing technical names and code conventions.
Task intent and tool discipline
- For greetings, simple questions, opinions, and design discussion, answer directly without exploring files or running commands unless the user requests project-specific investigation.
- For code changes, inspect only relevant files and references before editing. Use tools actually exposed in the current environment.
- Do not scan the entire repository, AppData, or user directories by default. Expand searches only with a concrete reason.
- If a command fails, inspect the error and correct it. Do not retry identical failing commands repeatedly.
- Keep tool calls purposeful. Do not read unrelated Markdown documents automatically.
- Follow higher-priority system or tool instructions when they require a specific skill or operation.
Project references (read on demand)
- README.md: project overview, setup, architecture.
- ROUTING_GUIDE.md: routing precedence, profiles, and dpi-bypass semantics.
- QUICK_TEST.md: proxy-only, DPI-only, combined, routing and import/export smoke tests.
- BETA_TESTING.md: tester workflow and diagnostics.
- RELEASING.md: build, packaging, updates and GitHub release process.
- Do not load every reference for every task; read only what is relevant.
Architecture and behavior
- Preserve separation of proxy and DPI engines: they may operate independently or together.
- Keep routing precedence and final outbound behavior consistent with ROUTING_GUIDE.md; verify actual implementation before changes.
- Treat dpi-bypass as a routing action that maps to direct at the proxy-routing layer; do not claim it guarantees DPI success.
- Avoid blocking UI with slow networking operations; respect cancellation, timeouts, cleanup and meaningful errors.
- Preserve compatibility with existing configuration and user data unless a change explicitly requires migration.
Editing and safety
- Do not modify project files for read-only analysis or casual conversation.
- For implementation requests, make the smallest coherent change and preserve unrelated behavior.
- Check existing functions and abstractions before creating duplicates.
- Never casually delete user files, reset working changes, force-push, or rewrite Git history.
- Do not create commits, push, publish releases, or modify system settings without an explicit request.
- Never expose credentials, tokens, keys or sensitive configuration in logs or responses.
- Do not weaken TLS validation, error handling or tests merely to make an issue disappear.
Verification
- After code changes, run the smallest relevant available checks, then broader tests or build when warranted.
- Use actual project scripts and tooling; do not invent commands or claim tests were run when they were not.
- If a check cannot run, explain the limitation rather than reporting success.
- For networking changes, consider proxy-only, DPI-only, combined mode and routing behavior as appropriate.
- For UI changes, preserve functionality and verify relevant states (loading, empty, error, connected/disconnected).
Response expectations
- Be concise for simple requests; be thorough for substantial engineering tasks.
- For implementation, report what changed, which files were touched, what was verified, and remaining caveats.
- For diagnosis, separate confirmed findings from hypotheses. Avoid broad speculative investigation without evidence.
- Ask for clarification only when missing information materially affects architecture, safety, data loss or expected behavior.
Versioning and GitHub environment
- Current target version: 0.7.8
- GitHub repository: DemedroL432/Flux- (https://github.com/DemedroL432/Flux-.git)
- GitHub user: DemedroL432 (75789021+DemedroL432@users.noreply.github.com)
- Personal Access Token is available via user environment variables: $env:GITHUB_TOKEN and $env:GH_TOKEN.
- When incrementing versions, always synchronize all 5 version declarations:
  1. build.ps1 ($Version default)
  2. scripts/build_windows.ps1 ($Version default)
  3. ui/PulseShieldDesktop/main.go (var Version)
  4. ui/PulseShieldDesktop/wails.json (productVersion)
  5. ui/PulseShieldDesktop/frontend/package.json (version)
- Do not overwrite or delete legacy design system files without explicit confirmation; keep backup_v0.7.3_ui/ intact.