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
	// 1. 初始化NodeServer组件
	//    让 InitNodeServer 自己去处理所有命令行参数。
	if _, err := server.InitNodeServer(ServerName); err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// 2. 创建NodeServer实例
	nodeServer, err := service.NewNodeServer()
	if err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// 3. 自动迁移数据库表
	service.RegisterTables(dbclient.GetMysqlDB())

	// 4. 将此节点注册到 Etcd 以进行服务发现
	if err = nodeServer.Register(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("register node into etcd error:%s", err.Error()))
		os.Exit(1)
	}

	// 5. 运行节点服务器的核心逻辑
	if err = nodeServer.Run(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node run error: %s", err.Error()))
		os.Exit(1)
	}

	// 6. 在后台启动通知服务
	go notify.Serve()

	logger.GetLogger().Info(fmt.Sprintf("pulse node %s service started, Ctrl+C or send kill sign to exit", nodeServer.String()))

	// 7. 设置正常关机
	event.OnEvent(event.EXIT, nodeServer.Stop) // 注册 EXIT 事件的回调函数
	event.WaitEvent()                          // 阻塞并等待终止信号
	event.EmitEvent(event.EXIT, nil)           // 触发 EXIT 事件

	logger.GetLogger().Info("exit success")
}
