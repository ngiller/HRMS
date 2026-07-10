/**
 * SSE Client for real-time server-sent events.
 * Connects to the backend SSE endpoint and dispatches custom DOM events.
 */

import config from '$lib/config';

/** @type {EventSource | null} */
let eventSource = null;
/** @type {ReturnType<typeof setTimeout> | null} */
let reconnectTimer = null;
let isConnected = false;

/**
 * Connect to the SSE endpoint.
 * @param {string} token - JWT access token
 * @param {{
 *   reconnectDelay?: number,
 *   onEvent?: (data: any) => void,
 *   onConnected?: () => void,
 *   onDisconnected?: () => void
 * }} [options]
 * @returns {EventSource|null}
 */
export function connectSSE(token, options = {}) {
	const { reconnectDelay = 5000, onEvent, onConnected, onDisconnected } = options;

	// Close existing connection
	disconnectSSE();

	if (!token) return null;

	const url = `${config.API_BASE_URL}/api/sse/subscribe?token=${encodeURIComponent(token)}`;

	eventSource = new EventSource(url);

	eventSource.addEventListener('connected', () => {
		isConnected = true;
		console.log('[SSE] Connected');
		if (onConnected) onConnected();
	});

	eventSource.addEventListener('approval_update', (event) => {
		try {
			const data = JSON.parse(event.data);
			console.log('[SSE] Approval update:', data);

			// Dispatch a custom DOM event for components to listen to
			window.dispatchEvent(
				new CustomEvent('sse:approval_update', { detail: data })
			);

			if (onEvent) onEvent(data);
		} catch (e) {
			console.error('[SSE] Failed to parse event:', e);
		}
	});

	eventSource.onerror = () => {
		isConnected = false;
		console.warn('[SSE] Connection error, will reconnect...');
		if (onDisconnected) onDisconnected();

		// Close the broken connection
		if (eventSource) {
			eventSource.close();
			eventSource = null;
		}

		// Auto-reconnect
		if (reconnectTimer) clearTimeout(reconnectTimer);
		reconnectTimer = setTimeout(() => {
			console.log('[SSE] Reconnecting...');
			connectSSE(token, options);
		}, reconnectDelay);
	};

	return eventSource;
}

/**
 * Disconnect from the SSE endpoint.
 */
export function disconnectSSE() {
	if (reconnectTimer) {
		clearTimeout(reconnectTimer);
		reconnectTimer = null;
	}
	if (eventSource) {
		eventSource.close();
		eventSource = null;
	}
	isConnected = false;
}

/**
 * Check if currently connected.
 */
export function isSSEConnected() {
	return isConnected;
}
