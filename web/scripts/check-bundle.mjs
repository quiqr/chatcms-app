import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative } from "node:path";

const root = new URL("..", import.meta.url).pathname;
const dist = process.argv[2] ?? join(root, "dist");

const forbidden = [
  "a long enough passphrase",
  "marieke@korenaar.example",
  "sam@korenaar.example",
  "owner@example.org",
  "admin@example.org",
];

function* files(dir) {
  for (const entry of readdirSync(dir)) {
    const path = join(dir, entry);
    if (statSync(path).isDirectory()) {
      yield* files(path);
    } else if (/\.(js|html|css|map)$/.test(entry)) {
      yield path;
    }
  }
}

const problems = [];
for (const path of files(dist)) {
  const text = readFileSync(path, "utf8");
  for (const word of forbidden) {
    if (text.includes(word)) problems.push(`${relative(root, path)} carries "${word}"`);
  }
}

const icon = join(dist, "favicon.png");
const page = readFileSync(join(dist, "index.html"), "utf8");
if (!/<link rel="icon" type="image\/png" href="\/favicon\.png"/.test(page)) {
  problems.push("index.html does not name favicon.png as the icon");
}
if (!existsSync(icon) || readFileSync(icon).subarray(1, 4).toString() !== "PNG") {
  problems.push("favicon.png is not in the build, or is not a PNG");
}

if (problems.length > 0) {
  console.error("check-bundle: the production build is not what ships");
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}
console.log("check-bundle: no development sign-in, and the Quiqr icon, in the production build");
