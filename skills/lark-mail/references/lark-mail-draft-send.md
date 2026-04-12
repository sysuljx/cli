# mail +draft-send

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

发送已有草稿，支持定时发送功能。

本 skill 对应 shortcut：`lark-cli mail +draft-send`。

## 定时发送

草稿发送支持两种定时方式：

- **`--send-time`**：指定 Unix 时间戳（秒），绝对时间
- **`--send-after`**：指定相对时间（如 `30m`、`2h`、`1d`），从现在起多久之后发送

定时发送要求：
- 发送时间必须距离当前至少 5 分钟
- `--send-time` 优先级高于 `--send-after`，两者同时设置时会打印警告并使用 `--send-time`

## 命令

```bash
# 立即发送草稿
lark-cli mail +draft-send --draft-id DR_xxx --mailbox me

# 定时发送（绝对时间，Unix 时间戳）
lark-cli mail +draft-send --draft-id DR_xxx --send-time 1775846400

# 定时发送（相对时间，1 小时后）
lark-cli mail +draft-send --draft-id DR_xxx --send-after 1h

# 定时发送（30 分钟后）
lark-cli mail +draft-send --draft-id DR_xxx --send-after 30m

# Dry Run
lark-cli mail +draft-send --draft-id DR_xxx --send-after 1h --dry-run
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--draft-id <id>` | 是 | 要发送的草稿 ID |
| `--mailbox <email>` | 否 | 邮箱地址（默认: me） |
| `--send-time <timestamp>` | 否 | Unix 时间戳（秒），定时发送的绝对时间 |
| `--send-after <duration>` | 否 | 相对时间，如 `30m`、`2h`、`1d`，定时发送的相对时间 |
| `--dry-run` | 否 | 仅打印请求，不执行 |

## 返回值

**立即发送：**

```json
{
  "ok": true,
  "data": {
    "message_id": "邮件ID",
    "thread_id": "会话ID"
  }
}
```

**定时发送：**

```json
{
  "ok": true,
  "data": {
    "message_id": "邮件ID",
    "thread_id": "会话ID",
    "state": "SCHEDULED",
    "send_time": 1775846400
  }
}
```

## 典型场景

### 场景 1：立即发送草稿

```bash
# 先通过 +draft-create 或 UI 创建草稿，获取 draft_id
lark-cli mail +draft-send --draft-id DR_abc123 --mailbox me
```

### 场景 2：定时发送会议提醒

```bash
# 1 小时后发送会议提醒邮件
lark-cli mail +draft-send --draft-id DR_meeting_reminder --send-after 1h
```

### 场景 3：定时在特定时间发送

```bash
# 在指定时间发送（需要计算 Unix 时间戳）
# 2026-04-13 00:00:00 UTC = 1775856000
lark-cli mail +draft-send --draft-id DR_xxx --send-time 1775856000
```

## 取消定时发送

定时发送的邮件在发送前可以取消：

```bash
# 取消定时发送，邮件将退回草稿状态
lark-cli mail +cancel-scheduled-send --message-id MSG_xxx --mailbox me
```

## 相关命令

- `lark-cli mail +draft-create` — 创建新草稿
- `lark-cli mail +draft-edit` — 编辑草稿
- `lark-cli mail +cancel-scheduled-send` — 取消定时发送
- `lark-cli mail user_mailbox.drafts list` — 列出草稿
