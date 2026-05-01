/*
 * ============================================================================
 *  Government Subsidy System — C4 Architecture Model (Structurizr DSL)
 * ============================================================================
 *  Version  : 2.0
 *  Date     : 2026-04-28
 *  Author   : Architecture Team
 *  Scope    : System Context, Container, Component, and Deployment views
 *
 *  This model represents the architecture of the Thai Government Subsidy
 *  System, designed as a modular monolith with clear domain boundaries
 *  and scalability-ready data layer separation.
 * ============================================================================
 */

workspace "Government Subsidy System" "ระบบสวัสดิการภาครัฐ — สถาปัตยกรรมระดับ Production-Ready สำหรับการจัดการสิทธิ์สวัสดิการประชาชน" {

    !identifiers hierarchical

    model {

        /* ================================================================
         *  ACTORS / PEOPLE
         * ================================================================ */

        citizen = person "ประชาชน (Citizen)" "ผู้ใช้งานระบบสวัสดิการ ลงทะเบียน ยืนยันตัวตน และยื่นขอรับสิทธิ์สวัสดิการจากภาครัฐ" "Citizen"
        officer = person "เจ้าหน้าที่ตรวจสอบ (Review Officer)" "เจ้าหน้าที่ภาครัฐผู้ตรวจสอบคำขอสวัสดิการและอนุมัติ/ปฏิเสธสิทธิ์" "Officer"
        admin = person "ผู้ดูแลระบบ (System Administrator)" "ผู้ดูแลระบบ จัดการโครงการสวัสดิการ และตรวจสอบ Audit Log" "Admin"

        /* ================================================================
         *  EXTERNAL SYSTEMS (Mock Providers)
         * ================================================================ */

        group "External / Mock Providers" {
            dopaApi = softwareSystem "DOPA Mock API" "กรมการปกครอง — ตรวจสอบข้อมูลทะเบียนราษฎร ยืนยันตัวตนประชาชนด้วยเลขบัตรประจำตัวประชาชน" "ExternalSystem"
            ssoApi = softwareSystem "SSO Mock API" "สำนักงานประกันสังคม — ตรวจสอบสถานะการประกันตน สิทธิ์ประกันสังคม และข้อมูลนายจ้าง" "ExternalSystem"
            ktbApi = softwareSystem "KTB Mock API" "ธนาคารกรุงไทย — ตรวจสอบสถานะทางการเงินและข้อมูลบัญชีธนาคารของผู้ขอรับสิทธิ์" "ExternalSystem"
        }

        /* ================================================================
         *  MAIN SOFTWARE SYSTEM
         * ================================================================ */

        govSubsidy = softwareSystem "Government Subsidy System" "ระบบจัดการสวัสดิการภาครัฐแบบครบวงจร รองรับการลงทะเบียน ยืนยันตัวตน (eKYC) ยื่นคำขอสิทธิ์ ตรวจสอบคุณสมบัติอัตโนมัติ และการตรวจสอบโดยเจ้าหน้าที่" "GovernmentSystem" {

            /* ------------------------------------------------------------
             *  FRONTEND
             * ------------------------------------------------------------ */

            frontend = container "Frontend Application" "เว็บแอปพลิเคชันสำหรับประชาชนและเจ้าหน้าที่ รองรับ Responsive Design และ Real-time Status Updates" "React / Expo Router" "WebApp"

            /* ------------------------------------------------------------
             *  API GATEWAY
             * ------------------------------------------------------------ */

            apiGateway = container "API Gateway" "จุดเชื่อมต่อกลาง จัดการ Routing, Rate Limiting, CORS, TLS Termination, และ Request Authentication" "Nginx / Reverse Proxy" "Gateway"

            /* ------------------------------------------------------------
             *  BACKEND — Go Modular Monolith
             * ------------------------------------------------------------ */

            backend = container "Government Subsidy Backend" "Backend หลักของระบบ พัฒนาด้วย Go แบบ Modular Monolith แยก Domain Boundaries ชัดเจน" "Go" "Backend" {

                /* ---- Business Modules ---- */

                group "Business Modules" {
                    claimMgmt = component "Claim Management" "จัดการคำขอสิทธิ์สวัสดิการ: สร้าง ติดตามสถานะ บันทึกผลการตรวจสอบ และแสดงประวัติคำขอ" "Go Module"
                    authEkyc = component "Auth & eKYC" "การลงทะเบียน เข้าสู่ระบบ ออกจากระบบ JWT Token Management, RBAC และยืนยันตัวตนด้วย eKYC" "Go Module"
                    projectMgmt = component "Project Management" "จัดการโครงการสวัสดิการ: สร้าง แก้ไข กำหนดเงื่อนไขคุณสมบัติ เปิด/ปิดรับสมัคร" "Go Module"
                    officerReview = component "Officer Review" "ระบบตรวจสอบคำขอโดยเจ้าหน้าที่ แสดงรายละเอียดผลตรวจ รองรับ Manual Approve/Reject" "Go Module"
                    adminDashboard = component "Admin Dashboard" "แดชบอร์ดสำหรับผู้ดูแลระบบ แสดงสถิติ จัดการโครงการ และเข้าถึง Audit Log" "Go Module"
                    statusStream = component "Status Stream (SSE)" "Server-Sent Events สำหรับแจ้งสถานะคำขอแบบ Real-time ให้ประชาชนและเจ้าหน้าที่" "Go Module"
                }

                /* ---- Async Processing ---- */

                group "Async Processing" {
                    jobQueue = component "Job Queue" "คิวงานสำหรับจัดการ Asynchronous Tasks รองรับ Priority Queue และ Retry Mechanism" "Go + Redis"
                    orchestratorWorker = component "Orchestrator Worker" "Background Worker ที่ประสานงานเรียก External APIs พร้อมกัน จัดการ Timeout และ Retry" "Go Worker"
                    decisionEngine = component "Decision Engine" "เอนจินตัดสินใจอัตโนมัติ ประเมินคุณสมบัติจากข้อมูลทุกแหล่ง ตัดสิน Approve/Reject/Pending" "Go Module"
                }

                /* ---- Integration Adapters ---- */

                group "Integration Adapters" {
                    dopaAdapter = component "DOPA Adapter" "Adapter สำหรับเชื่อมต่อ DOPA API แปลง Schema เป็น IdentityCheckResult พร้อม Timeout/Retry" "Go Adapter"
                    ssoAdapter = component "SSO Adapter" "Adapter สำหรับเชื่อมต่อ SSO API แปลง Schema เป็น InsuranceStatusResult พร้อม Timeout/Retry" "Go Adapter"
                    ktbAdapter = component "KTB Adapter" "Adapter สำหรับเชื่อมต่อ KTB API แปลง Schema เป็น FinancialStatusResult พร้อม Timeout/Retry" "Go Adapter"
                }
            }

            /* ------------------------------------------------------------
             *  DATA LAYER — Separated for Scalability
             * ------------------------------------------------------------ */

            group "Primary Data Store" {
                postgresWrite = container "PostgreSQL — Primary (Write)" "ฐานข้อมูลหลักสำหรับ Write Operations: claims, users, projects, eligibility results" "PostgreSQL 16" "DatabasePrimary"
                postgresRead = container "PostgreSQL — Read Replica" "Read Replica สำหรับ Query-heavy Operations: Dashboard, Reports, Officer Review Queries" "PostgreSQL 16 (Streaming Replication)" "DatabaseReplica"
            }

            group "Audit & Compliance Store" {
                auditLog = container "Audit Log Database" "ฐานข้อมูล Append-only สำหรับ Audit Trail เก็บทุก Action ที่เกิดขึ้นในระบบ รองรับ Compliance และ Forensics" "PostgreSQL 16 (Append-Only)" "DatabaseAudit"
            }

            group "Caching & Queue Layer" {
                redisCache = container "Redis — Cache" "Cache Layer สำหรับ Session Management, Rate Limiting, Token Blacklist และ Frequently Accessed Data" "Redis 7 (Cluster Mode)" "Cache"
                redisQueue = container "Redis — Queue" "Message Queue สำหรับ Async Job Processing: Claim Orchestration, Notification, และ Batch Tasks" "Redis 7 (Streams)" "Queue"
            }

            group "File Storage" {
                objectStorage = container "Object Storage / File Store" "จัดเก็บไฟล์ eKYC Documents, รูปบัตรประชาชน และเอกสารประกอบคำขอ รองรับ Encryption at Rest" "MinIO / S3-Compatible" "FileStore"
            }
        }

        /* ================================================================
         *  RELATIONSHIPS — System Context Level
         * ================================================================ */

        citizen -> govSubsidy "ลงทะเบียน ยืนยันตัวตน และยื่นขอสิทธิ์สวัสดิการ" "HTTPS"
        officer -> govSubsidy "ตรวจสอบและอนุมัติ/ปฏิเสธคำขอสวัสดิการ" "HTTPS"
        admin -> govSubsidy "จัดการโครงการสวัสดิการและตรวจสอบ Audit Log" "HTTPS"

        govSubsidy -> dopaApi "ตรวจสอบข้อมูลทะเบียนราษฎร" "HTTPS / REST"
        govSubsidy -> ssoApi "ตรวจสอบสถานะประกันสังคม" "HTTPS / REST"
        govSubsidy -> ktbApi "ตรวจสอบข้อมูลทางการเงิน" "HTTPS / REST"

        /* ================================================================
         *  RELATIONSHIPS — Container Level
         * ================================================================ */

        /* Users → Frontend */
        citizen -> govSubsidy.frontend "เข้าใช้งานระบบสวัสดิการ" "HTTPS"
        officer -> govSubsidy.frontend "เข้าถึงระบบตรวจสอบคำขอ" "HTTPS"
        admin -> govSubsidy.frontend "เข้าถึง Admin Dashboard" "HTTPS"

        /* Frontend → Gateway → Backend */
        govSubsidy.frontend -> govSubsidy.apiGateway "ส่ง API Requests" "HTTPS / REST"
        govSubsidy.apiGateway -> govSubsidy.backend "Route ไปยัง Backend Services" "HTTP / REST"

        /* Backend → Data Layer */
        govSubsidy.backend -> govSubsidy.postgresWrite "อ่าน/เขียนข้อมูลหลัก (Claims, Users, Projects)" "TCP/5432"
        govSubsidy.backend -> govSubsidy.postgresRead "อ่านข้อมูลสำหรับ Dashboard และ Reports" "TCP/5432"
        govSubsidy.backend -> govSubsidy.auditLog "บันทึก Audit Trail (Append-Only)" "TCP/5432"
        govSubsidy.backend -> govSubsidy.redisCache "Cache, Session, Rate Limiting" "TCP/6379"
        govSubsidy.backend -> govSubsidy.redisQueue "Enqueue Async Jobs" "TCP/6379"
        govSubsidy.backend -> govSubsidy.objectStorage "จัดเก็บและดึงไฟล์ eKYC" "S3 API"

        /* Replication */
        govSubsidy.postgresWrite -> govSubsidy.postgresRead "Streaming Replication" "PostgreSQL Replication Protocol"

        /* Backend → External Systems */
        govSubsidy.backend -> dopaApi "ตรวจสอบทะเบียนราษฎร (ผ่าน DOPA Adapter)" "HTTPS / REST"
        govSubsidy.backend -> ssoApi "ตรวจสอบประกันสังคม (ผ่าน SSO Adapter)" "HTTPS / REST"
        govSubsidy.backend -> ktbApi "ตรวจสอบข้อมูลการเงิน (ผ่าน KTB Adapter)" "HTTPS / REST"

        /* ================================================================
         *  RELATIONSHIPS — Component Level
         * ================================================================ */

        /* API Gateway → Business Modules */
        govSubsidy.apiGateway -> govSubsidy.backend.claimMgmt "จัดการคำขอสิทธิ์" "HTTP"
        govSubsidy.apiGateway -> govSubsidy.backend.authEkyc "Authentication & eKYC" "HTTP"
        govSubsidy.apiGateway -> govSubsidy.backend.projectMgmt "จัดการโครงการ" "HTTP"
        govSubsidy.apiGateway -> govSubsidy.backend.officerReview "Officer Review" "HTTP"
        govSubsidy.apiGateway -> govSubsidy.backend.adminDashboard "Admin Dashboard" "HTTP"
        govSubsidy.apiGateway -> govSubsidy.backend.statusStream "SSE Status Stream" "HTTP/SSE"

        /* Claim Management → Async Processing */
        govSubsidy.backend.claimMgmt -> govSubsidy.backend.jobQueue "ส่งงาน Claim ไปคิว" "Internal"
        govSubsidy.backend.jobQueue -> govSubsidy.backend.orchestratorWorker "ดึงงานมาประมวลผล" "Redis Streams"
        govSubsidy.backend.orchestratorWorker -> govSubsidy.backend.decisionEngine "ส่งผลตรวจสอบเพื่อตัดสินใจ" "Internal"

        /* Integration Adapters → External APIs */
        govSubsidy.backend.orchestratorWorker -> govSubsidy.backend.dopaAdapter "เรียกตรวจสอบทะเบียนราษฎร" "Internal"
        govSubsidy.backend.orchestratorWorker -> govSubsidy.backend.ssoAdapter "เรียกตรวจสอบประกันสังคม" "Internal"
        govSubsidy.backend.orchestratorWorker -> govSubsidy.backend.ktbAdapter "เรียกตรวจสอบข้อมูลการเงิน" "Internal"
        govSubsidy.backend.dopaAdapter -> dopaApi "HTTP Request + Retry" "HTTPS"
        govSubsidy.backend.ssoAdapter -> ssoApi "HTTP Request + Retry" "HTTPS"
        govSubsidy.backend.ktbAdapter -> ktbApi "HTTP Request + Retry" "HTTPS"

        /* Auth & eKYC → External */
        govSubsidy.backend.authEkyc -> govSubsidy.backend.dopaAdapter "ตรวจสอบตัวตนระหว่าง eKYC" "Internal"

        /* Business Modules → Data Stores */
        govSubsidy.backend.claimMgmt -> govSubsidy.postgresWrite "CRUD Claims" "SQL"
        govSubsidy.backend.claimMgmt -> govSubsidy.redisCache "Cache Claim Status" "Redis Protocol"
        govSubsidy.backend.authEkyc -> govSubsidy.postgresWrite "CRUD Users, Sessions" "SQL"
        govSubsidy.backend.authEkyc -> govSubsidy.redisCache "Token Blacklist, Rate Limit" "Redis Protocol"
        govSubsidy.backend.authEkyc -> govSubsidy.objectStorage "จัดเก็บเอกสาร eKYC" "S3 API"
        govSubsidy.backend.projectMgmt -> govSubsidy.postgresWrite "CRUD Projects" "SQL"
        govSubsidy.backend.officerReview -> govSubsidy.postgresRead "Query Pending Claims" "SQL"
        govSubsidy.backend.officerReview -> govSubsidy.postgresWrite "Update Claim Decision" "SQL"
        govSubsidy.backend.adminDashboard -> govSubsidy.postgresRead "Query Dashboard Data" "SQL"
        govSubsidy.backend.statusStream -> govSubsidy.redisCache "Subscribe Status Events" "Redis Pub/Sub"
        govSubsidy.backend.decisionEngine -> govSubsidy.postgresWrite "บันทึกผลตัดสิน" "SQL"
        govSubsidy.backend.orchestratorWorker -> govSubsidy.postgresWrite "บันทึกผลตรวจสอบ" "SQL"

        /* Audit Trail */
        govSubsidy.backend.claimMgmt -> govSubsidy.auditLog "บันทึก Claim Actions" "SQL"
        govSubsidy.backend.authEkyc -> govSubsidy.auditLog "บันทึก Auth Events" "SQL"
        govSubsidy.backend.officerReview -> govSubsidy.auditLog "บันทึก Officer Decisions" "SQL"
        govSubsidy.backend.adminDashboard -> govSubsidy.auditLog "อ่าน Audit Log" "SQL"
        govSubsidy.backend.decisionEngine -> govSubsidy.auditLog "บันทึก Decision Results" "SQL"

        /* Queue interactions */
        govSubsidy.backend.jobQueue -> govSubsidy.redisQueue "Enqueue/Dequeue Jobs" "Redis Streams"
        govSubsidy.backend.orchestratorWorker -> govSubsidy.redisQueue "Consume Jobs" "Redis Streams"

        /* ================================================================
         *  DEPLOYMENT — Production Environment
         * ================================================================ */

        prodEnvironment = deploymentEnvironment "Production" {

            deploymentNode "Government Data Center" "ศูนย์ข้อมูลภาครัฐ ระดับ Tier-3 รองรับ High Availability" "" "DataCenter" {

                deploymentNode "Load Balancer Cluster" "HAProxy / Nginx Plus สำหรับ Traffic Distribution" "HAProxy" "LoadBalancer" {
                    lbInstance = containerInstance govSubsidy.apiGateway
                }

                deploymentNode "Application Cluster" "กลุ่ม Application Servers สำหรับ Horizontal Scaling" "" "AppCluster" {

                    deploymentNode "App Server 01" "Primary Application Node" "Ubuntu 22.04 LTS" "Server" {
                        backendInstance1 = containerInstance govSubsidy.backend
                    }

                    deploymentNode "App Server 02" "Secondary Application Node (Scale-out)" "Ubuntu 22.04 LTS" "Server" {
                        backendInstance2 = containerInstance govSubsidy.backend
                    }

                    deploymentNode "Web Server" "Frontend Static Assets Hosting" "Nginx" "WebServer" {
                        frontendInstance = containerInstance govSubsidy.frontend
                    }
                }

                deploymentNode "Database Cluster" "กลุ่มฐานข้อมูลหลัก รองรับ Replication และ Failover" "" "DBCluster" {

                    deploymentNode "PostgreSQL Primary Node" "Write Master — ฐานข้อมูลหลักสำหรับ Write Operations" "PostgreSQL 16" "DatabaseNode" {
                        pgWriteInstance = containerInstance govSubsidy.postgresWrite
                    }

                    deploymentNode "PostgreSQL Read Replica Node" "Read Replica — สำหรับ Query-heavy Workloads" "PostgreSQL 16" "DatabaseNode" {
                        pgReadInstance = containerInstance govSubsidy.postgresRead
                    }

                    deploymentNode "Audit Database Node" "Append-Only Audit Store — แยกเพื่อ Compliance" "PostgreSQL 16" "DatabaseNode" {
                        auditInstance = containerInstance govSubsidy.auditLog
                    }
                }

                deploymentNode "Cache & Queue Cluster" "Redis Cluster สำหรับ Cache และ Message Queue" "" "CacheCluster" {

                    deploymentNode "Redis Cache Node" "Cache, Session, Rate Limiting" "Redis 7" "CacheNode" {
                        redisCacheInstance = containerInstance govSubsidy.redisCache
                    }

                    deploymentNode "Redis Queue Node" "Async Job Queue" "Redis 7" "QueueNode" {
                        redisQueueInstance = containerInstance govSubsidy.redisQueue
                    }
                }

                deploymentNode "Storage Cluster" "Object Storage สำหรับเอกสารและไฟล์" "" "StorageCluster" {
                    deploymentNode "MinIO Node" "S3-Compatible Object Storage" "MinIO" "StorageNode" {
                        storageInstance = containerInstance govSubsidy.objectStorage
                    }
                }
            }

            deploymentNode "External Network Zone" "External Mock Provider Zone — แยก Network Segment" "" "ExternalZone" {
                deploymentNode "DOPA Provider" "" "" "ExternalNode" {
                    dopaInstance = softwareSystemInstance dopaApi
                }
                deploymentNode "SSO Provider" "" "" "ExternalNode" {
                    ssoInstance = softwareSystemInstance ssoApi
                }
                deploymentNode "KTB Provider" "" "" "ExternalNode" {
                    ktbInstance = softwareSystemInstance ktbApi
                }
            }
        }
    }

    /* ====================================================================
     *  VIEWS
     * ==================================================================== */

    views {

        /* ----------------------------------------------------------------
         *  Level 1: System Context Diagram
         * ---------------------------------------------------------------- */

        systemContext govSubsidy "SystemContext" "ภาพรวมระบบสวัสดิการภาครัฐ — แสดงผู้ใช้งาน ระบบหลัก และระบบภายนอกที่เชื่อมต่อ" {
            include *
            autoLayout lr 400 200
        }

        /* ----------------------------------------------------------------
         *  Level 2: Container Diagram
         * ---------------------------------------------------------------- */

        container govSubsidy "ContainerDiagram" "สถาปัตยกรรมระดับ Container — แสดง Frontend, API Gateway, Backend, Data Stores และ External Systems" {
            include *
            autoLayout lr 350 200
        }

        /* ----------------------------------------------------------------
         *  Level 3: Component Diagram (Backend)
         * ---------------------------------------------------------------- */

        component govSubsidy.backend "ComponentDiagram" "Component Diagram ของ Backend — แสดง Business Modules, Async Processing, Integration Adapters และการเชื่อมต่อกับ Data Stores" {
            include *
            autoLayout lr 300 150
        }

        /* ----------------------------------------------------------------
         *  Level 4: Deployment Diagram (Production)
         * ---------------------------------------------------------------- */

        deployment govSubsidy "Production" "DeploymentDiagram" "Deployment View — การ Deploy ระบบใน Government Data Center ระดับ Production" {
            include *
            autoLayout lr 350 200
        }

        /* ----------------------------------------------------------------
         *  STYLES
         * ---------------------------------------------------------------- */

        styles {

            /* ---- People ---- */
            element "Person" {
                shape Person
                fontSize 24
            }

            element "Citizen" {
                background #0d9488
                color #ffffff
                shape Person
            }

            element "Officer" {
                background #0369a1
                color #ffffff
                shape Person
            }

            element "Admin" {
                background #7c3aed
                color #ffffff
                shape Person
            }

            /* ---- Software Systems ---- */
            element "Software System" {
                shape RoundedBox
                fontSize 22
            }

            element "GovernmentSystem" {
                background #115e59
                color #ffffff
                shape RoundedBox
                fontSize 24
            }

            element "ExternalSystem" {
                background #64748b
                color #ffffff
                shape RoundedBox
            }

            /* ---- Containers ---- */
            element "Container" {
                shape RoundedBox
                fontSize 20
            }

            element "WebApp" {
                background #0d9488
                color #ffffff
                shape WebBrowser
            }

            element "Gateway" {
                background #0f766e
                color #ffffff
                shape Hexagon
            }

            element "Backend" {
                background #115e59
                color #ffffff
                shape RoundedBox
            }

            element "DatabasePrimary" {
                background #1e40af
                color #ffffff
                shape Cylinder
            }

            element "DatabaseReplica" {
                background #3b82f6
                color #ffffff
                shape Cylinder
            }

            element "DatabaseAudit" {
                background #7c3aed
                color #ffffff
                shape Cylinder
            }

            element "Cache" {
                background #dc2626
                color #ffffff
                shape RoundedBox
            }

            element "Queue" {
                background #ea580c
                color #ffffff
                shape RoundedBox
            }

            element "FileStore" {
                background #ca8a04
                color #ffffff
                shape Folder
            }

            /* ---- Components ---- */
            element "Component" {
                background #0d9488
                color #ffffff
                shape RoundedBox
                fontSize 18
            }

            /* ---- Deployment ---- */
            element "DataCenter" {
                background #f0fdf4
                color #166534
            }

            element "AppCluster" {
                background #ecfdf5
                color #065f46
            }

            element "DBCluster" {
                background #eff6ff
                color #1e40af
            }

            element "CacheCluster" {
                background #fef2f2
                color #991b1b
            }

            element "StorageCluster" {
                background #fefce8
                color #854d0e
            }

            element "Server" {
                background #d1fae5
                color #065f46
            }

            element "LoadBalancer" {
                background #ccfbf1
                color #0f766e
            }

            element "ExternalZone" {
                background #f1f5f9
                color #475569
            }

            /* ---- Relationships ---- */
            relationship "Relationship" {
                thickness 2
                color #334155
                fontSize 14
                style solid
            }
        }

        /* ----------------------------------------------------------------
         *  THEMES
         * ---------------------------------------------------------------- */

        theme default
    }
}
