-- +goose Up
-- ============================================================
-- Migration 00018: Flexible Overtime Rate Configuration
-- ============================================================
-- Adds 3-level overtime rate configuration:
--   Level 1: Company default (JSONB in companies.hr_settings)
--   Level 2: Position grade override (position_grade_overtime_rates)
--   Level 3: Per employee override (employee_overtime_rates)
-- Also updates the overtime_calculation view to use this hierarchy.
-- ============================================================

-- ============================================================
-- 1. Position Grade Overtime Rates (Level 2)
-- Override untuk level jabatan tertentu
-- ============================================================
CREATE TABLE position_grade_overtime_rates (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    position_grade_id   UUID NOT NULL REFERENCES position_grades(id) ON DELETE CASCADE,
    overtime_type       overtime_type NOT NULL,
    hour_from           INTEGER NOT NULL CHECK (hour_from >= 1),     -- Mulai jam ke-
    hour_to             INTEGER CHECK (hour_to IS NULL OR hour_to >= hour_from), -- Sampai jam ke- (NULL = tak terbatas)
    multiplier          DECIMAL(4,2) NOT NULL CHECK (multiplier >= 0),   -- Pengali (1.0, 1.5, 2.0, 2.5, dst)
    is_active           BOOLEAN DEFAULT TRUE,
    effective_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(position_grade_id, overtime_type, hour_from)
);

CREATE TRIGGER set_position_grade_overtime_rates_updated_at
    BEFORE UPDATE ON position_grade_overtime_rates
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 2. Employee Overtime Rates (Level 3)
-- Override individu per karyawan
-- ============================================================
CREATE TABLE employee_overtime_rates (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    overtime_type       overtime_type NOT NULL,
    hour_from           INTEGER NOT NULL CHECK (hour_from >= 1),
    hour_to             INTEGER CHECK (hour_to IS NULL OR hour_to >= hour_from),
    multiplier          DECIMAL(4,2) NOT NULL CHECK (multiplier >= 0),
    is_active           BOOLEAN DEFAULT TRUE,
    effective_date      DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(employee_id, overtime_type, hour_from)
);

CREATE TRIGGER set_employee_overtime_rates_updated_at
    BEFORE UPDATE ON employee_overtime_rates
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- 3. Update companies.hr_settings to include overtime defaults
-- ============================================================
UPDATE companies
SET hr_settings = hr_settings || '{
    "overtime": {
        "hourly_rate_method": "base_salary / 173",
        "default_rates": {
            "weekday": [
                {"hour_from": 1, "hour_to": 1, "multiplier": 1.5},
                {"hour_from": 2, "hour_to": null, "multiplier": 2.0}
            ],
            "weekend": [
                {"hour_from": 1, "hour_to": 7, "multiplier": 2.0},
                {"hour_from": 8, "hour_to": null, "multiplier": 3.0}
            ],
            "holiday": [
                {"hour_from": 1, "hour_to": 7, "multiplier": 2.0},
                {"hour_from": 8, "hour_to": null, "multiplier": 3.0}
            ]
        }
    }
}'::jsonb;

-- ============================================================
-- 4. Update overtime_calculation view to use flexible rates
-- ============================================================
DROP VIEW IF EXISTS overtime_calculation;

CREATE OR REPLACE VIEW overtime_calculation AS
WITH employee_rate_segments AS (
    -- Get rate segments for each employee based on 3-level hierarchy
    SELECT 
        otr.id AS overtime_request_id,
        otr.employee_id,
        otr.total_hours,
        otr.overtime_type,
        esh.base_salary,
        ROUND(esh.base_salary / 173, 2) AS hourly_rate,
        COALESCE(
            -- Level 3: Per-employee override
            (SELECT jsonb_agg(
                jsonb_build_object(
                    'hour_from', eor.hour_from,
                    'hour_to', eor.hour_to,
                    'multiplier', eor.multiplier
                )
                ORDER BY eor.hour_from
             )
             FROM employee_overtime_rates eor
             WHERE eor.employee_id = otr.employee_id
               AND eor.overtime_type = otr.overtime_type
               AND eor.is_active = TRUE
               AND eor.effective_date <= otr.date
            ),
            -- Level 2: Position grade override
            (SELECT jsonb_agg(
                jsonb_build_object(
                    'hour_from', pgor.hour_from,
                    'hour_to', pgor.hour_to,
                    'multiplier', pgor.multiplier
                )
                ORDER BY pgor.hour_from
             )
             FROM employees e
             JOIN position_grade_overtime_rates pgor ON pgor.position_grade_id = e.position_grade_id
             WHERE e.id = otr.employee_id
               AND pgor.overtime_type = otr.overtime_type
               AND pgor.is_active = TRUE
               AND pgor.effective_date <= otr.date
            ),
            -- Level 1: Company default (from hr_settings)
            (SELECT jsonb_agg(
                jsonb_build_object(
                    'hour_from', (rate->>'hour_from')::INT,
                    'hour_to', (rate->>'hour_to')::INT,
                    'multiplier', (rate->>'multiplier')::DECIMAL
                )
                ORDER BY (rate->>'hour_from')::INT
             )
             FROM companies c,
             jsonb_array_elements(c.hr_settings->'overtime'->'default_rates'->otr.overtime_type::TEXT) AS rate
             LIMIT 1
            )
        ) AS rate_segments
    FROM overtime_requests otr
    LEFT JOIN LATERAL (
        SELECT base_salary FROM employee_salary_histories
        WHERE employee_id = otr.employee_id
        AND effective_date <= otr.date
        ORDER BY effective_date DESC
        LIMIT 1
    ) esh ON TRUE
    WHERE otr.status = 'approved'
)
SELECT 
    ers.overtime_request_id AS id,
    ers.employee_id,
    otr.date,
    ers.total_hours,
    ers.overtime_type,
    ers.base_salary,
    ers.hourly_rate,
    ers.rate_segments,
    -- Calculate overtime pay using rate segments
    (
        SELECT COALESCE(SUM(
            LEAST(
                -- Hours in this segment
                GREATEST(0, ers.total_hours - (seg->>'hour_from')::INT + 1),
                -- Max hours for this segment
                CASE 
                    WHEN (seg->>'hour_to') IS NULL OR (seg->>'hour_to')::INT >= ers.total_hours 
                    THEN ers.total_hours - (seg->>'hour_from')::INT + 1
                    ELSE (seg->>'hour_to')::INT - (seg->>'hour_from')::INT + 1
                END
            )
            * ers.hourly_rate 
            * (seg->>'multiplier')::DECIMAL
        ), 0)
        FROM jsonb_array_elements(ers.rate_segments) AS seg
    ) AS overtime_pay
FROM employee_rate_segments ers
JOIN overtime_requests otr ON otr.id = ers.overtime_request_id;

-- ============================================================
-- 5. Seed default position grade overtime rates
-- Staff (level 1-2): more granular weekday rates
-- ============================================================
INSERT INTO position_grade_overtime_rates (position_grade_id, overtime_type, hour_from, hour_to, multiplier)
SELECT pg.id, 'weekday', 1, 2, 1.5
FROM position_grades pg
WHERE pg.level <= 2
AND NOT EXISTS (
    SELECT 1 FROM position_grade_overtime_rates 
    WHERE position_grade_id = pg.id AND overtime_type = 'weekday' AND hour_from = 1
);

INSERT INTO position_grade_overtime_rates (position_grade_id, overtime_type, hour_from, hour_to, multiplier)
SELECT pg.id, 'weekday', 3, 4, 2.0
FROM position_grades pg
WHERE pg.level <= 2
AND NOT EXISTS (
    SELECT 1 FROM position_grade_overtime_rates 
    WHERE position_grade_id = pg.id AND overtime_type = 'weekday' AND hour_from = 3
);

INSERT INTO position_grade_overtime_rates (position_grade_id, overtime_type, hour_from, hour_to, multiplier)
SELECT pg.id, 'weekday', 5, NULL, 2.5
FROM position_grades pg
WHERE pg.level <= 2
AND NOT EXISTS (
    SELECT 1 FROM position_grade_overtime_rates 
    WHERE position_grade_id = pg.id AND overtime_type = 'weekday' AND hour_from = 5
);

-- ============================================================
-- 6. Indexes
-- ============================================================
CREATE INDEX idx_emp_overtime_rates_employee ON employee_overtime_rates(employee_id, overtime_type, is_active);
CREATE INDEX idx_emp_overtime_rates_active ON employee_overtime_rates(employee_id) WHERE is_active = TRUE;
CREATE INDEX idx_pos_grade_overtime_rates_grade ON position_grade_overtime_rates(position_grade_id, overtime_type, is_active);

-- +goose Down
DROP VIEW IF EXISTS overtime_calculation;
DROP TABLE IF EXISTS employee_overtime_rates;
DROP TABLE IF EXISTS position_grade_overtime_rates;

-- Restore old overtime_calculation view (from migration 00008)
CREATE OR REPLACE VIEW overtime_calculation AS
SELECT
    otr.id,
    otr.employee_id,
    otr.date,
    otr.total_hours,
    otr.overtime_type,
    esh.base_salary,
    ROUND(esh.base_salary / 173, 2) AS hourly_rate,
    CASE
        WHEN otr.overtime_type = 'weekday' THEN
            CASE
                WHEN otr.total_hours <= 1 THEN ROUND((esh.base_salary / 173) * 1.5 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 1.5 * 1 + (esh.base_salary / 173) * 2 * (otr.total_hours - 1), 2)
            END
        WHEN otr.overtime_type = 'weekend' THEN
            CASE
                WHEN otr.total_hours <= 7 THEN ROUND((esh.base_salary / 173) * 2 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 2 * 7 + (esh.base_salary / 173) * 3 * (otr.total_hours - 7), 2)
            END
        WHEN otr.overtime_type = 'holiday' THEN
            CASE
                WHEN otr.total_hours <= 7 THEN ROUND((esh.base_salary / 173) * 2 * otr.total_hours, 2)
                ELSE ROUND((esh.base_salary / 173) * 2 * 7 + (esh.base_salary / 173) * 3 * (otr.total_hours - 7), 2)
            END
    END AS overtime_pay
FROM overtime_requests otr
LEFT JOIN LATERAL (
    SELECT base_salary FROM employee_salary_histories
    WHERE employee_id = otr.employee_id
    AND effective_date <= otr.date
    ORDER BY effective_date DESC
    LIMIT 1
) esh ON TRUE
WHERE otr.status = 'approved';
