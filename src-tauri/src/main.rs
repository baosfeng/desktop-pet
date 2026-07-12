// PetCore — Tauri v2 桌面壳入口
//
// 负责：服务子进程管理、窗口管理、系统托盘、事件转发
#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

mod commands;
mod sidecar;
mod tray;
mod window;

#[allow(clippy::exit)]
fn main() {
    env_logger::init();

    if let Err(e) = tauri::Builder::default()
        .plugin(tauri_plugin_shell::init())
        .setup(|app| {
            // 创建并配置宠物窗口
            window::create_pet_window(app)?;

            // 创建系统托盘
            let app_handle = app.handle();
            tray::create_tray(app_handle)?;

            // 异步启动 sidecar 子进程
            let handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                sidecar::start_sidecar(&handle).await;
            });

            Ok(())
        })
        .invoke_handler(tauri::generate_handler![
            commands::toggle_clickthrough,
            commands::get_window_position,
            commands::set_window_position,
            commands::chat,
            commands::get_status,
            commands::log_from_frontend,
        ])
        .run(tauri::generate_context!())
    {
        log::error!("Tauri application error: {e}");
    }
}
