# Tauri IPC 桥接

## 职责

封装前端与 Tauri Rust 壳之间的 IPC 通信层。使用 `@tauri-apps/api` v2 提供的 `invoke` 和 `listen` 机制，替代旧架构中的 WebSocket 客户端。

## 核心文件

| 文件路径 | 职责 |
|---------|------|
| `frontend/src/lib/bridge.ts` | Tauri IPC 封装（调用命令 + 监听事件） |

## 前端调用后端命令

```typescript
// bridge.ts
import { invoke } from '@tauri-apps/api/core';
import { type UnlistenFn, listen } from '@tauri-apps/api/event';

// 发送聊天消息（通过 Tauri → sidecar → PetCore）
export async function sendMessage(text: string): Promise<void> {
    await invoke('chat', { text });
}

// 切换鼠标穿透
export async function toggleClickthrough(enabled: boolean): Promise<void> {
    await invoke('toggle_clickthrough', { enabled });
}

// 获取窗口位置
export async function getWindowPosition(): Promise<{ x: number; y: number }> {
    return await invoke('get_window_position');
}

// 设置窗口位置
export async function setWindowPosition(x: number, y: number): Promise<void> {
    await invoke('set_window_position', { x, y });
}

// 获取宠物状态
export async function getStatus(): Promise<Record<string, unknown>> {
    return await invoke('get_status');
}
```

## 事件流

```typescript
// bridge.ts — 监听宠物事件
export interface PetEvent {
  kind: string;
  data: Record<string, unknown>;
}

export function onPetEvent(handler: (event: PetEvent) => void): Promise<UnlistenFn> {
  return listen<PetEvent>('pet:event', (e) => {
    handler(e.payload);
  });
}

// App.tsx — 使用示例
useEffect(() => {
  const unlistenPromise = onPetEvent((event: PetEvent) => {
    if (event.kind === 'message' && typeof event.data.text === 'string') {
      // 显示消息
    }
  });
  return () => { void unlistenPromise.then((unlisten) => { unlisten(); }); };
}, []);
```

## 可用命令表

| 命令 | 参数 | 返回 | 说明 |
|------|------|------|------|
| `chat` | `{ text: string }` | `void` | 发送对话消息 |
| `toggle_clickthrough` | `{ enabled: boolean }` | `void` | 切换鼠标穿透 |
| `get_window_position` | 无 | `{ x: number, y: number }` | 获取窗口位置 |
| `set_window_position` | `{ x: number, y: number }` | `void` | 设置窗口位置 |
| `get_status` | 无 | `Record<string, unknown>` | 获取宠物状态 |

## 事件类型

| 事件名 | 方向 | Payload | 说明 |
|--------|------|---------|------|
| `pet:event` | PetCore → 前端 | `PetEvent` | 宠物事件（说话/动作/状态变化/思考） |
