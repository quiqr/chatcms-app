import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { afterEach, describe, expect, it, vi } from "vitest";

import { inboxAddress } from "../DevMailLink";
import { ForgotPasswordPage } from "../ForgotPasswordPage";
import { SignInPage } from "../SignInPage";

afterEach(() => {
  vi.unstubAllEnvs();
});

function inbox() {
  return screen.findByRole("link", { name: "Open the inbox of this instance" });
}

describe("the link to the inbox", () => {
  it("joins the page's host and the inbox port", () => {
    expect(inboxAddress("localhost", "8025")).toBe("http://localhost:8025");
    expect(inboxAddress("192.168.1.20", "8026")).toBe("http://192.168.1.20:8026");
  });

  it("is on the sign-in screen in development, at this page's host", async () => {
    render(
      <MemoryRouter>
        <SignInPage />
      </MemoryRouter>,
    );

    const link = await inbox();
    expect(link).toHaveAttribute("href", `http://${window.location.hostname}:8025`);
    expect(link).toHaveAttribute("target", "_blank");
  });

  it("is on the forgot-password screen in development", async () => {
    render(
      <MemoryRouter>
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    expect(await inbox()).toHaveAttribute("href", `http://${window.location.hostname}:8025`);
  });

  it("is on neither screen outside development", async () => {
    vi.stubEnv("DEV", false);
    render(
      <MemoryRouter>
        <SignInPage />
        <ForgotPasswordPage />
      </MemoryRouter>,
    );

    await screen.findByRole("button", { name: "Sign in" });
    expect(screen.queryByRole("link", { name: "Open the inbox of this instance" })).toBeNull();
  });
});
