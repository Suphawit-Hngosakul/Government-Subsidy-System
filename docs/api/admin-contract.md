# Admin Service Contract

Owner: ศุภวิชญ์ หงอสกุล 6609650053

Status: Version 1, in-memory mock implementation

## Purpose

Admin Service เป็นส่วนของระบบที่ผู้ดูแลระบบใช้จัดการ "โครงการสวัสดิการ" (Project) เช่น สร้างโครงการใหม่ ตั้งเกณฑ์สิทธิ์ (criteria) แก้ไข หรือลบโครงการ โครงการเหล่านี้คือสิ่งที่ประชาชนจะยื่นคำร้องเพื่อขอรับสิทธิ์ และเป็นข้อมูลอ้างอิงให้ Orchestrator/Officer ใช้ตรวจคุณสมบัติ

ใน Version 1 ระบบใช้ in-memory repository เพื่อให้ทีมอื่นทดสอบ API contract ได้ก่อน โดยยังไม่ต้องรอ database, authentication และ project-aware decision engine

## Ownership Scope

| Component | Responsibility |
|-----------|----------------|
| Admin Project Handler | เปิด HTTP endpoints สำหรับจัดการโครงการ |
| Project Service | ตรวจสอบความถูกต้อง, สร้าง ID, ใส่ค่า default และเรียก repository |
| Project Repository | จัดเก็บโครงการใน in-memory map (v1) — interface พร้อมเปลี่ยนเป็น PostgreSQL ภายหลัง |
| Project Domain | นิยาม `Project` และ `ProjectCriteria` |

ใน Version 1 ผู้ใช้ทุกคนเข้าถึง endpoint เหล่านี้ได้โดยไม่ต้อง authenticate รอ AUTH service (คนที่ 1) เสร็จแล้วจึงเพิ่ม middleware ตรวจสิทธิ์บทบาท admin

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไข endpoint, HTTP method, request body, response body, status code หรือ structure ของ `Project`/`ProjectCriteria` ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ

ถ้า endpoint ใหม่ถูกเพิ่มในส่วนของศุภวิชญ์ ให้เพิ่มรายการในเอกสารนี้ก่อน merge เข้า `develop`

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Admin Frontend | `POST/PUT/DELETE /api/v1/admin/project[/:id]` | จัดการโครงการ |
| Admin Frontend | `GET /api/v1/admin/projects` | แสดงรายการโครงการทั้งหมด |
| Citizen Benefit Service | `GET /api/v1/admin/projects` หรือ `GET /api/v1/admin/project/:id` | ดึงรายการโครงการที่เปิดรับ + ตรวจ criteria ก่อนยื่นคำร้อง |
| Orchestrator / Officer | `GET /api/v1/admin/project/:id` | ใช้ criteria เป็นข้อมูลอ้างอิงเวลาประเมินคำร้อง |

## Base URL

Local development:

```text
http://localhost:8080
```

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/projects` | รายการโครงการทั้งหมด เรียงตาม createdAt |
| `POST` | `/api/v1/admin/project` | สร้างโครงการใหม่พร้อมเกณฑ์สิทธิ์ |
| `GET` | `/api/v1/admin/project/{id}` | ดึงรายละเอียดโครงการตาม id |
| `PUT` | `/api/v1/admin/project/{id}` | แก้ไขโครงการ (replace ทั้งก้อน) |
| `DELETE` | `/api/v1/admin/project/{id}` | ลบโครงการ |

## Data Model

### Project

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | ID โครงการ ระบบสร้างให้เองในรูป `proj-<hex>` (client ส่งมาไม่ได้ จะถูก server เพิกเฉย) |
| `name` | string | ชื่อโครงการ (required) |
| `description` | string | คำอธิบายโครงการ |
| `active` | bool | true ถ้าโครงการเปิดรับคำร้อง |
| `criteria` | ProjectCriteria | เกณฑ์สิทธิ์ของโครงการ |
| `createdAt` | RFC3339 datetime | เวลาที่สร้าง (ระบบเซ็ตอัตโนมัติ) |
| `updatedAt` | RFC3339 datetime | เวลาที่แก้ไขล่าสุด (ระบบเซ็ตอัตโนมัติ) |

### ProjectCriteria

| Field | Type | Description |
|-------|------|-------------|
| `minAge` | int | อายุขั้นต่ำ ถ้าไม่ส่งหรือเป็น 0 ระบบจะใช้ค่า default 18 |
| `maxAge` | int | อายุสูงสุด 0 หมายถึงไม่จำกัด |
| `maxMonthlyIncome` | float | รายได้เฉลี่ยต่อเดือนสูงสุด 0 หมายถึงไม่จำกัด |
| `allowedSsoSections` | []string | section ของ SSO ที่อนุญาต (เช่น `["39","40"]`) ถ้าเว้นว่างหมายถึงไม่กรอง |
| `requirePromptPay` | bool | true ถ้าต้องมีบัญชีพร้อมเพย์ผูกอยู่ |

หมายเหตุ: ใน Version 1 ค่า criteria จะถูกจัดเก็บอย่างเดียว ยังไม่ถูกใช้ใน Decision Engine ของ Orchestrator (ใช้ rule แบบ hardcode ใน v1) จะรวมเข้าด้วยกันใน v2

## GET /api/v1/admin/projects

ดึงรายการโครงการทั้งหมด

Response `200 OK`:

```json
{
  "projects": [
    {
      "id": "proj-1b9f8a2d8c8a963a",
      "name": "Energy Subsidy",
      "description": "Help low-income citizens",
      "active": true,
      "criteria": {
        "minAge": 18,
        "maxMonthlyIncome": 15000,
        "requirePromptPay": true
      },
      "createdAt": "2026-05-13T12:33:33.580Z",
      "updatedAt": "2026-05-13T12:33:33.580Z"
    }
  ]
}
```

## POST /api/v1/admin/project

สร้างโครงการใหม่

Request body:

```json
{
  "name": "Energy Subsidy",
  "description": "Help low-income citizens",
  "active": true,
  "criteria": {
    "maxMonthlyIncome": 15000,
    "requirePromptPay": true
  }
}
```

Field: `name` required ที่เหลือ optional ถ้า client ส่ง `id` มาด้วย ระบบจะเพิกเฉยและสร้างให้เอง

Response `201 Created`: คืน `Project` ที่สร้างพร้อม `id`, `createdAt`, `updatedAt`

Error response `400 Bad Request`:

```json
{ "error": "project name is required" }
```

## GET /api/v1/admin/project/{id}

ดึงโครงการตาม id

Response `200 OK`: คืน `Project`

Error response `404 Not Found`:

```json
{ "error": "project not found" }
```

## PUT /api/v1/admin/project/{id}

แก้ไขโครงการแบบ **partial update** ส่งมาเฉพาะฟิลด์ที่อยากเปลี่ยน ฟิลด์ที่ไม่ส่ง (หรือเป็น `null`) ระบบจะไม่แตะของเดิม

ฟิลด์ที่อนุญาตให้แก้ได้: `name`, `description`, `active`, `criteria`

ถ้าส่ง `criteria` มาด้วย ระบบจะ **แทนที่ทั้งก้อน** ของ criteria (ไม่ merge ลึกถึงฟิลด์ย่อย)

ตัวอย่าง — แก้แค่ชื่อ:

```json
{ "name": "Energy Subsidy v2" }
```

ตัวอย่าง — ปิดโครงการ:

```json
{ "active": false }
```

ตัวอย่าง — เปลี่ยน criteria ทั้งชุด:

```json
{
  "criteria": {
    "minAge": 20,
    "maxMonthlyIncome": 20000,
    "requirePromptPay": true
  }
}
```

Response `200 OK`: คืน `Project` ที่แก้แล้ว (รวมฟิลด์เดิมที่ไม่ได้ถูกแก้)

Error responses:
- `400 Bad Request` ถ้าส่ง `name` มาเป็นสตริงว่าง หรือ JSON ผิดรูปแบบ
- `404 Not Found` ถ้าไม่พบ id

## DELETE /api/v1/admin/project/{id}

ลบโครงการ

Response `204 No Content`

Error response `404 Not Found`:

```json
{ "error": "project not found" }
```

## Run Locally

จากโฟลเดอร์ `Backend`:

```bash
go run .
```

ตัวอย่าง request:

```bash
curl -X POST http://localhost:8080/api/v1/admin/project \
  -H 'Content-Type: application/json' \
  -d '{"name":"Energy Subsidy","criteria":{"maxMonthlyIncome":15000,"requirePromptPay":true}}'
```

## Test

จากโฟลเดอร์ `Backend`:

```bash
go test ./...
```

## Roadmap (สิ่งที่จะเพิ่มในเวอร์ชันถัดไป)

- Auth middleware ตรวจ role admin (รอ AUTH ของคนที่ 1)
- เปลี่ยน in-memory เป็น PostgreSQL repository (interface เดิม)
- Officer endpoints (`/api/v1/officer/*`) — ตรวจสอบคำร้องที่อ้างถึง project นี้
- Dashboard endpoints (`/api/v1/admin/stats`, `/api/v1/admin/audit-log`)
- ส่ง criteria ของ project ให้ Orchestrator ใช้ใน Decision Engine แทน hardcoded rules
