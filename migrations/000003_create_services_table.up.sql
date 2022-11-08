CREATE TABLE IF NOT EXISTS public.services
(
    id bigint NOT NULL,
    title character varying COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT services_pkey PRIMARY KEY (id)
);