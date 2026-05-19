import Link from "next/link";

export default function AdminLoginPage() {
  return (
    <div className="app-shell auth-shell admin-auth">
      <main className="auth-layout">
        <section className="auth-hero glass">
          <Link className="brand" href="/">
            <div className="seal">RT</div>
            <div className="brand-text">
              <p className="brand-title">Admin Console</p>
              <p className="brand-subtitle">Manual fallback and program operations</p>
            </div>
          </Link>
          <div>
            <p className="eyebrow">Officer Workspace</p>
            <h1>Manage programs and exception cases.</h1>
            <p className="lead">
              Auto-decision handles normal claims. Officers use this workspace for projects, audit trails, and pending exceptions.
            </p>
          </div>
          <div className="auth-proof">
            <span>Projects</span>
            <span>Manual Review</span>
            <span>Audit Trail</span>
          </div>
        </section>

        <form className="auth-card neo-card">
          <div className="auth-card-header">
            <span className="badge">Officer Login</span>
            <h2>Admin Sign In</h2>
            <p>Enter the separated officer area for exception handling and project management.</p>
          </div>
          <label className="field">
            <span>Officer username</span>
            <input className="input neo-input" placeholder="policy.officer" />
          </label>
          <label className="field">
            <span>Password</span>
            <input className="input neo-input" type="password" placeholder="Password" />
          </label>
          <label className="field">
            <span>Agency code</span>
            <input className="input neo-input" placeholder="GOV-SUBSIDY-01" />
          </label>
          <Link className="btn btn-primary full-width" href="/admin">
            Continue to Admin Console
          </Link>
          <p className="auth-note">
            Back to role selection <Link href="/">Home</Link>
          </p>
        </form>
      </main>
    </div>
  );
}
