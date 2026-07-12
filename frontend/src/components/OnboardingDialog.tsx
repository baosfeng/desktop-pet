import { useCallback, useState } from "react";
import type React from "react";

import { Dialog, DialogContent, DialogHeader, DialogTitle } from "@/components/ui/dialog";
import { usePetStore } from "@/stores/petStore";

const STEPS = [
  {
    title: "🐾 欢迎来到桌面 AI 宠物",
    description: "我是你的桌面 AI 小伙伴！可以聊天、互动，越用越懂你。",
    icon: "🎉",
  },
  {
    title: "🔑 配置 API Key",
    description: "点击「设置」按钮，输入你的 API Key 并验证，我就能和你聊天啦！",
    icon: "⚙️",
  },
  {
    title: "🎨 自定义模型",
    description: "你可以在设置中选择不同的 AI 模型和 Live2D 宠物外观。",
    icon: "✨",
  },
  {
    title: "💬 开始对话",
    description: "在下方输入框打字和我聊天，点击宠物可以互动，现在就试试吧！",
    icon: "🚀",
  },
] as const;

export function OnboardingDialog(): React.JSX.Element {
  const [step, setStep] = useState(0);
  const updateSettings = usePetStore((s) => s.updateSettings);
  const saveSettings = usePetStore((s) => s.saveSettings);
  const toggleSettings = usePetStore((s) => s.toggleSettings);
  const showSettings = usePetStore((s) => s.showSettings);

  const currentStep: (typeof STEPS)[number] = STEPS[step] ?? STEPS[0];
  const isLastStep = step >= STEPS.length - 1;

  const handleNext = useCallback(() => {
    if (isLastStep) {
      updateSettings({ hasCompletedOnboarding: true });
      saveSettings();
    } else {
      setStep((p) => p + 1);
    }
  }, [isLastStep, updateSettings, saveSettings]);

  const handleOpenSettings = useCallback(() => {
    if (!showSettings) {
      toggleSettings();
    }
    updateSettings({ hasCompletedOnboarding: true });
    saveSettings();
  }, [showSettings, toggleSettings, updateSettings, saveSettings]);

  return (
    <Dialog open onOpenChange={undefined}>
      <DialogContent className="sm:max-w-[380px] p-0 gap-0 overflow-hidden">
        <DialogHeader className="px-6 pt-6 pb-2">
          <DialogTitle className="text-lg font-display font-bold text-center">
            {currentStep.title}
          </DialogTitle>
        </DialogHeader>

        <div className="px-6 pb-6 flex flex-col items-center gap-4">
          <div className="text-6xl py-4">{currentStep.icon}</div>
          <p className="text-sm text-muted-foreground text-center leading-relaxed">
            {currentStep.description}
          </p>

          {/* Step indicator */}
          <div className="flex gap-2 mt-2">
            {STEPS.map((_, i) => (
              <div
                key={i}
                className={`w-2 h-2 rounded-full transition-all ${
                  i === step ? "bg-primary w-4" : "bg-muted-foreground/20"
                }`}
              />
            ))}
          </div>

          <div className="flex gap-3 w-full mt-2">
            {isLastStep ? (
              <>
                <button
                  className="flex-1 px-4 py-2.5 rounded-lg text-sm font-medium bg-primary text-primary-foreground cursor-pointer border-none hover:opacity-90 active:scale-95 transition-all"
                  onClick={handleNext}
                  type="button"
                >
                  开始使用 🚀
                </button>
                <button
                  className="px-4 py-2.5 rounded-lg text-sm font-medium bg-muted text-foreground border border-border cursor-pointer hover:bg-muted/80 transition-all"
                  onClick={handleOpenSettings}
                  type="button"
                >
                  去设置
                </button>
              </>
            ) : (
              <button
                className="flex-1 px-4 py-2.5 rounded-lg text-sm font-medium bg-primary text-primary-foreground cursor-pointer border-none hover:opacity-90 active:scale-95 transition-all"
                onClick={handleNext}
                type="button"
              >
                下一步 →
              </button>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
