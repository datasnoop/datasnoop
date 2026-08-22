import { expect, test } from "vitest";

import { App } from "./App";

test("defines the Lounge application component", () => {
  expect(App).toBeTypeOf("function");
});
