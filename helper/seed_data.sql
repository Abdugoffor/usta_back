-- ============================================================
-- SEED: 2000 ta resume, 1000 ta vakansiya (kategoriya 1-18)
-- Ishlatish:  psql -U postgres -d main_service -f seed_data.sql
-- ============================================================

-- -------------------------
-- 1. RESUMES (2000 ta)
-- -------------------------
WITH inserted_resumes AS (
    INSERT INTO resumes (
        slug, user_id, region_id, district_id, mahalla_id,
        adress, name, photo, title, text, contact,
        price, experience_year, skills,
        views_count, is_active,
        created_at, updated_at
    )
    SELECT
        'resume-' || extract(epoch from clock_timestamp())::bigint || '-' || i,

        1::bigint,
        (floor(random() * 14)  + 1)::bigint,
        (floor(random() * 200) + 1)::bigint,
        (floor(random() * 500) + 1)::bigint,

        (ARRAY[
            'Toshkent shahar Chilonzor tumani',
            'Toshkent shahar Yunusobod tumani',
            'Samarqand shahar',
            'Andijon shahar',
            'Namangan shahar',
            'Farg''ona shahar'
        ])[1 + (i % 6)],

        (ARRAY[
            'Ali','Vali','Hasan','Husan','Jasur',
            'Sardor','Dilshod','Bekzod','Aziz','Sherzod',
            'Abdulloh','Muhammad','Oybek','Rustam','Shoxrux',
            'Nodir','Akmal','Sanjar','Umid','Temur'
        ])[1 + (i % 20)]
        || ' ' ||
        (ARRAY[
            'Karimov','Aliyev','Raximov','Qodirov','Abdullayev',
            'Tursunov','Ergashev','Sattorov','Bozorov','Yoqubov'
        ])[1 + ((i * 3) % 10)],

        '/uploads/1b702d0a4eebc2a5ecf2de0030e37872.png',

        (ARRAY[
            'Elektrik ustasi','Santexnik','Payvandchi','Kafelchi','Mebel ustasi',
            'Tom yopuvchi','Bo''yoqchi','G''isht teruvchi','Betonchi','Konditsioner ustasi',
            'Duradgor','Gidravlik usta','Kranovshchik','Ekskavatorchi','Quvur usta',
            'Isitish tizimi ustasi','Rozetkachi','Elevator ustasi'
        ])[1 + (i % 18)],

        (ARRAY[
            'Elektrik ustasi','Santexnik','Payvandchi','Kafelchi','Mebel ustasi',
            'Tom yopuvchi','Bo''yoqchi','G''isht teruvchi','Betonchi','Konditsioner ustasi',
            'Duradgor','Gidravlik usta','Kranovshchik','Ekskavatorchi','Quvur usta',
            'Isitish tizimi ustasi','Rozetkachi','Elevator ustasi'
        ])[1 + (i % 18)]
        || ' bo''yicha '
        || (floor(random() * 15) + 1)::int
        || ' yillik tajribaga ega ustaman',

        '+99890' || lpad((floor(random() * 9999999))::text, 7, '0'),

        ((floor(random() * 200) + 50)::bigint) * 10000,
        (floor(random() * 15) + 1)::int,

        (ARRAY[
            'Elektr montaj','Suv tizimi','Payvandlash','Kafel terish',
            'Mebel yig''ish','Tom yopish','Bo''yash','Beton ishlari'
        ])[1 + (i % 8)]
        || ', ' ||
        (ARRAY[
            'Elektr montaj','Suv tizimi','Payvandlash','Kafel terish',
            'Mebel yig''ish','Tom yopish','Bo''yash','Beton ishlari'
        ])[1 + ((i + 4) % 8)],

        (floor(random() * 10000))::bigint,
        (random() > 0.1),
        NOW() - (random() * interval '365 days'),
        NOW() - (random() * interval '30 days')

    FROM generate_series(1, 2000) AS t(i)
    RETURNING id
)
INSERT INTO category_resume (categorya_id, resume_id, created_at)
SELECT
    (floor(random() * 18) + 1)::bigint,
    id,
    NOW()
FROM inserted_resumes
ON CONFLICT (categorya_id, resume_id) DO NOTHING;

\echo '✅ 2000 ta resume va category_resume qoshildi'

-- -------------------------
-- 2. VACANCIES (1000 ta)
-- -------------------------
WITH inserted_vacancies AS (
    INSERT INTO vacancies (
        slug, user_id, region_id, district_id, mahalla_id,
        adress, name, title, text, contact,
        price, views_count, is_active,
        created_at, updated_at
    )
    SELECT
        'vacancy-' || extract(epoch from clock_timestamp())::bigint || '-' || i,

        1::bigint,
        (floor(random() * 14)  + 1)::bigint,
        (floor(random() * 200) + 1)::bigint,
        (floor(random() * 500) + 1)::bigint,

        (ARRAY[
            'Toshkent shahar Chilonzor tumani',
            'Toshkent shahar Yunusobod tumani',
            'Samarqand shahar',
            'Andijon shahar',
            'Namangan shahar',
            'Farg''ona shahar'
        ])[1 + (i % 6)],

        (ARRAY[
            'Ali','Vali','Hasan','Husan','Jasur',
            'Sardor','Dilshod','Bekzod','Aziz','Sherzod',
            'Abdulloh','Muhammad','Oybek','Rustam','Shoxrux',
            'Nodir','Akmal','Sanjar','Umid','Temur'
        ])[1 + (i % 20)]
        || ' ' ||
        (ARRAY[
            'Karimov','Aliyev','Raximov','Qodirov','Abdullayev',
            'Tursunov','Ergashev','Sattorov','Bozorov','Yoqubov'
        ])[1 + ((i * 7) % 10)],

        (ARRAY[
            'Elektrik kerak','Santexnik kerak','Payvandchi kerak','Kafelchi kerak',
            'Mebel ustasi kerak','Tom yopuvchi kerak','Bo''yoqchi kerak',
            'G''isht teruvchi kerak','Betonchi kerak','Konditsioner ustasi kerak',
            'Duradgor kerak','Gidravlik usta kerak','Kranovshchik kerak','Ekskavatorchi kerak',
            'Quvur usta kerak','Isitish tizimi ustasi kerak','Rozetkachi kerak','Elevator ustasi kerak'
        ])[1 + (i % 18)],

        (ARRAY[
            'Elektrik kerak','Santexnik kerak','Payvandchi kerak','Kafelchi kerak',
            'Mebel ustasi kerak','Tom yopuvchi kerak','Bo''yoqchi kerak',
            'G''isht teruvchi kerak','Betonchi kerak','Konditsioner ustasi kerak',
            'Duradgor kerak','Gidravlik usta kerak','Kranovshchik kerak','Ekskavatorchi kerak',
            'Quvur usta kerak','Isitish tizimi ustasi kerak','Rozetkachi kerak','Elevator ustasi kerak'
        ])[1 + (i % 18)]
        || ' uchun tajribali odam kerak',

        '+99891' || lpad((floor(random() * 9999999))::text, 7, '0'),

        ((floor(random() * 300) + 100)::bigint) * 10000,
        (floor(random() * 20000))::bigint,
        (random() > 0.1),
        NOW() - (random() * interval '365 days'),
        NOW() - (random() * interval '30 days')

    FROM generate_series(1, 1000) AS t(i)
    RETURNING id
)
INSERT INTO category_vacancy (categorya_id, vacancy_id, created_at)
SELECT
    (floor(random() * 18) + 1)::bigint,
    id,
    NOW()
FROM inserted_vacancies
ON CONFLICT (categorya_id, vacancy_id) DO NOTHING;

\echo '✅ 1000 ta vakansiya va category_vacancy qoshildi'
\echo '🎉 Jami: 2000 resume + 1000 vakansiya saqlandi!'
