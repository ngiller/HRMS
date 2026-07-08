<script lang="ts">
	import { onMount } from 'svelte';
	import { push, ApiError } from '$lib/api.js';

	let supported = $state(true);
	let subscribed = $state(false);
	let loading = $state(true);
	let errorMsg = $state('');
	let subs: any[] = $state([]);

	onMount(async () => {
		// Check if Push API is supported
		if (!('Notification' in window) || !('serviceWorker' in navigator) || !('PushManager' in window)) {
			supported = false;
			loading = false;
			return;
		}

		// Check permission
		if (Notification.permission === 'denied') {
			supported = true; // Still possible but blocked
			loading = false;
			return;
		}

		await checkSubscription();
	});

	async function checkSubscription() {
		loading = true;
		try {
			const reg = await navigator.serviceWorker.ready;
			const existingSub = await reg.pushManager.getSubscription();
			subscribed = !!existingSub;

			if (existingSub) {
				// Load existing subscriptions from server
				const res = await push.listSubscriptions();
				if (res.success) {
					subs = res.data || [];
				}
			}
		} catch (e) {
			console.warn('Push subscription check failed:', e);
		} finally {
			loading = false;
		}
	}

	async function subscribe() {
		errorMsg = '';
		try {
			// Request permission first
			const permission = await Notification.requestPermission();
			if (permission !== 'granted') {
				errorMsg = 'Izin notifikasi ditolak. Aktifkan dari pengaturan browser.';
				return;
			}

			// Get VAPID public key from server
			const keyRes = await push.getVapidPublicKey();
			if (!keyRes.success) {
				errorMsg = 'Gagal mendapatkan kunci VAPID';
				return;
			}
			const vapidPublicKey = keyRes.data.public_key;

			// Convert VAPID key to Uint8Array
			const convertedKey = urlBase64ToUint8Array(vapidPublicKey);

			// Register push subscription
			const reg = await navigator.serviceWorker.ready;
			const existingSub = await reg.pushManager.getSubscription();
			if (existingSub) {
				await existingSub.unsubscribe();
			}

			const newSub = await reg.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey: convertedKey,
			});

			// Send subscription to server
			const deviceName = getDeviceName();
			await push.subscribe(newSub, deviceName);

			subscribed = true;
			await checkSubscription();
		} catch (e) {
			if (e instanceof ApiError) {
				errorMsg = e.message;
			} else {
				errorMsg = 'Gagal berlangganan notifikasi: ' + (e as Error).message;
			}
		}
	}

	async function unsubscribe(subId: string) {
		errorMsg = '';
		try {
			await push.unsubscribe(subId);

			// If no more subscriptions, unsubscribe from push manager too
			const reg = await navigator.serviceWorker.ready;
			const existingSub = await reg.pushManager.getSubscription();
			if (existingSub && subs.length <= 1) {
				await existingSub.unsubscribe();
			}

			await checkSubscription();
		} catch (e) {
			if (e instanceof ApiError) {
				errorMsg = e.message;
			} else {
				errorMsg = 'Gagal berhenti berlangganan';
			}
		}
	}

	async function unsubscribeAll() {
		errorMsg = '';
		try {
			// Unsubscribe all from server
			for (const sub of subs) {
				await push.unsubscribe(sub.id);
			}

			// Unsubscribe from push manager
			const reg = await navigator.serviceWorker.ready;
			const existingSub = await reg.pushManager.getSubscription();
			if (existingSub) {
				await existingSub.unsubscribe();
			}

			subscribed = false;
			subs = [];
		} catch (e) {
			if (e instanceof ApiError) {
				errorMsg = e.message;
			} else {
				errorMsg = 'Gagal berhenti berlangganan';
			}
		}
	}

	function getDeviceName(): string {
		const ua = navigator.userAgent;
		if (ua.includes('Windows')) return 'Windows';
		if (ua.includes('Mac OS')) return 'macOS';
		if (ua.includes('Linux') && !ua.includes('Android')) return 'Linux';
		if (ua.includes('Android')) return 'Android';
		if (ua.includes('iPhone') || ua.includes('iPad')) return 'iOS';
		return 'Unknown';
	}

	/** Convert a base64url string to Uint8Array */
	function urlBase64ToUint8Array(base64String: string): Uint8Array {
		const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
		const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
		const rawData = window.atob(base64);
		const outputArray = new Uint8Array(rawData.length);
		for (let i = 0; i < rawData.length; ++i) {
			outputArray[i] = rawData.charCodeAt(i);
		}
		return outputArray;
	}
</script>

{#if !supported}
	<div class="p-4 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-xl text-sm text-yellow-700 dark:text-yellow-400">
		Browser Anda tidak mendukung notifikasi push. Gunakan browser modern seperti Chrome, Firefox, atau Edge.
	</div>
{:else}
	<div class="space-y-4">
		{#if loading}
			<div class="flex items-center gap-3 text-sm text-gray-500 dark:text-gray-400">
				<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
				</svg>
				Memeriksa status notifikasi...
			</div>
		{:else if subscribed && subs.length > 0}
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-2">
					<span class="w-2 h-2 bg-green-500 rounded-full animate-pulse" />
					<span class="text-sm font-medium text-green-700 dark:text-green-400">
						Notifikasi aktif di {subs.length} perangkat
					</span>
				</div>
				<button
					onclick={unsubscribeAll}
					class="text-xs text-red-600 dark:text-red-400 hover:text-red-800 dark:hover:text-red-300 font-medium transition-colors"
				>
					Nonaktifkan Semua
				</button>
			</div>
			<div class="space-y-2">
				{#each subs as sub}
					<div class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-gray-100 dark:border-gray-800">
						<div class="flex items-center gap-2">
							<svg class="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 18h.01M8 21h8a2 2 0 002-2V5a2 2 0 00-2-2H8a2 2 0 00-2 2v14a2 2 0 002 2z" />
							</svg>
							<span class="text-sm text-gray-700 dark:text-gray-300">
								{sub.device_name || 'Perangkat'}
							</span>
						</div>
						<button
							onclick={() => unsubscribe(sub.id)}
							class="text-xs text-gray-500 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors"
						>
							Lepaskan
						</button>
					</div>
				{/each}
			</div>
		{:else if subscribed && subs.length === 0}
			<div class="flex items-center gap-2">
				<span class="w-2 h-2 bg-yellow-500 rounded-full" />
				<span class="text-sm text-yellow-700 dark:text-yellow-400">
					Terdaftar namun belum sinkron dengan server
				</span>
			</div>
			<button
				onclick={subscribe}
				class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
			>
				Daftarkan Ulang
			</button>
		{:else}
			<div class="flex items-center justify-between">
				<div>
					<p class="text-sm font-medium text-gray-900 dark:text-white">Notifikasi Push</p>
					<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
						Dapatkan notifikasi real-time langsung di browser
					</p>
				</div>
				<button
					onclick={subscribe}
					class="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium shadow-sm"
				>
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
					</svg>
					Aktifkan Notifikasi
				</button>
			</div>
		{/if}

		{#if errorMsg}
			<div class="p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-sm text-red-700 dark:text-red-400">
				{errorMsg}
			</div>
		{/if}

		{#if Notification.permission === 'denied'}
			<div class="p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg text-sm text-yellow-700 dark:text-yellow-400">
				Izin notifikasi telah ditolak. Aktifkan melalui pengaturan browser Anda.
			</div>
		{/if}
	</div>
{/if}
