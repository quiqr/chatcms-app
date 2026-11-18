import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { AIUseSection } from "../AIUseSection";

const path = "/api/administration/organisations/bakery/ai";

function server(answer: unknown, sent: { method: string; body: unknown }[] = []) {
  return vi.fn().mockImplementation((asked: string, options?: RequestInit) => {
    const method = options?.method ?? "GET";
    sent.push({ method: `${method} ${asked}`, body: options?.body ? JSON.parse(String(options.body)) : null });
    return Promise.resolve({ ok: true, status: 200, json: async () => answer });
  });
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("AI use", () => {
  it("says this month's use and that there are no limits", async () => {
    vi.stubGlobal(
      "fetch",
      server({ requestsPerMinute: null, tokensPerMonth: null, requestsThisMonth: 12, tokensThisMonth: 34567 }),
    );
    render(<AIUseSection organisation="bakery" />);

    expect(await screen.findByText(/12 requests and 34,567 tokens this month/)).toBeInTheDocument();
    expect(screen.getByRole("checkbox", { name: "Limit requests a minute" })).not.toBeChecked();
    expect(screen.getByRole("checkbox", { name: "Limit tokens a month" })).not.toBeChecked();
  });

  it("sets a monthly budget and leaves the rate unlimited", async () => {
    const sent: { method: string; body: unknown }[] = [];
    vi.stubGlobal(
      "fetch",
      server({ requestsPerMinute: null, tokensPerMonth: null, requestsThisMonth: 0, tokensThisMonth: 0 }, sent),
    );
    render(<AIUseSection organisation="bakery" />);

    await userEvent.click(await screen.findByRole("checkbox", { name: "Limit tokens a month" }));
    const budget = screen.getByRole("spinbutton", { name: "Limit tokens a month" });
    await userEvent.clear(budget);
    await userEvent.type(budget, "2000000");
    await userEvent.click(screen.getByRole("button", { name: "Save AI limits" }));

    await waitFor(() => expect(sent.some((call) => call.method === `PUT ${path}`)).toBe(true));
    const put = sent.find((call) => call.method === `PUT ${path}`)!;
    expect(put.body).toEqual({ requestsPerMinute: null, tokensPerMonth: 2000000 });
    expect(await screen.findByRole("status")).toHaveTextContent("Saved");
  });

  it("shows nothing when the use cannot be read", async () => {
    vi.stubGlobal("fetch", vi.fn().mockResolvedValue({ ok: false, status: 404, json: async () => ({}) }));
    const { container } = render(<AIUseSection organisation="bakery" />);
    await new Promise((resolve) => setTimeout(resolve, 20));
    expect(container).toBeEmptyDOMElement();
  });
});
