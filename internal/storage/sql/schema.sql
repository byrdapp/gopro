CREATE TABLE profiles (
    id uuid PRIMARY KEY NOT NULL,
    user_id uuid NOT NULL,
    pro_level INTEGER NOT NULL
);

CREATE TABLE bookings (
    id uuid PRIMARY KEY NOT NULL,
    media_id uuid NOT NULL REFERENCES profile(id),
    photographer_id uuid NOT NULL,
    task TEXT NOT NULL,
    price BIGINT NOT NULL CHECK (price >= 0),
    credits INTEGER NOT NULL,
    accepted BOOLEAN NOT NULL DEFAULT FALSE,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    date_start DATE NOT NULL,
    date_end DATE NOT NULL CHECK (date_end >= date_start),
    created_at DATE NOT NULL DEFAULT CURRENT_DATE,
    lat NUMERIC(6,9) NOT NULL,
    lng NUMERIC(6,9) NOT NULL
);
