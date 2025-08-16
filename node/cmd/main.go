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
	if _, err := server.InitNodeServer(ServerName); err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}
	nodeServer, err := service.NewNodeServer()
	if err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}
	service.RegisterTables(dbclient.GetMysqlDB())
	if err = nodeServer.Register(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("register node into etcd error:%s", err.Error()))
		os.Exit(1)
	}
	if err = nodeServer.Run(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node run error: %s", err.Error()))
		os.Exit(1)
	}
	//notification operation
	go notify.Serve()
	logger.GetLogger().Info(fmt.Sprintf("pulse node %s service started, Ctrl+C or send kill sign to exit", nodeServer.String()))
	// Register the logout event
	event.OnEvent(event.EXIT, nodeServer.Stop)
	// Listen for exit signals
	event.WaitEvent()
	event.EmitEvent(event.EXIT, nil)
	logger.GetLogger().Info("exit success")
}
