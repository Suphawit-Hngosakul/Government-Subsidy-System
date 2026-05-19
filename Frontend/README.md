# Government Subsidy System Frontend

Next.js frontend for the Government Subsidy System, focused on production dashboards, citizen claim flow, officer/admin workflows, and real-time status preview.

## Run

```bash
npm run dev
```

Open `http://localhost:3000`.

## Routes

```text
/                 Role gateway
/citizen/login    Citizen login
/citizen/register Citizen registration
/citizen          Citizen claim dashboard
/admin/login      Officer/Admin login
/admin            Admin console
```

## Backend Integration

The claim page calls the Next.js API proxy:

```text
POST /api/orchestrate
```

By default, the proxy forwards requests to the Go backend:

```text
http://localhost:8080/internal/orchestrate
```

To change the backend URL, set this environment variable:

```bash
GOV_SUBSIDY_BACKEND_URL=http://localhost:8080 npm run dev
```

If the Go backend is not running, the UI falls back to a mock simulation so the frontend demo remains usable.

## Quality Check

```bash
npm run lint
npm run build
```
