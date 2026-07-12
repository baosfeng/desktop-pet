import { AnimatePresence, motion } from "motion/react";
import { useCallback, useState } from "react";
import type React from "react";

import { usePetStore } from "@/stores/petStore";
import { updateConfig } from "@/lib/bridge";

interface SettingsPanelProps {
  onClose: () => void;
}

export function SettingsPanel({ onClose }: SettingsPanelProps): React.JSX.Element {
  const storeSettings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);

  const [form, setForm] = useState({ ...storeSettings });

  const handleChange = useCallback(
    (field: keyof typeof form, value: string | number) => {
      setForm((prev) => ({ ...prev, [field]: value }));
    },
    [],
  );

  const handleSave = useCallback(() => {
    // 更新 store 并保存到 localStorage
    updateSettings(form);
    saveSettings();

    // 将 LLM 设置同步到 PetCore 后端
    const llmConfig = {
      apiKey: form.apiKey,
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

  const inputClass =
    "w-full px-3 py-2 border border-soft-brown/40 rounded-[8px] bg-cream text-text-brown text-[13px] outline-none focus:border-primary/60 transition-colors placeholder:text-text-brown/30 font-sans";

  const labelClass = "text-[12px] font-medium text-accent uppercase tracking-[0.5px]";

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

          {/* Base URL */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>Base URL</span>
            <input
              className={inputClass}
              type="text"
              placeholder="https://api.openai.com/v1"
              value={form.baseUrl}
              onChange={(e): void => { handleChange("baseUrl", e.target.value); }}
            />
          </label>

          {/* Model Name */}
          <label className="flex flex-col gap-1 mb-4">
            <span className={labelClass}>模型名</span>
            <input
              className={inputClass}
              type="text"
              placeholder="gpt-4o-mini"
              value={form.modelName}
              onChange={(e): void => { handleChange("modelName", e.target.value); }}
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
