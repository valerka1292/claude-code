import { afterEach, describe, it } from "node:test";
import assert from "node:assert/strict";
import { GlobTool } from "../GlobTool";

const ORIGINAL_WINDOW = globalThis.window;

function createContext(cwd: string) {
  return { cwd };
}

afterEach(() => {
  Object.defineProperty(globalThis, "window", {
    configurable: true,
    writable: true,
    value: ORIGINAL_WINDOW,
  });
});

describe("GlobTool", () => {
  it("returns a controlled tool error when window.electronAPI is missing", async () => {
    Object.defineProperty(globalThis, "window", {
      configurable: true,
      writable: true,
      value: {},
    });

    const result = await GlobTool.execute(
      { pattern: "**/*.ts" },
      createContext("/tmp/project")
    );

    assert.equal(
      result,
      "<tool_use_error>Glob tool is unavailable: electronAPI not initialized</tool_use_error>"
    );
  });
});
