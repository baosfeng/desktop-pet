import { useCallback, useState } from "react";
import type React from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Slider } from "@/components/ui/slider";

import { verifyApiKey as bridgeVerifyApiKey, setWindowOpacity, updateConfig } from "@/lib/bridge";

import { usePetStore } from "@/stores/petStore";

/* ─── API Key 验证 ──────────────────────────── */

type VerifyStatus = "idle" | "verifying" | "success" | "error";

async function verifyApiKey(
  baseUrl: string,
  apiKey: string,
  modelName: string,
): Promise<{ ok: true } | { ok: false; message: string }> {
  try {
    await bridgeVerifyApiKey({
      apiKey,
      provider: "",
      baseUrl,
      modelName,
      systemPrompt: "",
    });
    return { ok: true };
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err);
    return { ok: false, message: msg };
  }
}

/* ─── Provider 配置 ──────────────────────────── */

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
      { id: "deepseek-v4-flash", label: "V4 Flash（默认）" },
      { id: "deepseek-v4-pro", label: "V4 Pro（思考）" },
    ],
  },
  {
    id: "openai",
    label: "OpenAI",
    baseUrl: "https://api.openai.com/v1",
    models: [
      { id: "gpt-4o-mini", label: "GPT-4o Mini" },
      { id: "gpt-4o", label: "GPT-4o" },
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
    ],
  },
];

/* ─── 组件 ──────────────────────────────────── */

interface SettingsPanelProps {
  onClose: () => void;
}

export function SettingsPanel({ onClose }: SettingsPanelProps): React.JSX.Element {
  const storeSettings = usePetStore((s) => s.settings);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);

  const [form, setForm] = useState({ ...storeSettings });
  const [verifyStatus, setVerifyStatus] = useState<VerifyStatus>("idle");

  // eslint-disable-next-line @typescript-eslint/no-non-null-assertion
  const currentProvider = PROVIDERS.find((p) => p.id === form.provider) ?? PROVIDERS[0]!;

  const handleChange = useCallback((field: keyof typeof form, value: string | number) => {
    setForm((prev) => ({ ...prev, [field]: value }));
  }, []);

  const handleProviderChange = useCallback((value: string) => {
    const provider = PROVIDERS.find((p) => p.id === value);
    if (!provider) return;
    setForm((prev) => ({
      ...prev,
      provider: provider.id,
      baseUrl: provider.baseUrl,
      modelName: provider.models[0]?.id ?? prev.modelName,
    }));
  }, []);

  const handleVerify = useCallback(async () => {
    if (!form.apiKey) {
      setVerifyStatus("error");
      return;
    }
    setVerifyStatus("verifying");
    const result = await verifyApiKey(form.baseUrl, form.apiKey, form.modelName);
    setVerifyStatus(result.ok ? "success" : "error");
  }, [form]);

  const handleSave = useCallback(() => {
    updateSettings(form);
    saveSettings();
    void usePetStore.getState().saveApiKey();
    void updateConfig({
      apiKey: form.apiKey,
      provider: form.provider,
      baseUrl: form.baseUrl,
      modelName: form.modelName,
      systemPrompt: form.persona,
    });
    void setWindowOpacity(form.opacity);
    onClose();
  }, [form, updateSettings, saveSettings, onClose]);

  const handleClose = useCallback(() => {
    setForm({ ...storeSettings });
    onClose();
  }, [storeSettings, onClose]);

  const canSave = !form.apiKey || verifyStatus === "success";

  return (
    <Dialog
      open
      onOpenChange={(open) => {
        if (!open) handleClose();
      }}
    >
      <DialogContent className="sm:max-w-[420px] p-0 gap-0 max-h-[85vh] overflow-hidden">
        <DialogHeader className="px-6 pt-5 pb-2">
          <DialogTitle className="text-xl font-display font-bold">⚙️ 设置</DialogTitle>
        </DialogHeader>

        <div className="overflow-y-auto px-6 pb-6 space-y-4">
          {/* API Key */}
          <div className="space-y-1.5">
            <div className="flex items-center justify-between">
              <label className="text-xs font-semibold text-accent uppercase tracking-wider">
                API Key
              </label>
              <button
                className={`text-[11px] px-2.5 py-1 rounded-md border-none cursor-pointer font-medium transition-all
                  ${verifyStatus === "idle" ? "bg-accent/15 text-accent hover:bg-accent/25" : ""}
                  ${verifyStatus === "verifying" ? "bg-muted text-muted-foreground cursor-wait" : ""}
                  ${verifyStatus === "success" ? "bg-green-100 text-green-700" : ""}
                  ${verifyStatus === "error" ? "bg-red-100 text-red-600" : ""}`}
                // eslint-disable-next-line @typescript-eslint/no-misused-promises
                onClick={verifyStatus === "verifying" ? undefined : handleVerify}
                type="button"
                disabled={verifyStatus === "verifying"}
              >
                {verifyStatus === "idle" && "🔍 验证"}
                {verifyStatus === "verifying" && "⏳ 验证中..."}
                {verifyStatus === "success" && "✅ 已验证"}
                {verifyStatus === "error" && "🔄 重试"}
              </button>
            </div>
            <input
              className="w-full h-auto px-3 py-2 rounded-lg border border-input bg-background text-sm text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-ring focus:ring-2 focus:ring-ring/20 transition-all"
              type="password"
              placeholder="sk-..."
              value={form.apiKey}
              onChange={(e): void => {
                handleChange("apiKey", e.target.value);
                setVerifyStatus("idle");
              }}
            />
            {verifyStatus === "error" && !form.apiKey && (
              <p className="text-xs text-destructive mt-0.5">请先输入 API Key</p>
            )}
          </div>

          {/* Provider */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-accent uppercase tracking-wider">
              服务商
            </label>
            <Select
              value={form.provider}
              onValueChange={(v) => {
                if (v) handleProviderChange(v);
              }}
            >
              <SelectTrigger className="w-full h-auto px-3 py-2 rounded-lg text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {PROVIDERS.map((p) => (
                  <SelectItem key={p.id} value={p.id}>
                    {p.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Model */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-accent uppercase tracking-wider">
              模型
            </label>
            <Select
              value={form.modelName}
              onValueChange={(v) => {
                if (v) handleChange("modelName", v);
              }}
            >
              <SelectTrigger className="w-full h-auto px-3 py-2 rounded-lg text-sm">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {currentProvider.models.map((m) => (
                  <SelectItem key={m.id} value={m.id}>
                    {m.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {/* Base URL */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-accent uppercase tracking-wider">
              接口地址
            </label>
            <input
              className="w-full h-auto px-3 py-2 rounded-lg border border-input bg-muted/30 text-sm text-muted-foreground cursor-not-allowed outline-none"
              type="text"
              value={currentProvider.baseUrl}
              readOnly
              tabIndex={-1}
            />
          </div>

          {/* Persona */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-accent uppercase tracking-wider">
              角色人设
            </label>
            <textarea
              className="w-full h-auto min-h-[60px] px-3 py-2 rounded-lg border border-input bg-background text-sm text-foreground placeholder:text-muted-foreground/50 outline-none focus:border-ring focus:ring-2 focus:ring-ring/20 transition-all resize-y"
              placeholder="描述宠物的性格…"
              rows={3}
              value={form.persona}
              onChange={(e): void => {
                handleChange("persona", e.target.value);
              }}
            />
          </div>

          {/* Opacity */}
          <div className="space-y-1.5">
            <label className="text-xs font-semibold text-accent uppercase tracking-wider">
              窗口透明度: {Math.round(form.opacity * 100)}%
            </label>
            <Slider
              value={[form.opacity]}
              onValueChange={(v: number[]) => {
                const val = v[0];
                if (val !== undefined) handleChange("opacity", val);
              }}
              min={0.1}
              max={1}
              step={0.05}
              className="w-full"
            />
          </div>
        </div>

        {/* Action buttons */}
        <div className="px-6 pb-5 pt-2 border-t border-border/50">
          <div className="flex gap-3">
            <button
              className={`flex-1 px-4 py-2.5 rounded-lg text-sm font-medium border-none transition-all
                ${
                  canSave
                    ? "bg-primary text-primary-foreground cursor-pointer hover:opacity-90 active:scale-95"
                    : "bg-muted text-muted-foreground/50 cursor-not-allowed"
                }`}
              onClick={canSave ? handleSave : undefined}
              type="button"
            >
              保存
            </button>
            <button
              className="flex-1 px-4 py-2.5 rounded-lg text-sm font-medium bg-muted text-foreground border border-border cursor-pointer hover:bg-muted/80 active:scale-95 transition-all"
              onClick={handleClose}
              type="button"
            >
              取消
            </button>
          </div>
          {form.apiKey && verifyStatus !== "success" && (
            <p className="text-xs text-destructive/70 mt-1.5 ml-0.5">
              {verifyStatus === "idle" ? "🔍 请先验证 API Key 后再保存" : ""}
              {verifyStatus === "verifying" ? "⏳ 验证中，请稍候..." : ""}
              {verifyStatus === "error" ? "❌ 验证失败，请检查 API Key 后重试" : ""}
            </p>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
