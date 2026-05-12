# Orchestrator Service Contract

Owner: ธนธัช นิติเจริญ 6609650095

Status: Version 1, mock implementation

## Purpose

Orchestrator Service เป็น internal web service สำหรับประสานงานตรวจสอบสิทธิ์ของ claim หนึ่งรายการ โดยรับ `claimId` และ `nationalId` จาก Claim/Benefit service จากนั้นเรียกข้อมูลจำลองของ DOPA, SSO และ KTB แบบ parallel แล้วนำผลไปประเมิน eligibility rules เพื่อคืนสถานะ `approved`, `rejected` หรือ `pending`

ใน Version 1 ระบบยังใช้ mock data และ in-memory repository เพื่อให้ทีมอื่นทดสอบ API contract ได้ก่อน โดยยังไม่ต้องรอ database, Redis หรือ external mock provider ของทีมอื่นเสร็จสมบูรณ์

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไข endpoint, HTTP method, request body, response body, status code, SSE event format หรือ decision rule ของ Orchestrator Service ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ

ถ้า endpoint ใหม่ถูกเพิ่มในส่วนของธนธัช ให้เพิ่มรายการในเอกสารนี้ก่อน merge เข้า `develop`

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Benefit / Claim Service | `POST /internal/orchestrate` | Trigger การตรวจสอบสิทธิ์หลังประชาชนยื่น claim |
| Benefit / Claim Service หรือ Officer Service | `POST /internal/decision` | ดึงผลการตัดสินล่าสุดของ claim |
| Frontend | `GET /api/v1/claim/:claimId/stream` | รับสถานะ claim แบบ real-time ผ่าน SSE |
| Developer / Health Check | `GET /healthz` | ตรวจว่า service ยังทำงานอยู่ |

## Base URL

Local development:

```text
http://localhost:8080
```

## Endpoint Summary

| Method | Endpoint | Access | Description |
|--------|----------|--------|-------------|
| `GET` | `/healthz` | Public dev | ตรวจ health ของ service |
| `POST` | `/internal/orchestrate` | Internal service | เรียก mock DOPA, SSO, KTB แบบ parallel และบันทึกผล |
| `POST` | `/internal/decision` | Internal service | อ่านผล decision ล่าสุดจาก in-memory store |
| `GET` | `/api/v1/claim/:claimId/stream` | Frontend | เปิด SSE stream สำหรับ claim status events |

## GET /healthz

ใช้สำหรับตรวจว่า service พร้อมรับ request

Response `200 OK`:

```json
{
  "status": "ok"
}
```

## POST /internal/orchestrate

ใช้โดย Benefit / Claim Service หลังสร้าง claim แล้ว เพื่อให้ Orchestrator เริ่มตรวจสอบสิทธิ์

Request body:

```json
{
  "claimId": "claim-001",
  "nationalId": "1101700203451",
  "projectId": "project-001"
}
```

Field:

| Field | Required | Description |
|-------|----------|-------------|
| `claimId` | Yes | ID ของ claim ที่ต้องตรวจสอบ |
| `nationalId` | Yes | เลขบัตรประชาชน 13 หลัก ใช้เรียก provider |
| `projectId` | No | ID ของโครงการสวัสดิการ เผื่อใช้กับ rules เฉพาะโครงการใน version ถัดไป |

Current mock behavior:

| Provider | Mock response |
|----------|---------------|
| DOPA | valid citizen, age 35, alive, active card |
| SSO | section 40, contribution 12 months |
| KTB | deposit 12000, monthly income 15000, PromptPay linked |

Response `202 Accepted`:

```json
{
  "claimId": "claim-001",
  "status": "approved",
  "reasons": [
    "eligible by mock subsidy rules"
  ],
  "sources": {
    "dopa": {
      "valid": true,
      "age": 35,
      "alive": true,
      "cardActive": true
    },
    "sso": {
      "section": "40",
      "contributionMonths": 12
    },
    "ktb": {
      "depositTotal": 12000,
      "averageMonthlyIncome": 15000,
      "promptPayLinked": true
    }
  }
}
```

Error response `400 Bad Request`:

```json
{
  "error": "claimId and nationalId are required"
}
```

## POST /internal/decision

ใช้สำหรับอ่านผล decision ล่าสุดของ claim ที่เคยผ่าน `/internal/orchestrate` แล้ว

Request body:

```json
{
  "claimId": "claim-001"
}
```

Response `200 OK`:

```json
{
  "claimId": "claim-001",
  "status": "approved",
  "reasons": [
    "eligible by mock subsidy rules"
  ],
  "sources": {
    "dopa": {
      "valid": true,
      "age": 35,
      "alive": true,
      "cardActive": true
    },
    "sso": {
      "section": "40",
      "contributionMonths": 12
    },
    "ktb": {
      "depositTotal": 12000,
      "averageMonthlyIncome": 15000,
      "promptPayLinked": true
    }
  }
}
```

Error response `404 Not Found`:

```json
{
  "error": "claim result not found"
}
```

## GET /api/v1/claim/:claimId/stream

ใช้โดย frontend เพื่อรับสถานะ claim แบบ real-time ผ่าน Server-Sent Events

Request:

```text
GET /api/v1/claim/claim-001/stream
Accept: text/event-stream
```

Response content type:

```text
text/event-stream
```

Event format:

```text
event: claim-status
data: {"claimId":"claim-001","status":"processing","message":"started external verification","at":"2026-05-12T16:09:39Z"}
```

Current events:

| Status | Message | Trigger |
|--------|---------|---------|
| `processing` | `started external verification` | เมื่อเริ่ม `/internal/orchestrate` |
| `approved` | `claim approved automatically` | เมื่อ decision เป็น approved |
| `rejected` | `claim rejected by eligibility rules` | เมื่อ decision เป็น rejected |
| `pending` | `claim requires follow-up` | เมื่อ decision เป็น pending |

หมายเหตุ: Version 1 ใช้ in-memory event store ถ้า restart service ข้อมูล result และ event history จะหาย

## Decision Rules V1

ระบบเริ่มจากสถานะ `approved` แล้วเปลี่ยนสถานะตาม rules ต่อไปนี้

| Rule | Result |
|------|--------|
| Provider ใด provider หนึ่ง error | `pending` |
| DOPA invalid, citizen not alive หรือ card inactive | `rejected` |
| Age ต่ำกว่า 18 | `rejected` |
| SSO section เป็น `33` | `rejected` |
| PromptPay ยังไม่ผูกบัญชี | `pending` |
| Average monthly income มากกว่า 30000 | `rejected` |
| ไม่เข้า rule ปฏิเสธหรือรอตรวจเพิ่ม | `approved` |

ถ้ามีหลาย rule ตรงกัน ระบบจะสะสมเหตุผลทั้งหมดไว้ใน `reasons`

## Run Locally

จากโฟลเดอร์ `Backend`

```bash
go run .
```

ตัวอย่าง request:

```bash
curl -X POST http://localhost:8080/internal/orchestrate \
  -H 'Content-Type: application/json' \
  -d '{"claimId":"claim-001","nationalId":"1101700203451"}'
```

## Test

จากโฟลเดอร์ `Backend`

```bash
go test ./...
```
