// 系统托盘 — 菜单栏常驻图标
//
// 提供右键菜单：显示/隐藏、对话、设置、检查更新、退出
//
// ═══ 测试说明 ═══
// create_tray 完全依赖 Tauri AppHandle（菜单构建、托盘图标注册、
// 事件回调），在纯单元测试中无法构造 AppHandle。
// 托盘功能应在集成测试（带真实窗口上下文）或端到端测试中验证。
// 可作为占位：创建临时目录下的 Tauri app、验证菜单项数量和 ID。
use tauri::{
    AppHandle, Emitter, Manager,
    menu::{MenuBuilder, MenuItem, PredefinedMenuItem},
    tray::TrayIconBuilder,
};

/// 创建系统托盘和右键菜单。
///
/// # Errors
///
/// 当菜单构建或托盘创建失败时返回错误。
pub fn create_tray(app: &AppHandle) -> tauri::Result<()> {
    // 构建右键菜单
    let toggle = MenuItem::with_id(app, "toggle", "显示/隐藏宠物", true, None::<&str>)?;
    let chat = MenuItem::with_id(app, "chat", "对话", true, None::<&str>)?;
    let separator = PredefinedMenuItem::separator(app)?;
    let settings = MenuItem::with_id(app, "settings", "设置", true, None::<&str>)?;
    let check_update = MenuItem::with_id(app, "check_update", "检查更新", true, None::<&str>)?;
    let separator2 = PredefinedMenuItem::separator(app)?;
    let quit = MenuItem::with_id(app, "quit", "退出", true, None::<&str>)?;

    let menu = MenuBuilder::new(app)
        .item(&toggle)
        .item(&chat)
        .item(&separator)
        .item(&settings)
        .item(&check_update)
        .item(&separator2)
        .item(&quit)
        .build()?;

    let mut tray_builder = TrayIconBuilder::new().menu(&menu);

    // 如果有默认窗口图标则设置
    if let Some(icon) = app.default_window_icon() {
        tray_builder = tray_builder.icon(icon.clone());
    }

    tray_builder
        .on_menu_event(move |app_handle, event| match event.id().as_ref() {
            "toggle" => {
                if let Some(window) = app_handle.get_webview_window("pet") {
                    if window.is_visible().unwrap_or(true) {
                        let _ = window.hide();
                    } else {
                        let _ = window.show();
                    }
                }
            },
            "settings" => {
                // 触发前端打开设置面板
                let _ = app_handle.emit("pet:open-settings", ());
            },
            "chat" => {
                let _ = app_handle.emit("pet:open-chat", ());
            },
            "check_update" => {
                // 触发检查更新（Phase 2 实现）
                let _ = app_handle.emit("pet:check-update", ());
            },
            "quit" => {
                app_handle.exit(0);
            },
            _ => {},
        })
        .build(app)?;

    Ok(())
}
