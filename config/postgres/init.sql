-- Habilita extensão para UUIDs
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tipos ENUM
CREATE TYPE shift_enum AS ENUM ('morning', 'afternoon', 'evening');
CREATE TYPE gender_enum AS ENUM ('male', 'female', 'other');
CREATE TYPE attendance_status_enum AS ENUM ('present', 'absent', 'justified');
CREATE TYPE term_enum AS ENUM ('1', '2', '3', '4');
CREATE TYPE priority_enum AS ENUM ('high', 'medium', 'low');
CREATE TYPE notification_status_enum AS ENUM ('pending', 'sent', 'read');

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

-- Atividades
CREATE TABLE activities (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            user_id UUID NOT NULL,
                            school_id UUID NOT NULL,
                            title TEXT NOT NULL,
                            description TEXT,
                            priority priority_enum NOT NULL,
                            date DATE NOT NULL,
                            time TIME NOT NULL,
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
