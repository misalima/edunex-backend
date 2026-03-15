-- =====================================================
-- EXTENSIONS
-- =====================================================

CREATE
    EXTENSION IF NOT EXISTS "pgcrypto";

-- =====================================================
-- ENUM TYPES
-- =====================================================

CREATE TYPE role_enum AS ENUM ('admin', 'coordinator', 'principal');

CREATE TYPE shift_enum AS ENUM (
    'morning',
    'afternoon',
    'evening'
    );

CREATE TYPE gender_enum AS ENUM (
    'male',
    'female',
    'other'
    );

CREATE TYPE attendance_status_enum AS ENUM (
    'present',
    'absent',
    'justified'
    );

CREATE TYPE term_enum AS ENUM (
    '1',
    '2',
    '3',
    '4'
    );

CREATE TYPE priority_enum AS ENUM (
    'high',
    'medium',
    'low'
    );

CREATE TYPE notification_status_enum AS ENUM (
    'pending',
    'sent',
    'read'
    );

CREATE TYPE activity_status_enum AS ENUM (
    'pending',
    'in_progress',
    'done'
    );

-- Fila de processamento IA
CREATE TYPE job_status_enum AS ENUM (
    'pending',
    'processing',
    'done',
    'failed'
    );

CREATE TYPE lesson_plan_status AS ENUM (
    'pending',
    'approved',
    'needs_adjustment'
    );

-- =====================================================
-- CORE TABLES
-- =====================================================

CREATE TABLE schools
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name       TEXT NOT NULL,
    address    TEXT,
    inep       TEXT,
    image_url  TEXT,
    created_at TIMESTAMP        DEFAULT now(),
    updated_at TIMESTAMP        DEFAULT now()
);

CREATE TABLE users
(
    id         UUID PRIMARY KEY     DEFAULT gen_random_uuid(),
    name       TEXT        NOT NULL,
    email      TEXT UNIQUE NOT NULL,
    role       role_enum   NOT NULL DEFAULT 'coordinator',
    created_at TIMESTAMP            DEFAULT now(),
    updated_at TIMESTAMP            DEFAULT now()
);

CREATE TABLE users_schools
(
    user_id   UUID NOT NULL,
    school_id UUID NOT NULL,

    PRIMARY KEY (user_id, school_id),

    FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    FOREIGN KEY (school_id)
        REFERENCES schools (id)
        ON DELETE CASCADE
);

-- =====================================================
-- EDUCATION STRUCTURE
-- =====================================================

CREATE TABLE education_levels
(
    id   SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE academic_classes
(
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name               TEXT       NOT NULL,
    year               INT        NOT NULL,

    education_level_id INT        NOT NULL,

    grade_level        TEXT       NOT NULL,

    shift              shift_enum NOT NULL,

    school_id          UUID       NOT NULL,

    created_at         TIMESTAMP        DEFAULT now(),
    updated_at         TIMESTAMP        DEFAULT now(),

    FOREIGN KEY (school_id)
        REFERENCES schools (id)
        ON DELETE CASCADE,

    FOREIGN KEY (education_level_id)
        REFERENCES education_levels (id)
        ON DELETE CASCADE
);

CREATE TABLE disciplines
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name      TEXT NOT NULL,

    school_id UUID NOT NULL,

    FOREIGN KEY (school_id)
        REFERENCES schools (id)
        ON DELETE CASCADE
);

CREATE TABLE teachers
(
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name      TEXT NOT NULL,

    phone     TEXT,
    email     TEXT,

    school_id UUID NOT NULL,

    FOREIGN KEY (school_id)
        REFERENCES schools (id)
        ON DELETE CASCADE
);

CREATE TABLE classes_disciplines_teachers
(
    class_id      UUID NOT NULL,
    discipline_id UUID NOT NULL,
    teacher_id    UUID NOT NULL,

    PRIMARY KEY (
                 class_id,
                 discipline_id,
                 teacher_id
        ),

    FOREIGN KEY (class_id)
        REFERENCES academic_classes (id)
        ON DELETE CASCADE,

    FOREIGN KEY (discipline_id)
        REFERENCES disciplines (id)
        ON DELETE CASCADE,

    FOREIGN KEY (teacher_id)
        REFERENCES teachers (id)
        ON DELETE CASCADE
);

-- =====================================================
-- STUDENTS
-- =====================================================

CREATE TABLE students
(
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    name                TEXT        NOT NULL,

    registration_number TEXT UNIQUE NOT NULL,

    phone               TEXT,
    email               TEXT,

    birth_date          DATE,

    gender              gender_enum NOT NULL,

    class_id            UUID        NOT NULL,

    FOREIGN KEY (class_id)
        REFERENCES academic_classes (id)
        ON DELETE CASCADE
);

CREATE TABLE attendances
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    student_id UUID                   NOT NULL,

    date       DATE                   NOT NULL,

    status     attendance_status_enum NOT NULL,

    FOREIGN KEY (student_id)
        REFERENCES students (id)
        ON DELETE CASCADE
);

CREATE TABLE performance
(
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    student_id     UUID      NOT NULL,

    term           term_enum NOT NULL,

    grade          DOUBLE PRECISION,
    recovery_grade DOUBLE PRECISION,

    FOREIGN KEY (student_id)
        REFERENCES students (id)
        ON DELETE CASCADE
);

-- =====================================================
-- ACTIVITIES
-- =====================================================

CREATE TABLE activities
(
    id         UUID PRIMARY KEY              DEFAULT gen_random_uuid(),

    user_id    UUID                 NOT NULL,

    school_id  UUID                 NOT NULL,

    title      TEXT                 NOT NULL,

    date       DATE                 NOT NULL,

    status     activity_status_enum NOT NULL DEFAULT 'pending',

    created_at TIMESTAMP                     DEFAULT now(),
    updated_at TIMESTAMP                     DEFAULT now(),

    FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    FOREIGN KEY (school_id)
        REFERENCES schools (id)
        ON DELETE CASCADE
);

-- =====================================================
-- NOTIFICATIONS
-- =====================================================

CREATE TABLE notifications
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id    UUID                     NOT NULL,

    title      TEXT                     NOT NULL,

    message    TEXT                     NOT NULL,

    type       TEXT,

    status     notification_status_enum NOT NULL,

    created_at TIMESTAMP        DEFAULT now(),

    sent_at    TIMESTAMP,

    FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

-- =====================================================
-- LESSON PLANS (FILES)
-- =====================================================

CREATE TABLE lesson_plans
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    user_id    UUID NOT NULL,

    title      TEXT NOT NULL,

    file_path  TEXT NOT NULL,

    teacher    TEXT,

    discipline TEXT,

    grade_level TEXT,

    status     lesson_plan_status NOT NULL DEFAULT 'pending',

    created_at TIMESTAMP        DEFAULT now(),

    FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE
);

-- =====================================================
-- ANALYSIS JOBS (QUEUE)
-- =====================================================

CREATE TABLE analysis_jobs
(
    id             UUID PRIMARY KEY         DEFAULT gen_random_uuid(),

    lesson_plan_id UUID            NOT NULL UNIQUE,

    status         job_status_enum NOT NULL DEFAULT 'pending',

    attempts       INT             NOT NULL DEFAULT 0,

    error_message  TEXT,

    created_at     TIMESTAMP                DEFAULT now(),
    started_at     TIMESTAMP,
    finished_at    TIMESTAMP,

    FOREIGN KEY (lesson_plan_id)
        REFERENCES lesson_plans (id)
        ON DELETE CASCADE
);

-- =====================================================
-- ANALYSIS RESULTS
-- =====================================================

CREATE TABLE lesson_plan_analyses
(
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    lesson_plan_id UUID NOT NULL UNIQUE,

    analysis_text  TEXT NOT NULL,

    created_at     TIMESTAMP        DEFAULT now(),

    FOREIGN KEY (lesson_plan_id)
        REFERENCES lesson_plans (id)
        ON DELETE CASCADE
);

-- =====================================================
-- INDEXES
-- =====================================================

CREATE INDEX idx_users_schools_user_id
    ON users_schools (user_id);

CREATE INDEX idx_users_schools_school_id
    ON users_schools (school_id);

CREATE INDEX idx_academic_classes_school_id
    ON academic_classes (school_id);

CREATE INDEX idx_disciplines_school_id
    ON disciplines (school_id);

CREATE INDEX idx_teachers_school_id
    ON teachers (school_id);

CREATE INDEX idx_students_class_id
    ON students (class_id);

CREATE INDEX idx_attendances_student_id
    ON attendances (student_id);

CREATE INDEX idx_performance_student_id
    ON performance (student_id);

CREATE INDEX idx_activities_user_id
    ON activities (user_id);

CREATE INDEX idx_notifications_user_id
    ON notifications (user_id);

CREATE INDEX idx_lesson_plans_user_id
    ON lesson_plans (user_id);

CREATE INDEX idx_analysis_jobs_status
    ON analysis_jobs (status);

CREATE INDEX idx_analysis_jobs_created_at
    ON analysis_jobs (created_at);

CREATE INDEX idx_lesson_plan_analyses_lesson_plan_id
    ON lesson_plan_analyses (lesson_plan_id);

-- =====================================================
-- UPDATED_AT TRIGGER FUNCTION
-- =====================================================

CREATE
    OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at
        = NOW();
    RETURN NEW;
END;
$$
    LANGUAGE plpgsql;

-- =====================================================
-- TRIGGERS
-- =====================================================

CREATE TRIGGER update_schools_updated_at
    BEFORE UPDATE
    ON schools
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_academic_classes_updated_at
    BEFORE UPDATE
    ON academic_classes
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_activities_updated_at
    BEFORE UPDATE
    ON activities
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();