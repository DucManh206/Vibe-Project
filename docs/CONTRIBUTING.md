# 🤝 Contributing Guide

Cảm ơn bạn đã quan tâm đến việc đóng góp cho Captcha Platform! Tài liệu này hướng dẫn cách tham gia phát triển dự án.

---

## Code of Conduct

- Tôn trọng tất cả contributors
- Xây dựng môi trường hợp tác
- Feedback mang tính xây dựng
- Focus vào vấn đề, không phải cá nhân

---

## Getting Started

### 1. Fork Repository

```bash
# Fork trên GitHub, sau đó clone
git clone https://github.com/YOUR_USERNAME/captcha-platform.git
cd captcha-platform
```

### 2. Setup Development Environment

```bash
# Run setup script
chmod +x scripts/setup.sh
./scripts/setup.sh --build
```

### 3. Create Branch

```bash
# Từ main branch
git checkout main
git pull origin main
git checkout -b feature/your-feature-name
```

---

## Development Workflow

### Branch Naming

- `feature/` - Tính năng mới
- `fix/` - Bug fixes
- `docs/` - Documentation
- `refactor/` - Code refactoring
- `test/` - Tests

Examples:
- `feature/add-recaptcha-support`
- `fix/login-validation-error`
- `docs/update-api-reference`

### Commit Messages

Sử dụng [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat` - New feature
- `fix` - Bug fix
- `docs` - Documentation
- `style` - Formatting, no code change
- `refactor` - Code restructuring
- `test` - Adding tests
- `chore` - Maintenance

**Examples:**
```bash
feat(auth): add password reset functionality
fix(captcha): resolve memory leak in OCR solver
docs(api): update authentication examples
test(auth): add unit tests for JWT validation
```

---

## Code Standards

### Go (Backend)

**Style:**
- Follow [Effective Go](https://golang.org/doc/effective_go)
- Use `gofmt` for formatting
- Run `golint` and `go vet`

**Example:**
```go
// Package comment
package handlers

import (
    "net/http"
    
    "github.com/gin-gonic/gin"
)

// Handler handles HTTP requests
type Handler struct {
    service *Service
    logger  *Logger
}

// NewHandler creates a new Handler
func NewHandler(service *Service, logger *Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

// HandleRequest processes incoming requests
func (h *Handler) HandleRequest(c *gin.Context) {
    // Implementation
}
```

### Rust (Captcha Service)

**Style:**
- Follow [Rust Style Guide](https://doc.rust-lang.org/1.0.0/style/README.html)
- Use `cargo fmt` for formatting
- Run `cargo clippy` for linting

**Example:**
```rust
//! Module documentation

use crate::error::{CaptchaError, CaptchaResult};

/// Solver trait for captcha solving
pub trait Solver {
    /// Solve a captcha image
    fn solve(&self, image: &[u8]) -> CaptchaResult<String>;
}

/// CNN-based solver implementation
pub struct CnnSolver {
    model_path: String,
}

impl CnnSolver {
    /// Create a new CNN solver
    pub fn new(model_path: &str) -> Self {
        Self {
            model_path: model_path.to_string(),
        }
    }
}

impl Solver for CnnSolver {
    fn solve(&self, image: &[u8]) -> CaptchaResult<String> {
        // Implementation
        Ok("result".to_string())
    }
}
```

### TypeScript/React (Frontend)

**Style:**
- Follow [React Style Guide](https://reactjs.org/docs/getting-started.html)
- Use ESLint + Prettier
- Prefer functional components

**Example:**
```tsx
'use client';

import { useState } from 'react';
import { useTranslations } from 'next-intl';

interface Props {
  title: string;
  onSubmit: (value: string) => void;
}

export function MyComponent({ title, onSubmit }: Props) {
  const t = useTranslations('common');
  const [value, setValue] = useState('');

  const handleSubmit = () => {
    onSubmit(value);
    setValue('');
  };

  return (
    <div className="container">
      <h1>{title}</h1>
      <input 
        value={value} 
        onChange={(e) => setValue(e.target.value)} 
      />
      <button onClick={handleSubmit}>
        {t('submit')}
      </button>
    </div>
  );
}
```

---

## Testing

### Running Tests

```bash
# Backend - Go
cd backend/auth
go test ./...

# Backend - Rust
cd backend/captcha
cargo test

# Frontend
cd frontend
npm test
```

### Writing Tests

**Go:**
```go
func TestLogin_Success(t *testing.T) {
    // Arrange
    service := NewMockService()
    handler := NewHandler(service)
    
    // Act
    resp := handler.Login(mockRequest)
    
    // Assert
    assert.Equal(t, http.StatusOK, resp.Code)
}
```

**Rust:**
```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_solve_captcha() {
        let solver = CnnSolver::new("/path/to/model");
        let result = solver.solve(b"image_data");
        assert!(result.is_ok());
    }
}
```

**TypeScript:**
```tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { MyComponent } from './MyComponent';

describe('MyComponent', () => {
  it('should submit value', () => {
    const onSubmit = jest.fn();
    render(<MyComponent title="Test" onSubmit={onSubmit} />);
    
    fireEvent.change(screen.getByRole('textbox'), {
      target: { value: 'test' },
    });
    fireEvent.click(screen.getByRole('button'));
    
    expect(onSubmit).toHaveBeenCalledWith('test');
  });
});
```

---

## Pull Request Process

### 1. Before Submitting

- [ ] Code follows style guidelines
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] No sensitive data committed
- [ ] Rebased on latest main

### 2. PR Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
How was this tested?

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-reviewed code
- [ ] Added tests
- [ ] Updated documentation
- [ ] No breaking changes

## Screenshots (if applicable)
```

### 3. Review Process

1. Automated checks run (lint, tests)
2. At least 1 reviewer approval required
3. Address feedback promptly
4. Squash merge to main

---

## Project Structure

```
captcha-platform/
├── backend/
│   ├── auth/           # Auth Service (Go)
│   ├── gateway/        # API Gateway (Go)
│   └── captcha/        # Captcha Service (Rust)
├── frontend/           # Next.js App
├── database/           # SQL migrations
├── docs/               # Documentation
├── scripts/            # Utility scripts
└── docker-compose.yml  # Docker config
```

---

## Adding New Features

### 1. New API Endpoint

1. Define route in gateway
2. Implement handler
3. Add service logic
4. Write tests
5. Update API docs

### 2. New Frontend Page

1. Create page in `app/`
2. Add translations
3. Create components
4. Add to navigation
5. Write tests

### 3. New Captcha Solver

1. Implement `CaptchaSolver` trait
2. Add to `SolverManager`
3. Create model file
4. Write tests
5. Update docs

---

## Getting Help

- **Discord:** [Join our server](#)
- **Issues:** [GitHub Issues](https://github.com/your-org/captcha-platform/issues)
- **Discussions:** [GitHub Discussions](https://github.com/your-org/captcha-platform/discussions)

---

## Recognition

Contributors are recognized in:
- README.md
- Release notes
- Annual contributor highlights

Thank you for contributing! 🎉