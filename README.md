# Zap Logger

基于 [uber-go/zap](https://github.com/uber-go/zap) 和 [lumberjack](https://github.com/natefinch/lumberjack) 封装的高性能 Go 日志库。支持自动轮转、多输出源、动态级别调整。

## ✨ 特性

- **高性能**: 基于 Zap，生产环境首选。
- **自动轮转**: 集成 Lumberjack，支持按大小、日期切割日志。
- **多输出**: 支持同时输出到控制台（带颜色）和文件（JSON/Console）。
- **动态级别**: 运行时可动态调整日志级别，无需重启服务。
- **全局单例**: 提供全局 `L` (Logger) 和 `S` (SugaredLogger) 方便调用。

## 📦 安装

```bash
go get github.com/jxintelink/logger
```

导入：

```go
import "github.com/jxintelink/logger"
```

## 使用注意

- `InitLogger` 在进程内建议只调用一次；重复调用会覆盖全局 `L`、`S` 以及 `zap` 的全局 logger。
- `Filename` 为空且 `LogToStd` 为 `false` 时，日志不会输出到任何地方；库会在首次遇到这种情况时向 **stderr** 打印一条警告。
- `Level` 为无法识别的字符串时，实际级别为 **info**。