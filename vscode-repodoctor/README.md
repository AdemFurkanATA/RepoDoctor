# VS Code RepoDoctor Extension (Skeleton)

This directory is an isolated extension skeleton for RD-15004.

## Scope

- Runs `repodoctor analyze -path <workspace> -format json`.
- Parses JSON output and publishes diagnostics to Problems panel.
- Shows install guidance when `repodoctor` binary is missing.

## Architecture Boundary

- The extension is a separate TypeScript project under `vscode-repodoctor/`.
- No Go core module/runtime dependencies are introduced.
- Core CLI output remains authoritative; extension consumes CLI JSON only.

## Local Development

```bash
cd vscode-repodoctor
npm install
npm run compile
```

Then open this repository in VS Code and run Extension Host from the debugger.
