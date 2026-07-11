// Package event 定义 PetCore 事件类型和 Sink 接口。
//
// 这是整个架构中最基础的模块——零依赖其他 petcore 内部包。
// Shell 层（桌面/CLI/Server）通过实现 Sink 接口来消费 PetCore 输出的事件。
package event

import "time"

// Type 表示事件类型的字符串标识。
type Type string

// 预定义事件类型常量。
const (
	EventStateChanged  Type = "state.changed"  // 状态机切换
	EventPetSpeak      Type = "pet.speak"      // 宠物说话
	EventPetAction     Type = "pet.action"     // 宠物动作
	EventPetEmotion    Type = "pet.emotion"    // 情绪变化
	EventAgentThinking Type = "agent.thinking" // AI 思考中
	EventAgentReply    Type = "agent.reply"    // AI 回复片段
	EventMemoryUpdated Type = "memory.updated" // 记忆更新
	EventError         Type = "error"          // 错误
)

// Event 是系统中传递的标准事件结构。
type Event struct {
	Kind Type
	Data any
	Meta Meta
}

// Meta 包含事件的元信息。
type Meta struct {
	Timestamp int64  // Unix 毫秒时间戳
	Source    string // 事件来源：core / plugin / user
	SessionID string
}

// NewMeta 创建一个带当前时间戳的元信息。
func NewMeta(source, sessionID string) Meta {
	return Meta{
		Timestamp: time.Now().UnixMilli(),
		Source:    source,
		SessionID: sessionID,
	}
}

// Sink 是 Shell 层需要实现的消费者接口。
// 任何 Shell（桌面/CLI/Server）只需要实现 Sink，就能消费 PetCore 的所有事件。
type Sink interface {
	Send(event Event) error
	Close() error
}

// Ensure NoopSink implements Sink.
var _ Sink = (*NoopSink)(nil)

// NoopSink 是一个空实现，用于测试或不需要处理事件的场景。
type NoopSink struct{}

// Send 实现 Sink 接口，空操作返回 nil。
func (NoopSink) Send(Event) error { return nil }

// Close 实现 Sink 接口，空操作返回 nil。
func (NoopSink) Close() error { return nil }
