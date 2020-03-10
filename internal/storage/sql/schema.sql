CREATE TABLE IF NOT EXISTS profiles (
    id uuid PRIMARY KEY NOT NULL,
    user_id VARCHAR(40) NOT NULL,
    pro_level INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS bookings (
    id uuid PRIMARY KEY NOT NULL,
    media_id VARCHAR(40) NOT NULL REFERENCES profiles(id),
    photographer_id VARCHAR(40) NOT NULL,
    task TEXT NOT NULL,
    price INTEGER NOT NULL CHECK (price >= 0),
    credits INTEGER NOT NULL,
    accepted BOOLEAN NOT NULL DEFAULT FALSE,
    completed BOOLEAN NOT NULL DEFAULT FALSE,
    date_start DATE NOT NULL,
    date_end DATE NOT NULL CHECK (date_end >= date_start),
    created_at DATE NOT NULL DEFAULT CURRENT_DATE,
    lat NUMERIC(6,9) NOT NULL,
    lng NUMERIC(6,9) NOT NULL
);
