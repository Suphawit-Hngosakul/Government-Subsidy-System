# Provider Mock Database

Owner: ธนรัชต แก้วลอย 6609650426

Status: Version 1, local PostgreSQL for DOPA / SSO / KTB mock web services

## Purpose

ฐานข้อมูลชุดนี้ใช้สำหรับ mock provider APIs ของ 3 หน่วยงานที่ Orchestrator ต้องเรียก:

- DOPA: ตรวจตัวตน สถานะบุคคล และสถานะบัตรประชาชน
- SSO: ตรวจมาตราประกันสังคมและประวัติการส่งเงินสมทบ
- KTB: ตรวจข้อมูลทางการเงิน บัญชี และ PromptPay

Schema นี้เป็น **mock schema สำหรับงานกลุ่มเท่านั้น** ไม่ใช่ schema จริงของ DOPA, SSO หรือ KTB เพราะ schema ภายในของหน่วยงานไม่ได้เปิดเผยสาธารณะ

## Sources Used

ใช้ข้อมูลสาธารณะเป็นกรอบในการออกแบบ field:

- Thai National ID card มีเลขประจำตัว 13 หลัก และข้อมูลบนบัตร เช่น ชื่อ วันเกิด วันออกบัตร วันหมดอายุ และที่อยู่
- Social Security Office ใช้มาตรา 33, 39 และ 40 และมีแนวคิดเรื่องผู้ประกันตนกับการส่งเงินสมทบรายเดือน
- PromptPay ใช้ proxy เช่น citizen ID, mobile number หรือ e-wallet ID เพื่อผูกกับบัญชีธนาคาร

## Local Docker

จาก root project:

```bash
docker compose -f database/docker-compose.db.yml up -d
```

Connection:

```text
Host: localhost
Port: 5433
Database: gss_provider
User: gss_user
Password: gss_password
```

หยุด database:

```bash
docker compose -f database/docker-compose.db.yml down
```

ลบข้อมูลทั้งหมดแล้ว seed ใหม่:

```bash
docker compose -f database/docker-compose.db.yml down -v
docker compose -f database/docker-compose.db.yml up -d
```

## Schema Layout

ใช้ PostgreSQL database เดียวชื่อ `gss_provider` แล้วแยก schema ตาม provider:

| Schema | Purpose |
|--------|---------|
| `dopa` | ข้อมูลประชาชนและบัตรประชาชน |
| `sso` | ข้อมูลผู้ประกันตนและเงินสมทบ |
| `ktb` | ข้อมูลบัญชีธนาคารและ PromptPay |

## DOPA Tables

### `dopa.citizens`

ข้อมูลประชาชนหลัก keyed ด้วย `national_id`

| Field | Description |
|-------|-------------|
| `national_id` | เลขบัตรประชาชน 13 หลัก |
| `title_th`, `first_name_th`, `last_name_th` | ชื่อภาษาไทย |
| `title_en`, `first_name_en`, `last_name_en` | ชื่อภาษาอังกฤษ |
| `date_of_birth` | วันเกิด ใช้คำนวณอายุ |
| `gender` | `M`, `F`, `X` |
| `address_line`, `province`, `district`, `subdistrict`, `postal_code` | ที่อยู่ |
| `person_status` | `alive`, `deceased`, `missing` |

### `dopa.id_cards`

ข้อมูลบัตรประชาชนของ citizen

| Field | Description |
|-------|-------------|
| `national_id` | FK ไปที่ `dopa.citizens` |
| `laser_code` | เลขหลังบัตรสำหรับ mock eKYC |
| `issued_at`, `expired_at` | วันออกบัตรและวันหมดอายุ |
| `card_status` | `active`, `expired`, `revoked`, `lost` |
| `revoked_reason` | เหตุผลกรณีบัตรถูก revoke |

Endpoint mapping:

| Endpoint | Tables |
|----------|--------|
| `GET /api/v1/dopa/verify/:nationalId` | `dopa.citizens`, latest `dopa.id_cards` |
| `GET /api/v1/dopa/card-status/:nationalId` | `dopa.id_cards` |

## SSO Tables

### `sso.insured_persons`

ข้อมูลสถานะผู้ประกันตน

| Field | Description |
|-------|-------------|
| `national_id` | เลขบัตรประชาชน 13 หลัก |
| `insured_status` | `insured`, `uninsured`, `suspended`, `terminated` |
| `section` | `33`, `39`, `40` หรือ `NULL` ถ้าไม่เป็นผู้ประกันตน |
| `registered_at` | วันที่เริ่มขึ้นทะเบียน |
| `employer_id`, `employer_name` | ใช้กับมาตรา 33 |
| `contribution_months` | จำนวนเดือนที่ส่งเงินสมทบ |
| `latest_contribution_month` | เดือนล่าสุดที่มีข้อมูลสมทบ |

### `sso.contributions`

ประวัติการส่งเงินสมทบรายเดือน

| Field | Description |
|-------|-------------|
| `national_id` | FK ไปที่ `sso.insured_persons` |
| `contribution_month` | เดือนของงวดสมทบ |
| `employee_amount`, `employer_amount`, `government_amount` | จำนวนเงินแต่ละส่วน |
| `paid_at` | วันที่ชำระ |
| `payment_status` | `paid`, `late`, `missing`, `refunded` |

Endpoint mapping:

| Endpoint | Tables |
|----------|--------|
| `GET /api/v1/sso/status/:nationalId` | `sso.insured_persons` |
| `GET /api/v1/sso/contribution/:nationalId` | `sso.insured_persons`, `sso.contributions` |

## KTB Tables

### `ktb.bank_accounts`

ข้อมูลบัญชีและภาพรวมการเงิน

| Field | Description |
|-------|-------------|
| `national_id` | เลขบัตรประชาชน 13 หลัก |
| `bank_code` | default `006` สำหรับ KTB |
| `account_no`, `account_name` | เลขบัญชีและชื่อบัญชี |
| `account_type` | `savings`, `current`, `wallet` |
| `account_status` | `active`, `closed`, `frozen`, `dormant` |
| `balance` | ยอดเงินฝากปัจจุบันสำหรับ mock financial check |
| `average_monthly_income` | รายได้เฉลี่ยต่อเดือนสำหรับ eligibility rules |

### `ktb.promptpay_registrations`

ข้อมูลการผูก PromptPay

| Field | Description |
|-------|-------------|
| `national_id` | เลขบัตรประชาชน 13 หลัก |
| `account_id` | FK ไปที่ `ktb.bank_accounts` |
| `proxy_type` | `national_id`, `mobile`, `ewallet` |
| `proxy_value` | ค่า proxy ที่ผูกไว้ |
| `registration_status` | `active`, `inactive`, `revoked` |

Endpoint mapping:

| Endpoint | Tables |
|----------|--------|
| `GET /api/v1/ktb/financial-check/:nationalId` | `ktb.bank_accounts` |
| `GET /api/v1/ktb/account-status/:nationalId` | `ktb.bank_accounts`, `ktb.promptpay_registrations` |

## Seed Cases

| National ID | Expected Scenario |
|-------------|-------------------|
| `1101700203451` | eligible baseline: alive, active card, SSO section 40, PromptPay active |
| `1101700203452` | under 18 / uninsured, useful for rejected age rule |
| `1101700203453` | SSO section 33 and high income, useful for rejected rules |
| `1101700203454` | deceased / revoked card / frozen account, useful for hard reject |

## Runtime Integration

Version 1 เชื่อม database นี้เข้ากับ Go backend แล้วผ่าน provider module:

| Layer | File |
|-------|------|
| HTTP handler | `Backend/controller/provider_handler.go` |
| Service / Orchestrator adapter composition | `Backend/service/provider_service.go` |
| PostgreSQL repository | `Backend/repository/provider_repository.go` |
| Domain response models | `Backend/domain/provider.go` |

Orchestrator ใช้ `ProviderService` instance เดียวกันผ่าน interface `DOPAClient`, `SSOClient`, `KTBClient` ใน `service/orchestrator_service.go`

## Contract Link

API contract ของ provider อยู่ที่:

```text
docs/api/provider-contract.md
```

ถ้ามีการแก้ endpoint, response field, error code หรือ table/column ที่ response พึ่งพา ต้องอัปเดต contract และเอกสาร database ใน PR เดียวกัน
