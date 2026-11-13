import userEvent from "@testing-library/user-event";

export async function fill(field: HTMLElement, text: string) {
  await userEvent.click(field);
  await userEvent.paste(text);
}
