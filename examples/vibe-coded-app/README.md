# vibe-coded-app

**Intentionally vulnerable. Do not deploy, and do not reuse any of this code.**

A small app written the way AI coding assistants tend to write one. It is the reference target for
vibescan: the demo recordings scan it, and the [roadmap](../../ROADMAP.md) exit tests are defined
against it.

Every credential in this folder is fake.

## Planted issues

| Vibe Top 10 | Issue | Where |
|---|---|---|
| V01 Secrets shipped to the client | Supabase service-role key behind `NEXT_PUBLIC_`, with a hardcoded fallback | `src/lib/supabase.js` |
| V02 Missing authorization | Unauthenticated note lookup by ID | `server.js`, `GET /api/notes/:id` |
| V03 Open data layer | Table without RLS; policy of `using (true)` | `supabase/migrations/0001_init.sql` |
| V05 String-built injection | SQL injection, command injection, SSRF, reflected XSS | `server.js` |
| V06 Placeholder security | Hardcoded JWT secret, database password, `admin`/`admin`, wildcard CORS | `server.js` |
| V08 Missing abuse controls | No rate limiting on login | `server.js`, `POST /api/login` |
| V09 Unsafe LLM integration | User input in the system prompt; model output passed to `eval` | `server.js`, `POST /api/summarize` |

V04 (hallucinated dependencies) is deliberately not planted here. Listing a package name that does
not exist would invite someone to register it. Those fixtures use a mock registry instead.

V10 (poisoned agent context) fixtures will live in their own folder so that no agent working in
this repository loads them by accident.
