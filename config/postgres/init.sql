-- Habilita extensão para UUIDs
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tipos ENUM
CREATE TYPE role_enum AS ENUM ('admin', 'coordinator', 'principal');
CREATE TYPE shift_enum AS ENUM ('morning', 'afternoon', 'evening');
CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other');
CREATE TYPE attendance_status_enum AS ENUM ('present', 'absent', 'justified');
CREATE TYPE term_enum AS ENUM ('1', '2', '3', '4');
CREATE TYPE priority_enum AS ENUM ('high', 'medium', 'low');
CREATE TYPE notification_status_enum AS ENUM ('pending', 'sent', 'read');
CREATE TYPE activity_status_enum AS ENUM ('pending', 'in_progress', 'done');
CREATE TYPE process_status_enum AS ENUM ('pending', 'in_progress', 'done');

-- Tabela de Escolas
CREATE TABLE schools (
                         id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                         name TEXT NOT NULL,
                         address TEXT,
                         inep TEXT,
                         image_url TEXT,
                         created_at TIMESTAMP DEFAULT now(),
                         updated_at TIMESTAMP DEFAULT now()
);

-- Tabela de Usuários
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       name TEXT NOT NULL,
                       email TEXT UNIQUE NOT NULL,
                       password TEXT NOT NULL,
                       role role_enum NOT NULL DEFAULT 'coordinator',
                       created_at TIMESTAMP DEFAULT now(),
                       updated_at TIMESTAMP DEFAULT now()
);

-- Associação Usuários-Escolas
CREATE TABLE users_schools (
                               user_id UUID NOT NULL,
                               school_id UUID NOT NULL,
                               PRIMARY KEY (user_id, school_id),
                               FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
                               FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE CASCADE
);

-- Níveis de Ensino
CREATE TABLE education_levels (
                                  id SERIAL PRIMARY KEY,
                                  name TEXT NOT NULL
);

-- Turmas Acadêmicas
CREATE TABLE academic_classes (
                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  name TEXT NOT NULL,
                                  year INT NOT NULL,
                                  education_level_id INT NOT NULL,
                                  grade_level TEXT NOT NULL,
                                  shift shift_enum NOT NULL,
                                  school_id UUID NOT NULL,
                                  created_at TIMESTAMP DEFAULT now(),
                                  updated_at TIMESTAMP DEFAULT now(),
                                  FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE CASCADE,
                                  FOREIGN KEY (education_level_id) REFERENCES education_levels (id) ON DELETE CASCADE
);

-- Disciplinas
CREATE TABLE disciplines (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             name TEXT NOT NULL,
                             school_id UUID NOT NULL,
                             FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE CASCADE
);

-- Professores
CREATE TABLE teachers (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name TEXT NOT NULL,
                          phone TEXT,
                          email TEXT,
                          school_id UUID NOT NULL,
                          FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE CASCADE
);

-- Associação Turmas-Disciplinas-Professores
CREATE TABLE classes_disciplines_teachers (
                                              class_id UUID NOT NULL,
                                              discipline_id UUID NOT NULL,
                                              teacher_id UUID NOT NULL,
                                              PRIMARY KEY (class_id, discipline_id, teacher_id),
                                              FOREIGN KEY (class_id) REFERENCES academic_classes (id) ON DELETE CASCADE,
                                              FOREIGN KEY (discipline_id) REFERENCES disciplines (id) ON DELETE CASCADE,
                                              FOREIGN KEY (teacher_id) REFERENCES teachers (id) ON DELETE CASCADE
);

-- Alunos
CREATE TABLE students (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name TEXT NOT NULL,
                          registration_number TEXT UNIQUE NOT NULL,
                          phone TEXT,
                          email TEXT,
                          birth_date DATE,
                          gender gender_enum NOT NULL,
                          class_id UUID NOT NULL,
                          FOREIGN KEY (class_id) REFERENCES academic_classes (id) ON DELETE CASCADE
);

-- Frequência
CREATE TABLE attendances (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             student_id UUID NOT NULL,
                             date DATE NOT NULL,
                             status attendance_status_enum NOT NULL,
                             FOREIGN KEY (student_id) REFERENCES students (id) ON DELETE CASCADE
);

-- Notas
CREATE TABLE performance (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             student_id UUID NOT NULL,
                             term term_enum NOT NULL,
                             grade DOUBLE PRECISION,
                             recovery_grade DOUBLE PRECISION,
                             FOREIGN KEY (student_id) REFERENCES students (id) ON DELETE CASCADE
);

-- Atividades (Agenda Simplificada)
CREATE TABLE activities (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            user_id UUID NOT NULL,
                            school_id UUID NOT NULL,
                            title TEXT NOT NULL,
                            date DATE NOT NULL,
                            status activity_status_enum NOT NULL DEFAULT 'pending',
                            created_at TIMESTAMP DEFAULT now(),
                            updated_at TIMESTAMP DEFAULT now(),
                            FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
                            FOREIGN KEY (school_id) REFERENCES schools (id) ON DELETE CASCADE
);

-- Notificações
CREATE TABLE notifications (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                               user_id UUID NOT NULL,
                               title TEXT NOT NULL,
                               message TEXT NOT NULL,
                               type TEXT,
                               status notification_status_enum NOT NULL,
                               created_at TIMESTAMP DEFAULT now(),
                               sent_at TIMESTAMP,
                               FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Planos de Aula
CREATE TABLE lesson_plans (
                              id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                              user_id UUID NOT NULL,
                              title TEXT NOT NULL,
                              file_path TEXT NOT NULL,
                              status process_status_enum NOT NULL DEFAULT 'pending',
                              created_at TIMESTAMP DEFAULT now(),
                              FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Análises dos Planos de Aula
CREATE TABLE lesson_plan_analyses (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      lesson_plan_id UUID NOT NULL,
                                      analysis_text TEXT,
                                      status process_status_enum NOT NULL DEFAULT 'pending',
                                      created_at TIMESTAMP DEFAULT now(),
                                      FOREIGN KEY (lesson_plan_id) REFERENCES lesson_plans (id) ON DELETE CASCADE
);

-- Recuperação de Senha
CREATE TABLE password_resets (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 user_id UUID NOT NULL,
                                 token TEXT NOT NULL UNIQUE,
                                 expires_at TIMESTAMP NOT NULL,
                                 created_at TIMESTAMP DEFAULT now(),
                                 used_at TIMESTAMP,
                                 FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Índice para busca rápida de tokens de reset
CREATE INDEX idx_password_resets_token ON password_resets(token);

-- Refresh Tokens
CREATE TABLE refresh_tokens (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID NOT NULL,
                                token TEXT NOT NULL UNIQUE,
                                expires_at TIMESTAMP NOT NULL,
                                created_at TIMESTAMP DEFAULT now(),
                                FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

-- Índice para busca rápida de tokens
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);

-- Índices para Foreign Keys
CREATE INDEX idx_users_schools_user_id ON users_schools(user_id);
CREATE INDEX idx_users_schools_school_id ON users_schools(school_id);

CREATE INDEX idx_academic_classes_school_id ON academic_classes(school_id);
CREATE INDEX idx_disciplines_school_id ON disciplines(school_id);
CREATE INDEX idx_teachers_school_id ON teachers(school_id);

CREATE INDEX idx_students_class_id ON students(class_id);

CREATE INDEX idx_attendances_student_id ON attendances(student_id);
CREATE INDEX idx_performance_student_id ON performance(student_id);

CREATE INDEX idx_activities_user_id ON activities(user_id);
CREATE INDEX idx_notifications_user_id ON notifications(user_id);

CREATE INDEX idx_lesson_plans_user_id ON lesson_plans(user_id);
CREATE INDEX idx_lesson_plan_analyses_lesson_plan_id ON lesson_plan_analyses(lesson_plan_id);

CREATE INDEX idx_password_resets_user_id ON password_resets(user_id);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Função para atualizar updated_at automaticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers para atualizar updated_at
CREATE TRIGGER update_schools_updated_at BEFORE UPDATE ON schools FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_academic_classes_updated_at BEFORE UPDATE ON academic_classes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_activities_updated_at BEFORE UPDATE ON activities FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insere o usuário admin (você mesmo)
INSERT INTO users (id, name, email, password, role, created_at, updated_at)
VALUES (
           '8fbc2c20-70e6-4a27-87b6-c26b18d42551',
           'Admin EduNex',
           'admin@edunex.com.br',
           '$2a$10$S3E7WwETLd7gh6JpgqdqVe2wTNurTYW9zxDryX/NM/DHmBAEoeIxS', -- Use bcrypt para gerar o hash
           'admin',
           NOW(),
           NOW()
       )
    ON CONFLICT (email) DO NOTHING;