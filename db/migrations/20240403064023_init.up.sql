CREATE SEQUENCE IF NOT EXISTS people_id_seq;
CREATE SEQUENCE IF NOT EXISTS car_id_seq;

CREATE TABLE IF NOT EXISTS public."people"
(
    id              BIGINT                 DEFAULT NEXTVAL('people_id_seq'::regclass) NOT NULL PRIMARY KEY,
    name           TEXT                                                               NOT NULL CHECK (name <> '')
    CONSTRAINT max_len_name CHECK (LENGTH(name) <= 64),
    surname           TEXT                                                            NOT NULL CHECK (surname <> '')
    CONSTRAINT max_len_surname CHECK (LENGTH(surname) <= 64),
    patronymic           TEXT                                                         DEFAULT NULL
    CONSTRAINT max_len_patronymic CHECK (LENGTH(patronymic) <= 64),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()                            NOT NULL
);

CREATE TABLE IF NOT EXISTS public."car"
(
    id              BIGINT                   DEFAULT NEXTVAL('car_id_seq'::regclass)     NOT NULL PRIMARY KEY,
    owner_id        BIGINT                                                               NOT NULL REFERENCES public."people" (id) ON DELETE CASCADE,
    reg_num         TEXT                                                                 NOT NULL UNIQUE CHECK (reg_num <> '')
    CONSTRAINT len_reg_num CHECK (CHAR_LENGTH(reg_num) = 9),
    mark            TEXT                                                                 NOT NULL CHECK (mark <> '')
    CONSTRAINT max_len_mark CHECK (LENGTH(mark) <= 256),
    model           TEXT                                                                 NOT NULL CHECK (model <> '')
    CONSTRAINT max_len_model CHECK (LENGTH(model) <= 256),
    year            INT                      DEFAULT NULL
    CONSTRAINT correct_year CHECK (year >= 1885),
    created_at      TIMESTAMP WITH TIME ZONE DEFAULT NOW()                               NOT NULL
);