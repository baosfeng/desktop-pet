// Sidecar 管理 — Go PetCore 子进程生命周期
//
// 负责：
// - 启动 petcore sidecar 子进程
// - 从 stdout 读取 JSON 事件并转发给前端
// - 发送 ping 健康检查
// - 自动重启（最多 3 次）
// CodeQL's rust/unused-variable does not recognize Rust named capture format {e},
// so use _e/_status prefix to avoid false positives.
#![allow(clippy::used_underscore_binding)]


use std::sync::Mutex;

use log::{error, info, warn};
use serde::Serialize;
use tauri::{AppHandle, Emitter, Manager};
use tauri_plugin_shell::ShellExt;
use tauri_plugin_shell::process::CommandEvent;

/// Sidecar 的 stdin 写入器，供 commands.rs 使用
pub struct SidecarWriter {
    writer: Mutex<Option<tauri_plugin_shell::process::CommandChild>>,
}

impl SidecarWriter {
    /// 创建新的 [`SidecarWriter`]
    #[must_use]
    pub const fn new() -> Self {
        Self {
            writer: Mutex::new(None),
        }
    }

    /// 设置子进程写入器
    pub fn set_writer(&self, child: tauri_plugin_shell::process::CommandChild) {
        if let Ok(mut guard) = self.writer.lock() {
            *guard = Some(child);
        }
    }

    /// 发送 JSON 命令到 sidecar stdin
    pub fn send(&self, json: &str) {
        if let Ok(mut guard) = self.writer.lock()
            && let Some(ref mut child) = *guard
        {
            let _ = child.write(json.as_bytes());
            let _ = child.write(b"\n");
        }
    }
}

impl Default for SidecarWriter {
    fn default() -> Self {
        Self::new()
    }
}

/// Sidecar 事件（从 stdout 读取的 JSON 行）
#[derive(Debug, Serialize)]
struct SidecarEvent {
    kind: String,
    data: serde_json::Value,
}

/// 启动 sidecar 子进程并管理其生命周期。
pub async fn start_sidecar(handle: &AppHandle) {
    let writer = SidecarWriter::new();

    // 注册 SidecarWriter 到 Tauri 状态
    handle.manage(writer);

    // 最大重启次数
    let max_retries: u32 = 3;

    for attempt in 0..max_retries {
        info!(
            "starting petcore sidecar (attempt {}/{})",
            attempt + 1,
            max_retries
        );

        // 创建 sidecar 命令
        let shell = handle.shell();
        let sidecar_cmd = match shell.sidecar("petcore") {
            Ok(cmd) => cmd,
            Err(_e) => {
                error!("failed to create sidecar command: {_e}");
                if attempt < max_retries - 1 {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                }
                continue;
            },
        };

        // sidecar 默认模式，无需额外参数
        let (mut rx, child) = match sidecar_cmd.spawn() {
            Ok(spawned) => spawned,
            Err(_e) => {
                error!("failed to spawn sidecar: {_e}");
                if attempt < max_retries - 1 {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                }
                continue;
            },
        };

        // 保存子进程写入器
        if let Some(state) = handle.try_state::<SidecarWriter>() {
            state.set_writer(child);
        }

        // 发送启动后第一个 ping
        if let Some(state) = handle.try_state::<SidecarWriter>() {
            state.send(r#"{"type":"cmd","id":"ping-1","method":"ping","params":{}}"#);
        }

        // 读取 stdout 事件并转发到前端
        let handle_clone = handle.clone();
        let forward_handle = tokio::spawn(async move {
            loop {
                match rx.recv().await {
                    Some(CommandEvent::Stdout(line)) => {
                        let line_str = String::from_utf8_lossy(&line);
                        let trimmed = line_str.trim();
                        if trimmed.is_empty() {
                            continue;
                        }

                        // 解析 JSON 事件
                        match serde_json::from_str::<serde_json::Value>(trimmed) {
                            Ok(event_value) => {
                                let kind = event_value
                                    .get("event")
                                    .and_then(|v| v.as_str())
                                    .unwrap_or("unknown")
                                    .to_string();
                                let data = event_value
                                    .get("data")
                                    .cloned()
                                    .unwrap_or(serde_json::Value::Null);

                                let pet_event = SidecarEvent { kind, data };
                                let _ = handle_clone.emit("pet:event", &pet_event);
                            },
                            Err(_e) => {
                                warn!("failed to parse sidecar event: {_e}");
                            },
                        }
                    },
                    Some(CommandEvent::Stderr(line)) => {
                        let line_str = String::from_utf8_lossy(&line);
                        warn!("petcore stderr: {}", line_str.trim());
                    },
                    Some(CommandEvent::Terminated(_status)) => {
                        info!("petcore sidecar terminated: {_status:?}");
                        break;
                    },
                    None => {
                        // channel closed
                        break;
                    },
                    _ => {},
                }
            }
        });

        // 等待 sidecar 事件流结束（sidecar 退出时前向处理线程退出）
        let result = forward_handle.await;

        match result {
            Ok(()) => {
                info!("petcore sidecar exited normally");
                return;
            },
            Err(_e) => {
                error!("petcore sidecar forward error: {_e}");
                if attempt < max_retries - 1 {
                    tokio::time::sleep(std::time::Duration::from_secs(1)).await;
                }
            },
        }
    }

    error!("petcore sidecar failed after {max_retries} attempts");
}
