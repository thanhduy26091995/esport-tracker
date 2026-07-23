-- ASEAN Cup 2026 (ASEAN Hyundai Cup) — Group Stage Match Seed
-- Source: ESPN / Wikipedia / official AFF schedule
-- Run AFTER the tournament_type migration has been applied.
--
-- Times converted to UTC from local venue timezone:
--   UTC+7  → Thailand, Vietnam, Laos, Cambodia, Indonesia (West)
--   UTC+8  → Singapore, Malaysia, Philippines
--   UTC+6:30 → Myanmar
--
-- Idempotent: skips entirely if any asean_cup matches already exist.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM wc_matches WHERE tournament_type = 'asean_cup') THEN
    INSERT INTO wc_matches (
      external_id,
      home_team, away_team,
      home_team_code, away_team_code,
      match_date,
      group_name, stage,
      venue,
      status,
      tournament_type,
      predictions_open,
      created_at, updated_at
    ) VALUES
    -- GROUP A: Vietnam · Singapore · Indonesia · Cambodia · Timor-Leste
    -- Matchday 1 — July 24
    ('ac2026-gA-r1-1', 'Cambodia',    'Singapore',   'CAM', 'SGP', '2026-07-24 12:00:00+00', 'Group A', 'group', 'Olympic National Stadium, Phnom Penh',      'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gA-r1-2', 'Timor-Leste', 'Vietnam',     'TLS', 'VIE', '2026-07-24 13:30:00+00', 'Group A', 'group', '700th Anniversary Stadium, Chonburi',       'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 2 — July 27
    ('ac2026-gA-r2-1', 'Singapore',   'Timor-Leste', 'SGP', 'TLS', '2026-07-27 11:00:00+00', 'Group A', 'group', 'Jalan Besar Stadium, Singapore',            'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gA-r2-2', 'Indonesia',   'Cambodia',    'IDN', 'CAM', '2026-07-27 13:30:00+00', 'Group A', 'group', 'Pakansari Stadium, Bogor',                   'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 3 — July 31
    ('ac2026-gA-r3-1', 'Timor-Leste', 'Indonesia',   'TLS', 'IDN', '2026-07-31 10:00:00+00', 'Group A', 'group', '700th Anniversary Stadium, Chonburi',       'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gA-r3-2', 'Vietnam',     'Singapore',   'VIE', 'SGP', '2026-07-31 13:00:00+00', 'Group A', 'group', 'My Dinh National Stadium, Hanoi',            'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 4 — August 3
    ('ac2026-gA-r4-1', 'Cambodia',    'Timor-Leste', 'CAM', 'TLS', '2026-08-03 10:30:00+00', 'Group A', 'group', 'Olympic National Stadium, Phnom Penh',      'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gA-r4-2', 'Indonesia',   'Vietnam',     'IDN', 'VIE', '2026-08-03 13:30:00+00', 'Group A', 'group', 'Pakansari Stadium, Bogor',                   'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 5 — August 7
    ('ac2026-gA-r5-1', 'Vietnam',     'Cambodia',    'VIE', 'CAM', '2026-08-07 13:00:00+00', 'Group A', 'group', 'My Dinh National Stadium, Hanoi',            'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gA-r5-2', 'Singapore',   'Indonesia',   'SGP', 'IDN', '2026-08-07 13:00:00+00', 'Group A', 'group', 'Jalan Besar Stadium, Singapore',            'scheduled', 'asean_cup', false, now(), now()),
    -- GROUP B: Thailand · Malaysia · Philippines · Myanmar · Laos
    -- Matchday 1 — July 25
    ('ac2026-gB-r1-1', 'Myanmar',     'Malaysia',    'MYA', 'MAS', '2026-07-25 10:00:00+00', 'Group B', 'group', 'Thuwunna Stadium, Yangon',                   'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gB-r1-2', 'Laos',        'Thailand',    'LAO', 'THA', '2026-07-25 13:00:00+00', 'Group B', 'group', 'New Laos National Stadium, Vientiane',       'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 2 — July 28
    ('ac2026-gB-r2-1', 'Philippines', 'Myanmar',     'PHI', 'MYA', '2026-07-28 10:00:00+00', 'Group B', 'group', 'New Clark City Athletics Stadium, Capas',    'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gB-r2-2', 'Malaysia',    'Laos',        'MAS', 'LAO', '2026-07-28 13:00:00+00', 'Group B', 'group', 'Bukit Jalil National Stadium, Kuala Lumpur', 'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 3 — August 1
    ('ac2026-gB-r3-1', 'Laos',        'Philippines', 'LAO', 'PHI', '2026-08-01 10:00:00+00', 'Group B', 'group', 'New Laos National Stadium, Vientiane',       'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gB-r3-2', 'Thailand',    'Malaysia',    'THA', 'MAS', '2026-08-01 13:00:00+00', 'Group B', 'group', 'Rajamangala National Stadium, Bangkok',      'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 4 — August 4
    ('ac2026-gB-r4-1', 'Myanmar',     'Laos',        'MYA', 'LAO', '2026-08-04 10:00:00+00', 'Group B', 'group', 'Thuwunna Stadium, Yangon',                   'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gB-r4-2', 'Philippines', 'Thailand',    'PHI', 'THA', '2026-08-04 13:00:00+00', 'Group B', 'group', 'New Clark City Athletics Stadium, Capas',    'scheduled', 'asean_cup', false, now(), now()),
    -- Matchday 5 — August 8
    ('ac2026-gB-r5-1', 'Thailand',    'Myanmar',     'THA', 'MYA', '2026-08-08 13:00:00+00', 'Group B', 'group', 'Rajamangala National Stadium, Bangkok',      'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-gB-r5-2', 'Malaysia',    'Philippines', 'MAS', 'PHI', '2026-08-08 13:00:00+00', 'Group B', 'group', 'Bukit Jalil National Stadium, Kuala Lumpur', 'scheduled', 'asean_cup', false, now(), now()),
    -- KNOCKOUT STAGE (teams TBD after Aug 8 group stage)
    ('ac2026-sf1-leg1',   'TBD (Runner-up A)', 'TBD (Winner B)',    'TBD', 'TBD', '2026-08-15 13:00:00+00', NULL, 'sf',    'TBD', 'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-sf1-leg2',   'TBD (Winner B)',    'TBD (Runner-up A)', 'TBD', 'TBD', '2026-08-18 13:00:00+00', NULL, 'sf',    'TBD', 'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-sf2-leg1',   'TBD (Runner-up B)', 'TBD (Winner A)',    'TBD', 'TBD', '2026-08-16 13:00:00+00', NULL, 'sf',    'TBD', 'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-sf2-leg2',   'TBD (Winner A)',    'TBD (Runner-up B)', 'TBD', 'TBD', '2026-08-19 13:00:00+00', NULL, 'sf',    'TBD', 'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-final-leg1', 'TBD (SF1 winner)',  'TBD (SF2 winner)',  'TBD', 'TBD', '2026-08-22 13:00:00+00', NULL, 'final', 'TBD', 'scheduled', 'asean_cup', false, now(), now()),
    ('ac2026-final-leg2', 'TBD (SF2 winner)',  'TBD (SF1 winner)',  'TBD', 'TBD', '2026-08-26 13:00:00+00', NULL, 'final', 'TBD', 'scheduled', 'asean_cup', false, now(), now());
  END IF;
END;
$$;
