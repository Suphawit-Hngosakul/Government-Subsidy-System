# Branching Strategy

เอกสารนี้ใช้เป็นมาตรฐานการแตก branch ของโปรเจกต์ Government Subsidy System เพื่อให้ทุกคนในทีมทำงานไปในทิศทางเดียวกัน ลดการชนกันของโค้ด และทำให้การ promote งานจากฝั่งพัฒนาไปสู่สภาพแวดล้อมทดสอบและ production เป็นระบบ

## วัตถุประสงค์

- แยกสภาพแวดล้อมการทำงานให้ชัดเจน
- ทำให้การ review และ merge มีลำดับที่เข้าใจง่าย
- ลดความเสี่ยงจากการ push ตรงเข้า branch สำคัญ
- ทำให้สมาชิกใหม่ในทีมเข้าใจ workflow ได้เร็ว

## โครงสร้าง Branch หลัก

โปรเจกต์นี้ใช้ branch หลักดังนี้

- `main`
- `prod`
- `uat`
- `sit`
- `develop`

ลำดับความหมายของ branch คือ:

1. `develop` ใช้รวมงาน feature ที่พัฒนาเสร็จในระดับทีมพัฒนา
2. `sit` ใช้สำหรับรวมงานเพื่อทดสอบภาพรวมของระบบร่วมกัน
3. `uat` ใช้สำหรับทดสอบกับมุมมองผู้ใช้งานหรือผู้เกี่ยวข้องทางธุรกิจ
4. `prod` ใช้เป็น branch เตรียมปล่อยใช้งานจริง
5. `main` ใช้เป็น branch อ้างอิงหลักของโปรเจกต์ และควรสะท้อน baseline ที่นิ่งที่สุด

## ลำดับการแตก Branch เริ่มต้น

ก่อนเริ่มงาน ให้สร้าง branch ตามลำดับนี้จาก `main`

```bash
git checkout main
git pull origin main

git checkout -b prod
git push -u origin prod

git checkout -b uat
git push -u origin uat

git checkout -b sit
git push -u origin sit

git checkout -b develop
git push -u origin develop
```

แนวคิดสำคัญคือ:

- `prod` แตกจาก `main`
- `uat` แตกจาก `prod`
- `sit` แตกจาก `uat`
- `develop` แตกจาก `sit`

ถ้าทีมสร้าง branch หลักครบแล้ว ไม่ต้องสร้างซ้ำ ให้ใช้งาน branch เดิมต่อได้เลย

## การแตก Branch สำหรับงานของแต่ละคน

เมื่อ branch กลางพร้อมแล้ว สมาชิกแต่ละคนต้องแตก branch งานของตัวเองจาก `develop` เท่านั้น

ห้ามแตก feature branch จาก `main`, `prod`, `uat`, หรือ `sit` โดยตรง เว้นแต่ทีมตกลงร่วมกันในกรณีพิเศษ เช่น hotfix production

ตัวอย่าง:

```bash
git checkout develop
git pull origin develop

git checkout -b feat/auth-register
git push -u origin feat/auth-register
```

## รูปแบบการตั้งชื่อ Branch

แนะนำให้ตั้งชื่อ branch ให้สื่อความหมายและอ่านแล้วรู้ทันทีว่าทำอะไร

รูปแบบที่แนะนำ:

- `feat/<feature-name>` สำหรับฟีเจอร์ใหม่
- `fix/<issue-name>` สำหรับแก้ bug
- `hotfix/<issue-name>` สำหรับแก้ปัญหาเร่งด่วน
- `chore/<task-name>` สำหรับงานดูแลระบบหรือ config
- `docs/<document-name>` สำหรับงานเอกสาร
- `refactor/<scope-name>` สำหรับปรับโครงสร้างโค้ด
- `test/<scope-name>` สำหรับเพิ่มหรือปรับชุดทดสอบ

ตัวอย่างชื่อ branch ที่ดี:

- `feat/auth-register`
- `feat/benefit-claim`
- `feat/dopa-mock-api`
- `fix/login-validation`
- `docs/branching-strategy`

## Workflow ที่สมาชิกในทีมต้องใช้ทุกครั้ง

1. อัปเดต `develop` ให้ล่าสุดก่อนเริ่มงาน
2. แตก branch งานของตัวเองจาก `develop`
3. เขียนโค้ดและ commit เป็นช่วง ๆ ให้ข้อความ commit ชัดเจน
4. push branch ขึ้น remote
5. เปิด Pull Request เข้า `develop`
6. ให้เพื่อนในทีม review ก่อน merge
7. เมื่อรวมหลาย feature แล้ว ค่อย promote `develop` ไป `sit`
8. หลังผ่าน SIT แล้ว ค่อย promote `sit` ไป `uat`
9. หลังผ่าน UAT แล้ว ค่อย promote `uat` ไป `prod`
10. เมื่อพร้อม release จริง ค่อย merge `prod` เข้า `main`

## ภาพรวมการไหลของงาน

```text
main
  -> prod
     -> uat
        -> sit
           -> develop
              -> feat/... , fix/... , docs/... , chore/...
```

เมื่องานเสร็จ การไหลของการ merge จะย้อนกลับขึ้นไปตามลำดับ:

```text
feat/* -> develop -> sit -> uat -> prod -> main
```

## ตัวอย่างการทำงานจริง

สมมติสมาชิกชื่อ A รับผิดชอบระบบสมัครสมาชิก

```bash
git checkout develop
git pull origin develop
git checkout -b feat/auth-register
```

ทำงานเสร็จแล้ว:

```bash
git add .
git commit -m "feat: add citizen registration endpoint"
git push -u origin feat/auth-register
```

จากนั้นเปิด Pull Request จาก `feat/auth-register` เข้า `develop`

## กติกาในการ Merge

- ห้าม push ตรงเข้า `main`
- ห้าม push ตรงเข้า `prod` หากไม่ผ่านขั้นตอนจาก branch ก่อนหน้า
- ควรใช้ Pull Request ในการ merge ทุกครั้ง
- ทุก PR ควรมีอย่างน้อย 1 reviewer ถ้า workflow ของทีมรองรับ
- ควรทดสอบก่อน merge เสมอ
- ถ้า branch มี conflict ให้คนที่ทำ branch นั้นเป็นผู้ rebase หรือ merge latest `develop` เข้ามาแก้ก่อน

## แนวทาง Commit Message

แนะนำให้ใช้รูปแบบที่สม่ำเสมอ เช่น:

- `feat: add benefit claim endpoint`
- `fix: correct national id validation`
- `docs: add team branching strategy`
- `chore: update frontend project config`

## กรณี Hotfix

ถ้าเกิดปัญหาเร่งด่วนใน production:

1. แตก branch จาก `prod` โดยใช้ชื่อ `hotfix/...`
2. แก้ปัญหาและทดสอบ
3. merge กลับเข้า `prod`
4. จากนั้นต้อง sync การแก้ไขกลับลง `uat`, `sit`, และ `develop` ด้วย เพื่อไม่ให้โค้ดคนละชุด

ตัวอย่าง:

```bash
git checkout prod
git pull origin prod
git checkout -b hotfix/login-null-check
```

## สิ่งที่ควรหลีกเลี่ยง

- แตก feature branch จาก branch ของเพื่อน
- ทำหลาย feature ใน branch เดียว
- merge ข้ามลำดับ เช่น merge `develop` เข้า `prod` ตรงโดยไม่ผ่าน `sit` และ `uat`
- commit ใหญ่เกินไปจน review ยาก
- ใช้ชื่อ branch กว้างเกินไป เช่น `feat/update`

## ข้อแนะนำสำหรับทีม

- ให้ทุกคน sync `develop` ก่อนเริ่มงานทุกวัน
- ให้ตั้ง branch protection สำหรับ `main`, `prod`, `uat`, `sit`, และ `develop`
- ให้ใช้ PR template และ checklist ถ้าทีมเริ่มโตขึ้น
- ให้แยกงานเป็น feature branch เล็ก ๆ เพื่อ review ได้ง่าย

## สรุปสั้น

จำง่ายที่สุดคือ:

`main -> prod -> uat -> sit -> develop -> feat/*`

และเวลา merge ให้ไหลกลับขึ้นมาตามลำดับ:

`feat/* -> develop -> sit -> uat -> prod -> main`

ถ้าทุกคนยึด flow นี้เหมือนกัน ทีมจะทำงานร่วมกันได้ง่ายขึ้นมาก และลดปัญหา branch สับสนในระยะยาว
