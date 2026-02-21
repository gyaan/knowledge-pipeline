# Contributing

Contributions are welcome. This project has zero external dependencies by design — please keep it that way.

## Getting started

```bash
git clone https://github.com/gyaan/knowledge-pipeline.git
cd knowledge-pipeline
cp .env_example .env
# Set ANTHROPIC_API_KEY or OPENAI_API_KEY in .env
go build ./...
go vet ./...
```

## Ground rules

- **No external dependencies.** Pure Go standard library only. No third-party packages.
- **In-memory only.** No external databases or disk persistence in the core pipeline.
- **Standard Go layout.** Keep `cmd/`, `internal/`, `pkg/` conventions.
- **Backwards compatible config.** New environment variables must have sensible defaults so existing `.env` files keep working.

## Making changes

1. Fork the repository and create a branch: `git checkout -b feature/your-feature`
2. Make your changes
3. Run checks before committing:
   ```bash
   go build ./...
   go vet ./...
   go test ./...
   ```
4. Manually test with the server:
   ```bash
   go run cmd/server/main.go
   curl http://localhost:8080/api/health
   ```
5. Open a pull request against `master`

## Adding a new LLM provider

1. Implement the `LLMClient` interface in `internal/llm/` (see `claude.go` or `openai.go` for the pattern)
2. Add the new provider to the switch statement in `cmd/server/main.go`
3. Document the new environment variables in `README.md` and `.env_example`

## Adding knowledge base content

Drop `.txt` files into any subdirectory of `knowledge_base/`. The loader picks up all files recursively. The `category` metadata is set from the top-level subdirectory name.

## Reporting issues

Use the [issue templates](https://github.com/gyaan/knowledge-pipeline/issues/new/choose) — they help us understand and reproduce the problem faster.

## License

By contributing, you agree that your contributions will be licensed under the [MIT License](LICENSE).
