import { describe, expect, it } from "vitest";
import { motion, AnimatePresence } from "motion/react";
import { render } from "@testing-library/react";

describe("motion import", () => {
  it("can render motion.div", () => {
    const { container } = render(
      <AnimatePresence>
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
          hello
        </motion.div>
      </AnimatePresence>
    );
    expect(container.textContent).toBe("hello");
  });
});
