import type { ReactNode } from "react";

export type Endpoint = {
  route: string;
  requests: number;
  errors: number;
  durationMs: number;
};
export type OverviewState = "loading" | "empty" | "populated" | "degraded";

const endpoints: Endpoint[] = [
  { route: "/orders/:orderID", requests: 24, errors: 7, durationMs: 88 },
  { route: "/health", requests: 200, errors: 0, durationMs: 2 },
];

export function Overview({
  state = "populated",
  items = endpoints,
}: {
  state?: OverviewState;
  items?: Endpoint[];
}) {
  const content: Record<OverviewState, ReactNode> = {
    loading: <p>Loading endpoint activity…</p>,
    empty: <p>No request activity in this time window.</p>,
    degraded: <p>Some telemetry is delayed. Refresh history to reconcile.</p>,
    populated: (
      <ol>
        {[...items]
          .sort(
            (left, right) =>
              right.errors - left.errors || right.durationMs - left.durationMs,
          )
          .map((endpoint) => (
            <li key={endpoint.route}>
              <strong>{endpoint.route}</strong> — {endpoint.requests} requests,{" "}
              {endpoint.errors} errors, {endpoint.durationMs} ms
            </li>
          ))}
      </ol>
    ),
  };
  return <section aria-label="Endpoint overview">{content[state]}</section>;
}

export function App() {
  return (
    <main>
      <h1>DataSnoop Lounge</h1>
      <label>
        Service{" "}
        <select defaultValue="checkout">
          <option>checkout</option>
        </select>
      </label>
      <label>
        Time window{" "}
        <select defaultValue="Last hour">
          <option>Last hour</option>
          <option>Last 24 hours</option>
        </select>
      </label>
      <Overview />
    </main>
  );
}
