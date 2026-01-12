# Frontend Documentation

## Tổng Quan

Frontend của Captcha Platform được xây dựng với Next.js 14+ sử dụng App Router, TypeScript, Tailwind CSS, và hỗ trợ đa ngôn ngữ (i18n).

## Tech Stack

- **Framework**: Next.js 14 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: Zustand
- **Data Fetching**: TanStack Query (React Query)
- **Forms**: React Hook Form + Zod
- **i18n**: next-intl
- **UI Components**: Radix UI + Custom Components
- **Icons**: Lucide React

## Cấu Trúc Thư Mục

```
frontend/
├── src/
│   ├── app/                    # App Router pages
│   │   ├── (auth)/            # Auth pages (login, register)
│   │   ├── (dashboard)/       # Dashboard pages (protected)
│   │   ├── globals.css        # Global styles
│   │   ├── layout.tsx         # Root layout
│   │   └── page.tsx           # Landing page
│   ├── components/            # React components
│   │   ├── ui/               # Base UI components (Button, etc.)
│   │   ├── forms/            # Form components
│   │   ├── layouts/          # Layout components
│   │   └── providers.tsx     # Context providers
│   ├── hooks/                 # Custom React hooks
│   ├── lib/                   # Utility functions
│   ├── services/             # API services
│   │   └── api.ts            # Axios instance + API calls
│   ├── stores/               # Zustand stores
│   │   └── auth-store.ts     # Authentication state
│   ├── types/                # TypeScript types
│   └── i18n/                 # Internationalization
│       ├── request.ts        # i18n config
│       └── locales/          # Translation files
│           ├── vi.json       # Vietnamese
│           └── en.json       # English
├── public/                    # Static files
├── Dockerfile                # Docker config
├── package.json
├── tailwind.config.ts
├── tsconfig.json
└── next.config.js
```

## Pages

### Public Pages
- `/` - Landing page
- `/login` - Đăng nhập
- `/register` - Đăng ký

### Protected Pages (Dashboard)
- `/dashboard` - Dashboard chính
- `/solve` - Giải captcha
- `/models` - Quản lý models
- `/training` - Huấn luyện model
- `/api-keys` - Quản lý API keys
- `/logs` - Lịch sử xử lý
- `/settings` - Cài đặt

## State Management

### Auth Store (Zustand)

```typescript
interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  
  setAuth: (user, accessToken, refreshToken) => void;
  logout: () => void;
}
```

### React Query

Sử dụng TanStack Query cho data fetching:

```typescript
// Fetch stats
const { data, isLoading } = useQuery({
  queryKey: ['stats'],
  queryFn: () => captchaApi.getStats(),
});

// Mutation
const mutation = useMutation({
  mutationFn: captchaApi.solve,
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['logs'] });
  },
});
```

## API Services

API client được cấu hình trong `src/services/api.ts`:

```typescript
// Auth API
authApi.login(data)
authApi.register(data)
authApi.logout()
authApi.getMe()

// Captcha API
captchaApi.solve(data)
captchaApi.solveBatch(images)
captchaApi.getModels()
captchaApi.getStats()
captchaApi.getLogs(params)

// API Keys API
apiKeysApi.list()
apiKeysApi.create(data)
apiKeysApi.delete(id)

// Training API
trainingApi.start(data)
trainingApi.getStatus(jobId)
```

## Internationalization (i18n)

### Cấu hình

Ngôn ngữ được phát hiện tự động từ:
1. Cookie `NEXT_LOCALE`
2. Header `Accept-Language`
3. Default: `vi`

### Sử dụng

```typescript
import { useTranslations } from 'next-intl';

function Component() {
  const t = useTranslations('dashboard');
  return <h1>{t('title')}</h1>;
}
```

### Thêm ngôn ngữ mới

1. Tạo file `src/i18n/locales/{lang}.json`
2. Thêm locale vào `src/i18n/request.ts`
3. Cập nhật UI trong Settings page

## Components

### UI Components

- `Button` - Primary, secondary, outline, ghost variants
- `Toaster` - Toast notifications
- Modal dialogs với Radix UI

### Form Handling

```typescript
const { register, handleSubmit, formState } = useForm({
  resolver: zodResolver(schema),
});
```

## Styling

### Tailwind CSS

Sử dụng Tailwind CSS với custom design tokens:

```css
/* globals.css */
:root {
  --background: 0 0% 100%;
  --foreground: 222.2 84% 4.9%;
  --primary: 222.2 47.4% 11.2%;
  /* ... */
}
```

### Dark Mode

Hỗ trợ dark mode với `next-themes`:

```typescript
import { useTheme } from 'next-themes';

const { theme, setTheme } = useTheme();
setTheme('dark'); // 'light', 'dark', 'system'
```

## Development

### Cài đặt

```bash
cd frontend
npm install
```

### Chạy development

```bash
npm run dev
```

### Build production

```bash
npm run build
npm start
```

### Docker

```bash
docker build -t captcha-frontend .
docker run -p 3000:3000 captcha-frontend
```

## Environment Variables

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
NEXT_PUBLIC_APP_NAME=Captcha Platform
NEXT_PUBLIC_DEFAULT_LOCALE=vi
```

## Testing

```bash
# Unit tests
npm test

# E2E tests
npm run e2e
```

## Best Practices

1. **TypeScript**: Luôn định nghĩa types cho props và API responses
2. **Error Handling**: Sử dụng try-catch và hiển thị error messages
3. **Loading States**: Hiển thị loading indicators
4. **Form Validation**: Sử dụng Zod schemas
5. **Accessibility**: Thêm ARIA labels và keyboard navigation
6. **Performance**: Sử dụng React.memo, useMemo, useCallback khi cần