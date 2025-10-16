package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"pulse/common/pkg/config"
	"pulse/common/pkg/dbclient"
	"pulse/common/pkg/etcdclient"
	"pulse/common/pkg/logger"
	"pulse/common/pkg/notify"

	"github.com/gin-gonic/gin"
	"github.com/jessevdk/go-flags"
)

const (
	shutdownMaxAge = 15 * time.Second
	shutdownWait   = 1000 * time.Millisecond
)
const (
	// green   = "\033[97;42m"
	// white   = "\033[90;47m"
	// yellow  = "\033[90;43m"
	// red     = "\033[97;41m"
	// blue    = "\033[97;44m"
	// magenta = "\033[97;45m"
	// cyan    = "\033[97;46m"
	reset = "\033[0m" // reset 用于重置终端颜色
)

// ApiOptions 定义了 API 服务器的命令行启动参数
var (
	ApiOptions struct {
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
)

// healthCheckHttpServer 用于健康检查
type healthCheckHttpServer struct{}

// ServeHTTP 实现了 http.Handler 接口
func (server *healthCheckHttpServer) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	io.WriteString(response, "ok\n")
}

var healthCheckServer = &healthCheckHttpServer{}

// ApiServer 定义了 API 服务器的结构体
type ApiServer struct {
	Engine     *gin.Engine  // Gin 框架的引擎实例
	HttpServer *http.Server // HTTP 服务器
	Addr       string       // 监听地址和端口
	// mu          sync.Mutex
	// doneChan    chan struct{}
	Routers     []func(*gin.Engine) // 路有注册函数列表
	Middlewares []func(*gin.Engine) // 中间件注册函数列表
	Shutdowns   []func(*ApiServer)  // 关闭时需要执行的钩子函数列表
	Services    []func(*ApiServer)  // 服务启动时需要执行的函数列表
}

// get close Chan
// func (srv *ApiServer) getDoneChan() <-chan struct{} {
// 	srv.mu.Lock()
// 	defer srv.mu.Unlock()
// 	return srv.getDoneChanLocked()
// }

// func (srv *ApiServer) getDoneChanLocked() chan struct{} {
// 	if srv.doneChan == nil {
// 		srv.doneChan = make(chan struct{})
// 	}
// 	return srv.doneChan
// }

// Shutdown 定义了服务器的关闭逻辑
func (srv *ApiServer) Shutdown(ctx context.Context) {
	// 优先执行业务注册的关闭钩子函数
	if len(srv.Shutdowns) > 0 {
		for _, shutdown := range srv.Shutdowns {
			shutdown(srv)
		}
	}

	// 等待一段时间
	time.Sleep(shutdownWait)
	// 关闭 HTTP 服务器
	srv.HttpServer.Shutdown(ctx)
}

// apiRecoveryMiddleware 是一个 Gin 中间件，用于捕获任何 panic 防止程序崩溃
func (srv *ApiServer) apiRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 检查错误是否为 "broken pipe" 或 "connection reset by peer"，这类网络错误无需详细记录堆栈
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				// 获取堆栈信息，跳过前三个调用帧
				stack := stack(3)
				httpRequest, _ := httputil.DumpRequest(c.Request, false) // 获取 HTTP 请求的详细信息
				headers := strings.Split(string(httpRequest), "\r\n")    // 为了安全，隐藏请求头中的 Authorization 信息
				for idx, header := range headers {
					current := strings.Split(header, ":")
					if current[0] == "Authorization" {
						headers[idx] = current[0] + ": *"
					}
				}

				if brokenPipe {
					// 如果是 broken pipe 错误，只记录简单的错误信息
					logger.GetLogger().Error(fmt.Sprintf("%s\n%s%s", err, string(httpRequest), reset))
				} else {
					// 对于其他 panic，记录详细的恢复信息和堆栈
					logger.GetLogger().Error(fmt.Sprintf("[Recovery] %s panic recovered:\n%s\n%s%s",
						formatTime(time.Now()), err, stack, reset))
				}

				if brokenPipe {
					c.Error(err.(error)) // 将错误传递给 Gin 的错误处理
					c.Abort()
				} else {
					// 返回 500 内部服务器错误
					c.AbortWithStatus(http.StatusInternalServerError)
				}
			}
		}()
		c.Next() // 继续处理请求链中的下一个中间件或处理器
	}
}

// setupSignal 设置信号监听，用于实现优雅关闭
func (srv *ApiServer) setupSignal() {
	go func() {
		var sigChan = make(chan os.Signal, 1)
		// 监听 SIGINT, SIGHUP, SIGTERM 信号
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)
		// 创建一个带有超时的上下文，用于关闭操作
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownMaxAge)
		defer shutdownCancel()

		for sig := range sigChan {
			if sig == syscall.SIGINT || sig == syscall.SIGHUP || sig == syscall.SIGTERM {
				logger.GetLogger().Error(fmt.Sprintf("Graceful shutdown:signal %v to stop api-server ", sig))
				srv.Shutdown(shutdownCtx) // 调用优雅关闭函数
			} else {
				logger.GetLogger().Info(fmt.Sprintf("Caught signal %v", sig))
			}
		}
		logger.Shutdown() // 关闭日志系统
	}()
}

// NewApiServer 是创建 ApiServer 实例的构造函数
func NewApiServer(serverName string, inits ...func()) (*ApiServer, error) {
	// 解析命令行参数
	var parser = flags.NewParser(&ApiOptions, flags.Default)
	if _, err := parser.Parse(); err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0) // 如果是 --help 命令， 则正常退出
		}

		return nil, err
	}

	// 如果指定了 -v 或 --version，打印版本信息并退出
	if ApiOptions.Version {
		fmt.Printf("%s Version:%s\n", ApiModule, Version)
		os.Exit(0)
	}

	// 根据命令行参数，选择性地启动 pprof 和 healthcheck 服务
	if ApiOptions.EnablePProfile {
		go func() {
			fmt.Printf("enable pprof http server at:%d\n", ApiOptions.PProfilePort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.PProfilePort), nil))
		}()
	}

	if ApiOptions.EnableHealthCheck {
		go func() {
			fmt.Printf("enable healthcheck http server at:%d\n", ApiOptions.HealthCheckPort)
			fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", ApiOptions.HealthCheckPort), healthCheckServer))
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

	// 加载配置文件
	var configFile = ApiOptions.ConfigFileName
	if configFile == "" {
		configFile = "main"
	}
	defaultConfig, err := config.LoadConfig(env.String(), serverName, configFile)
	if err != nil {
		fmt.Printf("api-server:init config error:%s", err.Error())
		return nil, err
	}

	// 初始化各个组件：日志、通知、数据库、Etcd
	logConfig := defaultConfig.Log
	mysqlConfig := defaultConfig.Mysql
	etcdConfig := defaultConfig.Etcd
	// 初始化日志
	logger.Init(serverName, logConfig.Level, logConfig.Format, logConfig.Prefix, logConfig.Director, logConfig.ShowLine, logConfig.EncodeLevel, logConfig.StacktraceKey, logConfig.LogInConsole)
	// 初始化通知服务
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
	createSql := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 ;", mysqlConfig.Dbname)
	if err := dbclient.CreateDatabase(dsn, "mysql", createSql); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("create mysql database failed , error:%s", err.Error()))
	}
	_, err = dbclient.Init(mysqlConfig.Dsn(), mysqlConfig.LogMode, mysqlConfig.MaxIdleConns, mysqlConfig.MaxOpenConns)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server:init mysql failed , error:%s", err.Error()))
	} else {
		logger.GetLogger().Info("api-server:init mysql success")
	}
	// 初始化 Etcd 客户端
	_, err = etcdclient.Init(etcdConfig.Endpoints, etcdConfig.DialTimeout, etcdConfig.ReqTimeout)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("api-server:init etcd failed , error:%s", err.Error()))
	} else {
		logger.GetLogger().Info("api-server:init etcd success")
	}

	// 执行传入的额外初始化函数
	if len(inits) > 0 {
		for _, init := range inits {
			init()
		}
	}

	// 创建 ApiServer 实例
	apiServer := &ApiServer{
		Addr: fmt.Sprintf(":%d", defaultConfig.System.Addr),
	}

	apiServer.setupSignal() // 设置信号监听

	// 根据环境设置 Gin 的运行模式
	switch env {
	case config.EnvProduction:
		gin.SetMode(gin.ReleaseMode)
	case config.EnvTesting:
		gin.SetMode(gin.DebugMode)
	}
	return apiServer, nil
}

// ListenAndServe 启动 HTTP 服务器并开始监听请求
func (srv *ApiServer) ListenAndServe() error {
	srv.Engine = gin.New()                      // 创建一个新的 Gin 引擎
	srv.Engine.Use(srv.apiRecoveryMiddleware()) // 使用自定义的 panic 恢复中间件

	// 注册服务
	for _, service := range srv.Services {
		service(srv)
	}

	// 组 注册中间件
	for _, middleware := range srv.Middlewares {
		middleware(srv.Engine)
	}

	// 注册所有路由
	for _, c := range srv.Routers {
		c(srv.Engine)
	}

	// 配置并创建 HTTP 服务器
	srv.HttpServer = &http.Server{
		Handler:        srv.Engine,
		Addr:           srv.Addr,
		ReadTimeout:    20 * time.Second,
		WriteTimeout:   20 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// 启动服务器，这是一个阻塞操作
	if err := srv.HttpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// RegisterShutdown 注册一个或多个关闭钩子函数
func (srv *ApiServer) RegisterShutdown(handlers ...func(*ApiServer)) {
	srv.Shutdowns = append(srv.Shutdowns, handlers...)
}

// RegisterService 注册一个或多个服务
func (srv *ApiServer) RegisterService(handlers ...func(*ApiServer)) {
	srv.Services = append(srv.Services, handlers...)
}

// RegisterMiddleware 注册一个或多个中间件
func (srv *ApiServer) RegisterMiddleware(middlewares ...func(engine *gin.Engine)) {
	srv.Middlewares = append(srv.Middlewares, middlewares...)
}

// RegisterRouters 注册一个或多个路由组
func (srv *ApiServer) RegisterRouters(routers ...func(engine *gin.Engine)) *ApiServer {
	srv.Routers = append(srv.Routers, routers...)
	return srv
}
