# lark-slides xml_presentations create

## 用途

创建空白的飞书幻灯片（PPT）演示文稿。

## 命令

```bash
lark-cli slides xml_presentations create --data '<json_data>'
```

## 参数说明

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `--data` | JSON string | 是 | 请求体，字段结构以 `lark-cli schema slides.xml_presentations.create` 为准 |

### data JSON 结构

```json
{
  "xml_presentation": {
    "content": "<presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" width=\"960\" height=\"540\"><title>项目汇报</title></presentation>",
    "presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef",
    "revision_id": 1
  }
}
```

### 字段说明

| 字段 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `xml_presentation.content` | string | 否 | 演示文稿的 XML 内容 |
| `xml_presentation.presentation_id` | string | 否 | 演示文稿 ID |
| `xml_presentation.revision_id` | integer | 否 | 演示文稿版本号 |

> 常见新建场景只需要传 `xml_presentation.content`，而且建议内容保持为空白 PPT 模板。XML 协议本身以 [slides_xml_schema_definition.xml](slides_xml_schema_definition.xml) 为准。

> **当前命令行为说明**：`xml_presentations.create` 当前只适合创建空白 PPT，通常只支持指定标题和长宽。页面内容不要在这里一次性传入，后续 slide 请使用 [xml_presentation.slides create](lark-slides-xml-presentation-slides-create.md) 逐页添加。

## 使用示例

### 创建空白 PPT

```bash
lark-cli slides xml_presentations create --data '{
  "xml_presentation": {
    "content": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" width=\"960\" height=\"540\"><title>演示文稿标题</title></presentation>"
  }
}'
```

### 创建后再逐页添加 slide

```bash
PRESENTATION_ID=$(lark-cli slides xml_presentations create --data '{
  "xml_presentation": {
    "content": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" width=\"960\" height=\"540\"><title>项目汇报</title></presentation>"
  }
}' | jq -r '.xml_presentation_id')

lark-cli slides xml_presentation.slides create --params "{\"xml_presentation_id\":\"$PRESENTATION_ID\"}" --data '{
  "slide": {
    "content": "<slide xmlns=\"http://www.larkoffice.com/sml/2.0\"><style><fill><fillColor color=\"rgb(245, 245, 245)\"/></fill></style><data><shape type=\"text\" topLeftX=\"80\" topLeftY=\"72\" width=\"760\" height=\"90\"><content textType=\"title\"><p>Q3 项目汇报</p></content></shape><shape type=\"text\" topLeftX=\"80\" topLeftY=\"190\" width=\"520\" height=\"220\"><content textType=\"body\"><p>关键结论</p><ul><li><p>页面加载速度提升 40%</p></li><li><p>用户满意度提升</p></li><li><p>新增功能 12 个</p></li></ul></content></shape><shape type=\"rect\" topLeftX=\"660\" topLeftY=\"180\" width=\"180\" height=\"140\"><fill><fillColor color=\"rgba(100, 149, 237, 0.25)\"/></fill><border color=\"rgb(100, 149, 237)\" width=\"2\"/></shape></data></slide>"
  }
}'
```

### 从文件读取 XML

```bash
lark-cli slides xml_presentations create --data "$(jq -n --arg content "$(cat presentation.xml)" '{xml_presentation:{content:$content}}')"
```

其中 `presentation.xml` 建议保持为空白模板，例如：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<presentation xmlns="http://www.larkoffice.com/sml/2.0" width="960" height="540">
  <title>项目汇报</title>
</presentation>
```

## 关键 XML 结构说明

### presentation 根元素

| 属性 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `width` | positiveInteger | 是 | 演示文稿宽度（像素），如 960 |
| `height` | positiveInteger | 是 | 演示文稿高度（像素），如 540 |
| `id` | string | 否 | 演示文稿唯一标识符 |

**子元素：**
- `<title>` - 演示文稿标题（可选）
- `<theme>` - 全局主题设置（可选）
- `<slide>` - 幻灯片页面（必需，1-100页）

> 协议层仍然允许 `<presentation>` 包含 `<slide>`，但当前 `xml_presentations.create` 命令不建议承载 slide 内容；实际页面请改用 `xml_presentation.slides.create`。

## 返回值

成功时返回创建结果：

```json
{
  "xml_presentation_id": "abc",
  "revision_id": 1
}
```

## 常见错误

| 错误码 | 含义 | 解决方案 |
|--------|------|----------|
| 400 | XML 格式错误 | 检查 XML 语法是否正确，确保标签闭合 |
| 400 | 缺少必需属性 | 确保 `<presentation>` 包含 `width` 和 `height` 属性 |
| 400 | create 内容超出支持范围 | 只传标题和长宽，slide 内容改用 `xml_presentation.slides.create` |
| 400 | 请求体结构错误 | 检查是否按 `xml_presentation.content` 包装 XML |
| 403 | 权限不足 | 检查是否拥有 `slides:presentation:create` scope |

## 注意事项

1. **执行前必做**: 使用 `lark-cli schema slides.xml_presentations.create` 查看最新的参数结构
2. **命名空间建议**: 协议标准写法应带 `xmlns`，例如 `<presentation xmlns="http://www.larkoffice.com/sml/2.0" ...>`；当前服务端实现可能兼容不带 `xmlns` 的输入，但不作为协议保证
3. **create 只建空白 PPT**: 建议只传标题和长宽，不要把页面内容直接塞进 create
4. **页面内容添加方式**: 后续 slide 请用 [xml_presentation.slides create](lark-slides-xml-presentation-slides-create.md)
5. **JSON 转义**: 如果直接内联 XML，需要正确转义双引号
6. **创建成功后保存返回的 `xml_presentation_id` 和 `revision_id`**
7. **完整 Schema 定义**: 参考 [slides_xml_schema_definition.xml](slides_xml_schema_definition.xml)

## 相关命令

- [xml_presentations get](lark-slides-xml-presentations-get.md) - 读取 PPT 内容
- [xml_presentation.slides create](lark-slides-xml-presentation-slides-create.md) - 添加幻灯片页面
- [xml_presentation.slides delete](lark-slides-xml-presentation-slides-delete.md) - 删除幻灯片页面
- [xml-format-guide.md](xml-format-guide.md) - XML 格式详细规范
- [xml-schema-quick-ref.md](xml-schema-quick-ref.md) - Schema 快速参考
