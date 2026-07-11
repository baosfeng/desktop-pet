# Tauri IPC 桥接

## 职责

封装前端与 Tauri Rust 壳之间的 IPC 通信层。替代旧架构中的 WebSocket 客户端，使用 `@tauri-apps/api` v2 提供的 invoke 和 event 机制。

## 核心文件

| 文件路径 | 职责 |
|---------|------|
| `frontend/src/lib/bridge.ts` | Tauri IPC 封装（调用 Go 命令、监听事件） |
| `frontend/src/hooks/usePetCore.ts` | React hook，封装 invoke 调用 |
| `frontend/src/hooks/usePetEvents.ts` | React hook，封装事件监听 |

## 前端调用后端命令

```typescript
// bridge.ts
import { invoke } from '@tauri-apps/api/core';
import { listen } from '@tauri-apps/api/event';

// 调用 Go PetCore 命令（通过 Tauri → sidecar）
export async function chat(text: string): Promise<void> {
    await invoke('chat', { text });
}

export async function getStatus(): Promise<PetStatus> {
    return await invoke('get_status');
}

export async function queryMemory(query: string): Promise<Memory[]> {
    return await invoke('query_memory', { query });
}

// 监听 PetCore 事件（通过 Tauri → sidecar → event）
export function onPetEvent(handler: (event: PetEvent) => void): () => void {
    return listen('pet:event', (e) => {
        handler(e.payload as PetEvent);
    });
}
```

## 事件流

```typescript
// usePetEvents.ts — 组件中监听事件
function usePetEvents() {
    useEffect(() => {
        const unlisten = onPetEvent((event) => {
            switch (event.kind) {
                case 'pet.speak':
                    setBubbleText(event.data.text);
                    break;
                case 'pet.action':
                    petRef.current?.playAnimation(event.data.action);
                    break;
                case 'pet.emotion':
                    setEmotion(event.data.mood);
                    break;
                case 'agent.thinking':
                    setIsThinking(event.data.status);
                    break;
            }
        });
        return unlisten;
    }, []);
}
```

## 与旧 WebSocket 对比

| 对比项 | 旧 (WebSocket) | 新 (Tauri IPC) |
|--------|---------------|----------------|
| 通信方式 | WebSocket 127.0.0.1 | Tauri invoke + event |
| 延迟 | 毫秒级（TCP） | 微秒级（进程内） |
| 安全性 | 需处理端口冲突 | 零网络暴露 |
| 代码量 | 需管理连接/重连 | invoke 自动处理 |
| 状态管理 | 手动序列化/反序列化 | serde 自动处理 |
