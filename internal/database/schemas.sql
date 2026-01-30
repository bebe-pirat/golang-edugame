-- 1. Таблица типов уравнений
CREATE TABLE equation_types (
    id SERIAL PRIMARY KEY,
    class INTEGER NOT NULL, -- 3, 4 и т.д.
    name VARCHAR(100) NOT NULL, -- Человекочитаемое название
    description TEXT, -- Например: "Сложение/вычитание двузначных с однозначными"
    
    -- Поля для генерации
    operation VARCHAR(10) NOT NULL, -- '+', '-', '*', '/', '+-' (значит, случайный выбор + или -)
    num_operands INTEGER NOT NULL DEFAULT 2, -- Количество чисел в выражении (2, 3, 4)
    
    -- ОГРАНИЧЕНИЯ НА ЧИСЛА (основной подход)
    -- Для каждого операнда задаем диапазон и дополнительные условия
    operand1_min INTEGER DEFAULT 0 NULL,
    operand1_max INTEGER DEFAULT 0 NULL,
    operand2_min INTEGER DEFAULT 0 NULL,
    operand2_max INTEGER DEFAULT 0 NULL,
    operand3_min INTEGER DEFAULT NULL, -- NULL, если операнда нет
    operand3_max INTEGER DEFAULT NULL,
    operand4_min INTEGER DEFAULT NULL,
    operand4_max INTEGER DEFAULT NULL,
    
    -- Специальные условия
    no_remainder BOOLEAN DEFAULT FALSE, -- Для деления без остатка
    result_max INTEGER DEFAULT NULL -- Ограничение на результат (например, "до 90")
);

-- Пример заполнения для ваших типов:
INSERT INTO equation_types 
(class, name, description, operation, num_operands, 
 operand1_min, operand1_max, operand2_min, operand2_max, result_max, no_remainder) VALUES

-- 3 класс
(3, 'Сложение/вычитание (2-знач. с 1-знач.)', 'До 90', '+-', 2, 10, 90, 1, 9, 100, FALSE),
(3, 'Сложение/вычитание (2-знач. с 2-знач.)', 'До 50', '+-', 2, 10, 90, 10, 90, 100, FALSE),
(3, 'Умножение (2-знач. на 1-знач.)', 'До 100', '*', 2, 10, 99, 2, 9, 100, FALSE),
(3, 'Деление (без остатка)', 'До 100', '/', 2, 10, 100, 2, 10, NULL, TRUE), -- Особый случай, генерируется через обратное умножение

-- 4 класс
(4, 'Сложение/вычитание (3-знач. с 3-знач.)', 'До 100', '+-', 2, 100, 999, 100, 999, 1000, FALSE),
(4, 'Умножение (3-знач. на 1-знач.)', 'До 500', '*', 2, 100, 500, 2, 9, 1000, FALSE);

-- 2. Таблица пользователей (остается почти без изменений)
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    class INTEGER DEFAULT 1,
    total_correct INTEGER DEFAULT 0,
    total_attempts INTEGER DEFAULT 0,
    
    -- Текущий прогресс по типам уравнений
    current_equation_type_id INTEGER REFERENCES equation_types(id),
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Таблица попыток
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
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
);

-- 4. Таблица сессий
CREATE TABLE IF NOT EXISTS sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. НОВАЯ: Таблица прогресса пользователя по типам уравнений
CREATE TABLE IF NOT EXISTS user_progress (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    equation_type_id INTEGER REFERENCES equation_types(id),
    
    -- Статистика по конкретному типу
    attempts_count INTEGER DEFAULT 0,
    correct_count INTEGER DEFAULT 0,
    best_time_ms INTEGER,
    
    -- Разблокирован ли этот тип для пользователя
    is_unlocked BOOLEAN DEFAULT FALSE,
    
    -- Когда впервые открыли и последняя активность
    first_unlocked_at TIMESTAMP,
    last_attempt_at TIMESTAMP,
    
    UNIQUE(user_id, equation_type_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Функция для обработки создания пользователя
CREATE OR REPLACE FUNCTION create_user_progress_for_new_user()
RETURNS TRIGGER AS $$
BEGIN
    -- Для каждого типа уравнения, который соответствует классу пользователя
    INSERT INTO user_progress (user_id, equation_type_id, is_unlocked)
    SELECT NEW.id, et.id, 
           CASE WHEN et.class <= NEW.class THEN TRUE ELSE FALSE END
    FROM equation_types et
    WHERE et.class = NEW.class  -- создаем только для совпадающих классов
    ON CONFLICT (user_id, equation_type_id) DO NOTHING;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на вставку пользователя
CREATE TRIGGER trigger_create_user_progress
AFTER INSERT ON users
FOR EACH ROW
EXECUTE FUNCTION create_user_progress_for_new_user();

-- Функция для обработки создания типа уравнения
CREATE OR REPLACE FUNCTION create_user_progress_for_new_equation_type()
RETURNS TRIGGER AS $$
BEGIN
    -- Для каждого пользователя с совпадающим классом
    INSERT INTO user_progress (user_id, equation_type_id, is_unlocked)
    SELECT u.id, NEW.id, 
           CASE WHEN NEW.class <= u.class THEN TRUE ELSE FALSE END
    FROM users u
    WHERE u.class = NEW.class  -- создаем только для совпадающих классов
    ON CONFLICT (user_id, equation_type_id) DO NOTHING;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер на вставку типа уравнения
CREATE TRIGGER trigger_create_user_progress_for_eq_type
AFTER INSERT ON equation_types
FOR EACH ROW
EXECUTE FUNCTION create_user_progress_for_new_equation_type();

-- Индексы для производительности
CREATE INDEX IF NOT EXISTS idx_attempts_user_id ON attempts(user_id);
CREATE INDEX IF NOT EXISTS idx_attempts_equation_type_id ON attempts(equation_type_id);
CREATE INDEX IF NOT EXISTS idx_attempts_created_at ON attempts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_attempts_is_correct ON attempts(is_correct);

INSERT INTO equation_types 
(class, name, description, operation, num_operands, 
 operand1_min, operand1_max, operand2_min, operand2_max, result_max, no_remainder, display_order) VALUES

-- 3 класс (основные)
(3, 'Сложение/вычитание (2-знач. с 1-знач.)', 'До 90', '+-', 2, 10, 90, 1, 9, 90, FALSE, 1),
(3, 'Сложение/вычитание (2-знач. с 2-знач.)', 'До 50', '+-', 2, 10, 50, 10, 50, NULL, FALSE, 2),
(3, 'Умножение (2-знач. на 1-знач.)', 'До 100', '*', 2, 10, 99, 2, 9, 100, FALSE, 3),
(3, 'Деление (без остатка)', 'До 100', '/', 2, 10, 100, 2, 10, NULL, TRUE, 4),

-- 3 класс (будущие расширения - пока is_active = FALSE)
(3, 'Выражение из 3 операндов', 'До 33', '+-*/', 3, 1, 33, 1, 33, 100, FALSE, 5),
(3, 'Выражение из 4 операндов', 'До 20', '+-*/', 4, 1, 20, 1, 20, 100, FALSE, 6),

-- 4 класс
(4, 'Сложение/вычитание (3-знач. с 3-знач.)', 'До 500', '+-', 2, 100, 500, 100, 500, 500, FALSE, 7),
(4, 'Умножение (3-знач. на 1-знач.)', 'До 500', '*', 2, 100, 500, 2, 9, 500, FALSE, 8),

-- 4 класс (будущие расширения)
(4, 'Выражение из 3 чисел', 'До 333', '+-*/', 3, 100, 333, 100, 333, 1000, FALSE, 9);

INSERT INTO users (username,  class, total_correct, total_attempts, current_equation_type_id)
VALUES ('тест', 3, 15, 20, 1);


INSERT INTO users (username,  class, total_correct, total_attempts, current_equation_type_id)
VALUES ('тест', 3, 15, 20, 1);
