import { NextResponse } from "next/server";

const BACKEND_URL = process.env.GOV_SUBSIDY_BACKEND_URL ?? "http://localhost:8080";

type RouteContext = {
  params: Promise<{
    path: string[];
  }>;
};

async function proxy(request: Request, context: RouteContext) {
  const { path } = await context.params;
  const sourceURL = new URL(request.url);
  const targetURL = new URL(path.join("/"), `${BACKEND_URL}/`);
  targetURL.search = sourceURL.search;

  try {
    const response = await fetch(targetURL, {
      method: request.method,
      headers: forwardHeaders(request),
      body: request.method === "GET" || request.method === "HEAD" ? undefined : await request.arrayBuffer(),
      cache: "no-store",
    });

    const headers = new Headers(response.headers);
    headers.delete("content-encoding");
    headers.delete("content-length");

    return new Response(response.body, {
      status: response.status,
      headers,
    });
  } catch {
    return NextResponse.json(
      { error: "backend unavailable", backendUrl: BACKEND_URL },
      { status: 502 },
    );
  }
}

function forwardHeaders(request: Request) {
  const headers = new Headers();
  const contentType = request.headers.get("content-type");
  const authorization = request.headers.get("authorization");

  if (contentType) headers.set("content-type", contentType);
  if (authorization) headers.set("authorization", authorization);

  return headers;
}

export const GET = proxy;
export const POST = proxy;
export const PUT = proxy;
export const PATCH = proxy;
export const DELETE = proxy;
