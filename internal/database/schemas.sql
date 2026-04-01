-- 1. Таблица ролей
CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Таблица школ
CREATE TABLE IF NOT EXISTS schools (
    id SERIAL PRIMARY KEY,
    name VARCHAR(256) NOT NULL,
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Таблица диапазонов операндов (вынесена в отдельную таблицу)
CREATE TABLE IF NOT EXISTS operand_ranges (
    id SERIAL PRIMARY KEY,
    equation_type_id INTEGER NOT NULL REFERENCES equation_types(id) ON DELETE CASCADE,
    operand_order INTEGER NOT NULL CHECK (operand_order >= 1), -- Порядок операнда (1, 2, 3, ...)
    min_value INTEGER DEFAULT 0,
    max_value INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(equation_type_id, operand_order)
);

-- 3. Таблица типов уравнений (обновленная структура)
CREATE TABLE IF NOT EXISTS equation_types (
    id SERIAL PRIMARY KEY,
    class INTEGER NOT NULL, -- 3, 4 и т.д.
    name VARCHAR(100) NOT NULL, -- Человекочитаемое название
    description TEXT, -- Например: "Сложение/вычитание двузначных с однозначными"
    
    -- Поля для генерации
    operation VARCHAR(10) NOT NULL, -- '+', '-', '*', '/', '+-' (значит, случайный выбор + или -)
    num_operands INTEGER NOT NULL DEFAULT 2, -- Количество чисел в выражении (2, 3, 4, ...)
    
    -- Специальные условия
    no_remainder BOOLEAN DEFAULT FALSE, -- Для деления без остатка
    result_max INTEGER DEFAULT NULL, -- Ограничение на результат (например, "до 90")
    
    is_available BOOLEAN DEFAULT TRUE, -- Доступен ли тип уравнений
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Таблица пользователей
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    role_id INTEGER NOT NULL REFERENCES roles(id) DEFAULT 1,
    fullname VARCHAR(256) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Таблица классов
CREATE TABLE IF NOT EXISTS classes (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    grade INTEGER,
    teacher_id INTEGER REFERENCES users(id),
    school_id INTEGER REFERENCES schools(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 4. Связь учеников с классами
CREATE TABLE IF NOT EXISTS student_classes (
    student_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    class_id INTEGER REFERENCES classes(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (student_id, class_id)
);

-- 6. Таблица попыток
CREATE TABLE IF NOT EXISTS attempts (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,

    -- Тип уравнения 
    equation_type_id INTEGER REFERENCES equation_types(id),
    
    -- Само уравнение и ответы
    equation_text TEXT NOT NULL,
    correct_answer VARCHAR(50) NOT NULL,
    user_answer VARCHAR(50),
    is_correct BOOLEAN,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Таблица прогресса пользователя по типам уравнений
CREATE TABLE IF NOT EXISTS user_progress (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    equation_type_id INTEGER REFERENCES equation_types(id),
    
    -- Статистика по конкретному типу
    attempts_count INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
  
    -- Разблокирован ли этот тип для пользователя
    is_unlocked BOOLEAN DEFAULT FALSE,
    
    -- Когда впервые открыли и последняя активность
    first_unlocked_at TIMESTAMP,
    last_attempt_at TIMESTAMP,
    
    UNIQUE(user_id, equation_type_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. Таблица сессий для авторизации
CREATE TABLE IF NOT EXISTS user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(100) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_attempts_user_id ON attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_attempts_equation_type_id ON attempts(equation_type_id);
CREATE INDEX IF NOT EXISTS idx_attempts_created_at ON attempts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_is_correct ON attempts(is_correct);
CREATE INDEX IF NOT EXISTS idx_user_progress_user_id ON user_progress(user_id);
CREATE INDEX IF NOT EXISTS idx_user_progress_type_id ON user_progress(equation_type_id);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_classes_school_id ON classes(school_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(session_token);

-- Заполнение ролей
INSERT INTO roles (name, description) VALUES
('student', 'Ученик'),
('teacher', 'Учитель'),
('admin', 'Администратор'),
('director', 'Директор');

-- Заполнение типов уравнений
-- Сначала вставляем типы уравнений без диапазонов операндов
INSERT INTO equation_types 
(class, name, description, operation, num_operands, result_max, no_remainder, is_available) VALUES
-- 3 класс (основные)
(3, 'Сложение/вычитание (2-знач. с 1-знач.)', 'До 90', '+-', 2, 90, FALSE, TRUE),
(3, 'Сложение/вычитание (2-знач. с 2-знач.)', 'До 50', '+-', 2, NULL, FALSE, TRUE),
(3, 'Умножение (2-знач. на 1-знач.)', 'До 100', '*', 2, 100, FALSE, TRUE),
(3, 'Деление (без остатка)', 'До 100', '/', 2, NULL, TRUE, TRUE),

-- 3 класс (будущие расширения - пока is_active = FALSE)
(3, 'Выражение из 3 операндов', 'До 33', '+-*/', 3, 100, FALSE, FALSE),
(3, 'Выражение из 4 операндов', 'До 20', '+-*/', 4, 100, FALSE, FALSE),
(3, 'Сложение/вычитание (3-знач. с 3-знач.)', 'До 1000', '+-', 2, 1000, FALSE, FALSE),
(3, 'Сложение/вычитание (3-знач. с 2-знач.)', 'До 1000', '+-', 2, 1000, FALSE, FALSE),
(3, 'Умножение (3-знач. на 1-знач.)', 'До 1000', '*', 2, 1000, FALSE, FALSE),
(3, 'Деление (3-знач. на 1-знач.)', 'До 1000', '/', 2, 1000, FALSE, FALSE),

-- 4 класс
(4, 'Сложение/вычитание (3-знач. с 3-знач.)', 'До 500', '+-', 2, 500, FALSE, TRUE),
(4, 'Умножение (3-знач. на 1-знач.)', 'До 500', '*', 2, 500, FALSE, TRUE),

-- 4 класс (будущие расширения)
(4, 'Выражение из 3 чисел', 'До 333', '+-*/', 3, 1000, FALSE, FALSE);

-- Теперь добавляем диапазоны операндов для каждого типа уравнения
-- ID 1: Сложение/вычитание (2-знач. с 1-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(1, 1, 10, 90),  -- первый операнд: 10-90
(1, 2, 1, 9);    -- второй операнд: 1-9

-- ID 2: Сложение/вычитание (2-знач. с 2-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(2, 1, 10, 50),
(2, 2, 10, 50);

-- ID 3: Умножение (2-знач. на 1-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(3, 1, 10, 99),
(3, 2, 2, 9);

-- ID 4: Деление (без остатка)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(4, 1, 10, 100),
(4, 2, 2, 10);

-- ID 5: Выражение из 3 операндов
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(5, 1, 1, 33),
(5, 2, 1, 33),
(5, 3, 1, 33);

-- ID 6: Выражение из 4 операндов
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(6, 1, 1, 20),
(6, 2, 1, 20),
(6, 3, 1, 20),
(6, 4, 1, 20);

-- ID 7: Сложение/вычитание (3-знач. с 3-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(7, 1, 100, 1000),
(7, 2, 100, 1000);

-- ID 8: Сложение/вычитание (3-знач. с 2-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(8, 1, 100, 1000),
(8, 2, 10, 100);

-- ID 9: Умножение (3-знач. на 1-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(9, 1, 100, 999),
(9, 2, 2, 9);

-- ID 10: Деление (3-знач. на 1-знач.)
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(10, 1, 100, 999),
(10, 2, 2, 9);

-- ID 11: Сложение/вычитание (3-знач. с 3-знач.) 4 класс
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(11, 1, 100, 500),
(11, 2, 100, 500);

-- ID 12: Умножение (3-знач. на 1-знач.) 4 класс
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(12, 1, 100, 500),
(12, 2, 2, 9);

-- ID 13: Выражение из 3 чисел 4 класс
INSERT INTO operand_ranges (equation_type_id, operand_order, min_value, max_value) VALUES
(13, 1, 100, 333),
(13, 2, 100, 333),
(13, 3, 100, 333);

-- Функция для обработки создания ученика
CREATE OR REPLACE FUNCTION create_user_progress_for_new_student()
RETURNS TRIGGER AS $$
BEGIN
    -- Для учеников НЕ создаем прогресс при регистрации
    -- Прогресс будет создаваться, когда ученика добавят в класс
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на вставку пользователя (ученика)
CREATE TRIGGER trigger_create_user_progress
AFTER INSERT ON users
FOR EACH ROW
WHEN (NEW.role_id = 1)
EXECUTE FUNCTION create_user_progress_for_new_student();

-- Функция для обработки добавления ученика в класс
CREATE OR REPLACE FUNCTION create_progress_when_student_added_to_class()
RETURNS TRIGGER AS $$
DECLARE
    class_grade INTEGER;
BEGIN
    -- Получаем уровень (grade) класса
    SELECT grade INTO class_grade FROM classes WHERE id = NEW.class_id;
    
    -- Для каждого типа уравнения, который соответствует уровню класса
    INSERT INTO user_progress (user_id, equation_type_id, is_unlocked, first_unlocked_at)
    SELECT NEW.student_id, et.id, et.is_avalable, CURRENT_TIMESTAMP
    FROM equation_types et
    WHERE et.class = class_grade
    ON CONFLICT (user_id, equation_type_id) DO NOTHING;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на добавление ученика в класс
CREATE TRIGGER trigger_create_progress_on_class_assignment
AFTER INSERT ON student_classes
FOR EACH ROW
EXECUTE FUNCTION create_progress_when_student_added_to_class();

-- Функция для обработки создания типа уравнения
CREATE OR REPLACE FUNCTION create_user_progress_for_new_equation_type()
RETURNS TRIGGER AS $$
BEGIN
    -- Для каждого ученика, который находится в классе с таким уровнем
    INSERT INTO user_progress (user_id, equation_type_id, is_unlocked, first_unlocked_at)
    SELECT sc.student_id, NEW.id, NEW.is_available, CURRENT_TIMESTAMP
    FROM student_classes sc
    JOIN classes c ON c.id = sc.class_id
    WHERE c.grade = NEW.class
      AND EXISTS (SELECT 1 FROM users u WHERE u.id = sc.student_id AND u.role_id = 1)
    ON CONFLICT (user_id, equation_type_id) DO NOTHING;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на вставку типа уравнения
CREATE TRIGGER trigger_create_user_progress_for_eq_type
AFTER INSERT ON equation_types
FOR EACH ROW
EXECUTE FUNCTION create_user_progress_for_new_equation_type();

CREATE OR REPLACE FUNCTION sync_equation_type_availability()
RETURNS TRIGGER AS $$
BEGIN
    -- Если изменился статус доступности
    IF OLD.is_available IS DISTINCT FROM NEW.is_available THEN
        -- Обновляем статус разблокировки у всех пользователей
        UPDATE user_progress 
        SET is_unlocked = NEW.is_available,
            updated_at = CURRENT_TIMESTAMP
        WHERE equation_type_id = NEW.id;
        
        -- Если тип стал доступным, устанавливаем дату первого разблокирования
        IF NEW.is_available = TRUE AND OLD.is_available = FALSE THEN
            UPDATE user_progress 
            SET first_unlocked_at = CURRENT_TIMESTAMP
            WHERE equation_type_id = NEW.id 
              AND first_unlocked_at IS NULL;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 2. Триггер на обновление equation_types
CREATE TRIGGER trigger_sync_equation_type_availability
AFTER UPDATE ON equation_types
FOR EACH ROW
WHEN (OLD.is_available IS DISTINCT FROM NEW.is_available)
EXECUTE FUNCTION sync_equation_type_availability();
