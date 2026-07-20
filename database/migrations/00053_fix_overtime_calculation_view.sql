-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW overtime_calculation AS
WITH employee_rate_segments AS (
    SELECT 
        otr.id AS overtime_request_id,
        otr.employee_id,
        otr.total_hours,
        otr.overtime_type,
        esh.base_salary,
        ROUND(esh.base_salary / 173, 2) AS hourly_rate,
        COALESCE(
            -- 1. Check custom rate for employee
            (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'hour_from', eor.hour_from,
                        'hour_to', eor.hour_to,
                        'multiplier', eor.multiplier
                    ) ORDER BY eor.hour_from
                )
                FROM employee_overtime_rates eor
                WHERE eor.employee_id = otr.employee_id
                  AND eor.overtime_type = otr.overtime_type
                  AND eor.is_active = true
                  AND eor.effective_date <= otr.date
            ),
            -- 2. Check position grade rate
            (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'hour_from', pgor.hour_from,
                        'hour_to', pgor.hour_to,
                        'multiplier', pgor.multiplier
                    ) ORDER BY pgor.hour_from
                )
                FROM employees e
                JOIN position_grade_overtime_rates pgor ON pgor.position_grade_id = e.position_grade_id
                WHERE e.id = otr.employee_id
                  AND pgor.overtime_type = otr.overtime_type
                  AND pgor.is_active = true
                  AND pgor.effective_date <= otr.date
            ),
            -- 3. Fallback to company default rate
            (
                SELECT jsonb_agg(
                    jsonb_build_object(
                        'hour_from', (rate.value->>'hour_from')::INTEGER,
                        'hour_to', (rate.value->>'hour_to')::INTEGER,
                        'multiplier', (rate.value->>'multiplier')::DECIMAL
                    ) ORDER BY (rate.value->>'hour_from')::INTEGER
                )
                FROM companies c,
                     jsonb_array_elements(c.hr_settings->'overtime'->'default_rates'->(otr.overtime_type::TEXT)) AS rate
                LIMIT 1
            )
        ) AS rate_segments
    FROM overtime_requests otr
    LEFT JOIN LATERAL (
        SELECT base_salary
        FROM (
            (
                SELECT base_salary
                FROM employee_salary_histories
                WHERE employee_id = otr.employee_id AND effective_date <= otr.date
                ORDER BY effective_date DESC
                LIMIT 1
            )
            UNION ALL
            (
                SELECT base_salary
                FROM employee_salary_histories
                WHERE employee_id = otr.employee_id AND effective_date > otr.date
                ORDER BY effective_date ASC
                LIMIT 1
            )
        ) AS fallback_salary
        LIMIT 1
    ) esh ON true
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
    (
        SELECT COALESCE(SUM(
            -- Calculate hours applicable to this segment
            LEAST(
                GREATEST(0, ers.total_hours - (seg.value->>'hour_from')::INTEGER + 1),
                CASE 
                    WHEN (seg.value->>'hour_to') IS NULL OR (seg.value->>'hour_to')::INTEGER >= ers.total_hours THEN ers.total_hours - (seg.value->>'hour_from')::INTEGER + 1
                    ELSE (seg.value->>'hour_to')::INTEGER - (seg.value->>'hour_from')::INTEGER + 1
                END
            ) * ers.hourly_rate * (seg.value->>'multiplier')::DECIMAL
        ), 0)
        FROM jsonb_array_elements(ers.rate_segments) AS seg
    ) AS overtime_pay
FROM employee_rate_segments ers
JOIN overtime_requests otr ON otr.id = ers.overtime_request_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Not providing down migration for simplicity
-- +goose StatementEnd
