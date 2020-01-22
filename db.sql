CREATE SCHEMA forum;
--
-- PostgreSQL database dump
--

-- Dumped from database version 10.10
-- Dumped by pg_dump version 10.10

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

ALTER SCHEMA forum OWNER TO postgres;

SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: forum; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum.forum (
                             slug text NOT NULL,
                             threads integer DEFAULT 0 NOT NULL,
                             posts integer DEFAULT 0 NOT NULL,
                             title text NOT NULL,
                             "user" text NOT NULL
);


ALTER TABLE forum.forum OWNER TO postgres;

--
-- Name: forum_user; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum.forum_user (
                                  forum text NOT NULL,
                                  "user" text NOT NULL
);


ALTER TABLE forum.forum_user OWNER TO postgres;

--
-- Name: post; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum.post (
                            id integer NOT NULL,
                            author text NOT NULL,
                            created text NOT NULL,
                            forum text NOT NULL,
                            is_edited boolean DEFAULT false NOT NULL,
                            message text NOT NULL,
                            parent integer DEFAULT 0 NOT NULL,
                            thread integer NOT NULL,
                            path bigint[] DEFAULT '{0}'::bigint[] NOT NULL
);


ALTER TABLE forum.post OWNER TO postgres;

--
-- Name: post_id_seq; Type: SEQUENCE; Schema: forum; Owner: postgres
--

CREATE SEQUENCE forum.post_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE forum.post_id_seq OWNER TO postgres;

--
-- Name: post_id_seq; Type: SEQUENCE OWNED BY; Schema: forum; Owner: postgres
--

ALTER SEQUENCE forum.post_id_seq OWNED BY forum.post.id;


--
-- Name: thread; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum.thread (
                              id integer NOT NULL,
                              author text NOT NULL,
                              created timestamp with time zone DEFAULT now() NOT NULL,
                              forum text NOT NULL,
                              message text NOT NULL,
                              slug text,
                              title text NOT NULL,
                              votes integer DEFAULT 0
);


ALTER TABLE forum.thread OWNER TO postgres;

--
-- Name: thread_id_seq; Type: SEQUENCE; Schema: forum; Owner: postgres
--

CREATE SEQUENCE forum.thread_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE forum.thread_id_seq OWNER TO postgres;

--
-- Name: thread_id_seq; Type: SEQUENCE OWNED BY; Schema: forum; Owner: postgres
--

ALTER SEQUENCE forum.thread_id_seq OWNED BY forum.thread.id;


--
-- Name: user; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum."user" (
                              nick_name text COLLATE "C" NOT NULL,
                              email text NOT NULL,
                              full_name text NOT NULL,
                              about text
);


ALTER TABLE forum."user" OWNER TO postgres;

--
-- Name: vote; Type: TABLE; Schema: forum; Owner: postgres
--

CREATE TABLE forum.vote (
                            "user" text,
                            voice integer NOT NULL,
                            thread_id integer NOT NULL
);


ALTER TABLE forum.vote OWNER TO postgres;

--
-- Name: post id; Type: DEFAULT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.post ALTER COLUMN id SET DEFAULT nextval('forum.post_id_seq'::regclass);


--
-- Name: thread id; Type: DEFAULT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.thread ALTER COLUMN id SET DEFAULT nextval('forum.thread_id_seq'::regclass);


--
-- Name: post_id_seq; Type: SEQUENCE SET; Schema: forum; Owner: postgres
--

SELECT pg_catalog.setval('forum.post_id_seq', 11, true);


--
-- Name: thread_id_seq; Type: SEQUENCE SET; Schema: forum; Owner: postgres
--

SELECT pg_catalog.setval('forum.thread_id_seq', 11, true);


--
-- Name: forum forum_pk; Type: CONSTRAINT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.forum
    ADD CONSTRAINT forum_pk PRIMARY KEY (slug);


--
-- Name: forum_user forum_user_pk; Type: CONSTRAINT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.forum_user
    ADD CONSTRAINT forum_user_pk PRIMARY KEY ("user", forum);


--
-- Name: post post_pk; Type: CONSTRAINT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.post
    ADD CONSTRAINT post_pk PRIMARY KEY (id);


--
-- Name: thread thread_pk; Type: CONSTRAINT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum.thread
    ADD CONSTRAINT thread_pk PRIMARY KEY (id);


--
-- Name: user user_pk; Type: CONSTRAINT; Schema: forum; Owner: postgres
--

ALTER TABLE ONLY forum."user"
    ADD CONSTRAINT user_pk PRIMARY KEY (nick_name);


--
-- Name: forum_slug_uindex; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE UNIQUE INDEX forum_slug_uindex ON forum.forum USING btree (lower(slug));


--
-- Name: post_author_forum_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX post_author_forum_index ON forum.post USING btree (lower(author), lower(forum));


--
-- Name: post_forum_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX post_forum_index ON forum.post USING btree (lower(forum));


--
-- Name: post_parent_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX post_parent_index ON forum.post USING btree (parent);


--
-- Name: post_path_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX post_path_index ON forum.post USING gin (path);


--
-- Name: post_thread_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX post_thread_index ON forum.post USING btree (thread);


--
-- Name: thread_forum_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX thread_forum_index ON forum.thread USING btree (lower(forum));


--
-- Name: thread_id_uindex; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE UNIQUE INDEX thread_id_uindex ON forum.thread USING btree (id);


--
-- Name: thread_slug_index; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE INDEX thread_slug_index ON forum.thread USING btree (lower(slug));


--
-- Name: user_email_uindex; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE UNIQUE INDEX user_email_uindex ON forum."user" USING btree (lower(email));


--
-- Name: user_nick_name_uindex; Type: INDEX; Schema: forum; Owner: postgres
--

CREATE UNIQUE INDEX user_nick_name_uindex ON forum."user" USING btree (lower(nick_name));


--
-- Name: SCHEMA forum; Type: ACL; Schema: -; Owner: postgres
--

GRANT ALL ON SCHEMA forum TO PUBLIC;


--
-- PostgreSQL database dump complete
--

