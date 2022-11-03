CREATE TABLE IF NOT EXISTS public.logs
(
    id bigint NOT NULL DEFAULT nextval('id_sequence'::regclass),
    user_id bigint NOT NULL,
    date date NOT NULL,
    description character varying(100) COLLATE pg_catalog."default" NOT NULL,
    amount bigint NOT NULL,
    CONSTRAINT report_pkey PRIMARY KEY (id),
    CONSTRAINT "user" FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);