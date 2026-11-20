import { render, screen, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { setWidth } from "./width";
import { LogPage } from "../LogPage";

const base = "/api/organisations/technative/projects/korenaar/log";

const entries = [
  { entry: "start:7", kind: "start", at: new Date().toISOString(), state: "failed", what: "gvisor", who: "Joris", millis: 4100 },
  { entry: "publish:p1", kind: "publish", at: new Date().toISOString(), state: "succeeded", what: "3", who: "Marieke", millis: 22000 },
  { entry: "build:b1", kind: "build", at: new Date().toISOString(), state: "succeeded", what: "publish", millis: 18000 },
];

const start = {
  ...entries[0],
  hugo: { asked: "0.101.0", program: "/nix/store/h-hugo-0.101.0/bin/hugo", because: "pinned", printed: "hugo v0.101.0+extended" },
  steps: [{ step: "building-preview", state: "failed" }],
  log: "backend gvisor\nError: template failed",
  reason: "did-not-start: no answer",
};

function serve() {
  const asked: string[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn((url: string) => {
      asked.push(url);
      const body = url.startsWith(`${base}/start`) ? start : url.includes("kind=start") ? { entries: [entries[0]], more: false } : { entries, more: true };
      return Promise.resolve({ ok: true, status: 200, json: async () => body });
    }),
  );
  return asked;
}

function at(address: string) {
  return render(
    <Inside>
      <MemoryRouter initialEntries={[address]}>
        <Routes>
          <Route path="/o/:organisation/p/:project/log" element={<LogPage />} />
        </Routes>
      </MemoryRouter>
    </Inside>,
  );
}

beforeEach(() => setWidth(1440));
afterEach(() => vi.unstubAllGlobals());

describe("the project log", () => {
  it("lists starts, builds and publications together, and filters by kind", async () => {
    const asked = serve();
    at("/o/technative/p/korenaar/log");

    const list = await screen.findByRole("list", { name: "Log" });
    expect(within(list).getAllByRole("button")).toHaveLength(3);
    expect(list).toHaveTextContent("Preview start");
    expect(list).toHaveTextContent("Joris · gvisor");
    expect(list).toHaveTextContent("Marieke · 3 changes");
    expect(list).toHaveTextContent("Failed · 4.1 s");
    expect(screen.getByRole("button", { name: "Show older" })).toBeInTheDocument();

    await userEvent.click(screen.getByRole("radio", { name: "Preview starts" }));
    expect(await screen.findAllByRole("button", { expanded: false })).toHaveLength(1);
    expect(asked.some((url) => url.includes("kind=start"))).toBe(true);
  });

  it("opens an entry from its address, with what Hugo was asked, chosen and said", async () => {
    serve();
    at("/o/technative/p/korenaar/log?entry=start%3A7");

    const detail = await screen.findByTestId("log-detail");
    const hugo = within(detail).getByRole("region", { name: "Hugo" });
    expect(hugo).toHaveTextContent("0.101.0");
    expect(hugo).toHaveTextContent("/nix/store/h-hugo-0.101.0/bin/hugo");
    expect(hugo).toHaveTextContent("pinned through Nixhub");
    expect(hugo).toHaveTextContent("hugo v0.101.0+extended");
    expect(within(detail).getByRole("region", { name: "What went wrong" })).toHaveTextContent("did-not-start: no answer");
    expect(within(detail).getByText(/Error: template failed/)).toBeInTheDocument();
    expect(within(detail).getByRole("button", { name: "Copy" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: /Preview start/, expanded: true })).toBeInTheDocument();
  });

  it("fills a phone's screen with an opened entry, and goes back to the list", async () => {
    setWidth(390);
    serve();
    at("/o/technative/p/korenaar/log?entry=start%3A7");

    expect(await screen.findByTestId("log-detail")).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Log" })).toBeNull();
    await userEvent.click(screen.getByRole("button", { name: "Back to the log" }));
    expect(await screen.findByRole("list", { name: "Log" })).toBeInTheDocument();
  });
});
