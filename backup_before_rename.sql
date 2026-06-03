--
-- PostgreSQL database dump
--

-- Dumped from database version 14.22 (Debian 14.22-1.pgdg13+1)
-- Dumped by pg_dump version 16.4

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

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

-- *not* creating schema, since initdb creates it


ALTER SCHEMA public OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.config (
    key character varying(50) NOT NULL,
    value text NOT NULL,
    description text
);


ALTER TABLE public.config OWNER TO postgres;

--
-- Name: debt_settlements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.debt_settlements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    debtor_id uuid NOT NULL,
    debt_amount bigint NOT NULL,
    money_amount bigint NOT NULL,
    to_fund numeric(12,2),
    to_winners numeric(12,2),
    winner_distribution bigint NOT NULL,
    settled_at timestamp with time zone DEFAULT now(),
    fund_amount bigint NOT NULL,
    original_debt_points bigint NOT NULL,
    settlement_date timestamp with time zone DEFAULT now(),
    created_at timestamp with time zone
);


ALTER TABLE public.debt_settlements OWNER TO postgres;

--
-- Name: fund_transactions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.fund_transactions (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    amount bigint NOT NULL,
    transaction_type character varying(20) NOT NULL,
    description text,
    related_settlement_id uuid,
    created_at timestamp with time zone,
    transaction_date timestamp with time zone DEFAULT now()
);


ALTER TABLE public.fund_transactions OWNER TO postgres;

--
-- Name: match_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.match_participants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    match_id uuid NOT NULL,
    user_id uuid NOT NULL,
    team_number bigint NOT NULL,
    point_change bigint NOT NULL
);


ALTER TABLE public.match_participants OWNER TO postgres;

--
-- Name: matches; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.matches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    match_type character varying(10) NOT NULL,
    winner_team bigint NOT NULL,
    match_date timestamp with time zone DEFAULT now(),
    recorded_by character varying(100),
    created_at timestamp with time zone,
    is_locked boolean DEFAULT false,
    tournament_match_id uuid
);


ALTER TABLE public.matches OWNER TO postgres;

--
-- Name: settlement_winners; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.settlement_winners (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    settlement_id uuid NOT NULL,
    winner_id uuid NOT NULL,
    money_amount bigint NOT NULL,
    points_deducted bigint NOT NULL
);


ALTER TABLE public.settlement_winners OWNER TO postgres;

--
-- Name: tournament_matches; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournament_matches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tournament_id uuid NOT NULL,
    round bigint NOT NULL,
    match_order bigint NOT NULL,
    team1_player1_id uuid NOT NULL,
    team1_player2_id uuid,
    team2_player1_id uuid NOT NULL,
    team2_player2_id uuid,
    handicap_team1 numeric DEFAULT 0,
    handicap_team2 numeric DEFAULT 0,
    status character varying(20) DEFAULT 'pending'::character varying,
    actual_score1 bigint,
    actual_score2 bigint,
    effective_winner bigint DEFAULT 0,
    match_id uuid,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.tournament_matches OWNER TO postgres;

--
-- Name: tournament_participants; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournament_participants (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tournament_id uuid NOT NULL,
    user_id uuid NOT NULL,
    tier_snapshot character varying(10) DEFAULT 'normal'::character varying,
    handicap_rate_snapshot numeric DEFAULT 0
);


ALTER TABLE public.tournament_participants OWNER TO postgres;

--
-- Name: tournaments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.tournaments (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(200) NOT NULL,
    match_type character varying(10) NOT NULL,
    status character varying(20) DEFAULT 'active'::character varying,
    affects_score boolean DEFAULT true,
    entry_fee bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.tournaments OWNER TO postgres;

--
-- Name: users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    current_score bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    is_active boolean DEFAULT true,
    tier character varying(10) DEFAULT 'normal'::character varying,
    handicap_rate numeric DEFAULT 0
);


ALTER TABLE public.users OWNER TO postgres;

--
-- Name: wc_bets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_bets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wc_user_id uuid NOT NULL,
    match_id uuid NOT NULL,
    bet_type character varying(15) NOT NULL,
    stake bigint NOT NULL,
    odds_snapshot numeric(5,2) NOT NULL,
    bet_choice character varying(5),
    handicap_snapshot numeric(4,1),
    handicap_team_snapshot character varying(5),
    predicted_home_score bigint,
    predicted_away_score bigint,
    result character varying(10),
    payout bigint,
    created_at timestamp with time zone
);


ALTER TABLE public.wc_bets OWNER TO postgres;

--
-- Name: wc_config; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_config (
    id bigint NOT NULL,
    is_enabled boolean DEFAULT false,
    updated_at timestamp with time zone,
    updated_by uuid
);


ALTER TABLE public.wc_config OWNER TO postgres;

--
-- Name: wc_matches; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_matches (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    external_id character varying(64),
    home_team character varying(100) NOT NULL,
    away_team character varying(100) NOT NULL,
    home_team_code character(3),
    away_team_code character(3),
    match_date timestamp with time zone NOT NULL,
    group_name character varying(30),
    stage character varying(30) DEFAULT 'group'::character varying NOT NULL,
    venue character varying(100),
    home_score bigint,
    away_score bigint,
    status character varying(20) DEFAULT 'scheduled'::character varying NOT NULL,
    handicap_team character varying(5),
    handicap_value numeric(4,1),
    odds_handicap_home numeric(5,2),
    odds_handicap_away numeric(5,2),
    betting_open boolean DEFAULT false NOT NULL,
    bets_locked_at timestamp with time zone,
    settled_at timestamp with time zone,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.wc_matches OWNER TO postgres;

--
-- Name: wc_score_odds; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_score_odds (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    match_id uuid NOT NULL,
    home_score bigint NOT NULL,
    away_score bigint NOT NULL,
    odds numeric(5,2) NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.wc_score_odds OWNER TO postgres;

--
-- Name: wc_settlement_details; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_settlement_details (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    settlement_id uuid NOT NULL,
    wc_user_id uuid NOT NULL,
    final_balance bigint NOT NULL,
    amount numeric(12,2) NOT NULL,
    direction character varying(10) NOT NULL,
    status character varying(20) DEFAULT 'pending'::character varying NOT NULL,
    completed_at timestamp with time zone,
    done_note character varying(255),
    created_at timestamp with time zone
);


ALTER TABLE public.wc_settlement_details OWNER TO postgres;

--
-- Name: wc_settlements; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_settlements (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    point_rate numeric(10,2) NOT NULL,
    settled_by uuid NOT NULL,
    note character varying(255),
    created_at timestamp with time zone
);


ALTER TABLE public.wc_settlements OWNER TO postgres;

--
-- Name: wc_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_users (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    name character varying(100) NOT NULL,
    password_hash character varying(255) NOT NULL,
    is_admin boolean DEFAULT false,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.wc_users OWNER TO postgres;

--
-- Name: wc_wallet_logs; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_wallet_logs (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wc_user_id uuid NOT NULL,
    admin_id uuid NOT NULL,
    delta bigint NOT NULL,
    balance_before bigint NOT NULL,
    balance_after bigint NOT NULL,
    note character varying(255),
    created_at timestamp with time zone
);


ALTER TABLE public.wc_wallet_logs OWNER TO postgres;

--
-- Name: wc_wallets; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.wc_wallets (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    wc_user_id uuid NOT NULL,
    balance bigint DEFAULT 0,
    created_at timestamp with time zone,
    updated_at timestamp with time zone
);


ALTER TABLE public.wc_wallets OWNER TO postgres;

--
-- Data for Name: config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.config (key, value, description) FROM stdin;
pro_win_rate_threshold	0.85	Win rate threshold (0-1) required to reach the Pro tier
normal_win_rate_threshold	0.4	Win rate threshold (0-1) required to reach the Normal tier (below this = Noob)
debt_threshold	-6	Score threshold that triggers debt settlement
point_to_vnd	22000	Conversion rate: 1 point = X VND
fund_split_percent	50	Percentage of debt that goes to fund (rest to winners)
auto_settlement	false	Automatically trigger settlement when debt threshold is reached (true/false)
points_per_win	1	\N
min_matches_for_tier	5	Minimum matches a player must play before tier evaluation applies
\.


--
-- Data for Name: debt_settlements; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.debt_settlements (id, debtor_id, debt_amount, money_amount, to_fund, to_winners, winner_distribution, settled_at, fund_amount, original_debt_points, settlement_date, created_at) FROM stdin;
5fe004db-4309-4fa1-89b1-c55098e5ca1a	7a69c6be-ec53-4981-9128-ce313249496a	-6	132000	\N	\N	66000	2026-04-15 06:18:20.549308+00	66000	6	2026-04-15 06:18:20.555952+00	2026-04-15 06:18:20.555986+00
5919e9d9-4e34-49b2-9134-659717e4ea1c	72ef779d-4d4b-4369-a799-a9fbca34a888	-1	22000	\N	\N	11000	2026-04-15 06:19:45.345147+00	11000	1	2026-04-15 06:19:45.361797+00	2026-04-15 06:19:45.362638+00
2daf7934-2763-4572-9c1a-e4e5ff56469a	7a69c6be-ec53-4981-9128-ce313249496a	-6	132000	66000.00	66000.00	66000	2026-04-15 06:27:48.981715+00	66000	6	2026-04-15 06:27:48.998326+00	2026-04-15 06:27:48.998399+00
0f1e0718-8d80-413d-ae4a-cc1bc6bc10d0	7a69c6be-ec53-4981-9128-ce313249496a	-6	132000	66000.00	66000.00	66000	2026-04-15 06:28:29.075859+00	66000	6	2026-04-15 06:28:29.086197+00	2026-04-15 06:28:29.08623+00
53295c21-3e82-46da-ac76-1da20fb73689	baae4310-e82c-424a-b71f-b336e51d7311	-6	132000	66000.00	66000.00	66000	2026-04-15 07:28:46.639551+00	66000	6	2026-04-15 07:28:46.646042+00	2026-04-15 07:28:46.646442+00
e9c6984e-1b94-4e8f-981a-bd48d5447c25	b60f95f2-a951-408f-8c7f-98b86b5dc098	-7	154000	77000.00	77000.00	77000	2026-05-14 11:24:21.964468+00	77000	7	2026-05-14 11:24:21.97964+00	2026-05-14 11:24:21.979742+00
a5ff79c1-c144-4e19-84a4-6c620a8a7a0c	2978025f-c9a9-43d9-8953-f2e0e9a0232d	-9	198000	99000.00	99000.00	99000	2026-05-14 15:50:08.003448+00	99000	9	2026-05-14 15:50:08.040735+00	2026-05-14 15:50:08.041201+00
958e6b20-a417-4b50-9d0b-a4875cc6e481	2978025f-c9a9-43d9-8953-f2e0e9a0232d	-9	198000	99000.00	99000.00	99000	2026-05-14 15:51:14.86694+00	99000	9	2026-05-14 15:51:14.912472+00	2026-05-14 15:51:14.912524+00
6770e6b7-fa30-4eaf-8c3f-425e870c1af4	2978025f-c9a9-43d9-8953-f2e0e9a0232d	-9	132000	66000.00	66000.00	66000	2026-05-14 15:52:06.83018+00	66000	9	2026-05-14 15:52:06.881539+00	2026-05-14 15:52:06.881687+00
\.


--
-- Data for Name: fund_transactions; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.fund_transactions (id, amount, transaction_type, description, related_settlement_id, created_at, transaction_date) FROM stdin;
68dc6b29-fdd5-4ce0-bcf4-c51fb7b2c22b	100000	deposit	Initial deposit	\N	2026-04-14 13:57:31.672997+00	2026-04-14 13:57:31.670805+00
7b1f2f80-169d-4c30-bc3c-f53a602e23a4	30000	withdrawal	Equipment purchase	\N	2026-04-14 13:57:31.71189+00	2026-04-14 13:57:31.711005+00
10ab4421-0e1c-4ac6-a5ca-1b5db2eb4965	66000	deposit	Settlement: 50% fund share from Dennis's debt (132000 VND)	\N	2026-04-15 06:18:20.571163+00	2026-04-15 06:18:20.570547+00
4dbc1686-9cac-4512-84ee-452a1ca75255	11000	deposit	Settlement: 50% fund share from Cuban's debt (22000 VND)	\N	2026-04-15 06:19:45.385301+00	2026-04-15 06:19:45.384619+00
fd5155b0-aed7-4d30-a0bd-1a9a4f3a7ed8	66000	deposit	Settlement: 50% fund share from Dennis's debt (132000 VND)	\N	2026-04-15 06:27:49.01257+00	2026-04-15 06:27:49.011868+00
72151284-5f7d-458e-ace3-140e71e89148	66000	deposit	Settlement: 50% fund share from Dennis's debt (132000 VND)	\N	2026-04-15 06:28:29.089745+00	2026-04-15 06:28:29.089216+00
40fb09ae-45e2-4a2c-a202-be0597322670	10000	deposit	Manual deposit to fund	\N	2026-04-15 06:39:23.157635+00	2026-04-15 06:39:23.149212+00
998fbc93-7f4c-4207-a799-83722da8dda7	10000	deposit	Manual deposit to fund	\N	2026-04-15 06:51:08.285184+00	2026-04-15 06:51:08.266076+00
86eceeeb-050d-48c5-945a-a2182300eb7e	66000	deposit	Settlement: 50% fund share from Ric's debt (132000 VND)	\N	2026-04-15 07:28:46.65892+00	2026-04-15 07:28:46.658098+00
8ddb7d2e-35c3-44fd-a292-9b1895c62c90	10000	withdrawal	Manual withdrawal from fund	\N	2026-04-15 07:29:07.395433+00	2026-04-15 07:29:07.391698+00
b1f2f250-18ef-46af-af78-f5e1b65e4acc	10000	deposit	Manual deposit to funddsa	\N	2026-04-15 07:30:00.977564+00	2026-04-15 07:30:00.969886+00
043fa468-547c-4602-8084-2cee6a497c94	10000	withdrawal	Manual withdrawal from fund	\N	2026-04-15 07:30:04.255225+00	2026-04-15 07:30:04.254262+00
9d1b1c65-fd8a-4576-adab-7db2caae1922	77000	deposit	Settlement: 50% fund share from HAHHA's debt (154000 VND)	\N	2026-05-14 11:24:21.997777+00	2026-05-14 11:24:21.994308+00
9384afc7-e75d-4733-bb0a-76bf5b228088	99000	deposit	Settlement: 50% fund share from Ben's debt (198000 VND)	\N	2026-05-14 15:50:08.058815+00	2026-05-14 15:50:08.058482+00
6b7963cf-04ef-497f-9e9c-40942a31419a	99000	deposit	Settlement: 50% fund share from Ben's debt (198000 VND)	\N	2026-05-14 15:51:14.936117+00	2026-05-14 15:51:14.934668+00
aa00a05c-98b8-4fab-8f2b-0c589b3b41a6	66000	deposit	Settlement: 50% fund share from Ben's debt (132000 VND)	\N	2026-05-14 15:52:06.906069+00	2026-05-14 15:52:06.904643+00
\.


--
-- Data for Name: match_participants; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.match_participants (id, match_id, user_id, team_number, point_change) FROM stdin;
f10f5a2f-b80c-4a11-9332-30a9a87a9cd7	4fa858ca-f673-4f32-af15-c8bc84854a0e	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
1f80b329-3740-4463-8946-e9093786b260	4fa858ca-f673-4f32-af15-c8bc84854a0e	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
de207b54-4dd7-4c48-8497-0d4069dfe92e	b018be0c-e261-464c-9040-ba52bd8fa3b3	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
71206285-9fdc-47f6-9bc2-fdaef6abd9e6	b018be0c-e261-464c-9040-ba52bd8fa3b3	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
3b11e2d9-e854-4bb6-95f2-4e9aeff12d58	ecf9b76a-9c2b-4012-a273-19f6bd8bef71	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
caefea9c-4200-4a55-91d8-addfa2c01d76	ecf9b76a-9c2b-4012-a273-19f6bd8bef71	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
4dc1c906-c557-435f-ab6c-69d9a528df87	b5b67ef7-56b7-419b-894a-936b2102d53e	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
b4dd37ab-0942-4d9b-a0d5-817068551948	b5b67ef7-56b7-419b-894a-936b2102d53e	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
73d5f3bc-567d-48ad-ae96-c2dd30372fa0	841544ad-7221-4c2b-8147-1fbc7de15afe	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
8c0eea52-094e-4f1c-9b41-4f6e7c108f54	841544ad-7221-4c2b-8147-1fbc7de15afe	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
14a73871-3b3c-4bc2-b009-f29a133da65d	500fb241-1269-42b0-beb6-bfd90133897a	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
e1fd0875-46be-46ac-b23f-deef87ea74b3	500fb241-1269-42b0-beb6-bfd90133897a	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
d217e754-925b-43eb-b216-6777be3c6ffa	6abe7034-b0b7-4823-a3e7-4eeffd6ee4da	8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	1	-1
515119a7-51c0-400b-8355-ed3728550d93	6abe7034-b0b7-4823-a3e7-4eeffd6ee4da	240f778e-9eb6-44d8-b3ac-7ea615d76921	2	1
9c53e139-4969-4831-a1fd-adf548779ee0	a6110d3c-7dfa-499b-a99d-cecaa873ceb9	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
fe9e1658-0c42-497f-a52c-ceea7d741354	a6110d3c-7dfa-499b-a99d-cecaa873ceb9	baae4310-e82c-424a-b71f-b336e51d7311	1	1
eb65009a-c1f0-4619-9ed5-333cfeb0b1da	a6110d3c-7dfa-499b-a99d-cecaa873ceb9	1ab68412-aa4a-42ba-bc07-1bad509341db	2	-1
738ae4eb-73af-4353-9658-67d09bb99ec0	a6110d3c-7dfa-499b-a99d-cecaa873ceb9	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
7758ac06-412f-45c1-a81d-2ff7b85b5534	2b2a573f-b472-4b29-ae3e-3256b29b9350	72ef779d-4d4b-4369-a799-a9fbca34a888	1	-1
8cd8a16f-a01b-449c-8916-868ade9c586b	2b2a573f-b472-4b29-ae3e-3256b29b9350	7a69c6be-ec53-4981-9128-ce313249496a	1	-1
1e0c086d-c59c-4a23-a33d-a7a9a294325c	2b2a573f-b472-4b29-ae3e-3256b29b9350	baae4310-e82c-424a-b71f-b336e51d7311	2	1
2d2f7ba1-ab25-421d-aeac-57cf56d15ed3	2b2a573f-b472-4b29-ae3e-3256b29b9350	1ab68412-aa4a-42ba-bc07-1bad509341db	2	1
3d60c062-c313-4901-822a-cc310e477874	712121a6-983e-435e-ac36-f3bbae2adfc9	baae4310-e82c-424a-b71f-b336e51d7311	1	-1
01065089-e2f0-46d3-9a00-bf4ce39d5d99	712121a6-983e-435e-ac36-f3bbae2adfc9	72ef779d-4d4b-4369-a799-a9fbca34a888	1	-1
adc19480-331a-481d-9996-940ad63eebae	712121a6-983e-435e-ac36-f3bbae2adfc9	1ab68412-aa4a-42ba-bc07-1bad509341db	2	1
c64ce2af-66f8-4b44-b1d1-350f6f304631	712121a6-983e-435e-ac36-f3bbae2adfc9	7a69c6be-ec53-4981-9128-ce313249496a	2	1
647cea3c-cb94-4faa-bbec-8c8f7b5a48b6	24802a24-789b-4a87-b949-db9fb3c6d436	baae4310-e82c-424a-b71f-b336e51d7311	1	1
5975a45d-d8ec-419e-b02f-12600a9e0676	24802a24-789b-4a87-b949-db9fb3c6d436	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
9653c073-3ef3-481e-82d9-9f95671dc383	11b63b5d-8da4-4f9b-889b-9348ab73d6b0	baae4310-e82c-424a-b71f-b336e51d7311	1	1
f3a6df2e-8c45-4301-8f66-821e34f69d40	11b63b5d-8da4-4f9b-889b-9348ab73d6b0	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
60789d35-b805-46fc-bda5-93dbcc306694	0b533aa9-4866-4017-beea-4e9b68581136	baae4310-e82c-424a-b71f-b336e51d7311	1	1
8ac476c1-809d-4d8e-b79b-a30516770697	0b533aa9-4866-4017-beea-4e9b68581136	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
bfca3a64-d087-45de-aabf-d15f6dc88505	8920168a-0593-4fa4-be16-de652ac284d9	baae4310-e82c-424a-b71f-b336e51d7311	1	1
9c0a22a1-a7ca-406f-98c1-6d1b6b9e55b8	8920168a-0593-4fa4-be16-de652ac284d9	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
010fba49-c223-40af-aea2-83839fbcb2d1	7e9957df-e017-41d3-902c-a7274b1c0698	baae4310-e82c-424a-b71f-b336e51d7311	1	1
6a3f3a27-948a-47ec-9c6c-58a0c95ae836	7e9957df-e017-41d3-902c-a7274b1c0698	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
15576998-a039-4a70-8719-ff887ea0deea	29ed0ff9-7e6a-4e50-9670-eabbf54b0d96	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
ae32fe8a-2d9c-4ebc-9842-1457673211c9	29ed0ff9-7e6a-4e50-9670-eabbf54b0d96	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
9863986d-705d-4092-9ddc-841ed181ebf9	059cab77-abf8-4e55-a80c-0ddd00f52d99	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
d5462d70-eb1d-466b-8737-79de1dfd53de	059cab77-abf8-4e55-a80c-0ddd00f52d99	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
f584913a-4a70-4943-9442-adbb37b571aa	8540bb1f-8bf1-45a0-9435-28c6a5b3953a	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
c39e06fd-9e39-479c-b33f-04c79cf40303	8540bb1f-8bf1-45a0-9435-28c6a5b3953a	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
af7ca29c-921c-41dc-920f-469e9bdcf8d5	7eabed17-ecf9-49c9-b0cc-53df04f3d27e	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
aa62f4d6-c4f4-4a71-9739-b548d1e5a266	7eabed17-ecf9-49c9-b0cc-53df04f3d27e	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
df528ba7-58ed-4f67-80af-ef88f51b60bc	7d72e287-3530-482c-9bf8-f0ef4de46124	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
4250b6d0-afa7-468a-ba37-940d66629cce	7d72e287-3530-482c-9bf8-f0ef4de46124	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
196dd892-2fcc-431b-891f-c849bf983a56	ab070caa-e630-4472-8290-d69f664c41d8	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
05350893-8411-4ce4-9226-27108b19145b	ab070caa-e630-4472-8290-d69f664c41d8	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
6d1e04cd-e5ce-48df-a465-32d25e2a373b	6b74175a-3395-4e43-a7b4-60071b095254	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
41321349-db88-4a86-bf4a-3c33fc2a322f	6b74175a-3395-4e43-a7b4-60071b095254	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
4d89a3f4-0646-4b41-861a-8d7810e9cdc0	d87f3acb-48b9-4531-84c6-fde4834450b8	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
f550aa21-2f07-43f5-bc94-613612126017	d87f3acb-48b9-4531-84c6-fde4834450b8	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
c2959651-609d-4de6-9dee-0fc10f99d4d1	6cd4ae2b-e777-412c-8a6f-225876a93640	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
6c2ca8b6-00c6-4e5e-93cb-fdb4d893603a	6cd4ae2b-e777-412c-8a6f-225876a93640	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
319798aa-eaa8-4abe-942f-b99648bd1a78	0f712a06-3446-485a-a0de-77ec01860b37	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
c6a9ba95-0701-4c62-a5f9-f05f03c05abd	0f712a06-3446-485a-a0de-77ec01860b37	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
fb139078-3472-43d8-bcba-44875a3b9860	e4a73480-3c50-4ef6-a45f-b3b3eeaf96bd	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
690e13d4-04b9-40a5-a42d-06a1af274375	e4a73480-3c50-4ef6-a45f-b3b3eeaf96bd	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
4737e297-4bc2-4d5f-92a9-50d171279690	6b442e93-47d5-4cd1-b4a0-a59a965e3b79	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
6db8f65c-c159-4de1-b404-1e33ee64a814	6b442e93-47d5-4cd1-b4a0-a59a965e3b79	7a69c6be-ec53-4981-9128-ce313249496a	2	-1
07ac7430-4b9c-40b3-abdc-16463249b225	0509d541-4313-4f5d-ac44-c8e93abb3f6d	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
9c4148c5-79a5-4194-a4cc-f3693edaacba	0509d541-4313-4f5d-ac44-c8e93abb3f6d	baae4310-e82c-424a-b71f-b336e51d7311	2	-1
5c7ad080-0a33-49d5-81a1-e30e231d3e9b	ff249f8b-72e7-4bd4-b67a-64fb26ac1cf8	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
69ad965a-2d4a-426d-8083-b8d7708c1e9c	ff249f8b-72e7-4bd4-b67a-64fb26ac1cf8	baae4310-e82c-424a-b71f-b336e51d7311	2	-1
81d56b15-df6f-4ad9-a568-9b95c125df16	e26d0705-76a0-4cd7-9063-9fd793a66480	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
163bc60d-bdb9-4e0b-b506-f5abb8370dac	e26d0705-76a0-4cd7-9063-9fd793a66480	baae4310-e82c-424a-b71f-b336e51d7311	2	-1
de4fc6e7-9f27-4ccd-943b-f894c8f62838	637ebe7f-698b-43bc-8589-2c03f48cfd87	72ef779d-4d4b-4369-a799-a9fbca34a888	1	1
8b11c12e-a1b9-4771-b4ff-0751c9120203	637ebe7f-698b-43bc-8589-2c03f48cfd87	baae4310-e82c-424a-b71f-b336e51d7311	2	-1
4dbbb67a-a519-4532-ae81-91fee7550aa8	21c32f48-ae44-4b33-ba7b-a2a514cdeafa	72ef779d-4d4b-4369-a799-a9fbca34a888	1	3
27d6145c-6f9e-496b-b9e7-b8dc2b6f18f0	21c32f48-ae44-4b33-ba7b-a2a514cdeafa	7a69c6be-ec53-4981-9128-ce313249496a	2	-3
3610bf8b-b7b5-4e54-a660-c21f36fba416	defb0d3e-1c8b-4353-b905-54f7089dbb59	72ef779d-4d4b-4369-a799-a9fbca34a888	1	-1
7575bcf6-3a51-4ecc-b0c4-c80227fb95d2	defb0d3e-1c8b-4353-b905-54f7089dbb59	ece085e7-5db9-48f0-8b65-a16c65f286fb	1	-1
bfdedd89-ce7d-46c0-88d6-806aa0e0c925	defb0d3e-1c8b-4353-b905-54f7089dbb59	2978025f-c9a9-43d9-8953-f2e0e9a0232d	2	1
191064d2-5e74-457e-bd82-78d85ecbaea1	defb0d3e-1c8b-4353-b905-54f7089dbb59	1ab68412-aa4a-42ba-bc07-1bad509341db	2	1
651d5d32-7a5d-4459-8853-ea51727d1f74	98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	72ef779d-4d4b-4369-a799-a9fbca34a888	1	-1
ecd0d29a-fac3-4cbe-ad3e-c58b9c412ae4	98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	baae4310-e82c-424a-b71f-b336e51d7311	1	-1
fb27a9e9-5f2e-4da7-b45b-11e57fc36bed	98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	ece085e7-5db9-48f0-8b65-a16c65f286fb	2	1
98e894de-8767-4333-b552-1e412810e446	98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	7a69c6be-ec53-4981-9128-ce313249496a	2	1
0a35ebf7-ca63-4861-8e26-40a1efa9811f	4cd0422c-bbec-411d-b6d3-53cce38ba5e0	72ef779d-4d4b-4369-a799-a9fbca34a888	1	5
e32d59ef-90d0-4bed-82c1-f4b8ed70c688	4cd0422c-bbec-411d-b6d3-53cce38ba5e0	1ab68412-aa4a-42ba-bc07-1bad509341db	1	5
7dc751b1-ea8a-4d70-a323-09f1080a7a28	4cd0422c-bbec-411d-b6d3-53cce38ba5e0	2978025f-c9a9-43d9-8953-f2e0e9a0232d	2	-5
5ab8c279-e7ef-4cbb-b458-b0160d3cd2ac	4cd0422c-bbec-411d-b6d3-53cce38ba5e0	b60f95f2-a951-408f-8c7f-98b86b5dc098	2	-5
71fb6b29-b7d4-442c-8c8a-2a3674f55750	747bbec7-0fc4-4ac8-92d9-84876d232d16	b60f95f2-a951-408f-8c7f-98b86b5dc098	1	4
63ca0a3d-85d8-44d9-b67e-06341620d207	747bbec7-0fc4-4ac8-92d9-84876d232d16	44b66cce-e1a6-4f2b-969c-68cef1eb046a	2	-4
a6edd291-cb31-4b4b-af71-3e7503e5b37f	11f0a47e-4d43-4a9e-bdc9-1a5577db6196	b60f95f2-a951-408f-8c7f-98b86b5dc098	1	-6
de4602b1-02eb-4c5f-a58d-97478b487693	11f0a47e-4d43-4a9e-bdc9-1a5577db6196	1ab68412-aa4a-42ba-bc07-1bad509341db	2	6
38999e06-2bcc-4e23-895d-2ab7b826757c	cfe587b2-96d0-4e55-9fbd-a06040c4f556	74ddd73e-075e-422f-b398-24d718a92fc6	1	5
88934a8a-1e27-418b-b044-6edbca314911	cfe587b2-96d0-4e55-9fbd-a06040c4f556	2978025f-c9a9-43d9-8953-f2e0e9a0232d	2	-5
225d774d-c8f1-4c6d-a948-78ad9da9cc61	fe2dcc25-0fb3-44df-bbd1-a637540ce439	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1	-9
dfcc9843-f01d-4b6d-9db5-760d425349d7	fe2dcc25-0fb3-44df-bbd1-a637540ce439	1ab68412-aa4a-42ba-bc07-1bad509341db	2	9
17efc619-1b3d-4fa5-a57e-176dbd543b75	05edfe74-63da-41ef-853a-e9361071782c	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1	-9
a09e1ea6-0092-4c0f-b47b-2a2ceb84d175	05edfe74-63da-41ef-853a-e9361071782c	7335f039-e590-4407-ae4a-4eee517b7f94	2	9
\.


--
-- Data for Name: matches; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.matches (id, match_type, winner_team, match_date, recorded_by, created_at, is_locked, tournament_match_id) FROM stdin;
4fa858ca-f673-4f32-af15-c8bc84854a0e	1v1	2	2026-04-14 13:57:31.863371+00		2026-04-14 13:57:31.864027+00	f	\N
b018be0c-e261-464c-9040-ba52bd8fa3b3	1v1	2	2026-04-14 13:57:31.889113+00		2026-04-14 13:57:31.890598+00	f	\N
ecf9b76a-9c2b-4012-a273-19f6bd8bef71	1v1	2	2026-04-14 13:57:31.910661+00		2026-04-14 13:57:31.911034+00	f	\N
b5b67ef7-56b7-419b-894a-936b2102d53e	1v1	2	2026-04-14 13:57:31.931436+00		2026-04-14 13:57:31.931788+00	f	\N
841544ad-7221-4c2b-8147-1fbc7de15afe	1v1	2	2026-04-14 13:57:31.950555+00		2026-04-14 13:57:31.950865+00	f	\N
500fb241-1269-42b0-beb6-bfd90133897a	1v1	2	2026-04-14 13:57:31.97328+00		2026-04-14 13:57:31.973683+00	f	\N
6abe7034-b0b7-4823-a3e7-4eeffd6ee4da	1v1	2	2026-04-14 13:57:32.017725+00		2026-04-14 13:57:32.018089+00	f	\N
ab070caa-e630-4472-8290-d69f664c41d8	1v1	1	2026-04-15 06:27:48.971562+00		2026-04-15 06:27:48.97205+00	t	\N
7d72e287-3530-482c-9bf8-f0ef4de46124	1v1	1	2026-04-15 06:27:43.482859+00		2026-04-15 06:27:43.483585+00	t	\N
7eabed17-ecf9-49c9-b0cc-53df04f3d27e	1v1	1	2026-04-15 06:27:37.951874+00		2026-04-15 06:27:37.953102+00	t	\N
8540bb1f-8bf1-45a0-9435-28c6a5b3953a	1v1	1	2026-04-15 06:27:32.294975+00		2026-04-15 06:27:32.303415+00	t	\N
059cab77-abf8-4e55-a80c-0ddd00f52d99	1v1	1	2026-04-15 06:27:25.612514+00		2026-04-15 06:27:25.618082+00	t	\N
29ed0ff9-7e6a-4e50-9670-eabbf54b0d96	1v1	1	2026-04-15 06:27:16.87608+00		2026-04-15 06:27:16.878171+00	t	\N
7e9957df-e017-41d3-902c-a7274b1c0698	1v1	1	2026-04-15 04:25:45.951029+00		2026-04-15 04:25:45.956101+00	t	\N
8920168a-0593-4fa4-be16-de652ac284d9	1v1	1	2026-04-15 04:25:39.592232+00		2026-04-15 04:25:39.593379+00	t	\N
0b533aa9-4866-4017-beea-4e9b68581136	1v1	1	2026-04-15 04:25:32.85266+00		2026-04-15 04:25:32.853701+00	t	\N
11b63b5d-8da4-4f9b-889b-9348ab73d6b0	1v1	1	2026-04-15 04:25:26.133516+00		2026-04-15 04:25:26.134441+00	t	\N
24802a24-789b-4a87-b949-db9fb3c6d436	1v1	1	2026-04-15 04:25:18.740174+00		2026-04-15 04:25:18.740704+00	t	\N
712121a6-983e-435e-ac36-f3bbae2adfc9	2v2	2	2026-04-15 04:23:37.913875+00		2026-04-15 04:23:37.919213+00	t	\N
2b2a573f-b472-4b29-ae3e-3256b29b9350	2v2	2	2026-04-14 16:44:44.690857+00		2026-04-14 16:44:44.69183+00	t	\N
a6110d3c-7dfa-499b-a99d-cecaa873ceb9	2v2	1	2026-04-14 15:53:46.189911+00		2026-04-14 15:53:46.190981+00	t	\N
6b442e93-47d5-4cd1-b4a0-a59a965e3b79	1v1	1	2026-04-15 06:28:29.054523+00		2026-04-15 06:28:29.056535+00	t	\N
e4a73480-3c50-4ef6-a45f-b3b3eeaf96bd	1v1	1	2026-04-15 06:28:22.767227+00		2026-04-15 06:28:22.768438+00	t	\N
0f712a06-3446-485a-a0de-77ec01860b37	1v1	1	2026-04-15 06:28:17.499913+00		2026-04-15 06:28:17.500569+00	t	\N
6cd4ae2b-e777-412c-8a6f-225876a93640	1v1	1	2026-04-15 06:28:10.154313+00		2026-04-15 06:28:10.155183+00	t	\N
d87f3acb-48b9-4531-84c6-fde4834450b8	1v1	1	2026-04-15 06:28:05.65814+00		2026-04-15 06:28:05.659361+00	t	\N
6b74175a-3395-4e43-a7b4-60071b095254	1v1	1	2026-04-15 06:27:58.514495+00		2026-04-15 06:27:58.515212+00	t	\N
0509d541-4313-4f5d-ac44-c8e93abb3f6d	1v1	1	2026-04-15 07:24:21.15666+00		2026-04-15 07:24:21.159043+00	f	\N
ff249f8b-72e7-4bd4-b67a-64fb26ac1cf8	1v1	1	2026-04-15 07:24:27.065576+00		2026-04-15 07:24:27.067918+00	f	\N
e26d0705-76a0-4cd7-9063-9fd793a66480	1v1	1	2026-04-15 07:24:33.572211+00		2026-04-15 07:24:33.572566+00	f	\N
637ebe7f-698b-43bc-8589-2c03f48cfd87	1v1	1	2026-04-15 07:24:40.115509+00		2026-04-15 07:24:40.120464+00	f	\N
21c32f48-ae44-4b33-ba7b-a2a514cdeafa	1v1	1	2026-04-15 07:51:02.840541+00		2026-04-15 07:51:02.846746+00	f	\N
98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	2v2	2	2026-04-16 09:27:59.940387+00		2026-04-16 09:27:59.94149+00	f	ae3cfe9f-39d7-4d47-a3fc-f3da31ef4572
747bbec7-0fc4-4ac8-92d9-84876d232d16	1v1	1	2026-05-14 11:23:36.179779+00		2026-05-14 11:23:36.186377+00	f	\N
11f0a47e-4d43-4a9e-bdc9-1a5577db6196	1v1	2	2026-05-14 11:24:06.795801+00		2026-05-14 11:24:06.796336+00	f	\N
cfe587b2-96d0-4e55-9fbd-a06040c4f556	1v1	1	2026-05-14 15:34:35.718614+00		2026-05-14 15:34:35.722931+00	t	\N
4cd0422c-bbec-411d-b6d3-53cce38ba5e0	2v2	1	2026-04-17 14:01:04+00		2026-04-17 14:01:10.729593+00	t	\N
defb0d3e-1c8b-4353-b905-54f7089dbb59	2v2	2	2026-04-16 09:27:59.022589+00		2026-04-16 09:27:59.025008+00	t	73217694-9192-4ff3-b729-c99f50d68247
fe2dcc25-0fb3-44df-bbd1-a637540ce439	1v1	2	2026-05-14 15:50:55.234406+00		2026-05-14 15:50:55.245343+00	t	\N
05edfe74-63da-41ef-853a-e9361071782c	1v1	2	2026-05-14 15:51:53.628623+00		2026-05-14 15:51:53.631798+00	f	\N
\.


--
-- Data for Name: settlement_winners; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.settlement_winners (id, settlement_id, winner_id, money_amount, points_deducted) FROM stdin;
da287382-2d95-4b64-a245-0b69b1460811	5fe004db-4309-4fa1-89b1-c55098e5ca1a	baae4310-e82c-424a-b71f-b336e51d7311	66000	6
28f386db-fcb4-4ded-979b-a2b72f06adaa	5919e9d9-4e34-49b2-9134-659717e4ea1c	1ab68412-aa4a-42ba-bc07-1bad509341db	11000	1
fac41e5f-bcda-427d-aed2-84f43a23c0e4	2daf7934-2763-4572-9c1a-e4e5ff56469a	72ef779d-4d4b-4369-a799-a9fbca34a888	30800	2
f4d5bf6f-f61d-4e04-a1af-24eda87420d7	2daf7934-2763-4572-9c1a-e4e5ff56469a	baae4310-e82c-424a-b71f-b336e51d7311	30800	2
ed27e46c-e25a-47b3-8818-ecc0bf52f467	2daf7934-2763-4572-9c1a-e4e5ff56469a	1ab68412-aa4a-42ba-bc07-1bad509341db	4400	0
40f077cb-789a-4113-9aa4-7e97ca8dc6b3	0f1e0718-8d80-413d-ae4a-cc1bc6bc10d0	72ef779d-4d4b-4369-a799-a9fbca34a888	66000	6
fe62622f-d166-4e0d-8478-59774709a3f7	53295c21-3e82-46da-ac76-1da20fb73689	72ef779d-4d4b-4369-a799-a9fbca34a888	66000	6
d5c1a571-8aba-4b93-a6c5-aaf7431dad33	e9c6984e-1b94-4e8f-981a-bd48d5447c25	1ab68412-aa4a-42ba-bc07-1bad509341db	77000	7
291df41a-45aa-4f59-b1f0-62fa85249be9	a5ff79c1-c144-4e19-84a4-6c620a8a7a0c	74ddd73e-075e-422f-b398-24d718a92fc6	33000	3
80a3d169-bb2f-4ffd-a8bb-7a522209f3f0	a5ff79c1-c144-4e19-84a4-6c620a8a7a0c	72ef779d-4d4b-4369-a799-a9fbca34a888	33000	3
f5474e65-6f47-4ac2-8407-1140f97edbcf	a5ff79c1-c144-4e19-84a4-6c620a8a7a0c	1ab68412-aa4a-42ba-bc07-1bad509341db	33000	3
c2d99436-39d3-4ee7-9140-770d33d33961	958e6b20-a417-4b50-9d0b-a4875cc6e481	1ab68412-aa4a-42ba-bc07-1bad509341db	99000	9
0037463f-c11e-41ab-9b10-b0d4a78b6d69	6770e6b7-fa30-4eaf-8c3f-425e870c1af4	7335f039-e590-4407-ae4a-4eee517b7f94	44000	4
8d8f4740-a09a-4296-970b-edf6aedcbc88	6770e6b7-fa30-4eaf-8c3f-425e870c1af4	72ef779d-4d4b-4369-a799-a9fbca34a888	22000	2
\.


--
-- Data for Name: tournament_matches; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tournament_matches (id, tournament_id, round, match_order, team1_player1_id, team1_player2_id, team2_player1_id, team2_player2_id, handicap_team1, handicap_team2, status, actual_score1, actual_score2, effective_winner, match_id, created_at, updated_at) FROM stdin;
e4e1d17c-9940-4da3-9d9a-8b372b03f6cd	772e7743-1743-4d28-a1e0-7440875b627a	1	1	72ef779d-4d4b-4369-a799-a9fbca34a888	1ab68412-aa4a-42ba-bc07-1bad509341db	2978025f-c9a9-43d9-8953-f2e0e9a0232d	b60f95f2-a951-408f-8c7f-98b86b5dc098	0.5	0	pending	\N	\N	0	\N	2026-04-17 13:49:50.437931+00	2026-04-17 13:49:50.437931+00
ef57edc2-f931-46d2-9f5d-907811359eff	772e7743-1743-4d28-a1e0-7440875b627a	2	1	72ef779d-4d4b-4369-a799-a9fbca34a888	b60f95f2-a951-408f-8c7f-98b86b5dc098	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1ab68412-aa4a-42ba-bc07-1bad509341db	0.5	0	pending	\N	\N	0	\N	2026-04-17 13:49:50.437931+00	2026-04-17 13:49:50.437931+00
d8361e5d-47e7-4f43-95cd-6da0256a9b54	9027a923-d093-4a4e-9325-e2f28ad506e1	1	1	72ef779d-4d4b-4369-a799-a9fbca34a888	ece085e7-5db9-48f0-8b65-a16c65f286fb	2978025f-c9a9-43d9-8953-f2e0e9a0232d	baae4310-e82c-424a-b71f-b336e51d7311	0.5	0	pending	\N	\N	0	\N	2026-04-16 09:27:13.375377+00	2026-04-16 09:27:13.375377+00
58b10cac-173e-4da0-9dfa-e3df41782897	9027a923-d093-4a4e-9325-e2f28ad506e1	2	1	72ef779d-4d4b-4369-a799-a9fbca34a888	1ab68412-aa4a-42ba-bc07-1bad509341db	ece085e7-5db9-48f0-8b65-a16c65f286fb	7a69c6be-ec53-4981-9128-ce313249496a	0.5	0	pending	\N	\N	0	\N	2026-04-16 09:27:13.375377+00	2026-04-16 09:27:13.375377+00
3d87faac-9e55-4073-af40-e86cd06ecb6d	9027a923-d093-4a4e-9325-e2f28ad506e1	3	1	2978025f-c9a9-43d9-8953-f2e0e9a0232d	baae4310-e82c-424a-b71f-b336e51d7311	1ab68412-aa4a-42ba-bc07-1bad509341db	7a69c6be-ec53-4981-9128-ce313249496a	0	0	pending	\N	\N	0	\N	2026-04-16 09:27:13.375377+00	2026-04-16 09:27:13.375377+00
ec8e66e1-9a4d-4112-8300-7ec6080e2546	9027a923-d093-4a4e-9325-e2f28ad506e1	4	1	72ef779d-4d4b-4369-a799-a9fbca34a888	baae4310-e82c-424a-b71f-b336e51d7311	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1ab68412-aa4a-42ba-bc07-1bad509341db	0.5	0	pending	\N	\N	0	\N	2026-04-16 09:27:13.375377+00	2026-04-16 09:27:13.375377+00
8b55f468-f072-4423-8e6e-10f4ecf74936	9027a923-d093-4a4e-9325-e2f28ad506e1	5	1	2978025f-c9a9-43d9-8953-f2e0e9a0232d	ece085e7-5db9-48f0-8b65-a16c65f286fb	baae4310-e82c-424a-b71f-b336e51d7311	7a69c6be-ec53-4981-9128-ce313249496a	0	0	pending	\N	\N	0	\N	2026-04-16 09:27:13.375377+00	2026-04-16 09:27:13.375377+00
18f01031-0830-472d-b8f7-399b67c92e29	84e03db9-bb0e-4fbc-9491-e57fefc81c95	3	1	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1ab68412-aa4a-42ba-bc07-1bad509341db	baae4310-e82c-424a-b71f-b336e51d7311	7a69c6be-ec53-4981-9128-ce313249496a	0	0	pending	\N	\N	0	\N	2026-04-16 09:27:57.274653+00	2026-04-16 09:27:57.274653+00
054658f2-be46-434f-9cab-d6d044a42280	84e03db9-bb0e-4fbc-9491-e57fefc81c95	4	1	72ef779d-4d4b-4369-a799-a9fbca34a888	1ab68412-aa4a-42ba-bc07-1bad509341db	2978025f-c9a9-43d9-8953-f2e0e9a0232d	baae4310-e82c-424a-b71f-b336e51d7311	0.5	0	pending	\N	\N	0	\N	2026-04-16 09:27:57.274653+00	2026-04-16 09:27:57.274653+00
83fb219a-1e48-4c56-97be-b79053de18dc	84e03db9-bb0e-4fbc-9491-e57fefc81c95	5	1	2978025f-c9a9-43d9-8953-f2e0e9a0232d	ece085e7-5db9-48f0-8b65-a16c65f286fb	1ab68412-aa4a-42ba-bc07-1bad509341db	7a69c6be-ec53-4981-9128-ce313249496a	0	0	pending	\N	\N	0	\N	2026-04-16 09:27:57.274653+00	2026-04-16 09:27:57.274653+00
73217694-9192-4ff3-b729-c99f50d68247	84e03db9-bb0e-4fbc-9491-e57fefc81c95	1	1	72ef779d-4d4b-4369-a799-a9fbca34a888	ece085e7-5db9-48f0-8b65-a16c65f286fb	2978025f-c9a9-43d9-8953-f2e0e9a0232d	1ab68412-aa4a-42ba-bc07-1bad509341db	0.5	0	completed	0	0	2	defb0d3e-1c8b-4353-b905-54f7089dbb59	2026-04-16 09:27:57.274653+00	2026-04-16 09:27:59.052193+00
ae3cfe9f-39d7-4d47-a3fc-f3da31ef4572	84e03db9-bb0e-4fbc-9491-e57fefc81c95	2	1	72ef779d-4d4b-4369-a799-a9fbca34a888	baae4310-e82c-424a-b71f-b336e51d7311	ece085e7-5db9-48f0-8b65-a16c65f286fb	7a69c6be-ec53-4981-9128-ce313249496a	0.5	0	completed	0	0	2	98efab1d-ecd5-476d-a9d3-609fc4f2bdcb	2026-04-16 09:27:57.274653+00	2026-04-16 09:27:59.949737+00
\.


--
-- Data for Name: tournament_participants; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tournament_participants (id, tournament_id, user_id, tier_snapshot, handicap_rate_snapshot) FROM stdin;
cfdac65b-9bb8-4bb6-99c2-b5adc1f48f4d	9027a923-d093-4a4e-9325-e2f28ad506e1	72ef779d-4d4b-4369-a799-a9fbca34a888	pro	0.5
57ffa439-7cd1-4b4c-8bd4-4fda058de4fa	9027a923-d093-4a4e-9325-e2f28ad506e1	2978025f-c9a9-43d9-8953-f2e0e9a0232d	pro	0
81e22d5b-8f9e-4b5e-bda2-2559b22d2a37	9027a923-d093-4a4e-9325-e2f28ad506e1	ece085e7-5db9-48f0-8b65-a16c65f286fb	normal	0
c9bc73c8-3c0b-4c95-b0fe-07c7e4f23ff3	9027a923-d093-4a4e-9325-e2f28ad506e1	baae4310-e82c-424a-b71f-b336e51d7311	normal	0
2e4b859a-b1c6-49eb-90b0-851d28dd704b	9027a923-d093-4a4e-9325-e2f28ad506e1	1ab68412-aa4a-42ba-bc07-1bad509341db	normal	0
13998f90-c5b0-429b-a0d6-809941c6db63	9027a923-d093-4a4e-9325-e2f28ad506e1	7a69c6be-ec53-4981-9128-ce313249496a	pro	0
48b2f75c-8700-47d3-8aa6-ac699e5d2814	84e03db9-bb0e-4fbc-9491-e57fefc81c95	72ef779d-4d4b-4369-a799-a9fbca34a888	pro	0.5
80d07735-d973-496e-962d-76f883f9b1de	84e03db9-bb0e-4fbc-9491-e57fefc81c95	2978025f-c9a9-43d9-8953-f2e0e9a0232d	pro	0
a41095ce-3e4a-4394-9aea-1afc6e765d85	84e03db9-bb0e-4fbc-9491-e57fefc81c95	ece085e7-5db9-48f0-8b65-a16c65f286fb	normal	0
c0c0a6b1-1f2b-40de-9293-2a77eb495885	84e03db9-bb0e-4fbc-9491-e57fefc81c95	1ab68412-aa4a-42ba-bc07-1bad509341db	normal	0
87ecb556-a3b6-4c4d-9d2e-32e7e551a166	84e03db9-bb0e-4fbc-9491-e57fefc81c95	baae4310-e82c-424a-b71f-b336e51d7311	normal	0
d61f471b-7359-4f57-a4b2-5e9478ed5d88	84e03db9-bb0e-4fbc-9491-e57fefc81c95	7a69c6be-ec53-4981-9128-ce313249496a	pro	0
979cae81-082a-483f-a274-2ed4b0d741e4	772e7743-1743-4d28-a1e0-7440875b627a	72ef779d-4d4b-4369-a799-a9fbca34a888	pro	0.5
4b14cd6d-56a3-440f-8e10-5af80f76903e	772e7743-1743-4d28-a1e0-7440875b627a	2978025f-c9a9-43d9-8953-f2e0e9a0232d	pro	0
edc20a57-ea4d-4531-8044-09ac34e29e83	772e7743-1743-4d28-a1e0-7440875b627a	1ab68412-aa4a-42ba-bc07-1bad509341db	normal	0
b0e5ed7e-e0c7-44aa-b78c-45802d6b29a3	772e7743-1743-4d28-a1e0-7440875b627a	b60f95f2-a951-408f-8c7f-98b86b5dc098	normal	0
\.


--
-- Data for Name: tournaments; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.tournaments (id, name, match_type, status, affects_score, entry_fee, created_at, updated_at) FROM stdin;
9027a923-d093-4a4e-9325-e2f28ad506e1	TESADASDsa	2v2	completed	t	0	2026-04-16 09:27:13.355738+00	2026-04-16 09:27:17.106107+00
84e03db9-bb0e-4fbc-9491-e57fefc81c95	dsadsadsadsadsadasdsadasdsadsa	2v2	completed	t	10000	2026-04-16 09:27:57.267573+00	2026-04-16 09:28:02.947926+00
772e7743-1743-4d28-a1e0-7440875b627a	dsa	2v2	active	t	1000000	2026-04-17 13:49:50.401211+00	2026-04-17 13:49:50.401211+00
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.users (id, name, current_score, created_at, updated_at, is_active, tier, handicap_rate) FROM stdin;
33dd27fe-6c26-4bfd-a9c0-b990cd3ae246	Messi	0	2026-04-14 11:41:27.101133+00	2026-04-14 11:41:27.220207+00	f	normal	0
64f3ba33-5997-45ef-8322-79aca3a3ab7e	AB	0	2026-04-14 11:41:59.475403+00	2026-04-14 11:41:59.655832+00	f	normal	0
786ec911-006c-42ca-aa79-73b4ed16f32a	AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA	0	2026-04-14 11:41:59.500224+00	2026-04-14 11:41:59.670167+00	f	normal	0
0b34c567-1f6c-447f-bc0a-945ffa99d477	CR7	0	2026-04-14 11:41:27.070177+00	2026-04-14 11:50:48.299282+00	f	normal	0
f2fca01d-1455-4867-ba62-0680676d3545	Neymar	0	2026-04-14 11:41:27.125188+00	2026-04-14 11:50:48.322308+00	f	normal	0
7aeb6644-4b42-4011-9326-0e639ab93657	Spaced Name	0	2026-04-14 11:41:59.526129+00	2026-04-14 11:50:48.33719+00	f	normal	0
9f6398d2-6890-4e74-816e-4296aaa5511c	Alice	0	2026-04-14 11:50:48.358453+00	2026-04-14 11:53:30.058903+00	f	normal	0
17b3dee5-e347-4fb7-ae07-164b2c679e0d	Bob	0	2026-04-14 11:50:48.378717+00	2026-04-14 11:53:30.078235+00	f	normal	0
23c49fac-29ea-4249-9dd1-9733bca479ea	Charlie	0	2026-04-14 11:50:48.396093+00	2026-04-14 11:53:30.09189+00	f	normal	0
05e9533e-ff8c-40c8-bb29-b1fbbed3b8cd	David	0	2026-04-14 11:50:48.412197+00	2026-04-14 11:53:30.105541+00	f	normal	0
75cf6215-bde1-4c0e-b272-df0da4dae269	Alice	0	2026-04-14 12:02:03.736365+00	2026-04-14 13:57:31.504373+00	f	normal	0
2c2564c6-9305-4ff6-992f-5d7ee060cc2b	Bob	0	2026-04-14 12:02:03.753548+00	2026-04-14 13:57:31.521197+00	f	normal	0
c2acd170-5abe-454e-b996-f7acd5aa1609	Charlie	0	2026-04-14 12:02:03.773133+00	2026-04-14 13:57:31.536774+00	f	normal	0
665183e8-378f-4122-9560-1c0d3eff27b3	David	0	2026-04-14 12:02:03.789445+00	2026-04-14 13:57:31.552382+00	f	normal	0
e171022b-135c-4edd-ab73-42a2607d790f	AliceTest	0	2026-04-14 11:59:40.838519+00	2026-04-14 11:59:56.505295+00	f	normal	0
726a0dec-9b53-4f19-a28f-6a05a71ad703	TestManual	0	2026-04-14 11:59:17.355406+00	2026-04-14 11:59:56.521468+00	f	normal	0
c3f68df2-03ff-4005-bc83-5f152b246b62	Alice	0	2026-04-14 12:00:52.65319+00	2026-04-14 12:00:56.7694+00	f	normal	0
240f778e-9eb6-44d8-b3ac-7ea615d76921	Player2	7	2026-04-14 13:57:31.826268+00	2026-04-14 16:01:01.02884+00	f	normal	0
b709663b-27f0-4ef6-9b38-abc40e93caf6	Player3	0	2026-04-14 13:57:31.846416+00	2026-04-14 16:01:04.763226+00	f	normal	0
8e1d2811-cfba-4b76-a05a-4d1973c6c4cf	Player1	-7	2026-04-14 13:57:31.807118+00	2026-04-14 16:01:10.148523+00	f	normal	0
be121700-ec9e-40a8-9539-cf8760f1fd03	Charlie	0	2026-04-14 12:00:56.824458+00	2026-04-14 12:02:03.595388+00	f	normal	0
3c8857af-faea-463a-ac3e-7a3fc95300c9	David	0	2026-04-14 12:00:56.843428+00	2026-04-14 12:02:03.596575+00	f	normal	0
73e684fc-d39a-41b1-8811-785f36e43ea0	Alice	0	2026-04-14 12:00:56.785712+00	2026-04-14 12:02:03.612917+00	f	normal	0
b1982bb4-91e3-4409-8818-d90ec4df8a68	Bob	0	2026-04-14 12:00:56.807588+00	2026-04-14 12:02:03.613595+00	f	normal	0
3fb3db80-e09c-486e-a6ab-09ff6029b7d3	Debug1	0	2026-04-14 11:54:08.648196+00	2026-04-14 12:02:03.62876+00	f	normal	0
792d9ba1-6bfc-4278-8267-bfd95d8f0f45	Debug2	0	2026-04-14 11:54:08.664581+00	2026-04-14 12:02:03.629361+00	f	normal	0
72c6367b-5178-4449-841e-e9e7e2cd02c0	TestUser1	0	2026-04-14 11:53:35.334772+00	2026-04-14 12:02:03.649861+00	f	normal	0
0a5d3b58-0082-49a6-8dae-23eb45a853bc	TestUser2	0	2026-04-14 11:53:35.347947+00	2026-04-14 12:02:03.6503+00	f	normal	0
8a72101b-922d-4582-b0ee-341fffb5de57	Alice	0	2026-04-14 12:01:47.064581+00	2026-04-14 12:02:03.678958+00	f	normal	0
72748a0e-b005-4105-83e8-8615fa3b5053	Bob	0	2026-04-14 12:01:47.090529+00	2026-04-14 12:02:03.691812+00	f	normal	0
afd00ee3-5ae6-41a6-8e4b-48e3b71599e1	Charlie	0	2026-04-14 12:01:47.107277+00	2026-04-14 12:02:03.705436+00	f	normal	0
3333176b-4fac-4431-a1c5-b3652d167a83	David	0	2026-04-14 12:01:47.119715+00	2026-04-14 12:02:03.719643+00	f	normal	0
b60f95f2-a951-408f-8c7f-98b86b5dc098	HAHHA	0	2026-04-17 11:05:28.95108+00	2026-06-03 16:07:35.316511+00	t	normal	0
74ddd73e-075e-422f-b398-24d718a92fc6	111	2	2026-04-17 14:03:27.249361+00	2026-06-03 16:07:35.320274+00	t	normal	0
1ab68412-aa4a-42ba-bc07-1bad509341db	Theo	2	2026-04-14 15:53:04.190601+00	2026-06-03 16:07:35.322931+00	t	pro	0
7335f039-e590-4407-ae4a-4eee517b7f94	dasdasdsa	5	2026-04-17 14:03:12.54436+00	2026-06-03 16:07:35.325205+00	t	normal	0
72ef779d-4d4b-4369-a799-a9fbca34a888	Cuban	3	2026-04-14 15:53:11.388943+00	2026-06-03 16:07:35.328513+00	t	normal	0.5
baae4310-e82c-424a-b71f-b336e51d7311	Ric	-1	2026-04-14 15:53:07.009458+00	2026-06-03 16:07:35.331746+00	t	normal	0
ece085e7-5db9-48f0-8b65-a16c65f286fb	Hoang Le	0	2026-04-16 04:51:18.039364+00	2026-06-03 16:07:35.336912+00	t	normal	0
7a69c6be-ec53-4981-9128-ce313249496a	Dennis	-2	2026-04-14 15:52:59.937301+00	2026-06-03 16:07:35.34015+00	t	noob	0
bb3ed2e6-58bd-4a33-ba06-0a788c459341	TETS	0	2026-04-17 10:54:36.397698+00	2026-06-03 16:07:35.342943+00	t	normal	0
7e70bbb7-20dc-44a1-8592-066d1070af6f	NEW	0	2026-04-17 14:27:05.235219+00	2026-06-03 16:07:35.345483+00	t	normal	0
9864868f-0e03-48a8-83b5-c29f7bb1bf21	LATETS	0	2026-04-17 14:27:49.246501+00	2026-06-03 16:07:35.349261+00	t	normal	2
2978025f-c9a9-43d9-8953-f2e0e9a0232d	Ben	-3	2026-04-16 04:51:12.247894+00	2026-06-03 16:07:35.35232+00	t	noob	0
44b66cce-e1a6-4f2b-969c-68cef1eb046a	dsaiodaosa	-4	2026-04-17 14:26:45.782122+00	2026-06-03 16:07:35.354628+00	t	normal	0
\.


--
-- Data for Name: wc_bets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_bets (id, wc_user_id, match_id, bet_type, stake, odds_snapshot, bet_choice, handicap_snapshot, handicap_team_snapshot, predicted_home_score, predicted_away_score, result, payout, created_at) FROM stdin;
\.


--
-- Data for Name: wc_config; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_config (id, is_enabled, updated_at, updated_by) FROM stdin;
1	f	2026-06-03 16:07:22.371702+00	8a6225ba-2371-4f22-835e-e8a6104e25de
\.


--
-- Data for Name: wc_matches; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_matches (id, external_id, home_team, away_team, home_team_code, away_team_code, match_date, group_name, stage, venue, home_score, away_score, status, handicap_team, handicap_value, odds_handicap_home, odds_handicap_away, betting_open, bets_locked_at, settled_at, created_at, updated_at) FROM stdin;
fb3e67be-8d7b-4ae8-88bd-89567a75f5c1	537327	Mexico	South Africa	MEX	RSA	2026-06-11 19:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-11 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
6adc051d-7263-45b8-be7d-cbd6bfc3adc7	537328	South Korea	Czechia	KOR	CZE	2026-06-12 02:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-12 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
61558e94-e5af-48e6-89fd-0e6b5e1f96cf	537333	Canada	Bosnia-Herzegovina	CAN	BIH	2026-06-12 19:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-12 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a12dbb0f-7e73-4155-8bea-ec1eb8661d98	537345	United States	Paraguay	USA	PAR	2026-06-13 01:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-13 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
220bcc0e-2868-4c56-92e6-b0604be7d25b	537334	Qatar	Switzerland	QAT	SUI	2026-06-13 19:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-13 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
e1ae413e-957b-43f9-a507-4b6d296fc991	537339	Brazil	Morocco	BRA	MAR	2026-06-13 22:00:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-13 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
71711a9c-7e10-4d3e-8764-4e6274ae5a63	537340	Haiti	Scotland	HAI	SCO	2026-06-14 01:00:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-14 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
2921ed74-0e7a-48bf-bedc-df848699b97b	537346	Australia	Turkey	AUS	TUR	2026-06-14 04:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-14 04:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
6bfa528b-927b-437c-ae99-b5646f81d021	537351	Germany	Curaçao	GER	CUW	2026-06-14 17:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-14 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ab0dafa0-51c7-4006-9051-9fb7de68c78c	537357	Netherlands	Japan	NED	JPN	2026-06-14 20:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-14 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
9aa98f14-0898-4fa1-8a05-917e726705d3	537352	Ivory Coast	Ecuador	CIV	ECU	2026-06-14 23:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-14 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
02c5f717-0183-4b1a-b987-06df83d0dccf	537358	Sweden	Tunisia	SWE	TUN	2026-06-15 02:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-15 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
eca08fab-2a91-4059-a8e9-a7826fc11060	537369	Spain	Cape Verde Islands	ESP	CPV	2026-06-15 16:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-15 16:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
89e37e43-17e1-4ce7-9441-38b9a3163fdc	537363	Belgium	Egypt	BEL	EGY	2026-06-15 19:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-15 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d3aaf2c4-6226-4bbb-b9d8-0bb2dc09a751	537370	Saudi Arabia	Uruguay	KSA	URY	2026-06-15 22:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-15 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
7cd59632-9110-4b20-873c-9d0baf800ddd	537364	Iran	New Zealand	IRN	NZL	2026-06-16 01:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-16 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
2f9dca2b-5ab7-4b8a-961d-a0d4e2eba69e	537391	France	Senegal	FRA	SEN	2026-06-16 19:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-16 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
3b410d4f-7055-4ac3-a13b-b9aabae47675	537392	Iraq	Norway	IRQ	NOR	2026-06-16 22:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-16 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
283c249c-982b-42e8-8bc0-4a704d69e12d	537397	Argentina	Algeria	ARG	ALG	2026-06-17 01:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-17 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
6cf0e844-9b4f-4087-ad21-0d67b06e7cb7	537398	Austria	Jordan	AUT	JOR	2026-06-17 04:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-17 04:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
72ade035-ff23-4da7-8b63-2ba6745a0d44	537403	Portugal	Congo DR	POR	COD	2026-06-17 17:00:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-17 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
efe9aff0-3274-41ee-a3a8-e6956a119f8a	537409	England	Croatia	ENG	CRO	2026-06-17 20:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-17 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
11ec7102-0fe3-40f7-8cb8-1d0a24edc193	537410	Ghana	Panama	GHA	PAN	2026-06-17 23:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-17 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
534f11a4-5474-4a1b-bfd4-97054303a9f5	537404	Uzbekistan	Colombia	UZB	COL	2026-06-18 02:00:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-18 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
741166b8-c05d-4a69-939e-1a4b66696605	537329	Czechia	South Africa	CZE	RSA	2026-06-18 16:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-18 16:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a001107a-7615-4ab9-9fa7-a52b30fa4cbc	537335	Switzerland	Bosnia-Herzegovina	SUI	BIH	2026-06-18 19:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-18 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
8f3aa0cd-579f-49fe-8786-5f5391fc8d1a	537336	Canada	Qatar	CAN	QAT	2026-06-18 22:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-18 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ff0f65f5-d399-4ed2-9e7c-52bb8b2591b6	537330	Mexico	South Korea	MEX	KOR	2026-06-19 01:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-19 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
acc68fce-11e1-488b-9d96-25b856dc57f7	537348	United States	Australia	USA	AUS	2026-06-19 19:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-19 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
20b962b1-67db-4236-837f-8fadf1d922c1	537342	Scotland	Morocco	SCO	MAR	2026-06-19 22:00:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-19 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
209fb75c-ffad-42f7-a46f-723946af3992	537341	Brazil	Haiti	BRA	HAI	2026-06-20 00:30:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-20 00:30:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
0c769d3a-5c4c-4003-a3ae-be759ae2f91a	537347	Turkey	Paraguay	TUR	PAR	2026-06-20 03:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-20 03:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
cbef9678-92bb-48d4-bae4-b05b58b9c303	537359	Netherlands	Sweden	NED	SWE	2026-06-20 17:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-20 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d3ccc39d-bcdc-443b-95c2-64905dddcbd0	537353	Germany	Ivory Coast	GER	CIV	2026-06-20 20:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-20 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
86922c8a-1bd7-45dc-977f-5d37d68a3829	537354	Ecuador	Curaçao	ECU	CUW	2026-06-21 00:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-21 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a7713c23-2998-4c5b-bfa9-5b57bbafa2a0	537360	Tunisia	Japan	TUN	JPN	2026-06-21 04:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-21 04:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
41d070e0-3cea-4b51-b20c-38a9426010c8	537371	Spain	Saudi Arabia	ESP	KSA	2026-06-21 16:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-21 16:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
c66a1065-5be2-4eec-bf1a-29ccb1b4981b	537365	Belgium	Iran	BEL	IRN	2026-06-21 19:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-21 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
50a83aab-8e3b-491b-b3b0-60826ec7684b	537372	Uruguay	Cape Verde Islands	URY	CPV	2026-06-21 22:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-21 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
9a0ba694-1b19-4099-9ff2-94bde5f6ffbc	537366	New Zealand	Egypt	NZL	EGY	2026-06-22 01:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-22 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
cc5a2c22-e8d1-4994-84e3-f7d6e8fc2297	537399	Argentina	Austria	ARG	AUT	2026-06-22 17:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-22 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
99b53325-3fa8-402f-981c-9d92d86cff03	537393	France	Iraq	FRA	IRQ	2026-06-22 21:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-22 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
1a6450ba-4002-4413-8e19-f7951df5f452	537394	Norway	Senegal	NOR	SEN	2026-06-23 00:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-23 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a7fcee77-c4e7-45cc-b190-9c57ba01c7d8	537400	Jordan	Algeria	JOR	ALG	2026-06-23 03:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-23 03:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
f63a3132-aaca-4e4f-9b56-5d0e07831f3c	537405	Portugal	Uzbekistan	POR	UZB	2026-06-23 17:00:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-23 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a0c9eca7-abfa-4629-98e7-c4e28b7c46f8	537411	England	Ghana	ENG	GHA	2026-06-23 20:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-23 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
56efdef1-44ef-4620-a6dc-03a95d9bfa62	537412	Panama	Croatia	PAN	CRO	2026-06-23 23:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-23 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
7a1bcb8d-3e59-4dac-82b8-2610de4e72b7	537406	Colombia	Congo DR	COL	COD	2026-06-24 02:00:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-24 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
f66c1cd0-d232-4165-b001-86c228896edd	537337	Switzerland	Canada	SUI	CAN	2026-06-24 19:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-24 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
3fc54fd7-3580-42e5-bb4f-431cb15f3501	537338	Bosnia-Herzegovina	Qatar	BIH	QAT	2026-06-24 19:00:00+00	Group B	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-24 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
38efb3de-6697-445d-9f3b-0d47e6742285	537344	Morocco	Haiti	MAR	HAI	2026-06-24 22:00:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-24 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
8b7f1b1d-1476-440e-9ebf-d5bb32fcc8c2	537343	Scotland	Brazil	SCO	BRA	2026-06-24 22:00:00+00	Group C	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-24 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
dba69e4c-f670-4eaa-9c11-3f5a533421a5	537331	Czechia	Mexico	CZE	MEX	2026-06-25 01:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
5de45cd6-cfd6-461b-b535-ae8eb1a48e40	537332	South Africa	South Korea	RSA	KOR	2026-06-25 01:00:00+00	Group A	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
372a0c75-d863-4741-a99d-e200a0ae915f	537355	Ecuador	Germany	ECU	GER	2026-06-25 20:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
f946f079-b045-4144-a745-d1880c5ad139	537356	Curaçao	Ivory Coast	CUW	CIV	2026-06-25 20:00:00+00	Group E	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a2439eda-147b-400c-ac42-19ab6401ed8e	537361	Tunisia	Netherlands	TUN	NED	2026-06-25 23:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a2316b8b-d305-4427-a735-6db7aa64a4f9	537362	Japan	Sweden	JPN	SWE	2026-06-25 23:00:00+00	Group F	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-25 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
222d6103-4b35-400e-81f3-c76e5269b20e	537349	Turkey	United States	TUR	USA	2026-06-26 02:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-26 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ce0d4dbe-1a7e-41eb-90b9-063e8a4bb9ff	537350	Paraguay	Australia	PAR	AUS	2026-06-26 02:00:00+00	Group D	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-26 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
9f4096d0-131f-49c3-9b17-ee6dc52f1f66	537395	Norway	France	NOR	FRA	2026-06-26 19:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-26 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
5b158ca0-0f94-4a3b-8c0a-cfec99506d10	537396	Senegal	Iraq	SEN	IRQ	2026-06-26 19:00:00+00	Group I	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-26 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
2ff998a5-26b8-403e-9fca-cf223381e202	537373	Uruguay	Spain	URY	ESP	2026-06-27 00:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
07d916c8-6924-49c8-9d1d-64efec528c9c	537374	Cape Verde Islands	Saudi Arabia	CPV	KSA	2026-06-27 00:00:00+00	Group H	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
2f35df12-0166-4f31-ab28-fc3111c03e4f	537367	New Zealand	Belgium	NZL	BEL	2026-06-27 03:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 03:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
fb2b0057-ce40-4323-afe4-ea80828d2bed	537368	Egypt	Iran	EGY	IRN	2026-06-27 03:00:00+00	Group G	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 03:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
68915051-b27f-4136-a5d1-c682f84479f4	537413	Panama	England	PAN	ENG	2026-06-27 21:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d148dd8a-32a7-4454-997c-2a9ab49bed73	537414	Croatia	Ghana	CRO	GHA	2026-06-27 21:00:00+00	Group L	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
b46ebf7b-b9fd-4702-acd3-833ef6486d2c	537407	Colombia	Portugal	COL	POR	2026-06-27 23:30:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 23:30:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a32ff366-5e53-4cb5-9340-0bf78655c0f0	537408	Congo DR	Uzbekistan	COD	UZB	2026-06-27 23:30:00+00	Group K	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-27 23:30:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
aeb4a76e-fe27-4f2b-be18-12825c0b530e	537401	Jordan	Argentina	JOR	ARG	2026-06-28 02:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-28 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
bbc68a33-a922-4707-bd5a-e890f7159410	537402	Algeria	Austria	ALG	AUT	2026-06-28 02:00:00+00	Group J	group		\N	\N	scheduled		\N	\N	\N	f	2026-06-28 02:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
f7d82796-04c3-4237-88bc-c142dcb73cf0	537417			   	   	2026-06-28 19:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-28 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
e7aaf7cb-eb2e-4b88-a520-242eab5f176d	537423			   	   	2026-06-29 17:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-29 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ab0ff0cb-f400-431b-8bd5-eb89f920f58e	537415			   	   	2026-06-29 20:30:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-29 20:30:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d4382cf8-8f50-4d98-a004-1c16cd93fc6f	537418			   	   	2026-06-30 01:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-30 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
0b7273c1-65ee-452c-a5a6-a033b207d1a5	537424			   	   	2026-06-30 17:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-30 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
8fda71c0-8f0d-4ed4-8222-05997371d9d0	537416			   	   	2026-06-30 21:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-06-30 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
5cdd8f9f-9096-4abc-8164-fa77a6724e81	537425			   	   	2026-07-01 01:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-01 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
dabcbe2e-fe5d-4663-9304-f4378291ccb6	537426			   	   	2026-07-01 16:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-01 16:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
4c44098b-a88d-42c5-84d5-e81e432e95d7	537422			   	   	2026-07-01 20:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-01 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
25d8253a-b501-4bc7-ba82-5c7b2c4a9d91	537421			   	   	2026-07-02 00:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-02 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
f524b131-96cf-4518-8b11-fe3612c49dd1	537420			   	   	2026-07-02 19:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-02 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
527d188c-86ef-4a82-b255-93c022d1c105	537419			   	   	2026-07-02 23:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-02 23:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d214dc1f-e634-47ff-99b2-673b2e0110a9	537429			   	   	2026-07-03 03:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-03 03:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
5bd4d8cd-11bc-4e29-850b-367d049e12a6	537428			   	   	2026-07-03 18:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-03 18:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
9e04086a-36d2-4cbc-9b2a-06138347adcc	537427			   	   	2026-07-03 22:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-03 22:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
9326ebca-6d4b-4dcd-8a47-4e2122a0632b	537430			   	   	2026-07-04 01:30:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-04 01:30:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a5946fdb-498c-4b70-8393-f313564fdf9f	537376			   	   	2026-07-04 17:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-04 17:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
96119acd-c8e5-4b4a-8dbe-1911f6ffc08b	537375			   	   	2026-07-04 21:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-04 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
232a1011-cd8c-418d-bc92-104f399169ef	537377			   	   	2026-07-05 20:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-05 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
b50f8447-6f84-44c2-9620-16ede8f9616d	537378			   	   	2026-07-06 00:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-06 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
17ad5413-699b-4452-998a-d69b9f87f1f1	537379			   	   	2026-07-06 19:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-06 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
139c6d4c-e0ba-473c-a2fc-f5a03e811d82	537380			   	   	2026-07-07 00:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-07 00:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ced7520d-cefb-4876-a122-a0643184bbbd	537381			   	   	2026-07-07 16:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-07 16:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
1ae60efb-00d2-4910-9709-39eb29656519	537382			   	   	2026-07-07 20:00:00+00		group		\N	\N	scheduled		\N	\N	\N	f	2026-07-07 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
26ce5751-b4c3-4ccc-b4be-12af389203b8	537383			   	   	2026-07-09 20:00:00+00		qf		\N	\N	scheduled		\N	\N	\N	f	2026-07-09 20:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
015ef7c7-c944-4733-9f5a-baec395a2806	537384			   	   	2026-07-10 19:00:00+00		qf		\N	\N	scheduled		\N	\N	\N	f	2026-07-10 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a1af3876-9348-4e24-8db1-88b8693ad177	537385			   	   	2026-07-11 21:00:00+00		qf		\N	\N	scheduled		\N	\N	\N	f	2026-07-11 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
ff9cb484-b054-41fb-8f19-9d9d84c9e6d3	537386			   	   	2026-07-12 01:00:00+00		qf		\N	\N	scheduled		\N	\N	\N	f	2026-07-12 01:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
d4b5555d-7878-48f0-b45a-29cb057f4b1c	537387			   	   	2026-07-14 19:00:00+00		sf		\N	\N	scheduled		\N	\N	\N	f	2026-07-14 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
16467f56-2f97-4e5c-8b4b-b4ca05bce028	537388			   	   	2026-07-15 19:00:00+00		sf		\N	\N	scheduled		\N	\N	\N	f	2026-07-15 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
a30b2fc3-f9d9-4f57-952c-c9601a5a51e9	537389			   	   	2026-07-18 21:00:00+00		third_place		\N	\N	scheduled		\N	\N	\N	f	2026-07-18 21:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
269c9642-c30b-48c2-80a3-cf20edcb9c01	537390			   	   	2026-07-19 19:00:00+00		final		\N	\N	scheduled		\N	\N	\N	f	2026-07-19 19:00:00+00	\N	2026-06-03 16:00:26.880486+00	2026-06-03 16:07:47.40054+00
\.


--
-- Data for Name: wc_score_odds; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_score_odds (id, match_id, home_score, away_score, odds, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: wc_settlement_details; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_settlement_details (id, settlement_id, wc_user_id, final_balance, amount, direction, status, completed_at, done_note, created_at) FROM stdin;
\.


--
-- Data for Name: wc_settlements; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_settlements (id, name, point_rate, settled_by, note, created_at) FROM stdin;
\.


--
-- Data for Name: wc_users; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_users (id, name, password_hash, is_admin, created_at, updated_at) FROM stdin;
8a6225ba-2371-4f22-835e-e8a6104e25de	buivuthanhduy@gmail.com	$2a$12$TGt/3GLvzXgn7vRb54.NxOoM51sB4GYqdPlGIpL0yWKJf7RCbfk3i	t	2026-06-03 15:54:42.491481+00	2026-06-03 15:54:42.491481+00
\.


--
-- Data for Name: wc_wallet_logs; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_wallet_logs (id, wc_user_id, admin_id, delta, balance_before, balance_after, note, created_at) FROM stdin;
\.


--
-- Data for Name: wc_wallets; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.wc_wallets (id, wc_user_id, balance, created_at, updated_at) FROM stdin;
f77d7ada-b1fe-4184-b4bd-1520ed15a537	8a6225ba-2371-4f22-835e-e8a6104e25de	0	2026-06-03 15:54:42.49486+00	2026-06-03 15:54:42.49486+00
\.


--
-- Name: config config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.config
    ADD CONSTRAINT config_pkey PRIMARY KEY (key);


--
-- Name: debt_settlements debt_settlements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_settlements
    ADD CONSTRAINT debt_settlements_pkey PRIMARY KEY (id);


--
-- Name: fund_transactions fund_transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.fund_transactions
    ADD CONSTRAINT fund_transactions_pkey PRIMARY KEY (id);


--
-- Name: match_participants match_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.match_participants
    ADD CONSTRAINT match_participants_pkey PRIMARY KEY (id);


--
-- Name: matches matches_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.matches
    ADD CONSTRAINT matches_pkey PRIMARY KEY (id);


--
-- Name: settlement_winners settlement_winners_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_winners
    ADD CONSTRAINT settlement_winners_pkey PRIMARY KEY (id);


--
-- Name: tournament_matches tournament_matches_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matches
    ADD CONSTRAINT tournament_matches_pkey PRIMARY KEY (id);


--
-- Name: tournament_participants tournament_participants_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_participants
    ADD CONSTRAINT tournament_participants_pkey PRIMARY KEY (id);


--
-- Name: tournaments tournaments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournaments
    ADD CONSTRAINT tournaments_pkey PRIMARY KEY (id);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: wc_bets wc_bets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_bets
    ADD CONSTRAINT wc_bets_pkey PRIMARY KEY (id);


--
-- Name: wc_config wc_config_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_config
    ADD CONSTRAINT wc_config_pkey PRIMARY KEY (id);


--
-- Name: wc_matches wc_matches_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_matches
    ADD CONSTRAINT wc_matches_pkey PRIMARY KEY (id);


--
-- Name: wc_score_odds wc_score_odds_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_score_odds
    ADD CONSTRAINT wc_score_odds_pkey PRIMARY KEY (id);


--
-- Name: wc_settlement_details wc_settlement_details_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_settlement_details
    ADD CONSTRAINT wc_settlement_details_pkey PRIMARY KEY (id);


--
-- Name: wc_settlements wc_settlements_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_settlements
    ADD CONSTRAINT wc_settlements_pkey PRIMARY KEY (id);


--
-- Name: wc_users wc_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_users
    ADD CONSTRAINT wc_users_pkey PRIMARY KEY (id);


--
-- Name: wc_wallet_logs wc_wallet_logs_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_wallet_logs
    ADD CONSTRAINT wc_wallet_logs_pkey PRIMARY KEY (id);


--
-- Name: wc_wallets wc_wallets_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.wc_wallets
    ADD CONSTRAINT wc_wallets_pkey PRIMARY KEY (id);


--
-- Name: idx_bet_es_dedup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_bet_es_dedup ON public.wc_bets USING btree (wc_user_id, match_id, predicted_home_score, predicted_away_score);


--
-- Name: idx_bet_hc_dedup; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_bet_hc_dedup ON public.wc_bets USING btree (wc_user_id, match_id, bet_type, bet_choice);


--
-- Name: idx_match_participants_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_match_participants_user_id ON public.match_participants USING btree (user_id);


--
-- Name: idx_score_odds_scoreline; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_score_odds_scoreline ON public.wc_score_odds USING btree (match_id, home_score, away_score);


--
-- Name: idx_settlement_user; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_settlement_user ON public.wc_settlement_details USING btree (settlement_id, wc_user_id);


--
-- Name: idx_users_active_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_users_active_name ON public.users USING btree (name) WHERE (is_active = true);


--
-- Name: idx_wc_bets_match_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wc_bets_match_id ON public.wc_bets USING btree (match_id);


--
-- Name: idx_wc_bets_wc_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wc_bets_wc_user_id ON public.wc_bets USING btree (wc_user_id);


--
-- Name: idx_wc_matches_external_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wc_matches_external_id ON public.wc_matches USING btree (external_id);


--
-- Name: idx_wc_settlement_details_settlement_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wc_settlement_details_settlement_id ON public.wc_settlement_details USING btree (settlement_id);


--
-- Name: idx_wc_users_name; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wc_users_name ON public.wc_users USING btree (name);


--
-- Name: idx_wc_wallet_logs_wc_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX idx_wc_wallet_logs_wc_user_id ON public.wc_wallet_logs USING btree (wc_user_id);


--
-- Name: idx_wc_wallets_wc_user_id; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX idx_wc_wallets_wc_user_id ON public.wc_wallets USING btree (wc_user_id);


--
-- Name: debt_settlements fk_debt_settlements_debtor; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.debt_settlements
    ADD CONSTRAINT fk_debt_settlements_debtor FOREIGN KEY (debtor_id) REFERENCES public.users(id);


--
-- Name: settlement_winners fk_debt_settlements_winners; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_winners
    ADD CONSTRAINT fk_debt_settlements_winners FOREIGN KEY (settlement_id) REFERENCES public.debt_settlements(id);


--
-- Name: match_participants fk_match_participants_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.match_participants
    ADD CONSTRAINT fk_match_participants_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: match_participants fk_matches_participants; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.match_participants
    ADD CONSTRAINT fk_matches_participants FOREIGN KEY (match_id) REFERENCES public.matches(id) ON DELETE CASCADE;


--
-- Name: settlement_winners fk_settlement_winners_winner; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.settlement_winners
    ADD CONSTRAINT fk_settlement_winners_winner FOREIGN KEY (winner_id) REFERENCES public.users(id);


--
-- Name: tournament_participants fk_tournament_participants_user; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_participants
    ADD CONSTRAINT fk_tournament_participants_user FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: tournament_matches fk_tournaments_matches; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_matches
    ADD CONSTRAINT fk_tournaments_matches FOREIGN KEY (tournament_id) REFERENCES public.tournaments(id) ON DELETE CASCADE;


--
-- Name: tournament_participants fk_tournaments_participants; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.tournament_participants
    ADD CONSTRAINT fk_tournaments_participants FOREIGN KEY (tournament_id) REFERENCES public.tournaments(id) ON DELETE CASCADE;


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: postgres
--

REVOKE USAGE ON SCHEMA public FROM PUBLIC;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

