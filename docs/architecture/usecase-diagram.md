```mermaid
flowchart LR
    %% Actors Definition
    Citizen((ประชาชน<br/>Citizen))
    Officer((เจ้าหน้าที่รัฐ<br/>Officer))
    Admin((ผู้ดูแลระบบ<br/>Admin))

    subgraph Platform["ระบบจัดการสวัสดิการภาครัฐ (Government Subsidy Platform)"]
        %% Citizen Use Cases
        subgraph Citizen_Actions["ส่วนของประชาชน"]
            UC1(["UC1: ลงทะเบียน & ยืนยันตัวตน eKYC"])
            UC2(["UC2: ยื่นคำขอรับสิทธิ์สวัสดิการ"])
            UC3(["UC3: ติดตามสถานะคำขอแบบ Real-time"])
        end

        %% Admin Actions
        subgraph Admin_Actions["ส่วนของผู้ดูแลระบบ"]
            UC6(["UC6: สร้างโครงการและกำหนดเกณฑ์สิทธิ์"])
            UC7(["UC7: ตรวจสอบ Audit Log และรายงาน"])
        end

        %% Officer Actions (New Automated Focus)
        subgraph Officer_Actions["ส่วนของเจ้าหน้าที่ระดับนโยบาย"]
            UC9(["UC9: อนุมัติโครงการสวัสดิการ<br/>Project Approval"])
        end

        %% Internal System Process
        UC8(["UC8: ตัดสินสิทธิ์อัตโนมัติ 100%<br/>Auto Decision Engine"])
    end

    %% Citizen Connections
    Citizen --- UC1
    Citizen --- UC2
    Citizen --- UC3

    %% Admin Connections
    Admin --- UC6
    Admin --- UC7

    %% Officer Connections
    Officer --- UC9

    %% Logical Relationships
    UC2 -.->|include| UC1
    UC2 -.->|triggered| UC8
    UC9 -.->|authorize| UC6

    %% Styling
    style Citizen fill:#0d9488,stroke:#fff,color:#fff
    style Officer fill:#0369a1,stroke:#fff,color:#fff
    style Admin fill:#7c3aed,stroke:#fff,color:#fff
    
    style UC1 fill:#f0fdfa,stroke:#0d9488
    style UC2 fill:#f0fdfa,stroke:#0d9488
    style UC3 fill:#f0fdfa,stroke:#0d9488
    
    style UC6 fill:#f5f3ff,stroke:#7c3aed
    style UC7 fill:#f5f3ff,stroke:#7c3aed
    
    style UC9 fill:#eff6ff,stroke:#0369a1
    
    style UC8 fill:#fff7ed,stroke:#ea580c,stroke-dasharray: 5 5
```