# API rules

The goal is a small, generic API that many clients can reuse, not an endpoint per screen or use case.

## Public and internal API

- The public API listens on `:8080` and is the only port exposed through the Ingress. It serves `/api/v1/...` and the frontend.
- The internal API listens on `:8081` behind a ClusterIP Service only. It serves `/health`, `/ready` and `/internal/v1/...`. The cluster network (flannel) has no NetworkPolicy, so the only protection is that this port is never exposed on the public load balancer. Treat every in-cluster caller as trusted and keep destructive operations idempotent.
- Both share the domain services but have their own handlers and response types. When shared behaviour changes (validation, defaults, lifecycle), keep both contracts working.
- The public API exposes resources. Operations such as jobs and admin actions belong to the internal API.

## Resources

- Paths are plural nouns: `/servers`, `/servers/{id}`. No verbs in paths of the public API.
- Every resource has a stable, opaque `id` (UUID). Names are display data and may repeat across history.
- Use HTTP methods for their meaning: `GET` reads, `POST` on a collection creates, `PATCH` changes part of a resource, `DELETE` on `/{id}` removes one resource. Never overload a method with unrelated behaviour.
- A resource has one representation, returned the same way by `POST`, `GET /{id}` and list items.
- Before adding an endpoint, check whether a filter, a field or a sub-resource on an existing one covers the need.

## Data

- Return raw data and let clients format it: RFC 3339 timestamps (`createdAt`, `expiresAt`), enums for states, numbers with the unit in the field name (`durationMs`).
- Field names are camelCase, acronyms included (`ip`, not `IP`).
- Do not compute values that depend on the time of the request, like remaining time.
- Values the client has to choose from (tiers, presets, options) come from the API, not from constants copied into the frontend.

## Collections

- Lists support filtering, sorting and pagination through query params: `?status=running&sort=-createdAt&limit=20&cursor=...`.
- `limit` has a default and a maximum, and every sort ends with `id` as a tie-breaker so pages never skip or repeat items.
- Lists return an object, `{"items": [...], "nextCursor": "..."}`, so metadata can be added without breaking clients.
- Aggregates live next to the collection they describe (`/servers/stats`) and take the same filters, plus `groupBy` when needed.

## Errors

- Errors use `application/problem+json` (RFC 9457).
- Use the right status: 400 malformed request, 404 missing resource, 409 conflict (e.g. name in use), 422 failed validation, 500 only for unexpected failures.
- Never return internal error messages (Kubernetes, database) to clients. Log them instead.

## Versioning

- Paths are versioned (`/api/v1`, `/internal/v1`).
- A change is breaking when a request or client behaviour that works today stops working: removing or renaming fields, changing defaults, status codes or the meaning of a field.
- Adding response fields, optional params and enum values is allowed, so clients must ignore unknown fields and handle unknown enum values.
- Breaking changes need a new version, or a migration of every client first. Consumers outside this repo, like the cleanup CronJob in the deployments repo, keep their old route until they have been switched over.
