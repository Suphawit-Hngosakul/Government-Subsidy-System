import Link from "next/link";
import Image from "next/image";

const services = [
  {
    title: "Citizen Claim",
    body: "Submit a subsidy request, run auto-decision rules, and track status in real time.",
    href: "/citizen/login",
    tag: "Citizen",
  },
  {
    title: "eKYC Registration",
    body: "Create an account and verify identity with the Auth and eKYC backend.",
    href: "/citizen/register",
    tag: "Identity",
  },
  {
    title: "Officer Review",
    body: "Handle only pending exception cases that require manual approve or reject decisions.",
    href: "/admin/login",
    tag: "Officer",
  },
];

const stats = [
  ["3", "Connected agencies"],
  ["100%", "Integrated backend"],
  ["SSE", "Live status stream"],
];

const quickFilters = [
  ["Workspace", "Citizen"],
  ["Program", "Household Relief"],
  ["Status", "Auto-decision"],
  ["Action", "Exception fallback"],
];

export default function Home() {
  return (
    <div className="app-shell landing-shell">
      <header className="landing-nav">
        <Link className="brand" href="/">
          <div className="seal">RT</div>
          <div className="brand-text">
            <p className="brand-title">Ruam Thai Srang Chati</p>
            <p className="brand-subtitle">Government Subsidy System</p>
          </div>
        </Link>

        <nav className="nav-menu" aria-label="Primary navigation">
          <Link href="/citizen/login">Citizen</Link>
          <Link href="/citizen/register">Register</Link>
          <Link href="/admin/login">Admin</Link>
          <a href="#services">Services</a>
        </nav>

        <Link className="nav-cta" href="/citizen/login">
          Start Claim
        </Link>
      </header>

      <main>
        <section className="landing-hero">
          <div className="hero-copy">
            <p className="eyebrow">Digital Welfare Infrastructure</p>
            <h1>
              Public benefits,
              <span> verified in one flow.</span>
            </h1>
            <p className="lead">
              A production-ready service surface for citizen claims, auto-decision rules, exception review,
              and real-time eligibility checks across DOPA, SSO, and KTB.
            </p>
            <div className="hero-actions">
              <Link className="btn btn-primary hero-button" href="/citizen/login">
                Apply for benefits
              </Link>
              <Link className="play-action" href="/admin/login" aria-label="Open admin console">
                <span className="play-ring">→</span>
                Admin Console
              </Link>
            </div>
          </div>

          <div className="hero-visual" aria-label="Digital government service preview">
            <div className="paint-stroke" />
            <Image
              className="welfare-mark"
              src="/digital-welfare-mark.svg"
              alt=""
              width={520}
              height={520}
              priority
            />
            <div className="visual-band" />
            <div className="visual-orb orb-one" />
            <div className="visual-orb orb-two" />
          </div>

          <aside className="hero-stats" aria-label="System highlights">
            {stats.map(([value, label]) => (
              <div className="stat-row" key={label}>
                <span className="stat-icon">✦</span>
                <div>
                  <strong>{value}</strong>
                  <span>{label}</span>
                </div>
              </div>
            ))}
          </aside>
        </section>

        <section className="quick-panel neo-card" aria-label="Quick service search">
          {quickFilters.map(([label, value]) => (
            <div className="quick-field" key={label}>
              <span>{label}</span>
              <strong>{value}</strong>
            </div>
          ))}
          <Link className="quick-search" href="/citizen/login">
            Search
          </Link>
        </section>

        <section className="services-section" id="services">
          <div className="section-heading">
            <div>
              <h2>Recently connected services</h2>
              <p>Core workflows are separated by role while sharing one integrated backend.</p>
            </div>
            <div className="slider-actions" aria-hidden="true">
              <span>←</span>
              <span>→</span>
            </div>
          </div>

          <div className="service-cards">
            {services.map((service) => (
              <Link className="service-card neo-card" href={service.href} key={service.title}>
                <div className="service-image">
                  <span>{service.tag}</span>
                </div>
                <h3>{service.title}</h3>
                <p>{service.body}</p>
              </Link>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}
