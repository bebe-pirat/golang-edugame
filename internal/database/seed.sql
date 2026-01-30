-- Шаблоны для уровня 1 (простое сложение/вычитание)
INSERT INTO equation_templates (category, difficulty, template, min_value, max_value) VALUES
('addition', 1, '{a} + {b}', 1, 10),
('addition', 1, '{a} + {b} + {c}', 1, 5),
('subtraction', 1, '{a} - {b}', 1, 10),
('subtraction', 1, '{a} - {b} - {c}', 5, 15);

-- Шаблоны для уровня 2 (умножение, сложение 3 чисел)
INSERT INTO equation_templates (category, difficulty, template, min_value, max_value) VALUES
('multiplication', 2, '{a} × {b}', 1, 5),
('multiplication', 2, '{a} × {b} × {c}', 1, 3),
('addition', 2, '{a} + ? = {c}', 5, 20),
('subtraction', 2, '? - {b} = {c}', 10, 30);

-- Шаблоны для уровня 3 (деление, скобки)
INSERT INTO equation_templates (category, difficulty, template, min_value, max_value) VALUES
('division', 3, '{a} ÷ {b}', 4, 20),
('division', 3, '{a} ÷ ? = {c}', 10, 50),
('addition', 3, '({a} + {b}) × {c}', 1, 5),
('multiplication', 3, '{a} × ? = {c}', 2, 8);

-- Создаем анонимного пользователя для сессий
INSERT INTO users (username, current_level) VALUES ('anonymous', 1)
ON CONFLICT (username) DO NOTHING;