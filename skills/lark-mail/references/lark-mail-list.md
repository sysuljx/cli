# mail message list

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

列出指定文件夹或标签下的邮件列表。必须指定 `folder_id` 或 `label_id` 之一。

## 用法

```bash
# 列出收件箱邮件
lark-cli mail user_mailbox.messages list \
  --params '{"user_mailbox_id":"me","page_size":20,"folder_id":"INBOX"}'

# 列出已发送邮件
lark-cli mail user_mailbox.messages list \
  --params '{"user_mailbox_id":"me","page_size":20,"folder_id":"SENT"}'

# 列出草稿
lark-cli mail user_mailbox.messages list \
  --params '{"user_mailbox_id":"me","page_size":20,"folder_id":"DRAFT"}'

# 列出定时发送邮件（SCHEDULED）
lark-cli mail user_mailbox.messages list \
  --params '{"user_mailbox_id":"me","page_size":20,"label_id":"SCHEDULED"}'
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `user_mailbox_id` | 是 | 邮箱 ID，通常传 `me` |
| `page_size` | 是 | 每页返回条数 |
| `folder_id` | 否* | 文件夹 ID。系统值：`INBOX`、`SENT`、`DRAFT`、`TRASH`、`SPAM`、`ARCHIVED` |
| `label_id` | 否* | 标签 ID。系统值：`INBOX`、`SENT`、`DRAFT`、`STARRED`、`SCHEDULED`（定时发送邮件） |
| `page_token` | 否 | 分页令牌 |

> \* `folder_id` 和 `label_id` 必须且只能提供一个。

## --label / label_id 支持的值

| 值 | 说明 |
|----|------|
| `INBOX` | 收件箱 |
| `SENT` | 已发送 |
| `DRAFT` | 草稿 |
| `STARRED` | 星标邮件 |
| `SCHEDULED` | 定时发送邮件。设置了 `send_time` 的邮件在到达发送时间前处于此状态。可通过 `+cancel-scheduled-send` 取消定时发送并退回草稿状态 |

## 典型场景

### 查看待发送的定时邮件

```bash
# 方法 1：原生 API
lark-cli mail user_mailbox.messages list \
  --params '{"user_mailbox_id":"me","page_size":20,"label_id":"SCHEDULED"}'

# 方法 2：使用 +triage shortcut
lark-cli mail +triage --filter '{"folder":"scheduled"}'
```

### 取消某封定时邮件

```bash
# 先列出定时邮件获取 message_id
lark-cli mail +triage --filter '{"folder":"scheduled"}'

# 取消定时发送
lark-cli mail +cancel-scheduled-send --mailbox me --message-id <message_id>
```

## 相关命令

- `lark-cli mail +triage` — 邮件摘要浏览（支持 `--filter '{"folder":"scheduled"}'`）
- `lark-cli mail +cancel-scheduled-send` — 取消定时发送
- `lark-cli mail +message` — 查看单封邮件详情
