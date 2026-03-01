# EduNex

EduNex is a platform I am building to **streamline my own workflow as a Pedagogical Coordinator**. By combining my background in education with backend development in Go, I am creating a central hub to reduce manual work, organize school data, and experiment with AI-supported pedagogical analysis.

## 🎯 Vision

In my day-to-day work as a coordinator, I deal with fragmented information (spreadsheets, PDFs, messages) and a lot of manual checking of lesson plans and student performance.

EduNex is my personal tool to:

- Reduce time spent on repetitive coordination tasks.
- Keep student and school data consistent and queryable.
- Use AI to support lesson plan review and decision-making.

## 🚀 Key Features (Current & Planned)

- **Authentication:** Supabase Auth integration with local JWT (ES256) validation.
- **Lesson Plan Management:**
    - Upload and storage of lesson plans in Supabase Buckets.
    - Secure access via signed URLs for download and viewing.
- **AI-Powered Analysis (Planned / In Progress):**
    - Asynchronous job processing for lesson plan analysis.
    - Text extraction from PDF/DOCX and LLM-based feedback.
- **Coordinator Workflow (Planned):**
    - Activity and agenda tracking focused on my coordination routines.
    - Future modules for student performance and attendance tracking.

## 🛠 Architecture & Tech Stack

I am using this project to apply and refine my knowledge of **Hexagonal Architecture (Ports & Adapters)**, aiming to keep the pedagogical domain logic isolated from infrastructure details. You can find a detailed breakdown of the design and structure in the [**ARCHITECTURE.md**](./ARCHITECTURE.md) file.

- **Backend:** Go (Golang), focusing on clean code and separation of concerns.
- **Architecture:** Hexagonal (internal `core`, `api`, `infra` layout).
- **Database:** PostgreSQL (with GORM) for structured pedagogical and school data.
- **Cloud Storage:** Supabase Storage (Buckets) for lesson plan files.
- **Authentication:** Supabase Auth with local JWT validation (ES256) and Supabase API as source of truth.
- **Dependency Injection:** Custom, thread-safe lazy-loading container.
- **Frontend:** Next.js (under development, separate repository).

## 🏗 Project Structure (Backend)

High-level structure of this repository:

- `cmd/app` – Application entry point and configuration wiring.
- `config/postgres/init.sql` – Database schema and extensions.
- `internal/core` – Domain model, business rules, and Hexagonal ports (interfaces).
- `internal/api` – HTTP handlers, DTOs, middleware, router, and DI container.
- `internal/infra` – Infrastructure adapters (Postgres, Supabase Storage, security, logging).
- `docs/` – Documentation (including sprint planning for the AI analysis pipeline).

For a more detailed explanation of each layer and how they interact, see [ARCHITECTURE.md](./ARCHITECTURE.md).

## 🚦 Getting Started

### Prerequisites

- Docker and Docker Compose
- Go 1.21+ (if running locally)
- Supabase project (Auth + Storage enabled)

### Configuration

1. **Environment Variables**
   Copy the example environment file and fill in your Supabase credentials and database settings:
   ```bash
   cp .env.example .env
   ```

### Running the Project

#### Option A: Full Docker Environment (Recommended)
This will spin up both the PostgreSQL database and the Go API. The `init.sql` script will run automatically on the first startup to set up the schema.
```bash
docker compose up -d
```

#### Option B: Local API with Docker Database
If you want to run the API locally for development/debugging while keeping the database in a container:
1. Start only the database:
   ```bash
   docker compose up -d postgres
   ```
2. Run the API from the root:
   ```bash
   go run cmd/app/main.go
   ```

### Testing Endpoints

Use a tool like Insomnia/Postman to call the API. The default port is usually `:8080` (check your `.env`).
- **Health Check:** `GET /health`
- **Auth:** Login/Register via Supabase integration.
- **Lesson Plans:** Upload and list pedagogical documents.

## 📈 Roadmap (Backend Focus)

- [ ] **AI Analysis Pipeline**
    - Asynchronous job processing for lesson plan analysis.
    - Text extraction from PDF/DOCX and LLM-based feedback.
- [ ] **Lesson Plan Analysis Storage**
    - Persist structured AI output and human-readable feedback.
- [ ] **Coordinator Dashboard API**
    - Endpoints for activities, agenda, and school-level overviews.
- [ ] **Student & Performance Module**
    - APIs for students, classes, attendance, and performance data.
- [ ] **Frontend Integration**
    - Connect the Next.js frontend to the existing backend APIs.

## 📬 Contact

Developed by **Misael Lima**
Pedagogical Coordinator & Backend Developer
📧 [misael.alisson14@gmail.com](mailto:misael.alisson14@gmail.com)

---