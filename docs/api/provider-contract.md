# Provider Mock Service Contract

Owner: ธนรัชต แก้วลอย 6609650426

Status: Version 1, PostgreSQL-backed local mock implementation

## Purpose

Provider Mock Service เป็น web service จำลองของ 3 หน่วยงานที่ระบบ Government Subsidy System ต้องใช้ในการตรวจสอบสิทธิ์:

- DOPA: ตรวจตัวตน อายุ สถานะบุคคล และสถานะบัตรประชาชน
- SSO: ตรวจสถานะผู้ประกันตน มาตรา 33/39/40 และจำนวนเดือนที่ส่งเงินสมทบ
- KTB: ตรวจยอดเงินฝาก รายได้เฉลี่ย บัญชีธนาคาร และ PromptPay

Version 1 ใช้ PostgreSQL local ผ่าน Docker โดย seed data อยู่ที่ `database/postgres/init`

Schema และข้อมูลเป็น mock สำหรับงานกลุ่มเท่านั้น ไม่ใช่ schema จริงของหน่วยงานรัฐหรือธนาคาร

## Ownership Scope

| Component | Responsibility |
|-----------|----------------|
| Provider Handler | เปิด HTTP endpoints สำหรับ DOPA, SSO และ KTB |
| Provider Service | Validate `nationalId`, mapping response และ implement interface ที่ Orchestrator ใช้ |
| Provider Repository | Query PostgreSQL local จาก schema `dopa`, `sso`, `ktb` |
| Provider Database | Dockerized PostgreSQL พร้อม schema และ seed data |

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไขสิ่งต่อไปนี้ ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ:

- endpoint path
- HTTP method
- request parameter
- response body
- error status code
- database field ที่ response พึ่งพา
- behavior ที่ Orchestrator ใช้ตัดสินสิทธิ์

ถ้าเพิ่ม endpoint ใหม่ในส่วน DOPA/SSO/KTB ให้เพิ่มรายการในเอกสารนี้ก่อน merge เข้า `develop`

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Orchestrator Service | DOPA/SSO/KTB provider methods ผ่าน `ProviderService` | ดึงข้อมูล 3 แหล่งเพื่อประเมินสิทธิ์แบบ parallel |
| Developer / Tester | HTTP endpoints ใต้ `/api/v1/dopa`, `/api/v1/sso`, `/api/v1/ktb` | ทดสอบ mock provider โดยตรง |
| Future Benefit / Officer tooling | HTTP endpoints provider | ตรวจสอบข้อมูลต้นทางระหว่าง debug claim |

ใน code ปัจจุบัน Orchestrator ไม่ได้ยิง HTTP กลับเข้า provider endpoint แต่ใช้ `ProviderService` instance เดียวกันผ่าน composition เพื่อเลี่ยง network hop ใน process เดียวกัน และลด logic ซ้ำ

## Base URL

Local development:

```text
http://localhost:8080
```

Provider database default:

```text
postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable
```

Environment variable:

| Name | Default | Description |
|------|---------|-------------|
| `PROVIDER_DATABASE_URL` | `postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable` | PostgreSQL connection string ของ provider database |

## Endpoint Summary

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/dopa/verify/{nationalId}` | ตรวจสถานะบุคคล อายุ และบัตร active |
| `GET` | `/api/v1/dopa/card-status/{nationalId}` | ตรวจสถานะบัตรประชาชนล่าสุด |
| `GET` | `/api/v1/sso/status/{nationalId}` | ตรวจสถานะผู้ประกันตนและมาตรา |
| `GET` | `/api/v1/sso/contribution/{nationalId}` | ตรวจจำนวนเดือนและประวัติเงินสมทบล่าสุด |
| `GET` | `/api/v1/ktb/financial-check/{nationalId}` | ตรวจยอดเงินฝากรวมและรายได้เฉลี่ย |
| `GET` | `/api/v1/ktb/account-status/{nationalId}` | ตรวจบัญชี active และ PromptPay |

## Common Rules

`nationalId` ต้องเป็นตัวเลข 13 หลักเท่านั้น

Common error responses:

| Status | Body | Cause |
|--------|------|-------|
| `400 Bad Request` | `{"error":"nationalId must be 13 digits"}` | `nationalId` ไม่ใช่ตัวเลข 13 หลัก |
| `404 Not Found` | `{"error":"provider record not found"}` | ไม่พบข้อมูลใน provider database |
| `500 Internal Server Error` | `{"error":"provider service unavailable"}` | database หรือ provider service ใช้งานไม่ได้ |

## GET /api/v1/dopa/verify/{nationalId}

ใช้ตรวจสถานะ citizen สำหรับ decision rule ของ Orchestrator

Source tables:

- `dopa.citizens`
- latest `dopa.id_cards`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "valid": true,
  "age": 35,
  "alive": true,
  "personStatus": "alive",
  "cardActive": true
}
```

Field:

| Field | Type | Description |
|-------|------|-------------|
| `nationalId` | string | เลขบัตรประชาชน 13 หลัก |
| `valid` | bool | true เมื่อ `alive=true` และ `cardActive=true` |
| `age` | int | อายุคำนวณจาก `date_of_birth` |
| `alive` | bool | true เมื่อ `person_status=alive` |
| `personStatus` | string | `alive`, `deceased`, `missing` |
| `cardActive` | bool | true เมื่อบัตรล่าสุดมี `card_status=active` |

Orchestrator mapping:

```json
{
  "valid": true,
  "age": 35,
  "alive": true,
  "cardActive": true
}
```

## GET /api/v1/dopa/card-status/{nationalId}

ใช้ตรวจสถานะบัตรประชาชนล่าสุด

Source table:

- latest `dopa.id_cards`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "cardStatus": "active",
  "cardActive": true,
  "issuedAt": "2022-01-10T00:00:00Z",
  "expiredAt": "2030-01-09T00:00:00Z",
  "checkedAt": "2026-05-19T00:00:00Z"
}
```

ถ้าบัตรถูก revoke อาจมี `revokedReason`

## GET /api/v1/sso/status/{nationalId}

ใช้ตรวจสถานะผู้ประกันตนและมาตราประกันสังคม

Source table:

- `sso.insured_persons`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "insuredStatus": "insured",
  "section": "40",
  "insured": true,
  "registeredAt": "2024-01-01T00:00:00Z"
}
```

Field:

| Field | Type | Description |
|-------|------|-------------|
| `insuredStatus` | string | `insured`, `uninsured`, `suspended`, `terminated` |
| `section` | string | `33`, `39`, `40`; ไม่มี field นี้เมื่อไม่เป็นผู้ประกันตน |
| `insured` | bool | true เมื่อ `insured_status=insured` |
| `employerId`, `employerName` | string | ใช้กับมาตรา 33 ถ้ามีข้อมูลนายจ้าง |

Orchestrator mapping:

```json
{
  "section": "40",
  "contributionMonths": 12
}
```

หมายเหตุ: `contributionMonths` มาจาก endpoint contribution หรือ service method `Status` ที่รวมข้อมูลสองชุดให้ Orchestrator

## GET /api/v1/sso/contribution/{nationalId}

ใช้ตรวจจำนวนเดือนที่ส่งเงินสมทบและประวัติล่าสุดไม่เกิน 12 งวด

Source tables:

- `sso.insured_persons`
- `sso.contributions`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "contributionMonths": 12,
  "latestContributionMonth": "2026-04-01T00:00:00Z",
  "recentContributions": [
    {
      "contributionMonth": "2026-04-01T00:00:00Z",
      "employeeAmount": 100,
      "employerAmount": 0,
      "governmentAmount": 50,
      "paidAt": "2026-04-15T00:00:00Z",
      "paymentStatus": "paid"
    }
  ]
}
```

## GET /api/v1/ktb/financial-check/{nationalId}

ใช้ตรวจข้อมูลการเงินสำหรับ eligibility rules

Source table:

- `ktb.bank_accounts`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "depositTotal": 12000,
  "averageMonthlyIncome": 15000,
  "activeAccountCount": 1
}
```

Orchestrator mapping:

```json
{
  "depositTotal": 12000,
  "averageMonthlyIncome": 15000,
  "promptPayLinked": true
}
```

หมายเหตุ: `promptPayLinked` มาจาก endpoint account-status หรือ service method `FinancialCheck` ที่รวมข้อมูลสองชุดให้ Orchestrator

## GET /api/v1/ktb/account-status/{nationalId}

ใช้ตรวจบัญชีธนาคารและ PromptPay

Source tables:

- `ktb.bank_accounts`
- `ktb.promptpay_registrations`

Response `200 OK`:

```json
{
  "nationalId": "1101700203451",
  "hasActiveAccount": true,
  "promptPayLinked": true,
  "accounts": [
    {
      "bankCode": "006",
      "branchCode": "0001",
      "accountNo": "006123456789",
      "accountName": "Somchai Jaidee",
      "accountType": "savings",
      "accountStatus": "active",
      "balance": 12000
    }
  ],
  "promptPay": [
    {
      "proxyType": "national_id",
      "proxyValue": "1101700203451",
      "registrationStatus": "active",
      "registeredAt": "2020-01-10T00:00:00Z"
    }
  ]
}
```

## Run Locally

จาก root project ให้เปิด provider database ก่อน:

```bash
docker compose -f database/docker-compose.db.yml up -d
```

ตรวจ container:

```bash
docker compose -f database/docker-compose.db.yml ps
```

จากนั้นรัน backend:

```bash
cd Backend
go run .
```

ถ้าต้องการ override database URL:

```bash
PROVIDER_DATABASE_URL=postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable go run .
```

บน PowerShell:

```powershell
$env:PROVIDER_DATABASE_URL="postgres://gss_user:gss_password@localhost:5433/gss_provider?sslmode=disable"
go run .
```

## Reset Local Database

ใช้เมื่อต้องการล้างข้อมูลและ seed ใหม่:

```bash
docker compose -f database/docker-compose.db.yml down -v
docker compose -f database/docker-compose.db.yml up -d
```

## Test

จากโฟลเดอร์ `Backend`:

```bash
go test ./...
```

ชุดทดสอบ provider repository และ handler จะพยายามต่อ database ที่ `PROVIDER_DATABASE_URL` หรือ default `localhost:5433`

ถ้า database ไม่ได้รัน test integration จะถูก skip เฉพาะส่วนที่ต้องใช้ database แต่ unit test อื่นยังรันต่อได้

## Example Requests

```bash
curl http://localhost:8080/api/v1/dopa/verify/1101700203451
curl http://localhost:8080/api/v1/dopa/card-status/1101700203451
curl http://localhost:8080/api/v1/sso/status/1101700203451
curl http://localhost:8080/api/v1/sso/contribution/1101700203451
curl http://localhost:8080/api/v1/ktb/financial-check/1101700203451
curl http://localhost:8080/api/v1/ktb/account-status/1101700203451
```

## Seed Cases

| National ID | Expected Scenario |
|-------------|-------------------|
| `1101700203451` | baseline eligible: alive, active card, SSO section 40, PromptPay active |
| `1101700203452` | under 18 / uninsured |
| `1101700203453` | SSO section 33 and high income |
| `1101700203454` | deceased / revoked card / frozen account |

## Team Workflow

1. Pull latest `develop`
2. Start provider database with Docker
3. Run backend
4. Test provider endpoints with seed IDs
5. If endpoint behavior changes, update this contract and `docs/database/provider-database.md` when database shape changes
6. Run `go test ./...` before opening PR
