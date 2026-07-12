// PetCore — 桌面 AI 宠物内核。
//
// 双模式入口：
//   - sidecar 模式（默认）：通过 stdin/stdout JSON 与父进程通信
//   - CLI 模式（开发/调试）：交互式对话
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/desktop-pet/petcore/internal/agent"
	"github.com/desktop-pet/petcore/internal/config"
	"github.com/desktop-pet/petcore/internal/core"
	"github.com/desktop-pet/petcore/internal/event"
	"github.com/desktop-pet/petcore/internal/feature"
	"github.com/desktop-pet/petcore/internal/fsm"
	"github.com/desktop-pet/petcore/internal/llm"
	_ "github.com/desktop-pet/petcore/internal/llm/mock"
	_ "github.com/desktop-pet/petcore/internal/llm/ollama"
	_ "github.com/desktop-pet/petcore/internal/llm/openai"
	"github.com/desktop-pet/petcore/internal/log"
	"github.com/desktop-pet/petcore/internal/memory"
	"github.com/desktop-pet/petcore/internal/plugin"
	"github.com/desktop-pet/petcore/internal/server"
	"github.com/desktop-pet/petcore/internal/tool"
	"github.com/desktop-pet/petcore/internal/tool/builtin"
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
	engine := buildEngine(cfg, runEnv)

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

func buildEngine(cfg *config.Config, runEnv config.Environment) *core.Engine {
	// LLM Provider
	provider, err := llm.NewProvider(cfg.LLM.Provider, map[string]any{
		"model":    cfg.LLM.Model,
		"base_url": cfg.LLM.BaseURL,
		"api_key":  cfg.LLM.APIKey(),
	})
	if err != nil {
		log.Error("failed to create LLM provider", "error", err)
		provider, _ = llm.NewProvider("mock", nil) // fallback to mock
	}

	// FSM
	machine := fsm.NewMockMachine(fsm.StateIdle)

	// Memory
	var mem memory.Manager
	if runEnv == config.EnvProduction {
		var err error
		mem, err = memory.NewSQLiteManager("")
		if err != nil {
			log.Error("failed to create SQLite memory, falling back to in-memory", "error", err)
			mem = memory.NewInMemoryManager()
		}
		log.Info("using SQLite persistent memory", "env", runEnv)
	} else {
		mem = memory.NewInMemoryManager()
		log.Info("using in-memory memory", "env", runEnv)
	}

	// Tool Registry
	toolReg := tool.NewRegistry()
	_ = toolReg.Register(builtin.NewRemember(mem))

	// Feature Flags
	flags := feature.New(cfg.FeatureFlags)
	flags.RegisterDefaults()

	// Agent
	ag := agent.New(provider,
		agent.WithMemory(mem),
		agent.WithToolRegistry(toolReg),
		agent.WithFlags(flags),
	)

	// Plugin Registry
	pluginReg := plugin.NewRegistry()

	// 加载 L1 YAML 动作包（如果功能开关开启且目录存在）
	if flags.IsEnabled(feature.FlagL1YAML) && cfg.Plugin.ActionsDir != "" {
		count, err := plugin.LoadYAMLDir(pluginReg, cfg.Plugin.ActionsDir)
		if err != nil {
			log.Warn("failed to load YAML plugins", "error", err)
		} else if count > 0 {
			log.Info("loaded YAML action packs", "count", count)
		}
	}

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

// ─── CLI 模式 ─────────────────────────────────

// cliSink 捕获 Agent 事件并在终端显示。
type cliSink struct {
	mu            sync.Mutex
	replyBuilder  strings.Builder
	streamingDone chan struct{}
}

func (s *cliSink) Send(e event.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch e.Kind {
	case event.EventAgentThinking:
		fmt.Print("🤔 ")
	case event.EventAgentReply:
		if data, ok := e.Data.(map[string]any); ok {
			if text, ok := data["text"].(string); ok {
				s.replyBuilder.WriteString(text)
				fmt.Print(text)
			}
			if done, ok := data["done"].(bool); ok && done {
				fmt.Println()
				close(s.streamingDone)
			}
		}
	case event.EventError:
		if data, ok := e.Data.(map[string]any); ok {
			if msg, ok := data["error"].(string); ok {
				fmt.Fprintf(os.Stderr, "\n❌ 错误: %s\n", msg)
			}
		}
	case event.EventStateChanged:
		if data, ok := e.Data.(map[string]any); ok {
			if state, ok := data["state"].(string); ok {
				fmt.Fprintf(os.Stderr, "\n🐾 状态: %s\n", state)
			}
		}
	}
	return nil
}

func (s *cliSink) Close() error { return nil }

func runCLI(ctx context.Context, eng *core.Engine) {
	fmt.Println("🐾 PetCore CLI - 桌面 AI 宠物交互终端")
	fmt.Println("   输入消息开始对话，输入 /help 查看命令列表")
	fmt.Println()

	sink := &cliSink{streamingDone: make(chan struct{})}
	eng.SetSink(sink)

	// 启动引擎事件循环（goroutine）
	go func() {
		if err := eng.Run(ctx); err != nil {
			log.Error("engine run error", "error", err)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	prompt := "你 > "

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\n再见！👋")
			return
		default:
		}

		fmt.Print(prompt)
		if !scanner.Scan() {
			return
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 处理命令
		if strings.HasPrefix(line, "/") {
			if !handleCLICommand(ctx, line, eng) {
				return
			}
			continue
		}

		// 处理对话
		fmt.Print("🐱 ")
		sink.streamingDone = make(chan struct{})
		sink.mu.Lock()
		sink.replyBuilder.Reset()
		sink.mu.Unlock()

		if err := eng.HandleInput(ctx, line); err != nil {
			fmt.Fprintf(os.Stderr, "\n❌ 处理消息失败: %v\n", err)
		}

		// 等待流式回复完成
		select {
		case <-sink.streamingDone:
		case <-ctx.Done():
			return
		}

		fmt.Println()
	}
}

func handleCLICommand(ctx context.Context, line string, eng *core.Engine) bool {
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return true
	}

	switch parts[0] {
	case "/help":
		fmt.Println("可用命令:")
		fmt.Println("  /help      - 显示此帮助")
		fmt.Println("  /clear     - 清屏")
		fmt.Println("  /status    - 查看引擎状态")
		fmt.Println("  /memory    - 查看短期记忆")
		fmt.Println("  /exit      - 退出")
	case "/clear":
		fmt.Print("\033[H\033[2J")
	case "/status":
		status := eng.GetStatus()
		fmt.Printf("状态: %v\n", status["state"])
		fmt.Printf("插件数: %v\n", status["plugins"])
		fmt.Printf("工具数: %v\n", status["tools"])
	case "/memory":
		msgs := eng.GetShortTerm()
		if len(msgs) == 0 {
			fmt.Println("暂无短期记忆")
		} else {
			fmt.Printf("短期记忆 (%d 条):\n", len(msgs))
			for _, m := range msgs {
				fmt.Printf("  [%s] %s\n", m.Role, m.Content)
			}
		}
	case "/exit", "/quit":
		fmt.Println("再见！👋")
		return false
	default:
		fmt.Printf("未知命令: %s（输入 /help 查看帮助）\n", parts[0])
	}
	return true
}
