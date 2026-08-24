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
      <IncidentDetail />
      <LiveView />
      <Diagnostics />
    </main>
  );
}

export function IncidentDetail() {
  return (
    <section aria-label="Incident detail">
      <h2>Failed request</h2>
      <p>GET /orders/:orderID — 500 — 88 ms</p>
      <label>
        Status{" "}
        <select defaultValue="500">
          <option>500</option>
        </select>
      </label>
      <label>
        Log severity{" "}
        <select defaultValue="ERROR">
          <option>ERROR</option>
        </select>
      </label>
      <h3>Correlated logs</h3>
      <p>ERROR payment provider failed</p>
      <h3>Host context</h3>
      <p>
        CPU utilization: 82% near this request. This is context, not a cause.
      </p>
    </section>
  );
}

export function LiveView({
  state = "connected",
}: {
  state?: "connected" | "interrupted" | "recovering";
}) {
  if (state === "interrupted")
    return (
      <section aria-label="Live View">
        <h2>Live View</h2>
        <p>Live updates were interrupted and may be incomplete.</p>
        <button>Refresh history</button>
      </section>
    );
  if (state === "recovering")
    return (
      <section aria-label="Live View">
        <h2>Live View</h2>
        <p>Reconciling with persisted history…</p>
      </section>
    );
  return (
    <section aria-label="Live View">
      <h2>Live View</h2>
      <p>Receiving new errors and logs.</p>
    </section>
  );
}

export function Diagnostics({
  logs = false,
  metrics = false,
}: {
  logs?: boolean;
  metrics?: boolean;
}) {
  return (
    <section aria-label="Onboarding diagnostics">
      <h2>Connection diagnostics</h2>
      <p>Operations: connected</p>
      <p>
        Logs:{" "}
        {logs ? "connected" : "not observed — add the DataSnoop slog handler."}
      </p>
      <p>
        Host measurements:{" "}
        {metrics
          ? "connected"
          : "not observed — enable application host collection."}
      </p>
      <p>DataSnoop platform: healthy</p>
    </section>
  );
}
