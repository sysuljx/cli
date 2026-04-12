# mail +cancel-scheduled-send

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

取消尚未发送的定时邮件，将邮件退回草稿状态。

本 skill 对应 shortcut：`lark-cli mail +cancel-scheduled-send`。

## 状态机

```
DRAFT → SCHEDULED → SENT
         ↑
         └── +cancel-scheduled-send 可在此处取消，退回 DRAFT
```

## 命令

```bash
# 取消定时发送（使用 message-id）
lark-cli mail +cancel-scheduled-send --message-id MSG_xxx --mailbox me

# Dry Run
lark-cli mail +cancel-scheduled-send --message-id MSG_xxx --dry-run
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--message-id <id>` | 是 | 要取消的定时邮件的 message ID（messageBizID） |
| `--mailbox <email>` | 否 | 邮箱地址（默认: me） |
| `--dry-run` | 否 | 仅打印请求，不执行 |

## 返回值

**成功：**

```json
{
  "ok": true,
  "data": {}
}
```

**失败（邮件不在 SCHEDULED 状态）：**

```json
{
  "ok": false,
  "error": "message not in scheduled state"
}
```

## 典型场景

### 场景 1：取消定时发送

```bash
# 查看即将发送的定时邮件
lark-cli mail user_mailbox messages list --user_mailbox_id me --label_id SCHEDULED

# 取消其中一封的定时发送
lark-cli mail +cancel-scheduled-send --message-id MSG_abc123 --mailbox me
# → 邮件退回草稿状态，可用 +draft-send 重新发送或 +draft-edit 编辑
```

### 场景 2：修改定时邮件内容后重新发送

```bash
# 1. 取消定时发送
lark-cli mail +cancel-scheduled-send --message-id MSG_xxx --mailbox me

# 2. 编辑草稿内容
lark-cli mail +draft-edit --draft-id DR_xxx --patch-file ./edit.json

# 3. 重新设置定时发送
lark-cli mail +draft-send --draft-id DR_xxx --send-after 1h
```

## 限制

- 只能取消 `SCHEDULED` 状态的邮件
- 仅支持用户身份（user token），bot 身份无法调用此接口
- 取消后邮件退回草稿状态，可以重新编辑或立即发送

## 相关命令

- `lark-cli mail +draft-send` — 发送草稿（支持定时发送）
- `lark-cli mail +draft-edit` — 编辑草稿
- `lark-cli mail user_mailbox messages list --label_id SCHEDULED` — 列出定时邮件
