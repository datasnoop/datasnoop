import { expect, test } from "vitest";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";

import { App, Diagnostics, IncidentDetail, LiveView, Overview } from "./App";

test("defines the Lounge application component", () => {
  expect(App).toBeTypeOf("function");
  expect(Overview).toBeTypeOf("function");
  expect(IncidentDetail).toBeTypeOf("function");
  expect(LiveView).toBeTypeOf("function");
  expect(Diagnostics).toBeTypeOf("function");
});

test("offers actionable product guidance for missing logs and host measurements", () => {
  const diagnostics = renderToStaticMarkup(createElement(Diagnostics));
  expect(diagnostics).toContain("not observed");
  expect(diagnostics).toContain("slog handler");
  expect(diagnostics).toContain("application host collection");
});

test("shows interruption and history recovery states in Live View", () => {
  expect(
    renderToStaticMarkup(createElement(LiveView, { state: "connected" })),
  ).toContain("Receiving");
  expect(
    renderToStaticMarkup(createElement(LiveView, { state: "interrupted" })),
  ).toContain("Refresh history");
  expect(
    renderToStaticMarkup(createElement(LiveView, { state: "recovering" })),
  ).toContain("persisted history");
});

test("renders the endpoint-to-occurrence investigation path without query syntax", () => {
  const detail = renderToStaticMarkup(createElement(IncidentDetail));
  expect(detail).toContain("Failed request");
  expect(detail).toContain("Correlated logs");
  expect(detail).toContain("Host context");
  expect(detail).not.toContain("SELECT");
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
