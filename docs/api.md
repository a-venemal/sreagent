# SREAgent REST API Reference

> Auto-generated from source code. Last updated: 2026-04-04.

## Table of Contents

- [Conventions](#conventions)
- [Authentication](#1-authentication)
- [OIDC Single Sign-On](#2-oidc-single-sign-on)
- [DataSources](#3-datasources)
- [Alert Rules](#4-alert-rules)
- [Alert Events](#5-alert-events)
- [Mute Rules](#6-mute-rules)
- [Notify Rules (v2)](#7-notify-rules-v2)
- [Notify Media](#8-notify-media)
- [Message Templates](#9-message-templates)
- [Subscribe Rules](#10-subscribe-rules)
- [Business Groups](#11-business-groups)
- [Alert Channels](#12-alert-channels)
- [User Notify Configs](#13-user-notify-configs)
- [Notify Channels (v1)](#14-notify-channels-v1)
- [Notify Policies (v1)](#15-notify-policies-v1)
- [Users](#16-users)
- [Teams](#17-teams)
- [Schedules (On-Call)](#18-schedules-on-call)
- [Escalation Policies](#19-escalation-policies)
- [AI](#20-ai)
- [Lark Bot](#21-lark-bot)
- [Engine](#22-engine)
- [Dashboard](#23-dashboard)
- [Webhooks](#24-webhooks)
- [Alert Action Pages](#25-alert-action-pages)

---

## Conventions

### Base URL

All API routes are prefixed with `/api/v1` unless otherwise noted.

### Response Envelope

Every JSON endpoint returns a unified envelope:

```json
{
  "code": 0,
  "message": "ok",
  "data": { ... }
}
```

- `code = 0` — success
- `code != 0` — error (message field contains human-readable description)

### Pagination

Paginated list endpoints accept:

| Param | Type | Default | Range | Description |
|-------|------|---------|-------|-------------|
| `page` | int | 1 | >= 1 | Page number |
| `page_size` | int | 20 | 1–100 | Items per page |

Paginated responses wrap `data`:

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [ ... ],
    "total": 128,
    "page": 1,
    "page_size": 20
  }
}
```

### Authentication

Protected routes require a JWT token in the `Authorization` header:

```
Authorization: Bearer <token>
```

Tokens are obtained via `POST /api/v1/auth/login` or the OIDC callback flow.

### RBAC Roles

Five roles, in descending privilege order:

| Role | Description |
|------|-------------|
| `admin` | Full access to all resources |
| `team_lead` | Manage config objects (rules, channels, schedules, teams) |
| `member` | Operational actions (acknowledge, resolve, subscribe) |
| `viewer` | Read-only access to assigned resources |
| `global_viewer` | Read-only access to all resources |

Route access levels referenced below:
- **Public** — no authentication required
- **Any** — any authenticated user
- **Operate** — `admin`, `team_lead`, or `member`
- **Manage** — `admin` or `team_lead`
- **Admin** — `admin` only

### Common Model Fields

All entities include:

| Field | Type | Description |
|-------|------|-------------|
| `id` | uint | Auto-increment primary key |
| `created_at` | datetime | ISO 8601 |
| `updated_at` | datetime | ISO 8601 |

---

## 1. Authentication

### POST `/api/v1/auth/login` — Login

**Access:** Public

**Request:**

```json
{
  "username": "admin",
  "password": "secret123"
}
```

**Response:**

```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOi...",
    "expires_in": 86400
  }
}
```

### GET `/api/v1/auth/profile` — Get Current User

**Access:** Any

**Response:** User object (see [Users](#16-users) model). Password is never included.

### PUT `/api/v1/me/profile` — Update Own Profile

**Access:** Any

| Field | Type | Description |
|-------|------|-------------|
| `display_name` | string | |
| `email` | string | |
| `phone` | string | |
| `avatar` | string | Base64 data URL or preset key |

### POST `/api/v1/me/password` — Change Own Password

**Access:** Any

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `old_password` | string | yes | |
| `new_password` | string | yes | min 6 chars |

---

## 2. OIDC Single Sign-On

### GET `/api/v1/auth/oidc/config` — OIDC Status

**Access:** Public

**Response:**

```json
{
  "code": 0,
  "data": {
    "enabled": true,
    "login_url": "/api/v1/auth/oidc/login"
  }
}
```

Returns `{"enabled": false}` when OIDC is not configured.

### GET `/api/v1/auth/oidc/login` — Initiate OIDC Login

**Access:** Public

Redirects (302) to the configured IdP authorization endpoint. Sets an `oidc_state` cookie for CSRF protection.

### GET `/api/v1/auth/oidc/callback` — OIDC Callback

**Access:** Public

Called by the IdP after authentication. On success, redirects the browser to:

```
/?oidc_token=<jwt>&expires_in=<seconds>
```

The frontend router guard intercepts the `oidc_token` query parameter and stores it.

**Query Parameters:**

| Param | Description |
|-------|-------------|
| `code` | Authorization code from IdP |
| `state` | CSRF state (validated against cookie) |
| `error` | Error code (optional) |
| `error_description` | Error details (optional) |

### POST `/api/v1/auth/oidc/token` — Exchange Code for Token (JSON)

**Access:** Public

For SPA clients that prefer a JSON flow instead of redirects.

**Request:**

```json
{ "code": "abc123" }
```

**Response:** Same as login response (`token`, `expires_in`).

---

## 3. DataSources

Manage Prometheus, VictoriaMetrics, VictoriaLogs, and Zabbix data sources.

**Model fields:** `name`, `type` (prometheus | victoriametrics | zabbix | victorialogs), `endpoint`, `description`, `labels` (map), `status` (healthy | unhealthy | unknown), `auth_type` (none | basic | bearer | api_key), `auth_config` (JSON), `health_check_interval`, `is_enabled`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/datasources` | Any | List (paginated). Filter: `?type=prometheus` |
| GET | `/datasources/:id` | Any | Get by ID |
| POST | `/datasources` | Admin | Create |
| PUT | `/datasources/:id` | Admin | Update |
| DELETE | `/datasources/:id` | Admin | Delete |
| POST | `/datasources/:id/health-check` | Manage | Trigger health check |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Unique name |
| `type` | string | yes | One of the supported types |
| `endpoint` | string | yes | URL |
| `description` | string | no | |
| `labels` | map[string]string | no | Key-value metadata |
| `auth_type` | string | no | Default: `none` |
| `auth_config` | string (JSON) | no | Auth-specific config |
| `health_check_interval` | int | no | Seconds |

**Health Check Response:**

```json
{ "code": 0, "data": { "status": "healthy" } }
```

---

## 4. Alert Rules

Define evaluation rules using PromQL, LogsQL, or other query expressions.

**Model fields:** `name`, `display_name`, `description`, `datasource_id`, `expression`, `for_duration`, `severity` (critical | warning | info), `labels` (map), `annotations` (map), `status` (enabled | disabled | muted), `group_name`, `version`, `eval_interval`, `recovery_hold`, `nodata_enabled`, `nodata_duration`, `suppress_enabled`, `biz_group_id`, `created_by`, `updated_by`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/alert-rules` | Any | List (paginated). Filters: `?severity=critical&status=enabled&group_name=infra` |
| GET | `/alert-rules/:id` | Any | Get by ID |
| GET | `/alert-rules/export` | Any | Export as YAML. Filter: `?group_name=infra` |
| POST | `/alert-rules` | Manage | Create |
| PUT | `/alert-rules/:id` | Manage | Update |
| DELETE | `/alert-rules/:id` | Manage | Delete |
| PATCH | `/alert-rules/:id/status` | Manage | Toggle status |
| POST | `/alert-rules/import` | Manage | Import from YAML/JSON file |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | Rule identifier |
| `display_name` | string | no | Human-readable name |
| `description` | string | no | |
| `datasource_id` | uint | yes | FK to datasource |
| `expression` | string | yes | PromQL / LogsQL expression |
| `for_duration` | string | no | e.g. `"5m"` |
| `severity` | string | yes | critical, warning, info |
| `labels` | map[string]string | no | Additional labels |
| `annotations` | map[string]string | no | Annotations (summary, description) |
| `group_name` | string | no | Rule group |
| `eval_interval` | int | no | Evaluation interval in seconds |
| `recovery_hold` | int | no | Hold before auto-resolve (seconds) |
| `nodata_enabled` | bool | no | Fire on no-data |
| `nodata_duration` | int | no | No-data threshold (seconds) |
| `suppress_enabled` | bool | no | Enable level-based suppression |
| `biz_group_id` | *uint | no | Business group affiliation |

**Toggle Status Body:**

```json
{ "status": "enabled" }
```

**Import** — `multipart/form-data`:

| Field | Type | Description |
|-------|------|-------------|
| `file` | file | `.yaml` / `.yml` / `.json` (Prometheus rule file format) |
| `datasource_id` | string | Default datasource for imported rules |

**Import Response:**

```json
{ "code": 0, "data": { "total": 10, "success": 9, "failed": 1, "errors": ["..."] } }
```

**Export** — Returns `application/x-yaml` Content-Type with `Content-Disposition: attachment`.

---

## 5. Alert Events

Live and historical alert instances generated by the evaluation engine or received via webhooks.

**Model fields:** `fingerprint`, `rule_id`, `alert_name`, `severity`, `status` (firing | acknowledged | assigned | silenced | resolved | closed), `labels` (map), `annotations` (map), `source`, `generator_url`, `fired_at`, `acked_at`, `resolved_at`, `closed_at`, `acked_by`, `assigned_to`, `silenced_until`, `silence_reason`, `resolution`, `fire_count`, `oncall_user_id`, `is_dispatched`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/alert-events` | Any | List (paginated). Filters: `?status=firing&severity=critical&view_mode=mine` |
| GET | `/alert-events/:id` | Any | Get by ID |
| GET | `/alert-events/:id/timeline` | Any | Get event timeline (state changes, comments) |
| POST | `/alert-events/:id/acknowledge` | Operate | Acknowledge alert |
| POST | `/alert-events/:id/assign` | Operate | Assign to user |
| POST | `/alert-events/:id/resolve` | Operate | Resolve alert |
| POST | `/alert-events/:id/close` | Operate | Close alert |
| POST | `/alert-events/:id/comment` | Operate | Add comment |
| POST | `/alert-events/:id/silence` | Operate | Silence alert |
| POST | `/alert-events/batch/acknowledge` | Operate | Batch acknowledge |
| POST | `/alert-events/batch/close` | Operate | Batch close |

**List Filters:**

| Param | Type | Description |
|-------|------|-------------|
| `status` | string | firing, acknowledged, assigned, silenced, resolved, closed |
| `severity` | string | critical, warning, info |
| `view_mode` | string | `mine` (assigned to me), `unassigned`, `all` (default) |
| `user_id` | uint | Admin override for view_mode=mine |

**Assign Body:**

```json
{ "assign_to": 5, "note": "Please investigate" }
```

**Resolve Body:**

```json
{ "resolution": "Fixed the root cause by scaling the service" }
```

**Close Body:**

```json
{ "note": "False positive" }
```

**Comment Body:**

```json
{ "note": "Investigating the issue now" }
```

**Silence Body:**

```json
{ "duration_minutes": 60, "reason": "Maintenance window" }
```

**Batch Acknowledge / Close Body:**

```json
{ "ids": [1, 2, 3] }
```

**Batch Response:**

```json
{ "code": 0, "data": { "success": 3, "failed": 0 } }
```

---

## 6. Mute Rules

Suppress notifications for alerts matching specified criteria during defined time windows.

**Model fields:** `name`, `description`, `match_labels` (map), `severities` (comma-separated), `start_time`, `end_time`, `periodic_start`, `periodic_end`, `days_of_week`, `timezone`, `is_enabled`, `rule_ids` (comma-separated), `created_by`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/mute-rules` | Any | List (paginated) |
| GET | `/mute-rules/:id` | Any | Get by ID |
| POST | `/mute-rules` | Manage | Create |
| PUT | `/mute-rules/:id` | Manage | Update |
| DELETE | `/mute-rules/:id` | Manage | Delete |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `match_labels` | map[string]string | no | Label matchers |
| `severities` | string | no | e.g. `"critical,warning"` |
| `start_time` | datetime | no | One-time window start (ISO 8601) |
| `end_time` | datetime | no | One-time window end |
| `periodic_start` | string | no | Daily start, e.g. `"02:00"` |
| `periodic_end` | string | no | Daily end, e.g. `"06:00"` |
| `days_of_week` | string | no | e.g. `"1,2,3,4,5"` (Mon=1) |
| `timezone` | string | no | Default: `"Asia/Shanghai"` |
| `is_enabled` | bool | no | |
| `rule_ids` | string | no | Comma-separated alert rule IDs |

---

## 7. Notify Rules (v2)

Advanced notification rules with pipeline processing and per-rule notify configurations.

**Model fields:** `name`, `description`, `is_enabled`, `severities`, `match_labels` (map), `pipeline` (JSON), `notify_configs` (JSON), `repeat_interval`, `callback_url`, `created_by`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/notify-rules` | Any | List (paginated) |
| GET | `/notify-rules/:id` | Any | Get by ID |
| POST | `/notify-rules` | Manage | Create |
| PUT | `/notify-rules/:id` | Manage | Update |
| DELETE | `/notify-rules/:id` | Manage | Delete |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `is_enabled` | bool | no | |
| `severities` | string | no | Comma-separated |
| `match_labels` | map[string]string | no | Label matchers |
| `pipeline` | string (JSON) | no | Processing pipeline steps |
| `notify_configs` | string (JSON) | no | Notification config array |
| `repeat_interval` | int | no | Seconds between repeat notifications |
| `callback_url` | string | no | Webhook callback URL |

---

## 8. Notify Media

Notification media (delivery channels): Lark webhook, email, HTTP webhook, script.

**Model fields:** `name`, `type` (lark_webhook | email | http | script), `description`, `is_enabled`, `config` (JSON), `variables` (JSON), `is_builtin`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/notify-media` | Any | List (paginated) |
| GET | `/notify-media/:id` | Any | Get by ID |
| POST | `/notify-media` | Manage | Create |
| PUT | `/notify-media/:id` | Manage | Update |
| DELETE | `/notify-media/:id` | Manage | Delete |
| POST | `/notify-media/:id/test` | Manage | Send test notification |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `type` | string | yes | One of: lark_webhook, email, http, script |
| `description` | string | no | |
| `is_enabled` | bool | no | |
| `config` | string (JSON) | yes | Type-specific config |
| `variables` | string (JSON) | no | Template variables |

**Test Response:**

```json
{ "code": 0, "data": { "message": "test notification sent" } }
```

---

## 9. Message Templates

Go `text/template` based message templates for notifications.

**Model fields:** `name`, `description`, `content` (Go template string), `type` (text | html | markdown | lark_card), `is_builtin`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/message-templates` | Any | List (paginated) |
| GET | `/message-templates/:id` | Any | Get by ID |
| POST | `/message-templates` | Manage | Create |
| PUT | `/message-templates/:id` | Manage | Update |
| DELETE | `/message-templates/:id` | Manage | Delete |
| POST | `/message-templates/preview` | Any | Preview rendered template |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `content` | string | yes | Go template string |
| `type` | string | no | Default: `"text"` |

**Preview Body:**

```json
{ "content": "Alert {{ .AlertName }} is {{ .Status }}" }
```

**Preview Response:**

```json
{ "code": 0, "data": { "rendered": "Alert CPUHigh is firing" } }
```

---

## 10. Subscribe Rules

Allow users/teams to subscribe to alerts matching certain criteria and route them to a notify rule.

**Model fields:** `name`, `description`, `is_enabled`, `match_labels` (map), `severities`, `notify_rule_id`, `user_id`, `team_id`, `created_by`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/subscribe-rules` | Any | List (paginated) |
| GET | `/subscribe-rules/:id` | Any | Get by ID |
| POST | `/subscribe-rules` | Operate | Create |
| PUT | `/subscribe-rules/:id` | Operate | Update |
| DELETE | `/subscribe-rules/:id` | Operate | Delete |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `is_enabled` | bool | no | |
| `match_labels` | map[string]string | no | Label matchers |
| `severities` | string | no | Comma-separated |
| `notify_rule_id` | uint | yes | Target notify rule |
| `user_id` | uint | no | Subscribe for a specific user |
| `team_id` | uint | no | Subscribe for a team |

---

## 11. Business Groups

Hierarchical business group tree for organizing alert rules and access control.

**Model fields:** `name` (supports `/` for hierarchy), `description`, `parent_id`, `labels` (map), `members`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/biz-groups` | Any | List (paginated) |
| GET | `/biz-groups/tree` | Any | Get tree structure |
| GET | `/biz-groups/:id` | Any | Get by ID |
| GET | `/biz-groups/:id/members` | Any | List group members |
| POST | `/biz-groups` | Manage | Create |
| PUT | `/biz-groups/:id` | Manage | Update |
| DELETE | `/biz-groups/:id` | Manage | Delete |
| POST | `/biz-groups/:id/members` | Manage | Add member |
| DELETE | `/biz-groups/:id/members/:uid` | Manage | Remove member |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `parent_id` | uint | no | Parent group ID |
| `labels` | map[string]string | no | |

**Add Member Body:**

```json
{ "user_id": 5, "role": "admin" }
```

Role can be `"admin"` or `"member"`.

---

## 12. Alert Channels

Virtual alert routing channels that bind a notify media to optional template and label matchers.

**Model fields:** `name`, `description`, `match_labels` (map), `severities`, `media_id`, `template_id`, `throttle_min`, `is_enabled`, `created_by`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/alert-channels` | Any | List (paginated) |
| GET | `/alert-channels/:id` | Any | Get by ID |
| POST | `/alert-channels` | Manage | Create |
| PUT | `/alert-channels/:id` | Manage | Update |
| DELETE | `/alert-channels/:id` | Manage | Delete |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `match_labels` | map[string]string | no | |
| `severities` | string | no | Comma-separated |
| `media_id` | uint | yes | FK to notify media |
| `template_id` | uint | no | FK to message template |
| `throttle_min` | int | no | Minimum minutes between notifications |
| `is_enabled` | bool | no | |

---

## 13. User Notify Configs

Personal notification preferences for the current user (multi-media).

**Model fields:** `user_id`, `media_type` (lark_personal | email | webhook), `config` (JSON), `is_enabled`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/me/notify-configs` | Any | List current user's configs |
| PUT | `/me/notify-configs` | Any | Create or update (upsert by media_type) |
| DELETE | `/me/notify-configs/:mediaType` | Any | Delete by media type |

**Upsert Body:**

```json
{
  "media_type": "email",
  "config": "{\"address\": \"user@example.com\"}",
  "is_enabled": true
}
```

**Delete Path Param:** `:mediaType` — e.g., `email`, `lark_personal`, `webhook`.

---

## 14. Notify Channels (v1)

Legacy notification channels. Types: `lark_webhook`, `lark_bot`, `email`, `sms`, `custom_webhook`.

**Model fields:** `name`, `type`, `description`, `labels` (map), `config` (JSON, hidden on GET), `is_enabled`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/notify-channels` | Any | List (paginated) |
| GET | `/notify-channels/:id` | Any | Get by ID |
| POST | `/notify-channels` | Manage | Create |
| PUT | `/notify-channels/:id` | Manage | Update |
| DELETE | `/notify-channels/:id` | Manage | Delete |
| POST | `/notify-channels/:id/test` | Manage | Send test notification |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `type` | string | yes | Channel type |
| `description` | string | no | |
| `labels` | map[string]string | no | |
| `config` | string (JSON) | yes | Type-specific config |
| `is_enabled` | bool | no | |

---

## 15. Notify Policies (v1)

Legacy notification policies that route alerts to channels based on label matching.

**Model fields:** `name`, `description`, `match_labels` (map), `severities`, `channel_id`, `throttle_minutes`, `template_name`, `is_enabled`, `priority`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/notify-policies` | Any | List (paginated) |
| GET | `/notify-policies/:id` | Any | Get by ID |
| POST | `/notify-policies` | Manage | Create |
| PUT | `/notify-policies/:id` | Manage | Update |
| DELETE | `/notify-policies/:id` | Manage | Delete |

**Create / Update Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | yes | |
| `description` | string | no | |
| `match_labels` | map[string]string | yes | Label matchers |
| `severities` | string | no | Comma-separated |
| `channel_id` | uint | yes | FK to notify channel |
| `throttle_minutes` | int | no | Throttle window |
| `template_name` | string | no | Default: `"default"` |
| `is_enabled` | bool | no | |
| `priority` | int | no | Lower = higher priority |

---

## 16. Users

User management. Supports human users, bot users, and channel (virtual) users.

**Model fields:** `username`, `display_name`, `email`, `phone`, `lark_user_id`, `avatar`, `role` (admin | team_lead | member | viewer | global_viewer), `is_active`, `user_type` (human | bot | channel), `notify_target` (JSON), `oidc_subject`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/users` | Any | List (paginated). Filter: `?user_type=human` |
| GET | `/users/:id` | Any | Get by ID |
| POST | `/users` | Admin | Create human user |
| POST | `/users/virtual` | Admin | Create virtual (bot/channel) user |
| PUT | `/users/:id` | Admin | Update user |
| PATCH | `/users/:id/active` | Admin | Enable / disable user |
| PATCH | `/users/:id/password` | Admin | Admin reset password |
| DELETE | `/users/:id` | Admin | Delete user |

**Create Human User Body:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `username` | string | yes | Unique |
| `password` | string | yes | min 6 chars |
| `display_name` | string | no | |
| `email` | string | no | email format |
| `phone` | string | no | |
| `lark_user_id` | string | no | |
| `avatar` | string | no | |
| `role` | string | no | Default: `"member"` |

**Create Virtual User Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `username` | string | yes | |
| `display_name` | string | no | |
| `user_type` | string | yes | `"bot"` or `"channel"` |
| `notify_target` | string | no | JSON notification target config |
| `description` | string | no | |
| `role` | string | no | |

**Update Body:**

| Field | Type | Description |
|-------|------|-------------|
| `display_name` | string | |
| `email` | string | |
| `phone` | string | |
| `lark_user_id` | string | |
| `avatar` | string | |
| `role` | string | |

**Toggle Active Body:**

```json
{ "is_active": true }
```

**Change Password Body:**

| Field | Type | Required | Validation |
|-------|------|----------|------------|
| `old_password` | string | yes | |
| `new_password` | string | yes | min 6 chars |

---

## 17. Teams

Team management with member roles.

**Model fields:** `name`, `description`, `labels` (map). Members via join table with `role` (lead | member).

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/teams` | Any | List (paginated) |
| GET | `/teams/:id` | Any | Get by ID |
| GET | `/teams/:id/members` | Any | List team members |
| POST | `/teams` | Manage | Create |
| PUT | `/teams/:id` | Manage | Update |
| DELETE | `/teams/:id` | Manage | Delete |
| POST | `/teams/:id/members` | Manage | Add member |
| DELETE | `/teams/:id/members/:uid` | Manage | Remove member |

**Create / Update Body:**

| Field | Type | Required |
|-------|------|----------|
| `name` | string | yes |
| `description` | string | no |
| `labels` | map[string]string | no |

**Add Member Body:**

```json
{ "user_id": 5, "role": "lead" }
```

Role: `"lead"` or `"member"`.

---

## 18. Schedules (On-Call)

On-call schedule management with rotation, shifts, overrides, and participants.

### Schedule CRUD

**Model fields:** `name`, `team_id`, `description`, `rotation_type` (daily | weekly | custom), `timezone`, `handoff_time`, `handoff_day`, `is_enabled`, `severity_filter`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/schedules` | Any | List (paginated). Filter: `?team_id=1` |
| GET | `/schedules/:id` | Any | Get by ID |
| GET | `/schedules/:id/oncall` | Any | Get current on-call user |
| POST | `/schedules` | Manage | Create |
| PUT | `/schedules/:id` | Manage | Update |
| DELETE | `/schedules/:id` | Manage | Delete |
| PUT | `/schedules/:id/participants` | Manage | Set rotation participants |
| POST | `/schedules/:id/overrides` | Manage | Create override |
| DELETE | `/schedules/:id/overrides/:oid` | Manage | Delete override |

**Create / Update Body:**

| Field | Type | Required | Default |
|-------|------|----------|---------|
| `name` | string | yes | |
| `team_id` | uint | no | |
| `description` | string | no | |
| `rotation_type` | string | yes | |
| `timezone` | string | no | `"Asia/Shanghai"` |
| `handoff_time` | string | no | `"09:00"` |
| `handoff_day` | int | no | |
| `is_enabled` | bool | no | true |

**Set Participants Body:**

```json
{ "user_ids": [1, 2, 3] }
```

**Create Override Body:**

```json
{
  "user_id": 5,
  "start_time": "2026-04-05T00:00:00Z",
  "end_time": "2026-04-06T00:00:00Z",
  "reason": "Coverage swap"
}
```

### On-Call Shifts

**Model fields:** `schedule_id`, `user_id`, `start_time`, `end_time`, `severity_filter`, `source` (manual | rotation), `note`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/schedules/:id/shifts` | Any | List shifts. Filters: `?start=<RFC3339>&end=<RFC3339>` |
| POST | `/schedules/:id/shifts` | Manage | Create shift |
| PUT | `/schedules/:id/shifts/:shiftId` | Manage | Update shift |
| DELETE | `/schedules/:id/shifts/:shiftId` | Manage | Delete shift |
| POST | `/schedules/:id/generate-shifts` | Manage | Auto-generate shifts from rotation |

**Create / Update Shift Body:**

| Field | Type | Required |
|-------|------|----------|
| `user_id` | uint | yes |
| `start_time` | datetime (RFC 3339) | yes |
| `end_time` | datetime (RFC 3339) | yes |
| `severity_filter` | string | no |
| `note` | string | no |

**Generate Shifts Body:**

```json
{ "weeks": 4 }
```

Validation: 1–52 weeks.

**Generate Response:**

```json
{ "code": 0, "data": { "message": "shifts generated", "weeks": 4 } }
```

---

## 19. Escalation Policies

Multi-step escalation policies that define notification targets and delay intervals.

### Policy CRUD

**Model fields:** `name`, `team_id`, `is_enabled`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| GET | `/escalation-policies` | Any | List. Filter: `?team_id=1` |
| GET | `/escalation-policies/:id` | Any | Get policy with steps |
| POST | `/escalation-policies` | Manage | Create |
| PUT | `/escalation-policies/:id` | Manage | Update |
| DELETE | `/escalation-policies/:id` | Manage | Delete |

**Create / Update Body:**

```json
{ "name": "Critical P1", "team_id": 1, "is_enabled": true }
```

**Get Response:**

```json
{
  "code": 0,
  "data": {
    "policy": { "id": 1, "name": "Critical P1", "team_id": 1, "is_enabled": true },
    "steps": [ { "id": 1, "step_order": 1, "delay_minutes": 0, "target_type": "schedule", "target_id": 1 }, ... ]
  }
}
```

### Escalation Steps

**Model fields:** `policy_id`, `step_order`, `delay_minutes`, `target_type` (user | schedule | team), `target_id`, `notify_channel_id`.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| POST | `/escalation-policies/:id/steps` | Manage | Create step |
| PUT | `/escalation-policies/:id/steps/:stepId` | Manage | Update step |
| DELETE | `/escalation-policies/:id/steps/:stepId` | Manage | Delete step |

**Create / Update Step Body:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `step_order` | int | yes | Execution order (1-based) |
| `delay_minutes` | int | yes | Delay before this step fires |
| `target_type` | string | yes | user, schedule, or team |
| `target_id` | uint | yes | FK to target entity |
| `notify_channel_id` | uint | no | Override default channel |

---

## 20. AI

AI-powered alert analysis. Supports LLM-generated alert reports and SOP suggestions.

| Method | Route | Access | Description |
|--------|-------|--------|-------------|
| POST | `/ai/alert-report` | Any | Generate AI alert analysis report |
| POST | `/ai/suggest-sop` | Any | AI-suggested SOP for an alert |
| POST | `/ai/test` | Manage | Test AI provider connectivity |
| GET | `/ai/config` | Admin | Get AI configuration (API key masked) |
| PUT | `/ai/config` | Admin | Update AI configuration |

**Generate Report / Suggest SOP Body:**

```json
{ "event_id": 42 }
```

**Report Response:**

```json
{ "code": 0, "data": { "report": "## Analysis\n...", "event_id": 42 } }
```

**SOP Response:**

```json
{ "code": 0, "data": { "sop": "1. Check CPU usage\n2. ...", "event_id": 42 } }
```

**Test Response:**

```json
{ "code": 0, "data": { "message": "AI connection successful" } }
```

---

## 21. Lark Bot

Lark (Feishu) bot integration for interactive alert notifications.

### POST `/lark/event` — Lark Event Callback

**Access:** Public (verified by Lark verification token)

Receives Lark event subscription callbacks including URL verification challenges and message events. Returns raw JSON for Lark protocol compatibility.

### GET `/api/v1/lark/bot/config` — Get Lark Config

**Access:** Admin

Returns Lark bot configuration (app ID, webhook URL, etc.).

### PUT `/api/v1/lark/bot/config` — Update Lark Config

**Access:** Admin

Updates Lark bot configuration. Sensitive fields (app_secret, verification_token, encrypt_key) are stored with AES-GCM encryption.

---

## 22. Engine

### GET `/api/v1/engine/status` — Engine Status

**Access:** Any

Returns the alert evaluation engine status including active rule count, evaluation metrics, and state store connectivity.

---

## 23. Dashboard

### GET `/api/v1/dashboard/stats` — Dashboard Statistics

**Access:** Any

**Response:**

```json
{
  "code": 0,
  "data": {
    "total_datasources": 3,
    "total_rules": 45,
    "active_alerts": 12,
    "resolved_today": 8,
    "total_users": 20,
    "total_teams": 4
  }
}
```

---

## 24. Webhooks

### POST `/webhooks/alertmanager` — AlertManager Webhook

**Access:** Public (authenticated by shared secret or source IP at network level)

Receives alert payloads in [Alertmanager webhook format](https://prometheus.io/docs/alerting/latest/configuration/#webhook_config).

**Request Body:**

```json
{
  "version": "4",
  "status": "firing",
  "receiver": "sreagent",
  "alerts": [
    {
      "status": "firing",
      "labels": { "alertname": "HighCPU", "severity": "critical", "instance": "node1:9090" },
      "annotations": { "summary": "CPU usage above 90%", "description": "..." },
      "startsAt": "2026-04-04T10:00:00Z",
      "endsAt": "0001-01-01T00:00:00Z",
      "generatorURL": "http://prometheus:9090/graph?...",
      "fingerprint": "abc123"
    }
  ],
  "groupLabels": { "alertname": "HighCPU" },
  "commonLabels": { "alertname": "HighCPU", "severity": "critical" },
  "commonAnnotations": { "summary": "CPU usage above 90%" },
  "externalURL": "http://alertmanager:9093"
}
```

---

## 25. Alert Action Pages

Token-authenticated HTML pages linked from Lark notification cards. Allow one-click alert operations without requiring the full UI.

### GET `/alert-action/:token` — Render Action Page

**Access:** Token in URL path (JWT)

| Query Param | Description |
|-------------|-------------|
| `action` | Pre-select action: acknowledge, silence, resolve, close |
| `duration` | Pre-fill silence duration in minutes |

Returns an HTML page with an action form.

### POST `/alert-action/:token` — Execute Action

**Access:** Token in URL path (JWT)

**Form Fields** (`application/x-www-form-urlencoded`):

| Field | Type | Description |
|-------|------|-------------|
| `action` | string | acknowledge, silence, resolve, close |
| `operator_name` | string | Who performed the action |
| `note` | string | Optional note |
| `duration` | string | Silence duration in minutes (for silence action) |

Returns an HTML result page (success or error).

---

## Route Summary

| Category | Count | Access Level |
|----------|-------|-------------|
| Public (no auth) | 10 | Health, login, OIDC, webhooks, Lark callback, action pages |
| Read-only (any authenticated) | 36 | All GET/list endpoints |
| Operational (member+) | 12 | Alert actions, subscribe rules |
| Management (team_lead+) | 35 | Config CRUD, channels, rules, schedules, teams |
| Admin-only | 10 | User CRUD, system settings, AI/Lark config |
| **Total** | **~87** | |
