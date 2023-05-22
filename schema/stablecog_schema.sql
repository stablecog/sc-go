CREATE extension IF NOT EXISTS moddatetime schema extensions;
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" schema extensions;

--
-- Name: generate_upscale_status_enum; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.generate_upscale_status_enum AS ENUM (
    'queued',
    'started',
    'succeeded',
    'failed'
);


ALTER TYPE public.generate_upscale_status_enum OWNER TO postgres;

--
-- Name: user_role_names_enum; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.user_role_names_enum AS ENUM (
    'SUPER_ADMIN',
    'GALLERY_ADMIN'
);


ALTER TYPE public.user_role_names_enum OWNER TO postgres;

--
-- Name: generation_output_gallery_status_enum; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.generation_output_gallery_status_enum AS ENUM (
    'not_submitted',
    'submitted',
    'approved',
    'rejected'
);

ALTER TYPE public.generation_output_gallery_status_enum OWNER TO postgres;

--
-- Name: credit_type_enum; Type: TYPE; Schema: public; Owner: postgres
--
CREATE TYPE public.credit_type_enum AS ENUM (
    'free',
    'subscription',
    'one_time'
);

ALTER TYPE public.credit_type_enum OWNER TO postgres;

--
-- Name: api_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_tokens (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    hashed_token text NOT NULL,
    is_active boolean default true not null,
    uses bigint NOT NULL DEFAULT 0,
    credits_spent bigint NOT NULL DEFAULT 0,
    user_id uuid NOT NULL,
    name text NOT NULL,
    short_string text NOT NULL,
    last_used_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.api_tokens FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);


--
-- Name: credit_types; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.credit_types (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    type public.credit_type_enum NOT NULL,
    description text,
    amount integer NOT NULL,
    stripe_product_id text,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);


CREATE trigger handle_updated_at before
UPDATE
    ON public.credit_types FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.credit_types OWNER TO postgres;

--
-- Name: credits; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.credits (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    remaining_amount integer NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    credit_type_id uuid NOT NULL,
    user_id uuid NOT NULL,
    stripe_line_item_id character varying,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    replenished_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);


CREATE trigger handle_updated_at before
UPDATE
    ON public.credits FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.credits OWNER TO postgres;

--
-- Name: device_info; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.device_info (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    type text,
    os text,
    browser text,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.device_info FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.device_info OWNER TO postgres;

--
-- Name: generation_models; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.generation_models (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    name_in_worker text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    is_hidden boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.generation_models FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.generation_models OWNER TO postgres;

--
-- Name: generation_outputs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.generation_outputs (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    image_path text NOT NULL,
    upscaled_image_path text,
    generation_id uuid NOT NULL,
    gallery_status public.generation_output_gallery_status_enum DEFAULT 'not_submitted'::public.generation_output_gallery_status_enum NOT NULL,
    is_favorited DEFAULT false not null;
    has_embeddings DEFAULT false not null;
    deleted_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.generation_outputs FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.generation_outputs OWNER TO postgres;

--
-- Name: generations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.generations (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL,
    inference_steps integer NOT NULL,
    guidance_scale real NOT NULL,
    seed bigint NOT NULL,
    status public.generate_upscale_status_enum NOT NULL,
    failure_reason text,
    country_code text,
    init_image_url text,
    prompt_strength real,
    was_auto_submitted boolean DEFAULT false NOT NULL,
    num_outputs integer NOT NULL,
    nsfw_count integer DEFAULT 0 NOT NULL,
    stripe_product_id text,
    device_info_id uuid NOT NULL,
    model_id uuid NOT NULL,
    negative_prompt_id uuid,
    prompt_id uuid,
    scheduler_id uuid NOT NULL,
    user_id uuid NOT NULL,
    api_token_id uuid,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.generations FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.generations OWNER TO postgres;

--
-- Name: negative_prompts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.negative_prompts (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    text text NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);


CREATE trigger handle_updated_at before
UPDATE
    ON public.negative_prompts FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.negative_prompts OWNER TO postgres;

--
-- Name: prompts; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.prompts (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    text text NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.prompts FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);


ALTER TABLE public.prompts OWNER TO postgres;

--
-- Name: schedulers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schedulers (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    name_in_worker text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    is_hidden boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.schedulers FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.schedulers OWNER TO postgres;

--
-- Name: upscale_models; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.upscale_models (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    name_in_worker text NOT NULL,
    is_active boolean DEFAULT true NOT NULL,
    is_default boolean DEFAULT false NOT NULL,
    is_hidden boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.upscale_models FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);


ALTER TABLE public.upscale_models OWNER TO postgres;

--
-- Name: upscale_outputs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.upscale_outputs (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    image_path text NOT NULL,
    upscale_id uuid NOT NULL,
    deleted_at timestamp with time zone,
    input_image_url text,
    generation_output_id uuid,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.upscale_outputs FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.upscale_outputs OWNER TO postgres;

--
-- Name: upscales; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.upscales (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL,
    scale integer NOT NULL,
    country_code text,
    status public.generate_upscale_status_enum NOT NULL,
    system_generated boolean default false not null,
    failure_reason text,
    stripe_product_id text,
    device_info_id uuid NOT NULL,
    model_id uuid NOT NULL,
    user_id uuid NOT NULL,
    api_token_id uuid,
    started_at timestamp with time zone,
    completed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.upscales FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.upscales OWNER TO postgres;

--
-- Name: user_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.user_roles (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    role_name public.user_role_names_enum NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.user_roles FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.user_roles OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid NOT NULL,
    email text NOT NULL,
    stripe_customer_id text NOT NULL,
    active_product_id text,
    last_sign_in_at timestamp with time zone,
    banned_at timestamp with time zone,
    data_deleted_at timestamp with time zone,
    scheduled_for_deletion_on timestamp with time zone,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.users FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

--
-- Name: disposable_emails; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.disposable_emails (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    domain text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.disposable_emails FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);


-- To update last_sign_in_at

create or replace function handle_updated_user() returns trigger as $$ begin
update
    public.users
set
    last_sign_in_at = new.last_sign_in_at,
    email = new.email
where
    id = new.id;
return new;
end;
$$ language plpgsql security definer;

create trigger on_auth_user_updated
after
update
    on auth.users for each row execute procedure handle_updated_user();

ALTER TABLE public.users OWNER TO postgres;

--
-- Name: api_tokens api_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_tokens
    ADD CONSTRAINT api_tokens_pkey PRIMARY KEY (id);


--
-- Name: credit_types credit_types_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credit_types
    ADD CONSTRAINT credit_types_pkey PRIMARY KEY (id);


--
-- Name: credits credits_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_pkey PRIMARY KEY (id);


--
-- Name: device_info device_info_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.device_info
    ADD CONSTRAINT device_info_pkey PRIMARY KEY (id);


--
-- Name: generation_models generation_models_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generation_models
    ADD CONSTRAINT generation_models_pkey PRIMARY KEY (id);


--
-- Name: generation_outputs generation_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generation_outputs
    ADD CONSTRAINT generation_outputs_pkey PRIMARY KEY (id);


--
-- Name: generations generations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_pkey PRIMARY KEY (id);


--
-- Name: negative_prompts negative_prompts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.negative_prompts
    ADD CONSTRAINT negative_prompts_pkey PRIMARY KEY (id);


--
-- Name: prompts prompts_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.prompts
    ADD CONSTRAINT prompts_pkey PRIMARY KEY (id);


--
-- Name: schedulers schedulers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schedulers
    ADD CONSTRAINT schedulers_pkey PRIMARY KEY (id);


--
-- Name: upscale_models upscale_models_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscale_models
    ADD CONSTRAINT upscale_models_pkey PRIMARY KEY (id);


--
-- Name: upscale_outputs upscale_outputs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscale_outputs
    ADD CONSTRAINT upscale_outputs_pkey PRIMARY KEY (id);


--
-- Name: upscales upscales_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_pkey PRIMARY KEY (id);


--
-- Name: users user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT user_pkey PRIMARY KEY (id);

--
-- Name: users user_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.disposable_emails
    ADD CONSTRAINT disposable_emails_pkey PRIMARY KEY (id);


--
-- Name: user_roles user_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_pkey PRIMARY KEY (id);


--
-- Name: credit_expires_at_user_id_remaining_amount; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX credit_expires_at_user_id_remaining_amount ON public.credits USING btree (expires_at, user_id, remaining_amount);


--
-- Name: credit_stripe_line_item_id_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX credit_stripe_line_item_id_key ON public.credits USING btree (stripe_line_item_id);

--
-- Name: credit_types_name_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX credit_types_name_key ON public.credit_types USING btree (name);


--
-- Name: generation_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generation_created_at ON public.generations USING btree (created_at);

--
-- Name: generation_updated_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generation_updated_at ON public.generations USING btree (updated_at);


--
-- Name: generationoutput_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generationoutput_created_at ON public.generation_outputs USING btree (created_at);

--
-- Name: generationoutput_generation_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generationoutput_generation_id ON public.generation_outputs USING btree (generation_id);

--
-- Name: generationoutput_updated_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generationoutput_updated_at ON public.generation_outputs USING btree (updated_at);

--
-- Name: generationoutput_gallery_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generationoutput_gallery_status ON public.generation_outputs USING btree (gallery_status);

--
-- Name: generation_user_id_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generation_user_id_created_at ON public.generations USING btree (user_id, created_at);


--
-- Name: generation_user_id_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generation_user_id_status ON public.generations USING btree (user_id, status);


--
-- Name: generation_user_id_status_created_at; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generation_user_id_status_created_at ON public.generations USING btree (user_id, status, created_at);


--
-- Name: generationoutput_id_gallery_status; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX generationoutput_id_gallery_status ON public.generation_outputs USING btree (id, gallery_status);


--
-- Name: upscale_outputs_generation_output_id_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX upscale_outputs_generation_output_id_key ON public.upscale_outputs USING btree (generation_output_id);

--
-- Name: disposable_emails_domain_key; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX disposable_emails_domain_key ON public.disposable_emails USING btree (domain);

--
-- Name: upscaleoutput_image_path; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX upscaleoutput_image_path ON public.upscale_outputs USING btree (image_path);


--
-- Name: credits credits_credit_types_credits; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_credit_types_credits FOREIGN KEY (credit_type_id) REFERENCES public.credit_types(id) ON DELETE CASCADE;


--
-- Name: credits credits_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.credits
    ADD CONSTRAINT credits_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: generation_outputs generation_outputs_generations_generation_outputs; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generation_outputs
    ADD CONSTRAINT generation_outputs_generations_generation_outputs FOREIGN KEY (generation_id) REFERENCES public.generations(id) ON DELETE CASCADE;


--
-- Name: generations generations_device_info_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_device_info_generations FOREIGN KEY (device_info_id) REFERENCES public.device_info(id) ON DELETE CASCADE;


--
-- Name: generations generations_generation_models_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_generation_models_generations FOREIGN KEY (model_id) REFERENCES public.generation_models(id) ON DELETE CASCADE;

--
-- Name: generations generations_api_tokens_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_api_tokens_generations FOREIGN KEY (api_token_id) REFERENCES public.api_tokens(id) ON DELETE CASCADE;

--
-- Name: generations generations_negative_prompts_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_negative_prompts_generations FOREIGN KEY (negative_prompt_id) REFERENCES public.negative_prompts(id) ON DELETE CASCADE;


--
-- Name: generations generations_prompts_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_prompts_generations FOREIGN KEY (prompt_id) REFERENCES public.prompts(id) ON DELETE CASCADE;


--
-- Name: generations generations_schedulers_generations; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_schedulers_generations FOREIGN KEY (scheduler_id) REFERENCES public.schedulers(id) ON DELETE CASCADE;


--
-- Name: generations generations_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.generations
    ADD CONSTRAINT generations_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: upscale_outputs upscale_outputs_generation_outputs_upscale_outputs; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscale_outputs
    ADD CONSTRAINT upscale_outputs_generation_outputs_upscale_outputs FOREIGN KEY (generation_output_id) REFERENCES public.generation_outputs(id) ON DELETE SET NULL;


--
-- Name: upscale_outputs upscale_outputs_upscales_upscale_outputs; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscale_outputs
    ADD CONSTRAINT upscale_outputs_upscales_upscale_outputs FOREIGN KEY (upscale_id) REFERENCES public.upscales(id) ON DELETE CASCADE;


--
-- Name: upscales upscales_device_info_upscales; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_device_info_upscales FOREIGN KEY (device_info_id) REFERENCES public.device_info(id) ON DELETE CASCADE;


--
-- Name: upscales upscales_upscale_models_upscales; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_upscale_models_upscales FOREIGN KEY (model_id) REFERENCES public.upscale_models(id) ON DELETE CASCADE;

--
-- Name: upscales upscales_api_tokens_upscales; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_api_tokens_upscales FOREIGN KEY (api_token_id) REFERENCES public.api_tokens(id) ON DELETE CASCADE;

--
-- Name: upscales upscales_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.upscales
    ADD CONSTRAINT upscales_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: user_roles user_roles_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.user_roles
    ADD CONSTRAINT user_roles_user_id_fkey FOREIGN KEY (user_id) REFERENCES auth.users(id);


--
-- Name: users users_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_id_fkey FOREIGN KEY (id) REFERENCES auth.users(id);


--
-- Name: users Users can select their own entry; Type: POLICY; Schema: public; Owner: postgres
--

CREATE POLICY "Users can select their own entry" ON public.users FOR SELECT USING ((auth.uid() = id));


--
-- Name: credit_types; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.credit_types ENABLE ROW LEVEL SECURITY;

--
-- Name: credits; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.credits ENABLE ROW LEVEL SECURITY;

--
-- Name: device_info; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.device_info ENABLE ROW LEVEL SECURITY;

--
-- Name: generation_models; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.generation_models ENABLE ROW LEVEL SECURITY;

--
-- Name: generation_outputs; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.generation_outputs ENABLE ROW LEVEL SECURITY;

--
-- Name: generations; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.generations ENABLE ROW LEVEL SECURITY;

--
-- Name: negative_prompts; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.negative_prompts ENABLE ROW LEVEL SECURITY;

--
-- Name: prompts; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.prompts ENABLE ROW LEVEL SECURITY;

--
-- Name: schedulers; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.schedulers ENABLE ROW LEVEL SECURITY;

--
-- Name: upscale_models; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.upscale_models ENABLE ROW LEVEL SECURITY;

--
-- Name: upscale_outputs; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.upscale_outputs ENABLE ROW LEVEL SECURITY;

--
-- Name: upscales; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.upscales ENABLE ROW LEVEL SECURITY;

--
-- Name: user_roles; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_roles ENABLE ROW LEVEL SECURITY;

--
-- Name: users; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.users ENABLE ROW LEVEL SECURITY;

--
-- Name: disposable_emails; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.disposable_emails ENABLE ROW LEVEL SECURITY;