package main

import (
	"fmt"
	"os"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"
	"pulse/common/pkg/server"
	"pulse/common/pkg/utils/event"
	"pulse/node/internal/service"
)

const ServerName = "node"

func main() {
	// 1. 初始化节点服务的基础组件
	if _, err := server.InitNodeServer(ServerName); err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// 2. 创建 NodeServer 业务逻辑实例
	nodeServer, err := service.NewNodeServer()
	if err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// 3. 初始化数据库表结构
	service.RegisterTables(dbclient.GetMysqlDB())

	// 4. 将节点注册到 Etcd
	// 节点通过 Register 将自己的信息写入 Etcd， admin 服务就能发现它，实现了服务的动态上线
	if err = nodeServer.Register(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("register node into etcd error:%s", err.Error()))
		os.Exit(1)
	}

	// 5. 启用节点的核心运行逻辑
	// Run 方法会加载分配给本节点的任务，启动 cron 调度器，并开始监听 Etcd 中与任务相关的变化
	if err = nodeServer.Run(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node run error: %s", err.Error()))
		os.Exit(1)
	}

	// 6. 在后台启动通知服务
	go notify.Serve()

	logger.GetLogger().Info(fmt.Sprintf("pulse node %s service started, Ctrl+C or send kill sign to exit", nodeServer.String()))

	// 7. 设置优雅关闭逻辑
	event.OnEvent(event.EXIT, nodeServer.Stop) // 设置 “EXIT” 事件的监听器，当事件被触发时，调用 nodeServer.Stop 方法进行清理
	event.WaitEvent()                          // 阻塞主 goroutine， 直到接收到操作系统的终止信号
	event.EmitEvent(event.EXIT, nil)           // 接收到信号后，手动触发 “EXIT” 事件

	logger.GetLogger().Info("exit success")
}
