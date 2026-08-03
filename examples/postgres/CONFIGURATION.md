# Configuration Reference

## Postgres

**Path:** `postgres`

| Name | Type | Default | Required | Constraints | Description |
| --- | --- | --- | --- | --- | --- |
| `conn_max_lifetime` | integer | `300` | No | - | Maximum duration in seconds a connection survives |
| `database` | string | - | Yes | - | Target database name |
| `host` | string | `localhost` | No | - | Postgres server hostname or IP address |
| `max_idle_conns` | integer | `5` | No | Min: 1 | Maximum idle database connections in pool |
| `max_open_conns` | integer | `25` | No | Min: 1 | Maximum open database connections |
| `password` | string | - | Yes | - | Database authentication password |
| `port` | integer | `5432` | No | Min: 1, Max: 65535 | Postgres server port number |
| `retry_attempts` | integer | `3` | No | Min: 0, Max: 10 | Number of reconnection retry attempts |
| `retry_backoff_ms` | integer | `500` | No | Min: 0 | Delay in milliseconds between retry attempts |
| `ssl_mode` | string | `disable` | No | Enum: [disable, require, verify-ca, verify-full] | SSL connection security mode |
| `user` | string | - | Yes | - | Database user name |

