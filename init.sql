CREATE TABLE authors (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL
);

CREATE TABLE publishers (
    id SERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    city VARCHAR(100)
);

CREATE TABLE books (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    isbn VARCHAR(20) UNIQUE NOT NULL,
    publication_year INT,
    publisher_id INT REFERENCES publishers(id)
);

CREATE TABLE book_authors (
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    author_id INT REFERENCES authors(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, author_id)
);

CREATE TABLE genres (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE book_genres (
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    genre_id INT REFERENCES genres(id) ON DELETE CASCADE,
    PRIMARY KEY (book_id, genre_id)
);

CREATE TABLE readers (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(200) NOT NULL,
    email VARCHAR(150) UNIQUE NOT NULL,
    phone VARCHAR(20)
);

CREATE TABLE loans (
    id SERIAL PRIMARY KEY,
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    reader_id INT REFERENCES readers(id) ON DELETE CASCADE,
    loan_date DATE NOT NULL DEFAULT CURRENT_DATE,
    return_date DATE
);

-- Добавляем тестовые данные
INSERT INTO genres (name) VALUES 
('Программирование'), ('Фантастика'), ('Детектив'), ('Классика'), ('Фэнтези'), ('Научпоп'), ('Роман'), ('Антиутопия'), ('Триллер');

INSERT INTO publishers (name, city) VALUES 
('O-Reilly Media', 'Себастопол'), ('Манн, Иванов и Фербер', 'Москва'), ('АСТ', 'Москва'), ('Эксмо', 'Москва'), ('Азбука', 'Москва');

INSERT INTO authors (first_name, last_name) VALUES 
('Алан', 'Донован'), ('Брайан', 'Керниган'), ('Джордж', 'Оруэлл'), ('Агата', 'Кристи'), 
('Стивен', 'Кинг'), ('Михаил', 'Булгаков'), ('Джон', 'Толкин'), ('Рэй', 'Брэдбери'), 
('Айзек', 'Азимов'), ('Артур', 'Конан Дойл');

INSERT INTO books (title, isbn, publication_year, publisher_id) VALUES 
('Язык программирования Go', '978-5-8459-2051-5', 2016, 1),
('1984', '978-5-17-094451-6', 1949, 3),
('Убийство в "Восточном экспрессе"', '978-5-699-12345-6', 1934, 4),
('Сияние', '978-5-17-111111-1', 1977, 3),
('Мастер и Маргарита', '978-5-17-222222-2', 1967, 3),
('Властелин колец: Братство кольца', '978-5-17-333333-3', 1954, 3),
('451 градус по Фаренгейту', '978-5-699-44444-4', 1953, 4),
('Основание', '978-5-699-55555-5', 1951, 4),
('Собака Баскервилей', '978-5-389-66666-6', 1902, 5),
('Зеленая миля', '978-5-17-777777-7', 1996, 3),
('Десять негритят', '978-5-699-88888-8', 1939, 4),
('Скотный двор', '978-5-17-999999-9', 1945, 3),
('Марсианские хроники', '978-5-699-00000-0', 1950, 4),
('Хоббит, или Туда и обратно', '978-5-17-101010-1', 1937, 3),
('Я, робот', '978-5-699-20202-2', 1950, 4);

-- Связываем книги и авторов
INSERT INTO book_authors (book_id, author_id) VALUES 
(1, 1), (1, 2), (2, 3), (3, 4), (4, 5), (5, 6), (6, 7), 
(7, 8), (8, 9), (9, 10), (10, 5), (11, 4), (12, 3), (13, 8), (14, 7), (15, 9);

-- Связываем книги и жанры
INSERT INTO book_genres (book_id, genre_id) VALUES 
(1, 1), 
(2, 2), (2, 8), 
(3, 3), 
(4, 9), 
(5, 4), (5, 5), (5, 7),
(6, 5), 
(7, 2), (7, 8),
(8, 2), 
(9, 3), (9, 4),
(10, 7), (10, 9),
(11, 3), (11, 9),
(12, 8), (12, 4),
(13, 2), 
(14, 5), 
(15, 2);

-- Добавляем 50 тестовых читателей (клиентов)
INSERT INTO readers (full_name, email, phone) VALUES 
('Иванов Иван Иванович', 'ivan.ivanov@example.com', '+79001234501'),
('Петров Петр Петрович', 'petr.petrov@example.com', '+79001234502'),
('Смирнов Алексей Сергеевич', 'alex.smirnov@example.com', '+79001234503'),
('Кузнецов Дмитрий Александрович', 'd.kuznetsov@example.com', '+79001234504'),
('Попов Сергей Николаевич', 's.popov@example.com', '+79001234505'),
('Васильев Андрей Викторович', 'a.vasiliev@example.com', '+79001234506'),
('Соколов Михаил Владимирович', 'm.sokolov@example.com', '+79001234507'),
('Михайлов Александр Евгеньевич', 'a.mikhailov@example.com', '+79001234508'),
('Новиков Илья Дмитриевич', 'i.novikov@example.com', '+79001234509'),
('Федоров Максим Игоревич', 'm.fedorov@example.com', '+79001234510'),
('Морозов Владислав Олегович', 'v.morozov@example.com', '+79001234511'),
('Волков Никита Денисович', 'n.volkov@example.com', '+79001234512'),
('Алексеев Роман Антонович', 'r.alekseev@example.com', '+79001234513'),
('Лебедев Артем Павлович', 'a.lebedev@example.com', '+79001234514'),
('Семенов Егор Романович', 'e.semenov@example.com', '+79001234515'),
('Егоров Даниил Степанович', 'd.egorov@example.com', '+79001234516'),
('Павлов Кирилл Вячеславович', 'k.pavlov@example.com', '+79001234517'),
('Козлов Тимур Борисович', 't.kozlov@example.com', '+79001234518'),
('Степанов Марк Леонидович', 'm.stepanov@example.com', '+79001234519'),
('Николаев Глеб Аркадьевич', 'g.nikolaev@example.com', '+79001234520'),
('Орлов Лев Тимурович', 'l.orlov@example.com', '+79001234521'),
('Андреев Макар Константинович', 'm.andreev@example.com', '+79001234522'),
('Макаров Давид Артурович', 'd.makarov@example.com', '+79001234523'),
('Никитин Арсений Эдуардович', 'a.nikitin@example.com', '+79001234524'),
('Захаров Платон Русланович', 'p.zakharov@example.com', '+79001234525'),
('Иванова Мария Ивановна', 'm.ivanova@example.com', '+79001234526'),
('Петрова Анна Петровна', 'a.petrova@example.com', '+79001234527'),
('Смирнова Елена Сергеевна', 'e.smirnova@example.com', '+79001234528'),
('Кузнецова Дарья Александровна', 'd.kuznetsova@example.com', '+79001234529'),
('Попова Екатерина Николаевна', 'e.popova@example.com', '+79001234530'),
('Васильева Ольга Викторовна', 'o.vasilieva@example.com', '+79001234531'),
('Соколова Наталья Владимировна', 'n.sokolova@example.com', '+79001234532'),
('Михайлова Татьяна Евгеньевна', 't.mikhailova@example.com', '+79001234533'),
('Новикова Анастасия Дмитриевна', 'a.novikova@example.com', '+79001234534'),
('Федорова Юлия Игоревна', 'yu.fedorova@example.com', '+79001234535'),
('Морозова Ирина Олеговна', 'i.morozova@example.com', '+79001234536'),
('Волкова Светлана Денисовна', 's.volkova@example.com', '+79001234537'),
('Алексеева Ксения Антоновна', 'k.alekseeva@example.com', '+79001234538'),
('Лебедева Полина Павловна', 'p.lebedeva@example.com', '+79001234539'),
('Семенова Виктория Романовна', 'v.semenova@example.com', '+79001234540'),
('Егорова Алиса Степановна', 'a.egorova@example.com', '+79001234541'),
('Павлова Маргарита Вячеславовна', 'm.pavlova@example.com', '+79001234542'),
('Козлова Вера Борисовна', 'v.kozlova@example.com', '+79001234543'),
('Степанова Надежда Леонидовна', 'n.stepanova@example.com', '+79001234544'),
('Николаева Любовь Аркадьевна', 'l.nikolaeva@example.com', '+79001234545'),
('Орлова Алина Тимуровна', 'a.orlova@example.com', '+79001234546'),
('Андреева Елизавета Константиновна', 'e.andreeva@example.com', '+79001234547'),
('Макарова София Артуровна', 's.makarova@example.com', '+79001234548'),
('Никитина Валерия Эдуардовна', 'v.nikitina@example.com', '+79001234549'),
('Захарова Милана Руслановна', 'm.zakharova@example.com', '+79001234550');