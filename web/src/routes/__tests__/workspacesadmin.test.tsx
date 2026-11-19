import { render, screen, within } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { PlatformWorkspaces, seconds } from "../PlatformWorkspaces";
import { PlatformAdminShell } from "../../shells/PlatformAdminShell";

const theAnswer = {
  backends: ["gvisor"],
  limit: 8,
  now: { starting: 1, warm: 3, sleeping: 5 },
  organisations: [{ slug: "technative", name: "TechNative", counts: { starting: 1, warm: 3, sleeping: 5 } }],
  projects: [
    { organisation: "technative", slug: "korenaar", name: "Korenaar Bakery", counts: { starting: 1, warm: 2, sleeping: 1 }, diskBytes: 5 * 1024 * 1024, diskKnown: true },
    { organisation: "technative", slug: "big", name: "Big site", counts: { starting: 0, warm: 1, sleeping: 4 }, diskBytes: 0, diskKnown: false },
  ],
  days: 7,
  starts: 42,
  medianMillis: 1430,
  slowMillis: 3210,
  refused: 2,
  failed: [{ reason: "did-not-start: the preview did not answer", times: 3 }],
  peaks: [{ day: "2026-11-19", warm: 6, starting: 1 }],
  historyDays: 90,
};

function serve(body: unknown) {
  vi.stubGlobal("fetch", vi.fn(() => Promise.resolve({ ok: true, status: 200, json: async () => body })));
}

afterEach(() => vi.unstubAllGlobals());

describe("the Workspaces screen", () => {
  it("shows now against the limit, the past days and the disk per project", async () => {
    serve(theAnswer);
    render(
      <Inside>
        <MemoryRouter>
          <PlatformWorkspaces />
        </MemoryRouter>
      </Inside>,
    );
    const now = await screen.findByRole("region", { name: "Now" });
    expect(now).toHaveTextContent("4 of 8");
    expect(now).toHaveTextContent("Running on gvisor");

    const past = screen.getByRole("region", { name: "The past 7 days" });
    expect(past).toHaveTextContent("42");
    expect(past).toHaveTextContent("1.4 s");
    expect(past).toHaveTextContent("Slow: 3.2 s");
    expect(within(past).getByRole("list", { name: "Starts that failed" })).toHaveTextContent("did-not-start: the preview did not answer3");
    expect(within(past).getByRole("list", { name: "Most warm at once, per day" })).toHaveTextContent("2026-11-196 warm");
    expect(past).toHaveTextContent("History is kept 90 days.");

    const projects = screen.getByRole("list", { name: "Per project" });
    expect(projects).toHaveTextContent("Korenaar Bakery");
    expect(projects).toHaveTextContent("1 starting · 2 warm · 1 sleeping");
    expect(projects).toHaveTextContent("5 MB on disk");
    expect(projects).toHaveTextContent("Disk not measured in time");
  });

  it("says when there is nothing yet", async () => {
    serve({ ...theAnswer, backends: [], now: { starting: 0, warm: 0, sleeping: 0 }, organisations: [], projects: [], failed: [], peaks: [] });
    render(
      <Inside>
        <MemoryRouter>
          <PlatformWorkspaces />
        </MemoryRouter>
      </Inside>,
    );
    expect(await screen.findByText("Nothing is running.")).toBeInTheDocument();
    expect(screen.getAllByText("No workspaces yet.")).toHaveLength(2);
    expect(screen.getByText("None.")).toBeInTheDocument();
  });

  it("is in the platform admin's navigation", () => {
    render(
      <Inside>
        <MemoryRouter>
          <PlatformAdminShell organisations={1} users={2}>body</PlatformAdminShell>
        </MemoryRouter>
      </Inside>,
    );
    expect(screen.getByRole("link", { name: "Workspaces" })).toHaveAttribute("href", "/administration/workspaces");
  });

  it("says seconds with one decimal", () => {
    expect(seconds(1430)).toBe("1.4 s");
    expect(seconds(0)).toBe("0 s");
  });
});
