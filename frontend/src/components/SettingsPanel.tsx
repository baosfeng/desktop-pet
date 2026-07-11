import { useCallback, useState } from "react";
import type React from "react";

interface SettingsPanelProps {
  onClose: () => void;
}

interface Settings {
  apiKey: string;
  baseUrl: string;
  modelName: string;
  persona: string;
  opacity: number;
}

const STORAGE_KEY = "desktop-pet-settings";

const DEFAULT_SETTINGS: Settings = {
  apiKey: "",
  baseUrl: "https://api.openai.com/v1",
  modelName: "gpt-4o-mini",
  persona: "你是一只可爱的桌面宠物，性格活泼友善。",
  opacity: 0.9,
};

function loadSettings(): Settings {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw !== null) {
      const parsed = JSON.parse(raw) as Partial<Settings>;
      return { ...DEFAULT_SETTINGS, ...parsed };
    }
  } catch {
    // ignore parse errors, use defaults
  }
  return { ...DEFAULT_SETTINGS };
}

function saveSettings(settings: Settings): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(settings));
}

export function SettingsPanel({ onClose }: SettingsPanelProps): React.JSX.Element {
  const [settings, setSettings] = useState<Settings>(loadSettings);

  const handleChange = useCallback((field: keyof Settings, value: string | number) => {
    setSettings((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleSave = useCallback(() => {
    saveSettings(settings);
    onClose();
  }, [settings, onClose]);

  const handleClose = useCallback(() => {
    // Restore from storage to discard unsaved changes
    setSettings(loadSettings());
    onClose();
  }, [onClose]);

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
            value={settings.apiKey}
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
            value={settings.baseUrl}
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
            value={settings.modelName}
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
            value={settings.persona}
            onChange={(e) => {
              handleChange("persona", e.target.value);
            }}
          />
        </label>

        <label className="settings-field">
          <span className="settings-label">窗口透明度: {Math.round(settings.opacity * 100)}%</span>
          <input
            className="settings-slider"
            type="range"
            min={0.1}
            max={1}
            step={0.05}
            value={settings.opacity}
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
