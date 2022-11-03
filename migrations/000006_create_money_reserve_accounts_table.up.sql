CREATE TABLE IF NOT EXISTS public.money_reserve_accounts
(
    user_id bigint NOT NULL,
    balance bigint NOT NULL DEFAULT 0,
    CONSTRAINT money_reserve_account_pkey PRIMARY KEY (user_id),
    CONSTRAINT "user" FOREIGN KEY (user_id)
        REFERENCES public.users (id) MATCH SIMPLE
        ON UPDATE NO ACTION
        ON DELETE NO ACTION
        NOT VALID
);