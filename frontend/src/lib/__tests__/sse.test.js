// @ts-nocheck
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// Mock EventSource before importing the module
class MockEventSource {
	close = vi.fn(() => {
		this.readyState = 2; // CLOSED
	});

	constructor(url) {
		this.url = url;
		this.readyState = 0; // CONNECTING
		this.listeners = {};

		// Simulate async connection
		setTimeout(() => {
			this.readyState = 1; // OPEN
			if (this.listeners['connected']) {
				this.listeners['connected'].forEach((cb) => cb({}));
			}
		}, 10);
	}

	addEventListener(event, callback) {
		if (!this.listeners[event]) {
			this.listeners[event] = [];
		}
		this.listeners[event].push(callback);
	}

	removeEventListener(event, callback) {
		if (this.listeners[event]) {
			this.listeners[event] = this.listeners[event].filter((cb) => cb !== callback);
		}
	}

	// Helper to simulate events
	simulateEvent(event, data) {
		if (this.listeners[event]) {
			this.listeners[event].forEach((cb) =>
				cb({ data: typeof data === 'string' ? data : JSON.stringify(data) })
			);
		}
	}

	// Helper to simulate errors
	simulateError() {
		if (this.onerror) {
			this.onerror(new Event('error'));
		}
	}
}

let mockEventSourceInstance = null;

// @ts-ignore
globalThis.EventSource = vi.fn(function(url) {
	mockEventSourceInstance = new MockEventSource(url);
	return mockEventSourceInstance;
});

// Mock $lib/config
vi.mock('$lib/config', () => ({
	default: {
		API_BASE_URL: 'http://localhost:8900',
	},
}));

// Now import after mocks are set up
import { connectSSE, disconnectSSE, isSSEConnected } from '../sse.js';

describe('SSE Client', () => {
	beforeEach(() => {
		mockEventSourceInstance = null;
		vi.clearAllTimers();
		vi.useFakeTimers();
	});

	afterEach(() => {
		vi.useRealTimers();
		disconnectSSE();
		// Clear the mock EventSource instances
		vi.clearAllMocks();
	});

	it('should return null when no token provided', () => {
		const result = connectSSE('');
		expect(result).toBeNull();
	});

	it('should return null when token is undefined', () => {
		const result = connectSSE(undefined);
		expect(result).toBeNull();
	});

	it('should create EventSource with correct URL', () => {
		connectSSE('test-token-123');

		expect(EventSource).toHaveBeenCalledTimes(1);
		const url = EventSource.mock.calls[0][0];
		expect(url).toBe(
			'http://localhost:8900/api/sse/subscribe?token=test-token-123'
		);
	});

	it('should register connected event listener', () => {
		const onConnected = vi.fn();
		connectSSE('token', { onConnected });

		expect(mockEventSourceInstance).not.toBeNull();

		// Advance timers to trigger the connected event
		vi.advanceTimersByTime(20);
		expect(onConnected).toHaveBeenCalledTimes(1);
	});

	it('should handle approval_update events and dispatch custom event', () => {
		const onEvent = vi.fn();
		connectSSE('token', { onEvent });

		expect(mockEventSourceInstance).not.toBeNull();

		const testData = { type: 'approval_update', data: { count: 5, action: 'approved' } };
		mockEventSourceInstance.simulateEvent('approval_update', testData);

		expect(onEvent).toHaveBeenCalledTimes(1);
		expect(onEvent).toHaveBeenCalledWith(testData);
	});

	it('should dispatch custom DOM event on approval_update', () => {
		// Listen for the custom event
		const customEventHandler = vi.fn();
		window.addEventListener('sse:approval_update', customEventHandler);

		connectSSE('token');

		const testData = { type: 'approval_update', data: { count: 3 } };
		mockEventSourceInstance.simulateEvent('approval_update', testData);

		expect(customEventHandler).toHaveBeenCalledTimes(1);
		const event = customEventHandler.mock.calls[0][0];
		expect(event.detail).toEqual(testData);

		window.removeEventListener('sse:approval_update', customEventHandler);
	});

	it('should handle malformed JSON in approval_update gracefully', () => {
		const onEvent = vi.fn();
		connectSSE('token', { onEvent });

		// Simulate malformed JSON
		const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});

		// Manually trigger the event with invalid JSON
		if (mockEventSourceInstance && mockEventSourceInstance.listeners['approval_update']) {
			mockEventSourceInstance.listeners['approval_update'].forEach((cb) =>
				cb({ data: '{invalid json}' })
			);
		}

		// Should not crash, should log error
		expect(consoleSpy).toHaveBeenCalled();
		expect(onEvent).not.toHaveBeenCalled();

		consoleSpy.mockRestore();
	});

	it('should disconnect and clean up EventSource', () => {
		connectSSE('token');

		expect(mockEventSourceInstance).not.toBeNull();

		disconnectSSE();

		expect(mockEventSourceInstance.close).toHaveBeenCalledTimes(1);
		expect(isSSEConnected()).toBe(false);
	});

	it('should reconnect on error', () => {
		vi.useFakeTimers();
		const onDisconnected = vi.fn();
		const onConnected = vi.fn();

		connectSSE('token', {
			reconnectDelay: 5000,
			onDisconnected,
			onConnected,
		});

		// Advance timers to trigger initial connected event
		vi.advanceTimersByTime(20);
		expect(onConnected).toHaveBeenCalledTimes(1);

		// Simulate error during SSE
		if (mockEventSourceInstance) {
			mockEventSourceInstance.simulateError();
		}

		// Should have called onDisconnected
		expect(onDisconnected).toHaveBeenCalledTimes(1);

		// EventSource.close() should have been called (via the mock)
		expect(mockEventSourceInstance.close).toHaveBeenCalledTimes(1);

		// Advance timer for reconnect (5000ms)
		vi.advanceTimersByTime(5000);

		// Should create a new EventSource for reconnection
		expect(EventSource).toHaveBeenCalledTimes(2);

		// Advance to trigger connected event on second connection
		vi.advanceTimersByTime(20);

		// onConnected should be called again
		expect(onConnected).toHaveBeenCalledTimes(2);
	});

	it('should not reconnect when token becomes empty', () => {
		vi.useFakeTimers();

		// Connect first
		connectSSE('token');

		// Simulate error
		if (mockEventSourceInstance) {
			mockEventSourceInstance.simulateError();
		}

		// Now disconnect (this clears reconnect timer)
		disconnectSSE();

		// Advance past reconnect delay
		vi.advanceTimersByTime(5000);

		// Should NOT have created another EventSource after disconnect
		expect(EventSource).toHaveBeenCalledTimes(1);
	});

	it('should track connection status', () => {
		expect(isSSEConnected()).toBe(false);

		connectSSE('token');

		// Should still be false initially (not yet connected)
		expect(isSSEConnected()).toBe(false);

		// Advance timer to trigger connection
		vi.advanceTimersByTime(20);
		expect(isSSEConnected()).toBe(true);

		// Simulate error
		if (mockEventSourceInstance) {
			mockEventSourceInstance.simulateError();
		}

		expect(isSSEConnected()).toBe(false);

		disconnectSSE();
		expect(isSSEConnected()).toBe(false);
	});

	it('should close existing connection before creating new one', () => {
		connectSSE('token');
		const firstInstance = mockEventSourceInstance;

		// Connect again with same token
		connectSSE('token');

		// First EventSource should have been closed
		expect(firstInstance.close).toHaveBeenCalledTimes(1);

		// New EventSource should be created
		expect(EventSource).toHaveBeenCalledTimes(2);
	});
});
