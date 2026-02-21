# Security Policy

## Supported versions

Only the latest commit on `master` is actively maintained.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security vulnerabilities.

Instead, report them privately via [GitHub's private vulnerability reporting](https://github.com/gyaan/knowledge-pipeline/security/advisories/new).

Include:
- A description of the vulnerability and its potential impact
- Steps to reproduce or a proof-of-concept
- Any suggested mitigations

You can expect an acknowledgement within 72 hours and a resolution or status update within 14 days.

## Security considerations

This project is intended for self-hosted deployments. Keep the following in mind:

- **API keys** — Store `ANTHROPIC_API_KEY` and `OPENAI_API_KEY` in `.env` (gitignored). Never commit keys to source control.
- **CORS** — The server currently allows all origins (`Access-Control-Allow-Origin: *`). Restrict this in production if the API should not be publicly accessible.
- **No authentication** — The `/api/chat` endpoint has no built-in auth. Add an API gateway, reverse proxy, or middleware with authentication before exposing it to the internet.
- **Knowledge base** — Only serve content you trust. The RAG pipeline passes document content directly into LLM prompts.
