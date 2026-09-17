ALTER TABLE books
ADD CONSTRAINT books_title_not_empty
CHECK (
    LENGTH(TRIM(title)) > 0
);
ALTER TABLE authors
ADD CONSTRAINT authors_name_not_empty
CHECK (
    LENGTH(TRIM(name)) > 0
);