// Tauri 命令注册
//
// 将前端可调用的命令与底层实现连接。
use std::time::{SystemTime, UNIX_EPOCH};

use serde::Serialize;
use tauri::{AppHandle, Manager};

use crate::window;

/// 窗口位置结构
#[derive(Debug, Clone, Serialize)]
pub struct Position {
    pub x: f64,
    pub y: f64,
}

/// 切换鼠标穿透状态
#[tauri::command]
#[allow(clippy::needless_pass_by_value)]
pub fn toggle_clickthrough(app_handle: AppHandle, enabled: bool) -> tauri::Result<()> {
    window::toggle_clickthrough(&app_handle, enabled)
}

/// 获取窗口当前位置
#[tauri::command]
#[allow(clippy::needless_pass_by_value)]
pub fn get_window_position(app_handle: AppHandle) -> tauri::Result<Position> {
    window::get_window_position(&app_handle).map(|pos| Position { x: pos.x, y: pos.y })
}

/// 设置窗口位置
#[tauri::command]
#[allow(clippy::needless_pass_by_value)]
pub fn set_window_position(app_handle: AppHandle, x: f64, y: f64) -> tauri::Result<()> {
    window::set_window_position(&app_handle, x, y)
}

/// 发送聊天消息
#[tauri::command]
#[allow(clippy::needless_pass_by_value, clippy::unnecessary_wraps)]
pub fn chat(app_handle: AppHandle, text: String) -> tauri::Result<()> {
    // 通过 sidecar 转发给 PetCore
    if let Some(writer) = app_handle.try_state::<crate::sidecar::SidecarWriter>() {
        let cmd = serde_json::json!({
            "type": "cmd",
            "id": format!("chat-{}", SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_millis()),
            "method": "chat",
            "params": { "text": text }
        });
        writer.send(&cmd.to_string());
    }
    Ok(())
}

/// 获取宠物状态
#[tauri::command]
#[allow(clippy::needless_pass_by_value, clippy::unnecessary_wraps)]
pub fn get_status(app_handle: AppHandle) -> tauri::Result<serde_json::Value> {
    if let Some(writer) = app_handle.try_state::<crate::sidecar::SidecarWriter>() {
        let cmd = serde_json::json!({
            "type": "cmd",
            "id": "status-1",
            "method": "get_status",
            "params": {}
        });
        writer.send(&cmd.to_string());
        return Ok(serde_json::json!({"queued": true}));
    }
    Ok(serde_json::json!({"state": "offline"}))
}

/// 更新 `PetCore` LLM 配置（API Key / Base URL / Model / System Prompt）
#[tauri::command]
#[allow(clippy::needless_pass_by_value, clippy::unnecessary_wraps)]
pub fn update_config(app_handle: AppHandle, config: serde_json::Value) -> tauri::Result<()> {
    if let Some(writer) = app_handle.try_state::<crate::sidecar::SidecarWriter>() {
        let cmd = serde_json::json!({
            "type": "cmd",
            "id": format!("config-{}", SystemTime::now().duration_since(UNIX_EPOCH).unwrap_or_default().as_millis()),
            "method": "update_config",
            "params": config
        });
        writer.send(&cmd.to_string());
    }
    Ok(())
}

/// 从前端接收日志（调试用）
#[tauri::command]
#[allow(clippy::needless_pass_by_value)]
pub fn log_from_frontend(level: String, message: String) {
    match level.as_str() {
        "error" => log::error!("[FE] {message}"),
        "warn" => log::warn!("[FE] {message}"),
        _ => log::info!("[FE] {message}"),
    }
}
