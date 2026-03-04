package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"hackmitm/pkg/cert"
	"hackmitm/pkg/config"
	"hackmitm/pkg/logger"
	"hackmitm/pkg/plugin"
	"hackmitm/pkg/proxy"
)

// 版本信息，在构建时通过 ldflags 注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// 命令行参数
var (
	configFile = flag.String("config", "configs/config.json", "配置文件路径")
	logLevel   = flag.String("log-level", "", "日志级别 (debug, info, warn, error)")
	version    = flag.Bool("version", false, "显示版本信息")
	help       = flag.Bool("help", false, "显示帮助信息")
	daemon     = flag.Bool("daemon", false, "以守护进程模式运行")
	pidFile    = flag.String("pid-file", "", "PID 文件路径")
)

// 颜色常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorBold   = "\033[1m"
)

func main() {
	// 解析命令行参数
	flag.Parse()

	// 显示帮助信息
	if *help {
		showHelp()
		return
	}

	// 显示版本信息
	if *version {
		showVersion()
		return
	}

	// 处理守护进程模式
	if *daemon {
		if err := runAsDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "启动守护进程失败: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 显示启动横幅
	showBanner()

	// 初始化日志系统
	if err := initLogger(); err != nil {
		printError("初始化日志系统失败: %v", err)
		os.Exit(1)
	}

	printInfo("🚀 HackMITM v%s 启动中...", Version)
	printInfo("📅 构建时间: %s", BuildTime)
	printInfo("🔗 Git提交: %s", GitCommit)
	printInfo("🐹 Go版本: %s", runtime.Version())
	printInfo("💻 操作系统: %s/%s", runtime.GOOS, runtime.GOARCH)
	printSeparator()

	// 写入 PID 文件
	if *pidFile != "" {
		if err := writePidFile(*pidFile); err != nil {
			printError("写入PID文件失败: %v", err)
			os.Exit(1)
		}
		defer removePidFile(*pidFile)
	}

	// 加载配置文件
	printInfo("📋 正在加载配置文件...")
	cfg, err := loadConfig(*configFile)
	if err != nil {
		printError("加载配置文件失败: %v", err)
		os.Exit(1)
	}
	printSuccess("✅ 配置文件加载成功")

	// 初始化证书管理器
	printInfo("🔐 正在初始化证书管理器...")
	certMgr, err := initCertManager(cfg)
	if err != nil {
		printError("初始化证书管理器失败: %v", err)
		os.Exit(1)
	}
	printSuccess("✅ 证书管理器初始化成功")

	// 创建代理服务器
	printInfo("🌐 正在创建代理服务器...")
	server, err := proxy.NewServer(cfg, certMgr)
	if err != nil {
		printError("创建代理服务器失败: %v", err)
		os.Exit(1)
	}
	printSuccess("✅ 代理服务器创建成功")

	// 加载插件
	printInfo("🔌 正在加载插件...")
	if err := loadPlugins(server, cfg); err != nil {
		printWarning("⚠️  插件加载失败: %v", err)
	} else {
		printSuccess("✅ 插件加载完成")
	}

	// 创建上下文用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 启动服务器
	printInfo("🚀 正在启动服务器...")
	if err := server.Start(); err != nil {
		printError("启动服务器失败: %v", err)
		os.Exit(1)
	}

	// 显示启动成功信息
	printSeparator()
	printSuccess("🎉 HackMITM 代理服务器启动成功!")
	serverConfig := cfg.GetServer()
	printInfo("📡 监听地址: %s%s:%d%s", ColorBold+ColorCyan, serverConfig.ListenAddr, serverConfig.ListenPort, ColorReset)
	printInfo("🌍 代理地址: %shttp://%s:%d%s", ColorBold+ColorGreen, serverConfig.ListenAddr, serverConfig.ListenPort, ColorReset)
	printInfo("📊 监控地址: %shttp://%s:%d%s", ColorBold+ColorBlue, serverConfig.ListenAddr, cfg.GetMonitoring().Port, ColorReset)
	printSeparator()
	printInfo("💡 提示: 按 %sCtrl+C%s 停止服务器", ColorBold+ColorYellow, ColorReset)
	printSeparator()

	// 等待中断信号
	waitForShutdown(ctx, server)

	printSuccess("👋 HackMITM 已安全退出")
}

// showBanner 显示启动横幅
func showBanner() {
	banner := `
%s%s
██╗  ██╗ █████╗  ██████╗██╗  ██╗███╗   ███╗██╗████████╗███╗   ███╗
██║  ██║██╔══██╗██╔════╝██║ ██╔╝████╗ ████║██║╚══██╔══╝████╗ ████║
███████║███████║██║     █████╔╝ ██╔████╔██║██║   ██║   ██╔████╔██║
██╔══██║██╔══██║██║     ██╔═██╗ ██║╚██╔╝██║██║   ██║   ██║╚██╔╝██║
██║  ██║██║  ██║╚██████╗██║  ██╗██║ ╚═╝ ██║██║   ██║   ██║ ╚═╝ ██║
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝   ╚═╝     ╚═╝
%s
%s                    高性能 HTTP/HTTPS 代理服务器%s
%s                        Version: %s%s
`
	fmt.Printf(banner,
		ColorBold+ColorCyan,
		ColorReset,
		ColorReset,
		ColorBold+ColorGreen,
		ColorReset,
		ColorBold+ColorYellow,
		Version,
		ColorReset,
	)
}

// showHelp 显示帮助信息
func showHelp() {
	fmt.Printf(`%s%sHackMITM - 高性能 HTTP/HTTPS 代理服务器%s

`, ColorBold, ColorCyan, ColorReset)
	fmt.Printf("%s用法:%s\n", ColorBold, ColorReset)
	fmt.Printf("  %s [选项]\n\n", os.Args[0])
	fmt.Printf("%s选项:%s\n", ColorBold, ColorReset)
	flag.PrintDefaults()
	fmt.Printf("\n%s示例:%s\n", ColorBold, ColorReset)
	fmt.Printf("  %s%s -config configs/config.json%s\n", ColorGreen, os.Args[0], ColorReset)
	fmt.Printf("  %s%s -config configs/config-no-plugins.json -log-level debug%s\n", ColorGreen, os.Args[0], ColorReset)
	fmt.Printf("  %s%s -daemon -pid-file /var/run/hackmitm.pid%s\n", ColorGreen, os.Args[0], ColorReset)
	fmt.Printf("\n%s更多信息:%s https://github.com/your-org/hackmitm\n", ColorBold, ColorReset)
}

// showVersion 显示版本信息
func showVersion() {
	fmt.Printf("%s%sHackMITM 版本信息%s\n", ColorBold, ColorCyan, ColorReset)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("%s版本:%s     %s%s%s\n", ColorBold, ColorReset, ColorGreen, Version, ColorReset)
	fmt.Printf("%s构建时间:%s %s%s%s\n", ColorBold, ColorReset, ColorYellow, BuildTime, ColorReset)
	fmt.Printf("%sGit提交:%s  %s%s%s\n", ColorBold, ColorReset, ColorBlue, GitCommit, ColorReset)
	fmt.Printf("%sGo版本:%s   %s%s%s\n", ColorBold, ColorReset, ColorPurple, runtime.Version(), ColorReset)
	fmt.Printf("%s操作系统:%s %s%s%s\n", ColorBold, ColorReset, ColorCyan, runtime.GOOS, ColorReset)
	fmt.Printf("%s架构:%s     %s%s%s\n", ColorBold, ColorReset, ColorWhite, runtime.GOARCH, ColorReset)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
}

// 打印函数
func printInfo(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s[INFO]%s %s\n",
		ColorBlue, timestamp, ColorReset,
		ColorCyan, ColorReset,
		fmt.Sprintf(format, args...))
}

func printSuccess(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s[SUCCESS]%s %s\n",
		ColorBlue, timestamp, ColorReset,
		ColorGreen, ColorReset,
		fmt.Sprintf(format, args...))
}

func printWarning(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s[WARNING]%s %s\n",
		ColorBlue, timestamp, ColorReset,
		ColorYellow, ColorReset,
		fmt.Sprintf(format, args...))
}

func printError(format string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Printf("%s[%s]%s %s[ERROR]%s %s\n",
		ColorBlue, timestamp, ColorReset,
		ColorRed, ColorReset,
		fmt.Sprintf(format, args...))
}

func printSeparator() {
	fmt.Printf("%s%s%s\n", ColorBlue, strings.Repeat("─", 60), ColorReset)
}

// initLogger 初始化日志系统
func initLogger() error {
	// 如果指定了日志级别，则设置
	if *logLevel != "" {
		switch *logLevel {
		case "debug":
			logger.DefaultLogger.SetLevel(logger.DebugLevel)
		case "info":
			logger.DefaultLogger.SetLevel(logger.InfoLevel)
		case "warn":
			logger.DefaultLogger.SetLevel(logger.WarnLevel)
		case "error":
			logger.DefaultLogger.SetLevel(logger.ErrorLevel)
		}
	}

	return nil
}

// loadConfig 加载配置文件
func loadConfig(configPath string) (*config.Config, error) {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 获取绝对路径
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("获取配置文件绝对路径失败: %w", err)
	}

	// 加载配置
	cfg, err := config.LoadConfig(absPath)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 配置加载成功，无需额外验证

	return cfg, nil
}

// initCertManager 初始化证书管理器
func initCertManager(cfg *config.Config) (*cert.CertManager, error) {
	tlsConfig := cfg.GetTLS()

	// 确保证书目录存在
	if err := os.MkdirAll(tlsConfig.CertDir, 0755); err != nil {
		return nil, fmt.Errorf("创建证书目录失败: %w", err)
	}

	// 创建证书管理器
	certMgr, err := cert.NewCertManager(cert.CertOptions{
		CertDir:     tlsConfig.CertDir,
		EnableCache: true,
		CacheTTL:    24 * time.Hour,
	})
	if err != nil {
		return nil, fmt.Errorf("创建证书管理器失败: %w", err)
	}

	return certMgr, nil
}

// loadPlugins 加载插件
func loadPlugins(server *proxy.Server, cfg *config.Config) error {
	pluginsConfig := cfg.GetPlugins()
	if !pluginsConfig.Enabled {
		printInfo("🔌 插件系统已禁用")
		return nil
	}

	loadedCount := 0
	// 加载配置中的插件
	for _, pluginCfg := range pluginsConfig.Plugins {
		if !pluginCfg.Enabled {
			printInfo("⏭️  跳过已禁用的插件: %s", pluginCfg.Name)
			continue
		}

		printInfo("📦 正在加载插件: %s", pluginCfg.Name)
		// 转换配置类型
		pluginConfig := &plugin.PluginConfig{
			Name:     pluginCfg.Name,
			Enabled:  pluginCfg.Enabled,
			Path:     pluginCfg.Path,
			Config:   pluginCfg.Config,
			Priority: pluginCfg.Priority,
		}
		if err := server.GetPluginManager().LoadPlugin(pluginConfig); err != nil {
			printWarning("❌ 插件 %s 加载失败: %v", pluginCfg.Name, err)
			continue
		}
		printSuccess("✅ 插件 %s 加载成功", pluginCfg.Name)
		loadedCount++
	}

	// 启动所有插件
	if loadedCount > 0 {
		if err := server.StartPlugins(); err != nil {
			return fmt.Errorf("启动插件失败: %w", err)
		}
		printSuccess("🚀 已启动 %d 个插件", loadedCount)
	} else {
		printInfo("📦 没有可用的插件")
	}

	return nil
}

// waitForShutdown 等待关闭信号并优雅关闭
func waitForShutdown(ctx context.Context, server *proxy.Server) {
	// 创建信号通道
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	// 等待信号
	select {
	case sig := <-sigChan:
		printInfo("📨 收到信号: %v，开始优雅关闭...", sig)
	case <-ctx.Done():
		printInfo("📨 收到上下文取消信号，开始关闭...")
	}

	// 创建关闭超时上下文
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 优雅关闭服务器
	printInfo("🛑 正在关闭代理服务器...")
	if err := server.Stop(); err != nil {
		printError("❌ 关闭服务器时出错: %v", err)
	} else {
		printSuccess("✅ 代理服务器已关闭")
	}

	// 等待所有 goroutine 完成或超时
	done := make(chan struct{})
	go func() {
		// 这里可以添加额外的清理逻辑
		time.Sleep(1 * time.Second) // 给其他组件一些时间清理
		close(done)
	}()

	select {
	case <-done:
		printSuccess("✅ 所有组件已安全关闭")
	case <-shutdownCtx.Done():
		printWarning("⏰ 关闭超时，强制退出")
	}
}

// writePidFile 写入 PID 文件
func writePidFile(pidFile string) error {
	pid := os.Getpid()
	pidDir := filepath.Dir(pidFile)

	// 确保目录存在
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		return fmt.Errorf("创建PID文件目录失败: %w", err)
	}

	// 写入 PID
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d\n", pid)), 0644); err != nil {
		return fmt.Errorf("写入PID文件失败: %w", err)
	}

	printInfo("📄 PID文件已创建: %s (PID: %d)", pidFile, pid)
	return nil
}

// runAsDaemon 以守护进程模式运行
func runAsDaemon() error {
	// 检查是否已经是守护进程
	if os.Getppid() == 1 {
		// 已经是守护进程，直接运行主程序
		return runMainProgram()
	}

	// Fork 子进程
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 构建新的参数列表（移除 -daemon 参数）
	args := make([]string, 0, len(os.Args)-1)
	for _, arg := range os.Args[1:] {
		if arg != "-daemon" && arg != "--daemon" {
			args = append(args, arg)
		}
	}

	// 启动子进程
	cmd := exec.Command(executable, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动守护进程失败: %w", err)
	}

	fmt.Printf("守护进程已启动，PID: %d\n", cmd.Process.Pid)
	return nil
}

// runMainProgram 运行主程序逻辑
func runMainProgram() error {
	// 显示启动横幅
	showBanner()

	// 初始化日志系统
	if err := initLogger(); err != nil {
		printError("初始化日志系统失败: %v", err)
		return err
	}

	printInfo("🚀 HackMITM v%s 启动中...", Version)
	printInfo("📅 构建时间: %s", BuildTime)
	printInfo("🔗 Git提交: %s", GitCommit)
	printInfo("🐹 Go版本: %s", runtime.Version())
	printInfo("💻 操作系统: %s/%s", runtime.GOOS, runtime.GOARCH)
	printSeparator()

	// 写入 PID 文件
	if *pidFile != "" {
		if err := writePidFile(*pidFile); err != nil {
			printError("写入PID文件失败: %v", err)
			return err
		}
		defer removePidFile(*pidFile)
	}

	// 加载配置文件
	printInfo("📋 正在加载配置文件...")
	cfg, err := loadConfig(*configFile)
	if err != nil {
		printError("加载配置文件失败: %v", err)
		return err
	}
	printSuccess("✅ 配置文件加载成功")

	// 初始化证书管理器
	printInfo("🔐 正在初始化证书管理器...")
	certMgr, err := initCertManager(cfg)
	if err != nil {
		printError("初始化证书管理器失败: %v", err)
		return err
	}
	printSuccess("✅ 证书管理器初始化成功")

	// 创建代理服务器
	printInfo("🌐 正在创建代理服务器...")
	server, err := proxy.NewServer(cfg, certMgr)
	if err != nil {
		printError("创建代理服务器失败: %v", err)
		return err
	}
	printSuccess("✅ 代理服务器创建成功")

	// 加载插件
	printInfo("🔌 正在加载插件...")
	if err := loadPlugins(server, cfg); err != nil {
		printWarning("⚠️  插件加载失败: %v", err)
	} else {
		printSuccess("✅ 插件加载成功")
	}

	// 启动服务器
	ctx := context.Background()
	if err := server.Start(); err != nil {
		printError("启动服务器失败: %v", err)
		return err
	}

	// 等待中断信号
	waitForShutdown(ctx, server)

	printSuccess("👋 HackMITM 已安全退出")
	return nil
}

// removePidFile 删除 PID 文件
func removePidFile(pidFile string) {
	if err := os.Remove(pidFile); err != nil {
		printWarning("⚠️  删除PID文件失败: %v", err)
	} else {
		printInfo("🗑️  PID文件已删除: %s", pidFile)
	}
}
