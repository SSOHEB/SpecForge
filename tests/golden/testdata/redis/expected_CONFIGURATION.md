# Configuration Reference

## Redis

**Path:** `redis`

| Name | Type | Default | Required | Constraints | Description |
| --- | --- | --- | --- | --- | --- |
| `db` | integer | `0` | No | Min: 0, Max: 15 | Redis logical database index |
| `dial_timeout` | integer | `5` | No | - | Connection establishment timeout in seconds |
| `host` | string | `localhost` | No | - | Redis server host name or IP address |
| `password` | string | - | No | - | Optional password for authentication |
| `pool_size` | integer | `10` | No | Min: 1, Max: 1000 | Maximum number of socket connections in pool |
| `port` | integer | `6379` | No | Min: 1, Max: 65535 | Redis server port number |
| `read_timeout` | integer | `3` | No | - | Socket read timeout in seconds |
| `sentinel_addrs` | array of strings | - | No | - | List of Redis Sentinel addresses for HA configurations |
| `tls_enabled` | boolean | `false` | No | - | Enables secure TLS connection |
| `write_timeout` | integer | `3` | No | - | Socket write timeout in seconds |

