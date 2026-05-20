# Auth & eKYC Service Contract

Owner: ชญานนท์ ปูรณะปุระ 6609650251

Status: Version 1, mock implementation

## Purpose

Auth & eKYC Service รับผิดชอบการลงทะเบียน, ยืนยันตัวตนด้วย eKYC และออก JWT token สำหรับประชาชน (role: citizen) ใน Version 1 ใช้ in-memory repository และ mock OCR โดยไม่ต้องพึ่ง Tesseract หรือ external provider จริง

## Contract Maintenance Rule

ทุกครั้งที่มีการแก้ไข endpoint, HTTP method, request body, response body, status code หรือ auth flow ต้องอัปเดตไฟล์นี้ใน PR เดียวกันเสมอ

## Callers

| Caller | Endpoint ที่เรียก | เหตุผล |
|--------|-------------------|--------|
| Frontend / Citizen | `POST /api/v1/auth/register` | ลงทะเบียนบัญชีประชาชนใหม่ |
| Frontend / Citizen | `POST /api/v1/auth/ekyc/ocr` | ส่งรูปบัตรเพื่อ OCR auto-fill ข้อมูล |
| Frontend / Citizen | `POST /api/v1/auth/ekyc/confirm` | ยืนยัน OCR data เพื่อ set `kyc_verified = true` |
| Frontend / Citizen | `POST /api/v1/auth/login` | เข้าสู่ระบบ → ได้รับ JWT token |
| Frontend / Citizen | `POST /api/v1/auth/logout` | ยกเลิก token |

## Base URL

Local development:

```text
http://localhost:8080
```

## Endpoint Summary

| Method | Endpoint | Auth Required | Description |
|--------|----------|---------------|-------------|
| `POST` | `/api/v1/auth/register` | No | ลงทะเบียนบัญชีใหม่ด้วยเลขบัตร + รหัสผ่าน + เบอร์โทร |
| `POST` | `/api/v1/auth/ekyc/ocr` | No | รับรูปบัตรหรือ nationalId → คืน mock OCR data |
| `POST` | `/api/v1/auth/ekyc/confirm` | No | ยืนยัน OCR data → เปลี่ยน kycStatus เป็น verified |
| `POST` | `/api/v1/auth/login` | No | ตรวจรหัส → คืน JWT token |
| `POST` | `/api/v1/auth/logout` | Yes (Bearer) | revoke token |

---

## POST /api/v1/auth/register

ลงทะเบียนบัญชีประชาชนใหม่

Request body:

```json
{
  "nationalId": "1234567890123",
  "password": "mysecretpassword",
  "phone": "0812345678"
}
```

Field:

| Field | Required | Description |
|-------|----------|-------------|
| `nationalId` | Yes | เลขบัตรประชาชน 13 หลัก |
| `password` | Yes | รหัสผ่าน |
| `phone` | Yes | เบอร์โทรศัพท์ |

Response `201 Created`:

```json
{
  "message": "registered successfully"
}
```

Error `400 Bad Request`:

```json
{
  "error": "citizen already registered"
}
```

```json
{
  "error": "national ID must be 13 digits"
}
```

---

## POST /api/v1/auth/ekyc/ocr

รับรูปบัตรประชาชน แล้วอ่านข้อมูลผ่าน Gemini 2.0 Flash Vision เพื่อ auto-fill ฟอร์ม

**OCR Mode** (มี `GEMINI_API_KEY` + ส่งรูปมา):
- ส่งรูปไป Gemini Vision → อ่านบัตรจริง → คืนข้อมูลที่อ่านได้

**Seed Fallback Mode** (ไม่มี key หรือไม่ส่งรูป):
- คืน mock data จาก nationalId (สำหรับ dev/test ที่ไม่มี key)

รองรับ 2 รูปแบบ:

**รูปแบบที่ 1 — multipart/form-data** (แนะนำ, รองรับ Gemini OCR):

```
Content-Type: multipart/form-data

image: <file>       ← รูปบัตรประชาชน (jpg/png)
nationalId: 1234567890123  ← ใช้เป็น fallback ถ้า OCR ล้มเหลว
```

**รูปแบบที่ 2 — JSON body** (Seed fallback เท่านั้น):

```json
{
  "nationalId": "1234567890123"
}
```

Response `200 OK`:

```json
{
  "nationalId": "1234567890123",
  "fullName": "นาย ทดสอบ ระบบ",
  "dateOfBirth": "01/01/2533",
  "laserCode": "AA1-1234567-89",
  "address": "1/1 ถ.สุขุมวิท แขวงคลองเตย เขตคลองเตย กรุงเทพมหานคร 10110"
}
```

Error `400 Bad Request`:

```json
{
  "error": "nationalId or image is required"
}
```

### Environment Variable

| Variable | Required | Description |
|----------|----------|-------------|
| `GEMINI_API_KEY` | No | Google AI Studio API key — ถ้าไม่มีจะใช้ seed fallback |

### OCR Decision Logic

```
มีรูป + มี GEMINI_API_KEY  →  Gemini Vision (อ่านจริง)
Gemini error              →  seed fallback อัตโนมัติ (demo ไม่พัง)
ไม่มีรูป หรือ ไม่มี key   →  seed fallback จาก nationalId
```

---

## POST /api/v1/auth/ekyc/confirm

ยืนยันข้อมูลที่ได้จาก OCR เพื่อเปลี่ยน kycStatus ของประชาชนเป็น `verified`

Request body:

```json
{
  "nationalId": "1234567890123",
  "laserCode": "AA1-1234567-89",
  "fullName": "นาย ทดสอบ ระบบ",
  "dateOfBirth": "01/01/2533"
}
```

Field:

| Field | Required | Description |
|-------|----------|-------------|
| `nationalId` | Yes | เลขบัตรประชาชน 13 หลัก |
| `laserCode` | Yes | รหัส laser code ด้านหลังบัตร |
| `fullName` | No | ชื่อ-นามสกุลจาก OCR |
| `dateOfBirth` | No | วันเกิดจาก OCR |

Response `200 OK`:

```json
{
  "message": "KYC verified successfully",
  "kycStatus": "verified"
}
```

Error `400 Bad Request`:

```json
{
  "error": "citizen not found"
}
```

> **Prerequisite**: ต้องเคย `POST /api/v1/auth/register` ก่อน ถึงจะ confirm KYC ได้

---

## POST /api/v1/auth/login

ตรวจสอบ credentials แล้วออก JWT token (role: citizen)

Request body:

```json
{
  "nationalId": "1234567890123",
  "password": "mysecretpassword"
}
```

Response `200 OK`:

```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "role": "citizen"
}
```

Error `401 Unauthorized`:

```json
{
  "error": "invalid credentials"
}
```

Token spec:

| Field | Value |
|-------|-------|
| Algorithm | HS256 (HMAC-SHA256) |
| Payload fields | `sub` (citizenID), `role`, `exp` (24h), `jti` (random) |
| Transport | `Authorization: Bearer <token>` |

---

## POST /api/v1/auth/logout

ยกเลิก token ที่ใช้งานอยู่ (in-memory revocation)

Request header:

```text
Authorization: Bearer <token>
```

Response `200 OK`:

```json
{
  "message": "logged out successfully"
}
```

Error `401 Unauthorized`:

```json
{
  "error": "missing authorization token"
}
```

```json
{
  "error": "token not found or already revoked"
}
```

> **Note v1**: token revocation เก็บใน memory ถ้า restart service token ที่ revoke แล้วจะถูกยืนยันได้อีกครั้ง

---

## eKYC Flow (ลำดับการเรียก)

```text
1. POST /api/v1/auth/register      → สร้างบัญชี (kycStatus: pending)
2. POST /api/v1/auth/ekyc/ocr      → รับข้อมูล mock OCR
3. POST /api/v1/auth/ekyc/confirm  → ยืนยัน → kycStatus: verified
4. POST /api/v1/auth/login         → เข้าสู่ระบบ → ได้ JWT token
```

## Run Locally

จากโฟลเดอร์ `Backend`:

```bash
go run .
```

ตัวอย่าง request:

```bash
# Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"nationalId":"1234567890123","password":"secret","phone":"0812345678"}'

# eKYC OCR
curl -X POST http://localhost:8080/api/v1/auth/ekyc/ocr \
  -H 'Content-Type: application/json' \
  -d '{"nationalId":"1234567890123"}'

# eKYC Confirm
curl -X POST http://localhost:8080/api/v1/auth/ekyc/confirm \
  -H 'Content-Type: application/json' \
  -d '{"nationalId":"1234567890123","laserCode":"AA1-1234567-89","fullName":"นาย ทดสอบ ระบบ","dateOfBirth":"01/01/2533"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"nationalId":"1234567890123","password":"secret"}'

# Logout
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H 'Authorization: Bearer <token>'
```

## Test

จากโฟลเดอร์ `Backend`:

```bash
go test ./...
```
