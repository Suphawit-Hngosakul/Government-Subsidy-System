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
  state: "Waiting" | "Checking" | "Passed" | "Failed";
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
  { key: "dopa", name: "DOPA (กรมการปกครอง)", state: "Waiting", detail: "ตรวจสอบสถานะบุคคลภาพและบัตรประชาชน" },
  { key: "sso", name: "SSO (สำนักงานประกันสังคม)", state: "Waiting", detail: "ตรวจสอบการขึ้นทะเบียนผู้ประกันตน มาตรา 33/39/40" },
  { key: "ktb", name: "KTB (ธนาคารกรุงไทย)", state: "Waiting", detail: "ตรวจสอบบัญชีรับเงินผูกพร้อมเพย์และรายได้เฉลี่ย" },
];

function nowTime() {
  return new Intl.DateTimeFormat("en-US", { hour: "2-digit", minute: "2-digit", second: "2-digit", hour12: false }).format(
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
  const [events, setEvents] = useState([{ status: "idle", message: "พร้อมสำหรับการยื่นสิทธิ์โครงการ", time: "--:--:--" }]);
  const [toast, setToast] = useState("");
  const [activeTab, setActiveTab] = useState<"claim" | "history" | "info">("claim");
  const eventSourceRef = useRef<EventSource | null>(null);

  const maskedNationalId = useMemo(() => {
    if (!nationalId) return "";
    const clean = nationalId.replace(/\D/g, "");
    if (clean.length < 13) return nationalId;
    return `${clean[0]}-${clean.slice(1, 5)}-XXXXX-${clean.slice(10, 12)}-${clean[12]}`;
  }, [nationalId]);

  const selectedProjectDescription = useMemo(() => {
    const proj = projects.find((p) => p.id === projectId);
    return proj ? proj.description : "กรุณาเลือกโครงการที่ต้องการยื่นสิทธิ์";
  }, [projectId, projects]);

  const progress = useMemo(() => {
    if (status === "idle") return 0;
    if (status === "processing") return 65;
    return 100;
  }, [status]);

  async function loadProjects() {
    try {
      const response = await fetch("/api/backend/api/v1/benefit/projects", { cache: "no-store" });
      const payload = await response.json();
      const nextProjects = (payload.projects ?? []) as Project[];
      setProjects(nextProjects);
      if (nextProjects[0]) setProjectId(nextProjects[0].id);
    } catch {
      setToast("เชื่อมต่อข้อมูลโครงการไม่สำเร็จ");
    }
  }

  async function loadHistory(nextNationalId = nationalId) {
    if (!nextNationalId) return;
    try {
      const response = await fetch(`/api/backend/api/v1/benefit/history/${nextNationalId}`, { cache: "no-store" });
      const payload = await response.json();
      setHistory((payload.claims ?? []) as ClaimResponse[]);
    } catch {
      // Swallowed gracefully
    }
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

  // ดึงประวัติคำร้องอัตโนมัติเมื่อกรอกเลขบัตรประชาชนครบ 13 หลัก
  useEffect(() => {
    const cleanId = nationalId.replace(/\D/g, "");
    if (cleanId.length === 13) {
      localStorage.setItem("citizenNationalId", cleanId);
      void loadHistory(cleanId);
    }
  }, [nationalId]);

  // กู้คืนสถานะการยื่นสิทธิ์โครงการอัตโนมัติจากข้อมูลประวัติ (State Auto-Restoration)
  useEffect(() => {
    if (!projectId || history.length === 0) {
      setClaim(null);
      setStatus("idle");
      setProviderState(initialProviders);
      setEvents([{ status: "idle", message: "พร้อมสำหรับการยื่นสิทธิ์โครงการ", time: "--:--:--" }]);
      return;
    }

    const projectClaims = history.filter((h) => h.projectId === projectId);
    if (projectClaims.length > 0) {
      // ค้นหาคำร้องสิทธิ์ล่าสุดสำหรับโครงการนี้ (อิงตามเวลา submittedAt ล่าสุด)
      const latestClaim = projectClaims.reduce((prev, current) => {
        const timePrev = new Date(prev.submittedAt).getTime();
        const timeCurr = new Date(current.submittedAt).getTime();
        return timePrev > timeCurr ? prev : current;
      });

      setClaim(latestClaim);
      setStatus(latestClaim.status);

      // กู้คืนสถานะผู้ให้บริการตรวจสอบสิทธิ์ (DOPA, SSO, KTB) ตามเหตุผลความล้มเหลว (ถ้าปฏิเสธ)
      setProviderState(
        initialProviders.map((p) => {
          let nextState: "Passed" | "Failed" | "Checking" | "Waiting" = "Passed";
          if (latestClaim.status === "rejected") {
            const reasons = latestClaim.eligibility?.reasons ?? [];
            const isDopaReject = reasons.some((r) => r.includes("citizen identity") || r.includes("citizen age"));
            const isSsoReject = reasons.some((r) => r.includes("insured under"));
            const isKtbReject = reasons.some((r) => r.includes("monthly income") || r.includes("deposit balance") || r.includes("PromptPay"));
            
            if (p.key === "dopa" && isDopaReject) nextState = "Failed";
            if (p.key === "sso" && isSsoReject) nextState = "Failed";
            if (p.key === "ktb" && isKtbReject) nextState = "Failed";
          } else if (latestClaim.status === "processing") {
            nextState = "Checking";
          } else if (latestClaim.status === "pending") {
            // Pending หมายถึงผ่านด่านอัตโนมัติหมดแล้ว แต่รอการอนุมัติแบบ manual เพิ่มเติม
            nextState = "Passed";
          }
          return { ...p, state: nextState };
        })
      );

      // สร้าง Live Logs กู้คืนสถานะจริงล่าสุด
      const displayTime = new Intl.DateTimeFormat("en-US", {
        hour: "2-digit",
        minute: "2-digit",
        second: "2-digit",
        hour12: false,
      }).format(new Date(latestClaim.updatedAt || latestClaim.submittedAt));

      setEvents([
        {
          status: latestClaim.status,
          message: `กู้คืนผลคำร้องล่าสุดสำหรับโครงการ: ${
            latestClaim.status === "approved"
              ? "อนุมัติผ่านสิทธิ์เรียบร้อย"
              : latestClaim.status === "rejected"
              ? "ปฏิเสธ/ไม่ผ่านเกณฑ์การประเมิน"
              : latestClaim.status === "pending"
              ? "อยู่ระหว่างรอพิจารณาข้อยกเว้นพิเศษ"
              : "กำลังประมวลผลข้อมูลอัตโนมัติ"
          }`,
          time: displayTime,
        },
      ]);
    } else {
      setClaim(null);
      setStatus("idle");
      setProviderState(initialProviders);
      setEvents([{ status: "idle", message: "พร้อมสำหรับการยื่นสิทธิ์โครงการ", time: "--:--:--" }]);
    }
  }, [projectId, history]);

  async function submitClaim(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!projectId) {
      setToast("กรุณาเลือกโครงการที่เปิดรับสมัครก่อนยื่นคำร้อง");
      return;
    }

    setStatus("processing");
    setClaim(null);
    setToast("กำลังเชื่อมต่อระบบพิจารณาผลอัตโนมัติ...");
    setProviderState(initialProviders.map((p) => ({ ...p, state: "Checking" })));
    setEvents([{ status: "processing", message: "ระบบเริ่มประมวลผลการยื่นสิทธิ์อัตโนมัติ", time: nowTime() }]);

    const response = await fetch("/api/backend/api/v1/benefit/claim", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ nationalId, projectId }),
    });
    const payload = await response.json();
    if (!response.ok) {
      setToast(payload.error ?? "ยื่นคำขอสิทธิ์ไม่สำเร็จ");
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
        setProviderState(
          initialProviders.map((p) => {
            let nextState: "Passed" | "Failed" = "Passed";
            // Check if there was any rejection related to this provider
            if (latest.status === "rejected") {
              const reasons = latest.eligibility?.reasons ?? [];
              const isDopaReject = reasons.some((r) => r.includes("citizen identity") || r.includes("citizen age"));
              const isSsoReject = reasons.some((r) => r.includes("insured under"));
              const isKtbReject = reasons.some((r) => r.includes("monthly income") || r.includes("deposit balance") || r.includes("PromptPay"));
              
              if (p.key === "dopa" && isDopaReject) nextState = "Failed";
              if (p.key === "sso" && isSsoReject) nextState = "Failed";
              if (p.key === "ktb" && isKtbReject) nextState = "Failed";
            }
            return { ...p, state: nextState };
          })
        );
        setEvents((current) => [...current, { status: latest.status, message: `ผลการพิจารณาสิ้นสุด: ${latest.status === "approved" ? "ผ่านเกณฑ์" : latest.status === "rejected" ? "ไม่ผ่านเกณฑ์" : "อยู่ระหว่างพิจารณาชดเชย"}`, time: nowTime() }]);
        eventSourceRef.current?.close();
        return;
      }
    }
  }

  function handleLogout() {
    localStorage.removeItem("citizenNationalId");
    localStorage.removeItem("citizenToken");
    localStorage.removeItem("citizenRole");
    setNationalId("");
    setClaim(null);
    setStatus("idle");
    setHistory([]);
    setProviderState(initialProviders);
    setEvents([{ status: "idle", message: "พร้อมสำหรับการยื่นสิทธิ์โครงการ", time: "--:--:--" }]);
    setToast("ออกจากระบบเรียบร้อยแล้ว");
    
    window.setTimeout(() => {
      window.location.href = "/citizen/login";
    }, 400);
  }

  return (
    <div className="app-shell pao-shell">
      {/* Top Main Navigation Bar for Global Routing */}
      <header className="topbar glass pao-top-nav">
        <Link className="brand" href="/">
          <div className="seal">RT</div>
          <div className="brand-text">
            <p className="brand-title">Ruam Thai Portal</p>
            <p className="brand-subtitle">Government Services</p>
          </div>
        </Link>
        <div className="nav-actions">
          <Link className="pill pao-pill-active" href="/citizen">Citizen Area</Link>
          <Link className="pill" href="/admin">Admin Console</Link>
        </div>
      </header>

      {/* Pao Tang Style Centered Mobile Wallet Container */}
      <div className="pao-container">
        {/* Header bar blue gradient */}
        <div className="pao-app-header">
          <div className="pao-user-row">
            <div className="pao-user-avatar">🧑</div>
            <div className="pao-user-info">
              <span className="pao-hello">สวัสดี ประชาชนผู้มีสิทธิ์</span>
              <h2 className="pao-uid">{maskedNationalId || "รอกำหนดรหัสบัตรประชาชน"}</h2>
            </div>
            <div style={{ display: "flex", flexDirection: "column", alignItems: "flex-end", gap: "6px" }}>
              <span className="pao-badge-ekyc">✓ eKYC ยืนยันสิทธิ์</span>
              <button 
                onClick={handleLogout} 
                className="pao-logout-btn"
                type="button"
              >
                🚪 ออกจากระบบ
              </button>
            </div>
          </div>
          <div className="pao-balance-tag">
            <span>ถุงเงินประชารัฐ · G-Wallet สวัสดิการภาครัฐ</span>
          </div>
        </div>

        {/* G-Wallet Metallic Luxury Card */}
        <div className="pao-wallet-card-container">
          <div className="pao-wallet-card">
            <div className="pao-card-brand">
              <span className="pao-flag">🇹🇭</span>
              <span className="pao-wallet-title">G-Wallet สิทธิ์ช่วยเหลือ</span>
            </div>
            <div className="pao-card-main">
              <p className="pao-card-lbl">โครงการเปิดรับสมัครล่าสุด</p>
              <h3 className="pao-card-campaign">
                {projects.length > 0 ? projects.find((p) => p.id === projectId)?.name ?? "คนละครึ่ง / บรรเทาทุกข์" : "รอผู้ดูแลระบบสร้างโครงการ"}
              </h3>
            </div>
            <div className="pao-card-footer">
              <span className="pao-card-id-label">สิทธิ์ประโยชน์เฉพาะบุคคล</span>
              <span className="pao-card-chip"></span>
            </div>
          </div>
        </div>

        {/* Quick Menu Grid */}
        <div className="pao-quick-grid">
          <button
            className={`pao-quick-item ${activeTab === "claim" ? "active" : ""}`}
            onClick={() => setActiveTab("claim")}
            type="button"
          >
            <div className="pao-quick-icon">📝</div>
            <span>ยื่นสิทธิ์ประชารัฐ</span>
          </button>
          <button
            className={`pao-quick-item ${activeTab === "history" ? "active" : ""}`}
            onClick={() => setActiveTab("history")}
            type="button"
          >
            <div className="pao-quick-icon">📜</div>
            <span>ประวัติสวัสดิการ</span>
          </button>
          <button
            className={`pao-quick-item ${activeTab === "info" ? "active" : ""}`}
            onClick={() => setActiveTab("info")}
            type="button"
          >
            <div className="pao-quick-icon">🏢</div>
            <span>ข้อมูลหน่วยงานรัฐ</span>
          </button>
          <button
            className="pao-quick-item"
            onClick={() => {
              void loadProjects();
              void loadHistory();
              setToast("ปรับปรุงข้อมูลจากหน่วยงานรัฐเรียบร้อย");
            }}
            type="button"
          >
            <div className="pao-quick-icon">🔄</div>
            <span>ซิงค์ข้อมูล</span>
          </button>
        </div>

        {/* Main Content Area */}
        <div className="pao-body">
          {activeTab === "claim" && (
            <div className="pao-pane">
              {/* Form Input Section */}
              <section className="pao-card shadow-sm">
                <div className="pao-pane-title">
                  <h3>กรอกข้อมูลเพื่อประมวลผลคำขอสิทธิ์</h3>
                  <span className="pao-indicator green">เซิร์ฟเวอร์ออนไลน์</span>
                </div>
                <form className="pao-form" onSubmit={submitClaim}>
                  <label className="pao-label-field">
                    <span>เลขประจำตัวประชาชน (13 หลัก)</span>
                    <input
                      className="pao-input"
                      inputMode="numeric"
                      maxLength={13}
                      value={nationalId}
                      onChange={(event) => setNationalId(event.target.value)}
                    />
                  </label>
                  <label className="pao-label-field">
                    <span>เลือกโครงการสวัสดิการภาครัฐ</span>
                    <select
                      className="pao-select"
                      value={projectId}
                      onChange={(event) => setProjectId(event.target.value)}
                    >
                      {projects.length === 0 ? <option value="">ไม่มีโครงการสวัสดิการที่เปิดรับในขณะนี้</option> : null}
                      {projects.map((p) => (
                        <option key={p.id} value={p.id}>{p.name}</option>
                      ))}
                    </select>
                  </label>

                  <div className="pao-campaign-preview">
                    <strong>รายละเอียดของสิทธิ์:</strong>
                    <p>{selectedProjectDescription}</p>
                  </div>

                  <button
                    className="pao-btn-primary"
                    disabled={status === "processing"}
                    type="submit"
                  >
                    {status === "processing" ? "กำลังหมุนตรวจสอบ..." : "สมัครโครงการและตรวจสอบสิทธิ์"}
                  </button>
                </form>
              </section>

              {/* Status Visual Tracker Engine */}
              {status !== "idle" && (
                <section className="pao-card pao-status-engine shadow-sm">
                  {/* Absolute Stamped Seal representation for verified results */}
                  {status === "approved" && (
                    <div className="pao-stamp approved">อนุมัติสิทธิ์<br/><span>APPROVED</span></div>
                  )}
                  {status === "rejected" && (
                    <div className="pao-stamp rejected">ไม่ผ่านเกณฑ์<br/><span>REJECTED</span></div>
                  )}
                  {status === "pending" && (
                    <div className="pao-stamp pending">รอการตรวจสอบ<br/><span>PENDING</span></div>
                  )}

                  <div className="pao-status-header">
                    <div>
                      <span className="pao-lbl-sec">สถานะระบบพิจารณาอัตโนมัติ</span>
                      <div className={`pao-status-text status-${status}`}>
                        {status === "processing" ? "⏳ อยู่ระหว่างประมวลผล..." : ""}
                        {status === "approved" ? "🟢 อนุมัติสิทธิ์เรียบร้อย" : ""}
                        {status === "rejected" ? "🔴 ไม่ผ่านเกณฑ์พิจารณา" : ""}
                        {status === "pending" ? "🟡 อยู่ระหว่างพิจารณาข้อยกเว้น" : ""}
                      </div>
                    </div>
                    <span className="pao-badge-id">คำร้อง: {claim?.id ? claim.id.slice(0, 8) : "กำลังขอ ID..."}</span>
                  </div>

                  {/* Progress fill */}
                  <div className="pao-progress-bar">
                    <div className="pao-progress-fill" style={{ width: `${progress}%` }} />
                  </div>

                  {/* Provider List showing connected state checkmarks */}
                  <div className="pao-checks-grid">
                    {providerState.map((p) => (
                      <div className="pao-check-row" key={p.key}>
                        <div className="pao-check-info">
                          <span className="pao-check-title">{p.name}</span>
                          <span className="pao-check-desc">{p.detail}</span>
                        </div>
                        <div className={`pao-check-status status-${p.state}`}>
                          {p.state === "Passed" && "✓ ผ่านการเช็ค"}
                          {p.state === "Failed" && "✗ ไม่ผ่านเกณฑ์"}
                          {p.state === "Checking" && "🔍 กำลังสแกน..."}
                          {p.state === "Waiting" && "⏳ รอคิว"}
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Decision reasons block */}
                  {claim?.eligibility && (
                    <div className="pao-reasons-block">
                      <strong>เหตุผลประกอบการวินิจฉัยสิทธิ์:</strong>
                      {claim.eligibility.reasons.map((reason) => (
                        <div className="pao-reason-row" key={reason}>
                          <span>• {reason}</span>
                          <span className="badge-rule">เกณฑ์วินิจฉัย</span>
                        </div>
                      ))}
                    </div>
                  )}

                  {/* Timeline Logs block */}
                  <div className="pao-timeline">
                    <h4>ความเคลื่อนไหวล่าสุด (Live Logs)</h4>
                    {events.slice(-3).map((event, index) => (
                      <div className="pao-timeline-event" key={`${event.message}-${index}`}>
                        <div className="pao-marker" />
                        <div className="pao-event-detail">
                          <p>{event.message}</p>
                          <span>{event.time} · สถานะ: {event.status}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </section>
              )}
            </div>
          )}

          {activeTab === "history" && (
            <div className="pao-pane">
              <section className="pao-card shadow-sm">
                <div className="pao-pane-title">
                  <h3>ประวัติการยื่นขอสิทธิ์สวัสดิการ</h3>
                  <span className="badge-count">{history.length} โครงการ</span>
                </div>
                <div className="pao-history-list">
                  {history.map((item) => (
                    <div className="pao-history-row" key={item.id}>
                      <div className="pao-row-details">
                        <strong>{item.project?.name ?? item.projectId}</strong>
                        <span>รหัสคำร้อง: {item.id}</span>
                        <span className="pao-row-time">ยื่นสิทธิ์เมื่อ: {new Date(item.submittedAt).toLocaleDateString("th-TH")}</span>
                      </div>
                      <span className={`pao-row-badge ${item.status}`}>
                        {item.status === "approved" ? "ผ่านสิทธิ์" : ""}
                        {item.status === "rejected" ? "ปฏิเสธ" : ""}
                        {item.status === "pending" ? "รอตรวจเพิ่ม" : ""}
                        {item.status === "processing" ? "กำลังตรวจ" : ""}
                      </span>
                    </div>
                  ))}
                  {history.length === 0 ? (
                    <div className="pao-empty">
                      <p>ไม่พบประวัติการยื่นสิทธิ์โครงการใด ๆ ของรหัสบัตรประชาชนนี้</p>
                    </div>
                  ) : null}
                </div>
              </section>
            </div>
          )}

          {activeTab === "info" && (
            <div className="pao-pane">
              <section className="pao-card shadow-sm">
                <div className="pao-pane-title">
                  <h3>หน่วยงานเชื่อมต่อโครงสร้างพื้นฐาน</h3>
                </div>
                <div className="pao-info-list">
                  <div className="pao-info-row">
                    <strong>กรมการปกครอง (DOPA)</strong>
                    <p>ตรวจสอบสิทธิความเป็นพลเมืองไทยทางกฎหมาย ปัจจุบันทำงานผ่าน Gateway แบบเรียลไทม์</p>
                  </div>
                  <div className="pao-info-row">
                    <strong>สำนักงานประกันสังคม (SSO)</strong>
                    <p>คัดแยกส่วนผู้ใช้แรงงาน โดย blacklists สำหรับมาตรา 33 หรือคัดกรองช่วยเหลือเฉพาะมาตรา 39 และ 40 ตามเกณฑ์ที่ผู้ดูแลระบบตั้งค่า</p>
                  </div>
                  <div className="pao-info-row">
                    <strong>ธนาคารกรุงไทย (KTB)</strong>
                    <p>ยืนยันบัญชีเงินฝากตรวจสอบความเชื่อมโยงกับ พร้อมเพย์ และประเมินรายรับรายเดือนโดยเฉลี่ย</p>
                  </div>
                </div>
              </section>
            </div>
          )}
        </div>
      </div>

      {toast ? (
        <div className="pao-toast" role="status" onAnimationEnd={() => window.setTimeout(() => setToast(""), 2200)}>
          {toast}
        </div>
      ) : null}

      {/* Styled JSX Styles specifically block for this premium Pao Tang interface */}
      <style jsx global>{`
        /* Pao Tang Design Variables mapping with original colors */
        .pao-shell {
          background: #f1ede2 !important; /* Premium light cream */
        }
        
        .pao-top-nav {
          max-width: 520px !important;
          margin: 0 auto 12px !important;
        }

        .pao-pill-active {
          border-color: var(--indigo) !important;
          background: var(--blue-soft) !important;
          color: var(--indigo) !important;
          font-weight: 700 !important;
          box-shadow: 0 2px 8px rgba(31,65,109,0.08) !important;
        }

        .pao-container {
          max-width: 520px;
          margin: 0 auto;
          background: #f7f5ef;
          min-height: calc(100vh - 120px);
          border-radius: 20px;
          border: 1px solid rgba(135,117,78,0.18);
          box-shadow: 0 15px 45px rgba(23,32,51,0.08);
          overflow: hidden;
          font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "Noto Sans Thai", sans-serif;
          position: relative;
        }

        /* App bar style */
        .pao-app-header {
          background: linear-gradient(135deg, var(--indigo) 0%, #112a4e 100%);
          color: #fff;
          padding: 26px 20px 76px 20px;
          border-bottom-left-radius: 24px;
          border-bottom-right-radius: 24px;
          position: relative;
        }

        .pao-user-row {
          display: flex;
          align-items: center;
          gap: 12px;
          position: relative;
        }

        .pao-user-avatar {
          width: 46px;
          height: 46px;
          border-radius: 50%;
          background: rgba(255,255,255,0.18);
          display: grid;
          place-items: center;
          font-size: 22px;
          border: 1.5px solid rgba(255,255,255,0.4);
        }

        .pao-user-info {
          flex: 1;
        }

        .pao-hello {
          font-size: 11px;
          color: rgba(255,255,255,0.7);
          display: block;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .pao-uid {
          font-size: 17px;
          font-weight: 700;
          margin: 2px 0 0 0;
          color: #fff;
          letter-spacing: 0.5px;
        }

        .pao-badge-ekyc {
          background: var(--jade);
          color: #fff;
          font-size: 10px;
          padding: 4px 8px;
          border-radius: 99px;
          font-weight: 600;
          letter-spacing: 0.3px;
        }

        .pao-logout-btn {
          background: rgba(255, 255, 255, 0.12);
          border: 1px solid rgba(255, 255, 255, 0.22);
          color: #fff;
          font-size: 9.5px;
          padding: 4px 8px;
          border-radius: 99px;
          font-weight: 600;
          cursor: pointer;
          transition: background-color 0.2s, border-color 0.2s, transform 0.1s;
          outline: none;
        }

        .pao-logout-btn:hover {
          background: rgba(255, 255, 255, 0.22);
          border-color: rgba(255, 255, 255, 0.4);
        }

        .pao-logout-btn:active {
          transform: scale(0.96);
        }

        .pao-balance-tag {
          margin-top: 14px;
          font-size: 11px;
          color: rgba(255,255,255,0.65);
          display: flex;
          align-items: center;
          gap: 6px;
        }

        /* G Wallet Card */
        .pao-wallet-card-container {
          padding: 0 18px;
          margin-top: -54px;
          position: relative;
          z-index: 10;
        }

        .pao-wallet-card {
          background: linear-gradient(135deg, var(--indigo) 0%, var(--gold) 100%);
          color: #fff;
          border-radius: 16px;
          padding: 20px;
          box-shadow: 0 10px 24px rgba(31,65,109,0.22);
          position: relative;
          overflow: hidden;
          border: 1px solid rgba(255,255,255,0.15);
          display: flex;
          flex-direction: column;
          gap: 22px;
        }

        .pao-wallet-card::before {
          content: "";
          position: absolute;
          width: 250px;
          height: 250px;
          top: -70%;
          right: -40%;
          background: radial-gradient(circle, rgba(255,255,255,0.14) 0%, transparent 65%);
          border-radius: 50%;
          pointer-events: none;
        }

        .pao-card-brand {
          display: flex;
          align-items: center;
          gap: 8px;
        }

        .pao-flag {
          font-size: 18px;
        }

        .pao-wallet-title {
          font-size: 12px;
          font-weight: 600;
          letter-spacing: 0.5px;
          color: rgba(255,255,255,0.9);
        }

        .pao-card-lbl {
          font-size: 11px;
          color: rgba(255,255,255,0.7);
          margin: 0 0 4px 0;
          text-transform: uppercase;
        }

        .pao-card-campaign {
          font-size: 19px;
          font-weight: 750;
          margin: 0;
          color: #fff;
          text-shadow: 0 2px 4px rgba(0,0,0,0.15);
        }

        .pao-card-footer {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 10px;
          color: rgba(255,255,255,0.6);
        }

        .pao-card-chip {
          width: 32px;
          height: 24px;
          background: linear-gradient(135deg, #e3c07b 0%, #a47621 100%);
          border-radius: 4px;
          position: relative;
        }

        /* Quick Grid */
        .pao-quick-grid {
          display: grid;
          grid-template-columns: repeat(4, 1fr);
          gap: 8px;
          padding: 16px 18px 8px 18px;
        }

        .pao-quick-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          gap: 6px;
          background: transparent;
          cursor: pointer;
          transition: transform 0.2s cubic-bezier(0.4, 0, 0.2, 1);
          border: 0;
          outline: none;
        }

        .pao-quick-item:hover {
          transform: translateY(-2px);
        }

        .pao-quick-icon {
          width: 44px;
          height: 44px;
          background: #fff;
          border-radius: 12px;
          box-shadow: 0 6px 12px rgba(23,32,51,0.05);
          display: grid;
          place-items: center;
          font-size: 20px;
          transition: background-color 0.2s;
          border: 1px solid rgba(135,117,78,0.1);
        }

        .pao-quick-item.active .pao-quick-icon {
          background: var(--blue-soft);
          border-color: var(--indigo);
        }

        .pao-quick-item span {
          font-size: 10.5px;
          font-weight: 600;
          color: var(--ink);
          text-align: center;
        }

        /* Content Area */
        .pao-body {
          padding: 8px 18px 24px 18px;
        }

        .pao-card {
          background: #fff;
          border-radius: 16px;
          padding: 18px;
          margin-bottom: 16px;
          border: 1px solid rgba(135,117,78,0.12);
          position: relative;
        }

        .pao-pane-title {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          border-bottom: 1px solid rgba(23,32,51,0.06);
          padding-bottom: 10px;
        }

        .pao-pane-title h3 {
          margin: 0;
          font-size: 13.5px;
          font-weight: 700;
          color: var(--indigo);
        }

        .pao-indicator {
          font-size: 10px;
          padding: 3px 8px;
          border-radius: 99px;
          font-weight: 600;
        }

        .pao-indicator.green {
          background: rgba(31,122,104,0.12);
          color: var(--jade);
        }

        /* Form */
        .pao-form {
          display: flex;
          flex-direction: column;
          gap: 14px;
        }

        .pao-label-field {
          display: flex;
          flex-direction: column;
          gap: 6px;
        }

        .pao-label-field span {
          font-size: 11.5px;
          font-weight: 600;
          color: var(--muted);
        }

        .pao-input,
        .pao-select {
          width: 100%;
          border: 1px solid rgba(135,117,78,0.22);
          border-radius: 8px;
          padding: 10px 12px;
          font-size: 13.5px;
          color: var(--ink);
          background: rgba(247,245,239,0.4);
          outline: none;
          transition: border-color 0.2s, background-color 0.2s;
        }

        .pao-input:focus,
        .pao-select:focus {
          border-color: var(--indigo);
          background: #fff;
        }

        .pao-campaign-preview {
          background: var(--blue-soft);
          border-radius: 8px;
          padding: 12px;
          font-size: 11.5px;
          border-left: 3px solid var(--indigo);
        }

        .pao-campaign-preview strong {
          color: var(--indigo);
          display: block;
          margin-bottom: 4px;
        }

        .pao-campaign-preview p {
          margin: 0;
          color: var(--ink);
          line-height: 1.4;
        }

        .pao-btn-primary {
          background: linear-gradient(135deg, var(--indigo) 0%, #153257 100%);
          color: #fff;
          font-weight: 700;
          padding: 12px;
          border-radius: 8px;
          font-size: 14px;
          cursor: pointer;
          transition: transform 0.1s, opacity 0.2s;
          border: 0;
          box-shadow: 0 4px 10px rgba(31,65,109,0.15);
        }

        .pao-btn-primary:active {
          transform: scale(0.98);
        }

        .pao-btn-primary:disabled {
          opacity: 0.6;
          cursor: not-allowed;
        }

        /* Status Visual Engine */
        .pao-status-engine {
          overflow: hidden;
        }

        /* Celeb Stamp */
        .pao-stamp {
          position: absolute;
          top: 14px;
          right: 14px;
          border: 3px double;
          border-radius: 8px;
          padding: 6px 12px;
          font-size: 13px;
          font-weight: 800;
          text-align: center;
          transform: rotate(12deg);
          text-shadow: 0 2px 4px rgba(0,0,0,0.05);
          letter-spacing: 0.5px;
          line-height: 1.1;
          pointer-events: none;
        }
        
        .pao-stamp span {
          font-size: 9px;
          font-weight: 600;
          display: block;
          margin-top: 1px;
        }

        .pao-stamp.approved {
          border-color: var(--jade);
          color: var(--jade);
          background: rgba(31,122,104,0.06);
          box-shadow: 0 0 0 2px rgba(31,122,104,0.04);
        }

        .pao-stamp.rejected {
          border-color: var(--crimson);
          color: var(--crimson);
          background: rgba(158,47,60,0.06);
          box-shadow: 0 0 0 2px rgba(158,47,60,0.04);
        }

        .pao-stamp.pending {
          border-color: var(--gold);
          color: var(--gold);
          background: rgba(185,139,47,0.06);
          box-shadow: 0 0 0 2px rgba(185,139,47,0.04);
        }

        .pao-status-header {
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          margin-bottom: 14px;
        }

        .pao-lbl-sec {
          font-size: 10px;
          color: var(--muted);
          text-transform: uppercase;
        }

        .pao-status-text {
          font-size: 15px;
          font-weight: 750;
          margin-top: 2px;
        }

        .pao-status-text.status-approved { color: var(--jade); }
        .pao-status-text.status-rejected { color: var(--crimson); }
        .pao-status-text.status-pending { color: var(--gold); }
        .pao-status-text.status-processing { color: var(--indigo); }

        .pao-badge-id {
          font-size: 9.5px;
          background: var(--paper);
          border: 1px solid rgba(135,117,78,0.15);
          color: var(--muted);
          padding: 3px 6px;
          border-radius: 4px;
        }

        .pao-progress-bar {
          height: 6px;
          background: #eee;
          border-radius: 3px;
          overflow: hidden;
          margin-bottom: 18px;
        }

        .pao-progress-fill {
          height: 100%;
          background: linear-gradient(90deg, var(--indigo) 0%, var(--gold) 100%);
          transition: width 0.4s ease;
        }

        /* Check rows */
        .pao-checks-grid {
          display: flex;
          flex-direction: column;
          gap: 10px;
          margin-bottom: 16px;
        }

        .pao-check-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 10px 12px;
          background: rgba(247,245,239,0.35);
          border-radius: 8px;
          border: 1px solid rgba(135,117,78,0.06);
        }

        .pao-check-info {
          display: flex;
          flex-direction: column;
          gap: 2px;
        }

        .pao-check-title {
          font-size: 11px;
          font-weight: 700;
          color: var(--ink);
        }

        .pao-check-desc {
          font-size: 9px;
          color: var(--muted);
        }

        .pao-check-status {
          font-size: 10px;
          font-weight: 700;
          padding: 3px 8px;
          border-radius: 99px;
        }

        .pao-check-status.status-Passed {
          background: rgba(31,122,104,0.12);
          color: var(--jade);
        }

        .pao-check-status.status-Failed {
          background: rgba(158,47,60,0.12);
          color: var(--crimson);
        }

        .pao-check-status.status-Checking {
          background: rgba(185,139,47,0.12);
          color: var(--gold);
          animation: pulse 1.2s infinite;
        }

        .pao-check-status.status-Waiting {
          background: #eee;
          color: #666;
        }

        /* Reasons */
        .pao-reasons-block {
          background: rgba(158,47,60,0.04);
          border: 1px solid rgba(158,47,60,0.12);
          border-radius: 8px;
          padding: 12px;
          font-size: 11.5px;
          margin-bottom: 14px;
        }

        .pao-reasons-block strong {
          color: var(--crimson);
          display: block;
          margin-bottom: 6px;
        }

        .pao-reason-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 4px;
          color: var(--ink);
        }

        .pao-reason-row:last-child {
          margin-bottom: 0;
        }

        .badge-rule {
          font-size: 9px;
          background: var(--crimson);
          color: #fff;
          padding: 1px 4px;
          border-radius: 3px;
        }

        /* History pane */
        .pao-history-list {
          display: flex;
          flex-direction: column;
          gap: 10px;
        }

        .pao-history-row {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 12px 14px;
          background: rgba(247,245,239,0.45);
          border-radius: 10px;
          border: 1px solid rgba(135,117,78,0.1);
        }

        .pao-row-details {
          display: flex;
          flex-direction: column;
          gap: 3px;
        }

        .pao-row-details strong {
          font-size: 12.5px;
          color: var(--ink);
        }

        .pao-row-details span {
          font-size: 9.5px;
          color: var(--muted);
        }

        .pao-row-time {
          font-size: 9px !important;
          color: var(--gold) !important;
        }

        .pao-row-badge {
          font-size: 10px;
          font-weight: 700;
          padding: 4px 8px;
          border-radius: 6px;
        }

        .pao-row-badge.approved { background: rgba(31,122,104,0.12); color: var(--jade); }
        .pao-row-badge.rejected { background: rgba(158,47,60,0.12); color: var(--crimson); }
        .pao-row-badge.pending { background: rgba(185,139,47,0.12); color: var(--gold); }
        .pao-row-badge.processing { background: var(--blue-soft); color: var(--indigo); }

        /* Timeline */
        .pao-timeline {
          border-top: 1px solid rgba(23,32,51,0.06);
          padding-top: 12px;
          margin-top: 14px;
        }

        .pao-timeline h4 {
          margin: 0 0 10px 0;
          font-size: 11px;
          font-weight: 750;
          color: var(--muted);
          text-transform: uppercase;
        }

        .pao-timeline-event {
          display: flex;
          gap: 10px;
          margin-bottom: 8px;
          position: relative;
        }

        .pao-timeline-event:last-child {
          margin-bottom: 0;
        }

        .pao-marker {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          background: var(--indigo);
          margin-top: 5px;
          flex: 0 0 auto;
        }

        .pao-event-detail {
          display: flex;
          flex-direction: column;
        }

        .pao-event-detail p {
          margin: 0;
          font-size: 10.5px;
          font-weight: 600;
          color: var(--ink);
        }

        .pao-event-detail span {
          font-size: 8.5px;
          color: var(--muted);
        }

        /* Toast & Info */
        .pao-toast {
          position: fixed;
          top: 24px;
          left: 50%;
          transform: translateX(-50%);
          background: var(--ink);
          color: #fff;
          padding: 10px 20px;
          border-radius: 99px;
          font-size: 12.5px;
          font-weight: 600;
          box-shadow: 0 8px 30px rgba(0,0,0,0.2);
          z-index: 999;
          animation: slideDown 0.3s forwards, slideUp 0.3s 2s forwards;
        }

        .pao-info-list {
          display: flex;
          flex-direction: column;
          gap: 12px;
        }

        .pao-info-row {
          border-bottom: 1px solid rgba(23,32,51,0.05);
          padding-bottom: 10px;
        }

        .pao-info-row:last-child {
          border-bottom: 0;
          padding-bottom: 0;
        }

        .pao-info-row strong {
          font-size: 12px;
          color: var(--indigo);
          display: block;
          margin-bottom: 4px;
        }

        .pao-info-row p {
          margin: 0;
          font-size: 11px;
          color: var(--muted);
          line-height: 1.4;
        }

        /* Animations */
        @keyframes pulse {
          0% { opacity: 0.6; }
          50% { opacity: 1; }
          100% { opacity: 0.6; }
        }

        @keyframes slideDown {
          from { top: -40px; opacity: 0; }
          to { top: 24px; opacity: 1; }
        }
      `}</style>
    </div>
  );
}
