// Package server 提供 PetCore 的 sidecar 通信层。
//
// 通过 stdin/stdout JSON 与 Tauri 壳或其他父进程通信。
// 接收命令（cmd），返回响应（resp）和推送事件（event）。
package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"

	"github.com/desktop-pet/petcore/internal/core"
	"github.com/desktop-pet/petcore/internal/event"
)

// ─── Wire 协议类型 ──────────────────────────

// Command 是父进程发给 PetCore 的命令。
type Command struct {
	Type   string          `json:"type"`            // "cmd"
	ID     string          `json:"id"`              // 请求 ID
	Method string          `json:"method"`           // chat / get_status / ping
	Params json.RawMessage `json:"params,omitempty"` // 参数
}

// Response 是 PetCore 返回给父进程的响应。
type Response struct {
	Type   string `json:"type"`            // "resp"
	ID     string `json:"id"`              // 对应 Command.ID
	Result any    `json:"result,omitempty"` // 成功结果
	Error  string `json:"error,omitempty"` // 错误信息
}

// WireEvent 是 PetCore 主动推送给父进程的事件。
type WireEvent struct {
	Type  string `json:"type"`  // "event"
	Event string `json:"event"` // 事件名称（pet.speak / state.changed 等）
	Data  any    `json:"data"`  // 事件数据
}

// ─── Server ─────────────────────────────────

// Server 处理 sidecar 通信。
// 从 stdin 读取命令，将响应写入 stdout，通过 Sink 输出事件。
type Server struct {
	engine *core.Engine
	reader io.Reader
	writer io.Writer
	log    *slog.Logger
}

// New 创建一个新的 Server 实例。
func New(engine *core.Engine, reader io.Reader, writer io.Writer) *Server {
	return &Server{
		engine: engine,
		reader: reader,
		writer: writer,
		log:    slog.Default(),
	}
}

// Run 启动 sidecar 通信循环，阻塞直到上下文取消或 stdin 关闭。
func (s *Server) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(s.reader)
	// 增大 buffer 以支持大消息
	scanner.Buffer(make([]byte, 1024*64), 1024*1024)

	// 启动引擎（异步）
	engineCtx, engineCancel := context.WithCancel(ctx)
	defer engineCancel()
	go func() {
		if err := s.engine.Run(engineCtx); err != nil {
			s.log.Error("engine stopped", "error", err)
		}
	}()

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var cmd Command
		if err := json.Unmarshal([]byte(line), &cmd); err != nil {
			s.writeError("", fmt.Sprintf("invalid command: %v", err))
			continue
		}

		s.handleCommand(ctx, cmd)
	}

	return scanner.Err()
}

func (s *Server) handleCommand(ctx context.Context, cmd Command) {
	switch cmd.Method {
	case "ping":
		s.writeResult(cmd.ID, map[string]string{"pong": "ok"})

	case "chat":
		var params struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal(cmd.Params, &params); err != nil {
			s.writeError(cmd.ID, fmt.Sprintf("invalid params: %v", err))
			return
		}
		if err := s.engine.HandleInput(ctx, params.Text); err != nil {
			s.writeError(cmd.ID, err.Error())
			return
		}
		s.writeResult(cmd.ID, map[string]bool{"done": true})

	case "get_status":
		status := s.engine.GetStatus()
		s.writeResult(cmd.ID, status)

	default:
		s.writeError(cmd.ID, fmt.Sprintf("unknown method: %s", cmd.Method))
	}
}

func (s *Server) writeResult(id string, result any) {
	resp := Response{Type: "resp", ID: id, Result: result}
	s.writeJSON(resp)
}

func (s *Server) writeError(id string, msg string) {
	resp := Response{Type: "resp", ID: id, Error: msg}
	s.writeJSON(resp)
}

func (s *Server) writeJSON(v any) {
	data, err := json.Marshal(v)
	if err != nil {
		s.log.Error("failed to marshal response", "error", err)
		return
	}
	data = append(data, '\n')
	_, _ = s.writer.Write(data)
}

// SinkAdapter 将 Server 的 writeJSON 适配为 Sink 接口，
// 使得 PetCore 的事件可以通过 sidecar 推送给父进程。
type SinkAdapter struct {
	server *Server
}

func NewSinkAdapter(s *Server) *SinkAdapter {
	return &SinkAdapter{server: s}
}

func (a *SinkAdapter) Send(e event.Event) error {
	we := WireEvent{
		Type:  "event",
		Event: string(e.Kind),
		Data:  e.Data,
	}
	a.server.writeJSON(we)
	return nil
}

func (a *SinkAdapter) Close() error { return nil }
