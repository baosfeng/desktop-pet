import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { AnimatePresence, motion } from "motion/react";

describe("render motion", () => {
  it("renders motion with AnimatePresence", () => {
    render(
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
            <h2>⚙️ 设置</h2>
          </motion.div>
        </motion.div>
      </AnimatePresence>
    );
    expect(screen.getByText("⚙️ 设置")).toBeDefined();
  });
});
