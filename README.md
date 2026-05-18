# Ruam Thai Srang Chati (รวมไทยสร้างชาติ)
ระบบนี้คือ **Mini X-Road** แบบไทย แนวคิดมาจาก X-Road ของเอสโตเนีย ซึ่งเป็นโครงสร้างพื้นฐานดิจิทัลที่เชื่อมข้อมูลระหว่างหน่วยงานรัฐทั้งประเทศเข้าด้วยกัน

### ปัญหาที่ระบบนี้แก้
ปัจจุบันเวลาประชาชนจะขอรับสวัสดิการจากรัฐ ต้องไปติดต่อหลายหน่วยงาน ยื่นเอกสารซ้ำซ้อน และรอนานเพราะแต่ละหน่วยงานข้อมูลไม่เชื่อมกัน ระบบนี้แก้ปัญหานั้นด้วยหลักการ **"Once-Only"** ประชาชนให้ข้อมูลครั้งเดียว ระบบดึงข้อมูลที่เหลือเองจากหน่วยงานที่เกี่ยวข้อง

### ระบบทำงานอย่างไร
จำลอง 3 หน่วยงานรัฐ ได้แก่ กรมการปกครอง (DOPA), สำนักงานประกันสังคม (SSO) และธนาคารกรุงไทย (KTB) แล้วเชื่อมทั้ง 3 เข้าหากันผ่าน **Middleware กลาง** เมื่อประชาชนกดขอรับสิทธิ์ ระบบจะยิง API ไปถามทั้ง 3 แหล่งพร้อมกัน แล้วนำผลมาตัดสินโดยอัตโนมัติว่า Approve หรือ Reject โดยไม่ต้องให้เจ้าหน้าที่ตรวจทีละคน

## Member
| ชื่อ-นามสกุล | รหัสนักศึกษา | GitHub | 
|---|---|---|
 | ศุภวิชญ์ หงอสกุล | 6609650053 |  [@วิว](https://github.com/Suphawit-Hngosakul)  | 
 | ธนธัช นิติเจริญ | 6609650095 |  [@น้ำปั่น](https://github.com/thanattouth)  | 
 | ชญานนท์ ปูรณะปุระ | 6609650251 |  [@เก็ท](https://github.com/Yonna4248)  |
 | ณัฏฐ์นลิน บุญทรัพย์ทีปกร | 6609650343 |  [@นิว](https://github.com/NutnarinBunsupteepakorn-6609650343)  |
 | ธนรัชต แก้วลอย | 6609650426 |  [@บอส](https://github.com/btkkkkkkk)  |

---


## คนที่ 1 — ชญานนท์ ปูรณะปุระ 6609650251

#### AUTH & EKYC
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/auth/register` | กรอกเลข 13 หลัก + password + phone |
| `POST` | `/api/v1/auth/ekyc/ocr` | ส่งรูปบัตร → Tesseract / seed fallback → auto-fill |
| `POST` | `/api/v1/auth/ekyc/confirm` | ยืนยันข้อมูล OCR → `kyc_verified = true` |
| `POST` | `/api/v1/auth/login` | login → JWT token (role: citizen) |
| `POST` | `/api/v1/auth/logout` | revoke token |
---

## คนที่ 2 — ณัฏฐ์นลิน บุญทรัพย์ทีปกร 6609650343

#### BENEFIT (CITIZEN)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/api/v1/benefit/claim` | ยื่นคำร้องขอสิทธิ์ → trigger orchestrator |
| `GET` | `/api/v1/benefit/status/:claimId` | เช็คสถานะ real-time |
| `GET` | `/api/v1/benefit/history/:citizenId` | ประวัติคำร้องทั้งหมด |
| `GET` | `/api/v1/benefit/projects` | รายการโครงการที่เปิดรับสิทธิ์ |
---

## คนที่ 3 — ธนรัชต แก้วลอย 6609650426

#### DOPA MOCK
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/dopa/verify/:nationalId` | ตรวจ validity + อายุ + สถานะบุคคล |
| `GET` | `/api/v1/dopa/card-status/:nationalId` | เช็คบัตรหมดอายุ / ถูก revoke หรือไม่ |

#### SSO MOCK
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/sso/status/:nationalId` | มาตรา 33 / 39 / 40 / ไม่เป็นผู้ประกันตน |
| `GET` | `/api/v1/sso/contribution/:nationalId` | ระยะเวลาส่งเงินสมทบ (เดือน) |

#### KTB MOCK
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/ktb/financial-check/:nationalId` | ยอดเงินฝากรวม + รายได้เฉลี่ย/เดือน |
| `GET` | `/api/v1/ktb/account-status/:nationalId` | มีบัญชีพร้อมเพย์ผูกไว้หรือไม่ |
---

## คนที่ 4 — ธนธัช นิติเจริญ 6609650095

#### ORCHESTRATOR (INTERNAL)
| Method | Endpoint | Description |
|--------|----------|-------------|
| `POST` | `/internal/orchestrate` | เรียก DOPA + SSO + KTB แบบ parallel |
| `POST` | `/internal/decision` | ประเมิน eligibility rules → Approve/Reject/Pending |

#### STATUS STREAM
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/claim/:claimId/stream` | SSE — push สถานะ real-time ให้ frontend |

Contract: [`docs/api/orchestrator-contract.md`](docs/api/orchestrator-contract.md)
---

## คนที่ 5 — ศุภวิชญ์ หงอสกุล 6609650053

#### ADMIN — จัดการโครงการ
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/projects` | รายการโครงการทั้งหมด |
| `POST` | `/api/v1/admin/project` | สร้างโครงการ + ตั้งเกณฑ์สิทธิ์ |
| `PUT` | `/api/v1/admin/project/:id` | แก้ไขโครงการ |
| `DELETE` | `/api/v1/admin/project/:id` | ลบโครงการ |

#### OFFICER — ตรวจสอบคำร้อง
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/officer/claims` | รายการ Pending ทั้งหมด |
| `GET` | `/api/v1/officer/claim/:id` | รายละเอียดคำร้อง + ผล 3 sources |
| `PATCH` | `/api/v1/officer/claim/:id/approve` | อนุมัติด้วยตนเอง |
| `PATCH` | `/api/v1/officer/claim/:id/reject` | ปฏิเสธพร้อม reason |

#### DASHBOARD
| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/admin/stats` | ยอด Approve / Reject / Pending รวม |
| `GET` | `/api/v1/admin/audit-log` | log ทุก action ใครทำอะไร เมื่อไหร่ |
