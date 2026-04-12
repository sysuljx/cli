# mail +draft-send / +cancel-scheduled-send

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

草稿发送与定时发送管理，支持：
- 立即发送草稿
- 定时发送草稿（`--send-time` 或 `--send-after`）
- 取消定时发送（`+cancel-scheduled-send`）

## +draft-send

发送已有草稿，可选择立即发送或定时发送。

### 命令

```bash
# 立即发送草稿
lark-cli mail +draft-send --mailbox me --draft-id DR_xxx

# 定时发送：1 小时后
lark-cli mail +draft-send --mailbox me --draft-id DR_xxx --send-after 1h

# 定时发送：30 分钟后
lark-cli mail +draft-send --mailbox me --draft-id DR_xxx --send-after 30m

# 定时发送：指定绝对时间（Unix 秒时间戳）
lark-cli mail +draft-send --mailbox me --draft-id DR_xxx --send-time 1775846400

# Dry Run
lark-cli mail +draft-send --mailbox me --draft-id DR_xxx --send-after 2h --dry-run
```

### 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--mailbox <id>` | 否 | 邮箱地址，默认 `me` |
| `--draft-id <id>` | 是 | 要发送的草稿 ID |
| `--send-time <unix_ts>` | 否 | 定时发送的绝对时间（Unix 秒时间戳）。为空或 0 时立即发送。必须至少为当前时间 5 分钟后 |
| `--send-after <duration>` | 否 | 定时发送的相对时间（如 `30m`、`2h`、`1d`）。CLI 内部转换为绝对 Unix 秒。必须至少 5 分钟 |
| `--dry-run` | 否 | 仅打印请求，不执行 |

> **优先级**：`--send-time` > `--send-after` > 立即发送。两者同时给出时，以 `--send-time` 为准并打印 warning。

### 状态流转

```
DRAFT ---(+draft-send)---> SENT          (立即发送)
DRAFT ---(+draft-send --send-time/--send-after)---> SCHEDULED ---(到达时间)---> SENT
SCHEDULED ---(+cancel-scheduled-send)---> DRAFT
```

## +send 定时发送

`+send` 同样支持 `--send-time` / `--send-after`，用于创建新邮件并定时发送。

```bash
# 创建新邮件并定时 2 小时后发送
lark-cli mail +send --to alice@example.com --subject '周报' \
  --body '<p>本周进展...</p>' --send-after 2h --confirm-send

# 创建新邮件并在指定时间发送
lark-cli mail +send --to alice@example.com --subject '通知' \
  --body '<p>会议提醒</p>' --send-time 1775846400 --confirm-send
```

## +cancel-scheduled-send

取消处于 `SCHEDULED` 状态的定时发送邮件，将其退回 `DRAFT` 状态。

### 命令

```bash
# 取消定时发送
lark-cli mail +cancel-scheduled-send --mailbox me --message-id MSG_xxx
```

### 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--mailbox <id>` | 否 | 邮箱地址，默认 `me` |
| `--message-id <id>` | 是 | 要取消定时发送的邮件 ID（messageBizID），必须处于 SCHEDULED 状态 |

### 错误场景

| 场景 | 错误信息 |
|------|---------|
| `--send-after 30s`（不足 5 分钟） | `scheduled send must be at least 5 minutes from now` |
| `--send-time` 为过去时间 | 后端：`BAD_REQUEST: invalid send_time` |
| 对 DRAFT 状态邮件调用 cancel | 后端：`message not in scheduled state` |

## 相关命令

- `lark-cli mail +send` — 创建新邮件（支持定时发送）
- `lark-cli mail +triage --filter '{"folder":"scheduled"}'` — 查看定时发送邮件列表
- `lark-cli mail +message --message-id <id>` — 查看邮件详情
