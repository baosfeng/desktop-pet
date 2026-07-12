import { AnimatePresence, motion } from "motion/react";
import { useCallback, useState } from "react";
import type React from "react";

import { usePetStore } from "@/stores/petStore";
import { updateConfig } from "@/lib/bridge";

/* ─── Provider 预设配置 ────────────────────── */

interface ModelOption {
  id: string;
  label: string;
}

interface ProviderOption {
  id: string;
  label: string;
  baseUrl: string;
  models: ModelOption[];
}

const PROVIDERS: ProviderOption[] = [
  {
    id: "deepseek",
    label: "DeepSeek",
    baseUrl: "https://api.deepseek.com/v1",
    models: [
      { id: "deepseek-chat", label: "DeepSeek Chat (V3)" },
      { id: "deepseek-reasoner", label: "DeepSeek Reasoner (R1)" },
    ],
  },
  {
    id: "openai",
    label: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    models: [
      { id: "gpt-4o-mini", label: "GPT-4o Mini" },
      { id: "gpt-4o", label: "GPT-4o" },
      { id: "gpt-4-turbo", label: "GPT-4 Turbo" },
    ],
  },
  {
    id: "siliconflow",
    label: "硅基流动",
    baseUrl: "https://api.siliconflow.cn/v1",
    models: [
      { id: "deepseek-ai/DeepSeek-V3", label: "DeepSeek V3" },
      { id: "deepseek-ai/DeepSeek-R1", label: "DeepSeek R1" },
      { id: "Qwen/Qwen2.5-7B-Instruct", label: "Qwen2.5-7B" },
      { id: "Qwen/Qwen2.5-14B-Instruct", label: "Qwen2.5-14B" },
    ],
  },
  {
    id: "ollama",
    label: "Ollama（本地）",
    baseUrl: "http://localhost:11434/v1",
    models: [
      { id: "llama3.2", label: "Llama 3.2" },
      { id: "qwen2.5", label: "Qwen 2.5" },
      { id: "mistral", label: "Mistral" },
    ],
  },
];

/* ─── 通用样式 ──────────────────────────────── */

const inputClass =
  "w-full px-3 py-2 border border-soft-brown/40 rounded-[8px] bg-cream text-text-brown text-[13px] outline-none focus:border-primary/60 transition-colors placeholder:text-text-brown/30 font-sans";

const selectClass =
  "w-full px-3 py-2 border border-soft-brown/40 rounded-[8px] bg-cream text-text-brown text-[13px] outline-none focus:border-primary/60 transition-colors font-sans appearance-none cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed";

const labelClass = "text-[12px] font-medium text-accent uppercase tracking-[0.5px]";

/* ─── 组件 ──────────────────────────────────── */

interface SettingsPanelProps {
  onClose: () => void;
}

export function SettingsPanel({ onClose }: SettingsPanelProps): React.JSX.Element {
  const storeSettings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);

  const [form, setForm] = useState({ ...storeSettings });

  // 根据当前 form.provider 查找 ProviderOption（必定存在，因为下拉选项与 PROVIDERS 一致）
  const currentProvider = (PROVIDERS.find((p) => p.id === form.provider) ?? PROVIDERS[0]) as ProviderOption;

  const handleChange = useCallback(
    (field: keyof typeof form, value: string | number) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  // 切换 Provider 时自动更新 baseUrl 和 modelName
  const handleProviderChange = useCallback(
    (providerId: string) => {
      const provider = PROVIDERS.find((p) => p.id === providerId);
      if (!provider) return;
      setForm((prev) => ({
        ...prev,
        provider: provider.id,
        baseUrl: provider.baseUrl,
        modelName: provider.models[0]?.id ?? prev.modelName,
      }));
    },
    [],
  );

  const handleSave = useCallback(() => {
    updateSettings(form);
    saveSettings();

    const llmConfig = {
      apiKey: form.apiKey,
      provider: form.provider,
      baseUrl: form.baseUrl,
      modelName: form.modelName,
      systemPrompt: form.persona,
    };
    void updateConfig(llmConfig);

    onClose();
  }, [form, updateSettings, saveSettings, onClose]);

  const handleClose = useCallback(() => {
    setForm({ ...storeSettings });
    onClose();
  }, [storeSettings, onClose]);

  return (
    <AnimatePresence>
      <motion.div
        className="fixed inset-0 z-200 flex items-center justify-center bg-black/20"
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        transition={{ duration: 0.2 }}
      >
        <motion.div
          className="w-[340px] max-h-[80vh] overflow-y-auto p-6 rounded-[24px] bg-cream border border-soft-brown/40 shadow-xl"
          initial={{ opacity: 0, scale: 0.95, y: 8 }}
          animate={{ opacity: 1, scale: 1, y: 0 }}
          exit={{ opacity: 0, scale: 0.95, y: 8 }}
          transition={{ duration: 0.2 }}
        >
          <h2 className="font-display text-[20px] font-bold text-text-brown mb-5">⚙️ 设置</h2>

          {/* API Key */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>API Key</span>
            <input
              className={inputClass}
              type="password"
              placeholder="sk-..."
              value={form.apiKey}
              onChange={(e): void => { handleChange("apiKey", e.target.value); }}
            />
          </label>

          {/* Provider */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>服务商</span>
            <select
              className={selectClass}
              value={form.provider}
              onChange={(e): void => { handleProviderChange(e.target.value); }}
            >
              {PROVIDERS.map((p) => (
                <option key={p.id} value={p.id}>{p.label}</option>
              ))}
            </select>
          </label>

          {/* Model */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>模型</span>
            <select
              className={selectClass}
              value={form.modelName}
              onChange={(e): void => { handleChange("modelName", e.target.value); }}
            >
              {currentProvider.models.map((m) => (
                <option key={m.id} value={m.id}>{m.label}</option>
              ))}
            </select>
          </label>

          {/* Base URL（只读展示，不可编辑） */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>接口地址</span>
            <input
              className={`${inputClass} opacity-60 cursor-not-allowed`}
              type="text"
              value={currentProvider.baseUrl}
              readOnly
              tabIndex={-1}
            />
          </label>

          {/* Persona */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>角色人设</span>
            <textarea
              className={`${inputClass} resize-y min-h-[60px]`}
              placeholder="描述宠物的性格…"
              rows={4}
              value={form.persona}
              onChange={(e): void => { handleChange("persona", e.target.value); }}
            />
          </label>

          {/* Opacity Slider */}
          <label className="flex flex-col gap-1 mb-5">
            <span className={labelClass}>
              窗口透明度: {Math.round(form.opacity * 100)}%
            </span>
            <input
              type="range"
              min={0.1}
              max={1}
              step={0.05}
              value={form.opacity}
              onChange={(e): void => { handleChange("opacity", Number.parseFloat(e.target.value)); }}
              className="w-full h-[6px] appearance-none bg-soft-brown/50 rounded-[3px] outline-none cursor-pointer accent-primary
                [&::-webkit-slider-thumb]:appearance-none [&::-webkit-slider-thumb]:w-[16px] [&::-webkit-slider-thumb]:h-[16px] [&::-webkit-slider-thumb]:rounded-full [&::-webkit-slider-thumb]:bg-primary [&::-webkit-slider-thumb]:border-2 [&::-webkit-slider-thumb]:border-cream [&::-webkit-slider-thumb]:cursor-pointer
                [&::-moz-range-thumb]:w-[16px] [&::-moz-range-thumb]:h-[16px] [&::-moz-range-thumb]:rounded-full [&::-moz-range-thumb]:bg-primary [&::-moz-range-thumb]:border-2 [&::-moz-range-thumb]:border-cream [&::-moz-range-thumb]:cursor-pointer"
            />
          </label>

          {/* Action buttons */}
          <div className="flex gap-3 mt-5">
            <motion.button
              className="flex-1 py-2.5 border-none rounded-[8px] bg-primary text-primary-content text-[14px] font-medium cursor-pointer"
              onClick={handleSave}
              type="button"
              whileHover={{ scale: 1.03 }}
              whileTap={{ scale: 0.97 }}
            >
              保存
            </motion.button>
            <motion.button
              className="flex-1 py-2.5 rounded-[8px] bg-soft-brown/30 text-text-brown text-[14px] font-medium cursor-pointer border border-soft-brown/30"
              onClick={handleClose}
              type="button"
              whileHover={{ scale: 1.03 }}
              whileTap={{ scale: 0.97 }}
            >
              取消
            </motion.button>
          </div>
        </motion.div>
      </motion.div>
    </AnimatePresence>
  );
}
