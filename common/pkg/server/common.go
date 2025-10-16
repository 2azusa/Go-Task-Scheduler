package server

import (
	"bytes"
	"fmt"
	"os"
	"runtime"
	"time"
)

// 定义预先分配的字节切片，用于字符串拼接
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

// formatTime 格式化时间为 "年/月/日 - 时:分:秒" 的字符串格式
func formatTime(t time.Time) string {
	var timeString = t.Format("2006/01/02 - 15:04:05")
	return timeString
}

// stack 函数用于获取并格式化当前的调用堆栈信息，用于 panic 恢复时记录详细错误
func stack(skip int) []byte {
	buf := new(bytes.Buffer) // 创建一个 buffer 来存储堆栈信息
	var lines [][]byte
	var lastFile string
	// 无限循环，直到 runtime.Caller 报告没有更多的调用帧
	for i := skip; ; i++ {
		// 获取调用者的程序计数器、文件名、行号和是否成功
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// 打印文件、行号和程序计数器地址
		fmt.Fprintf(buf, "%s:%d (0x%x)\n", file, line, pc)
		if file != lastFile {
			// 如果是新的文件，则读取该文件的内容
			data, err := os.ReadFile(file)
			if err != nil {
				continue
			}
			lines = bytes.Split(data, []byte{'\n'})
			lastFile = file
		}
		// 打印导致 panic 的具体函数名和源代码行
		fmt.Fprintf(buf, "\t%s: %s\n", function(pc), source(lines, line))
	}
	return buf.Bytes()
}

// source 用于从源代码行中获取指定行号的内容
func source(lines [][]byte, n int) []byte {
	n-- // 堆栈跟踪中的行号是 1-based，而切片索引是 0-based
	if n < 0 || n >= len(lines) {
		return dunno // 如果行号无效，返回未知占位符
	}
	// 返回对应行的源码，并去除前后的空白字符
	return bytes.TrimSpace(lines[n])
}

// function 函数用于根据程序计数器返回函数名
func function(pc uintptr) []byte {
	// 获取 pc 对应的函数信息
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return dunno // 找不到则返回未知
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
