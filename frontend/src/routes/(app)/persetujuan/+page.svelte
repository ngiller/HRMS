<script lang="ts">
	import { onMount } from 'svelte';
	import { approvals as approvalsApi } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';

	type PendingApproval = {
		tracking_id: string;
		entity_type: string;
		entity_id: string;
		current_step: number;
		total_steps: number;
		requestor_name: string;
		title: string;
		description: string;
		amount: number;
		created_at: string;
	};

	let pendingApprovals = $state<PendingApproval[]>([]);
	let total = $state(0);
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Process state
	let processingId = $state<string | null>(null);
	let showRejectModal = $state(false);
	let rejectEntityType = $state('');
	let rejectEntityId = $state('');
	let rejectReason = $state('');
	let actionLoading = $state(false);
	let actionError = $state('');

	const ENTITY_LABELS: Record<string, string> = {
		leave: 'Cuti',
		overtime: 'Lembur',
		reimbursement: 'Reimbursement',
		shift_change: 'Permintaan Shift',
		loan: 'Pinjaman',
		mutation: 'Mutasi',
	};

	const ENTITY_ICONS: Record<string, string> = {
		leave: 'M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z',
		overtime: 'M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z',
		reimbursement: 'M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z',
		shift_change: 'M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182',
		loan: 'M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Z',
		mutation: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z',
	};

	const ENTITY_COLORS: Record<string, string> = {
		leave: 'from-blue-50 to-blue-100 text-blue-600 ring-blue-200',
		overtime: 'from-amber-50 to-amber-100 text-amber-600 ring-amber-200',
		reimbursement: 'from-emerald-50 to-emerald-100 text-emerald-600 ring-emerald-200',
		shift_change: 'from-purple-50 to-purple-100 text-purple-600 ring-purple-200',
		loan: 'from-rose-50 to-rose-100 text-rose-600 ring-rose-200',
		mutation: 'from-indigo-50 to-indigo-100 text-indigo-600 ring-indigo-200',
	};

	onMount(() => {
		loadPendingApprovals(false);
		// Auto-refresh setiap 30 detik (tanpa loading skeleton)
		const interval = setInterval(() => loadPendingApprovals(true), 30000);
		return () => clearInterval(interval);
	});

	async function loadPendingApprovals(isBackgroundRefresh = false) {
		if (!isBackgroundRefresh) isLoading = true;
		errorMessage = '';
		try {
			const res: any = await approvalsApi.getPending();
			if (res.success && res.data) {
				const d = res.data as { items: PendingApproval[]; total: number };
				pendingApprovals = d.items || [];
				total = d.total || 0;
			} else {
				pendingApprovals = [];
				total = 0;
			}
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memuat data persetujuan';
		} finally {
			isLoading = false;
		}
	}

	function openRejectModal(entityType: string, entityId: string) {
		rejectEntityType = entityType;
		rejectEntityId = entityId;
		rejectReason = '';
		actionError = '';
		showRejectModal = true;
	}

	async function handleApprove(entityType: string, entityId: string) {
		processingId = entityId;
		actionError = '';
		try {
			await approvalsApi.process(entityType, entityId, { action: 'approve', notes: '' });
			await loadPendingApprovals();
		} catch (err: unknown) {
			actionError = (err as { message?: string }).message || 'Gagal menyetujui';
		} finally {
			processingId = null;
		}
	}

	async function handleReject() {
		actionLoading = true;
		actionError = '';
		try {
			await approvalsApi.process(rejectEntityType, rejectEntityId, { action: 'reject', notes: rejectReason });
			showRejectModal = false;
			await loadPendingApprovals();
		} catch (err: unknown) {
			actionError = (err as { message?: string }).message || 'Gagal menolak';
		} finally {
			actionLoading = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function getInitials(name: string): string {
		if (!name) return 'NA';
		return name.split(' ').slice(0, 2).map(s => s[0]).join('').toUpperCase() || name.substring(0, 2).toUpperCase();
	}

	function getColorBase(type: string): string {
		const c = ENTITY_COLORS[type];
		if (!c) return 'gray';
		const match = c.match(/text-([a-z]+)-/);
		return match ? match[1] : 'gray';
	}
</script>

<div class="w-full">
	<div class="flex items-center gap-4 mb-8">
		<div class="w-12 h-12 bg-emerald-600 rounded-2xl flex items-center justify-center text-white font-bold text-lg shrink-0">
			<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
			</svg>
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Persetujuan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">Setujui atau tolak pengajuan yang menunggu persetujuan Anda</p>
		</div>
		{#if total > 0}
			<div class="ml-auto flex items-center gap-2">
				<button onclick={() => loadPendingApprovals(false)} disabled={isLoading}
					class="p-2 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer disabled:opacity-50"
					aria-label="Refresh">
					<svg class="w-4 h-4 {isLoading ? 'animate-spin' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182" /></svg>
				</button>
				<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-semibold bg-emerald-50 dark:bg-emerald-900/30 text-emerald-700 dark:text-emerald-400 ring-1 ring-emerald-600/20">
					<span class="w-2 h-2 bg-emerald-500 rounded-full animate-pulse"></span>
					{total} menunggu
				</span>
			</div>
		{/if}
	</div>

	{#if isLoading}
		<PulseLoader variant="card" count={3} />
	{:else if errorMessage}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-4">
			<div class="flex items-center gap-2.5">
				<svg class="w-5 h-5 text-red-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				<p class="text-sm font-medium text-red-800 dark:text-red-200">{errorMessage}</p>
			</div>
		</div>
	{:else if pendingApprovals.length === 0}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-12 text-center">
			<div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-emerald-50 dark:bg-emerald-900/20 flex items-center justify-center">
				<svg class="w-8 h-8 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
			</div>
			<h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-1">Tidak ada yang menunggu</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400">Semua pengajuan sudah diproses. Anda akan melihat notifikasi jika ada pengajuan baru.</p>
		</div>
	{:else}
		{#if actionError}
			<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-3">
				<p class="text-sm font-medium text-red-800 dark:text-red-200">{actionError}</p>
			</div>
		{/if}

		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-5">
			{#each pendingApprovals as item}
				{@const color = getColorBase(item.entity_type)}
				<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 p-5 flex flex-col relative overflow-hidden group hover:shadow-xl hover:-translate-y-1 transition-all duration-300">
					<!-- Top decorative border -->
					<div class="absolute top-0 left-0 right-0 h-1.5 bg-{color}-500 opacity-80 group-hover:opacity-100 transition-opacity"></div>
					
					<div class="flex items-start justify-between gap-4 mb-4 mt-1">
						<div class="flex items-center gap-3">
							<div class="w-12 h-12 rounded-xl bg-gradient-to-br {ENTITY_COLORS[item.entity_type] || 'from-gray-50 to-gray-100 text-gray-600'} flex items-center justify-center shrink-0 shadow-inner">
								<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d={ENTITY_ICONS[item.entity_type] || 'M12 6v12m6-6H6'} />
								</svg>
							</div>
							<div>
								<span class="inline-block px-2.5 py-0.5 rounded-md text-[10px] font-bold uppercase tracking-wider mb-1 bg-{color}-100 text-{color}-700 dark:bg-{color}-900/40 dark:text-{color}-400">
									{ENTITY_LABELS[item.entity_type] || item.entity_type}
								</span>
								<h3 class="text-base font-bold text-gray-900 dark:text-white line-clamp-1" title={item.title || item.description || 'Pengajuan'}>
									{item.title || item.description || 'Pengajuan'}
								</h3>
							</div>
						</div>
					</div>

					<div class="flex-1 space-y-3.5 mb-6">
						<div class="flex items-center gap-2.5 text-sm text-gray-700 dark:text-gray-300 bg-gray-50 dark:bg-gray-800/80 p-2.5 rounded-lg border border-gray-100 dark:border-gray-700/50">
							<div class="w-7 h-7 rounded-full bg-white dark:bg-gray-700 flex items-center justify-center text-[10px] font-bold text-gray-600 dark:text-gray-300 shadow-sm shrink-0">
								{getInitials(item.requestor_name)}
							</div>
							<span class="font-semibold line-clamp-1">{item.requestor_name}</span>
						</div>
						
						{#if item.amount > 0}
							<div class="flex items-center gap-2 text-sm px-1">
								<svg class="w-4 h-4 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v12m-3-2.818l.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
								<span class="font-bold text-gray-900 dark:text-white truncate">{formatCurrency(item.amount)}</span>
							</div>
						{/if}

						<div class="text-xs text-gray-500 flex items-center gap-2 px-1">
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
							<span class="truncate">{formatDate(item.created_at)}</span>
						</div>

						<div class="pt-2 px-1">
							<div class="flex justify-between text-[11px] font-medium mb-2">
								<span class="text-gray-500 dark:text-gray-400">Progres Persetujuan</span>
								<span class="text-blue-600 dark:text-blue-400 font-bold">{item.current_step} <span class="text-gray-400">/ {item.total_steps}</span></span>
							</div>
							<div class="h-1.5 w-full bg-gray-100 dark:bg-gray-700 rounded-full overflow-hidden flex gap-0.5">
								{#each Array(item.total_steps) as _, i}
									<div class="h-full flex-1 rounded-full {i < item.current_step ? 'bg-blue-500' : 'bg-gray-200 dark:bg-gray-600'}"></div>
								{/each}
							</div>
						</div>
					</div>

					<div class="grid grid-cols-2 gap-3 mt-auto pt-4 border-t border-gray-100 dark:border-gray-700/50">
						<button
							onclick={() => openRejectModal(item.entity_type, item.entity_id)}
							disabled={processingId === item.entity_id}
							class="flex items-center justify-center gap-2 px-4 py-2.5 bg-red-50 hover:bg-red-100 text-red-600 dark:bg-red-500/10 dark:hover:bg-red-500/20 dark:text-red-400 rounded-xl text-sm font-bold transition-colors disabled:opacity-50 outline-none focus:ring-2 focus:ring-red-500 focus:ring-offset-1 dark:focus:ring-offset-gray-800"
						>
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
							Tolak
						</button>
						<button
							onclick={() => handleApprove(item.entity_type, item.entity_id)}
							disabled={processingId === item.entity_id}
							class="flex items-center justify-center gap-2 px-4 py-2.5 bg-[#1A56DB] hover:bg-[#1e40af] text-white rounded-xl text-sm font-bold transition-all hover:shadow-lg hover:shadow-blue-500/30 disabled:opacity-50 outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1 dark:focus:ring-offset-gray-800 active:scale-[0.98]"
						>
							{#if processingId === item.entity_id}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							{:else}
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
							{/if}
							Setujui
						</button>
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>

<!-- Reject Modal -->
<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={() => { if (!actionLoading) showRejectModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') showRejectModal = false; }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak pengajuan" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-12 h-12 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/30 flex items-center justify-center">
					<svg class="w-6 h-6 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 text-center mb-2">Tolak Pengajuan</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 text-center mb-5">Berikan alasan penolakan</p>
				
				{#if actionError}
					<div class="mb-3 text-xs text-red-600 bg-red-50 dark:bg-red-900/20 rounded-lg px-3 py-2">{actionError}</div>
				{/if}
				
				<textarea bind:value={rejectReason} rows="3" placeholder="Alasan penolakan..."
					class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-red-500/20 focus:border-red-500 transition bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 placeholder:text-gray-400 resize-none"></textarea>
				
				<div class="flex items-center justify-center gap-3 mt-5">
					<button onclick={() => showRejectModal = false} disabled={actionLoading}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={actionLoading}
						class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if actionLoading}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						{/if}
						Tolak Pengajuan
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
