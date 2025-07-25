CREATE TABLE "users" (
    "id" bigserial PRIMARY KEY,
    "username" varchar NOT NULL UNIQUE,
    "email" varchar NOT NULL UNIQUE,
    "password_hash" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "rooms" (
    "id" bigserial PRIMARY KEY,
    "name" varchar NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

CREATE TABLE "messages" (
    "id" bigserial PRIMARY KEY,
    "room_id" bigint NOT NULL,
    "user_id" bigint NOT NULL,
    "content" text NOT NULL,
    "created_at" timestamptz NOT NULL DEFAULT (now())
);

ALTER TABLE "messages" ADD FOREIGN KEY ("room_id") REFERENCES "rooms" ("id");
ALTER TABLE "messages" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");

CREATE INDEX ON "rooms" ("name");
CREATE INDEX ON "messages" ("room_id");
CREATE INDEX ON "messages" ("user_id");