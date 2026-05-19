# Benefit Service API Contract

Owner: ณัฏฐ์นลิน บุญทรัพย์ทีปกร 6609650343

Status: Version 1, fully implemented

## Overview

Benefit Service provides APIs for citizens to submit benefit claims, check eligibility status, view claim history, and browse available projects. The service integrates with the Orchestrator to perform async eligibility verification across DOPA, SSO, and KTB.

## Base URL

```text
http://localhost:8080
```

## Endpoints

### 1. POST /api/v1/benefit/claim - Submit Claim

**Purpose**: Citizens submit a benefit claim for a specific project. The service creates a claim record and triggers the Orchestrator to verify eligibility asynchronously.

**Request**:
```json
{
  "nationalId": "1101700203451",
  "projectId": "proj-abc123"
}
```

**Fields**:
| Field | Required | Type | Description |
|-------|----------|------|-------------|
| `nationalId` | Yes | string | Thai national ID (13 digits) |
| `projectId` | Yes | string | ID of the benefit project |

**Response** `201 Created`:
```json
{
  "id": "claim-a1b2c3d4",
  "nationalId": "1101700203451",
  "projectId": "proj-abc123",
  "status": "processing",
  "submittedAt": "2026-05-19T18:10:30Z",
  "updatedAt": "2026-05-19T18:10:30Z",
  "eligibility": null,
  "project": {
    "id": "proj-abc123",
    "name": "Program Name",
    "description": "Program Description",
    "active": true,
    "criteria": {
      "minAge": 18,
      "maxAge": 60,
      "maxMonthlyIncome": 30000,
      "allowedSsoSections": ["40"],
      "requirePromptPay": true
    },
    "createdAt": "2026-05-19T10:00:00Z",
    "updatedAt": "2026-05-19T10:00:00Z"
  }
}
```

**Error** `400 Bad Request`:
```json
{
  "error": "nationalId is required"
}
```

**Error** `404 Not Found`:
```json
{
  "error": "project not found"
}
```

---

### 2. GET /api/v1/benefit/status/:claimId - Check Claim Status

**Purpose**: Citizens check the current eligibility decision for their claim. The response includes the latest decision from the Orchestrator if available.

**Response** `200 OK`:
```json
{
  "id": "claim-a1b2c3d4",
  "nationalId": "1101700203451",
  "projectId": "proj-abc123",
  "status": "approved",
  "submittedAt": "2026-05-19T18:10:30Z",
  "updatedAt": "2026-05-19T18:10:35Z",
  "eligibility": {
    "claimId": "claim-a1b2c3d4",
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
  },
  "project": {
    "id": "proj-abc123",
    "name": "Program Name",
    "description": "Program Description",
    "active": true,
    "criteria": {
      "minAge": 18,
      "maxAge": 60,
      "maxMonthlyIncome": 30000,
      "allowedSsoSections": ["40"],
      "requirePromptPay": true
    },
    "createdAt": "2026-05-19T10:00:00Z",
    "updatedAt": "2026-05-19T10:00:00Z"
  }
}
```

**Error** `404 Not Found`:
```json
{
  "error": "claim not found"
}
```

**Note**: While Orchestrator processes the claim, `status` will be `"processing"` and `eligibility` may be `null`. Poll this endpoint to get the final decision.

---

### 3. GET /api/v1/benefit/history/:citizenId - View Claim History

**Purpose**: Citizens view all their benefit claims (submitted, approved, rejected, pending).

**Response** `200 OK`:
```json
{
  "nationalId": "1101700203451",
  "claims": [
    {
      "id": "claim-a1b2c3d4",
      "nationalId": "1101700203451",
      "projectId": "proj-abc123",
      "status": "approved",
      "submittedAt": "2026-05-19T18:10:30Z",
      "updatedAt": "2026-05-19T18:10:35Z",
      "eligibility": {
        "claimId": "claim-a1b2c3d4",
        "status": "approved",
        "reasons": ["eligible by mock subsidy rules"],
        "sources": {
          "dopa": { "valid": true, "age": 35, "alive": true, "cardActive": true },
          "sso": { "section": "40", "contributionMonths": 12 },
          "ktb": { "depositTotal": 12000, "averageMonthlyIncome": 15000, "promptPayLinked": true }
        }
      },
      "project": { "id": "proj-abc123", "name": "Program 1", ... }
    },
    {
      "id": "claim-x9y8z7w6",
      "nationalId": "1101700203451",
      "projectId": "proj-def456",
      "status": "rejected",
      "submittedAt": "2026-05-18T10:00:00Z",
      "updatedAt": "2026-05-18T10:05:00Z",
      "eligibility": {
        "claimId": "claim-x9y8z7w6",
        "status": "rejected",
        "reasons": ["citizen is insured under SSO section 33"],
        "sources": { ... }
      },
      "project": { "id": "proj-def456", "name": "Program 2", ... }
    }
  ]
}
```

**Error** `400 Bad Request`:
```json
{
  "error": "nationalId is required"
}
```

---

### 4. GET /api/v1/benefit/projects - List Available Projects

**Purpose**: Retrieve all active benefit projects that citizens can apply for.

**Response** `200 OK`:
```json
{
  "projects": [
    {
      "id": "proj-abc123",
      "name": "Senior Citizen Basic Allowance",
      "description": "Monthly allowance for citizens age 60 and above",
      "active": true,
      "criteria": {
        "minAge": 60,
        "maxAge": null,
        "maxMonthlyIncome": 30000,
        "allowedSsoSections": ["40"],
        "requirePromptPay": true
      },
      "createdAt": "2026-05-01T09:00:00Z",
      "updatedAt": "2026-05-19T12:00:00Z"
    },
    {
      "id": "proj-def456",
      "name": "Low-Income Family Support",
      "description": "Financial support for families with income below threshold",
      "active": true,
      "criteria": {
        "minAge": 18,
        "maxAge": 65,
        "maxMonthlyIncome": 20000,
        "allowedSsoSections": ["33", "39", "40"],
        "requirePromptPay": false
      },
      "createdAt": "2026-05-05T10:30:00Z",
      "updatedAt": "2026-05-19T12:00:00Z"
    }
  ]
}
```

**Note**: Only active projects are returned. Inactive projects are excluded.

---

## Claim Status Values

| Status | Meaning |
|--------|---------|
| `processing` | Orchestrator is verifying eligibility (async) |
| `approved` | Claim approved by eligibility rules |
| `rejected` | Claim rejected by eligibility rules |
| `pending` | Requires manual officer review |

---

## Eligibility Decision Rules

The Orchestrator evaluates the following rules:

### ✅ APPROVED
- DOPA: valid, alive, active card
- Age: >= 18
- SSO: Section 40 (or eligible sections per project)
- KTB: PromptPay linked
- Income: <= max monthly income

### ❌ REJECTED
- DOPA: invalid, deceased, or revoked card
- Age: < 18
- SSO: Section 33 (self-employed)
- Income: > max monthly income

### ⏳ PENDING
- Orchestrator unreachable (requires officer review)
- PromptPay not linked (conditional approval)

---

## Error Handling

All endpoints return standard error responses:

```json
{
  "error": "error message describing what went wrong"
}
```

HTTP Status Codes:
- `201 Created` - Claim successfully created
- `200 OK` - Data retrieved successfully
- `400 Bad Request` - Invalid input or validation failed
- `404 Not Found` - Resource (claim or project) not found
- `500 Internal Server Error` - Unexpected server error

---

## Integration with Orchestrator

When a claim is submitted, the Benefit Service triggers an asynchronous orchestration:

1. Create claim with status `"processing"`
2. Call `POST /internal/orchestrate` on Orchestrator service
3. Orchestrator calls DOPA, SSO, KTB in parallel
4. Orchestrator makes decision and publishes events
5. Frontend polls `GET /api/v1/benefit/status/:claimId` to get updated status

---

## Usage Examples

### Submit a Claim
```bash
curl -X POST http://localhost:8080/api/v1/benefit/claim \
  -H "Content-Type: application/json" \
  -d '{
    "nationalId": "1101700203451",
    "projectId": "proj-abc123"
  }'
```

### Check Claim Status
```bash
curl http://localhost:8080/api/v1/benefit/status/claim-a1b2c3d4
```

### View Claim History
```bash
curl http://localhost:8080/api/v1/benefit/history/1101700203451
```

### List Available Projects
```bash
curl http://localhost:8080/api/v1/benefit/projects
```

---

## Testing

All endpoints have been unit tested:
- ✅ Submit claim with valid data
- ✅ Reject submit with missing nationalId or projectId
- ✅ Reject submit with non-existent projectId
- ✅ Fetch claim status successfully
- ✅ Fetch claim history for citizen
- ✅ Return only active projects
