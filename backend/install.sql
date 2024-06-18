CREATE TABLE "scheduler" (
	"id"	INTEGER,
	"date"	INTEGER,
	"title"	TEXT NOT NULL,
	"comment"	TEXT,
	"repeat"	TEXT,
	CHECK(
        typeof("repeat") = "text" AND
        length("repeat") <= 128
    )
	PRIMARY KEY("id" AUTOINCREMENT)
);

CREATE INDEX "scheduler_date" ON "scheduler" (
	"date"	DESC
);