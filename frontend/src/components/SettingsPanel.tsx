import { useCallback, useEffect, useState } from "react";
import type React from "react";

import { usePetStore } from "@/stores/petStore";

interface SettingsPanelProps {
  onClose: () => void;
}

export function SettingsPanel({ onClose }: SettingsPanelProps): React.JSX.Element {
  const storeSettings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);

  const [form, setForm] = useState({ ...storeSettings });

  // 当 store 设置变化时同步到表单
  useEffect(() => {
    setForm({ ...storeSettings });
  }, [storeSettings]);

  const handleChange = useCallback(
    (field: keyof typeof form, value: string | number) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleSave = useCallback(() => {
    updateSettings(form);
    saveSettings();
    onClose();
  }, [form, updateSettings, saveSettings, onClose]);

  const handleClose = useCallback(() => {
    setForm({ ...storeSettings });
    onClose();
  }, [storeSettings, onClose]);

  return (
    <div className="settings-overlay">
      <div className="settings-panel">
        <h2 className="settings-title">设置</h2>

        <label className="settings-field">
          <span className="settings-label">API Key</span>
          <input
            className="settings-input"
            type="password"
            placeholder="sk-..."
            value={form.apiKey}
            onChange={(e) => {
              handleChange("apiKey", e.target.value);
            }}
          />
        </label>

        <label className="settings-field">
          <span className="settings-label">Base URL</span>
          <input
            className="settings-input"
            type="text"
            placeholder="https://api.openai.com/v1"
            value={form.baseUrl}
            onChange={(e) => {
              handleChange("baseUrl", e.target.value);
            }}
          />
        </label>

        <label className="settings-field">
          <span className="settings-label">模型名</span>
          <input
            className="settings-input"
            type="text"
            placeholder="gpt-4o-mini"
            value={form.modelName}
            onChange={(e) => {
              handleChange("modelName", e.target.value);
            }}
          />
        </label>

        <label className="settings-field">
          <span className="settings-label">角色人设</span>
          <textarea
            className="settings-textarea"
            placeholder="描述宠物的性格…"
            rows={4}
            value={form.persona}
            onChange={(e) => {
              handleChange("persona", e.target.value);
            }}
          />
        </label>

        <label className="settings-field">
          <span className="settings-label">窗口透明度: {Math.round(form.opacity * 100)}%</span>
          <input
            className="settings-slider"
            type="range"
            min={0.1}
            max={1}
            step={0.05}
            value={form.opacity}
            onChange={(e) => {
              handleChange("opacity", Number.parseFloat(e.target.value));
            }}
          />
        </label>

        <div className="settings-actions">
          <button className="settings-btn settings-btn--save" onClick={handleSave} type="button">
            保存
          </button>
          <button className="settings-btn settings-btn--cancel" onClick={handleClose} type="button">
            取消
          </button>
        </div>
      </div>
    </div>
  );
}
