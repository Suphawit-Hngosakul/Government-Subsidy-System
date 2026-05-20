# Officer Service Contract

Owner: ศุภวิชญ์ หงอสกุล 6609650053

Status: Version 1, in-memory mock claims + HTTP adapter to Orchestrator

## Purpose

Officer Service เป็นส่วนของเจ้าหน้าที่ที่ใช้ตรวจสอบคำร้อง (claim) ที่อยู่ในสถานะ pending — ดูรายการคำร้องค้าง, เปิดดูรายละเอียดพร้อมผลการตรวจสอบจาก DOPA/SSO/KTB และตัดสินใจอนุมัติ/ปฏิเสธด้วยตนเอง

ใน Version 1 ระบบใช้ in-memory repository พร้อมข้อมูล mock 3 คำร้อง (`claim-001`, `claim-002`, `claim-003`) สำหรับให้ทีมอื่นทดสอบได้ทันที และเรียกข้อมูล eligibility จาก Orchestrator Service ของคนที่ 4 ผ่าน HTTP adapter

## Ownership Scope

| Component | Responsibility |
|-----------|----------------|
| Officer Claim Handler | เปิด HTTP endpoints สำหรับเจ้าหน้าที่ |
| Officer Service | ตรวจสอบความถูกต้อง, จัดการ state machine ของ claim (pending → approved/rejected) |
| Officer Claim Repository | จัดเก็บ claim ในรูปแบบที่ officer มองเห็น (in-memory v1 พร้อม seed) |
| HTTP Orchestrator Adapter | เรียก `POST /internal/decision` ของ Orchestrator เพื่อดึงผล eligibility |

## Architectural Note: HTTP Adapter

Officer Service พึ่งพา **interface `OrchestratorClient`** ที่ประกาศในแพ็กเกจ `service` — ไม่ผูกกับ implementation ใด ในรันไทม์ `main.go` ฉีด `adapter.HTTPOrchestratorAdapter` ที่ห่อหุ้ม HTTP call เข้าไป

ผลที่ได้:

- Officer Service เปลี่ยน implementation ได้โดยไม่ต้องแก้ตัวเอง (เช่น เปลี่ยนเป็น in-process call, gRPC client, หรือ message bus)
- Unit test ของ Officer Service ใช้ stub `OrchestratorClient` (ดู `service/officer_service_test.go`) — ไม่ต้องสตาร์ท HTTP server
- Adapter test ใช้ `httptest.Server` จำลอง orchestrator (ดู `adapter/orchestrator_adapter_test.go`) — ทดสอบ encoding/decoding และ status-code handling ได้โดยไม่ต้องรัน orchestrator จริง

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไข endpoint, HTTP method, request body, response body, status code, state machine ของ claim หรือ shape ของ `EligibilityResult` ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ

ถ้า adapter เปลี่ยน endpoint ของ Orchestrator ที่เรียก ต้องประสานงานกับเจ้าของ orchestrator-contract.md (คนที่ 4) ก่อน

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Officer Frontend | `GET /api/v1/officer/claims` | แสดงรายการคำร้องที่รออนุมัติ |
| Officer Frontend | `GET /api/v1/officer/claim/:id` | เปิดรายละเอียดคำร้องเพื่อตรวจสอบ |
| Officer Frontend | `PATCH /api/v1/officer/claim/:id/approve|reject` | บันทึกการตัดสิน |

## Downstream Dependencies

| Service | Endpoint ที่ Officer เรียก | เหตุผล |
|---------|----------------------------|--------|
| Orchestrator (คนที่ 4) | `POST /internal/decision` | ดึงผล eligibility จาก DOPA/SSO/KTB |

## Configuration

| Environment Variable | Default | Description |
|----------------------|---------|-------------|
| `ORCHESTRATOR_BASE_URL` | `http://localhost:8080` | Base URL ของ Orchestrator Service |

## Base URL

Local development:

```text
http://localhost:8080
```

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/officer/claims` | รายการคำร้องสถานะ `pending` ทั้งหมด |
| `GET` | `/api/v1/officer/claim/{id}` | รายละเอียดคำร้อง + ผล 3 sources |
| `PATCH` | `/api/v1/officer/claim/{id}/approve` | เจ้าหน้าที่อนุมัติ |
| `PATCH` | `/api/v1/officer/claim/{id}/reject` | เจ้าหน้าที่ปฏิเสธ (ต้องระบุ reason) |

## Data Model

### OfficerClaim

| Field | Type | Description |
|-------|------|-------------|
| `claimId` | string | ID คำร้อง |
| `nationalId` | string | เลขบัตรประชาชน 13 หลัก |
| `projectId` | string | ID โครงการที่ยื่นขอ |
| `status` | `pending` \| `approved` \| `rejected` | สถานะปัจจุบัน |
| `submittedAt` | RFC3339 datetime | เวลาที่ประชาชนยื่น |
| `officerDecision` | OfficerDecision \| null | ผลการตัดสินของเจ้าหน้าที่ ถ้ามี |

### OfficerDecision

| Field | Type | Description |
|-------|------|-------------|
| `officerId` | string | ID ของเจ้าหน้าที่ที่ตัดสิน |
| `reason` | string | เหตุผลประกอบการตัดสิน (required สำหรับ reject) |
| `decidedAt` | RFC3339 datetime | เวลาที่ตัดสิน |

### EligibilityResult

ดึงมาจาก Orchestrator (ผ่าน HTTP adapter) ถ้า Orchestrator ยังไม่มีผลของ claim นี้ ระบบจะคืนค่า zero ทั้งหมด (`status` ว่าง, `claimId` ว่าง)

| Field | Type | Description |
|-------|------|-------------|
| `claimId` | string | ID คำร้องที่ Orchestrator ตอบกลับ |
| `status` | string | ผลจาก Decision Engine: `approved` / `rejected` / `pending` / `processing` |
| `reasons` | []string | เหตุผลของผลข้างต้น |
| `sources.dopa` | EligibilityDOPA | ผลตรวจ DOPA: `valid`, `age`, `alive`, `cardActive` |
| `sources.sso` | EligibilitySSO | ผลตรวจ SSO: `section`, `contributionMonths` |
| `sources.ktb` | EligibilityKTB | ผลตรวจ KTB: `depositTotal`, `averageMonthlyIncome`, `promptPayLinked` |

## GET /api/v1/officer/claims

ดึงรายการคำร้องที่ยังไม่ได้ตัดสิน (status = `pending`) เรียงตาม `submittedAt` เก่าก่อน

Response `200 OK`:

```json
{
  "claims": [
    {
      "claimId": "claim-001",
      "nationalId": "1101700203451",
      "projectId": "proj-energy",
      "status": "pending",
      "submittedAt": "2026-05-13T10:59:27Z"
    }
  ]
}
```

## GET /api/v1/officer/claim/{id}

ดึงรายละเอียดคำร้อง + ผล eligibility (ถ้ามี) จาก Orchestrator

Response `200 OK`:

```json
{
  "claim": {
    "claimId": "claim-001",
    "nationalId": "1101700203451",
    "projectId": "proj-energy",
    "status": "pending",
    "submittedAt": "2026-05-13T10:59:27Z"
  },
  "eligibility": {
    "claimId": "claim-001",
    "status": "approved",
    "reasons": ["eligible by mock subsidy rules"],
    "sources": {
      "dopa": {"valid": true, "age": 35, "alive": true, "cardActive": true},
      "sso": {"section": "40", "contributionMonths": 12},
      "ktb": {"depositTotal": 12000, "averageMonthlyIncome": 15000, "promptPayLinked": true}
    }
  }
}
```

ฟิลด์ `eligibility` เป็น **optional** — จะไม่ปรากฏใน response (หรือเป็น `null`) ในกรณีต่อไปนี้:

- Orchestrator ยังไม่ orchestrate claim นี้ (ตอบ 404 มา)
- Orchestrator ติดต่อไม่ได้ / timeout / ตอบ 5xx — Officer Service จะ log error และคืน claim โดยไม่มี eligibility เจ้าหน้าที่ยังสามารถดูข้อมูล claim และตัดสินใจด้วยตนเองได้

วิธีเช็คจากฝั่ง client: ถ้า key `eligibility` หายไปหรือเป็น `null` แสดงว่ายังไม่มีผลการตรวจสอบจาก Orchestrator

Error responses:
- `400 Bad Request` — `claimId` ใน path ว่าง
- `404 Not Found` — ไม่พบ claim ใน officer repository

## PATCH /api/v1/officer/claim/{id}/approve

เจ้าหน้าที่อนุมัติคำร้อง

Request body:

```json
{
  "officerId": "officer-1",
  "reason": "verified manually"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `officerId` | Yes | ID เจ้าหน้าที่ |
| `reason` | No | เหตุผล (optional สำหรับ approve) |

Response `200 OK`: คืน `OfficerClaim` ที่อัปเดตแล้ว status = `approved` พร้อม `officerDecision`

Error responses:
- `400 Bad Request` — `officerId` ว่าง หรือ JSON ผิด
- `404 Not Found` — ไม่พบ claim
- `409 Conflict` — claim ถูกตัดสินไปแล้ว

## PATCH /api/v1/officer/claim/{id}/reject

เจ้าหน้าที่ปฏิเสธคำร้อง

Request body:

```json
{
  "officerId": "officer-2",
  "reason": "income exceeds threshold"
}
```

| Field | Required | Description |
|-------|----------|-------------|
| `officerId` | Yes | ID เจ้าหน้าที่ |
| `reason` | **Yes** | เหตุผลที่ปฏิเสธ |

Response `200 OK`: คืน `OfficerClaim` ที่อัปเดตแล้ว status = `rejected` พร้อม `officerDecision`

Error responses:
- `400 Bad Request` — `officerId` หรือ `reason` ว่าง
- `404 Not Found` — ไม่พบ claim
- `409 Conflict` — claim ถูกตัดสินไปแล้ว

## State Machine

```text
pending ──approve──► approved   (terminal)
pending ──reject───► rejected   (terminal)
```

`approved` และ `rejected` เป็น terminal state — เปลี่ยนกลับไม่ได้ (จะได้ 409)

## Mock Seed Data

Repository ใน v1 ถูก seed ด้วย 3 claim เริ่มต้นเมื่อ server เริ่มทำงาน:

| ClaimID | NationalID | ProjectID |
|---------|------------|-----------|
| `claim-001` | `1101700203451` | `proj-energy` |
| `claim-002` | `1101700203452` | `proj-energy` |
| `claim-003` | `1101700203453` | `proj-water` |

ทั้ง 3 ทำสถานะ `pending` หลัง restart server seed จะถูกสร้างใหม่ทุกครั้ง

## Run Locally

จากโฟลเดอร์ `Backend`:

```bash
go run .
```

ถ้า Orchestrator รันคนละพอร์ต:

```bash
ORCHESTRATOR_BASE_URL=http://localhost:9090 go run .
```

ตัวอย่าง request:

```bash
curl http://localhost:8080/api/v1/officer/claims

curl -X PATCH http://localhost:8080/api/v1/officer/claim/claim-001/approve \
  -H 'Content-Type: application/json' \
  -d '{"officerId":"officer-1","reason":"verified manually"}'
```

## Test

จากโฟลเดอร์ `Backend`:

```bash
go test ./...
```

- Service test ใช้ stub `OrchestratorClient` ไม่ต้องมี orchestrator จริง
- Adapter test ใช้ `httptest.Server` จำลอง orchestrator

## Roadmap

- Auth middleware ตรวจ role `officer` (รอ AUTH ของคนที่ 1)
- เปลี่ยน in-memory เป็น PostgreSQL repository
- รับ claim ใหม่จาก Benefit Service (คนที่ 2) — ปัจจุบันใช้ seed
- เก็บ audit log ทุกครั้งที่มีการตัดสิน (ส่งต่อให้ Dashboard module)
