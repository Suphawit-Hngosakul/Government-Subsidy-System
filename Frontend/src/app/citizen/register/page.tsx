"use client";

import Link from "next/link";
import { FormEvent, useState } from "react";

type OCRResponse = {
  nationalId: string;
  fullName: string;
  dateOfBirth: string;
  laserCode: string;
  address?: string;
  error?: string;
};

type LoginResponse = {
  token: string;
  role: string;
  error?: string;
};

export default function CitizenRegisterPage() {
  const [nationalId, setNationalId] = useState("1101700203451");
  const [phone, setPhone] = useState("0812345678");
  const [password, setPassword] = useState("password");
  const [laserCode, setLaserCode] = useState("AA1-1234567-89");
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);

  async function register(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    setMessage("Registering citizen with backend...");

    try {
      const registerResponse = await fetch("/api/backend/api/v1/auth/register", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nationalId, password, phone }),
      });
      const registerPayload = await registerResponse.json();
      if (!registerResponse.ok) throw new Error(registerPayload.error ?? "Registration failed");

      const ocrResponse = await fetch("/api/backend/api/v1/auth/ekyc/ocr", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nationalId }),
      });
      const ocrPayload = (await ocrResponse.json()) as OCRResponse;
      if (!ocrResponse.ok) throw new Error(ocrPayload.error ?? "OCR failed");

      const confirmResponse = await fetch("/api/backend/api/v1/auth/ekyc/confirm", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          nationalId,
          laserCode: ocrPayload.laserCode || laserCode,
          fullName: ocrPayload.fullName,
          dateOfBirth: ocrPayload.dateOfBirth,
        }),
      });
      const confirmPayload = await confirmResponse.json();
      if (!confirmResponse.ok) throw new Error(confirmPayload.error ?? "KYC confirmation failed");

      const loginResponse = await fetch("/api/backend/api/v1/auth/login", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ nationalId, password }),
      });
      const loginPayload = (await loginResponse.json()) as LoginResponse;
      if (!loginResponse.ok) throw new Error(loginPayload.error ?? "Login failed");

      localStorage.setItem("citizenToken", loginPayload.token);
      localStorage.setItem("citizenNationalId", nationalId);
      localStorage.setItem("citizenRole", loginPayload.role);
      window.location.href = "/citizen";
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Registration failed");
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
              <p className="brand-title">Citizen Registration</p>
              <p className="brand-subtitle">Identity first, claim next</p>
            </div>
          </Link>
          <div>
            <p className="eyebrow">eKYC Onboarding</p>
            <h1>Create your verified citizen account.</h1>
            <p className="lead">
              Register once, confirm identity, then continue directly to the citizen auto-decision workspace.
            </p>
          </div>
          <div className="auth-flow">
            <span>Register</span>
            <span>OCR</span>
            <span>KYC Confirm</span>
            <span>Login</span>
          </div>
        </section>

        <form className="auth-card neo-card" onSubmit={register}>
          <div className="auth-card-header">
            <span className="badge gold">Backend eKYC Flow</span>
            <h2>Citizen Registration</h2>
            <p>The UI calls the real Auth and eKYC endpoints, then signs the citizen in automatically.</p>
          </div>
          <label className="field">
            <span>National ID</span>
            <input className="input neo-input" inputMode="numeric" value={nationalId} onChange={(event) => setNationalId(event.target.value)} />
          </label>
          <label className="field">
            <span>Phone number</span>
            <input className="input neo-input" inputMode="tel" value={phone} onChange={(event) => setPhone(event.target.value)} />
          </label>
          <label className="field">
            <span>Laser code fallback</span>
            <input className="input neo-input" value={laserCode} onChange={(event) => setLaserCode(event.target.value)} />
          </label>
          <label className="field">
            <span>Create password</span>
            <input className="input neo-input" type="password" value={password} onChange={(event) => setPassword(event.target.value)} />
          </label>
          <button className="btn btn-primary full-width" disabled={loading} type="submit">
            {loading ? "Creating account..." : "Create account and continue"}
          </button>
          {message ? <p className="auth-note">{message}</p> : null}
          <p className="auth-note">
            Already have an account? <Link href="/citizen/login">Sign in</Link>
          </p>
        </form>
      </main>
    </div>
  );
}
