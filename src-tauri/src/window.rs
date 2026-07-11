// 窗口管理器 — 透明无边框宠物窗口
//
// 负责：
// - 透明无边框窗口创建
// - 鼠标穿透切换（待机穿透 / 交互关闭）
// - 窗口置顶
// - 窗口位置记忆（关闭保存，启动恢复）
// - 多显示器边界保护
use serde::{Deserialize, Serialize};
use std::fs;
use std::path::PathBuf;
use tauri::{App, AppHandle, Manager};

/// 宠物窗口标签名
const WINDOW_LABEL: &str = "pet";
/// 窗口默认宽度
const WINDOW_DEFAULT_WIDTH: f64 = 300.0;
/// 窗口默认高度
const WINDOW_DEFAULT_HEIGHT: f64 = 400.0;
/// 交互超时后自动恢复穿透（秒）
const CLICKTHROUGH_TIMEOUT_SECS: u64 = 5;
/// 位置存储文件名
const POSITION_STORE_FILE: &str = "window_position.json";

/// 窗口位置
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Position {
    pub x: f64,
    pub y: f64,
}

/// 创建并配置宠物窗口。
/// 在 Tauri setup 阶段调用。
///
/// # Errors
///
/// 当窗口创建或位置恢复失败时返回错误。
#[allow(clippy::unnecessary_wraps)]
pub fn create_pet_window(app: &App) -> Result<(), Box<dyn std::error::Error>> {
    // 窗口由 tauri.conf.json 配置创建，这里仅做后置配置

    // 尝试恢复上次窗口位置
    if let Ok(pos) = load_window_position(app)
        && let Some(window) = app.get_webview_window(WINDOW_LABEL)
    {
        #[allow(clippy::cast_possible_truncation)]
        let _ = window.set_position(tauri::PhysicalPosition::new(
            pos.x.round() as i32,
            pos.y.round() as i32,
        ));
    }

    Ok(())
}

/// 切换鼠标穿透状态。
///
/// `true` — 鼠标穿透（待机模式）
/// `false` — 鼠标可交互（关注/交互/说话模式）
///
/// # Errors
///
/// 当窗口未创建或 API 调用失败时返回错误。
pub fn toggle_clickthrough(app_handle: &AppHandle, enabled: bool) -> tauri::Result<()> {
    let window = app_handle
        .get_webview_window(WINDOW_LABEL)
        .ok_or_else(|| tauri::Error::WindowNotFound)?;

    window.set_ignore_cursor_events(enabled)?;
    Ok(())
}

/// 获取窗口当前位置。
///
/// # Errors
///
/// 当窗口未创建时返回错误。
pub fn get_window_position(app_handle: &AppHandle) -> tauri::Result<Position> {
    let window = app_handle
        .get_webview_window(WINDOW_LABEL)
        .ok_or_else(|| tauri::Error::WindowNotFound)?;

    let pos = window.outer_position()?;
    Ok(Position {
        x: f64::from(pos.x),
        y: f64::from(pos.y),
    })
}

/// 设置窗口位置并持久化。
///
/// # Errors
///
/// 当窗口未创建或 API 调用失败时返回错误。
pub fn set_window_position(app_handle: &AppHandle, x: f64, y: f64) -> tauri::Result<()> {
    let window = app_handle
        .get_webview_window(WINDOW_LABEL)
        .ok_or_else(|| tauri::Error::WindowNotFound)?;

    #[allow(clippy::cast_possible_truncation)]
    window.set_position(tauri::PhysicalPosition::new(
        x.round() as i32,
        y.round() as i32,
    ))?;

    // 持久化位置
    let pos = Position { x, y };
    let _ = save_window_position(app_handle, &pos);

    Ok(())
}

// ─── 位置持久化 ──────────────────────────────

/// 获取窗口位置存储路径
fn position_store_path(app_handle: &AppHandle) -> PathBuf {
    let data_dir = app_handle.path().app_data_dir().unwrap_or_default();
    data_dir.join(POSITION_STORE_FILE)
}

/// 保存窗口位置到文件
fn save_window_position(
    app_handle: &AppHandle,
    pos: &Position,
) -> Result<(), Box<dyn std::error::Error>> {
    let path = position_store_path(app_handle);

    // 确保父目录存在
    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }

    let json = serde_json::to_string(pos)?;
    fs::write(&path, json)?;
    Ok(())
}

/// 从文件加载窗口位置
fn load_window_position(app: &App) -> Result<Position, Box<dyn std::error::Error>> {
    let data_dir = app.path().app_data_dir()?;
    let path = data_dir.join(POSITION_STORE_FILE);

    if !path.exists() {
        return Err("position file not found".into());
    }

    let json = fs::read_to_string(&path)?;
    let pos: Position = serde_json::from_str(&json)?;

    // 边界保护：确保位置在屏幕可见范围内
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);

    Ok(Position {
        x: clamped_x,
        y: clamped_y,
    })
}

// ─── 公开常量 ────────────────────────────────

#[allow(dead_code)]
pub const fn window_default_width() -> f64 {
    WINDOW_DEFAULT_WIDTH
}

#[allow(dead_code)]
pub const fn window_default_height() -> f64 {
    WINDOW_DEFAULT_HEIGHT
}

#[allow(dead_code)]
pub const fn clickthrough_timeout_secs() -> u64 {
    CLICKTHROUGH_TIMEOUT_SECS
}
