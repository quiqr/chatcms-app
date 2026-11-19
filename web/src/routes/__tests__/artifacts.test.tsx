import { render, screen } from "@testing-library/react";
import { MemoryRouter, Route, Routes } from "react-router";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { Inside } from "./here";
import { ArtifactsPage, sizeOf } from "../ArtifactsPage";

const base = "/api/organisations/technative/projects/korenaar";

function server(answer: unknown, status = 200) {
  return vi.fn().mockImplementation(() => Promise.resolve({ ok: status < 400, status, json: async () => answer }));
}

function renderPage() {
  return render(
    <Inside>
      <MemoryRouter initialEntries={["/o/technative/p/korenaar/artifacts"]}>
        <Routes>
          <Route path="/o/:organisation/p/:project/artifacts" element={<ArtifactsPage />} />
        </Routes>
      </MemoryRouter>
    </Inside>,
  );
}

beforeEach(() => {
  document.cookie = "chatcms_request_token=a-token; path=/";
});

afterEach(() => {
  vi.unstubAllGlobals();
});

describe("the artifacts screen", () => {
  it("lists each kept file with a download", async () => {
    vi.stubGlobal(
      "fetch",
      server([
        { id: "a2", name: "Korenaar-Recipes.pdf", size: 2_400_000, createdAt: "2026-11-19T09:00:00Z" },
        { id: "a1", name: "Korenaar-Recipes.pdf", size: 900, createdAt: "2026-11-18T09:00:00Z" },
      ]),
    );
    renderPage();
    const links = await screen.findAllByRole("link", { name: "Download Korenaar-Recipes.pdf" });
    expect(links).toHaveLength(2);
    expect(links[0]).toHaveAttribute("href", `${base}/artifacts/a2`);
    expect(screen.getByText(/2\.3 MB/)).toBeInTheDocument();
    expect(screen.getByText(/900 bytes/)).toBeInTheDocument();
  });

  it("says when nothing has been kept", async () => {
    vi.stubGlobal("fetch", server([]));
    renderPage();
    expect(await screen.findByText(/Nothing has been kept yet/)).toBeInTheDocument();
  });

  it("says when the list cannot be read", async () => {
    vi.stubGlobal("fetch", server({ error: "down" }, 500));
    renderPage();
    expect(await screen.findByRole("alert")).toHaveTextContent("The files could not be listed");
  });
});

describe("a file's size", () => {
  it("is said in the unit that reads best", () => {
    expect(sizeOf(12)).toBe("12 bytes");
    expect(sizeOf(20_480)).toBe("20 KB");
    expect(sizeOf(5 * 1024 * 1024)).toBe("5 MB");
  });
});
