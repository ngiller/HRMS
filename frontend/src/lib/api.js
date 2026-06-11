/**
 * HRMS API Client
 *
 * Wrapper untuk fetch API dengan JWT authentication,
 * error handling, dan token management.
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
 * Base API request function
 * @param {string} endpoint - API endpoint path
 * @param {RequestOptions} [options] - Request options
 * @returns {Promise<ApiResponse>}
 */
async function request(endpoint, options = {}) {
	const url = `${config.API_BASE_URL}${endpoint}`;
	const { body, method = 'GET', headers = {}, auth = true } = options;

	/** @type {Record<string, any>} */
	const requestHeaders = {
		'Content-Type': 'application/json',
		...headers,
	};

	// Attach auth token
	if (auth) {
		const token = typeof localStorage !== 'undefined' ? localStorage.getItem(config.ACCESS_TOKEN_KEY) : null;
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
		fetchOptions.body = JSON.stringify(body);
	}

	try {
		const response = await fetch(url, fetchOptions);
		const data = await response.json();

		if (!response.ok) {
			throw new ApiError(
				data.error || 'Terjadi kesalahan',
				response.status,
				data
			);
		}

		return data;
	} catch (error) {
		if (error instanceof ApiError) {
			throw error;
		}
		throw new ApiError('Gagal terhubung ke server', 0);
	}
}

/**
 * Custom error class for API errors
 */
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

/**
 * Auth API
 */
export const auth = {
	/**
	 * @param {string} email
	 * @param {string} password
	 * @returns {Promise<ApiResponse>}
	 */
	login(email, password) {
		return request('/api/auth/login', {
			method: 'POST',
			body: { email, password },
			auth: false,
		});
	},

	/**
	 * @param {string} email
	 * @returns {Promise<ApiResponse>}
	 */
	forgotPassword(email) {
		return request('/api/auth/forgot-password', {
			method: 'POST',
			body: { email },
			auth: false,
		});
	},

	/**
	 * @param {string} token
	 * @param {string} newPassword
	 * @returns {Promise<ApiResponse>}
	 */
	resetPassword(token, newPassword) {
		return request('/api/auth/reset-password', {
			method: 'POST',
			body: { token, new_password: newPassword },
			auth: false,
		});
	},

	/**
	 * @returns {Promise<ApiResponse>}
	 */
	getMe() {
		return request('/api/auth/me');
	},

	/**
	 * Simpan token & user data setelah login
	 * @param {ApiResponse} response
	 */
	saveSession(response) {
		if (!response || !response.data) return;
		const { access_token, refresh_token, user } = response.data;
		if (access_token) localStorage.setItem(config.ACCESS_TOKEN_KEY, access_token);
		if (refresh_token) localStorage.setItem(config.REFRESH_TOKEN_KEY, refresh_token);
		if (user) localStorage.setItem(config.USER_DATA_KEY, JSON.stringify(user));
	},

	/** Hapus session (logout) */
	clearSession() {
		localStorage.removeItem(config.ACCESS_TOKEN_KEY);
		localStorage.removeItem(config.REFRESH_TOKEN_KEY);
		localStorage.removeItem(config.USER_DATA_KEY);
	},

	/**
	 * Ambil user data dari local storage
	 * @returns {Object|null}
	 */
	getUser() {
		try {
			const data = typeof localStorage !== 'undefined' ? localStorage.getItem(config.USER_DATA_KEY) : null;
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

/**
 * Employee API
 */
export const employees = {
	/**
	 * @param {number} [page]
	 * @param {number} [perPage]
	 * @param {string} [search]
	 * @returns {Promise<ApiResponse>}
	 */
	list(page = 1, perPage = 25, search = '') {
		const params = new URLSearchParams();
		params.set('page', String(page));
		params.set('per_page', String(perPage));
		if (search) params.set('search', search);
		return request(`/api/employees?${params}`);
	},

	/**
	 * @param {string} id
	 * @returns {Promise<ApiResponse>}
	 */
	get(id) {
		return request(`/api/employees/${id}`);
	},
};

/**
 * Dashboard API
 */
export const dashboard = {
	/** @returns {Promise<ApiResponse>} */
	get() {
		return request('/api/dashboard');
	},
};

export { ApiError };
export default { auth, employees, dashboard, ApiError };
