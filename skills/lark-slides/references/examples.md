# 完整操作示例

本文档提供和 CLI schema 一致的调用示例，XML 内容均遵循 [slides_xml_schema_definition.xml](slides_xml_schema_definition.xml)。

> **重要**：`xml_presentations.create` 当前只用于创建空白 PPT；实际页面内容请使用 `xml_presentation.slides.create` 逐页添加。

## 目录

- [示例 1: 创建空白演示文稿](#示例-1-创建空白演示文稿)
- [示例 2: 创建后添加第一页](#示例-2-创建后添加第一页)
- [示例 3: 读取 XML 内容](#示例-3-读取-xml-内容)
- [示例 4: 在指定页面前插入新幻灯片](#示例-4-在指定页面前插入新幻灯片)
- [示例 5: 删除幻灯片](#示例-5-删除幻灯片)
- [示例 6: 从文件读取 XML 再创建](#示例-6-从文件读取-xml-再创建)

## 示例 1: 创建空白演示文稿

```bash
lark-cli slides xml_presentations create --data '{
  "xml_presentation": {
    "content": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" width=\"960\" height=\"540\"><title>项目汇报</title></presentation>"
  }
}'
```

预期返回结构：

```json
{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef",
  "revision_id": 1
}
```

## 示例 2: 创建后添加第一页

```bash
PRESENTATION_ID=$(lark-cli slides xml_presentations create --data '{
  "xml_presentation": {
    "content": "<?xml version=\"1.0\" encoding=\"UTF-8\"?><presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" width=\"960\" height=\"540\"><title>季度复盘</title></presentation>"
  }
}' | jq -r '.xml_presentation_id')

lark-cli slides xml_presentation.slides create --params "{\"xml_presentation_id\":\"$PRESENTATION_ID\"}" --data '{
  "slide": {
    "content": "<slide xmlns=\"http://www.larkoffice.com/sml/2.0\"><style><fill><fillColor color=\"rgb(245, 245, 245)\"/></fill></style><data><shape type=\"text\" topLeftX=\"80\" topLeftY=\"72\" width=\"760\" height=\"90\"><content textType=\"title\"><p>2024 Q3 季度复盘</p></content></shape><shape type=\"text\" topLeftX=\"80\" topLeftY=\"190\" width=\"520\" height=\"220\"><content textType=\"body\"><p>关键结论</p><ul><li><p>收入增长 30%</p></li><li><p>重点项目全部上线</p></li><li><p>用户满意度持续提升</p></li></ul></content></shape><shape type=\"rect\" topLeftX=\"660\" topLeftY=\"180\" width=\"180\" height=\"140\"><fill><fillColor color=\"rgba(100, 149, 237, 0.25)\"/></fill><border color=\"rgb(100, 149, 237)\" width=\"2\"/></shape></data><note><content textType=\"body\"><p>讲述时先给结论，再补充数据。</p></content></note></slide>"
  }
}'
```

## 示例 3: 读取 XML 内容

```bash
lark-cli slides xml_presentations get --params '{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef"
}'
```

提取 XML 内容：

```bash
lark-cli slides xml_presentations get --params '{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef"
}' | jq -r '.xml_presentation.content'
```

预期返回结构：

```json
{
  "xml_presentation": {
    "presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef",
    "revision_id": 3,
    "content": "<presentation xmlns=\"http://www.larkoffice.com/sml/2.0\" height=\"540\" width=\"960\">...</presentation>"
  }
}
```

## 示例 4: 在指定页面前插入新幻灯片

```bash
lark-cli slides xml_presentation.slides create --params '{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef"
}' --data '{
  "slide": {
    "content": "<slide xmlns=\"http://www.larkoffice.com/sml/2.0\"><data><shape type=\"text\" topLeftX=\"80\" topLeftY=\"80\" width=\"800\" height=\"120\"><content textType=\"title\"><p>新增页面</p></content></shape><shape type=\"text\" topLeftX=\"80\" topLeftY=\"200\" width=\"800\" height=\"180\"><content textType=\"body\"><p>这是新增页面的正文。</p></content></shape></data></slide>"
  },
  "before_slide_id": "sld_before_target"
}'
```

预期返回结构：

```json
{
  "slide_id": "slidecn20m4XJ1hJXXxXXX",
  "revision_id": 100
}
```

## 示例 5: 删除幻灯片

```bash
lark-cli slides xml_presentation.slides delete --params '{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef",
  "slide_id": "sld_xxx"
}'
```

预期返回结构：

```json
{
  "revision_id": 101
}
```

## 示例 6: 从文件读取 XML 再创建

先准备 `presentation.xml`：

```xml
<?xml version="1.0" encoding="UTF-8"?>
<presentation xmlns="http://www.larkoffice.com/sml/2.0" width="960" height="540">
  <title>从文件创建的演示文稿</title>
</presentation>
```

再用 `jq` 组装请求体创建空白 PPT：

```bash
lark-cli slides xml_presentations create --data "$(jq -n --arg content "$(cat presentation.xml)" '{xml_presentation:{content:$content}}')"
```

后续页面内容请继续使用 `xml_presentation.slides create` 添加。

## 常见处理技巧

### 获取最新 revision_id

```bash
lark-cli slides xml_presentations get --params '{
  "xml_presentation_id": "S7YwsFIGIlnS2qdscKDc1Yabcef"
}' | jq -r '.xml_presentation.revision_id'
```

### 批量插入多页

```bash
#!/bin/bash

PRESENTATION_ID="S7YwsFIGIlnS2qdscKDc1Yabcef"

slides=(
  '<slide xmlns="http://www.larkoffice.com/sml/2.0"><data><shape type="text" topLeftX="80" topLeftY="80" width="800" height="120"><content textType="title"><p>页面 1</p></content></shape></data></slide>'
  '<slide xmlns="http://www.larkoffice.com/sml/2.0"><data><shape type="text" topLeftX="80" topLeftY="80" width="800" height="120"><content textType="title"><p>页面 2</p></content></shape></data></slide>'
)

for slide_xml in "${slides[@]}"; do
  payload=$(jq -n --arg content "$slide_xml" '{slide:{content:$content}}')
  lark-cli slides xml_presentation.slides create --params "{\"xml_presentation_id\":\"$PRESENTATION_ID\"}" --data "$payload"
done
```

### 本地校验 XML 基本语法

```bash
xmllint --noout presentation.xml
```

### 真实示例

- [slides_demo.xml](slides_demo.xml) 提供了更完整的页面示例，包含 `theme`、渐变填充、图片、图标和备注内容。
