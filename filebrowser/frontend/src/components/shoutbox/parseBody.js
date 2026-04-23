// frontend/src/components/shoutbox/parseBody.js
// Parses a body string into an array of tokens:
//   { type: "text", value: string }
//   { type: "link", path: string, line: number, rangeEnd?: number, snippet?: string, raw: string }
// A link may optionally be followed by a quoted snippet on the same line,
// which gets folded into the same token: [[path:42]] "selected text"
const LINK_RE = /\[\[([^\]]+?):(\d+)(?:-(\d+))?\]\](?:\s+"([^"\n]{1,200})")?/g;

export function parseBody(body) {
  const tokens = [];
  let last = 0;
  let m;
  LINK_RE.lastIndex = 0;
  while ((m = LINK_RE.exec(body)) !== null) {
    if (m.index > last) tokens.push({ type: "text", value: body.slice(last, m.index) });
    const path = m[1].trim();
    const line = parseInt(m[2], 10);
    const rangeEnd = m[3] ? parseInt(m[3], 10) : undefined;
    const snippet = m[4] || undefined;
    tokens.push({ type: "link", path, line, rangeEnd, snippet, raw: m[0] });
    last = m.index + m[0].length;
  }
  if (last < body.length) tokens.push({ type: "text", value: body.slice(last) });
  return tokens;
}
