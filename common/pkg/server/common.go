package server

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

var (
	dunno     = []byte("???")
	centerDot = []byte("·")
	dot       = []byte(".")
	slash     = []byte("/")
)

const (
	Version    = "v1.1.0"            // 版本号
	ApiModule  = "pulse/api-server"  // API 服务器模块名
	NodeModule = "pulse/node-server" // Node 服务器模块名
)

// "年/月/日 - 时:分:秒"
func formatTime(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}

// 获取当前调用堆栈信息，用于panic恢复时记录详细错误
func stack(skip int) []byte {
	buf := new(bytes.Buffer)
	var lines [][]byte
	var lastFile string
	// 无限循环，直到runtime.Caller报告没有更多的调用帧
	for i := skip; ; i++ {
		// 获取调用者的程序计数器、文件、行号
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)

		if file != lastFile {
			// 如果是新文件，则读取该文件的内容
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		// 打印导致panic的具体函数名和源代码行
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}

	return buf.Bytes()
}

// 用于从源代码行中获取指定行号的内容
func source(lines [][]byte, n int) []byte {
	n-- // 堆栈跟踪的行号是 1-based，而切片索引是 0-based
	if n < 0 || n >= len(lines) {
		return dunno
	}

	return bytes.TrimSpace(lines[n])
}

// 用于根据程序计数器返回函数名
func function(pc uintptr) []byte {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno
	}
	name := []byte(fn.Name())

	// 移除路径前缀
	if lastSlash := bytes.LastIndex(name, slash); lastSlash >= 0 {
		name = name[lastSlash+1:]
	}
	// 移除包名
	if period := bytes.Index(name, dot); period >= 0 {
		name = name[period+1:]
	}
	// 将中间点换成普通点
	name = bytes.ReplaceAll(name, centerDot, dot)

	return name
}
