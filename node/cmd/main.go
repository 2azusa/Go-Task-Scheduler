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
	// Initialize node server components.
	if _, err := server.InitNodeServer(ServerName); err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// Create a new NodeServer instance.
	nodeServer, err := service.NewNodeServer()
	if err != nil {
		fmt.Println("init node server error:", err.Error())
		os.Exit(1)
	}

	// Auto-migrate database tables.
	service.RegisterTables(dbclient.GetMysqlDB())

	// Register this node to Etcd for service discovery.
	if err = nodeServer.Register(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("register node into etcd error:%s", err.Error()))
		os.Exit(1)
	}

	// Run the node server's core logic, including cron scheduler and Etcd watchers.
	if err = nodeServer.Run(); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node run error: %s", err.Error()))
		os.Exit(1)
	}

	// Start the notification service in the background.
	go notify.Serve()

	logger.GetLogger().Info(fmt.Sprintf("pulse node %s service started, Ctrl+C or send kill sign to exit", nodeServer.String()))

	// Set up graceful shutdown.
	event.OnEvent(event.EXIT, nodeServer.Stop) // Call nodeServer.Stop on EXIT event.
	event.WaitEvent()                          // Block until a termination signal is received.
	event.EmitEvent(event.EXIT, nil)           // Trigger the EXIT event.

	logger.GetLogger().Info("exit success")
}
