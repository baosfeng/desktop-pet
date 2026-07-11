// PetCore — 桌面 AI 宠物内核。
//
// 双模式入口：
//   - sidecar 模式（默认）：通过 stdin/stdout JSON 与父进程通信
//   - CLI 模式（开发/调试）：交互式对话
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/desktop-pet/petcore/internal/agent"
	"github.com/desktop-pet/petcore/internal/config"
	"github.com/desktop-pet/petcore/internal/core"
	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/fsm"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	"github.com/desktop-pet/petcore/internal/log"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/plugin"
	"github.com/desktop-pet/petcore/internal/server"
	"github.com/desktop-pet/petcore/internal/tool"
)

func main() {
	cliMode := flag.Bool("cli", false, "启动 CLI 模式（交互式对话）")
	configPath := flag.String("config", "", "配置文件路径（默认 ~/.desktop-pet/config.toml）")
	envFlag := flag.String("env", "", "运行环境：development / production（默认 auto，从 PETCORE_ENV 环境变量读取）")
	flag.Parse()

	log.InitLogger()

	// 检测运行环境
	runEnv := config.CurrentEnv()
	if *envFlag != "" {
		switch config.Environment(*envFlag) {
		case config.EnvDevelopment, config.EnvProduction:
			runEnv = config.Environment(*envFlag)
		default:
			log.Warn("invalid env flag, falling back to auto-detected", "flag", *envFlag, "detected", runEnv)
		}
	}
	log.Info("PetCore starting", "cli", *cliMode, "env", runEnv)

	// 加载配置
	cfgPath := *configPath
	if cfgPath == "" {
		if d := os.Getenv("PETCORE_DATA_DIR"); d != "" {
			cfgPath = d + "/config.toml"
		} else {
			home, _ := os.UserHomeDir()
			cfgPath = home + "/.desktop-pet/config.toml"
		}
	}
	cfg, err := config.Load(cfgPath, runEnv)
	if err != nil {
		log.Error("failed to load config", "error", err, "path", cfgPath)
		os.Exit(1)
	}

	// 初始化各模块
	engine := buildEngine(cfg)

	// 信号处理
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Info("shutting down")
		cancel()
	}()

	if *cliMode {
		runCLI(ctx, engine)
	} else {
		runSidecar(ctx, engine)
	}
}

func buildEngine(cfg *config.Config) *core.Engine {
	// LLM Provider
	provider, err := llm.NewProvider(cfg.LLM.Provider, map[string]any{
		"model":    cfg.LLM.Model,
		"base_url": cfg.LLM.BaseURL,
	})
	if err != nil {
		log.Error("failed to create LLM provider", "error", err)
		provider, _ = llm.NewProvider("mock", nil) // fallback to mock
	}

	// FSM
	machine := fsm.NewMockMachine(fsm.StateIdle)

	// Memory
	mem := memory.NewInMemoryManager()

	// Tool Registry
	toolReg := tool.NewRegistry()

	// Agent
	ag := agent.New(provider,
		agent.WithMemory(mem),
		agent.WithToolRegistry(toolReg),
	)

	// Plugin Registry
	pluginReg := plugin.NewRegistry()

	// 注入 NoopSink（sidecar 模式下会被 SinkAdapter 替换）
	return core.New(machine, ag, mem, pluginReg, toolReg, cfg, event.NoopSink{})
}

func runSidecar(ctx context.Context, eng *core.Engine) {
	srv := server.New(eng, os.Stdin, os.Stdout)

	// 创建 SinkAdapter 并注入到 engine
	sinkAdapter := server.NewSinkAdapter(srv)
	eng.SetSink(sinkAdapter)

	log.Info("sidecar mode ready")
	if err := srv.Run(ctx); err != nil {
		log.Error("sidecar error", "error", err)
	}
}

func runCLI(ctx context.Context, _ *core.Engine) {
	log.Info("PetCore CLI mode (not yet implemented)")
	log.Info("Run with --cli to enter interactive mode")
	<-ctx.Done()
}
