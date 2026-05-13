# Dashboard Service Contract

Owner: ศุภวิชญ์ หงอสกุล 6609650053

Status: Version 1, read aggregator over project + officer claim + audit repositories

## Purpose

Dashboard Service เป็นส่วนของผู้ดูแลระบบที่ใช้ดูภาพรวมของระบบ — ยอดรวมโครงการและคำร้องตามสถานะ และประวัติการตัดสินของเจ้าหน้าที่ (audit log) เพื่อตรวจสอบย้อนหลังว่าใครทำอะไรเมื่อไหร่

ใน Version 1 ระบบใช้ข้อมูลจาก in-memory repository ทั้ง 3 ก้อนของ backend ตัวเดียวกัน — Project, Officer Claim, Audit — โดย Dashboard ไม่ได้เก็บข้อมูลใดเอง เป็น "read aggregator" บริสุทธิ์

## Ownership Scope

| Component | Responsibility |
|-----------|----------------|
| Admin Dashboard Handler | เปิด HTTP endpoints `/api/v1/admin/stats` และ `/api/v1/admin/audit-log` |
| Dashboard Service | รวบรวมข้อมูลจาก 3 repositories ตามที่ frontend ต้องใช้ |
| Audit Repository | จัดเก็บ audit entries (in-memory v1) — implements ทั้ง `AuditLogger` (write) และ `AuditReader` (read) |

## Architectural Note: Audit Logging (Write Side)

`AuditLogger` interface ถูกใช้โดย `OfficerService` เมื่อ approve/reject สำเร็จ ก็ append entry หนึ่งรายการ — ผ่าน DI ใน `main.go`

แยกเป็น 2 interfaces ตาม Interface Segregation Principle:

| Interface | Consumer | Methods |
|-----------|----------|---------|
| `AuditLogger` | Officer Service (write) | `Append(ctx, entry)` |
| `AuditReader` | Dashboard Service (read) | `List(ctx, limit)` |

ทั้งสอง interface implement โดย `MemoryAuditRepository` ตัวเดียวกัน main.go inject ก้อนเดียวเข้าทั้ง 2 consumer

ใน Version 1 audit log ครอบคลุมเฉพาะ **claim decisions** (approved/rejected) ของเจ้าหน้าที่เท่านั้น Project CRUD ยังไม่ถูก audit เพราะยังไม่มี admin auth identifier (รอคนที่ 1)

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไข endpoint, response shape, รายการ action ที่ audit หรือ rule ของ limit ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Admin Frontend | `GET /api/v1/admin/stats` | แสดงกราฟสรุปบน dashboard |
| Admin Frontend | `GET /api/v1/admin/audit-log` | แสดงตารางประวัติการอนุมัติ/ปฏิเสธ |

## Base URL

Local development:

```text
http://localhost:8080
```

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/stats` | ภาพรวม projects + claim counts |
| `GET` | `/api/v1/admin/audit-log?limit=N` | รายการ audit entries เรียงใหม่สุดก่อน |

## Data Model

### Stats

| Field | Type | Description |
|-------|------|-------------|
| `projects.total` | int | จำนวนโครงการทั้งหมด |
| `projects.active` | int | จำนวนโครงการที่ `active=true` |
| `claims.total` | int | จำนวน claim ทั้งหมดที่ officer มองเห็น |
| `claims.pending` | int | claim ที่ยังไม่ตัดสิน |
| `claims.approved` | int | claim ที่อนุมัติแล้ว |
| `claims.rejected` | int | claim ที่ปฏิเสธแล้ว |

### AuditEntry

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | ID ของ entry (`audit-<hex>`) ระบบสร้างให้ |
| `at` | RFC3339 datetime | เวลาที่ action เกิดขึ้น (UTC) |
| `actor` | string | ผู้กระทำ ใน v1 คือ `officerId` |
| `action` | `claim.approved` \| `claim.rejected` | ประเภท action |
| `entityId` | string | ID ของ entity ที่ถูกกระทำ (เช่น claimId) |
| `metadata` | object | ข้อมูลเสริม (เช่น `{"reason": "..."}`) — `omitempty` |

## GET /api/v1/admin/stats

ดึงภาพรวมสถิติของระบบ

Response `200 OK`:

```json
{
  "projects": {
    "total": 2,
    "active": 1
  },
  "claims": {
    "total": 3,
    "pending": 1,
    "approved": 1,
    "rejected": 1
  }
}
```

## GET /api/v1/admin/audit-log

ดึงรายการ audit entries เรียงตามเวลาใหม่สุดก่อน

Query parameter:

| Name | Type | Default | Max | Description |
|------|------|---------|-----|-------------|
| `limit` | int | 50 | 200 | จำนวน entry สูงสุดที่ต้องการ |

ถ้า `limit <= 0` จะใช้ค่า default หาก `limit > 200` จะถูก cap ไว้ที่ 200

Response `200 OK`:

```json
{
  "entries": [
    {
      "id": "audit-507c2191747b",
      "at": "2026-05-13T15:09:28Z",
      "actor": "officer-2",
      "action": "claim.rejected",
      "entityId": "claim-002",
      "metadata": {"reason": "income high"}
    },
    {
      "id": "audit-38a2ccca4296",
      "at": "2026-05-13T15:09:28Z",
      "actor": "officer-1",
      "action": "claim.approved",
      "entityId": "claim-001",
      "metadata": {"reason": "verified"}
    }
  ]
}
```

ถ้ายังไม่มี entry เลย จะคืน `{"entries": []}`

## Run Locally

จากโฟลเดอร์ `Backend`:

```bash
go run .
```

ตัวอย่าง request:

```bash
curl http://localhost:8080/api/v1/admin/stats
curl "http://localhost:8080/api/v1/admin/audit-log?limit=20"
```

## Test

จากโฟลเดอร์ `Backend`:

```bash
go test ./...
```

- Dashboard service test ใช้ in-memory repositories ทั้ง 3 ก้อนโดยตรง — ไม่มี dependency external
- Officer audit assertions อยู่ใน `service/officer_service_test.go` (`TestOfficerServiceApproveAppendsAuditEntry`, `TestOfficerServiceRejectAppendsAuditEntry`)

## Roadmap

- Auth middleware ตรวจ role `admin` (รอ AUTH ของคนที่ 1)
- Audit Project CRUD เมื่อมี admin auth identifier
- เปลี่ยน in-memory เป็น PostgreSQL repository (interface เดิม)
- เพิ่ม filter ใน `/audit-log` (เช่น by actor, by action, by entityId, date range)
- เพิ่ม `claims-per-project` breakdown ใน `/stats`
