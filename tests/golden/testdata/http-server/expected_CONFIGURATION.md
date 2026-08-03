# Configuration Reference

## Instrumentation

**Path:** `instrumentation`

### Http

**Path:** `instrumentation.http`

| Name | Type | Default | Required | Constraints | Description |
| --- | --- | --- | --- | --- | --- |
| `api_key` | string | - | No | Pattern: `^[A-Z0-9]{16}$` | API key for external service integration |
| `capture_headers` | array of strings | - | No | - | Headers to capture |
| `enabled` | boolean | `true` | No | - | Enables HTTP instrumentation |
| `host` | string | `127.0.0.1` | No | - | HTTP server bind address |
| `log_level` | string | `info` | No | Enum: [debug, info, warn, error] | - |
| `port` | integer | `8080` | No | Min: 1, Max: 65535 | - |
| `redact_query` | array of strings | - | No | - | Query parameters to redact from logs |
| `timeout` | integer | `30` | No | Min: 1 | Request timeout in seconds |

