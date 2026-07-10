/**
 * HRMS Frontend Configuration
 * 
 * Ubah API_BASE_URL sesuai dengan alamat backend server.
 * Untuk development: http://localhost:8900
 * Untuk production: https://api.hrms.company.com
 */

const config = {
	// Backend API URL (dynamic via VITE_API_BASE_URL env var)
	// - Local dev (make dev-backend): http://localhost:8900 (default)
	// - Docker dev (docker compose): http://localhost:8080 — set VITE_API_BASE_URL=http://localhost:8080
	API_BASE_URL: import.meta.env.DEV 
		? (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8900') 
		: '',

	// JWT Token keys (localStorage)
	ACCESS_TOKEN_KEY: 'hrms_access_token',
	REFRESH_TOKEN_KEY: 'hrms_refresh_token',
	USER_DATA_KEY: 'hrms_user',

	// App info
	APP_NAME: 'HRMS',
	COMPANY_NAME: 'PT Maju Jaya',
	APP_VERSION: '1.0.0',

	// Pagination defaults
	DEFAULT_PAGE_SIZE: 25,
};

export default config;
