"use client";

import Link from "next/link";
import { FormEvent, useEffect, useMemo, useRef, useState } from "react";

type ClaimStatus = "idle" | "processing" | "approved" | "pending" | "rejected";

type Project = {
  id: string;
  name: string;
  description: string;
  active: boolean;
};

type ProviderState = {
  key: string;
  name: string;
  state: "Waiting" | "Checking" | "Passed";
  detail: string;
};

type ClaimResponse = {
  id: string;
  nationalId: string;
  projectId: string;
  status: Exclude<ClaimStatus, "idle">;
  submittedAt: string;
  updatedAt: string;
  eligibility?: {
    claimId: string;
    status: Exclude<ClaimStatus, "idle">;
    reasons: string[];
  } | null;
  project?: Project | null;
};

const initialProviders: ProviderState[] = [
  { key: "dopa", name: "DOPA", state: "Waiting", detail: "Civil registry and ID card status" },
  { key: "sso", name: "SSO", state: "Waiting", detail: "Insurance status and contribution history" },
  { key: "ktb", name: "KTB", state: "Waiting", detail: "PromptPay account and income profile" },
];

function nowTime() {
  return new Intl.DateTimeFormat("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit" }).format(
    new Date(),
  );
}

export default function CitizenDashboardPage() {
  const [nationalId, setNationalId] = useState("1101700203451");
  const [projects, setProjects] = useState<Project[]>([]);
  const [projectId, setProjectId] = useState("");
  const [claim, setClaim] = useState<ClaimResponse | null>(null);
  const [history, setHistory] = useState<ClaimResponse[]>([]);
  const [status, setStatus] = useState<ClaimStatus>("idle");
  const [providerState, setProviderState] = useState(initialProviders);
  const [events, setEvents] = useState([{ status: "idle", message: "Ready for a new claim", time: "--:--:--" }]);
  const [toast, setToast] = useState("");
  const eventSourceRef = useRef<EventSource | null>(null);

  const progress = useMemo(() => {
    if (status === "idle") return 8;
    if (status === "processing") return 62;
    return 100;
  }, [status]);

  async function loadProjects() {
    const response = await fetch("/api/backend/api/v1/benefit/projects", { cache: "no-store" });
    const payload = await response.json();
    const nextProjects = (payload.projects ?? []) as Project[];
    setProjects(nextProjects);
    if (nextProjects[0]) setProjectId(nextProjects[0].id);
  }

  async function loadHistory(nextNationalId = nationalId) {
    if (!nextNationalId) return;
    const response = await fetch(`/api/backend/api/v1/benefit/history/${nextNationalId}`, { cache: "no-store" });
    const payload = await response.json();
    setHistory((payload.claims ?? []) as ClaimResponse[]);
  }

  useEffect(() => {
    const storedNationalId = localStorage.getItem("citizenNationalId");
    if (storedNationalId) {
      window.setTimeout(() => setNationalId(storedNationalId), 0);
    }
    window.setTimeout(() => {
      void loadProjects();
    }, 0);
    if (storedNationalId) {
      void fetch(`/api/backend/api/v1/benefit/history/${storedNationalId}`, { cache: "no-store" })
        .then((response) => response.json())
        .then((payload) => setHistory((payload.claims ?? []) as ClaimResponse[]));
    }

    return () => eventSourceRef.current?.close();
  }, []);

  async function submitClaim(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!projectId) {
      setToast("Create an active program in Admin Console before submitting a claim");
      return;
    }

    setStatus("processing");
    setClaim(null);
    setToast("Starting auto-decision flow");
    setProviderState(initialProviders.map((provider) => ({ ...provider, state: "Checking" })));
    setEvents([{ status: "processing", message: "Auto-decision flow started", time: nowTime() }]);

    const response = await fetch("/api/backend/api/v1/benefit/claim", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nationalId, projectId }),
    });
    const payload = await response.json();
    if (!response.ok) {
      setToast(payload.error ?? "Claim submission failed");
      setStatus("idle");
      return;
    }

    const createdClaim = payload as ClaimResponse;
    setClaim(createdClaim);
    setStatus(createdClaim.status);
    connectSSE(createdClaim.id);
    await pollClaim(createdClaim.id);
    await loadHistory();
  }

  function connectSSE(claimId: string) {
    eventSourceRef.current?.close();
    const source = new EventSource(`/api/backend/api/v1/claim/${claimId}/stream`);
    eventSourceRef.current = source;

    source.addEventListener("claim-status", (event) => {
      const payload = JSON.parse((event as MessageEvent).data) as { status: ClaimStatus; message: string };
      setStatus(payload.status);
      setEvents((current) => [...current, { status: payload.status, message: payload.message, time: nowTime() }]);
    });
  }

  async function pollClaim(claimId: string) {
    for (let attempt = 0; attempt < 8; attempt += 1) {
      await new Promise((resolve) => setTimeout(resolve, 700));
      const response = await fetch(`/api/backend/api/v1/benefit/status/${claimId}`, { cache: "no-store" });
      if (!response.ok) continue;

      const latest = (await response.json()) as ClaimResponse;
      setClaim(latest);
      setStatus(latest.status);

      if (latest.status !== "processing") {
        setProviderState(initialProviders.map((provider) => ({ ...provider, state: "Passed" })));
        setEvents((current) => [...current, { status: latest.status, message: `Benefit status resolved: ${latest.status}`, time: nowTime() }]);
        eventSourceRef.current?.close();
        return;
      }
    }
  }

  return (
    <div className="app-shell">
      <header className="topbar glass">
        <Link className="brand" href="/">
          <div className="seal">RT</div>
          <div className="brand-text">
            <p className="brand-title">Citizen Portal</p>
            <p className="brand-subtitle">Auto-decision with live status updates</p>
          </div>
        </Link>
        <div className="nav-actions">
          <button className="pill" onClick={() => loadHistory()} type="button">Refresh history</button>
          <Link className="pill" href="/citizen/login">Sign out</Link>
        </div>
      </header>

      <main className="citizen-layout">
        <section className="claim-card glass neo-inset">
          <div className="card-title-row">
            <div>
              <h1 className="page-title">Submit a subsidy claim</h1>
              <p className="caption">
                This form starts the auto-decision flow. Only pending exceptions move to officer review.
              </p>
            </div>
            <span className="badge green">Backend Connected</span>
          </div>
          <form className="form-grid" onSubmit={submitClaim}>
            <label className="field">
              <span>National ID</span>
              <input className="input neo-input" inputMode="numeric" value={nationalId} onChange={(event) => setNationalId(event.target.value)} />
            </label>
            <label className="field">
              <span>Active program</span>
              <select className="select neo-input" value={projectId} onChange={(event) => setProjectId(event.target.value)}>
                {projects.length === 0 ? <option value="">No active programs available</option> : null}
                {projects.map((project) => <option key={project.id} value={project.id}>{project.name}</option>)}
              </select>
            </label>
            <button className="btn btn-primary full-width" type="submit">Run auto-decision check</button>
          </form>
        </section>

        <section className="glass panel">
          <div className="status-header">
            <div>
              <span className="status-label">Auto-decision status</span>
              <div className={`status-value status-${status}`}>{status}</div>
            </div>
            <span className="badge">{claim?.id ?? "No claim yet"}</span>
          </div>
          <div className="progress-track">
            <div className="progress-fill" style={{ width: `${progress}%` }} />
          </div>
          <div className="provider-grid separated-grid">
            {providerState.map((provider) => (
              <div className="provider neo-card" key={provider.key}>
                <div className="provider-top">
                  <span className="provider-name">{provider.name}</span>
                  <span className="check">{provider.state === "Passed" ? "✓" : "…"}</span>
                </div>
                <div className="provider-state">{provider.state}</div>
                <p className="caption">{provider.detail}</p>
              </div>
            ))}
          </div>
          {claim?.eligibility ? (
            <div className="status-card">
              <span className="status-label">Decision reasons</span>
              {claim.eligibility.reasons.map((reason) => (
                <div className="list-row" key={reason}>
                  <span>{reason}</span>
                  <span className="badge green">rule</span>
                </div>
              ))}
            </div>
          ) : null}
          <div className="timeline">
            {events.map((event, index) => (
              <div className="event" key={`${event.message}-${index}`}>
                <span className="event-marker" />
                <div>
                  <strong>{event.message}</strong>
                  <span>{event.time} · {event.status}</span>
                </div>
              </div>
            ))}
          </div>
        </section>

        <section className="glass panel citizen-history">
          <div className="section-title">
            <h2>Claim history</h2>
            <span className="badge">{history.length} claims</span>
          </div>
          <div className="list">
            {history.map((item) => (
              <div className="list-row" key={item.id}>
                <div>
                  <strong>{item.id}</strong>
                  <p className="caption">{item.project?.name ?? item.projectId}</p>
                </div>
                <span className={item.status === "approved" ? "badge green" : "badge gold"}>{item.status}</span>
              </div>
            ))}
            {history.length === 0 ? <p className="caption">No claims yet.</p> : null}
          </div>
        </section>
      </main>

      {toast ? <div className="toast" role="status" onAnimationEnd={() => window.setTimeout(() => setToast(""), 2200)}>{toast}</div> : null}
    </div>
  );
}
