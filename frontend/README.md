# Frontend — Social Network

Vue 3 single-page application for the social network platform. Communicates with the Go backend via REST API and WebSocket.

---

## Tech Stack

| Concern | Choice |
|---------|--------|
| Framework | Vue 3 (Composition API) |
| Language | TypeScript |
| Build tool | Vite |
| Routing | Vue Router |
| Real-time | Native WebSocket |
| Styling | Scoped CSS (CSS custom properties) |

---

## Project Structure

```
frontend/
├── src/
│   ├── api/          # API route constants and base URL helpers
│   ├── assets/       # Static assets (icons, images)
│   ├── components/   # Reusable UI components
│   ├── composables/  # Shared Vue composables (WebSocket, auth, etc.)
│   ├── router/       # Vue Router configuration and route guards
│   ├── stores/       # Global state (auth session, notifications)
│   ├── utils/        # Utility helpers (formatting, validation, debounce)
│   └── views/        # Page-level components (one per route)
├── public/           # Static files served as-is
├── index.html
└── Dockerfile        # Nginx-based production image
```

---

## Running Locally

**Prerequisites:** Node.js 18+

```bash
cd frontend
npm install
npm run dev
```

App runs at `http://localhost:5173` by default.

---

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_API_BASE_URL` | `http://localhost:8080` | Base URL of the backend API |

Create a `.env.local` file in the `frontend/` directory to override:

```env
VITE_API_BASE_URL=http://localhost:8080
```

---

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server with hot-reload |
| `npm run build` | Type-check and build for production |
| `npm run lint` | Lint source files with ESLint |

---

## Running with Docker

The frontend is served via Nginx in production. The recommended way is via `docker compose` from the repo root:

```bash
docker compose up --build
```

The Vite bundle is compiled at image build time using `VITE_API_BASE_URL=http://localhost:8080`. The Nginx container exposes port 80 and proxies API/WebSocket requests to the backend container.
