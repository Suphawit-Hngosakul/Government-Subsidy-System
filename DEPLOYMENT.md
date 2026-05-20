# 🚀 คู่มือการติดตั้งระบบด้วย Docker & Cloud Deployment (Render / Local)

คู่มือการติดตั้งระบบจัดการทุนสนับสนุนจากรัฐบาล (Government Subsidy System) ในรูปแบบ Container ด้วย **Docker Compose** สำหรับเครื่องพัฒนาในเครื่องตนเอง (Local Development) และการติดตั้งขึ้นคลาวด์ **Render** ด้วยระบบ Blueprint IaC (Infrastructure as Code) หรือคลาวด์ทางเลือกอื่นๆ เช่น **Railway** และ **Fly.io**

---

## 📂 โครงสร้าง DevOps ในโปรเจกต์
ระบบถูกออกแบบโครงสร้างโมดูลการตั้งค่า Docker ไว้แยกสัดส่วนเพื่อความโปร่งใสและบำรุงรักษาง่าย:
```
Government-Subsidy-System/
├── Backend/
│   └── Dockerfile          # Multi-stage Go production runner (Alpine-based, ~25MB)
├── Frontend/
│   └── Dockerfile          # Next.js standalone runner (Node-Alpine, ~150MB)
├── docker-compose.yml      # ตัวควบคุมจัดการ Postgres + Go Backend + Next.js ในเครื่องตนเอง
├── .env.example            # เทมเพลตสำหรับจัดการค่าตัวแปรสิ่งแวดล้อม (Environment Variables)
├── render.yaml             # Render Blueprint สำหรับคลิกติดตั้งระบบขึ้น Cloud ทั้งหมดในคลิกเดียว
└── DEPLOYMENT.md           # [คู่มือฉบับนี้]
```

---

## 🛠️ ส่วนที่ 1: การติดตั้งและใช้งานแบบ Local (ด้วย Docker Compose)

เพื่อให้คุณสามารถรันและทดสอบระบบทั้งหมด (หน้าบ้าน, หลังบ้าน, และฐานข้อมูลพร้อมข้อมูลจำลอง) ในเครื่องของคุณได้ทันทีโดยไม่ต้องลงซอฟต์แวร์เสริมอื่นใดนอกจาก Docker:

### 1. การจัดเตรียมไฟล์ตัวแปรสิ่งแวดล้อม (Environment File)
คัดลอกไฟล์เทมเพลต `.env.example` ไปสร้างเป็นไฟล์ `.env` ที่โฟลเดอร์ Root ของโปรเจกต์:
```bash
cp .env.example .env
```
เปิดไฟล์ `.env` และเพิ่มคีย์ **Google Gemini API Key** ของคุณ (จำเป็นต้องใช้สำหรับโมดูล eKYC และสแกนหน้าบัตรประชาชนด้วย OCR):
```env
GEMINI_API_KEY=AIzaSyxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```
*(คุณสามารถสร้างคีย์ฟรียังไม่มีค่าใช้จ่ายได้จาก [Google AI Studio](https://aistudio.google.com/))*

---

### 2. สั่งเปิดใช้งานระบบ (Start Containers)
พิมพ์คำสั่งด้านล่างนี้ที่โฟลเดอร์ Root ของโปรเจกต์เพื่อบิลด์และบูสระบบทั้งหมดขึ้นมา:
```bash
docker compose up --build
```

**กลไกการทำงานอัตโนมัติเบื้องหลัง:**
1. ระบบจะสั่งเปิดเซิร์วิสฐานข้อมูล `postgres` และรันสคริปต์โครงสร้างฐานข้อมูล (`001_provider_schema.sql`) และตัวข้อมูล Seed จำลองจำพวกกองทุนรวมและผู้ให้บริการตรวจสอบบัตรประชาชน (`002_provider_seed.sql`) โดยอัตโนมัติ
2. ระบบจะตรวจเช็คสถานะความพร้อมใช้งาน (Healthcheck) ของฐานข้อมูล PostgreSQL จนกว่าจะพร้อมรับคำสั่งอย่างสมบูรณ์
3. เซิร์วิสหลังบ้าน `backend` (Go) จะเริ่มรันเชื่อมโยงฐานข้อมูล PostgreSQL
4. เซิร์วิสหน้าบ้าน `frontend` (Next.js) จะเริ่มต้นทำงานพร้อมให้คุณเรียกใช้งาน

---

### 3. ที่อยู่การเชื่อมต่อใช้งานในเครื่อง (Local Access URLs)

| บริการ / Service | พอร์ต / Port | ที่อยู่การเข้าถึง / URL |
| :--- | :--- | :--- |
| **Frontend (Next.js)** | `3000` | [http://localhost:3000](http://localhost:3000) |
| **Backend API (Go)** | `8080` | [http://localhost:8080](http://localhost:8080) |
| **PostgreSQL Database** | `5433` | Binds to `localhost:5433` *(ภายนอกโฮสต์ใช้พอร์ต 5433 แต่ภายใน Docker เครือข่ายใช้ 5432)* |

---

### 4. การปิดใช้งานและคืนค่าพื้นที่ (Stop & Cleanup)
หากต้องการปิดแอปพลิเคชันและลบสถานะการทำงานออกทั้งหมดรวมถึง Volume ข้อมูล:
```bash
docker compose down -v
```

---

## ☁️ ส่วนที่ 2: การติดตั้งขึ้น Cloud ด้วย Render (ผ่าน render.yaml Blueprint)

[Render.com](https://render.com) รองรับระบบ **Blueprint Infrastructure as Code (IaC)** ซึ่งช่วยให้คุณระบุการตั้งค่าของ Database และ Web Service ของทั้งระบบไว้ในโค้ด และสั่งเปิดใช้งานทั้งหมดได้ใน "คลิกเดียว"

### 📋 ขั้นตอนการติดตั้งบน Render:

1. **อัปโหลดโค้ดขึ้น Git Repository:**
   พุชโค้ดของโปรเจกต์ `Government-Subsidy-System` ทั้งหมดขึ้นหน้า GitHub, GitLab หรือ Bitbucket ส่วนตัวของคุณ

2. **เชื่อมโยง Blueprint บน Render:**
   * ล็อกอินเข้าใช้งานบนคลาวด์ [Render Dashboard](https://dashboard.render.com/)
   * ไปที่แถบเมนูด้านบน คลิกปุ่ม **New +** ➔ เลือก **Blueprint**
   * เชื่อมโยงบัญชีและเลือก Repository ที่เก็บโค้ดตัวแปรโปรเจกต์ของคุณ

3. **กรอกข้อมูลความปลอดภัยและการอนุมัติ:**
   * Render จะสแกนหาไฟล์ `render.yaml` ในโปรเจกต์โดยอัตโนมัติ และจะแสดงแผนภาพระบบบริการ 3 อย่างที่จะถูกสร้างขึ้น:
     * **`gss-db`** (Managed PostgreSQL Database Instance)
     * **`gss-backend`** (Go Web Service - Docker-based)
     * **`gss-frontend`** (Next.js Web Service - Docker-based)
   * **กำหนด Gemini Key:** ในหน้าเว็บจะแสดงช่องให้ระบุค่าของ `GEMINI_API_KEY` ให้คุณคัดลอกคีย์จาก Google AI Studio ไปใส่ไว้เพื่อความปลอดภัย
   * คลิกปุ่ม **Apply** เพื่ออนุมัติการดำเนินงาน

4. **การรัน SQL Seed บนคลาวด์ Render:**
   เนื่องจากฐานข้อมูล PostgreSQL แบบจัดการให้ (Managed) บน Render จะเปิดขึ้นมาใหม่โดยไม่มีตารางข้อมูล คุณสามารถรันสคริปต์เริ่มต้นข้อมูลได้ง่าย ๆ ดังนี้:
   * เมื่อบริการ `gss-db` และ `gss-backend` สร้างเสร็จสมบูรณ์แล้ว ให้ไปที่เซิร์วิส `gss-backend` ในหน้า Render
   * ไปที่แถบเมนู **Shell** เพื่อพิมพ์สั่งรัน SQL เริ่มต้นเข้าสู่ฐานข้อมูลคลาวด์โดยตรง หรือจะเชื่อมโยงเข้าฐานข้อมูลด้วยเครื่องมือภายนอก (เช่น DBeaver) ผ่านสาย Connection String ที่ Render มอบให้ แล้วสั่งรัน SQL สคริปต์ตามขั้นตอน:
     1. [001_provider_schema.sql](file:///Users/hamin/Documents/CS367/Government-Subsidy-System/database/postgres/init/001_provider_schema.sql)
     2. [002_provider_seed.sql](file:///Users/hamin/Documents/CS367/Government-Subsidy-System/database/postgres/init/002_provider_seed.sql)

---

## 🚀 ส่วนที่ 3: คลาวด์ทางเลือกอื่นๆ (Railway & Fly.io)

### 🚂 ทางเลือก A: Railway.app (แนะนำอย่างยิ่งสำหรับมือใหม่)
[Railway](https://railway.app) สามารถวิเคราะห์ Dockerfile และสร้างฐานข้อมูล PostgreSQL เพื่อผูกมัดให้อัตโนมัติในลักษณะที่ง่ายมาก:
1. คลิก **New Project** ➔ **Deploy from GitHub repo**
2. เลือก Repository ของคุณ
3. Railway จะวิเคราะห์และสร้างเซิร์วิส Backend และ Frontend ขึ้นมา
4. คลิก **Add Service +** ➔ เลือก **Database** ➔ เลือก **Add PostgreSQL**
5. **ตั้งค่าตัวแปร (Variables):**
   * ไปที่เซิร์วิส **Backend** ➔ ไปที่แถบ **Variables** ➔ คลิก **Add Variable**
   * ใส่ `PROVIDER_DATABASE_URL` ให้เชื่อมกับคีย์ `${{Postgres.DATABASE_URL}}` (ระบบ Railway จะส่งสายเชื่อมต่อจริงของฐานข้อมูลมาให้โดยอัตโนมัติ!)
   * ใส่ `GEMINI_API_KEY` และ `PORT` เป็น `8080`
   * ในแถบตั้งค่าของ **Frontend** ให้เพิ่มตัวแปร `GOV_SUBSIDY_BACKEND_URL` ชี้ไปที่ URL ปลายทางของเซิร์วิส Backend บน Railway ของคุณ
6. เริ่มทดสอบและใช้งานได้ทันที!

### 🎈 ทางเลือก B: Fly.io (สำหรับงานที่ต้องการประสิทธิภาพสูง)
[Fly.io](https://fly.io) รันด้วยระบบ Firecracker microVMs ซึ่งมอบความเร็วในการโหลดสูงและมีความหน่วงต่ำมาก:
1. ติดตั้ง Fly CLI ในเครื่องคอมพิวเตอร์ของคุณ
2. ล็อกอินผ่านเทอร์มินัล: `fly auth login`
3. ไปที่โฟลเดอร์ `Backend/` แล้วพิมพ์คำสั่งเริ่มต้นเซิร์วิสหลังบ้าน:
   ```bash
   fly launch --dockerfile Dockerfile --name gss-backend-api
   ```
   *(Fly.io จะสร้าง Postgres Cluster ในคลาวด์และเชื่อมสายคอนเนคชัน DSN ให้อัตโนมัติ)*
4. ตั้งค่า Gemini API Key ไว้เป็นคีย์ความลับ:
   ```bash
   fly secrets set GEMINI_API_KEY=AIzaSyxxxxxxxxxxxxxxxxx
   ```
5. เข้าไปที่โฟลเดอร์ `Frontend/` และสั่งดีพลอยหน้าบ้าน:
   ```bash
   fly launch --dockerfile Dockerfile --name gss-frontend-app
   ```
   *(ระบุ Environment variables ในส่วน `fly.toml` ในตัวแปร `GOV_SUBSIDY_BACKEND_URL` ให้ชี้ไปยังโดเมนหลังบ้านของคุณ)*

---

## 💡 คำแนะนำสำหรับการพัฒนาและแก้ไขปัญหา (Troubleshooting)

### 🔒 ปัญหาเรื่อง CORS (Cross-Origin Resource Sharing)
หากนำแอปพลิเคชัน Frontend และ Backend ไปวางต่างเซิร์ฟเวอร์หรือต่างโดเมนกัน (เช่น `gss-frontend.onrender.com` และ `gss-backend.onrender.com`) เบราว์เซอร์อาจสกัดการรับส่งข้อมูลผ่าน CORS:
* ระบบ Backend ในโปรเจกต์นี้ใช้ Go Standard Library ซึ่งดักจัดการ Header ได้ที่ระดับตัวครอบ Middleware คุณสามารถเพิ่มหรือตรวจสอบการส่ง Header `Access-Control-Allow-Origin: *` หรือชี้เฉพาะโดเมนหน้าบ้านของคุณใน Middleware ของ Controller เพื่อความปลอดภัยขั้นสูง

### 🩺 การตรวจเช็คสถานะการทำงาน (Debugging in Production)
หากตัวแอปพลิเคชันไม่ยอมเริ่มต้นทำงาน หรือไม่แสดงข้อมูล:
* **เช็คสายเชื่อมต่อฐานข้อมูล:** ให้แน่ใจว่าได้ระบุสาย DSN `sslmode=require` หรือ `sslmode=disable` ให้ถูกต้องตามมาตรฐานของผู้ให้บริการคลาวด์แต่ละเจ้า (เช่น Render จะบังคับ SSL เสมอ ส่วนการรัน Local Docker ของเราจะปิดการใช้งานไว้)
* **การเช็คระดับหน่วยความจำ (Out of Memory):** บิลด์ Next.js ในแบบปกติจะใช้แรมค่อนข้างมากในการคอมไพล์โค้ด แต่เนื่องจากเราใช้งานตัวเลือก **Standalone** ส่งผลให้ระบบประหยัดหน่วยความจำในการรันอย่างเห็นได้ชัด ทำให้สามารถประหยัดค่าบริการคลาวด์และใช้งานบนแผนประหยัด (Free Tier) ได้อย่างสบาย
