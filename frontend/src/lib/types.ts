/**
 * Shared TypeScript types for HRMS Frontend
 */

// ── API Response Types ──

export interface ApiResponse<T = unknown> {
	success: boolean;
	data?: T;
	error?: string;
	meta?: PaginationMeta;
}

export interface PaginationMeta {
	total: number;
	page: number;
	per_page: number;
	total_pages?: number;
}

// ── AG Grid Types ──

export interface AgGridCellParams<T = Record<string, unknown>> {
	value: unknown;
	data: T;
	rowIndex: number;
	api: unknown;
	colDef: unknown;
	context: unknown;
	node: unknown;
	column: unknown;
}

export interface AgGridValueParams {
	value: unknown;
	data: Record<string, unknown>;
}

// ── Common Entity Types ──

export interface Employee {
	id: string;
	full_name: string;
	employee_name?: string;
	email: string;
	employee_id?: string;
	department_name?: string;
	position_name?: string;
	status?: string;
	photo_url?: string;
}

export interface Department {
	id: string;
	name: string;
	code: string;
	parent_name: string;
	head_name: string;
	description: string;
	is_active: boolean;
	work_schedule_name: string;
	sort_order: number;
	employee_count: number;
	created_at: string;
	parent_id?: string;
	head_id?: string;
	work_schedule_id?: string;
}

export interface WorkSchedule {
	id: string;
	name: string;
	description?: string;
	start_time?: string;
	end_time?: string;
}

export interface LeaveRequest {
	id: string;
	employee_id: string;
	employee_name: string;
	leave_type_name: string;
	start_date: string;
	end_date: string;
	total_days: number;
	is_half_day: boolean;
	reason: string;
	status: string;
	created_at: string;
	contact_during_leave?: string;
	rejection_reason?: string;
	approval_trail?: string;
}

export interface LeaveType {
	id: string;
	name: string;
	code: string;
	default_quota: number;
	is_paid: boolean;
	is_active: boolean;
}

export interface LeaveBalance {
	id: string;
	leave_type_name: string;
	total_quota: number;
	used: number;
	remaining: number;
}

export interface Document {
	id: string;
	employee_id: string;
	employee_name: string;
	doc_type: string;
	file_name: string;
	title: string;
	status: string;
	expiry_date: string;
	created_at: string;
	description?: string;
	file_url?: string;
	verified_by_name?: string;
	verified_at?: string;
	rejection_reason?: string;
}

export interface Announcement {
	id: string;
	title: string;
	content: string;
	announcement_type: string;
	is_pinned: boolean;
	pin_priority: number;
	published_at: string;
	expired_at?: string;
	created_at: string;
	author_name?: string;
	target_all?: boolean;
	target_department_id?: string;
	attachment_urls?: string[];
}

export interface AttendanceLocation {
	id: string;
	name: string;
	address: string;
	latitude: number;
	longitude: number;
	radius: number;
	is_active: boolean;
}

export interface Loan {
	id: string;
	employee_id: string;
	employee_name: string;
	amount: number;
	installment: number;
	remaining: number;
	status: string;
	reason: string;
	created_at: string;
}

export interface OvertimeRequest {
	id: string;
	employee_id: string;
	employee_name?: string;
	date: string;
	start_time: string;
	end_time: string;
	total_hours: number;
	reason: string;
	status: string;
	created_at: string;
}

export interface Reimbursement {
	id: string;
	employee_id: string;
	employee_name?: string;
	title: string;
	amount: number;
	category: string;
	status: string;
	created_at: string;
	receipt_url?: string;
}

export interface ShiftChangeRequest {
	id: string;
	employee_id: string;
	employee_name?: string;
	target_employee_id?: string;
	target_employee_name?: string;
	date: string;
	shift_from?: string;
	shift_to?: string;
	reason: string;
	status: string;
	created_at: string;
}

export interface PayrollPeriod {
	id: string;
	name: string;
	period_month: number;
	period_year: number;
	start_date: string;
	end_date: string;
	status: string;
	created_at: string;
}

export interface Payslip {
	id: string;
	period_id: string;
	employee_id: string;
	employee_name?: string;
	gross_salary: number;
	net_salary: number;
	deductions: Record<string, number>;
	allowances: Record<string, number>;
	status: string;
}

export interface KpiTemplate {
	id: string;
	name: string;
	year: number;
	description?: string;
	indicators?: KpiIndicator[];
	is_active: boolean;
}

export interface KpiIndicator {
	id: string;
	name: string;
	target: number;
	weight: number;
	measurement_unit?: string;
}

export interface Reprimand {
	id: string;
	employee_id: string;
	employee_name?: string;
	title: string;
	description: string;
	level: number;
	status: string;
	issued_date: string;
	created_at: string;
}

export interface DailyJournal {
	id: string;
	employee_id: string;
	employee_name?: string;
	journal_date: string;
	work_description: string;
	status: string;
	created_at: string;
}

export interface OrganizationNode {
	id: string;
	name: string;
	type: string;
	children?: OrganizationNode[];
	employee_count?: number;
}

export interface Position {
	id: string;
	name: string;
	code: string;
	position_grade_id?: string;
	position_grade_name?: string;
	department_id?: string;
	department_name?: string;
	description?: string;
	is_active: boolean;
}

export interface PositionGrade {
	id: string;
	name: string;
	level: number;
	min_salary: number;
	max_salary: number;
	description?: string;
	is_active: boolean;
}

export interface Role {
	id: string;
	name: string;
	slug: string;
	description?: string;
	permissions?: Record<string, Record<string, boolean>>;
	is_system?: boolean;
}

export interface Notification {
	id: string;
	title: string;
	message: string;
	type: string;
	is_read: boolean;
	created_at: string;
}

export interface ActivityLog {
	id: string;
	user_id: string;
	user_name?: string;
	action: string;
	entity_type: string;
	entity_id: string;
	old_value?: Record<string, unknown>;
	new_value?: Record<string, unknown>;
	created_at: string;
}

export interface AttendanceRecord {
	id: string;
	employee_id: string;
	employee_name?: string;
	date: string;
	check_in: string;
	check_out: string | null;
	status: string;
	check_in_location_id?: string;
	check_out_location_id?: string;
}

export interface Holiday {
	id: string;
	name: string;
	date: string;
	type: string;
	is_recurring: boolean;
	description?: string;
}

export interface SalaryComponent {
	id: string;
	name: string;
	type: 'allowance' | 'deduction';
	amount: number;
	is_active: boolean;
}

// ── Form Types ──

export interface DepartmentForm {
	name: string;
	code: string;
	parent_id: string;
	head_id: string;
	work_schedule_id: string;
	description: string;
	is_active: boolean;
}

export interface LeaveForm {
	leave_type_id: string;
	start_date: string;
	end_date: string;
	total_days: number;
	is_half_day: boolean;
	reason: string;
	document_url: string;
	contact_during_leave: string;
}

export interface DocumentForm {
	employee_id: string;
	doc_type: string;
	title: string;
	description: string;
	expiry_date: string;
	file: File | null;
	file_name: string;
	file_url: string;
}
