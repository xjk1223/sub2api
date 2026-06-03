-- Migration 145: daily check-in feature (checkin_logs table + default settings)

CREATE TABLE IF NOT EXISTS "checkin_logs" (
    "id"               bigserial     NOT NULL,
    "user_id"          bigint        NOT NULL,
    "amount"           decimal(20,8) NOT NULL DEFAULT 0,
    "consecutive_days" integer       NOT NULL DEFAULT 1,
    "checkin_date"     date          NOT NULL,
    "created_at"       timestamptz   NOT NULL DEFAULT now(),
    PRIMARY KEY ("id"),
    CONSTRAINT "checkin_logs_user_id_checkin_date_key" UNIQUE ("user_id", "checkin_date")
);

CREATE INDEX IF NOT EXISTS "checkinlog_user_id"      ON "checkin_logs" ("user_id");
CREATE INDEX IF NOT EXISTS "checkinlog_checkin_date" ON "checkin_logs" ("checkin_date");

-- default settings
--   checkin_enabled            : feature on/off
--   checkin_base_amount        : fixed reward per check-in
--   checkin_consecutive_bonus  : enable streak bonus
--   checkin_bonus_per_day      : extra reward per consecutive day
--   checkin_max_bonus_days     : cap on bonus days
INSERT INTO "settings" ("key", "value", "updated_at") VALUES
    ('checkin_enabled',           'true', now()),
    ('checkin_base_amount',       '0.1',  now()),
    ('checkin_consecutive_bonus', 'true', now()),
    ('checkin_bonus_per_day',     '0.05', now()),
    ('checkin_max_bonus_days',    '7',    now())
ON CONFLICT ("key") DO NOTHING;
