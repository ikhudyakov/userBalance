CREATE TABLE IF NOT EXISTS public.users
(
    id bigint NOT NULL,
    balance bigint NOT NULL,
    CONSTRAINT users_pkey PRIMARY KEY (id)
);