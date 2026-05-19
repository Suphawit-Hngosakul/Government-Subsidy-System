"use client";

import Link from "next/link";
import { FormEvent, useEffect, useState } from "react";

type Project = {
  id: string;
  name: string;
  description: string;
  active: boolean;
};

type OfficerClaim = {
  claimId: string;
  nationalId: string;
  projectId: string;
  status: "pending" | "approved" | "rejected";
  submittedAt: string;
};

type Stats = {
  projects: { total: number; active: number };
  claims: { total: number; pending: number; approved: number; rejected: number };
};

type AuditEntry = {
  id: string;
  at: string;
  actor: string;
  action: string;
  entityId: string;
  metadata?: { reason?: string };
};

const emptyStats: Stats = {
  projects: { total: 0, active: 0 },
  claims: { total: 0, pending: 0, approved: 0, rejected: 0 },
};

export default function AdminDashboardPage() {
  const [stats, setStats] = useState<Stats>(emptyStats);
  const [projects, setProjects] = useState<Project[]>([]);
  const [claims, setClaims] = useState<OfficerClaim[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [projectName, setProjectName] = useState("Household Relief Program");
  const [projectDescription, setProjectDescription] = useState("Financial support for eligible households");
  const [message, setMessage] = useState("");

  useEffect(() => {
    void refreshAdminData();
  }, []);

  async function refreshAdminData() {
    const [statsResponse, projectsResponse, claimsResponse, auditResponse] = await Promise.all([
      fetch("/api/backend/api/v1/admin/stats", { cache: "no-store" }),
      fetch("/api/backend/api/v1/admin/projects", { cache: "no-store" }),
      fetch("/api/backend/api/v1/officer/claims", { cache: "no-store" }),
      fetch("/api/backend/api/v1/admin/audit-log?limit=20", { cache: "no-store" }),
    ]);

    if (statsResponse.ok) setStats((await statsResponse.json()) as Stats);
    if (projectsResponse.ok) {
      const payload = await projectsResponse.json();
      setProjects((payload.projects ?? []) as Project[]);
    }
    if (claimsResponse.ok) {
      const payload = await claimsResponse.json();
      setClaims((payload.claims ?? []) as OfficerClaim[]);
    }
    if (auditResponse.ok) {
      const payload = await auditResponse.json();
      setAudit((payload.entries ?? []) as AuditEntry[]);
    }
  }

  async function createProject(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setMessage("Creating project...");
    const response = await fetch("/api/backend/api/v1/admin/project", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        name: projectName,
        description: projectDescription,
        active: true,
        criteria: {
          minAge: 18,
          maxMonthlyIncome: 30000,
          allowedSsoSections: ["40"],
          requirePromptPay: true,
        },
      }),
    });
    const payload = await response.json().catch(() => ({}));
    setMessage(response.ok ? "Project created from backend" : payload.error ?? "Create project failed");
    await refreshAdminData();
  }

  async function decideClaim(claimId: string, decision: "approve" | "reject") {
    const response = await fetch(`/api/backend/api/v1/officer/claim/${claimId}/${decision}`, {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({
        officerId: "officer-frontend",
        reason: decision === "approve" ? "verified from admin console" : "rejected from admin console",
      }),
    });
    const payload = await response.json().catch(() => ({}));
    setMessage(
      response.ok
        ? `Exception case ${decision === "approve" ? "approved" : "rejected"} manually`
        : payload.error ?? `Manual ${decision} failed`,
    );
    await refreshAdminData();
  }

  return (
    <div className="app-shell admin-area">
      <header className="topbar glass">
        <Link className="brand" href="/">
          <div className="seal">RT</div>
          <div className="brand-text">
            <p className="brand-title">Admin Console</p>
            <p className="brand-subtitle">Exception handling for pending cases only</p>
          </div>
        </Link>
        <div className="nav-actions">
          <button className="pill" onClick={refreshAdminData} type="button">Refresh backend data</button>
          <Link className="pill" href="/admin/login">Sign out</Link>
        </div>
      </header>

      <main className="admin-layout">
        <section className="glass panel admin-hero">
          <span className="badge">Manual Fallback</span>
          <h1 className="page-title">Exception Review Console</h1>
          <p className="lead">Auto-decision is the default path. Officers only handle pending cases that need human judgment.</p>
          <div className="admin-grid">
            <div className="admin-tile neo-card"><span>Total projects</span><strong>{stats.projects.total}</strong></div>
            <div className="admin-tile neo-card"><span>Active projects</span><strong>{stats.projects.active}</strong></div>
            <div className="admin-tile neo-card"><span>Manual review</span><strong>{stats.claims.pending}</strong></div>
            <div className="admin-tile neo-card"><span>Resolved approved</span><strong>{stats.claims.approved}</strong></div>
          </div>
        </section>

        <section className="glass panel">
          <div className="section-title">
            <h2>Project Management</h2>
            <span className="badge green">{projects.length} loaded</span>
          </div>
          <form className="form-grid admin-create-form" onSubmit={createProject}>
            <label className="field">
              <span>Project name</span>
              <input className="input neo-input" value={projectName} onChange={(event) => setProjectName(event.target.value)} />
            </label>
            <label className="field">
              <span>Description</span>
              <input className="input neo-input" value={projectDescription} onChange={(event) => setProjectDescription(event.target.value)} />
            </label>
            <button className="btn btn-secondary" type="submit">Create project via backend</button>
          </form>
          <div className="list">
            {projects.map((project) => (
              <div className="list-row" key={project.id}>
                <div>
                  <strong>{project.name}</strong>
                  <p className="caption">{project.description || project.id}</p>
                </div>
                <span className={project.active ? "badge green" : "badge gold"}>{project.active ? "Active" : "Inactive"}</span>
              </div>
            ))}
          </div>
        </section>

        <section className="glass panel">
          <div className="section-title">
            <h2>Manual Review Queue</h2>
            <span className="badge gold">Pending exceptions</span>
          </div>
          <div className="list">
            {claims.map((claim) => (
              <div className="list-row" key={claim.claimId}>
                <div>
                  <strong>{claim.claimId}</strong>
                  <p className="caption">{claim.nationalId} · {claim.projectId}</p>
                </div>
                <div className="button-row compact-actions">
                  <button className="btn btn-secondary" onClick={() => decideClaim(claim.claimId, "approve")} type="button">Approve manually</button>
                  <button className="btn btn-secondary" onClick={() => decideClaim(claim.claimId, "reject")} type="button">Reject manually</button>
                </div>
              </div>
            ))}
            {claims.length === 0 ? <p className="caption">No pending exception cases.</p> : null}
          </div>
        </section>

        <section className="glass panel">
          <div className="section-title">
            <h2>Audit Trail</h2>
            <span className="badge">Append-only</span>
          </div>
          <div className="audit">
            {audit.map((entry) => (
              <div className="audit-row" key={entry.id}>
                <span className="audit-code">{entry.action.split(".").at(-1)?.slice(0, 3).toUpperCase() ?? "LOG"}</span>
                <div>
                  <strong>{entry.action} · {entry.entityId}</strong>
                  <span>{entry.actor} · {entry.metadata?.reason ?? "no reason"} · {new Date(entry.at).toLocaleString("en-US")}</span>
                </div>
              </div>
            ))}
            {audit.length === 0 ? <p className="caption">No audit entries yet.</p> : null}
          </div>
        </section>
      </main>

      {message ? <div className="toast" role="status" onAnimationEnd={() => window.setTimeout(() => setMessage(""), 2200)}>{message}</div> : null}
    </div>
  );
}
