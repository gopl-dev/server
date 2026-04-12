ALTER TABLE books ADD COLUMN release_date_sort DATE;

DO $$
    BEGIN
        SET LOCAL lc_time = 'en_US.UTF-8';

        UPDATE books
        SET release_date_sort = CASE
            -- "January 2, 2006"
                                    WHEN release_date ~* '^[a-z]+ \d{1,2}, \d{4}$'
                                        THEN TO_DATE(release_date, 'Month DD, YYYY')

            -- "January 2006"
                                    WHEN release_date ~* '^[a-z]+ \d{4}$'
                                        THEN TO_DATE(release_date, 'Month YYYY')

            -- "2006"
                                    ELSE TO_DATE(release_date, 'YYYY')
            END;
    END $$;

ALTER TABLE books ALTER COLUMN release_date_sort SET NOT NULL;

CREATE INDEX idx_books_release_date_sort ON books (release_date_sort DESC);