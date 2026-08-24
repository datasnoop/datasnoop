import { expect, test } from "vitest";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import { App, Overview } from "./App";

test("defines the Lounge application component", () => {
  expect(App).toBeTypeOf("function");
  expect(Overview).toBeTypeOf("function");
});

test("renders loading, empty, populated, and degraded overview states", () => {
  expect(
    renderToStaticMarkup(createElement(Overview, { state: "loading" })),
  ).toContain("Loading");
  expect(
    renderToStaticMarkup(createElement(Overview, { state: "empty" })),
  ).toContain("No request activity");
  expect(
    renderToStaticMarkup(createElement(Overview, { state: "degraded" })),
  ).toContain("delayed");
  const populated = renderToStaticMarkup(
    createElement(Overview, {
      items: [
        { route: "/healthy", requests: 4, errors: 0, durationMs: 2 },
        { route: "/failing", requests: 2, errors: 1, durationMs: 3 },
      ],
    }),
  );
  expect(populated.indexOf("/failing")).toBeLessThan(
    populated.indexOf("/healthy"),
  );
});
