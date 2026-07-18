// verify.dreamtree.org — the verification service (DT-23, v0).
//
// Zero-auth public reads; graded verdicts (observed | converged | not_found),
// never a bare boolean. The index is LAZY: a miss queues the hash and the m3
// resolver — which only ever dials OUT — answers within a poll cycle.
// The same Worker speaks MCP (POST /mcp, authless JSON-RPC) with one tool.

// Minimal D1 surface (self-contained — no @cloudflare/workers-types dep).
interface D1Result<T> { results?: T[] }
interface D1Stmt {
  bind(...args: unknown[]): D1Stmt;
  first<T>(): Promise<T | null>;
  run(): Promise<unknown>;
  all<T>(): Promise<D1Result<T>>;
}
interface D1 { prepare(sql: string): D1Stmt }

export interface Env {
  DB: D1;
  VERIFY_SYNC_TOKEN: string;
}

const HEX64 = /^[0-9a-f]{64}$/;
const NOT_FOUND_TTL_S = 3600; // a not_found may become observed later — re-ask hourly
const RETRY_AFTER_S = 4;

const json = (obj: unknown, status = 200, extra: Record<string, string> = {}) =>
  new Response(JSON.stringify(obj), {
    status,
    headers: { "content-type": "application/json", "access-control-allow-origin": "*", ...extra },
  });

function authed(req: Request, env: Env): boolean {
  const t = (req.headers.get("authorization") || "").replace(/^Bearer\s+/i, "");
  if (!env.VERIFY_SYNC_TOKEN || !t || t.length !== env.VERIFY_SYNC_TOKEN.length) return false;
  // constant-time-ish compare
  let diff = 0;
  for (let i = 0; i < t.length; i++) diff |= t.charCodeAt(i) ^ env.VERIFY_SYNC_TOKEN.charCodeAt(i);
  return diff === 0;
}

async function lookup(env: Env, hash: string): Promise<{ code: number; body: unknown }> {
  if (!HEX64.test(hash)) return { code: 400, body: { error: "hash must be 64 lowercase hex chars (sha256)" } };
  const row = await env.DB.prepare("SELECT status, verdict, resolved_at FROM verdicts WHERE hash = ?")
    .bind(hash).first<{ status: string; verdict: string | null; resolved_at: string | null }>();
  if (!row) {
    await env.DB.prepare(
      "INSERT INTO verdicts (hash, status) VALUES (?, 'pending') ON CONFLICT(hash) DO NOTHING",
    ).bind(hash).run();
    return { code: 202, body: { status: "pending", retry_after_s: RETRY_AFTER_S } };
  }
  if (row.status === "pending") return { code: 202, body: { status: "pending", retry_after_s: RETRY_AFTER_S } };
  if (row.status === "not_found" && row.resolved_at) {
    const age = (Date.now() - Date.parse(row.resolved_at + "Z")) / 1000;
    if (age > NOT_FOUND_TTL_S) {
      await env.DB.prepare("UPDATE verdicts SET status='pending', verdict=NULL WHERE hash=?").bind(hash).run();
      return { code: 202, body: { status: "pending", retry_after_s: RETRY_AFTER_S } };
    }
  }
  return { code: 200, body: row.verdict ? JSON.parse(row.verdict) : { status: row.status } };
}

// ---- MCP (authless, single tool) -------------------------------------------

const C2PA_TOOL = {
  name: "verify_c2pa_url",
  description:
    "Validate C2PA Content Credentials for an asset at a URL (image/video/document). Returns a " +
    "graded verdict — trusted (chains to the C2PA trust list), valid_untrusted (signature valid, " +
    "signer unknown), invalid, or no_credentials — with signer details, plus whether the asset " +
    "bytes are a recorded observation on the dreamtree chain. 'pending' means resolving; retry " +
    "in a few seconds with the returned verify_key via verify_observation.",
  inputSchema: {
    type: "object",
    properties: { url: { type: "string", description: "http(s) URL of the asset" } },
    required: ["url"],
  },
};

const TOOL = {
  name: "verify_observation",
  description:
    "Verify whether a sha256 content hash is a recorded observation on the dreamtree chain. " +
    "Returns a graded verdict: observed (recorded), converged (independently re-observed — the " +
    "strongest signal), or not_found — with who observed it, when, a self-contained Merkle " +
    "inclusion proof, the on-chain anchor (batch/tx/height), and the stamper's standing. " +
    "A 'pending' result means the index is resolving; retry in a few seconds.",
  inputSchema: {
    type: "object",
    properties: { hash: { type: "string", description: "64-char lowercase hex sha256" } },
    required: ["hash"],
  },
};

async function mcp(req: Request, env: Env): Promise<Response> {
  let rpc: { jsonrpc?: string; id?: unknown; method?: string; params?: any };
  try { rpc = await req.json(); } catch { return json({ jsonrpc: "2.0", id: null, error: { code: -32700, message: "parse error" } }); }
  const reply = (result: unknown) => json({ jsonrpc: "2.0", id: rpc.id ?? null, result });
  switch (rpc.method) {
    case "initialize":
      return reply({
        protocolVersion: rpc.params?.protocolVersion ?? "2025-06-18",
        capabilities: { tools: {} },
        serverInfo: { name: "dreamtree-verify", version: "0.1.0" },
      });
    case "notifications/initialized":
      return new Response(null, { status: 202 });
    case "tools/list":
      return reply({ tools: [TOOL, C2PA_TOOL] });
    case "tools/call": {
      const name = rpc.params?.name;
      if (name === TOOL.name) {
        const { body } = await lookup(env, String(rpc.params?.arguments?.hash ?? "").toLowerCase());
        return reply({
          content: [{ type: "text", text: JSON.stringify(body) }],
          structuredContent: body, isError: false,
        });
      }
      if (name === C2PA_TOOL.name) {
        const target = String(rpc.params?.arguments?.url ?? "");
        if (!/^https?:\/\//.test(target) || target.length > 2048)
          return json({ jsonrpc: "2.0", id: rpc.id ?? null, error: { code: -32602, message: "url must be http(s)" } });
        const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(target));
        const key = [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
        await env.DB.prepare(
          "INSERT INTO verdicts (hash, status, kind, url) VALUES (?, 'pending', 'c2pa_url', ?) " +
            "ON CONFLICT(hash) DO NOTHING",
        ).bind(key, target).run();
        const { body } = await lookup(env, key);
        const out = { verify_key: key, ...(body as object) };
        return reply({
          content: [{ type: "text", text: JSON.stringify(out) }],
          structuredContent: out, isError: false,
        });
      }
      return json({ jsonrpc: "2.0", id: rpc.id ?? null, error: { code: -32602, message: "unknown tool" } });
    }
    default:
      return json({ jsonrpc: "2.0", id: rpc.id ?? null, error: { code: -32601, message: "method not found" } });
  }
}

// ---- router ----------------------------------------------------------------

export default {
  async fetch(req: Request, env: Env): Promise<Response> {
    const url = new URL(req.url);
    const path = url.pathname;

    if (req.method === "GET" && path.startsWith("/verify/")) {
      const { code, body } = await lookup(env, path.slice("/verify/".length).toLowerCase());
      return json(body, code, code === 202 ? { "retry-after": String(RETRY_AFTER_S) } : {});
    }

    // DT-22: C2PA-by-URL. The job key is sha256(url); the resolver fetches the
    // asset, validates Content Credentials, and checks the asset bytes against
    // the observation log in the same verdict.
    if (req.method === "POST" && path === "/verify/url") {
      const body = (await req.json().catch(() => null)) as { url?: string } | null;
      const target = body?.url ?? "";
      if (!/^https?:\/\//.test(target) || target.length > 2048)
        return json({ error: "url must be http(s), <= 2048 chars" }, 400);
      const digest = await crypto.subtle.digest("SHA-256", new TextEncoder().encode(target));
      const key = [...new Uint8Array(digest)].map((b) => b.toString(16).padStart(2, "0")).join("");
      await env.DB.prepare(
        "INSERT INTO verdicts (hash, status, kind, url) VALUES (?, 'pending', 'c2pa_url', ?) " +
          "ON CONFLICT(hash) DO NOTHING",
      ).bind(key, target).run();
      const { code, body: out } = await lookup(env, key);
      return json({ verify_key: key, ...(out as object) }, code === 202 ? 202 : code);
    }

    if (req.method === "POST" && path === "/mcp") return mcp(req, env);

    if (req.method === "GET" && path === "/healthz") {
      const c = await env.DB.prepare(
        "SELECT status, count(*) AS n FROM verdicts GROUP BY status",
      ).all<{ status: string; n: number }>();
      const counts: Record<string, number> = {};
      for (const r of c.results ?? []) counts[(r as { status: string; n: number }).status] = (r as { status: string; n: number }).n;
      return json({ ok: true, counts });
    }

    // m3 resolver seam — both calls INITIATED by m3 (outbound); Bearer-gated.
    if (req.method === "POST" && path === "/queue/pull") {
      if (!authed(req, env)) return json({ error: "unauthorized" }, 401);
      const max = Math.min(Number((await req.json().catch(() => ({})) as any).max ?? 25), 100);
      const rows = await env.DB.prepare(
        "SELECT hash, kind, url FROM verdicts WHERE status='pending' ORDER BY requested_at LIMIT ?",
      ).bind(max).all<{ hash: string; kind: string; url: string | null }>();
      return json({
        hashes: (rows.results ?? []).map((r) => r.hash), // v0 resolver compat
        jobs: rows.results ?? [],
      });
    }
    if (req.method === "POST" && path === "/queue/resolve") {
      if (!authed(req, env)) return json({ error: "unauthorized" }, 401);
      const body = (await req.json().catch(() => null)) as
        | { results?: { hash: string; status: string; verdict?: unknown }[] }
        | null;
      if (!body?.results) return json({ error: "results required" }, 400);
      let n = 0;
      for (const r of body.results) {
        const STATUSES = ["observed", "converged", "not_found",
          "trusted", "valid_untrusted", "invalid", "no_credentials", "fetch_failed"];
        if (!HEX64.test(r.hash) || !STATUSES.includes(r.status)) continue;
        await env.DB.prepare(
          "UPDATE verdicts SET status=?, verdict=?, resolved_at=datetime('now') WHERE hash=?",
        ).bind(r.status, r.verdict ? JSON.stringify(r.verdict) : null, r.hash).run();
        n++;
      }
      return json({ ok: true, resolved: n });
    }

    if (req.method === "GET" && path === "/") {
      return new Response(
        "dreamtree verify — GET /verify/{sha256} for a graded verdict " +
          "(observed | converged | not_found) with Merkle proof, on-chain anchor, and stamper " +
          "standing. MCP at POST /mcp (tool: verify_observation). Spec: " +
          "github.com/blong-dev/dreamtree docs/specs/verify-service.md\n",
        { headers: { "content-type": "text/plain" } },
      );
    }
    return json({ error: "not found" }, 404);
  },
};
