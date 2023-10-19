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
CREATE TYPE public. AS ENUM (
    'free',
    'subscription',
    'one_time',
    'tippable'
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
    discord_id text,
    email text NOT NULL,
    stripe_customer_id text NOT NULL,
    active_product_id text,
    last_sign_in_at timestamp with time zone,
    banned_at timestamp with time zone,
    data_deleted_at timestamp with time zone,
    scheduled_for_deletion_on timestamp with time zone,
    wants_email boolean default false not null,
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

create index generation_user_id_idx on public.generations(user_id);
CREATE INDEX generations_status_idx ON generations (status);
CREATE INDEX generations_negative_prompt_id_idx ON generations (negative_prompt_id);
CREATE INDEX generations_prompt_id_idx ON generations (prompt_id);
CREATE INDEX generation_outputs_deleted_at_is_public_idx ON generation_outputs (deleted_at, is_public);
CREATE INDEX idx_generations_status_user_id ON generations (status, user_id);
CREATE INDEX idx_generation_outputs_generation_id_includes ON generation_outputs (generation_id) INCLUDE (deleted_at, is_public);
CREATE INDEX idx_generation_outputs_generation_id_is_public ON generation_outputs (generation_id, is_public);
CREATE INDEX generations_status_user_id_idx on public.generations(status, user_id);




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

CREATE INDEX generation_user_id_status_created_at ON public.generations USING btree (status, user_id, created_at);


drop index generation_user_id_status_created_at;
drop index generation_user_id_status;
drop index generation_user_id_created_at;

CREATE INDEX generation_user_id_status_created_at ON generations (user_id, created_at)
   WHERE deleted_at is null AND status='succeeded';


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

-- M2M tables
CREATE TABLE "public"."generation_model_compatible_schedulers"
  (
     "generation_model_id" UUID NOT NULL,
     "scheduler_id"        UUID NOT NULL,
     PRIMARY KEY ("generation_model_id", "scheduler_id"),
     CONSTRAINT "generation_model_compatible_schedulers_generation_model_id"
     FOREIGN KEY ("generation_model_id") REFERENCES "public"."generation_models"
     ("id") ON UPDATE no action ON DELETE CASCADE,
     CONSTRAINT "generation_model_compatible_schedulers_scheduler_id" FOREIGN
     KEY ("scheduler_id") REFERENCES "public"."schedulers" ("id") ON UPDATE no
     action ON DELETE CASCADE
  );

ALTER TABLE public.generation_model_compatible_schedulers ENABLE ROW LEVEL SECURITY;

ALTER TABLE public.generation_models add column default_scheduler_id UUID REFERENCES public.schedulers(id) ON DELETE SET NULL;
ALTER TABLE public.generation_models add column default_width INTEGER NOT NULL DEFAULT 512;
ALTER TABLE public.generation_models add column default_height INTEGER NOT NULL DEFAULT 512;

-- Create "roles" table
CREATE TABLE "public"."roles"
  (
     "id"         uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
     "name"       CHARACTER VARYING NOT NULL,
     PRIMARY KEY ("id")
  );

CREATE trigger handle_updated_at before
UPDATE
    ON public.roles FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

-- Create "user_role_users" table
CREATE TABLE "public"."user_role_users"
  (
     "role_model_id" UUID NOT NULL,
     "user_id"       UUID NOT NULL,
     PRIMARY KEY ("role_model_id", "user_id"),
     CONSTRAINT "user_role_users_role_model_id" FOREIGN KEY ("role_model_id")
     REFERENCES "public"."roles" ("id") ON UPDATE no action ON DELETE CASCADE,
     CONSTRAINT "user_role_users_user_id" FOREIGN KEY ("user_id") REFERENCES
     "auth"."users" ("id") ON UPDATE no action ON DELETE CASCADE
  ); 

--
-- Name: roles; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.roles ENABLE ROW LEVEL SECURITY;

--
-- Name: user_role_users; Type: ROW SECURITY; Schema: public; Owner: postgres
--

ALTER TABLE public.user_role_users ENABLE ROW LEVEL SECURITY;

-- Create "voiceover_models" table
CREATE TABLE "public"."voiceover_models"
  (
     "id"             UUID DEFAULT extensions.uuid_generate_v4() NOT NULL,
     "name_in_worker" TEXT NOT NULL,
     "is_active"      BOOLEAN NOT NULL DEFAULT TRUE,
     "is_default"     BOOLEAN NOT NULL DEFAULT FALSE,
     "is_hidden"      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
     PRIMARY KEY ("id")
  );

CREATE trigger handle_updated_at before
UPDATE
    ON public.voiceover_models FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

-- Create "voiceover_speakers" table
CREATE TABLE "public"."voiceover_speakers"
  (
     "id"             UUID DEFAULT extensions.uuid_generate_v4() NOT NULL,
     "name_in_worker" TEXT NOT NULL,
     "is_active"      BOOLEAN NOT NULL DEFAULT TRUE,
     "is_default"     BOOLEAN NOT NULL DEFAULT FALSE,
     "is_hidden"      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
     "model_id"       UUID NOT NULL,
     "locale"         TEXT NOT NULL DEFAULT 'en',
     PRIMARY KEY ("id"),
     CONSTRAINT "voiceover_speakers_voiceover_models_voiceover_speakers" FOREIGN
     KEY ("model_id") REFERENCES "public"."voiceover_models" ("id") ON UPDATE no
     action ON DELETE CASCADE
  );
alter table public.voiceover_speakers add column name text;

CREATE trigger handle_updated_at before
UPDATE
    ON public.voiceover_speakers FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

-- Create "voiceovers" table
CREATE TABLE "public"."voiceovers"
  (
     "id"                UUID DEFAULT extensions.uuid_generate_v4() NOT NULL,
     "country_code"      TEXT NULL,
     "status"            CHARACTER varying NOT NULL,
     "failure_reason"    TEXT NULL,
     "stripe_product_id" TEXT NULL,
     temperature real NOT NULL,
     seed bigint NOT NULL,
     was_auto_submitted boolean DEFAULT false NOT NULL,
     denoise_audio boolean DEFAULT true NOT NULL,
     remove_silence boolean DEFAULT true NOT NULL,
     "started_at"        TIMESTAMPTZ NULL,
     "completed_at"      TIMESTAMPTZ NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
     "api_token_id"      UUID NULL,
     "device_info_id"    UUID NOT NULL,
     "user_id"           UUID NOT NULL,
     "model_id"          UUID NOT NULL,
     "speaker_id"        UUID NOT NULL,
     "cost" integer NOT NULL,
     "prompt_id"         UUID constraint voiceovers_prompt_id_fk references prompts(id),
     PRIMARY KEY ("id"),
     CONSTRAINT "voiceovers_api_tokens_voiceovers" FOREIGN KEY ("api_token_id")
     REFERENCES "public"."api_tokens" ("id") ON UPDATE no action ON DELETE
     CASCADE,
     CONSTRAINT "voiceovers_device_info_voiceovers" FOREIGN KEY (
     "device_info_id") REFERENCES "public"."device_info" ("id") ON UPDATE no
     action ON DELETE CASCADE,
     CONSTRAINT "voiceovers_users_voiceovers" FOREIGN KEY ("user_id") REFERENCES
     "auth"."users" ("id") ON UPDATE no action ON DELETE CASCADE,
     CONSTRAINT "voiceovers_voiceover_models_voiceovers" FOREIGN KEY ("model_id"
     ) REFERENCES "public"."voiceover_models" ("id") ON UPDATE no action ON
     DELETE CASCADE,
     CONSTRAINT "voiceovers_voiceover_speakers_voiceovers" FOREIGN KEY (
     "speaker_id") REFERENCES "public"."voiceover_speakers" ("id") ON UPDATE no
     action ON DELETE CASCADE
  ); 

CREATE trigger handle_updated_at before
UPDATE
    ON public.voiceovers FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

-- Create "voiceover_outputs" table
CREATE TABLE "public"."voiceover_outputs"
  (
     "id"           UUID DEFAULT extensions.uuid_generate_v4() NOT NULL,
     "audio_path"   TEXT NOT NULL,
     "deleted_at"   TIMESTAMPTZ NULL,
     is_favorited boolean DEFAULT false NOT NULL,
     audio_duration real NOT NULL,
    gallery_status public.generation_output_gallery_status_enum DEFAULT 'not_submitted'::public.generation_output_gallery_status_enum NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
     "voiceover_id" UUID NOT NULL,
     PRIMARY KEY ("id"),
     CONSTRAINT "voiceover_outputs_voiceovers_voiceover_outputs" FOREIGN KEY (
     "voiceover_id") REFERENCES "public"."voiceovers" ("id") ON UPDATE no action
     ON DELETE CASCADE
  );
alter table public.voiceover_outputs ADD COLUMN "video_path" text NULL, ADD COLUMN "audio_array" jsonb NULL;

CREATE trigger handle_updated_at before
UPDATE
    ON public.voiceover_outputs FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

-- Create index "voiceoveroutput_audio_path" to table: "voiceover_outputs"
CREATE INDEX "voiceoveroutput_audio_path"
  ON "public"."voiceover_outputs" ("audio_path"); 

CREATE UNIQUE INDEX "voiceoverspeaker_name_in_worker_model_id" ON "public"."voiceover_speakers" ("name_in_worker", "model_id");

ALTER TABLE public.voiceover_models ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.voiceover_outputs ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.voiceovers ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.voiceover_speakers ENABLE ROW LEVEL SECURITY;
CREATE TYPE public.prompt_type_enum AS ENUM (
    'image',
    'voiceover'
);

ALTER TYPE public.prompt_type_enum OWNER TO postgres;

alter table prompts add column type public.prompt_type_enum not null default 'image';
alter table prompts alter column type drop default;

-- Add enum
create type public.operation_source_type_enum as enum (
    'web-ui',
    'api',
    'discord'
);
ALTER TYPE public.operation_source_type_enum OWNER TO postgres;

alter table public.generations add column source_type public.operation_source_type_enum DEFAULT 'web-ui'::public.operation_source_type_enum NOT NULL;
alter table public.upscales add column source_type public.operation_source_type_enum DEFAULT 'web-ui'::public.operation_source_type_enum NOT NULL;
alter table public.voiceovers add column source_type public.operation_source_type_enum DEFAULT 'web-ui'::public.operation_source_type_enum NOT NULL;

ALTER TYPE operation_source_type_enum ADD VALUE 'internal';

DROP index credit_stripe_line_item_id_key;
CREATE UNIQUE INDEX "credit_stripe_line_item_id_credit_type_id" ON "public"."credits" ("stripe_line_item_id", "credit_type_id");

CREATE UNIQUE INDEX "user_email_idx" ON "public"."users" ("email");

-- Create "tip_log" table
CREATE TABLE "public"."tip_log" (
    "id"                UUID DEFAULT extensions.uuid_generate_v4() NOT NULL,
  "amount" integer NOT NULL, 
  "tipped_by" uuid NOT NULL, 
  "tipped_to" uuid NULL, 
    "tipped_to_discord_id" text NOT NULL, 
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
  PRIMARY KEY ("id"), 
  CONSTRAINT "tip_log_users_tips_given" FOREIGN KEY ("tipped_by") REFERENCES "auth"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE, 
  CONSTRAINT "tip_log_users_tips_received" FOREIGN KEY ("tipped_to") REFERENCES "auth"."users" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.tip_log FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.tip_log ENABLE ROW LEVEL SECURITY;


-- Usernames
alter table public.users add column username text null;
CREATE UNIQUE INDEX users_username_key ON public.users USING btree (username);

-- Make usernames not null
alter table public.users alter column username set not null;

-- IP Blacklist

CREATE TABLE public.ip_blacklist (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    ip text NOT NULL,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.ip_blacklist FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.ip_blacklist ENABLE ROW LEVEL SECURITY;

ALTER TABLE ONLY public.ip_blacklist
    ADD CONSTRAINT ip_blacklist_pkey PRIMARY KEY (id);

-- Add username_changed_at
alter table public.users add column username_changed_at timestamp with time zone null;

-- add is_public
alter table public.generation_outputs add column is_public boolean not null default false;

-- Banned words

CREATE TABLE public.banned_words (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    words jsonb not null,
    reason text not null,
    split_match boolean not null default false,
    created_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at timestamp with time zone DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE trigger handle_updated_at before
UPDATE
    ON public.banned_words FOR each ROW EXECUTE PROCEDURE moddatetime (updated_at);

ALTER TABLE public.banned_words ENABLE ROW LEVEL SECURITY;

ALTER TABLE ONLY public.banned_words ADD CONSTRAINT banned_words_pkey PRIMARY KEY (id);

--
-- Name: auth_clients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.auth_clients (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    name text NOT NULL,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

ALTER TABLE ONLY public.auth_clients
    ADD CONSTRAINT auth_clients_pkey PRIMARY KEY (id);

ALTER TABLE public.api_tokens add column auth_client_id uuid;


ALTER TABLE ONLY public.api_tokens
    ADD CONSTRAINT api_tokens_auth_clients_api_tokens FOREIGN KEY (auth_client_id) REFERENCES public.auth_clients(id) ON DELETE CASCADE;

ALTER TABLE public.auth_clients ENABLE ROW LEVEL SECURITY;

--
-- Name: mq_log; Type: TABLE; Schema: public; Owner: postgres
--
CREATE TABLE public.mq_log (
    id uuid DEFAULT extensions.uuid_generate_v4() NOT NULL,
    message_id text not null unique,
    priority bigint NOT NULL,
    is_processing boolean default false not null,
    created_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL,
    updated_at timestamp with time zone DEFAULT (now() AT TIME ZONE 'utc'::text) NOT NULL
);

ALTER TABLE ONLY public.mq_log
    ADD CONSTRAINT mq_log_pkey PRIMARY KEY (id);

ALTER TABLE public.mq_log ENABLE ROW LEVEL SECURITY;
