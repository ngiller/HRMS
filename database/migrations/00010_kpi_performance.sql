-- +goose Up
-- ============================================================
-- Migration 00010: KPI & Performance Review
-- ============================================================

-- ============================================================
-- KPI Templates (per position / department)
-- ============================================================
CREATE TABLE kpi_templates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title           VARCHAR(255) NOT NULL,
    position_id     UUID REFERENCES positions(id) ON DELETE SET NULL,
    department_id   UUID REFERENCES departments(id) ON DELETE SET NULL,
    period_type     VARCHAR(20) NOT NULL DEFAULT 'yearly',  -- yearly, quarterly
    year            INTEGER NOT NULL,
    description     TEXT,
    is_active       BOOLEAN DEFAULT TRUE,
    created_by      UUID REFERENCES employees(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE TRIGGER set_kpi_templates_updated_at
    BEFORE UPDATE ON kpi_templates
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- KPI Indicators (target per template)
-- ============================================================
CREATE TABLE kpi_indicators (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kpi_template_id     UUID NOT NULL REFERENCES kpi_templates(id) ON DELETE CASCADE,
    name                VARCHAR(255) NOT NULL,
    description         TEXT,
    target              DECIMAL(15,2) NOT NULL,              -- Target value
    weight              DECIMAL(5,2) NOT NULL,              -- Bobot dalam persen (e.g., 20.00 = 20%)
    measurement_unit    VARCHAR(50),                         -- %, Rp, unit, score, dll
    sort_order          INTEGER DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TRIGGER set_kpi_indicators_updated_at
    BEFORE UPDATE ON kpi_indicators
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- KPI Reviews (penilaian per employee per periode)
-- ============================================================
CREATE TABLE kpi_reviews (
    id                  UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    employee_id         UUID NOT NULL REFERENCES employees(id) ON DELETE CASCADE,
    kpi_template_id     UUID NOT NULL REFERENCES kpi_templates(id) ON DELETE RESTRICT,
    
    period              VARCHAR(10) NOT NULL,                 -- Q1, Q2, Q3, Q4, or 'yearly'
    year                INTEGER NOT NULL,
    
    -- Self Assessment (JSON: [{indicator_id, score, note}])
    self_rating         JSONB DEFAULT '[]'::jsonb,
    self_score          DECIMAL(5,2),                         -- Calculated weighted score
    self_note           TEXT,
    self_submitted_at   TIMESTAMPTZ,
    
    -- Manager Review
    manager_rating      JSONB DEFAULT '[]'::jsonb,
    manager_score       DECIMAL(5,2),
    manager_note        TEXT,
    manager_id          UUID REFERENCES employees(id) ON DELETE SET NULL,
    manager_reviewed_at TIMESTAMPTZ,
    
    -- HR Review / Final
    hr_rating           JSONB DEFAULT '[]'::jsonb,
    final_score         DECIMAL(5,2),
    final_category      VARCHAR(30),                          -- outstanding, above, meets, needs_improvement, underperform
    hr_note             TEXT,
    hr_id               UUID REFERENCES employees(id) ON DELETE SET NULL,
    hr_reviewed_at      TIMESTAMPTZ,
    
    -- Status
    status              kpi_review_status NOT NULL DEFAULT 'draft',
    
    -- Impact
    salary_increase     DECIMAL(5,2),                         -- % kenaikan gaji
    bonus_amount        DECIMAL(15,2),                        -- Bonus kinerja
    promotion_recommended BOOLEAN DEFAULT FALSE,
    
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ,
    
    UNIQUE(employee_id, kpi_template_id, period, year)
);

CREATE TRIGGER set_kpi_reviews_updated_at
    BEFORE UPDATE ON kpi_reviews
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_updated_at();

-- ============================================================
-- Function: Calculate final KPI score & category
-- ============================================================
CREATE OR REPLACE FUNCTION calculate_kpi_final_score()
RETURNS TRIGGER AS $$
DECLARE
    total_weight DECIMAL(5,2);
    weighted_score DECIMAL(5,2);
BEGIN
    -- If HR has submitted review, calculate final score
    IF NEW.hr_reviewed_at IS NOT NULL THEN
        -- Calculate weighted score from HR rating
        SELECT SUM((r.value->>'score')::DECIMAL * (ki.weight / 100))
        INTO weighted_score
        FROM jsonb_array_elements(NEW.hr_rating) AS r(value)
        JOIN kpi_indicators ki ON ki.id = (r.value->>'indicator_id')::UUID
        WHERE ki.kpi_template_id = NEW.kpi_template_id;
        
        NEW.final_score := ROUND(weighted_score, 2);
        
        -- Determine category
        NEW.final_category := CASE
            WHEN NEW.final_score >= 4.5 THEN 'outstanding'
            WHEN NEW.final_score >= 3.5 THEN 'above'
            WHEN NEW.final_score >= 2.5 THEN 'meets'
            WHEN NEW.final_score >= 1.5 THEN 'needs_improvement'
            ELSE 'underperform'
        END;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_kpi_reviews_calculate_score
    BEFORE INSERT OR UPDATE OF hr_reviewed_at ON kpi_reviews
    FOR EACH ROW
    EXECUTE FUNCTION calculate_kpi_final_score();

-- Indexes
CREATE INDEX idx_kpi_templates_position_id ON kpi_templates(position_id);
CREATE INDEX idx_kpi_templates_year ON kpi_templates(year);
CREATE INDEX idx_kpi_indicators_template_id ON kpi_indicators(kpi_template_id);
CREATE INDEX idx_kpi_reviews_employee_id ON kpi_reviews(employee_id);
CREATE INDEX idx_kpi_reviews_status ON kpi_reviews(status);
CREATE INDEX idx_kpi_reviews_period ON kpi_reviews(period, year);

-- +goose Down
DROP TRIGGER IF EXISTS trg_kpi_reviews_calculate_score ON kpi_reviews;
DROP FUNCTION IF EXISTS calculate_kpi_final_score;
DROP TABLE IF EXISTS kpi_reviews;
DROP TABLE IF EXISTS kpi_indicators;
DROP TABLE IF EXISTS kpi_templates;
