CREATE TABLE public.api_tokens (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    hashed_token text NOT NULL,
    is_active boolean default true not null,
    user_id uuid NOT NULL,
    uses bigint NOT NULL DEFAULT 0,
    credits_spent bigint NOT NULL DEFAULT 0,
    name text NOT NULL,
    short_string text NOT NULL,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

ALTER TABLE ONLY public.api_tokens
    ADD CONSTRAINT api_tokens_pkey PRIMARY KEY (id);

CREATE trigger handle_updated_at before
UPDATE
    ON public.api_tokens FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.generations add column api_token_id uuid;
ALTER TABLE public.upscales add column api_token_id uuid;

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_api_tokens_generations FOREIGN KEY (api_token_id) REFERENCES public.api_tokens(id) ON DELETE CASCADE;

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_api_tokens_upscales FOREIGN KEY (api_token_id) REFERENCES public.api_tokens(id) ON DELETE CASCADE;
