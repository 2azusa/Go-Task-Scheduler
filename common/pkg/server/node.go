package server

import (
	"fmt"
	"net/http"
	"os"
	"pulse/common/models"
	"pulse/common/pkg/config"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"

	"github.com/jessevdk/go-flags"
)

// 节点服务器的命令行启动参数
var NodeOptions struct {
	flags.Options
	Environment    string `short:"e" long:"environment" description:"Use NodeServer environment"`
	Version        bool   `short:"v" long:"version"  description:"Show NodeServer version"`
	EnablePProfile bool   `short:"p" long:"enable-pprof"  description:"enable pprof"`
	PProfilePort   int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8288"`
	ConfigFileName string `short:"c" long:"configfilename" description:"Use NodeServer config file"`
	EnableDevMode  bool   `short:"m" long:"enable-dev-mode"  description:"enable dev mode"`
}

// 初始化节点服务器
// 不创建实例，而是初始化所有必要的依赖
func InitNodeServer(serverName string, inits ...func()) (*models.Config, error) {
	var parser = flags.NewParser(&NodeOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return nil, err
	}

	// 打印版本信息
	if NodeOptions.Version {
		fmt.Printf("%s Version:%s\n", NodeModule, Version)
		os.Exit(0)
	}
	// 启动pprof
	if NodeOptions.EnablePProfile {
		go func() {
			fmt.Printf("enable pprof http server at:%d\n", NodeOptions.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", NodeOptions.PProfilePort), nil))
		}()
	}

	// 初始化配置环境
	var env = config.Environment(NodeOptions.Environment)
	if env.Invalid() {
		var err error
		env, err = config.NewGlobalEnvironment()
		if err != nil {
			return nil, err
		}
	}

	var configFile = NodeOptions.ConfigFileName
	if configFile == "" {
		configFile = "main"
	}
	defaultConfig, err := config.LoadConfig(env.String(), serverName, configFile)
	if err != nil {
		fmt.Printf("node-server:init config error:%s", err.Error())
		return nil, err
	}

	// 初始化各个公共组件
	logConfig := defaultConfig.Log
	mysqlConfig := defaultConfig.Mysql
	etcdConfig := defaultConfig.Etcd
	// 初始化日志
	logger.Init(
		serverName,
		logConfig.Level,
		logConfig.Format,
		logConfig.Prefix,
		logConfig.Director,
		logConfig.ShowLine,
		logConfig.EncodeLevel,
		logConfig.StacktraceKey,
		logConfig.LogInConsole,
	)
	// 初始化通知
	notify.Init(&notify.Mail{
		Port:     defaultConfig.Email.Port,
		From:     defaultConfig.Email.From,
		Host:     defaultConfig.Email.Host,
		Secret:   defaultConfig.Email.Secret,
		Nickname: defaultConfig.Email.Nickname,
	}, &notify.WebHook{
		Kind: defaultConfig.WebHook.Kind,
		Url:  defaultConfig.WebHook.Url,
	})
	// 初始化数据库
	dsn := mysqlConfig.EmptyDsn()
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAUlt CHARACTER SET utf8mb4;", mysqlConfig.Dbname)
	if err := dbclient.CreateDatabase(dsn, "mysql", createSql); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create mysql database failed, error: %s", err.Error()))
	}
	_, err = dbclient.Init(mysqlConfig.Dsn(), mysqlConfig.LogMode, mysqlConfig.MaxIdleConns, mysqlConfig.MaxOpenConns)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node-server: init mysql failed, error: %s", err.Error()))
	} else {
		logger.GetLogger().Info("node-server: init mysql success")
	}
	// 初始化Etcd
	_, err = etcdclient.Init(etcdConfig.Endpoints, etcdConfig.DialTimeout, etcdConfig.ReqTimeout)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("node-server: init etcd failed, error: %s", err.Error()))
	} else {
		logger.GetLogger().Info("node-server: init etcd success")
	}
	if len(inits) > 0 {
		for _, init := range inits {
			init()
		}
	}

	return defaultConfig, nil
}
