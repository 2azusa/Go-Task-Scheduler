package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"pulse/common/pkg/config"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond

	// green   = "\033[97;42m"
	// white   = "\033[90;47m"
	// yellow  = "\033[90;43m"
	// red     = "\033[97;41m"
	// blue    = "\033[97;44m"
	// magenta = "\033[97;45m"
	// cyan    = "\033[97;46m"
	reset = "\033[0m"
)

// 定义了API服务器的命令行启动参数
var ApiOptions struct {
	flags.Options
	Environment       string `short:"e" long:"env" description:"Use ApiServer environment" default:"testing"`
	Version           bool   `short:"v" long:"verbose"  description:"Show ApiServer version"`
	EnablePProfile    bool   `short:"p" long:"enable-pprof"  description:"enable pprof"` // 是否启用 pprof 性能分析
	PProfilePort      int    `short:"d" long:"pprof-port"  description:"pprof port" default:"8188"`
	EnableHealthCheck bool   `short:"a" long:"enable-health-check"  description:"enable health check"`
	HealthCheckURI    string `short:"i" long:"health-check-uri"  description:"health check uri" default:"/health" ` // 是否启用健康检查
	HealthCheckPort   int    `short:"f" long:"health-check-port"  description:"health check port" default:"8186"`
	ConfigFileName    string `short:"c" long:"config" description:"Use ApiServer config file" default:"main"` // 配置文件名
	EnableDevMode     bool   `short:"m" long:"enable-dev-mode"  description:"enable dev mode"`
}

// // 用于健康检查
// type healthCheckHttpServer struct{}

// func (server *healthCheckHttpServer) ServerHTTP(response http.ResponseWriter, request *http.Request) {
// 	io.WriteString(response, "ok\n")
// }

// var healthCheckServer = &healthCheckHttpServer{}

// 定义了API服务器的结构体
type ApiServer struct {
	Engine      *gin.Engine         // Gin框架的引擎实例
	HttpServer  *http.Server        // Http服务器
	Addr        string              // 监听地址和端口
	Routers     []func(*gin.Engine) // 路由注册函数列表
	Middlewares []func(*gin.Engine) // 中间注册函数列表
	Shutdowns   []func(*ApiServer)  // 关闭时执行的函数列表
	Services    []func(*ApiServer)  // 启动时执行的函数列表
}

// 创建API服务器的构造函数
func NewApiServer(serverName string, inits ...func()) (*ApiServer, error) {
	// 解析命令行参数
	var parser = flags.NewParser(&ApiOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		return nil, err
	}

	// 打印版本信息
	if ApiOptions.Version {
		fmt.Printf("%s Version: %s\n", ApiModule, Version)
		os.Exit(0)
	}

	// 启用pprof
	if ApiOptions.EnablePProfile {
		go func() {
		}()
	}

	// 启动healthcheck
	if ApiOptions.EnableHealthCheck {
		go func() {
			fmt.Printf("enable healthcheck http server at: %d\n", ApiOptions.HealthCheckPort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.HealthCheckPort), nil))
		}()
	}

	// 初始化配置信息
	var env = config.Environment(ApiOptions.Environment)
	if env.Invalid() {
		var err error
		env, err = config.NewGlobalEnvironment()
		if err != nil {
			return nil, err
		}
	}

	var configFile = ApiOptions.ConfigFileName
	if configFile == "" {
		configFile = "main"
	}
	defaultConfig, err := config.LoadConfig(env.String(), serverName, configFile)
	if err != nil {
		fmt.Printf("api-server: init config error: %s", err.Error())
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
		Url:  defaultConfig.WebHook.Url,
		Kind: defaultConfig.WebHook.Kind,
	})
	// 初始化数据库
	dsn := mysqlConfig.EmptyDsn()
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4;", mysqlConfig.Dbname)
	if err := dbclient.CreateDatabase(dsn, "mysql", createSql); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create mysql database failed , error:%s", err.Error()))
	}
	_, err = dbclient.Init(mysqlConfig.Dsn(), mysqlConfig.LogMode, mysqlConfig.MaxIdleConns, mysqlConfig.MaxOpenConns)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server: init mysql failed , error:%s", err.Error()))
	} else {
		logger.GetLogger().Info("api-server: init mysql success")
	}
	// 初始化 Etcd 客户端
	_, err = etcdclient.Init(etcdConfig.Endpoints, etcdConfig.DialTimeout, etcdConfig.ReqTimeout)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server: init etcd failed , error:%s", err.Error()))
	} else {
		logger.GetLogger().Info("api-server: init etcd success")
	}

	// 执行传入的额外初始化函数
	if len(inits) > 0 {
		for _, init := range inits {
			init()
		}
	}

	// 创建ApiServer实例
	apiServer := &ApiServer{
		Addr: fmt.Sprintf(":%d", defaultConfig.System.Addr),
	}
	apiServer.setupSignal() // 设置信号监听

	// 根据环境设置Gin的运行模式
	switch env {
	case config.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case config.EnvTesting:
		gin.SetMode(gin.DebugMode)
	}

	return apiServer, nil
}

// 关闭API服务器
func (srv *ApiServer) Shutdown(ctx context.Context) {
	if len(srv.Shutdowns) > 0 {
		for _, shutdown := range srv.Shutdowns {
			shutdown(srv)
		}
	}

	time.Sleep(shutdownWait)
	srv.HttpServer.Shutdown(ctx)
}

// 启动http服务器并监听
func (srv *ApiServer) ListenAndServe() error {
	srv.Engine = gin.New() // 创建gin引擎
	srv.Engine.Use(srv.apiRecoveryMiddleware())

	for _, service := range srv.Services {
		service(srv)
	}
	for _, middleware := range srv.Middlewares {
		middleware(srv.Engine)
	}
	for _, router := range srv.Routers {
		router(srv.Engine)
	}

	srv.HttpServer = &http.Server{
		Handler:        srv.Engine,
		Addr:           srv.Addr,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// 启动服务器，阻塞操作
	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// gin中间件，用于捕获panic防止程序崩溃
func (srv *ApiServer) apiRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 网络错误无需详细记录堆栈
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by pper") {
							brokenPipe = true
						}
					}
				}

				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				headers := strings.Split(string(httpRequest), "\r\n")
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				if brokenPipe {
					logger.GetLogger().Error(fmt.Sprintf("%s\n%s%s", err, string(httpRequest), reset))
				} else {
					logger.GetLogger().Error(fmt.Sprintf("[Recovery] %s panic recoverd:\n%s\n%s%s", formatTime(time.Now()), err, stack, reset))
				}

				if brokenPipe {
					c.Error(err.(error))
					c.Abort()
				} else {
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next()
	}
}

// 设置信号监听，用于实现优雅关闭
func (srv *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown: signal %v to stop api-server", sig))
				srv.Shutdown(shutdownCtx) // 优雅关闭
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown() // 关闭日志系统
	}()
}

// 注册一个或多个关闭钩子函数
func (srv *ApiServer) RegisterShutdown(handlers ...func(*ApiServer)) {
	srv.Shutdowns = append(srv.Shutdowns, handlers...)
}

// 注册一个或多个服务
func (srv *ApiServer) RegisterService(handlers ...func(*ApiServer)) {
	srv.Services = append(srv.Services, handlers...)
}

// 注册一个或多个中间件
func (srv *ApiServer) RegisterMiddleware(middlewares ...func(engine *gin.Engine)) {
	srv.Middlewares = append(srv.Middlewares, middlewares...)
}

// 注册一个或多个路由组
func (srv *ApiServer) RegisterRouters(routers ...func(engine *gin.Engine)) *ApiServer {
	srv.Routers = append(srv.Routers, routers...)
	return srv
}
