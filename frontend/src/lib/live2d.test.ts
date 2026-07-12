/* ------------------------------------------------------------------ */
/*  Imports after mocks                                                */
/* ------------------------------------------------------------------ */
import { Application } from "pixi.js";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { initLive2D, setModelExpression, setModelMotion } from "./live2d";

// ---------------------------------------------------------------------------
// Hoisted mocks — accessible inside vi.mock factories even after hoisting
// ---------------------------------------------------------------------------

const { mockLive2DModel, mockLive2DModelFrom, mockAppInit, mockStageAddChild, MockApplication } =
  vi.hoisted(() => {
    const model = {
      anchor: { set: vi.fn() },
      position: { set: vi.fn() },
      scale: { set: vi.fn() },
      internalModel: { expressions: [], motions: {} },
      expression: vi.fn<[string], Promise<void>>().mockResolvedValue(undefined),
      motion: vi.fn<[string, number], Promise<void>>().mockResolvedValue(undefined),
    };

    const appInit = vi.fn<[Record<string, unknown>], Promise<void>>().mockResolvedValue(undefined);
    const stageAddChild = vi.fn();

    class App {
      init = appInit;
      stage = { addChild: stageAddChild };
    }

    return {
      mockLive2DModel: model,
      mockLive2DModelFrom: vi.fn<[string], Promise<typeof model>>().mockResolvedValue(model),
      mockAppInit: appInit,
      mockStageAddChild: stageAddChild,
      MockApplication: App,
    };
  });

vi.mock("pixi-live2d-display/cubism4", () => ({
  Live2DModel: {
    from: mockLive2DModelFrom,
  },
  CubismFramework: {
    cubismIdManager: {},
    cubismAllocate: {},
    cubismDeallocate: {},
    startUp: vi.fn(),
    initialize: vi.fn(),
  },
}));

vi.mock("pixi.js", () => ({
  Application: MockApplication,
}));

/* ------------------------------------------------------------------ */
/*  Helpers                                                            */
/* ------------------------------------------------------------------ */

function createMockCanvas(parentElement: HTMLElement | null): HTMLCanvasElement {
  const canvas = document.createElement("canvas");
  Object.defineProperty(canvas, "parentElement", {
    value: parentElement,
    writable: true,
    configurable: true,
  });
  return canvas;
}

/* ------------------------------------------------------------------ */
/*  Tests                                                              */
/* ------------------------------------------------------------------ */

beforeEach(() => {
  vi.clearAllMocks();
});

describe("initLive2D", () => {
  it("initializes PixiJS Application and returns it", async () => {
    const parent = document.createElement("div");
    Object.defineProperty(parent, "clientWidth", { value: 400 });
    Object.defineProperty(parent, "clientHeight", { value: 500 });
    const canvas = createMockCanvas(parent);

    const app = await initLive2D(canvas, {});

    expect(mockAppInit).toHaveBeenCalledOnce();
    expect(app).toBeInstanceOf(MockApplication);
  });

  it("uses parentElement dimensions when present", async () => {
    const parent = document.createElement("div");
    Object.defineProperty(parent, "clientWidth", { value: 800 });
    Object.defineProperty(parent, "clientHeight", { value: 600 });
    const canvas = createMockCanvas(parent);

    await initLive2D(canvas, {});

    const initCall = mockAppInit.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(initCall.width).toBe(800);
    expect(initCall.height).toBe(600);
    expect(initCall.resizeTo).toBe(parent);
  });

  it("defaults to 300x400 when canvas has no parent element", async () => {
    const canvas = createMockCanvas(null);

    await initLive2D(canvas, {});

    const initCall = mockAppInit.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(initCall.width).toBe(300);
    expect(initCall.height).toBe(400);
    expect(initCall.resizeTo).toBeUndefined();
  });

  it("loads a Live2D model when modelPath is provided", async () => {
    const parent = document.createElement("div");
    Object.defineProperty(parent, "clientWidth", { value: 400 });
    Object.defineProperty(parent, "clientHeight", { value: 400 });
    const canvas = createMockCanvas(parent);

    const modelPath = "/models/haru/haru.model3.json";
    await initLive2D(canvas, { modelPath });

    expect(mockLive2DModelFrom).toHaveBeenCalledWith(modelPath);
    expect(mockStageAddChild).toHaveBeenCalledWith(mockLive2DModel);
    expect(mockLive2DModel.anchor.set).toHaveBeenCalledWith(0.5, 0.5);
  });

  it("applies custom scale, x, y options to the model", async () => {
    const parent = document.createElement("div");
    Object.defineProperty(parent, "clientWidth", { value: 400 });
    Object.defineProperty(parent, "clientHeight", { value: 400 });
    const canvas = createMockCanvas(parent);

    await initLive2D(canvas, {
      modelPath: "/test.model3.json",
      scale: 0.8,
      x: 100,
      y: 200,
    });

    expect(mockLive2DModel.position.set).toHaveBeenCalledWith(100, 200);
    expect(mockLive2DModel.scale.set).toHaveBeenCalledWith(0.8);
  });

  it("does not attempt to load a model when modelPath is empty", async () => {
    const parent = document.createElement("div");
    Object.defineProperty(parent, "clientWidth", { value: 400 });
    Object.defineProperty(parent, "clientHeight", { value: 400 });
    const canvas = createMockCanvas(parent);

    await initLive2D(canvas, {});

    expect(mockLive2DModelFrom).not.toHaveBeenCalled();
    expect(mockStageAddChild).not.toHaveBeenCalled();
  });

  it("throws when init fails", async () => {
    mockAppInit.mockRejectedValueOnce(new Error("WebGL not available"));

    const canvas = createMockCanvas(null);

    await expect(initLive2D(canvas, {})).rejects.toThrow("WebGL not available");
  });

  it("re-throws non-Error failures as Error", async () => {
    mockAppInit.mockRejectedValueOnce("unknown string error");

    const canvas = createMockCanvas(null);

    await expect(initLive2D(canvas, {})).rejects.toThrow("unknown string error");
  });
});

/* ------------------------------------------------------------------ */
/*  setModelExpression                                                 */
/* ------------------------------------------------------------------ */

describe("setModelExpression", () => {
  it("calls model.expression when model exists", async () => {
    const app = {
      __live2dModel: mockLive2DModel,
    } as unknown as Application;

    await setModelExpression(app, "happy");

    expect(mockLive2DModel.expression).toHaveBeenCalledWith("happy");
  });

  it("does nothing when __live2dModel is missing", async () => {
    const app = {} as unknown as Application;

    await setModelExpression(app, "sad");

    expect(mockLive2DModel.expression).not.toHaveBeenCalled();
  });

  it("does nothing when internalModel is missing", async () => {
    const modelWithoutInternal = {
      expression: vi.fn<[string], Promise<void>>().mockResolvedValue(undefined),
    };
    const app = {
      __live2dModel: modelWithoutInternal,
    } as unknown as Application;

    await setModelExpression(app, "idle");

    expect(modelWithoutInternal.expression).not.toHaveBeenCalled();
  });

  it("swallows errors from model.expression", async () => {
    mockLive2DModel.expression.mockRejectedValueOnce(new Error("missing expression"));
    const app = {
      __live2dModel: mockLive2DModel,
    } as unknown as Application;

    await expect(setModelExpression(app, "nonexistent")).resolves.toBeUndefined();
  });
});

/* ------------------------------------------------------------------ */
/*  setModelMotion                                                     */
/* ------------------------------------------------------------------ */

describe("setModelMotion", () => {
  it("calls model.motion with group and default index", async () => {
    const app = {
      __live2dModel: mockLive2DModel,
    } as unknown as Application;

    await setModelMotion(app, "tap_body");

    expect(mockLive2DModel.motion).toHaveBeenCalledWith("tap_body", 0);
  });

  it("calls model.motion with custom index", async () => {
    const app = {
      __live2dModel: mockLive2DModel,
    } as unknown as Application;

    await setModelMotion(app, "idle", 2);

    expect(mockLive2DModel.motion).toHaveBeenCalledWith("idle", 2);
  });

  it("does nothing when __live2dModel is missing", async () => {
    const app = {} as unknown as Application;

    await setModelMotion(app, "tap_body");

    expect(mockLive2DModel.motion).not.toHaveBeenCalled();
  });

  it("does nothing when internalModel is missing", async () => {
    const modelWithoutInternal = {
      motion: vi.fn<[string, number], Promise<void>>().mockResolvedValue(undefined),
    };
    const app = {
      __live2dModel: modelWithoutInternal,
    } as unknown as Application;

    await setModelMotion(app, "tap_body");

    expect(modelWithoutInternal.motion).not.toHaveBeenCalled();
  });

  it("swallows errors from model.motion", async () => {
    mockLive2DModel.motion.mockRejectedValueOnce(new Error("missing motion"));
    const app = {
      __live2dModel: mockLive2DModel,
    } as unknown as Application;

    await expect(setModelMotion(app, "nonexistent")).resolves.toBeUndefined();
  });
});
