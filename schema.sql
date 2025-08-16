-- TODO: la capacité de traiter des fichiers qui sont composé de plusieurs fichiers (des images de scanner par exemple)

-- TODO: ajouter language_id au document

CREATE TABLE document(
id integer PRIMARY KEY,
created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
user_date TIMESTAMP,
sha256 bytes UNIQUE,
"binary" bytes,
filename string,
mime_type string,
content_length integer
);

CREATE TABLE tag(id integer PRIMARY KEY, created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP, name string UNIQUE);

CREATE TABLE tag_document(tag_id int, document_id int, UNIQUE(tag_id, document_id));

CREATE TABLE "language"(id integer PRIMARY KEY, name string UNIQUE);

CREATE TABLE document_inverted_index(document_id int, word string, positions bytes, UNIQUE(document_id, word));