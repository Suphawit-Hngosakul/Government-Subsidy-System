# 🧪 Backend Unit Testing & Interactive Coverage Dashboard Guide

Bilingual Guide: 🇹🇭 ภาษาไทย | 🇺🇸 English

Welcome to the backend testing and reporting directory of the Government Subsidy System. This guide provides comprehensive documentation on the system's unit testing architecture, advanced mocking techniques, standard Go test commands, and the premium interactive JMeter-style HTML test & coverage dashboard.

ยินดีต้อนรับสู่โฟลเดอร์สำหรับจัดการงานทดสอบและระบบรายงานของระบบจัดการทุนสนับสนุนจากรัฐบาล (Backend) เอกสารฉบับนี้จัดทำขึ้นแบบสองภาษาเพื่อแนะนำสถาปัตยกรรมการทำ Unit Test, เทคนิคการจำลองข้อมูลขั้นสูง (Mocking), คำสั่งมาตรฐานในการทดสอบ และการใช้งานแดชบอร์ดสรุปผลการทดสอบและความครอบคลุมของโค้ดแบบโต้ตอบในสไตล์ JMeter ระดับพรีเมียม

---

## 📂 Directory Layout / โครงสร้างโฟลเดอร์ทดสอบ

All testing utilities, reports, and execution guides are consolidated inside the `test/` directory to maintain a clean root workspace:
เครื่องมือรันรายงาน, โฟลเดอร์ผลลัพธ์ และคู่มือทั้งหมดถูกรวมไว้ในโฟลเดอร์ `test/` เพื่อแยกออกจากซอร์สโค้ดหลักอย่างเป็นสัดส่วน:

```
Backend/test/
├── README.md               # [THIS FILE] Bilingual developer guide / เอกสารคู่มือชุดนี้
├── run_report.sh           # Main executable runner for tests + dashboard / สคริปต์รันการทดสอบและสรุปผล
├── reports/                # Destination directory for test artifacts / โฟลเดอร์เก็บผลลัพธ์
│   ├── test_report.html    # Standalone premium interactive dashboard / แดชบอร์ดสรุปผลแบบโต้ตอบระดับพรีเมียม
│   └── coverage.out        # Raw Go statement coverage profile / ข้อมูลความครอบคลุมดิบระดับ Statement
└── scripts/
    └── generate_report.go  # Dashboard compilation and parser / สคริปต์ประมวลผลและสร้างรายงาน HTML
```

---

## 1. 🏗️ Go Testing Philosophy & Architecture / สถาปัตยกรรมและการทำ Unit Test ใน Go

### 🇺🇸 Package-Level Co-location
In Go, best practices dictate that test files (`*_test.go`) must live in the **same directory (package)** as the production code they are testing (e.g. `Backend/service/`).
* **White-box Testing:** Placing tests inside the same package permits direct access to unexported (private) functions, structures, and variables. This allows us to test internal package logic without exposing private implementation details to external importers.
* **Separation of Concerns:** While the files sit together, they are excluded from production builds. Only files ending with `_test.go` are compiled when running `go test`.

### 🇹🇭 การจัดวางไฟล์ Unit Test ในแพ็คเกจ (Package-Level Co-location)
ตามแนวทางปฏิบัติที่ดีที่สุด (Best Practices) ของภาษา Go ไฟล์ทดสอบ (`*_test.go`) จะต้องจัดวางอยู่ใน **โฟลเดอร์ (แพ็คเกจ) เดียวกัน** กับซอร์สโค้ดผลิตจริง (เช่น `Backend/service/`) ด้วยเหตุผลหลักดังนี้:
* **การทดสอบแบบกล่องขาว (White-box Testing):** การวางไฟล์ทดสอบในแพ็คเกจเดียวกันทำให้สามารถเข้าถึงตัวแปร, โครงสร้างข้อมูล (structs), และฟังก์ชันภายในที่เป็นแบบปิด (unexported / private) ได้โดยตรง ช่วยให้ตรวจสอบตรรกะภายในได้โดยไม่ต้องเปิดเผยรายละเอียดเหล่านั้นให้แพ็คเกจภายนอกมองเห็น
* **การแยกขอบเขตงาน:** ถึงแม้ไฟล์ทดสอบจะวางอยู่ร่วมโฟลเดอร์เดียวกัน แต่ระบบคอมไพเลอร์ของ Go จะแยกไฟล์ที่ลงท้ายด้วย `_test.go` ออกโดยอัตโนมัติเมื่อทำการสร้างแอปพลิเคชันเพื่อส่งมอบงานจริง (Production build) และจะนำมาคอมไพล์และทำงานเฉพาะตอนรันคำสั่ง `go test` เท่านั้น

---

## 2. ⚡ Advanced Mocking & Interception Techniques / ยุทธวิธีการจำลองข้อมูลขั้นสูง

To achieve **90-100% statement coverage** safely without relying on active databases or live third-party network APIs, we implemented several advanced mocking architectures inside the service tests:

เพื่อบรรลุ **ความครอบคลุมบรรทัดคำสั่ง (Statement Coverage) ในระดับสูงถึง 90-100%** อย่างปลอดภัยโดยไม่ต้องพึ่งพาระบบฐานข้อมูลจริงหรือการเชื่อมต่อเครือข่าย API ภายนอก เราจึงได้ออกแบบสถาปัตยกรรมการจำลองข้อมูลขั้นสูงในส่วนต่าง ๆ ดังนี้:

### 🌐 A. HTTP RoundTripper Interception (Gemini OCR Mocking)
In `gemini_ocr_test.go`, instead of calling the live Google Gemini Vision APIs (which requires API keys and network latency), we intercepted outbound HTTP requests at the standard library transport layer:
* **Mechanism:** We wrapped `http.DefaultClient.Transport` using a custom struct implementing the `http.RoundTripper` interface.
* **Verification:** The interceptor captures the outgoing HTTP request, decodes the multipart form-data (containing the uploaded file and prompt metadata), and asserts that proper parameters are sent.
* **Stubbed Responses:** It responds with custom, hardcoded HTTP responses (mocking both successful OCR parses and error conditions). This allows testing of `callGeminiVision` parsing failures, JSON unmarshalling errors, and bad status codes with **100% test reliability and sub-millisecond execution times**.

### 🇹🇭 การควบคุม HTTP Transport (RoundTripper Interception) ใน Gemini OCR
ในไฟล์ `gemini_ocr_test.go` เพื่อหลีกเลี่ยงการเรียกใช้งาน Google Gemini Vision API จริง (ซึ่งต้องการ API Key และมีความหน่วงเวลาเครือข่าย) เราได้ดักจับ HTTP ขาออกที่ส่งออกจาก Standard Library:
* **กลไกการทำงาน:** ทำการครอบ `http.DefaultClient.Transport` ด้วยโครงสร้างข้อมูลที่พัฒนาขึ้นเองเพื่อตอบสนองต่ออินเตอร์เฟส `http.RoundTripper`
* **การตรวจสอบความถูกต้อง:** ระบบดักจับจะสกัดโครงสร้าง Multipart Form-data ของคำขอ HTTP เพื่อเช็คความถูกต้องของไฟล์รูปภาพและ Prompt คำสั่งที่ส่งไป
* **การจำลองผลตอบกลับ (Stubbing):** ส่งกลับ HTTP Response ที่เตรียมไว้ล่วงหน้า (เช่น ผลการสแกน OCR สำเร็จ หรือกรณีโครงสร้างข้อมูลผิดเพี้ยน) ทำให้สามารถทดสอบกรณีเกิดข้อผิดพลาดของ JSON parser หรือกรณีได้ HTTP Status Code ล้มเหลวได้อย่างรวดเร็วและแม่นยำ 100% โดยใช้เวลาไม่ถึง 1 มิลลิวินาที

---

### 🔀 B. Asynchronous Channel Stream Testing (Orchestrator Service)
The `orchestrator_service.go` processes parallel subsidy allocations, publishing updates to internal Go channels (`Events` subscriber streams).
* **Buffer Isolation:** Tests utilize dedicated Go routines and sync mechanisms (`sync.WaitGroup` or short time-controlled sleep buffers) to publish and consume statuses in parallel.
* **Race Condition Avoidance:** Mocks simulate both single-threaded decision logic and high-concurrency event loops to confirm subscriber channels close properly, avoiding memory leaks and resource locks during application shutdown.

### 🇹🇭 การทดสอบช่องทางส่งข้อมูลแบบอะซิงโครนัส (Asynchronous Channels ใน Orchestrator)
ไฟล์ `orchestrator_service.go` มีกลไกกระจายงานคัดกรองทุนสนับสนุนแบบขนาน โดยส่งผลลัพธ์ผ่านทาง Go Channels (`Events` subscriber streams):
* **การจัดการบัฟเฟอร์คู่ขนาน:** ชุดทดสอบใช้ Go Routine ควบคู่ไปกับกลไกซิงโครไนซ์อย่าง `sync.WaitGroup` หรือบัฟเฟอร์หน่วงเวลาสั้น ๆ เพื่อช่วยทดสอบการรับส่งข้อมูลและการส่งสัญญาณแบบขนาน
* **การป้องกันสภาวะแข่งขัน (Race Conditions):** มีการจำลองสถานการณ์ทั้งแบบ Single-thread และแบบ Concurrency สูง เพื่อให้มั่นใจว่าช่องทางผู้ติดตาม (Subscriber Channels) จะถูกปิดตัวลงอย่างถูกต้อง ไม่ทำให้เกิด Memory Leaks หรือสถานะติดล็อกค้าง (Deadlocks)

---

## 3. ⚙️ Basic Go Testing Commands / คำสั่งการทดสอบพื้นฐาน

Developers can execute standard tests manually from the `Backend/` workspace using the following commands:
นักพัฒนาสามารถรันงานทดสอบพื้นฐานด้วยตนเองผ่าน Terminal จากโฟลเดอร์ `Backend/` โดยใช้คำสั่งเหล่านี้:

* **Run all tests (Silent Mode) / รันผลการทดสอบทั้งหมดแบบกระชับ:**
  ```bash
  go test ./service/...
  ```

* **Run all tests with verbose logs / รันผลการทดสอบทั้งหมดพร้อมแสดงขั้นตอนอย่างละเอียด:**
  ```bash
  go test -v ./service/...
  ```

* **Generate coverage profile / ส่งออกสถิติความครอบคลุมดิบ (coverage.out):**
  ```bash
  go test -coverprofile=test/reports/coverage.out ./service/...
  ```

* **View coverage in default Go HTML viewer / ตรวจดูความครอบคลุมบน HTML พื้นฐานของ Go:**
  ```bash
  go tool cover -html=test/reports/coverage.out
  ```

---

## 4. 📊 Premium JMeter-Style HTML Dashboard / แดชบอร์ดสรุปผลระดับพรีเมียม

For an executive-level overview of our test suites, we built a premium, standalone, and completely offline-capable **JMeter-Style Test & Coverage Dashboard**.

สำหรับมุมมองรายงานในระดับผู้บริหารและผู้ควบคุมงาน เราได้สร้างระบบรายงานสรุปผลพร้อมสถิติอย่างละเอียดในรูปแบบ **JMeter-Style Test & Coverage Dashboard** ซึ่งทำงานแบบออฟไลน์ได้ 100%

### 🚀 How to Run the Dashboard / วิธีการรันระบบสรุปผลรายงาน
Run the single executable runner script from the `Backend/` directory:
รันสคริปต์ควบคุมการทำงานในคำสั่งเดียวจากโฟลเดอร์ `Backend/`:

```bash
chmod +x test/run_report.sh
./test/run_report.sh
```

This script will automatically:
1. Compile and execute `go test -json -coverprofile=test/reports/coverage.out ./service/...`.
2. Analyze timings (min, max, average) for all 71 unit tests.
3. Parse statement-level block coverage and query function names using `go tool cover`.
4. Compile and output a stunning glassmorphic HTML report file at **`Backend/test/reports/test_report.html`**.

สคริปต์นี้จะทำงานเหล่านี้โดยอัตโนมัติ:
1. คอมไพล์และเรียกทดสอบระบบในรูปแบบ JSON พร้อมเก็บค่าความครอบคลุมไว้ที่ `test/reports/coverage.out`
2. คำนวณความเร็วในการประมวลผล (ต่ำสุด, สูงสุด, และความเร็วเฉลี่ย) สำหรับกรณีทดสอบทั้ง 71 ชุด
3. ตรวจประเมินความครอบคลุมราย Statement พร้อมดึงชื่อฟังก์ชันที่เกี่ยวข้องด้วย `go tool cover`
4. รวบรวมข้อมูลทั้งหมดแล้วคอมไพล์ออกมาเป็นหน้าเว็บแดชบอร์ดสไตล์กระจกโปร่งแสง (Glassmorphic) ที่สวยงามล้ำสมัย ณ ที่เก็บ **`Backend/test/reports/test_report.html`**

---

### 📈 Core Metrics Explained / รายละเอียดตัวชี้วัดสำคัญ

The dashboard blending functional correctness and statement coverage depth into three distinct metrics:
แดชบอร์ดสรุปประสิทธิภาพแบบรวมศูนย์ที่ประเมินทั้งคุณภาพผลการรันและสถิติครอบคลุมโค้ดผ่าน 3 ดัชนีหลัก:

1. **Pass Rate (อัตราการผ่าน):** 
   * **Formula:** `(Passed Tests / Total Tests) * 100`
   * **Significance:** Measures functional correctness. (Target: **100%**).
   * **ความสำคัญ:** แสดงความถูกต้องของระบบ โดยชุดการทดสอบทั้งหมด 71 ชุดต้องผ่านการรับรองทั้งหมด (เป้าหมาย: **100%**).

2. **Statement Coverage % (สัดส่วนความครอบคลุมบรรทัดคำสั่ง):** 
   * **Formula:** `(Covered Statements / Total Statements) * 100`
   * **Significance:** Measures test completeness. Ensure that critical service layers have no hidden logic unchecked. (Target: **90-100%**).
   * **ความสำคัญ:** แสดงถึงความครอบคลุมในการตรวจสอบโค้ด บ่งบอกว่าไม่มีช่องโหว่ใดในเซิร์วิสที่เล็ดลอดสายตาไปได้ (เป้าหมาย: **90-100%**).

3. **Apdex Score (คะแนนดัชนีคุณภาพความครอบคลุมแอปพลิเคชัน):** 
   * **Formula:** `Pass Rate * (Coverage % / 100)`
   * **Significance:** Blends functional success and coverage depth. An application with 100% passing tests but only 50% coverage receives an Apdex of 50.0%. To achieve a premium Apdex of **90.0% or above**, developers must maintain both 100% functional pass rate and over 90.0% statement coverage.
   * **ความสำคัญ:** คะแนนผสมผสานระหว่างอัตราความสำเร็จและระดับความครอบคลุม หากแอปพลิเคชันรันผ่าน 100% แต่โค้ดครอบคลุมเพียง 50% จะได้คะแนน Apdex เพียง 50.0% เท่านั้น การได้เกรด Apdex คุณภาพระดับยอดเยี่ยม (**90.0%+**) บ่งชี้ว่าระบบมีอัตราการรันผ่านครบถ้วนและมีสถิติความครอบคลุมโค้ดสูงเกิน 90% ด้วยเช่นกัน

---

## 🎨 Premium UI Features in `test_report.html` / คุณสมบัติการใช้งานแดชบอร์ดโต้ตอบ

* **Overall Quality Gauge:** An animated circular SVG progress gauge rendering the global Apdex score, color-coded dynamically (Green: Excellent >=90%, Blue: Tolerated >=80%, Red: Poor <80%).
* **Coverage Heatmap:** Clear visual table detailing statements covered per service, displaying status pills showing progress.
* **Bespoke Function Breakdown:** Click on any service row (e.g. `auth_service.go`) to expand and inspect individual functions, showing their exact declaration line and statement coverage.
* **Instant Logs Filtering & Clipboard:** An interactive terminal logger in dark mode, allowing developers to filter by test names or actions (`PASS`, `FAIL`) in real-time, accompanied by a single-click "Copy Logs" clipboard tool.
* **Fully Portable & Offline Ready:** Includes built-in visual rendering assets and styles. No network connections, internet CDNs (such as Tailwind or chart libraries) are utilized, ensuring complete operational privacy and security inside disconnected air-gapped development environments.

* **เกจแสดงผลลัพธ์คุณภาพสูงสุด (Overall Quality Gauge):** เกจวงกลม SVG ที่มีแอนิเมชันคำนวณคะแนน Apdex โดยเปลี่ยนโทนสีตามระดับคะแนนโดยอัตโนมัติ (สีเขียว: ยอดเยี่ยม >=90%, สีฟ้า: ยอมรับได้ >=80%, สีแดง: ต้องปรับปรุง <80%)
* **แผนภาพความครอบคลุมแบบ Heatmap:** รายการตารางแสดงความครอบคลุมของ Statement ในแต่ละฟิลเตอร์ย่อยของเซิร์วิสพร้อมระดับความสมบูรณ์เป็นป้ายสีสถานะ
* **ตารางแจกแจงระดับฟังก์ชัน (Function-level Breakdown):** สามารถคลิกแถวใด ๆ ในตารางไฟล์ (เช่น `auth_service.go`) เพื่อขยายดูรายชื่อฟังก์ชัน บรรทัดประกาศฟังก์ชัน และอัตราการครอบคลุมภายในของฟังก์ชันนั้น ๆ ได้ทันที
* **ตัวกรองประวัติการรันและคัดลอกประวัติ (Logs Filter & Copy):** คอนโซลจำลองแบบมืดที่มีตัวกรองข้อความเพื่อค้นหาชื่อเคสทดสอบหรือสถานะผลการทำงานแบบเรียลไทม์ และปุ่มกดยกยอดคัดลอก Logs ลง Clipboard ในคลิกเดียว
* **สามารถใช้งานออฟไลน์สมบูรณ์แบบ:** ตัวหน้าเว็บฝังตรรกะประมวลผลและการจัดรูปแบบทั้งหมดไว้ภายในหน้าเดียว โดยไม่เรียกใช้สคริปต์ภายนอกหรือ CDN (เช่น Tailwind หรือไลบรารีกราฟิก) ช่วยให้ใช้งานได้บนทุกเครื่องคอมพิวเตอร์อย่างปลอดภัยและเป็นส่วนตัวสูงสุด แม้จะไม่ได้เชื่อมต่ออินเทอร์เน็ตก็ตาม
