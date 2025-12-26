-- Table: books
CREATE TABLE IF NOT EXISTS books
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    title      VARCHAR(255) NOT NULL,
    author     VARCHAR(255) NOT NULL,
    stock      INT          NOT NULL DEFAULT 0,
    created_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP             DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

    -- MENGAPA index pada stock?
    -- Query "cek stok > 0" sangat sering, index mempercepat lookup
    INDEX idx_stock (stock)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

-- Table: members
CREATE TABLE IF NOT EXISTS members
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(255)        NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

-- Table: loans
CREATE TABLE IF NOT EXISTS loans
(
    id          INT AUTO_INCREMENT PRIMARY KEY,
    member_id   INT       NOT NULL,
    book_id     INT       NOT NULL,
    borrowed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP NULL,

    FOREIGN KEY (member_id) REFERENCES members (id) ON DELETE CASCADE,
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE CASCADE,

    -- MENGAPA composite index (member_id, returned_at)?
    -- Query "hitung pinjaman aktif member" sangat sering (validasi kuota)
    -- WHERE member_id = X AND returned_at IS NULL
    INDEX idx_member_active (member_id, returned_at),

    -- MENGAPA composite index (member_id, book_id, returned_at)?
    -- Query "cek apakah member sedang pinjam buku ini" untuk prevent double borrow
    -- WHERE member_id = X AND book_id = Y AND returned_at IS NULL
    INDEX idx_member_book_active (member_id, book_id, returned_at)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;

-- Seed Data: Books
INSERT INTO books (title, author, stock)
VALUES ('Clean Code', 'Robert C. Martin', 5),
       ('The Pragmatic Programmer', 'Andrew Hunt', 3),
       ('Design Patterns', 'Gang of Four', 2),
       ('Refactoring', 'Martin Fowler', 4),
       ('Head First Design Patterns', 'Eric Freeman', 1),
       ('Code Complete', 'Steve McConnell', 6),
       ('The Clean Coder', 'Robert C. Martin', 3),
       ('Working Effectively with Legacy Code', 'Michael Feathers', 2);

-- Seed Data: Members
INSERT INTO members (name, email)
VALUES ('John Doe', 'john@example.com'),
       ('Jane Smith', 'jane@example.com'),
       ('Bob Johnson', 'bob@example.com'),
       ('Alice Williams', 'alice@example.com'),
       ('Charlie Brown', 'charlie@example.com');

-- Seed Data: Sample Loans (untuk testing history)
INSERT INTO loans (member_id, book_id, borrowed_at, returned_at)
VALUES (1, 1, DATE_SUB(NOW(), INTERVAL 10 DAY), DATE_SUB(NOW(), INTERVAL 3 DAY)),
       (2, 2, DATE_SUB(NOW(), INTERVAL 7 DAY), NULL),
       (3, 3, DATE_SUB(NOW(), INTERVAL 5 DAY), NULL);