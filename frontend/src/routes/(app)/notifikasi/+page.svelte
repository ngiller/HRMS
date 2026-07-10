<script lang="ts">
	import { onMount } from 'svelte';
	import { notifications, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import PushSubscriptionManager from '$lib/components/PushSubscriptionManager.svelte';

	let loading = $state(true);
	let items = $state<any[]>([]);
	let total = $state(0);
	let pageNum = $state(1);
	const perPage = 25;
	let unreadCount = $state(0);
	let expandedId = $state<string | null>(null);
	let error = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			console.log("Loading notifications...");
			const res = await notifications.list(pageNum, perPage);
			console.log("Notifications response:", res);
			if (res.success) {
				items = res.data.notifications || [];
				total = res.data.total || 0;
				unreadCount = res.data.unread_count || 0;
			}
		} catch (e) {
			console.error("Notifications error:", e);
			if (e instanceof ApiError) error = e.message;
			else error = 'Gagal memuat notifikasi';
		} finally {
			loading = false;
		}
	}

	async function markAsRead(id?: string) {
		try {
			await notifications.markAsRead(id ? [id] : []);
			// Update locally for instant feedback instead of full reload
			if (id) {
				const item = items.find(i => i.id === id);
				if (item && !item.is_read) {
					item.is_read = true;
					unreadCount = Math.max(0, unreadCount - 1);
				}
			} else {
				items.forEach(i => i.is_read = true);
				unreadCount = 0;
			}
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
		}
	}

	function formatDate(dateStr: string) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('id-ID', {
			day: 'numeric', month: 'long', year: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	// Returns an SVG string based on notification type
	function getTypeIcon(type: string): { svg: string, colorClass: string, bgClass: string } {
		if (type.includes('approved') || type === 'reimbursement_approved') {
			return {
				svg: '<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
				colorClass: 'text-emerald-600 dark:text-emerald-400',
				bgClass: 'bg-emerald-50 dark:bg-emerald-500/10'
			};
		}
		if (type.includes('rejected')) {
			return {
				svg: '<path stroke-linecap="round" stroke-linejoin="round" d="M9.75 9.75l4.5 4.5m0-4.5l-4.5 4.5M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
				colorClass: 'text-rose-600 dark:text-rose-400',
				bgClass: 'bg-rose-50 dark:bg-rose-500/10'
			};
		}
		if (type === 'reprimand_issued') {
			return {
				svg: '<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2.25m0 2.25h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />',
				colorClass: 'text-amber-600 dark:text-amber-400',
				bgClass: 'bg-amber-50 dark:bg-amber-500/10'
			};
		}
		// Default (Info)
		return {
			svg: '<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />',
			colorClass: 'text-blue-600 dark:text-blue-400',
			bgClass: 'bg-blue-50 dark:bg-blue-500/10'
		};
	}

	const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	onMount(() => { load(); });
</script>

<div class="w-full space-y-6">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight flex items-center gap-2">
				Notifikasi
				{#if unreadCount > 0}
					<span class="inline-flex items-center justify-center w-6 h-6 text-xs font-bold text-white bg-red-500 rounded-full">{unreadCount}</span>
				{/if}
			</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
				Pembaruan terbaru terkait aktivitas Anda
			</p>
		</div>
		{#if unreadCount > 0}
			<button
				onclick={() => markAsRead()}
				class="inline-flex items-center justify-center gap-2 px-4 py-2.5 bg-blue-50 text-blue-600 dark:bg-blue-900/30 dark:text-blue-400 rounded-xl hover:bg-blue-100 dark:hover:bg-blue-900/50 transition-colors text-sm font-semibold whitespace-nowrap"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" />
				</svg>
				Tandai Semua Dibaca
			</button>
		{/if}
	</div>

	{#if error}
		<div class="p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 rounded-xl text-sm flex items-start gap-3">
			<svg class="w-5 h-5 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/></svg>
			{error}
		</div>
	{/if}

	<!-- Push Notification Settings -->
	<div class="bg-white dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-2xl shadow-sm p-4 md:p-6">
		<h2 class="text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2 mb-4">
			<svg class="w-5 h-5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
				<path stroke-linecap="round" stroke-linejoin="round" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
			</svg>
			Notifikasi Browser (Push)
		</h2>
		<PushSubscriptionManager />
	</div>

	<!-- List -->
	<div class="bg-white dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-2xl shadow-sm overflow-hidden">
		{#if loading}
			<PulseLoader variant="card" count={5} />
		{:else if items.length === 0}
			<div class="text-center py-20 px-4">
				<div class="inline-flex items-center justify-center w-16 h-16 rounded-full bg-gray-100 dark:bg-gray-800 mb-4 text-gray-400">
					<svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
						<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75v-.7V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
					</svg>
				</div>
				<h3 class="text-lg font-medium text-gray-900 dark:text-white mb-1">Semua Tenang!</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400">Belum ada notifikasi baru untuk Anda saat ini.</p>
			</div>
		{:else}
			<div class="divide-y divide-gray-100 dark:divide-gray-800/50">
				{#each items as item}
					{@const style = getTypeIcon(item.notification_type)}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div
						onclick={() => { 
							if (!item.is_read) markAsRead(item.id); 
							expandedId = expandedId === item.id ? null : item.id;
						}}
						class="group flex items-start gap-4 p-4 md:p-6 transition-all duration-200 hover:bg-gray-50 dark:hover:bg-gray-700/30 cursor-pointer relative {item.is_read ? 'opacity-70' : 'bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border-b border-blue-100/50 dark:border-blue-900/20 shadow-sm'}"
					>
						{#if !item.is_read}
							<div class="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 rounded-r-full shadow-[0_0_8px_rgba(59,130,246,0.5)]"></div>
						{/if}
						
						<div class="{style.bgClass} {style.colorClass} w-10 h-10 rounded-full flex items-center justify-center shrink-0 mt-0.5 ring-4 ring-white dark:ring-gray-800">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
								{@html style.svg}
							</svg>
						</div>
						
						<div class="flex-1 min-w-0 pr-4">
							<div class="flex items-center gap-2 mb-1">
								<h3 class="font-semibold text-sm md:text-base text-gray-900 dark:text-white group-hover:text-blue-600 dark:group-hover:text-blue-400 transition-colors">
									{item.title}
								</h3>
								{#if !item.is_read}
									<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium bg-blue-100 text-blue-700 dark:bg-blue-900/50 dark:text-blue-300">Baru</span>
								{/if}
							</div>
							<div class="text-sm text-gray-600 dark:text-gray-300 leading-relaxed mb-3 pr-2 {expandedId === item.id ? '' : 'line-clamp-1'}">
								{item.body}
							</div>
							
							{#if expandedId === item.id}
								<div class="mt-4 mb-4 p-4 bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-gray-100 dark:border-gray-800 animate-in slide-in-from-top-2 duration-200">
									<h4 class="text-xs font-semibold text-gray-500 dark:text-gray-400 mb-2 uppercase tracking-wider">Rincian Pesan</h4>
									<p class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap">{item.body}</p>
								</div>
							{/if}

							<div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500">
								<span class="flex items-center gap-1.5">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
									{formatDate(item.created_at)}
								</span>
								<div class="flex items-center gap-3">
									{#if !item.is_read}
										<button
											onclick={(e) => { e.stopPropagation(); markAsRead(item.id); }}
											class="font-medium text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 opacity-0 group-hover:opacity-100 transition-opacity"
										>
											Tandai dibaca
										</button>
									{/if}
									<svg class="w-4 h-4 transition-transform duration-200 {expandedId === item.id ? "rotate-180" : ""}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
								</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>

	<!-- Pagination -->
	{#if totalPages > 1}
		<div class="flex items-center justify-center gap-2 pt-4">
			<button
				disabled={pageNum === 1}
				onclick={() => { pageNum--; load(); }}
				class="p-2 border border-gray-200 dark:border-gray-700 rounded-lg disabled:opacity-50 hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-400 transition-colors"
				aria-label="Halaman sebelumnya"
			>
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"/></svg>
			</button>
			<div class="text-sm text-gray-600 dark:text-gray-400 font-medium px-4">
				Halaman {pageNum} dari {totalPages}
			</div>
			<button
				disabled={pageNum === totalPages}
				onclick={() => { pageNum++; load(); }}
				class="p-2 border border-gray-200 dark:border-gray-700 rounded-lg disabled:opacity-50 hover:bg-gray-50 dark:hover:bg-gray-800 text-gray-600 dark:text-gray-400 transition-colors"
				aria-label="Halaman berikutnya"
			>
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/></svg>
			</button>
		</div>
	{/if}
</div>
