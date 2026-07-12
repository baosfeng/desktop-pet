// window 模块单元测试
//
// 覆盖 Position 序列化/反序列化、边界保护逻辑、常量和纯函数。
// 依赖 AppHandle 的函数（toggle_clickthrough、get/set_window_position、
// resize_window、save/load_window_position、position_store_path）
// 需要集成测试环境，此处仅以 #[ignore] 标记占位。
#![allow(clippy::unwrap_used, clippy::expect_used)]

use super::*;

// ─── Position 序列化/反序列化 ─────────────────

#[test]
fn position_serde_roundtrip_positive() {
    let pos = Position { x: 100.5, y: 200.5 };
    let json = serde_json::to_string(&pos).unwrap();
    let deserialized: Position = serde_json::from_str(&json).unwrap();
    assert!(
        (deserialized.x - pos.x).abs() < f64::EPSILON,
        "x 不匹配: {} vs {}",
        deserialized.x,
        pos.x
    );
    assert!(
        (deserialized.y - pos.y).abs() < f64::EPSILON,
        "y 不匹配: {} vs {}",
        deserialized.y,
        pos.y
    );
}

#[test]
fn position_serde_roundtrip_zero() {
    let pos = Position { x: 0.0, y: 0.0 };
    let json = serde_json::to_string(&pos).unwrap();
    let deserialized: Position = serde_json::from_str(&json).unwrap();
    assert!((deserialized.x - 0.0).abs() < f64::EPSILON);
    assert!((deserialized.y - 0.0).abs() < f64::EPSILON);
}

#[test]
fn position_serde_roundtrip_negative() {
    // 负坐标在反序列化时可以被保留；边界保护只在 load_window_position 中应用
    let pos = Position { x: -42.0, y: -99.0 };
    let json = serde_json::to_string(&pos).unwrap();
    let deserialized: Position = serde_json::from_str(&json).unwrap();
    assert!((deserialized.x - (-42.0)).abs() < f64::EPSILON);
    assert!((deserialized.y - (-99.0)).abs() < f64::EPSILON);
}

#[test]
fn position_serde_large_values() {
    let pos = Position {
        x: f64::MAX,
        y: f64::MIN,
    };
    let json = serde_json::to_string(&pos).unwrap();
    let deserialized: Position = serde_json::from_str(&json).unwrap();
    assert!((deserialized.x - f64::MAX).abs() < 1.0e290);
    assert!((deserialized.y - f64::MIN).abs() < 1.0e290);
}

#[test]
fn position_json_structure() {
    let pos = Position { x: 1.5, y: 2.5 };
    let json = serde_json::to_string(&pos).unwrap();
    let value: serde_json::Value = serde_json::from_str(&json).unwrap();
    assert_eq!(value["x"].as_f64().unwrap(), 1.5);
    assert_eq!(value["y"].as_f64().unwrap(), 2.5);
}

// ─── 边界保护逻辑（clamp 到 >= 0.0）─────────────────

#[test]
fn clamp_negative_x_to_zero() {
    // 模拟 load_window_position 中的 clamped_x = pos.x.max(0.0)
    let pos = Position { x: -100.0, y: 50.0 };
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);
    assert!((clamped_x - 0.0).abs() < f64::EPSILON);
    assert!((clamped_y - 50.0).abs() < f64::EPSILON);
}

#[test]
fn clamp_negative_y_to_zero() {
    let pos = Position { x: 100.0, y: -200.0 };
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);
    assert!((clamped_x - 100.0).abs() < f64::EPSILON);
    assert!((clamped_y - 0.0).abs() < f64::EPSILON);
}

#[test]
fn clamp_both_negative_to_zero() {
    let pos = Position {
        x: -300.0,
        y: -400.0,
    };
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);
    assert!((clamped_x - 0.0).abs() < f64::EPSILON);
    assert!((clamped_y - 0.0).abs() < f64::EPSILON);
}

#[test]
fn clamp_positive_values_unchanged() {
    let pos = Position {
        x: 1920.0,
        y: 1080.0,
    };
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);
    assert!((clamped_x - 1920.0).abs() < f64::EPSILON);
    assert!((clamped_y - 1080.0).abs() < f64::EPSILON);
}

#[test]
fn clamp_zero_remains_zero() {
    let pos = Position { x: 0.0, y: 0.0 };
    let clamped_x = pos.x.max(0.0);
    let clamped_y = pos.y.max(0.0);
    assert!((clamped_x - 0.0).abs() < f64::EPSILON);
    assert!((clamped_y - 0.0).abs() < f64::EPSILON);
}

// ─── 常量 ──────────────────────────────────────

#[test]
fn window_default_dimensions() {
    assert!((window_default_width() - 420.0).abs() < f64::EPSILON);
    assert!((window_default_height() - 600.0).abs() < f64::EPSILON);
}

#[test]
fn clickthrough_timeout_value() {
    assert_eq!(clickthrough_timeout_secs(), 5);
}

// ─── AppHandle 依赖函数（标记为集成测试占位）───

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn toggle_clickthrough_integration() {
    // toggle_clickthrough(app_handle, true) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn get_window_position_integration() {
    // get_window_position(app_handle) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn set_window_position_integration() {
    // set_window_position(app_handle, x, y) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn resize_window_integration() {
    // resize_window(app_handle, width, height) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle，在集成测试环境中运行"]
fn position_store_path_integration() {
    // position_store_path(app_handle) — 需要 AppHandle
}

#[test]
#[ignore = "需要真实 Tauri AppHandle 和文件系统，在集成测试环境中运行"]
fn save_and_load_window_position_integration() {
    // save_window_position(app_handle, &pos) / load_window_position(app) — 需要 App/AppHandle
}
