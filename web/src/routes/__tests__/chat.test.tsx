import { act, render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { ChatPanel } from "../ChatPanel";

const base = "/api/organisations/technative/projects/korenaar";

class FakeSource {
  static last: FakeSource | null = null;
  onmessage: ((message: { data: string }) => void) | null = null;
  closed = false;
  constructor(public url: string) {
    FakeSource.last = this;
  }
  close() {
    this.closed = true;
  }
  say(event: unknown) {
    act(() => this.onmessage?.({ data: JSON.stringify(event) }));
  }
}

function server(available: boolean, sent: { method: string; path: string; body: unknown }[] = []) {
  return vi.fn().mockImplementation((path: string, options?: RequestInit) => {
    const method = options?.method ?? "GET";
    sent.push({ method, path, body: options?.body ? JSON.parse(String(options.body)) : null });
    if (method === "GET" && path === `${base}/chat`) {
      return Promise.resolve({ ok: true, status: 200, json: async () => ({ available }) });
    }
    return Promise.resolve({ ok: true, status: method === "POST" ? 202 : 200, json: async () => ({}) });
  });
}

function panel(onChanged = () => undefined) {
  return render(
    <MemoryRouter>
      <ChatPanel organisation="technative" project="korenaar" onChanged={onChanged} />
    </MemoryRouter>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
  vi.stubGlobal("EventSource", FakeSource);
  FakeSource.last = null;
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the assistant panel", () => {
  it("says when there is no assistant, and opens no stream", async () => {
    vi.stubGlobal("fetch", server(false));
    panel();
    expect(await screen.findByText("The assistant is not available on this installation.")).toBeInTheDocument();
    expect(FakeSource.last).toBeNull();
  });

  it("sends a message and shows the reply as it streams", async () => {
    const sent: { method: string; path: string; body: unknown }[] = [];
    vi.stubGlobal("fetch", server(true, sent));
    panel();

    const message = await screen.findByRole("textbox", { name: "Your message" });
    await waitFor(() => expect(FakeSource.last?.url).toBe(`${base}/chat/events`));
    await userEvent.type(message, "Change the title{Enter}");

    await waitFor(() => expect(sent.some((call) => call.method === "POST" && call.path === `${base}/chat`)).toBe(true));
    expect(sent.find((call) => call.method === "POST")!.body).toEqual({ text: "Change the title" });
    expect(screen.getByText("Change the title")).toBeInTheDocument();
    expect(screen.getByText("Working on it…")).toBeInTheDocument();

    FakeSource.last!.say({ type: "text", text: "Hel" });
    FakeSource.last!.say({ type: "text", text: "lo" });
    expect(screen.getByText("Hello")).toBeInTheDocument();
    FakeSource.last!.say({ type: "done" });
    expect(screen.getByRole("button", { name: "Send" })).toBeDisabled();
  });

  it("lists a change with Show and Undo, and Undo goes back without publishing", async () => {
    const sent: { method: string; path: string; body: unknown }[] = [];
    const changed = vi.fn();
    vi.stubGlobal("fetch", server(true, sent));
    panel(changed);
    await waitFor(() => expect(FakeSource.last).not.toBeNull());

    FakeSource.last!.say({ type: "change", change: { page: "home", title: "Home page", before: "v1", after: "v2" } });
    expect(await screen.findByText("Changed Home page")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Show" })).toBeInTheDocument();
    expect(changed).toHaveBeenCalledTimes(1);

    await userEvent.click(screen.getByRole("button", { name: "Undo" }));
    await waitFor(() => expect(screen.getByText("Home page is back as it was")).toBeInTheDocument());
    const undone = sent.find((call) => call.path === `${base}/versions/v1/go-back`);
    expect(undone?.body).toEqual({ publishRightAway: false });
    expect(changed).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole("button", { name: "Undo" })).toBeNull();
  });

  it("says when the assistant could not answer", async () => {
    vi.stubGlobal("fetch", server(true));
    panel();
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "error", message: "this organisation has used its AI budget" });
    expect(await screen.findByRole("alert")).toHaveTextContent("used its AI budget");
  });

  it("uses no git words", async () => {
    vi.stubGlobal("fetch", server(true));
    const { container } = panel();
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "change", change: { page: "home", title: "Home page", before: "v1", after: "v2" } });
    expect(container.textContent).not.toMatch(/commit|branch|push|pull|merge|repository|conflict/i);
  });
});
