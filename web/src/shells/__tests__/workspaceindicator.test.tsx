import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, describe, expect, it, vi } from "vitest";

import { WorkspaceIndicator, shownOf } from "../WorkspaceIndicator";

type Answer = { status?: number; body: unknown };

function serve(routes: Record<string, Answer>) {
  const fetched: string[] = [];
  vi.stubGlobal(
    "fetch",
    vi.fn((url: string, init?: RequestInit) => {
      const asked = `${init?.method ?? "GET"} ${url}`;
      fetched.push(asked);
      const answer = routes[asked] ?? { status: 404, body: {} };
      const status = answer.status ?? 200;
      return Promise.resolve({ ok: status < 300, status, json: async () => answer.body });
    }),
  );
  return fetched;
}

const base = "/api/organisations/technative/projects/korenaar";

afterEach(() => vi.unstubAllGlobals());

describe("the workspace symbol", () => {
  it("tells the states apart", () => {
    expect(shownOf(null)).toBe("sleeping");
    expect(shownOf({ state: "warm" })).toBe("ready");
    expect(shownOf({ state: "starting" })).toBe("starting");
    expect(shownOf({ state: "sleeping" })).toBe("sleeping");
    expect(shownOf({ state: "sleeping", reason: "preview.did-not-start" })).toBe("did-not-start");
  });

  it("says plain words to an editor", async () => {
    serve({
      [`GET ${base}/workspace`]: { body: { state: "warm", since: new Date().toISOString() } },
      [`GET ${base}/chat`]: { body: { available: true } },
    });
    render(<WorkspaceIndicator organisation="technative" project="korenaar" />);

    const symbol = await screen.findByRole("button", { name: "Your preview: Preview ready" });
    await userEvent.click(symbol);

    expect(await screen.findByText("The assistant can be asked.")).toBeInTheDocument();
    expect(screen.getByText("Started just now")).toBeInTheDocument();
    expect(screen.queryByText(/Runs on/)).not.toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Start again" })).not.toBeInTheDocument();
  });

  it("says a sleeping preview wakes when opened, and an absent assistant", async () => {
    serve({
      [`GET ${base}/workspace`]: { body: { state: "sleeping" } },
      [`GET ${base}/chat`]: { body: { available: false } },
    });
    render(<WorkspaceIndicator organisation="technative" project="korenaar" />);

    await userEvent.click(await screen.findByRole("button", { name: "Your preview: Preview asleep" }));
    expect(await screen.findByText("The assistant is not available.")).toBeInTheDocument();
    expect(screen.getByText("It starts again when you open the preview.")).toBeInTheDocument();
  });

  it("gives an implementer the backend, the steps and a way to start again", async () => {
    const fetched = serve({
      [`GET ${base}/workspace`]: {
        body: {
          state: "warm",
          backend: "gvisor",
          since: new Date().toISOString(),
          steps: [{ step: "building-preview", state: "done" }],
          log: ["hugo server started"],
        },
      },
      [`GET ${base}/chat`]: { body: { available: true } },
      [`POST ${base}/workspace/restart`]: { body: { state: "starting", backend: "gvisor" } },
    });
    const restarted = vi.fn();
    window.addEventListener("chatcms:workspace-restarted", restarted);
    render(<WorkspaceIndicator organisation="technative" project="korenaar" />);

    await userEvent.click(await screen.findByRole("button", { name: "Your preview: Preview ready" }));
    expect(await screen.findByText("Runs on a gVisor sandbox")).toBeInTheDocument();
    expect(screen.getByText("Building the preview")).toBeInTheDocument();
    expect(screen.getByText("hugo server started")).toBeInTheDocument();

    await userEvent.click(screen.getByRole("button", { name: "Start again" }));
    await waitFor(() => expect(fetched).toContain(`POST ${base}/workspace/restart`));
    await waitFor(() => expect(restarted).toHaveBeenCalled());
    window.removeEventListener("chatcms:workspace-restarted", restarted);
  });

  it("says a start that failed, with its reason", async () => {
    serve({
      [`GET ${base}/workspace`]: { body: { state: "sleeping", reason: "preview.did-not-start" } },
      [`GET ${base}/chat`]: { body: { available: false } },
    });
    render(<WorkspaceIndicator organisation="technative" project="korenaar" />);
    await userEvent.click(await screen.findByRole("button", { name: "Your preview: Preview did not start" }));
    expect(await screen.findByText("Preview did not start", { selector: "p" })).toBeInTheDocument();
  });
});
