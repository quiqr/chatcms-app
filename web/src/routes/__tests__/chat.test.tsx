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
    FakeSource.last!.say({ type: "ready" });
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
    FakeSource.last!.say({ type: "ready" });

    FakeSource.last!.say({ type: "change", change: { page: "home", title: "Home page", before: "v1", after: "v2" } });
    expect(await screen.findByText("Changed Home page")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Show changes" })).toBeInTheDocument();
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
    FakeSource.last!.say({ type: "ready" });
    FakeSource.last!.say({ type: "error", message: "this organisation has used its AI budget" });
    expect(await screen.findByRole("alert")).toHaveTextContent("used its AI budget");
  });

  it("uses no git words", async () => {
    vi.stubGlobal("fetch", server(true));
    const { container } = panel();
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "ready" });
    FakeSource.last!.say({ type: "change", change: { page: "home", title: "Home page", before: "v1", after: "v2" } });
    expect(container.textContent).not.toMatch(/commit|branch|push|pull|merge|repository|conflict/i);
  });
});

describe("a quote from the preview", () => {
  const quote = { text: "Fresh bread every morning", before: "Korenaar ", after: ", baked daily.", path: "/posts/one/" };

  it("is shown, sent with the message, and given back with the change", async () => {
    const sent: { method: string; path: string; body: unknown }[] = [];
    vi.stubGlobal("fetch", server(true, sent));
    const onChanged = vi.fn();
    const dropped = vi.fn();
    render(
      <MemoryRouter>
        <ChatPanel organisation="technative" project="korenaar" onChanged={onChanged} quote={quote} onQuoteDropped={dropped} />
      </MemoryRouter>,
    );

    expect(await screen.findByLabelText("About this part of the page")).toHaveTextContent("Fresh bread every morning");
    const message = screen.getByRole("textbox", { name: "Your message" });
    expect(message).toHaveAttribute("placeholder", "What should change here?");
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "ready" });
    await userEvent.type(message, "Make it shorter{Enter}");

    await waitFor(() => expect(sent.some((one) => one.method === "POST")).toBe(true));
    expect(sent.find((one) => one.method === "POST")!.body).toEqual({ text: "Make it shorter", selection: quote });
    expect(dropped).toHaveBeenCalled();

    FakeSource.last!.say({ type: "change", change: { page: "posts/one", title: "One", before: "a", after: "b" } });
    expect(onChanged).toHaveBeenCalledWith(quote);
  });

  it("shows as a card in its context, which marks the place in the preview, and is counted while it runs", async () => {
    vi.stubGlobal("fetch", server(true));
    const shown = vi.fn();
    const running: number[] = [];
    render(
      <MemoryRouter>
        <ChatPanel
          organisation="technative"
          project="korenaar"
          onChanged={() => undefined}
          quote={quote}
          onQuoteDropped={() => undefined}
          onShowPlace={shown}
          onRunning={(count) => running.push(count)}
          pageTitle="One"
        />
      </MemoryRouter>,
    );
    const message = await screen.findByRole("textbox", { name: "Your message" });
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "ready" });
    await userEvent.type(message, "Make it shorter{Enter}");

    const card = await screen.findByRole("button", { name: "Show in the preview: Fresh bread every morning" });
    expect(card).toHaveTextContent("Korenaar Fresh bread every morning, baked daily.");
    expect(card.querySelector("mark")).toHaveTextContent("Fresh bread every morning");
    expect(screen.getByText("From the preview · One")).toBeInTheDocument();
    await waitFor(() => expect(running.at(-1)).toBe(1));

    await userEvent.click(card);
    expect(shown).toHaveBeenCalledWith(expect.objectContaining({ before: "Korenaar ", after: ", baked daily." }));

    FakeSource.last!.say({ type: "change", change: { page: "posts/one", title: "One", before: "a", after: "b" } });
    FakeSource.last!.say({ type: "done" });
    await waitFor(() => expect(running.at(-1)).toBe(0));
    await userEvent.click(screen.getByRole("button", { name: "Show changes" }));
    expect(shown).toHaveBeenCalledTimes(2);
  });

  it("sends a request from the overlay with its selection, and queues one that comes while busy", async () => {
    const sent: { method: string; path: string; body: unknown }[] = [];
    vi.stubGlobal("fetch", server(true, sent));
    const view = render(
      <MemoryRouter>
        <ChatPanel organisation="technative" project="korenaar" onChanged={() => undefined} request={{ about: quote, text: "Make it shorter", id: 1 }} pageTitle="One" />
      </MemoryRouter>,
    );
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "ready" });
    await waitFor(() => expect(sent.filter((one) => one.method === "POST")).toHaveLength(1));
    expect(sent.find((one) => one.method === "POST")!.body).toEqual({ text: "Make it shorter", selection: quote });
    expect(await screen.findByRole("button", { name: /^Show in the preview: Fresh bread/ })).toBeInTheDocument();

    view.rerender(
      <MemoryRouter>
        <ChatPanel organisation="technative" project="korenaar" onChanged={() => undefined} request={{ about: quote, text: "Fix spelling", id: 2 }} pageTitle="One" />
      </MemoryRouter>,
    );
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(sent.filter((one) => one.method === "POST")).toHaveLength(1);
    FakeSource.last!.say({ type: "done" });
    await waitFor(() => expect(sent.filter((one) => one.method === "POST")).toHaveLength(2));
    expect(sent.filter((one) => one.method === "POST")[1]!.body).toEqual({ text: "Fix spelling", selection: quote });
  });

  it("puts a draft from the overlay in the message box", async () => {
    vi.stubGlobal("fetch", server(true));
    render(
      <MemoryRouter>
        <ChatPanel organisation="technative" project="korenaar" onChanged={() => undefined} quote={quote} drafted={{ text: "Half a thought", id: 1 }} />
      </MemoryRouter>,
    );
    expect(await screen.findByRole("textbox", { name: "Your message" })).toHaveValue("Half a thought");
    expect(screen.getByLabelText("About this part of the page")).toHaveTextContent("Fresh bread every morning");
  });

  it("can be left out", async () => {
    vi.stubGlobal("fetch", server(true));
    const dropped = vi.fn();
    render(
      <MemoryRouter>
        <ChatPanel organisation="technative" project="korenaar" onChanged={() => undefined} quote={quote} onQuoteDropped={dropped} />
      </MemoryRouter>,
    );
    await userEvent.click(await screen.findByRole("button", { name: "Leave out the quote" }));
    expect(dropped).toHaveBeenCalled();
  });

  it("is not sent once it is gone", async () => {
    const sent: { method: string; path: string; body: unknown }[] = [];
    vi.stubGlobal("fetch", server(true, sent));
    panel();
    const message = await screen.findByRole("textbox", { name: "Your message" });
    await waitFor(() => expect(FakeSource.last).not.toBeNull());
    FakeSource.last!.say({ type: "ready" });
    await userEvent.type(message, "Change the title{Enter}");
    await waitFor(() => expect(sent.some((one) => one.method === "POST")).toBe(true));
    expect(sent.find((one) => one.method === "POST")!.body).toEqual({ text: "Change the title" });
  });
});

it("holds a message until the assistant is listening", async () => {
  const sent: { method: string; path: string; body: unknown }[] = [];
  vi.stubGlobal("fetch", server(true, sent));
  panel();
  const message = await screen.findByRole("textbox", { name: "Your message" });
  await waitFor(() => expect(FakeSource.last).not.toBeNull());
  await userEvent.type(message, "Early{Enter}");
  expect(sent.some((one) => one.method === "POST")).toBe(false);

  FakeSource.last!.say({ type: "ready" });
  await waitFor(() => expect(sent.some((one) => one.method === "POST")).toBe(true));
  expect(sent.find((one) => one.method === "POST")!.body).toEqual({ text: "Early" });
});
