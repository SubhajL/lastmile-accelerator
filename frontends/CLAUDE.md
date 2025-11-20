# Frontends - Next.js Applications

**Parent Context:** This extends [../CLAUDE.md](../CLAUDE.md)

This directory contains all frontend applications for the LMA platform. All frontends use **Next.js 14.2.6 (App Router)** with **pnpm** as the package manager.

## Frontends Overview

### Applications

1. **Mode A Dashboard** ([mode-a-dashboard/](mode-a-dashboard/))
   - Primary admin dashboard for project management
   - Full-featured interface for all LMA capabilities
   - User: Power users, team leads, admins

2. **Mode B Privacy Portal** ([mode-b-privacy-portal/](mode-b-privacy-portal/))
   - Privacy-focused interface for sensitive operations
   - Minimal tracking, local-first architecture
   - User: Privacy-conscious developers, enterprises

3. **Mode C ZIP Uploader** ([mode-c-zip-uploader/](mode-c-zip-uploader/))
   - Standalone ZIP upload interface
   - Drag-and-drop project ingestion
   - User: Developers without Git access, quick tests

## Technology Stack

### Core Framework
- **Next.js:** 14.2.6 (App Router)
- **React:** 18.x
- **TypeScript:** 5.x (strict mode)
- **Package Manager:** pnpm (NOT bun - frontends only)

### UI Libraries
- **Styling:** Tailwind CSS 3.x
- **Components:** Radix UI primitives
- **Icons:** Lucide React
- **Forms:** React Hook Form + Zod validation
- **State Management:** Zustand (global) + React Query (server state)

### Data Fetching
- **TanStack Query (React Query):** Server state management
- **Axios/Fetch:** HTTP client
- **SWR:** Alternative for some real-time data

### Testing
- **Vitest:** Unit testing
- **Testing Library:** Component testing
- **Playwright:** E2E testing
- **MSW:** API mocking

## Universal Frontend Patterns

### 1. Project Structure

```
<frontend-name>/
├── src/
│   ├── app/                  # Next.js App Router pages
│   │   ├── layout.tsx        # Root layout
│   │   ├── page.tsx          # Home page
│   │   ├── (auth)/           # Auth route group
│   │   └── (dashboard)/      # Dashboard route group
│   ├── components/           # React components
│   │   ├── ui/               # Base UI components (Radix)
│   │   ├── forms/            # Form components
│   │   ├── layout/           # Layout components
│   │   └── features/         # Feature-specific components
│   ├── hooks/                # Custom React hooks
│   ├── lib/                  # Utilities and helpers
│   │   ├── api/              # API client setup
│   │   ├── auth/             # Authentication utilities
│   │   └── utils.ts          # General utilities
│   ├── stores/               # Zustand stores
│   ├── types/                # TypeScript type definitions
│   ├── styles/               # Global styles and theme
│   └── config/               # App configuration
├── public/                   # Static assets
├── tests/                    # E2E tests (Playwright)
├── .env.local                # Local environment variables (gitignored)
├── .env.example              # Example environment variables
├── next.config.js            # Next.js configuration
├── tailwind.config.ts        # Tailwind configuration
├── tsconfig.json             # TypeScript configuration
├── package.json
└── pnpm-lock.yaml
```

### 2. Development Commands

**All frontends use pnpm:**

```bash
# Install dependencies
pnpm install

# Development server
pnpm dev                      # Starts on http://localhost:3000 (or configured port)

# Build
pnpm build                    # Production build
pnpm start                    # Start production server

# Testing
pnpm test                     # Run unit tests
pnpm test:watch               # Watch mode
pnpm test:e2e                 # Run Playwright E2E tests
pnpm test:e2e:ui              # Playwright UI mode

# Quality
pnpm typecheck                # TypeScript type checking
pnpm lint                     # ESLint
pnpm lint:fix                 # Auto-fix linting issues
pnpm format                   # Prettier formatting
```

**From root (using Turbo):**
```bash
bunx turbo run dev --filter=<frontend-name>
bunx turbo run build --filter=<frontend-name>
bunx turbo run test --filter=<frontend-name>
```

### 3. Code Organization Patterns

#### Components

**✅ DO:** Use functional components with hooks
```tsx
// ✅ Good
export function UserProfile({ userId }: UserProfileProps) {
  const { data: user } = useUser(userId)
  return <div>{user?.name}</div>
}
```

**❌ DON'T:** Use class components
```tsx
// ❌ Bad
class UserProfile extends React.Component {}
```

#### File Naming
- Components: `PascalCase.tsx` (e.g., `UserProfile.tsx`)
- Hooks: `camelCase.ts` with `use` prefix (e.g., `useUser.ts`)
- Utilities: `camelCase.ts` (e.g., `formatDate.ts`)
- Pages: `page.tsx` (Next.js App Router convention)
- Layouts: `layout.tsx` (Next.js App Router convention)

#### Server vs Client Components

**Default:** Server Components (Next.js 13+)

**Use Client Components when:**
- Using React hooks (`useState`, `useEffect`, etc.)
- Handling browser events (onClick, onChange, etc.)
- Using browser-only APIs (localStorage, window, etc.)

```tsx
// Server Component (default, no directive)
export default function Page() {
  return <div>Static content</div>
}

// Client Component (explicit directive)
'use client'
export default function InteractivePage() {
  const [count, setCount] = useState(0)
  return <button onClick={() => setCount(c => c + 1)}>{count}</button>
}
```

#### State Management

**Local State:** `useState`
```tsx
const [isOpen, setIsOpen] = useState(false)
```

**Global State:** Zustand
```tsx
// stores/userStore.ts
export const useUserStore = create<UserStore>((set) => ({
  user: null,
  setUser: (user) => set({ user }),
}))

// Component
const user = useUserStore((state) => state.user)
```

**Server State:** TanStack Query
```tsx
export function useUser(userId: string) {
  return useQuery({
    queryKey: ['user', userId],
    queryFn: () => fetchUser(userId),
  })
}
```

#### Data Fetching

**✅ DO:** Use TanStack Query for server data
```tsx
// hooks/useProjects.ts
export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: async () => {
      const response = await fetch('/api/projects')
      return response.json()
    },
  })
}

// Component
function ProjectsList() {
  const { data: projects, isLoading } = useProjects()
  if (isLoading) return <Loading />
  return <ul>{projects?.map(p => <li key={p.id}>{p.name}</li>)}</ul>
}
```

**❌ DON'T:** Fetch in useEffect
```tsx
// ❌ Bad
const [projects, setProjects] = useState([])
useEffect(() => {
  fetch('/api/projects').then(r => r.json()).then(setProjects)
}, [])
```

#### Forms

**Use React Hook Form + Zod:**

```tsx
// schemas/projectSchema.ts
export const projectSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
})

// components/forms/ProjectForm.tsx
'use client'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'

export function ProjectForm() {
  const form = useForm({
    resolver: zodResolver(projectSchema),
    defaultValues: { name: '', description: '' },
  })

  const onSubmit = form.handleSubmit(async (data) => {
    await createProject(data)
  })

  return (
    <form onSubmit={onSubmit}>
      <input {...form.register('name')} />
      {form.formState.errors.name && <span>{form.formState.errors.name.message}</span>}
      <button type="submit">Create</button>
    </form>
  )
}
```

#### Styling

**Use Tailwind CSS utility classes:**

```tsx
// ✅ Good
<div className="flex items-center gap-4 rounded-lg bg-slate-100 p-4">
  <h2 className="text-xl font-semibold text-slate-900">Title</h2>
</div>
```

**Don't hardcode colors:**
```tsx
// ❌ Bad
<div className="bg-blue-500">

// ✅ Good - Use design tokens
<div className="bg-primary">
```

**Design Tokens:**
Define in `tailwind.config.ts`:
```ts
theme: {
  extend: {
    colors: {
      primary: 'hsl(var(--primary))',
      secondary: 'hsl(var(--secondary))',
    }
  }
}
```

### 4. API Integration

**API Client Setup:**
```ts
// lib/api/client.ts
import axios from 'axios'

export const apiClient = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  timeout: 10000,
})

// Add auth token interceptor
apiClient.interceptors.request.use((config) => {
  const token = getAuthToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})
```

**TanStack Query Integration:**
```ts
// lib/api/projects.ts
export const projectsApi = {
  getAll: () => apiClient.get('/projects').then(r => r.data),
  getById: (id: string) => apiClient.get(`/projects/${id}`).then(r => r.data),
  create: (data: CreateProjectDto) => apiClient.post('/projects', data).then(r => r.data),
}

// hooks/useProjects.ts
export function useProjects() {
  return useQuery({
    queryKey: ['projects'],
    queryFn: projectsApi.getAll,
  })
}

export function useCreateProject() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: projectsApi.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
  })
}
```

### 5. Authentication

**JWT Token Management:**
```ts
// lib/auth/tokens.ts
export function getAuthToken(): string | null {
  return localStorage.getItem('auth_token')
}

export function setAuthToken(token: string): void {
  localStorage.setItem('auth_token', token)
}

export function removeAuthToken(): void {
  localStorage.removeItem('auth_token')
}
```

**Protected Routes:**
```tsx
// app/(dashboard)/layout.tsx
'use client'
import { useAuth } from '@/hooks/useAuth'
import { redirect } from 'next/navigation'

export default function DashboardLayout({ children }) {
  const { user, isLoading } = useAuth()

  if (isLoading) return <Loading />
  if (!user) redirect('/login')

  return <div>{children}</div>
}
```

### 6. Environment Variables

**Client-Side Variables (MUST have `NEXT_PUBLIC_` prefix):**
```bash
# .env.local
NEXT_PUBLIC_API_URL=http://localhost:7000
NEXT_PUBLIC_OIDC_ISSUER=http://localhost:8080/realms/lma
NEXT_PUBLIC_SENTRY_DSN=https://...
```

**Server-Side Variables (no prefix):**
```bash
DATABASE_URL=postgresql://...
SECRET_KEY=...
```

**Usage:**
```ts
// Client component
const apiUrl = process.env.NEXT_PUBLIC_API_URL

// Server component or API route
const dbUrl = process.env.DATABASE_URL
```

### 7. Testing

**Unit Tests (Vitest + Testing Library):**
```tsx
// components/Button.test.tsx
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { Button } from './Button'

describe('Button', () => {
  it('renders with text', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByText('Click me')).toBeInTheDocument()
  })

  it('calls onClick when clicked', async () => {
    const onClick = vi.fn()
    render(<Button onClick={onClick}>Click</Button>)
    await userEvent.click(screen.getByText('Click'))
    expect(onClick).toHaveBeenCalledOnce()
  })
})
```

**E2E Tests (Playwright):**
```ts
// tests/projects.spec.ts
import { test, expect } from '@playwright/test'

test('user can create project', async ({ page }) => {
  await page.goto('/projects')
  await page.click('text=New Project')
  await page.fill('input[name="name"]', 'Test Project')
  await page.click('button[type="submit"]')
  await expect(page.locator('text=Test Project')).toBeVisible()
})
```

## Pre-PR Checklist

**Run before creating a PR:**
```bash
pnpm typecheck && pnpm lint && pnpm test && pnpm build
```

**Or from root:**
```bash
bunx turbo run typecheck lint test build --filter=<frontend-name>
```

## Common Gotchas

- **Environment Variables:** Client-side vars MUST have `NEXT_PUBLIC_` prefix
- **Import Paths:** Use `@/` alias for `src/` imports (configured in `tsconfig.json`)
- **Server Components:** Default in Next.js 13+; add `"use client"` directive when needed
- **Dynamic Routes:** Params are now async in Next.js 15: `await params.id`
- **Hydration Errors:** Ensure server and client render the same initial HTML
- **Image Optimization:** Use `next/image` for automatic optimization
- **Font Optimization:** Use `next/font` for automatic font loading
- **API Routes:** Use `/app/api/` directory, not `/pages/api/`

## Related Services

Frontends communicate with:
- **projects-service** (REST: 7002) - Project management
- **notification-service** (REST: 7902) - User notifications
- **observability-service** (REST: 7301) - Logs, metrics, traces
- **authz-matrix-service** (REST: 7203) - Authorization

See `service_catalog.yaml` for complete service directory.

## Useful Links

- **Next.js Docs:** https://nextjs.org/docs
- **TanStack Query:** https://tanstack.com/query
- **Tailwind CSS:** https://tailwindcss.com
- **Radix UI:** https://radix-ui.com
- **Playwright:** https://playwright.dev
- **Service Catalog:** `../service_catalog.yaml`
- **CI Workflows:** `../.github/workflows/`
