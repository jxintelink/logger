package logger

import (
	"os"
	"testing"
)

func TestInitLogger(t *testing.T) {
	logPath := "./test_logs/test.log"
	// 测试完成后清理文件
	defer os.RemoveAll("./test_logs")

	cfg := Config{
		Filename:   logPath,
		MaxSize:    1,
		MaxBackups: 3,
		MaxAge:     7,
		Compress:   false,
		Level:      "debug",
		Format:     "console",
		ShowLine:   true,
		LogToStd:   false, // 测试中关闭控制台输出
	}

	// 1. 测试初始化
	InitLogger(&cfg)
	if L == nil || S == nil {
		t.Fatal("Logger 初始化失败，全局变量为 nil")
	}

	// 2. 测试基本写入
	L.Info("Testing info log")
	S.Debugf("Testing debug log with format: %s", "hello")

	// 3. 验证文件是否生成
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("日志文件未成功创建: %s", logPath)
	}

	// 4. 测试动态级别调整
	L.SetLevel("error")
	if L.AtomicLevel().Level() != parseLogLevel("error") {
		t.Errorf("日志级别调整失败，期望 error，得到 %v", L.AtomicLevel().Level())
	}
}
