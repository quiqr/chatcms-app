import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { PlatformOrganisationsScreen } from "../PlatformOrganisationsScreen";

afterEach(() => {
  vi.unstubAllGlobals();
  document.cookie = "chatcms_request_token=; max-age=0; path=/";
});

describe("starting an organisation", () => {
  it("reads again who is signed in, so the organisation menu holds the new one", async () => {
    document.cookie = "chatcms_request_token=t; path=/";
    vi.stubGlobal(
      "fetch",
      vi.fn((_url: string, init?: RequestInit) => {
        const body = init?.method === "POST" ? { slug: "bakkerij-zuid", name: "Bakkerij Zuid" } : { organisations: [] };
        return Promise.resolve({ ok: true, status: 200, json: async () => body });
      }),
    );
    const again = vi.fn();
    render(
      <Inside administers again={again}>
        <MemoryRouter>
          <PlatformOrganisationsScreen />
        </MemoryRouter>
      </Inside>,
    );

    await userEvent.click(await screen.findByRole("button", { name: /New organisation/ }));
    await userEvent.type(screen.getByLabelText("Name", { exact: true }), "Bakkerij Zuid");
    await userEvent.type(screen.getByLabelText("Who administers it"), "joris@example.org");
    await userEvent.click(screen.getByRole("button", { name: "Start it" }));

    await waitFor(() => expect(again).toHaveBeenCalled());
  });
});
