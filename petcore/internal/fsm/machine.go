// Package fsm 定义宠物行为状态机。
//
// FSM 是纯逻辑模块，仅依赖 event 包输出状态变更事件。
// 不依赖 agent / memory / config 等其他 petcore 模块。
package fsm

import (
	"sync"

	"github.com/desktop-pet/petcore/internal/event"
)

// State 表示宠物的行为状态。
type State string

// 预定义状态常量。
const (
	StateIdle        State = "idle"        // 待机 — 鼠标穿透开启，随机小动作
	StateAttention   State = "attention"   // 关注 — 检测到用户，转向鼠标
	StateInteraction State = "interaction" // 交互 — 对话/点击互动中
	StateSpeaking    State = "speaking"    // 说话 — AI 回复播放中
)

// Machine 是状态机接口，所有 Shell 无关的 FSM 实现都通过此接口使用。
type Machine interface {
	// Current 返回当前状态。
	Current() State

	// Transition 尝试根据事件类型进行一次状态转换。
	// 如果转换非法，返回 ErrTransitionNotAllowed。
	Transition(evt event.Type) error

	// OnTransition 注册状态转换回调，用于状态变更后的副作用（如输出事件、切换窗口模式）。
	OnTransition(fn func(from, to State))
}

// ErrTransitionNotAllowed 表示非法的状态转换。
type ErrTransitionNotAllowed struct {
	From State
	Evt  event.Type
}

func (e *ErrTransitionNotAllowed) Error() string {
	return "fsm: transition not allowed from " + string(e.From) + " on " + string(e.Evt)
}

// ─── 内置状态转换表 ────────────────────────────

// transitionRules 定义了状态机转换规则。
// 格式: [当前状态][事件] = 下一状态
var transitionRules = map[State]map[event.Type]State{
	StateIdle: {
		event.EventStateChanged: StateAttention, // 检测到用户
		event.EventPetSpeak:     StateSpeaking,  // 直接触发说话
	},
	StateAttention: {
		event.EventStateChanged: StateInteraction, // 用户对话/点击
		event.EventPetSpeak:     StateSpeaking,    // 说话
		event.EventError:        StateIdle,        // 超时回到待机
	},
	StateInteraction: {
		event.EventAgentReply: StateSpeaking, // AI 开始回复
		event.EventError:      StateIdle,     // 出错回到待机
	},
	StateSpeaking: {
		event.EventStateChanged: StateIdle, // 说话结束
	},
}

// IsValidTransition 检查从 from 状态经过 evt 事件是否能到达 to 状态。
func IsValidTransition(from State, evt event.Type) bool {
	_, ok := transitionRules[from][evt]
	return ok
}

// TransitionsFrom 返回从指定状态出发的所有合法事件。
func TransitionsFrom(s State) []event.Type {
	var evts []event.Type
	for evt := range transitionRules[s] {
		evts = append(evts, evt)
	}
	return evts
}

// ─── MockMachine — 供 main.go 和其他模块使用 ──────

// MockMachine 是 Machine 接口的简单内存实现。
type MockMachine struct {
	mu      sync.RWMutex
	current State
	onFn    func(from, to State)
}

// Current 返回当前状态（线程安全）。
func (m *MockMachine) Current() State {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current
}

// Transition 尝试根据事件类型进行状态转换。
// 如果转换合法，更新当前状态并调用已注册的回调。
// 如果转换非法，返回 ErrTransitionNotAllowed。
func (m *MockMachine) Transition(evt event.Type) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	next, ok := transitionRules[m.current][evt]
	if !ok {
		return &ErrTransitionNotAllowed{From: m.current, Evt: evt}
	}
	from := m.current
	m.current = next
	if m.onFn != nil {
		m.onFn(from, next)
	}
	return nil
}

// OnTransition 注册状态转换回调（线程安全）。
func (m *MockMachine) OnTransition(fn func(from, to State)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onFn = fn
}

// NewMockMachine 创建一个从指定初始状态开始的 MockMachine。
func NewMockMachine(initial State) *MockMachine {
	return &MockMachine{current: initial}
}
