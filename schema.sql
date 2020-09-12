--
-- PostgreSQL database dump
--

-- Dumped from database version 12.3
-- Dumped by pg_dump version 12.3

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: auth_keys; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auth_keys (
    id integer NOT NULL,
    auth_key text NOT NULL,
    user_id integer NOT NULL,
    created time with time zone NOT NULL,
    accessed time with time zone NOT NULL
);


--
-- Name: auth_keys_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.auth_keys ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.auth_keys_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: conversations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations (
    id integer NOT NULL,
    created timestamp with time zone NOT NULL
);


--
-- Name: conversations_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.conversations ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.conversations_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: conversations_users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.conversations_users (
    id integer NOT NULL,
    conversation_id integer NOT NULL,
    user_id integer NOT NULL
);


--
-- Name: conversations_users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.conversations_users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.conversations_users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: messages; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.messages (
    id integer NOT NULL,
    conversation_id integer NOT NULL,
    user_id integer NOT NULL,
    sent timestamp with time zone NOT NULL,
    body text NOT NULL
);


--
-- Name: messages_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.messages ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.messages_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id integer NOT NULL,
    email text NOT NULL,
    password text NOT NULL,
    name text DEFAULT 'Goomba'::text NOT NULL,
    created timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

ALTER TABLE public.users ALTER COLUMN id ADD GENERATED ALWAYS AS IDENTITY (
    SEQUENCE NAME public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1
);


--
-- Name: auth_keys auth_keys_auth_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_keys
    ADD CONSTRAINT auth_keys_auth_key_key UNIQUE (auth_key);


--
-- Name: auth_keys auth_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_keys
    ADD CONSTRAINT auth_keys_pkey PRIMARY KEY (id);


--
-- Name: conversations conversations_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations
    ADD CONSTRAINT conversations_pkey1 PRIMARY KEY (id);


--
-- Name: conversations_users conversations_users_conversation_id_user_id_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations_users
    ADD CONSTRAINT conversations_users_conversation_id_user_id_key UNIQUE (conversation_id, user_id);


--
-- Name: conversations_users conversations_users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations_users
    ADD CONSTRAINT conversations_users_pkey PRIMARY KEY (id);


--
-- Name: messages messages_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_pkey PRIMARY KEY (id);


--
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: auth_keys auth_keys_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auth_keys
    ADD CONSTRAINT auth_keys_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) MATCH FULL;


--
-- Name: conversations_users conversations_users_conversation_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations_users
    ADD CONSTRAINT conversations_users_conversation_id_fkey FOREIGN KEY (conversation_id) REFERENCES public.conversations(id) DEFERRABLE;


--
-- Name: conversations_users conversations_users_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.conversations_users
    ADD CONSTRAINT conversations_users_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id) DEFERRABLE;


--
-- Name: messages messages_conversation_id_fkey1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_conversation_id_fkey1 FOREIGN KEY (conversation_id) REFERENCES public.conversations(id);


--
-- Name: messages messages_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.messages
    ADD CONSTRAINT messages_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- PostgreSQL database dump complete
--

