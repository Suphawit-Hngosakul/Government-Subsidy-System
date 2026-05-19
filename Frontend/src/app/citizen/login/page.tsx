"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

type LoginResponse = {
  token: string;
  role: string;
  error?: string;
};

export default function CitizenLoginPage() {
  const [nationalId, setNationalId] = useState("1101700203451");
  const [password, setPassword] = useState("password");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function login(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setMessage("Signing in with backend...");

    try {
      const response = await fetch("/api/backend/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nationalId, password }),
      });
      const payload = (await response.json()) as LoginResponse;
      if (!response.ok) throw new Error(payload.error ?? "Login failed");

      localStorage.setItem("citizenToken", payload.token);
      localStorage.setItem("citizenNationalId", nationalId);
      localStorage.setItem("citizenRole", payload.role);
      window.location.href = "/citizen";
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="app-shell auth-shell citizen-auth">
      <main className="auth-layout">
        <section className="auth-hero glass">
          <Link className="brand" href="/">
            <div className="seal">RT</div>
            <div className="brand-text">
              <p className="brand-title">Citizen Portal</p>
              <p className="brand-subtitle">Auto-decision starts here</p>
            </div>
          </Link>
          <div>
            <p className="eyebrow">Citizen Workspace</p>
            <h1>Sign in to continue your benefit claim.</h1>
            <p className="lead">
              Access the citizen dashboard, submit a claim, and follow the auto-decision status in real time.
            </p>
          </div>
          <div className="auth-proof">
            <span>Auth Service</span>
            <span>Benefit Service</span>
            <span>Live SSE</span>
          </div>
        </section>

        <form className="auth-card neo-card" onSubmit={login}>
          <div className="auth-card-header">
            <span className="badge green">Citizen Login</span>
            <h2>Citizen Sign In</h2>
            <p>Use a registered National ID to enter the auto-decision claim flow.</p>
          </div>
          <label className="field">
            <span>National ID</span>
            <input className="input neo-input" inputMode="numeric" value={nationalId} onChange={(event) => setNationalId(event.target.value)} />
          </label>
          <label className="field">
            <span>Password</span>
            <input className="input neo-input" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </label>
          <button className="btn btn-primary full-width" disabled={loading} type="submit">
            {loading ? "Signing in..." : "Continue to Citizen Portal"}
          </button>
          {message ? <p className="auth-note">{message}</p> : null}
          <p className="auth-note">
            Need an account? <Link href="/citizen/register">Create citizen account</Link>
          </p>
        </form>
      </main>
    </div>
  );
}
