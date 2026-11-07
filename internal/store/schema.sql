BEGIN TRANSACTION;
CREATE TABLE IF NOT EXISTS "document" (
	"id"	INTEGER,
	"name"	TEXT NOT NULL,
	"filename"	TEXT NOT NULL,
	"source_id"	INTEGER NOT NULL,
	PRIMARY KEY("id"),
	FOREIGN KEY("source_id") REFERENCES "source"("id")
) STRICT;

CREATE TABLE IF NOT EXISTS "source" (
	"id"	INTEGER NOT NULL,
	"hash"	TEXT UNIQUE,
	"status"	TEXT NOT NULL DEFAULT 'pending' CHECK("status" IN ('pending', 'processing', 'completed', 'failed')),
	"size"	INTEGER NOT NULL,
	"wordcount" INTEGER,
	"content"	BLOB NOT NULL,
	PRIMARY KEY("id")
) STRICT;

CREATE TABLE IF NOT EXISTS "token" (
	"id"	INTEGER PRIMARY KEY,
	"word"	TEXT UNIQUE,
	"folded"	TEXT
) STRICT;

CREATE INDEX IF NOT EXISTS idx_token_word_nocase ON token(word COLLATE NOCASE);
CREATE INDEX IF NOT EXISTS idx_token_folded ON token(folded);

CREATE TABLE IF NOT EXISTS "ngram1" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1", "source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram2" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram3" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram4" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram5" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram6" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"word6"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","word6","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram7" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"word6"	INTEGER,
	"word7"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","word6","word7","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram8" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"word6"	INTEGER,
	"word7"	INTEGER,
	"word8"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","word6","word7","word8","source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram9" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"word6"	INTEGER,
	"word7"	INTEGER,
	"word8"	INTEGER,
	"word9"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","word6","word7","word8","word9", "source_id")
) WITHOUT ROWID,STRICT;

CREATE TABLE IF NOT EXISTS "ngram10" (
	"source_id"	INTEGER,
	"word1"	INTEGER,
	"word2"	INTEGER,
	"word3"	INTEGER,
	"word4"	INTEGER,
	"word5"	INTEGER,
	"word6"	INTEGER,
	"word7"	INTEGER,
	"word8"	INTEGER,
	"word9"	INTEGER,
	"word10"	INTEGER,
	"occurrences" INTEGER NOT NULL DEFAULT 1,
	PRIMARY KEY("word1","word2","word3","word4","word5","word6","word7","word8","word9","word10","source_id")
) WITHOUT ROWID,STRICT;

COMMIT;
