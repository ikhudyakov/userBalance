CREATE TABLE IF NOT EXISTS public.money_reserve_details
(
    id bigint NOT NULL DEFAULT nextval('id_sequence'::regclass),
    user_id bigint NOT NULL,
    service_id bigint NOT NULL,
    order_id bigint NOT NULL,
    amount bigint NOT NULL,
    date date NOT NULL,
    CONSTRAINT "moneyReserveAccount_pkey" PRIMARY KEY (id),
    CONSTRAINT service FOREIGN KEY (service_id)
        REFERENCES public.services (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID,
    CONSTRAINT "user" FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);