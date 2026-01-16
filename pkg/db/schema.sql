-- public.r4_mobile_payments definition
-- Drop table
-- DROP TABLE public.r4_mobile_payments;
CREATE TABLE IF NOT EXISTS public.r4_mobile_payments
(
    id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
    id_commerce varchar(255) NOT NULL,
    commerce_phone varchar(255) NOT NULL,
    sender_phone varchar(255) NOT NULL,
    -- concept varchar(255) NOT NULL,
    issuing_bank varchar(255) NOT NULL,
    amount decimal(10,2) NOT NULL,
    -- date_time varchar(255) NOT NULL,
    reference varchar(255) NOT NULL UNIQUE,
    -- red_code varchar(255) NOT NULL,
    order_id int4 UNIQUE,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    CONSTRAINT r4_mobile_payments_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_r4_mobile_payments_on_reference ON public.r4_mobile_payments (reference);
CREATE INDEX idx_r4_mobile_payments_on_order_id ON public.r4_mobile_payments (order_id);
CREATE INDEX idx_r4_mobile_payments_on_sender_phone ON public.r4_mobile_payments (sender_phone);

-- public.r4_mobile_payments_previews definition
-- Drop table
-- DROP TABLE public.r4_mobile_payments_previews;
CREATE TABLE IF NOT EXISTS public.r4_mobile_payments_previews
(
    id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
    amount decimal(10,2) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    CONSTRAINT r4_mobile_payments_previews_pkey PRIMARY KEY (id)
);

-- public.r4_appa_mobile_payments definition
-- Drop table
-- DROP TABLE public.r4_appa_mobile_payments;
CREATE TABLE IF NOT EXISTS public.r4_appa_mobile_payments
(
    id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
    id_commerce varchar(255) NOT NULL,
    commerce_phone varchar(255) NOT NULL,
    sender_phone varchar(255) NOT NULL,
    -- concept varchar(255) NOT NULL,
    issuing_bank varchar(255) NOT NULL,
    amount decimal(10,2) NOT NULL,
    -- date_time varchar(255) NOT NULL,
    reference varchar(255) NOT NULL UNIQUE,
    -- red_code varchar(255) NOT NULL,
    order_id int4 UNIQUE,
    date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    CONSTRAINT r4_appa_mobile_payments_pkey PRIMARY KEY (id)
);
CREATE INDEX idx_r4_appa_mobile_payments_on_reference ON public.r4_appa_mobile_payments (reference);
CREATE INDEX idx_r4_appa_mobile_payments_on_order_id ON public.r4_appa_mobile_payments (order_id);
CREATE INDEX idx_r4_appa_mobile_payments_on_sender_phone ON public.r4_appa_mobile_payments (sender_phone);

-- public.r4_appa_mobile_payments_previews definition
-- Drop table
-- DROP TABLE public.r4_appa_mobile_payments_previews;
CREATE TABLE IF NOT EXISTS public.r4_appa_mobile_payments_previews
(
    id int4 GENERATED ALWAYS AS IDENTITY( INCREMENT BY 1 MINVALUE 1 MAXVALUE 2147483647 START 1 CACHE 1 NO CYCLE) NOT NULL,
    amount decimal(10,2) NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW(),
    CONSTRAINT r4_appa_mobile_payments_previews_pkey PRIMARY KEY (id)
);