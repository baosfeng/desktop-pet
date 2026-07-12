// commands 模块单元测试
//
// 覆盖 log_from_frontend 日志级别分支、Position 序列化。
// 依赖 AppHandle 的 Tauri command 函数需要集成测试环境，
// 此处仅以 #[ignore] 标记占位。
#![allow(clippy::unwrap_used, clippy::expect_used)]

use super::*;

// ─── log_from_frontend — 纯函数，各日志级别分支 ───

#[test]
fn log_from_frontend_error_level() {
    // 验证 error 级别不 panic；实际日志输出由 env_logger 控制
    super::log_from_frontend("error".into(), "something broke".into());
}

#[test]
fn log_from_frontend_warn_level() {
    super::log_from_frontend("warn".into(), "deprecation notice".into());
}

#[test]
fn log_from_frontend_info_level() {
    super::log_from_frontend("info".into(), "user clicked button".into());
}

#[test]
fn log_from_frontend_default_fallback() {
    // 非 error/warn 的任意值都走 info 分支
    super::log_from_frontend("debug".into(), "trace data".into());
    super::log_from_frontend("".into(), "empty level".into());
    super::log_from_frontend("UNKNOWN".into(), "random level".into());
}

#[test]
fn log_from_frontend_empty_message() {
    super::log_from_frontend("error".into(), String::new());
    super::log_from_frontend("warn".into(), String::new());
    super::log_from_frontend("info".into(), String::new());
}

// 注意：log_from_frontend 是纯函数，不依赖 AppHandle，
// 是 commands 模块中唯一可在单元测试中直接验证的函数。

// ─── Position 结构体序列化 ──────────────────────

#[test]
fn commands_position_serialize() {
    let pos = Position { x: 10.0, y: 20.0 };
    let json = serde_json::to_string(&pos).unwrap();
    assert!(json.contains("10"));
    assert!(json.contains("20"));
}

#[test]
fn commands_position_json_keys() {
    let pos = Position { x: 3.14, y: 2.718 };
    let json = serde_json::to_value(&pos).unwrap();
    assert_eq!(json["x"].as_f64().unwrap(), 3.14);
    assert_eq!(json["y"].as_f64().unwrap(), 2.718);
}

#[test]
fn commands_position_zero() {
    let pos = Position { x: 0.0, y: 0.0 };
    let json = serde_json::to_value(&pos).unwrap();
    assert_eq!(json["x"].as_f64().unwrap(), 0.0);
    assert_eq!(json["y"].as_f64().unwrap(), 0.0);
}

// ─── AppHandle 依赖的 Tauri command（标记为集成测试占位）───

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn resize_window_command_integration() {
    // resize_window(app_handle, w, h) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn toggle_clickthrough_command_integration() {
    // toggle_clickthrough(app_handle, enabled) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn get_window_position_command_integration() {
    // get_window_position(app_handle) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn set_window_position_command_integration() {
    // set_window_position(app_handle, x, y) — 需要 AppHandle
}
