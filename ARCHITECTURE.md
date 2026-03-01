# EduNex - Architecture

## 1. Overview
EduNex is a web platform designed for pedagogical coordinators to centralize school management tasks. The project follows the **Hexagonal Architecture (Ports & Adapters)** pattern, organized under the `internal` directory to enforce encapsulation. This design ensures that core pedagogical logic remains independent of frameworks, databases, and external identity providers, allowing for incremental MVP development with a stable and testable core.

## 2. Project Structure and Encapsulation
The backend is entirely contained within the `internal/` directory to prevent external packages from importing private logic, following Go’s idiomatic conventions. The system is divided into three main pillars: `core` for business logic, `api` for request handling, and `infra` for external integrations.

### Hexagonal Diagram:
![EduNex](https://github.com/user-attachments/assets/b89b01da-86c2-4dc0-b96a-56c3a0bfdc6b)



### The Application Core (`internal/core`)
The `core` is the heart of the application and contains the domain model and use cases. It is strictly decoupled from any specific technology, such as HTTP or SQL.

*   **Domain Entities:** Located in `internal/core/entities`, these are pure Go structs representing pedagogical concepts like `User` and `LessonPlan`.
*   **Domain Services:** Found in `internal/core/services`, they implement the actual business workflows, such as processing lesson plan uploads and coordinating AI analysis.
*   **Ports:** The core defines its boundaries through interfaces. **Primary Ports** (`iservice`) define what the core offers to the outside world, while **Secondary Ports** (`irepository`) define what the core requires from external systems, such as persistence, file storage and soon LLM processing.

### Driving Adapters (`internal/api`)
The `api` layer serves as the entry point for the application, exposing its functionality via HTTP. It is responsible for request/response marshalling and applying cross-cutting concerns. The **HTTP Handlers** within this layer translate incoming requests into calls to the Core's primary ports. A critical component here is the **Auth Middleware**, which intercepts requests to validate identity. It extracts the authenticated `UserID` and injects it into the Go `context.Context`, allowing the Core to execute business logic without being aware of the underlying authentication mechanism.

### Driven Adapters and Utilities (`internal/infra`)
The `infra` layer provides concrete implementations for the secondary ports defined by the Core. It handles all communication with external services:

*   **Persistence:** Uses **GORM** to interact with PostgreSQL, implementing the repository interfaces.
*   **Storage:** Integrates with **Supabase Buckets** to manage lesson plan files (PDF/DOCX).
*   **AI Integration:** Communicates directly with **LLM APIs** from Go to perform pedagogical analysis on uploaded documents.
*   **Security:** The `internal/infra/security` package contains the **JWT Manager**. This utility validates Supabase JWTs and extracts user identity. It is consumed exclusively by the API layer (Middleware and Handlers), reinforcing the principle that authentication is a transport-level infrastructure concern rather than a core business rule.

## 3. Authentication and Identity Flow
EduNex leverages **Supabase Auth** as its identity provider. The frontend authenticates directly with Supabase to obtain a JWT, which is then sent in the `Authorization` header. The backend's Auth Middleware uses the JWT Manager to verify the token and resolve the user's UUID. By the time a request reaches a Core Service, the user's identity is already resolved and available in the context, keeping the business logic clean and decoupled from Supabase-specific details.

## 4. Tech Stack Summary
| Layer | Technology |
|---|---|
| Language | Go (Golang) |
| Architecture | Hexagonal (Internalized) |
| ORM | GORM |
| Database | PostgreSQL |
| Auth & Storage | Supabase (Auth & Buckets) |
| AI | LLM API (Direct integration via Go) |
| Frontend | Next.js |

## 5. Implementation Status and Roadmap
The database schema (`init.sql`) is fully defined for the MVP scope. The implementation of domain services and adapters is progressing incrementally.

**Core Domains:**
- [x] **User** — Coordinator profile management linked to Supabase Auth.
- [x] **LessonPlan** — File upload to Supabase Buckets and AI-powered analysis.
- [ ] **School** — School management and user association (`users_schools`).
- [ ] **Activity** — Coordinator's agenda and task management.
- [ ] **Notification** — System alerts and messages.

**Academic Structure:**
- [ ] **EducationLevel** — Levels of education (e.g., High School, Elementary).
- [ ] **AcademicClass** — Class groupings by year, grade, and shift.
- [ ] **Discipline** — Subjects taught in the school.
- [ ] **Teacher** — Teacher profiles and assignments (`classes_disciplines_teachers`).

**Student & Performance:**
- [ ] **Student** — Enrollment, personal info, and class association.
- [ ] **Attendance** — Daily attendance records.
- [ ] **Performance** — Grades and recovery grades per term.

**AI & Background Jobs:**
- [ ] **AnalysisJob** — Queue management for AI processing (`pending`, `processing`, `done`).
- [ ] **LessonPlanAnalysis** — Storage for AI-generated feedback on lesson plans.

## 9. Architectural Decision Records (ADRs)

**ADR-01 — Hexagonal Architecture**
Adopted to isolate pedagogical domain complexity from infrastructure. This allows the core logic to remain unchanged even if we switch databases or AI providers.

**ADR-02 — Supabase Integration**
Used for Auth and Storage to accelerate MVP development. By delegating identity management and file hosting, the project focuses on unique pedagogical features.

**ADR-03 — AI Integration in Go**
Initially considered a Python microservice. Decided to integrate LLM APIs directly into the Go backend to reduce operational overhead and latency for the MVP (still not implemented).

**ADR-04 — JWT Manager Placement**
Authentication is treated as an infrastructure utility. The Core does not depend on JWT logic, ensuring that the business domain remains focused solely on pedagogical workflows.

**ADR-05 — Asynchronous AI Processing**
Implemented a job queue table (`analysis_jobs`) to handle AI analysis asynchronously. This prevents long-running AI requests from blocking the HTTP response, improving user experience and system reliability.
