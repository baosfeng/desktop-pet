/**
 * Live2D 渲染引擎
 *
 * 使用 PixiJS 8 + pixi-live2d-display 加载和渲染 Live2D 模型。
 * 需要自行准备 Live2D 模型文件（.model3.json 或 .moc3 文件）。
 *
 * 使用示例：
 * ```ts
 * import { initLive2D } from "./live2d";
 *
 * const app = await initLive2D(document.getElementById("live2d-canvas")!, {
 *   modelPath: "/models/haru/haru.model3.json",
 * });
 * ```
 */

import { Application } from "pixi.js";

/* eslint-disable @typescript-eslint/no-explicit-any, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-call, no-console -- pixi-live2d-display 缺少 TS 类型 */

// pixi-live2d-display 的 TS 类型支持可能不完整，使用动态导入
let Live2DModel: any;
let CubismFramework: any;

export interface Live2DOptions {
  /** Live2D 模型文件路径（.model3.json 或 .moc3） */
  modelPath?: string;
  /** 是否启用调试模式 */
  debug?: boolean;
  /** 模型缩放比例（默认 0.5） */
  scale?: number;
  /** 模型 X 轴偏移 */
  x?: number;
  /** 模型 Y 轴偏移 */
  y?: number;
}

/**
 * 初始化 Live2D 渲染
 * @param canvas canvas DOM 元素
 * @param options 配置选项
 * @returns PixiJS Application 实例（可用来销毁/暂停）
 */
export async function initLive2D(
  canvas: HTMLCanvasElement,
  options: Live2DOptions = {},
): Promise<Application | null> {
  try {
    // 仅加载 Cubism 4 专有包（不需要 Cubism 2.1 的 live2d.min.js）
    const live2dModule = await import("pixi-live2d-display/cubism4");
    Live2DModel = live2dModule.Live2DModel;
    CubismFramework = (live2dModule as any).CubismFramework;

    if (CubismFramework) {
      CubismFramework.startUp(
        CubismFramework.cubismIdManager,
        CubismFramework.cubismAllocate,
        CubismFramework.cubismDeallocate,
      );
      CubismFramework.initialize();
    }

    const {
      modelPath = "",
      scale = 0.5,
      x = 0,
      y = 0,
    } = options;

    // 创建 PixiJS Application
    const app = new Application();

    // Build init options, omitting resizeTo when parent is null (exactOptionalPropertyTypes compat)
    const initOpts: Record<string, unknown> = {
      canvas,
      width: canvas.parentElement?.clientWidth ?? 300,
      height: canvas.parentElement?.clientHeight ?? 400,
      backgroundAlpha: 0,
      antialias: true,
      autoDensity: true,
    };
    if (canvas.parentElement) {
      initOpts.resizeTo = canvas.parentElement;
    }
    await app.init(initOpts);

    // 加载 Live2D 模型
    if (modelPath) {
      const model = await Live2DModel.from(modelPath);
      model.anchor.set(0.5, 0.5);
      model.position.set(x, y);
      model.scale.set(scale);
      app.stage.addChild(model);

      // 存储模型引用
      (app as any).__live2dModel = model;
    }

    return app;
  } catch (err) {
    const errorMsg = err instanceof Error ? err.message : String(err);
    console.warn("[Live2D] Failed to initialize:", errorMsg);
    throw new Error(errorMsg);
  }
}

/**
 * 设置模型动画（表情/动作）
 */
export async function setModelExpression(
  app: Application,
  expressionId: string,
): Promise<void> {
  const model = (app as any).__live2dModel;
  if (!model?.internalModel) return;

  try {
    await model.expression(expressionId);
  } catch {
    // 忽略表情缺失错误
  }
}

/**
 * 设置模型动作
 */
export async function setModelMotion(
  app: Application,
  motionGroup: string,
  motionIndex = 0,
): Promise<void> {
  const model = (app as any).__live2dModel;
  if (!model?.internalModel) return;

  try {
    await model.motion(motionGroup, motionIndex);
  } catch {
    // 忽略动作缺失错误
  }
}
