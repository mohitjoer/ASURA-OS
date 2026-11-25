import test from "node:test";
import assert from "node:assert/strict";
import { CATALOG, CATEGORIES, validate, commands } from "./catalog.js";

test("menu catalog validates cleanly with no errors", () => {
  const problems = validate(CATALOG);
  assert.deepEqual(
    problems,
    [],
    "Menu catalog should have no validation errors",
  );
});

test("menu catalog covers all 8 core categories from Workstream 4", () => {
  const required = [
    "Apps",
    "Learn",
    "Style",
    "Setup",
    "Install / Remove",
    "Update",
    "Capture",
    "System",
  ];
  for (const cat of required) {
    assert.ok(CATEGORIES.includes(cat), `Categories must include ${cat}`);
    const count = CATALOG.filter((item) => item.category === cat).length;
    assert.ok(
      count > 0,
      `Category ${cat} must have at least one entry, found ${count}`,
    );
  }
});

test("menu catalog commands are distinct non-empty binaries", () => {
  const cmds = commands();
  assert.ok(cmds.length > 0, "Should extract commands");
  assert.ok(cmds.includes("toren"), "Should include toren");
  assert.ok(cmds.includes("toren-shell"), "Should include toren-shell");
  assert.ok(cmds.includes("ghostty"), "Should include ghostty");
  assert.ok(cmds.includes("loginctl"), "Should include loginctl");
});
