import { NextResponse } from "next/server";

const BACKEND_URL = process.env.GOV_SUBSIDY_BACKEND_URL ?? "http://localhost:8080";

export async function POST(request: Request) {
  const body = await request.json();

  const response = await fetch(`${BACKEND_URL}/internal/orchestrate`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
    cache: "no-store",
  });

  const payload = await response.json();
  return NextResponse.json(payload, { status: response.status });
}
