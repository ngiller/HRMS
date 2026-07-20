/**
 * HRMS API Client
 *
 * Wrapper untuk fetch API dengan JWT authentication,
 * auto-refresh token, error handling, dan session management.
 */

import config from '$lib/config.js';

/**
 * @typedef {Object} ApiResponse
 * @property {boolean} success
 * @property {*} [data]
 * @property {string} [error]
 * @property {Object} [meta]
 */

/**
 * @typedef {Object} RequestOptions
 * @property {*} [body]
 * @property {string} [method]
 * @property {Object} [headers]
 * @property {boolean} [auth]
 */

/**
 * @typedef {Object} HistoryOptions
 * @property {number} [page]
 * @property {number} [per_page]
 */

/**
 * @typedef {Object} PaginatedResponse
 * @property {boolean} success
 * @property {*} data
 * @property {Object} meta
 * @property {number} meta.total
 * @property {number} meta.page
 * @property {number} meta.per_page
 */

// --------------- utility helpers ---------------

/** @type {Promise<boolean>|null} */
let refreshPromise = null;

/**
 * Coba refresh token lewat API backend (token rotation).
 * @returns {Promise<boolean>} true jika berhasil
 */
async function tryRefreshToken() {
	const refreshToken =
		typeof localStorage !== 'undefined'
			? localStorage.getItem(config.REFRESH_TOKEN_KEY)
			: null;
	if (!refreshToken) return false;

	try {
		const res = await fetch(`${config.API_BASE_URL}/api/auth/refresh`, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ refresh_token: refreshToken }),
		});
		const data = await res.json();
		if (!res.ok || !data.success) return false;

		const { access_token, refresh_token: newRefresh } = data.data;
		if (access_token) localStorage.setItem(config.ACCESS_TOKEN_KEY, access_token);
		if (newRefresh) localStorage.setItem(config.REFRESH_TOKEN_KEY, newRefresh);
		return true;
	} catch {
		/* silently fail */
		return false;
	}
}

/** Redirect ke halaman login (gunakan window.location) */
function redirectToLogin() {
	clearSession();
	if (typeof window !== 'undefined') {
		window.location.href = '/login';
	}
}

// --------------- localStorage helpers ---------------

function getAccessToken() {
	return typeof localStorage !== 'undefined'
		? localStorage.getItem(config.ACCESS_TOKEN_KEY)
		: null;
}

function getRefreshToken() {
	return typeof localStorage !== 'undefined'
		? localStorage.getItem(config.REFRESH_TOKEN_KEY)
		: null;
}

function clearSession() {
	if (typeof localStorage === 'undefined') return;
	localStorage.removeItem(config.ACCESS_TOKEN_KEY);
	localStorage.removeItem(config.REFRESH_TOKEN_KEY);
	localStorage.removeItem(config.USER_DATA_KEY);
}

// --------------- core request ---------------

/**
 * Base API request function with auto-refresh on 401.
 * Supports both JSON and FormData bodies.
 * @param {string} endpoint
 * @param {RequestOptions} [options]
 * @returns {Promise<ApiResponse>}
 */
async function request(endpoint, options = {}) {
	const url = `${config.API_BASE_URL}${endpoint}`;
	const { body, method = 'GET', headers = {}, auth = true } = options;

	const isFormData = body instanceof FormData;

	/** @type {Record<string, any>} */
	const requestHeaders = {
		...headers,
	};
	// Only set Content-Type for JSON, not FormData (browser sets with boundary)
	if (!isFormData) {
		requestHeaders['Content-Type'] = 'application/json';
	}

	let token;
	if (auth) {
		token = getAccessToken();
		if (token) {
			requestHeaders['Authorization'] = `Bearer ${token}`;
		}
	}

	/** @type {RequestInit} */
	const fetchOptions = {
		method,
		headers: requestHeaders,
	};
	if (body) {
		fetchOptions.body = isFormData ? body : JSON.stringify(body);
	}

	async function doFetch() {
		const response = await fetch(url, fetchOptions);
		const isJson = response.headers.get('content-type')?.includes('json');
		const data = isJson ? await response.json() : await response.text();

		if (!response.ok) {
			const rawError = data?.message || data?.error || data || 'Terjadi kesalahan';
			const errorMsg =
				typeof rawError === 'object' && rawError !== null
					? rawError.message || 'Terjadi kesalahan'
					: rawError;
			throw new ApiError(errorMsg, response.status, data);
		}
		return data;
	}

	try {
		return await doFetch();
	} catch (error) {
		// Auto-refresh: hanya untuk 401 + ada refresh token + request pake auth
		if (error instanceof ApiError && error.status === 401 && auth) {
			const rt = getRefreshToken();
			if (rt) {
				// Deduplicate concurrent refresh attempts
				if (!refreshPromise) {
					refreshPromise = tryRefreshToken();
				}

				try {
					const ok = await refreshPromise;
					refreshPromise = null;

					if (ok) {
						// Update auth header with new token & retry
						const newToken = getAccessToken();
						if (newToken) {
							fetchOptions.headers = {
								...requestHeaders,
								Authorization: `Bearer ${newToken}`,
							};
						}
						return await doFetch();
					}
				} catch {
					refreshPromise = null;
				}
			}

			// Refresh gagal atau tidak ada refresh token
			redirectToLogin();
		}

		if (error instanceof ApiError) {
			throw error;
		}
		throw new ApiError('Gagal terhubung ke server', 0);
	}
}

// --------------- custom error ---------------

class ApiError extends Error {
	/**
	 * @param {string} message
	 * @param {number} status
	 * @param {*} [data]
	 */
	constructor(message, status, data = null) {
		super(message);
		this.name = 'ApiError';
		this.status = status;
		this.data = data;
	}
	/** @type {number} */
	status;
	/** @type {*} */
	data;
}

// --------------- public API ---------------

export const auth = {
	/** @param {string} email @param {string} password */
	login(email, password) {
		return request('/api/auth/login', {
			method: 'POST',
			body: { email, password },
			auth: false,
		});
	},

	/** @returns {string|null} */
	getAccessToken() {
		return getAccessToken();
	},

	/** @param {string} email */
	forgotPassword(email) {
		return request('/api/auth/forgot-password', {
			method: 'POST',
			body: { email },
			auth: false,
		});
	},

	/** @param {string} token @param {string} newPassword */
	resetPassword(token, newPassword) {
		return request('/api/auth/reset-password', {
			method: 'POST',
			body: { token, new_password: newPassword },
			auth: false,
		});
	},

	/** @returns {Promise<ApiResponse>} */
	getMe() {
		return request('/api/auth/me');
	},

	async refreshSession() {
		try {
			const res = await request('/api/auth/me');
			if (res && res.data) {
				const currentUser = this.getUser();
				const newUser = { ...currentUser, ...res.data };
				localStorage.setItem(config.USER_DATA_KEY, JSON.stringify(newUser));
				if (typeof window !== 'undefined') {
					window.dispatchEvent(new CustomEvent('auth:session_update'));
				}
			}
		} catch (e) {
			console.error('Failed to refresh session:', e);
		}
	},

	/**
	 * Refresh token secara manual (internal auto-refresh pake tryRefreshToken()).
	 * @returns {Promise<ApiResponse>}
	 */
	refreshToken() {
		const rt = getRefreshToken();
		if (!rt) {
			return Promise.reject(new ApiError('Session tidak ditemukan, silakan login ulang', 401));
		}
		return request('/api/auth/refresh', {
			method: 'POST',
			body: { refresh_token: rt },
			auth: false,
		});
	},

	/**
	 * Logout — invalidate session di backend + hapus localstorage.
	 * @returns {Promise<void>}
	 */
	async logout() {
		const rt = getRefreshToken();
		try {
			await request('/api/auth/logout', {
				method: 'POST',
				body: rt ? { refresh_token: rt } : undefined,
			});
		} catch {
			// Tetap hapus session walau request gagal
		} finally {
			clearSession();
		}
	},

	/**
	 * Simpan token & user data setelah login.
	 * @param {ApiResponse} response
	 */
	saveSession(response) {
		if (!response || !response.data) return;
		const { access_token, refresh_token, user } = response.data;
		if (access_token) localStorage.setItem(config.ACCESS_TOKEN_KEY, access_token);
		if (refresh_token) localStorage.setItem(config.REFRESH_TOKEN_KEY, refresh_token);
		if (user) localStorage.setItem(config.USER_DATA_KEY, JSON.stringify(user));
	},

	/** Hapus session dari localstorage */
	clearSession() {
		clearSession();
	},

	/** @returns {Object|null} */
	getUser() {
		try {
			const data =
				typeof localStorage !== 'undefined'
					? localStorage.getItem(config.USER_DATA_KEY)
					: null;
			return data ? JSON.parse(data) : null;
		} catch {
			return null;
		}
	},

	/** @returns {boolean} */
	isAuthenticated() {
		return typeof localStorage !== 'undefined' && !!localStorage.getItem(config.ACCESS_TOKEN_KEY);
	},
};

export const employees = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] @param {string} [departmentID] @param {string} [status] @param {boolean} [includeDeleted] */
	list(page = 1, perPage = 25, search = '', departmentID = '', status = '', includeDeleted = false) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		if (departmentID) params.set('department_id', departmentID);
		if (status) params.set('status', status);
		if (includeDeleted) params.set('include_deleted', 'true');
		return request(`/api/employees?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/employees/${id}`);
	},

	/** @param {string} id @param {HistoryOptions} [opts] */
	getHistory(id, opts = {}) {
		const params = new URLSearchParams();
		if (opts.page) params.set('page', String(opts.page));
		if (opts.per_page) params.set('per_page', String(opts.per_page));
		return request(`/api/employees/${id}/history?${params}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/employees', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/employees/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/employees/${id}`, {
			method: 'DELETE',
		});
	},

	/** Restore soft-deleted employee
	 * @param {string} id */
	restore(id) {
		return request(`/api/employees/${id}/restore`, {
			method: 'PUT',
		});
	},

	/**
	 * Export employees to Excel — returns blob
	 * @returns {Promise<Blob>}
	 */
	async exportExcel() {
		const token = getAccessToken();
		const response = await fetch(`${config.API_BASE_URL}/api/employees/export`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		});
		if (!response.ok) {
			const data = await response.json().catch(() => ({}));
			throw new ApiError(data.message || 'Gagal export data', response.status);
		}
		return response.blob();
	},

	/**
	 * Download employee import template — returns blob
	 * @returns {Promise<Blob>}
	 */
	async downloadTemplate() {
		const token = getAccessToken();
		const response = await fetch(`${config.API_BASE_URL}/api/employees/template`, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		});
		if (!response.ok) {
			const data = await response.json().catch(() => ({}));
			throw new ApiError(data.message || 'Gagal download template', response.status);
		}
		return response.blob();
	},

	/**
	 * Import employees from Excel file
	 * @param {File} file
	 * @returns {Promise<ApiResponse>}
	 */
	importExcel(file) {
		const formData = new FormData();
		formData.append('file', file);
		return request('/api/employees/import', {
			method: 'POST',
			body: formData,
		});
	},

	/**
	 * Update individual work schedule override
	 * @param {string} id
	 * @param {string} workScheduleId
	 * @returns {Promise<ApiResponse>}
	 */
	updateWorkSchedule(id, workScheduleId) {
		return request(`/api/employees/${id}/work-schedule`, {
			method: 'PUT',
			body: { work_schedule_id: workScheduleId },
		});
	},

	/**
	 * Upload employee photo
	 * @param {string} id
	 * @param {File} file
	 * @returns {Promise<ApiResponse>}
	 */
	uploadPhoto(id, file) {
		const formData = new FormData();
		formData.append('photo', file);
		return request(`/api/employees/${id}/photo`, {
			method: 'POST',
			body: formData,
		});
	},
};

export const departments = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/departments?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/departments/${id}`);
	},

	/**
	 * Get all departments (for dropdowns, no pagination).
	 * @returns {Promise<ApiResponse>}
	 */
	getAll() {
		return request('/api/departments/all');
	},

	/**
	 * Get all work schedules (for dropdown).
	 * @returns {Promise<ApiResponse>}
	 */
	getWorkSchedules() {
		return request('/api/departments/work-schedules');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/departments', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/departments/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/**
	 * Update work schedule for a department
	 * @param {string} id
	 * @param {string} workScheduleId
	 */
	updateWorkSchedule(id, workScheduleId) {
		return request(`/api/departments/${id}/work-schedule`, {
			method: 'PUT',
			body: { work_schedule_id: workScheduleId },
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/departments/${id}`, {
			method: 'DELETE',
		});
	},
};

export const roles = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/roles?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/roles/${id}`);
	},

	/**
	 * Get available permissions template.
	 * @returns {Promise<ApiResponse>}
	 */
	getPermissionTemplate() {
		return request('/api/roles/permissions/template');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/roles', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/roles/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/roles/${id}`, {
			method: 'DELETE',
		});
	},
};

/** @typedef {{ id: string; name: string; [key: string]: any }} BasicItem */

export const positionGrades = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/position-grades?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/position-grades/${id}`);
	},

	/** @returns {Promise<ApiResponse>} */
	getAll() {
		return request('/api/position-grades/all');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/position-grades', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/position-grades/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/position-grades/${id}`, {
			method: 'DELETE',
		});
	},
};

export const positions = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/positions?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/positions/${id}`);
	},

	/** @returns {Promise<ApiResponse>} */
	getAll() {
		return request('/api/positions/all');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/positions', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/positions/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/positions/${id}`, {
			method: 'DELETE',
		});
	},
};

export const workSchedules = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/work-schedules?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/work-schedules/${id}`);
	},

	/** @returns {Promise<ApiResponse>} */
	getAll() {
		return request('/api/work-schedules/all');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/work-schedules', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/work-schedules/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/work-schedules/${id}`, {
			method: 'DELETE',
		});
	},
};

export const attendanceLocations = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/attendance-locations?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/attendance-locations/${id}`);
	},

	/** @returns {Promise<ApiResponse>} */
	getAll() {
		return request('/api/attendance-locations/all');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/attendance-locations', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/attendance-locations/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/attendance-locations/${id}`, {
			method: 'DELETE',
		});
	},
};	export const attendance = {
	/** @returns {Promise<ApiResponse>} */
	getTodayStatus() {
		return request(`/api/attendance/today?_t=${Date.now()}`);
	},

	/** @param {Object} data */
	checkIn(data) {
		return request('/api/attendance/check-in', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {Object} data */
	checkOut(data) {
		return request('/api/attendance/check-out', {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {number} [page] @param {number} [perPage] @param {string} [dateFrom] @param {string} [dateTo] */
	myHistory(page = 1, perPage = 25, dateFrom = '', dateTo = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (dateFrom) params.set('date_from', dateFrom);
		if (dateTo) params.set('date_to', dateTo);
		params.set('_t', String(Date.now()));
		return request(`/api/attendance/my-history?${params}`);
	},

	/** @param {number} [page] @param {number} [perPage] @param {string} [deptId] @param {string} [employeeId] @param {string} [status] @param {string} [dateFrom] @param {string} [dateTo] */
	report(page = 1, perPage = 25, deptId = '', employeeId = '', status = '', dateFrom = '', dateTo = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (deptId) params.set('department_id', deptId);
		if (employeeId) params.set('employee_id', employeeId);
		if (status) params.set('status', status);
		if (dateFrom) params.set('date_from', dateFrom);
		if (dateTo) params.set('date_to', dateTo);
		return request(`/api/attendance/report?${params}`);
	},

	/** @param {string} id */
	getRecord(id) {
		return request(`/api/attendance/${id}`, { auth: true });
	},

	/**
	 * Export attendance report as Excel
	 * @param {string} [deptId] @param {string} [employeeId] @param {string} [status] @param {string} [dateFrom] @param {string} [dateTo]
	 * @returns {Promise<Blob>}
	 */
	async exportReport(deptId = '', employeeId = '', status = '', dateFrom = '', dateTo = '') {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem(config.ACCESS_TOKEN_KEY) : null;
		const params = new URLSearchParams();
		if (deptId) params.set('department_id', deptId);
		if (employeeId) params.set('employee_id', employeeId);
		if (status) params.set('status', status);
		if (dateFrom) params.set('date_from', dateFrom);
		if (dateTo) params.set('date_to', dateTo);
		const qs = params.toString();
		const url = `${config.API_BASE_URL}/api/attendance/report/export${qs ? '?' + qs : ''}`;
		const response = await fetch(url, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		});
		if (!response.ok) {
			const data = await response.json().catch(() => ({}));
			throw new ApiError(data.message || 'Gagal export laporan absensi', response.status);
		}
		return response.blob();
	},
};

export const salaryComponents = {
	/** @param {string} employeeId @param {number} [page] @param {number} [perPage] */
	list(employeeId, page = 1, perPage = 25) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		return request(`/api/employees/${employeeId}/salary-components?${params}`);
	},

	/** @param {string} employeeId @param {Object} data */
	create(employeeId, data) {
		return request(`/api/employees/${employeeId}/salary-components`, {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} employeeId @param {string} componentId @param {Object} data */
	update(employeeId, componentId, data) {
		return request(`/api/employees/${employeeId}/salary-components/${componentId}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} employeeId @param {string} componentId */
	remove(employeeId, componentId) {
		return request(`/api/employees/${employeeId}/salary-components/${componentId}`, {
			method: 'DELETE',
		});
	},
};

export const organization = {
	/** @returns {Promise<ApiResponse>} */
	getTree() {
		return request('/api/organization/tree');
	},
};

export const shiftChangeRequests = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/shift-change-requests?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/shift-change-requests/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/shift-change-requests', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/shift-change-requests/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/shift-change-requests/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	confirmSwap(id) {
		return request(`/api/shift-change-requests/${id}/confirm-swap`, {
			method: 'PUT',
		});
	},

	/** @param {string} id */
	cancel(id) {
		return request(`/api/shift-change-requests/${id}/cancel`, {
			method: 'PUT',
		});
	},
};

export const scheduleTemplates = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/schedule-templates?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/schedule-templates/${id}`);
	},

	/** @returns {Promise<ApiResponse>} */
	getAll() {
		return request('/api/schedule-templates/all');
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/schedule-templates', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/schedule-templates/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/schedule-templates/${id}`, {
			method: 'DELETE',
		});
	},
};

export const employeeSchedules = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [employeeId] */
	list(page = 1, perPage = 25, employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/employee-schedules?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/employee-schedules/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/employee-schedules', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/employee-schedules/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/employee-schedules/${id}`, {
			method: 'DELETE',
		});
	},

	/**
	 * Resolve effective schedule for an employee on a specific date
	 * @param {string} employeeId
	 * @param {string} date YYYY-MM-DD
	 * @returns {Promise<ApiResponse>}
	 */
	resolve(employeeId, date) {
		const params = new URLSearchParams();
		params.set('employee_id', employeeId);
		params.set('date', date);
		return request(`/api/employee-schedules/resolve?${params}`);
	},
};

export const overtime = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/overtime-requests?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/overtime-requests/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/overtime-requests', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/overtime-requests/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/overtime-requests/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	cancel(id) {
		return request(`/api/overtime-requests/${id}/cancel`, {
			method: 'PUT',
		});
	},

	/** @param {string} id */
	getCalculation(id) {
		return request(`/api/overtime-requests/${id}/calculation`);
	},
};	export const reimbursements = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/reimbursements?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/reimbursements/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/reimbursements', {
			method: 'POST',
			body: data,
		});
	},

	/**
	 * Upload receipt file
	 * @param {File} file
	 * @returns {Promise<ApiResponse>}
	 */
	uploadReceipt(file) {
		const formData = new FormData();
		formData.append('file', file);
		return request('/api/reimbursements/upload', {
			method: 'POST',
			body: formData,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/reimbursements/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/reimbursements/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id @param {Object} [data] */
	pay(id, data = {}) {
		return request(`/api/reimbursements/${id}/pay`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	cancel(id) {
		return request(`/api/reimbursements/${id}/cancel`, {
			method: 'PUT',
		});
	},
};

export const payroll = {
	/** @param {number} [page] @param {number} [perPage] */
	listPeriods(page = 1, perPage = 25) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		return request(`/api/payroll/periods?${params}`);
	},

	/** @param {Object} data */
	createPeriod(data) {
		return request('/api/payroll/periods', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	getPeriod(id) {
		return request(`/api/payroll/periods/${id}`);
	},

	/** @param {string} id @param {Object} [data] */
	calculatePayroll(id, data = {}) {
		return request(`/api/payroll/periods/${id}/calculate`, {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {number} [page] @param {number} [perPage] */
	listItems(id, page = 1, perPage = 50) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		return request(`/api/payroll/periods/${id}/items?${params}`);
	},

	/** @param {string} id */
	approvePeriod(id) {
		return request(`/api/payroll/periods/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id */
	payPeriod(id) {
		return request(`/api/payroll/periods/${id}/pay`, {
			method: 'PUT',
		});
	},

	/** @param {string} periodId @param {string} employeeId */
	getPayslip(periodId, employeeId) {
		return request(`/api/payroll/payslips/${periodId}/${employeeId}`);
	},

	/** @returns {Promise<ApiResponse>} */
	listMyPayslips() {
		return request('/api/payroll/my-payslips');
	},

	/** @param {string} periodId */
	getMyPayslip(periodId) {
		return request(`/api/payroll/my-payslips/${periodId}`);
	},

	/** @param {string} periodId */
	calculateTHR(periodId) {
		return request(`/api/payroll/periods/${periodId}/calculate-thr`);
	},
};

export const company = {
	/** @returns {Promise<ApiResponse>} */
	getSettings() {
		return request('/api/company/settings');
	},

	/** @param {Object} data */
	updateSettings(data) {
		return request('/api/company/settings', {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} employeeId */
	getEmployeeBPJSConfig(employeeId) {
		return request(`/api/employees/${employeeId}/bpjs-config`);
	},

	/** @param {string} employeeId @param {Object} bpjsConfig */
	updateEmployeeBPJSConfig(employeeId, bpjsConfig) {
		return request(`/api/employees/${employeeId}/bpjs-config`, {
			method: 'PUT',
			body: { bpjs_config: bpjsConfig },
		});
	},

	/** @param {string} employeeId */
	getEmployeeTaxConfig(employeeId) {
		return request(`/api/employees/${employeeId}/tax-config`);
	},

	/** @param {string} employeeId @param {Object} taxConfig */
	updateEmployeeTaxConfig(employeeId, taxConfig) {
		return request(`/api/employees/${employeeId}/tax-config`, {
			method: 'PUT',
			body: { tax_config: taxConfig },
		});
	},

	/** @param {string} employeeId */
	getEmployeeOvertimeConfig(employeeId) {
		return request(`/api/employees/${employeeId}/overtime-config`);
	},

	/** @param {string} employeeId @param {Object} overtimeConfig */
	updateEmployeeOvertimeConfig(employeeId, overtimeConfig) {
		return request(`/api/employees/${employeeId}/overtime-config`, {
			method: 'PUT',
			body: { overtime_config: overtimeConfig },
		});
	},
};

export const dashboard = {
	/** @returns {Promise<ApiResponse>} */
	get() {
		return request('/api/dashboard');
	},

	/** @returns {Promise<ApiResponse>} */
	getManager() {
		return request('/api/dashboard/manager');
	},

	/** @returns {Promise<ApiResponse>} */
	getHR() {
		return request('/api/dashboard/hr');
	},
};

export const leaveRequests = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/leaves?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/leaves/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/leaves', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/leaves/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/leaves/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id @param {Object} [data] */
	cancel(id, data = {}) {
		return request(`/api/leaves/${id}/cancel`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @returns {Promise<ApiResponse>} */
	getTypes() {
		return request('/api/leaves/types');
	},

	/** @returns {Promise<ApiResponse>} */
	getMyBalances() {
		return request('/api/leaves/my-balances');
	},

	/** @param {number} [year] */
	getAllBalances(year = 2026) {
		return request(`/api/leaves/balances?year=${year}`);
	},
};

export const documents = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] @param {string} [docType] */
	list(page = 1, perPage = 25, status = '', employeeId = '', docType = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		if (docType) params.set('doc_type', docType);
		return request(`/api/documents?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/documents/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/documents', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	verify(id) {
		return request(`/api/documents/${id}/verify`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/documents/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/documents/${id}`, {
			method: 'DELETE',
		});
	},
};

export const announcements = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [type] */
	list(page = 1, perPage = 25, type = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (type) params.set('type', type);
		return request(`/api/announcements?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/announcements/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/announcements', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/announcements/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/announcements/${id}`, {
			method: 'DELETE',
		});
	},

	/** @param {string} id */
	markRead(id) {
		return request(`/api/announcements/${id}/read`, {
			method: 'POST',
		});
	},
};

export const holidays = {
	/** @param {number} [page] @param {number} [perPage] @param {number} [year] @param {string} [type] */
	list(page = 1, perPage = 25, year = 0, type = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (year) params.set('year', String(year));
		if (type) params.set('type', type);
		return request(`/api/holidays?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/holidays/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/holidays', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/holidays/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/holidays/${id}`, {
			method: 'DELETE',
		});
	},

	/** @param {number} year */
	getByYear(year) {
		return request(`/api/holidays/year/${year}`);
	},
};

export const loans = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/loans?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/loans/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/loans', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/loans/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/loans/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id @param {Object} [data] */
	disburse(id, data = {}) {
		return request(`/api/loans/${id}/disburse`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @returns {Promise<ApiResponse>} */
	getStats() {
		return request('/api/loans/stats');
	},
};

export const kpi = {
	/** @param {number} [page] @param {number} [perPage] @param {number} [year] */
	listTemplates(page = 1, perPage = 25, year = 0) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (year) params.set('year', String(year));
		return request(`/api/kpi/templates?${params}`);
	},

	/** @param {string} id */
	getTemplate(id) {
		return request(`/api/kpi/templates/${id}`);
	},

	/** @param {Object} data */
	createTemplate(data) {
		return request('/api/kpi/templates', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	updateTemplate(id, data) {
		return request(`/api/kpi/templates/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	deleteTemplate(id) {
		return request(`/api/kpi/templates/${id}`, {
			method: 'DELETE',
		});
	},

	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	listReviews(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/kpi/reviews?${params}`);
	},

	/** @param {string} id */
	getReview(id) {
		return request(`/api/kpi/reviews/${id}`);
	},

	/** @param {Object} data */
	createReview(data) {
		return request('/api/kpi/reviews', {
			method: 'POST',
			body: data,
		});
	},
};

export const reprimands = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/reprimands?${params}`);
	},
	/** @param {string} id */
	get(id) { return request(`/api/reprimands/${id}`); },
	/** @param {Object} data */
	create(data) { return request('/api/reprimands', { method: 'POST', body: data }); },
	/** @param {string} id @param {Object} [data] */
	acknowledge(id, data = {}) { return request(`/api/reprimands/${id}/acknowledge`, { method: 'PUT', body: data }); },
};

export const dailyJournals = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [deptId] @param {string} [employeeId] @param {string} [dateFrom] @param {string} [dateTo] */
	list(page = 1, perPage = 25, deptId = '', employeeId = '', dateFrom = '', dateTo = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (deptId) params.set('department_id', deptId);
		if (employeeId) params.set('employee_id', employeeId);
		if (dateFrom) params.set('date_from', dateFrom);
		if (dateTo) params.set('date_to', dateTo);
		return request(`/api/daily-journals?${params}`);
	},
	/** @param {string} id */
	get(id) { return request(`/api/daily-journals/${id}`); },
	/** @param {Object} data */
	create(data) { return request('/api/daily-journals', { method: 'POST', body: data }); },
	/** @param {string} id */
	acknowledge(id) { return request(`/api/daily-journals/${id}/acknowledge`, { method: 'PUT' }); },
};

export const notifications = {
	/** @param {number} [page] @param {number} [perPage] */
	list(page = 1, perPage = 25) {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		return request(`/api/notifications?${params}`);
	},
	/** @returns {Promise<ApiResponse>} */
	getUnreadCount() {
		return request('/api/notifications/unread-count');
	},
	/** @param {string[]} [ids] */
	markAsRead(ids = []) {
		return request('/api/notifications/mark-read', {
			method: 'PUT',
			body: { ids },
		});
	},
};

export const activityLogs = {
	/** @param {Object} [filters] @param {number} [filters.page] @param {number} [filters.perPage] @param {string} [filters.action] @param {string} [filters.entityType] @param {string} [filters.userId] @param {string} [filters.startDate] @param {string} [filters.endDate] */
	list(filters = {}) {
		const params = new URLSearchParams();
		params.set('page', String(filters.page || 1));
		params.set('per_page', String(filters.perPage || 25));
		if (filters.action) params.set('action', filters.action);
		if (filters.entityType) params.set('entity_type', filters.entityType);
		if (filters.userId) params.set('user_id', filters.userId);
		if (filters.startDate) params.set('start_date', filters.startDate);
		if (filters.endDate) params.set('end_date', filters.endDate);
		return request(`/api/activity-logs?${params}`);
	},
	/** @param {string} id */
	get(id) { return request(`/api/activity-logs/${id}`); },
	/** @returns {Promise<ApiResponse>} */
	getEntityTypes() { return request('/api/activity-logs/entity-types'); },
	/** @returns {Promise<ApiResponse>} */
	getActions() { return request('/api/activity-logs/actions'); },
};

export const reports = {
	/** @param {number} [year] */
	headcount(year = 0) { return request(`/api/reports/headcount?year=${year}`); },
	/** @param {number} [year] */
	payrollSummary(year = 0) { return request(`/api/reports/payroll-summary?year=${year}`); },
	/** @param {number} [year] @param {number} [month] @param {string} [deptId] */
	attendanceSummary(year = 0, month = 0, deptId = '') {
		const params = new URLSearchParams();
		if (year) params.set('year', String(year));
		if (month) params.set('month', String(month));
		if (deptId) params.set('department_id', deptId);
		return request(`/api/reports/attendance-summary?${params}`);
	},
	/** @param {number} [year] @param {string} [deptId] */
	leaveSummary(year = 0, deptId = '') {
		const params = new URLSearchParams();
		if (year) params.set('year', String(year));
		if (deptId) params.set('department_id', deptId);
		return request(`/api/reports/leave-summary?${params}`);
	},
	/** @param {number} [year] @param {number} [month] */
	overtimeSummary(year = 0, month = 0) {
		return request(`/api/reports/overtime-summary?year=${year}&month=${month}`);
	},
};

export const approvalWorkflows = {
	/** @param {string} [entityType] */
	list(entityType = '') {
		const params = new URLSearchParams();
		if (entityType) params.set('entity_type', entityType);
		return request(`/api/approval-workflows?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/approval-workflows/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/approval-workflows', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/approval-workflows/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/approval-workflows/${id}`, {
			method: 'DELETE',
		});
	},

	/** @param {string} workflowId @param {Object} data */
	addStep(workflowId, data) {
		return request(`/api/approval-workflows/${workflowId}/steps`, {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} stepId @param {Object} data */
	updateStep(stepId, data) {
		return request(`/api/approval-workflow-steps/${stepId}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} stepId */
	removeStep(stepId) {
		return request(`/api/approval-workflow-steps/${stepId}`, {
			method: 'DELETE',
		});
	},
};

export const approvals = {
	/** @returns {Promise<ApiResponse>} */
	getPending() {
		return request('/api/approvals/pending');
	},

	/** @param {string} entityType @param {string} entityId @param {Object} data */
	process(entityType, entityId, data) {
		return request(`/api/approvals/${entityType}/${entityId}/process`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} entityType @param {string} entityId */
	initTracking(entityType, entityId) {
		return request(`/api/approvals/${entityType}/${entityId}/init`, {
			method: 'POST',
		});
	},
};

export const manualAttendance = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/manual-attendance?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/manual-attendance/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/manual-attendance', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/manual-attendance/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/manual-attendance/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	cancel(id) {
		return request(`/api/manual-attendance/${id}/cancel`, {
			method: 'PUT',
		});
	},
};

export const resign = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/resign?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/resign/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/resign', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	approve(id) {
		return request(`/api/resign/${id}/approve`, {
			method: 'PUT',
		});
	},

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) {
		return request(`/api/resign/${id}/reject`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	listClearance(id) {
		return request(`/api/resign/${id}/clearance`);
	},

	/** @param {string} itemId @param {Object} data */
	updateClearance(itemId, data) {
		return request(`/api/resign/clearance/${itemId}`, {
			method: 'PUT',
			body: data,
		});
	},
};

export const push = {
	/**
	 * Get VAPID public key for push subscription
	 * @returns {Promise<ApiResponse>}
	 */
	getVapidPublicKey() {
		return request('/api/push/vapid-public-key');
	},

	/**
	 * Subscribe to push notifications
	 * @param {PushSubscription} subscription
	 * @param {string} [deviceName]
	 * @returns {Promise<ApiResponse>}
	 */
	subscribe(subscription, deviceName = '') {
		return request('/api/push/subscribe', {
			method: 'POST',
			body: {
				endpoint: subscription.endpoint,
				p256dh_key: arrayBufferToBase64(subscription.getKey('p256dh')),
				auth_key: arrayBufferToBase64(subscription.getKey('auth')),
				device_name: deviceName,
			},
		});
	},

	/**
	 * Unsubscribe from push notifications
	 * @param {string} id
	 * @returns {Promise<ApiResponse>}
	 */
	unsubscribe(id) {
		return request(`/api/push/subscribe/${id}`, {
			method: 'DELETE',
		});
	},

	/**
	 * List active push subscriptions
	 * @returns {Promise<ApiResponse>}
	 */
	listSubscriptions() {
		return request('/api/push/subscriptions');
	},
};

/** 
 * Helper: Convert ArrayBuffer to Base64 string 
 * @param {ArrayBuffer | null} buffer
 */
function arrayBufferToBase64(buffer) {
	if (!buffer) return '';
	const bytes = new Uint8Array(buffer);
	let binary = '';
	for (let i = 0; i < bytes.byteLength; i++) {
		binary += String.fromCharCode(bytes[i]);
	}
	return btoa(binary);
}	export const employeesApi = {
	/**
	 * Register face descriptor for an employee
	 * @param {string} id Employee UUID
	 * @param {string} descriptor JSON string of face descriptor array
	 * @returns {Promise<ApiResponse>}
	 */
	registerFaceDescriptor(id, descriptor) {
		return request(`/api/employees/${id}/face-descriptor`, {
			method: 'POST',
			body: { descriptor },
		});
	},
};

	export const mutations = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [status] @param {string} [employeeId] */
	list(page = 1, perPage = 25, status = '', employeeId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		return request(`/api/mutations?${params}`);
	},

	/** @param {string} id */
	get(id) { return request(`/api/mutations/${id}`); },

	/** @param {Object} data */
	create(data) { return request('/api/mutations', { method: 'POST', body: data }); },

	/** @param {string} id */
	approve(id) { return request(`/api/mutations/${id}/approve`, { method: 'PUT' }); },

	/** @param {string} id @param {Object} [data] */
	reject(id, data = {}) { return request(`/api/mutations/${id}/reject`, { method: 'PUT', body: data }); },

	/** @param {string} id */
	cancel(id) { return request(`/api/mutations/${id}/cancel`, { method: 'PUT' }); },

	/**
	 * Export mutations as Excel file
	 * @param {string} [status]
	 * @param {string} [employeeId]
	 * @returns {Promise<Blob>}
	 */
	async exportExcel(status = '', employeeId = '') {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem(config.ACCESS_TOKEN_KEY) : null;
		const params = new URLSearchParams();
		if (status) params.set('status', status);
		if (employeeId) params.set('employee_id', employeeId);
		const qs = params.toString();
		const url = `${config.API_BASE_URL}/api/mutations/export${qs ? '?' + qs : ''}`;
		const response = await fetch(url, {
			headers: token ? { Authorization: `Bearer ${token}` } : {},
		});
		if (!response.ok) {
			const data = await response.json().catch(() => ({}));
			throw new ApiError(data.message || 'Gagal export data mutasi', response.status);
		}
		return response.blob();
	},
};

export const shifts = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [search] @param {string} [departmentId] */
	list(page = 1, perPage = 25, search = '', departmentId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		if (departmentId) params.set('department_id', departmentId);
		return request(`/api/shifts?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/shifts/${id}`);
	},

	/** @param {string} [departmentId] @returns {Promise<ApiResponse>} */
	getAll(departmentId = '') {
		const params = new URLSearchParams();
		if (departmentId) params.set('department_id', departmentId);
		return request(`/api/shifts/all?${params}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/shifts', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/shifts/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/shifts/${id}`, {
			method: 'DELETE',
		});
	},
};

export const rosters = {
	/** @param {number} [page] @param {number} [perPage] @param {string} [departmentId] */
	list(page = 1, perPage = 25, departmentId = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (departmentId) params.set('department_id', departmentId);
		return request(`/api/rosters?${params}`);
	},

	/** @param {string} id */
	get(id) {
		return request(`/api/rosters/${id}`);
	},

	/** @param {Object} data */
	create(data) {
		return request('/api/rosters', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id @param {Object} data */
	update(id, data) {
		return request(`/api/rosters/${id}`, {
			method: 'PUT',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/rosters/${id}`, {
			method: 'DELETE',
		});
	},

	/** @param {string} rosterId */
	getEntries(rosterId) {
		return request(`/api/rosters/${rosterId}/entries`);
	},

	/** @param {string} rosterId */
	getCalendar(rosterId) {
		return request(`/api/rosters/${rosterId}/calendar`);
	},

	/** @param {string} rosterId @param {Object} data */
	bulkCreateEntries(rosterId, data) {
		return request(`/api/rosters/${rosterId}/entries/bulk`, {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} rosterId @param {string} employeeId */
	removeEmployee(rosterId, employeeId) {
		return request(`/api/rosters/${rosterId}/employees/${employeeId}`, {
			method: 'DELETE',
		});
	},
};

export const rosterEntries = {
	/** @param {Object} data */
	create(data) {
		return request('/api/roster-entries', {
			method: 'POST',
			body: data,
		});
	},

	/** @param {string} id */
	remove(id) {
		return request(`/api/roster-entries/${id}`, {
			method: 'DELETE',
		});
	},
};

export { ApiError };
export default { auth, employees, dashboard, shiftChangeRequests, overtime, reimbursements, attendance, leaveRequests, documents, announcements, holidays, loans, kpi, reprimands, dailyJournals, notifications, activityLogs, reports, company, approvalWorkflows, approvals, manualAttendance, resign, push, mutations, employeesApi, shifts, rosters, rosterEntries, ApiError };
