# id3-tags

解析 ID3v2 标签链：同步安全长度、帧标志、文本编码、帧去重，并按帧 ID 建索引。只用标准库。

```
go test ./...
go vet ./...
go run ./cmd/id3tags --sample=synchsafe
go run ./cmd/id3tags --sample=dli
go run ./cmd/id3tags --sample=utf16
go run ./cmd/id3tags --sample=dupframe
go run ./cmd/id3tags --sample=chain
go run ./cmd/id3tags --sample=work
```

## 口径

- **标签长度**：头部第 6 到第 10 字节是同步安全整数（每字节低 7 位有效）。
- **标签链**：文件里可以接着放多个 ID3v2 标签，逐个解析。
- **帧长度**：v2.3 用普通大端整数，v2.4 用同步安全整数。
- **帧标志**：v2.4 的帧标志最低位置位时，帧数据开头有 4 字节同步安全的数据长度指示符，取数据时要跳过它。
- **文本帧**：ID 以 `T` 开头的帧，数据首字节是编码：0 拉丁、1 带 BOM 的 UTF-16、2 大端 UTF-16、3 UTF-8；解码成文本。
- **去重**：同一个帧 ID 只保留第一次出现的那一帧。
- **查找**：按帧 ID 走索引，不逐帧比较。

## 不变量

- 标签链长度等于文件里真实标签数；帧列表不含填充。
- 文本帧的 `Text` 不含编码字节；同一帧 ID 只对应第一帧。
- `scanned` 不随帧数乘查询次数放大：八百个帧查八百次的 `scanned` 不超过 6000。

## 输出契约

`Tag` 含版本、标签长度与帧列表；`Frame` 含 `ID` / `Size` / `Data` / `Text`；
`Find(id)` 按 ID 取第一帧。场景打印一行 JSON。`--sample=work` 打印 `{"frames": N, "scanned": N}`。
