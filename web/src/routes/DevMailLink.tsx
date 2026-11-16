import { say } from "../lib/say";

export function inboxAddress(hostname: string, port: string) {
  return `http://${hostname}:${port}`;
}

export function DevMailLink() {
  return (
    <p className="mt-3 text-[13px] text-subtle-foreground">
      {say("dev-mail.mail-sent-by-this-instance")}{" "}
      <a
        className="inline-flex min-h-11 items-center text-primary"
        href={inboxAddress(window.location.hostname, import.meta.env.VITE_CHATCMS_INBOX_PORT)}
        target="_blank"
        rel="noreferrer"
      >
        {say("dev-mail.open-the-inbox")}
      </a>
    </p>
  );
}

export default function DevMailBlock() {
  return (
    <section
      aria-labelledby="dev-mail"
      className="rounded-md border border-dashed border-input p-4"
    >
      <h2 id="dev-mail" className="text-[13px] font-semibold text-muted-foreground">
        {say("dev-sign-in.development-only")}
      </h2>
      <DevMailLink />
    </section>
  );
}
