## Sprint: AI-Powered Lesson Plan Analysis
**Sprint Goal:** Implement a pedagogical analysis pipeline using LLMs.
**Sprint Start:** March 1, 2026 \
**Sprint End:** March 21, 2026

#### 1. Objective
Implement an asynchronous data pipeline to extract, structure, and analyze pedagogical content from uploaded lesson plans (PDF/DOCX) using LLM integration.

#### 2. Scope & User Story
**User Story:** "As a Pedagogical Coordinator, I want the system to automatically analyze my uploaded lesson plans so that I can verify alignment with the BNCC (National Common Curricular Base) and receive pedagogical improvement suggestions."

**Scope:**
- Text extraction from PDF and DOCX files.
- Single-call LLM integration using **Structured Output (JSON Mode)**.
- Asynchronous job management to prevent API timeouts.
- Persistence of structured data and pedagogical feedback.

---

#### 3. Technical Workflow (The Pipeline)
1. **Trigger:** A new `lesson_plan` is created -> An `analysis_job` is inserted with `status: pending`.
2. **Extraction:** The background worker picks the job -> Downloads the file from Supabase -> Extracts raw text.
3. **AI Processing:** Raw text is sent to the LLM with a specialized System Prompt.
4. **Structuring:** The LLM returns a single JSON containing both the structured metadata and the pedagogical analysis.
5. **Finalization:** Data is saved to `lesson_plan_analyses` -> Job status updated to `done`.

---

#### 4. Task Breakdown

**Layer: Infrastructure (`internal/infra`)**
- [ ] **File Extractor:** Implement `internal/infra/extractor` using libraries for PDF (`ledongthuc/pdf`) and DOCX (`unidoc/unioffice` or similar).
- [ ] **AI Client:** Implement `internal/infra/ai` to communicate with the LLM API (OpenAI/Gemini/Groq).
- [ ] **Structured Prompt:** Define the "Mega Prompt" that enforces JSON output mode.

**Layer: Core (`internal/core`)**
- [ ] **Analysis Service:** Create `internal/core/services/analysis_service.go` to orchestrate the extraction and AI calls.
- [ ] **Job Runner:** Implement a background worker (using Goroutines and a Ticker or a simple queue) to process `pending` jobs.

**Layer: API (`internal/api`)**
- [ ] **Status Endpoint:** Create `GET /lesson-plans/:id/analysis` to return the current job status and the final result.
- [ ] **Trigger Integration:** Ensure the `POST /lesson-plans` handler initiates the `analysis_job`.

---

#### 5. Data Contract (LLM JSON Schema)
The AI must return strictly the following structure:
```json
{
  "metadata": {
    "title": "string",
    "subject": "string",
    "grade_level": "string",
    "objectives": ["string"],
    "bncc_skills": ["string"]
  },
  "analysis": {
    "pedagogical_feedback": "string (markdown)",
    "alignment_score": "number (0-100)",
    "suggestions": ["string"],
    "missing_elements": ["string"]
  }
}
```

---

#### 6. Definition of Done (DoD)
- [ ] Text is successfully extracted from both PDF and DOCX formats.
- [ ] The AI call is handled asynchronously (no blocking the main request).
- [ ] Failed jobs are updated with an error message and incremented attempts.
- [ ] The final analysis is correctly linked to the `lesson_plan` in the database.
- [ ] API returns a valid JSON response with the analysis status.

---