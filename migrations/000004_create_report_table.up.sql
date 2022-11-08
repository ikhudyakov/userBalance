CREATE TABLE IF NOT EXISTS public.report
(
    id bigint NOT NULL DEFAULT nextval('id_sequence'::regclass),
    user_id bigint NOT NULL,
    service_id bigint NOT NULL,
    date date NOT NULL,
    amount bigint NOT NULL,
    CONSTRAINT log_pkey PRIMARY KEY (id),
    CONSTRAINT service FOREIGN KEY (service_id)
        REFERENCES public.services (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION,
    CONSTRAINT "user" FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
);