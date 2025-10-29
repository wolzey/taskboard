# Configuration

Taskboard supports configuration via both YAML config files and environment variables. Environment variables take precedence over config file values.

## Configuration Methods

### 1. Config File (config.yaml)

Create a `config.yaml` file in one of these locations:
- Current directory (`.`)
- `./config/`
- `$HOME/.taskboard/`

Copy `config.example.yaml` to `config.yaml` and adjust values as needed:

```bash
cp config.example.yaml config.yaml
```

### 2. Environment Variables

All configuration values can be set via environment variables with the `TASKBOARD_` prefix.

The format is: `TASKBOARD_<SECTION>_<KEY>`

Examples:
```bash
export TASKBOARD_REDIS_HOST=redis.example.com
export TASKBOARD_REDIS_PORT=6380
export TASKBOARD_REDIS_PASSWORD=mypassword
export TASKBOARD_REDIS_USE_TLS=true
export TASKBOARD_REDIS_TLS_CA_FILE=/path/to/ca.crt
export TASKBOARD_API_PORT=8080
```

## Configuration Options

### Redis Configuration

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `redis.host` | `TASKBOARD_REDIS_HOST` | `localhost` | Redis server hostname |
| `redis.port` | `TASKBOARD_REDIS_PORT` | `6379` | Redis server port |
| `redis.password` | `TASKBOARD_REDIS_PASSWORD` | `""` | Redis password (if required) |
| `redis.username` | `TASKBOARD_REDIS_USERNAME` | `""` | Redis username (Redis 6+ ACL) |
| `redis.db` | `TASKBOARD_REDIS_DB` | `0` | Redis database number |
| `redis.use_tls` | `TASKBOARD_REDIS_USE_TLS` | `false` | Enable TLS/SSL connection |

### Redis TLS Configuration

Only used when `redis.use_tls` is `true`:

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `redis.tls.cert_file` | `TASKBOARD_REDIS_TLS_CERT_FILE` | `""` | Path to client certificate file |
| `redis.tls.key_file` | `TASKBOARD_REDIS_TLS_KEY_FILE` | `""` | Path to client key file |
| `redis.tls.ca_file` | `TASKBOARD_REDIS_TLS_CA_FILE` | `""` | Path to CA certificate file |
| `redis.tls.insecure_skip_verify` | `TASKBOARD_REDIS_TLS_INSECURE_SKIP_VERIFY` | `false` | Skip TLS certificate verification (not recommended for production) |

### API Configuration

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `api.port` | `TASKBOARD_API_PORT` | `1337` | API server port |

### Queue Configuration

| Config Key | Environment Variable | Default | Description |
|------------|---------------------|---------|-------------|
| `queue.prefix` | `TASKBOARD_QUEUE_PREFIX` | `bull` | Queue prefix for job keys in Redis (change if using different prefix) |

## Examples

### Basic Configuration (No TLS)

**config.yaml:**
```yaml
redis:
  host: localhost
  port: 6379
api:
  port: 1337
queue:
  prefix: bull
```

**or via environment variables:**
```bash
export TASKBOARD_REDIS_HOST=localhost
export TASKBOARD_REDIS_PORT=6379
export TASKBOARD_API_PORT=1337
export TASKBOARD_QUEUE_PREFIX=bull
```

### Redis with Authentication

**config.yaml:**
```yaml
redis:
  host: redis.example.com
  port: 6379
  username: default
  password: mysecretpassword
  db: 0
api:
  port: 1337
queue:
  prefix: bull
```

**or via environment variables:**
```bash
export TASKBOARD_REDIS_HOST=redis.example.com
export TASKBOARD_REDIS_USERNAME=default
export TASKBOARD_REDIS_PASSWORD=mysecretpassword
```

### Redis with TLS

**config.yaml:**
```yaml
redis:
  host: redis.example.com
  port: 6380
  password: mysecretpassword
  use_tls: true
  tls:
    ca_file: /path/to/ca.crt
    cert_file: /path/to/client.crt
    key_file: /path/to/client.key
api:
  port: 1337
queue:
  prefix: bull
```

**or via environment variables:**
```bash
export TASKBOARD_REDIS_HOST=redis.example.com
export TASKBOARD_REDIS_PORT=6380
export TASKBOARD_REDIS_PASSWORD=mysecretpassword
export TASKBOARD_REDIS_USE_TLS=true
export TASKBOARD_REDIS_TLS_CA_FILE=/path/to/ca.crt
export TASKBOARD_REDIS_TLS_CERT_FILE=/path/to/client.crt
export TASKBOARD_REDIS_TLS_KEY_FILE=/path/to/client.key
```

### Redis Cloud/Managed Service (TLS without client certs)

**config.yaml:**
```yaml
redis:
  host: my-redis.cloud.provider.com
  port: 6380
  password: mysecretpassword
  use_tls: true
api:
  port: 1337
queue:
  prefix: bull
```

**or via environment variables:**
```bash
export TASKBOARD_REDIS_HOST=my-redis.cloud.provider.com
export TASKBOARD_REDIS_PORT=6380
export TASKBOARD_REDIS_PASSWORD=mysecretpassword
export TASKBOARD_REDIS_USE_TLS=true
```

### Custom Queue Prefix

If your BullMQ implementation uses a custom prefix instead of "bull", you can configure it:

**config.yaml:**
```yaml
redis:
  host: localhost
  port: 6379
api:
  port: 1337
queue:
  prefix: myapp  # Custom prefix
```

**or via environment variables:**
```bash
export TASKBOARD_QUEUE_PREFIX=myapp
```

This will make the application look for queues like `myapp:queuename` instead of `bull:queuename`.

## Precedence

Configuration sources are loaded in this order (later sources override earlier ones):
1. Default values
2. Config file (`config.yaml`)
3. Environment variables
