package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 全局变量，方便直接使用
var (
	L *Logger
	S *zap.SugaredLogger
)

// Logger 封装 zap.Logger，并持有动态级别控制
type Logger struct {
	*zap.Logger
	atomicLevel zap.AtomicLevel
}

// Config 结构体，用于初始化参数（参数多时建议用结构体，方便后期扩展）
type Config struct {
	Filename   string // 日志文件路径
	MaxSize    int    // MB
	MaxBackups int
	MaxAge     int
	Compress   bool
	Level      string // debug, info, warn, error...
	Format     string // json 或 console
	ShowLine   bool   // 是否显示行号
	LogToStd   bool   // 是否同时输出到控制台
	Colorize   bool   // 是否颜色化输出
}

func NewConfig(filename string, maxSize int, maxBackups int, maxAge int, compress bool, level string, format string, showLine bool, logToStd bool, colorize bool) *Config {
	return &Config{
		Filename:   filename,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   compress,
		Level:      level,
		Format:     format,
		ShowLine:   showLine,
		LogToStd:   logToStd,
		Colorize:   colorize,
	}
}

// warnNoOutputOnce 在无输出目标时仅向 stderr 提示一次，避免误配后静默丢日志
var warnNoOutputOnce sync.Once

// InitLogger 初始化全局日志单例（进程内建议只调用一次，否则会覆盖全局 L/S 与 zap 全局 logger）
func InitLogger(cfg *Config) {
	L = NewLogger(cfg)
	S = L.Sugar()
	// 替换 zap 的全局 logger
	zap.ReplaceGlobals(L.Logger)
}

// NewLogger 创建一个新的 Logger 实例
func NewLogger(cfg *Config) *Logger {
	// 1. 设置动态日志级别
	atomicLevel := zap.NewAtomicLevelAt(parseLogLevel(cfg.Level))

	// 2. 获取编码器配置
	// 如果输出到文件且格式是 JSON，强制不带颜色
	encoder := newEncoder(cfg.Format, cfg.Colorize)

	var cores []zapcore.Core

	// 3. 核心 1：文件输出 (Lumberjack 轮转)
	if cfg.Filename != "" {
		// 自动创建日志目录
		dir := filepath.Dir(cfg.Filename)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			_ = os.MkdirAll(dir, 0755)
		}

		lj := &lumberjack.Logger{
			Filename:   cfg.Filename,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(lj), atomicLevel))
	}

	// 4. 核心 2：控制台输出
	if cfg.LogToStd {
		// 控制台通常建议用 console 格式且带颜色
		consoleEnc := newEncoder("console", cfg.Colorize)
		cores = append(cores, zapcore.NewCore(consoleEnc, zapcore.Lock(os.Stdout), atomicLevel))
	}

	if len(cores) == 0 {
		warnNoOutputOnce.Do(func() {
			_, _ = fmt.Fprintf(os.Stderr, "logger: unset Filename and LogToStd, no output will be written\n")
		})
	}

	// 5. 合并并构建 Logger
	core := zapcore.NewTee(cores...)

	options := []zap.Option{}
	if cfg.ShowLine {
		options = append(options, zap.AddCaller())
	}

	zl := zap.New(core, options...)

	return &Logger{
		Logger:      zl,
		atomicLevel: atomicLevel,
	}
}

// --- 内部辅助函数 ---

func newEncoder(format string, colorize bool) zapcore.Encoder {
	// 通用配置
	zec := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder, // 易读的时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	format = strings.ToLower(strings.TrimSpace(format))
	if format == "json" {
		zec.EncodeLevel = zapcore.CapitalLevelEncoder
		return zapcore.NewJSONEncoder(zec)
	}

	// 默认为 console 格式；是否带 ANSI 颜色由 colorize 控制
	zec.EncodeLevel = zapcore.CapitalLevelEncoder
	if colorize {
		zec.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	return zapcore.NewConsoleEncoder(zec)
}

// parseLogLevel 解析级别字符串；无法识别时默认为 info
func parseLogLevel(level string) zapcore.Level {
	level = strings.ToLower(strings.TrimSpace(level))
	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// --- 外部可调用方法 ---

func (l *Logger) SetLevel(level string) {
	l.atomicLevel.SetLevel(parseLogLevel(level))
}

func (l *Logger) SetColorize(colorize bool) {

}

func (l *Logger) AtomicLevel() zap.AtomicLevel {
	return l.atomicLevel
}

// Sync 刷新缓冲区，建议在 main 退出前调用
func Sync() {
	if L != nil {
		_ = L.Sync()
	}
}
