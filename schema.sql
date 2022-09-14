CREATE TABLE folders (
    id TEXT,
    name TEXT
);

CREATE TABLE documents (
    id TEXT,
    name TEXT,
    parent TEXT
);

CREATE TABLE groups (
    id TEXT,
    name TEXT,
    member TEXT
);