| Package | Source | Passing | Warnings | Failing |
|---------|--------|--------:|---------:|--------:|
| `psr/log` | `src` | 6 | 0 | 2 |
| `psr/container` | `src` | 3 | 0 | 0 |
| `psr/http-message` | `src` | 7 | 0 | 0 |
| `guzzlehttp/psr7` | `src` | 2 | 80 | 11 |
| `guzzlehttp/guzzle` | `src` | 20 | 27 | 19 |
| `symfony/console` | `.` | 48 | 36 | 66 |
| `symfony/finder` | `.` | 12 | 7 | 6 |
| `monolog/monolog` | `src` | 30 | 246 | 40 |
| `nikic/php-parser` | `lib` | 167 | 780 | 52 |
| `doctrine/inflector` | `src` | 29 | 24 | 11 |

Parse errors, by cause:

- 99 files: parse error: unexpected token 8("?")
- 32 files: parse error: a namespaced file may only declare classes and functions, because it is scanned for the symbols i...
- 19 files: parse error: unexpected token 8(".")
- 11 files: parse error: unexpected token 8("=>")
- 9 files: parse error: expected ";", got 8(",")
- 9 files: parse error: expected ")", got 8("=")
- 7 files: parse error: unexpected token 8("=")
- 6 files: parse error: expected ")", got 8(":")
- 3 files: parse error: expected ")", got 2("record")
- 2 files: parse error: expected ")", got 8("::")
- 1 files: parse error: unterminated string
- 1 files: parse error: unexpected token in class body: 8("?")

Undefined functions (the binding gap), by calls:

- `trigger_deprecation()` x35
- `curl_setopt()` x18
- `match()` x10
- `current()` x5
- `curl_init()` x5
- `assert()` x5
- `trigger_error()` x3
- `stream_select()` x3
- `stream_set_chunk_size()` x2
- `parse_url()` x2
- `newrelic_notice_error()` x2
- `newrelic_add_custom_parameter()` x2
- `flock()` x2
- `syslog()` x1
- `stream_set_blocking()` x1
- `stream_get_meta_data()` x1
- `stream_filter_append()` x1
- `stream_context_create()` x1
- `socket_send()` x1
- `socket_create()` x1
- `socket_connect()` x1
- `proc_open()` x1
- `proc_close()` x1
- `pcntl_signal_get_handler()` x1
- `openlog()` x1
- `next()` x1
- `newrelic_set_appname()` x1
- `newrelic_name_transaction()` x1
- `mt_rand()` x1
- `mb_strcut()` x1
- `key()` x1
- `from()` x1
- `fileinode()` x1
- `curl_setopt_array()` x1
- `curl_exec()` x1
- `curl_error()` x1
- `curl_errno()` x1
- `closelog()` x1
- `array_walk_recursive()` x1
