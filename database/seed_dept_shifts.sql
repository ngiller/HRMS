-- Seed Department-Specific Shifts
DO $$
DECLARE
    it_dept_id UUID;
    fin_dept_id UUID;
    sales_dept_id UUID;
BEGIN
    -- Get Department IDs
    SELECT id INTO it_dept_id FROM departments WHERE code = 'IT' LIMIT 1;
    SELECT id INTO fin_dept_id FROM departments WHERE code = 'FIN' LIMIT 1;
    SELECT id INTO sales_dept_id FROM departments WHERE code = 'SALES' LIMIT 1;

    -- Seed IT Shifts
    IF it_dept_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'it-morning' AND department_id = it_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (it_dept_id, 'IT Pagi', 'it-morning', '06:00', '14:00', '10:00', '10:30', '#4F46E5', 'Shift Pagi khusus Departemen IT');
        END IF;
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'it-afternoon' AND department_id = it_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (it_dept_id, 'IT Siang', 'it-afternoon', '14:00', '22:00', '18:00', '18:30', '#F59E0B', 'Shift Siang khusus Departemen IT');
        END IF;
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'it-night' AND department_id = it_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (it_dept_id, 'IT Malam', 'it-night', '22:00', '06:00', '01:00', '01:30', '#10B981', 'Shift Malam khusus Departemen IT');
        END IF;
    END IF;

    -- Seed FIN Shifts
    IF fin_dept_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'fin-regular' AND department_id = fin_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (fin_dept_id, 'FIN Reguler', 'fin-regular', '08:00', '17:00', '12:00', '13:00', '#EC4899', 'Jam Kerja Reguler Keuangan');
        END IF;
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'fin-overtime' AND department_id = fin_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (fin_dept_id, 'FIN Overtime', 'fin-overtime', '17:00', '21:00', '19:00', '19:30', '#EF4444', 'Shift Lembur Keuangan');
        END IF;
    END IF;

    -- Seed Sales Shifts
    IF sales_dept_id IS NOT NULL THEN
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'sales-morning' AND department_id = sales_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (sales_dept_id, 'Sales Pagi', 'sales-morning', '07:00', '15:00', '11:00', '11:30', '#3B82F6', 'Shift Pagi Penjualan');
        END IF;
        IF NOT EXISTS (SELECT 1 FROM shifts WHERE code = 'sales-afternoon' AND department_id = sales_dept_id AND deleted_at IS NULL) THEN
            INSERT INTO shifts (department_id, name, code, start_time, end_time, break_start, break_end, color, description)
            VALUES (sales_dept_id, 'Sales Sore', 'sales-afternoon', '15:00', '23:00', '19:00', '19:30', '#8B5CF6', 'Shift Sore Penjualan');
        END IF;
    END IF;
END $$;
