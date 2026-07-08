import { describe, it, expect, beforeEach, vi } from 'vitest';

// Mock the config module
vi.mock('$lib/config.js', () => ({
	default: {
		USER_DATA_KEY: 'hrms_user_data',
	},
}));

// Re-import after mocking
const { getUser, hasPermission, canAccess, getAccessibleMenus } = await import('../permissions.js');

describe('getUser', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns null when localStorage is empty', () => {
		expect(getUser()).toBeNull();
	});

	it('returns null when no session data exists', () => {
		localStorage.removeItem('hrms_user_data');
		expect(getUser()).toBeNull();
	});

	it('returns parsed user data when valid JSON is stored', () => {
		const mockUser = {
			id: '123',
			employee_id: 'EMP001',
			full_name: 'John Doe',
			email: 'john@example.com',
			role_slug: 'admin',
			role_name: 'Administrator',
			permissions: { employee: { create: true, read: true } },
		};
		localStorage.setItem('hrms_user_data', JSON.stringify(mockUser));
		expect(getUser()).toEqual(mockUser);
	});

	it('returns null when stored data is invalid JSON', () => {
		localStorage.setItem('hrms_user_data', '{invalid}');
		expect(getUser()).toBeNull();
	});

	it('returns parsed value for valid JSON string (caller should validate object shape)', () => {
		localStorage.setItem('hrms_user_data', '"string"');
		const result = getUser();
		// JSON.parse returns the parsed value; for valid JSON strings it returns the string
		expect(result).toBe('string');
	});
});

describe('hasPermission', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns false when no user is logged in', () => {
		expect(hasPermission('employee', 'create')).toBe(false);
	});

	it('returns false when user has no permissions', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
			}),
		);
		expect(hasPermission('employee', 'create')).toBe(false);
	});

	it('returns true when user has the specific permission', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'admin',
				role_name: 'Admin',
				permissions: { employee: { create: true } },
			}),
		);
		expect(hasPermission('employee', 'create')).toBe(true);
	});

	it('returns false when user does not have the specific permission', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: { employee: { read: true } },
			}),
		);
		expect(hasPermission('employee', 'create')).toBe(false);
	});

	it('returns false when module does not exist in permissions', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: { employee: { read: true } },
			}),
		);
		expect(hasPermission('payroll', 'read')).toBe(false);
	});

	it('returns false when action exists but is false', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: { employee: { delete: false } },
			}),
		);
		expect(hasPermission('employee', 'delete')).toBe(false);
	});
});

describe('canAccess', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns true when user has read permission for the module', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'admin',
				role_name: 'Admin',
				permissions: { employee: { read: true } },
			}),
		);
		expect(canAccess('employee')).toBe(true);
	});

	it('returns false when user lacks read permission', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: {},
			}),
		);
		expect(canAccess('employee')).toBe(false);
	});
});

describe('getAccessibleMenus', () => {
	beforeEach(() => {
		localStorage.clear();
	});

	it('returns empty groups when user has no permissions', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: {},
			}),
		);
		const menus = getAccessibleMenus();
		// Only Dashboard (no module requirement) should be accessible
		const dashboardGroup = menus.find((g) => g.group === '');
		expect(dashboardGroup?.items).toHaveLength(1);
		expect(dashboardGroup?.items[0].label).toBe('Dashboard');
	});

	it('includes menus from modules the user can access', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'admin',
				role_name: 'Admin',
				permissions: {
					employee: { read: true },
					department: { read: true },
				},
			}),
		);
		const menus = getAccessibleMenus();
		const kepegawaianGroup = menus.find((g) => g.group === 'Kepegawaian');
		expect(kepegawaianGroup).toBeDefined();
		expect(kepegawaianGroup!.items.some((m) => m.label === 'Karyawan')).toBe(true);
		expect(kepegawaianGroup!.items.some((m) => m.label === 'Departemen')).toBe(true);
		// Dokumen is also in Kepegawaian but requires "document" module
		expect(kepegawaianGroup!.items.some((m) => m.label === 'Dokumen Karyawan')).toBe(false);
	});

	it('excludes groups where all items are filtered out', () => {
		localStorage.setItem(
			'hrms_user_data',
			JSON.stringify({
				id: '1',
				employee_id: 'EMP001',
				full_name: 'John',
				email: 'john@test.com',
				role_slug: 'user',
				role_name: 'User',
				permissions: {},
			}),
		);
		const menus = getAccessibleMenus();
		const kompensasiGroup = menus.find((g) => g.group === 'Kompensasi & Benefit');
		expect(kompensasiGroup).toBeUndefined();
	});
});
