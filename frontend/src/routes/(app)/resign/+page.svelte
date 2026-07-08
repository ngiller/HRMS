<script lang="ts">
	import { onMount } from 'svelte';
	import { resign } from '$lib/api.js';
	import { hasPermission, getUser } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';

	interface ResignItem {
		id: string;
		employee_id: string;
		employee_name: string;
		resign_date: string;
		last_working_date: string;
		reason: string;
		resign_type: string;
		status: string;
		approved_by_name: string;
		rejection_reason: string;
		created_at: string;
	}

	interface ClearanceItem {
		id: string;
		resign_id: string;
		item_name: string;
		description: string;
		is_checked: boolean;
		checked_by_name: string;
		sort_order: number;
	}

	let items = $state<ResignItem[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let form = $state({ last_working_date: '', reason: '', resign_type: 'voluntary' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailData = $state<ResignItem | null>(null);
	let clearanceItems = $state<ClearanceItem[]>([]);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	let currentUser = $state(getUser());

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		processed: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
	};

	const typeLabels: Record<string, string> = {
		voluntary: 'Mengundurkan Diri',
		termination: 'PHK',
		retirement: 'Pensiun',
		mutual: 'Kesepakatan Bersama',
	};

	function getStatusBadge(status: string): string {
		const labels: Record<string, string> = {
			pending: 'Menunggu', approved: 'Disetujui', rejected: 'Ditolak', processed: 'Diproses'
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600'}">${labels[status] || status}</span>`;
	}

	onMount(() => load());

	async function load() {
		isLoading = true; errorMessage = '';
		try {
			const res = await resign.list(page, perPage, statusFilter);
			if (res.success) {
				items = res.data || [];
				total = (res as any).meta?.total || 0;
				page = (res as any).meta?.page || 1;
				perPage = (res as any).meta?.per_page || 25;
				totalPages = Math.ceil(total / perPage);
			}
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	function openCreateForm() {
		form = { last_working_date: '', reason: '', resign_type: 'voluntary' };
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.last_working_date) { formError = 'Tanggal terakhir kerja harus diisi'; return; }
		if (!form.reason.trim()) { formError = 'Alasan resign harus diisi'; return; }

		isSaving = true; formError = '';
		try {
			await resign.create(form);
			cancelForm();
			load();
		} catch (e: unknown) {
			formError = (e as { message?: string }).message || 'Gagal menyimpan';
		} finally {
			isSaving = false;
		}
	}

	async	function openDetail(id: string) {
		showDetail = true; detailData = null; clearanceItems = [];
		try {
			const res = await resign.get(id);
			const data: ResignItem | null = res.data ?? null;
			detailData = data;
			const clearRes = await resign.listClearance(id);
			if (clearRes.success && clearRes.data) clearanceItems = clearRes.data as ClearanceItem[] || [];
		} catch {
			detailData = null;
		}
	}

	function closeDetail() { showDetail = false; detailData = null; clearanceItems = []; }

	async function handleApprove(id: string) {
		try { await resign.approve(id); load(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menyetujui'; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		try { await resign.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; load(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
	}

	async function toggleClearanceItem(item: ClearanceItem) {
		try {
			await resign.updateClearance(item.id, { is_checked: !item.is_checked });
			if (detailData) await openDetail((detailData as ResignItem).id);
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal update clearance';
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight dark:text-white">Resign & Exit</h1>
			<p class="text-sm text-gray-500 mt-0.5">Kelola pengajuan resign dan exit clearance karyawan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail && hasPermission('employee', 'create')}
				<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					Ajukan Resign
				</button>
			{/if}
		</div>
	</div>

	{#if !showForm && !showDetail}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 dark:bg-gray-800 dark:border-gray-700">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">Semua</button>
				{#each ['pending', 'approved', 'rejected', 'processed'] as status}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">{status}</button>
				{/each}
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} pengajuan ditemukan` : ''}</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Ajukan Resign</h2>
				<button onclick={cancelForm} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jenis Resign</label>
						<select bind:value={form.resign_type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="voluntary">Mengundurkan Diri</option>
							<option value="termination">PHK</option>
							<option value="retirement">Pensiun</option>
							<option value="mutual">Kesepakatan Bersama</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tanggal Terakhir Kerja <span class="text-red-500">*</span></label>
						<input type="date" bind:value={form.last_working_date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Alasan Resign <span class="text-red-500">*</span></label>
					<textarea bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white resize-none" placeholder="Jelaskan alasan resign..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer dark:border-gray-600 dark:text-gray-300">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Resign
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Resign</h2>
				<button onclick={closeDetail} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if detailData}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3 dark:text-gray-500">Informasi Resign</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Jenis</span><p class="text-sm text-gray-700 dark:text-gray-300">{typeLabels[detailData.resign_type] || detailData.resign_type}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(detailData.status)}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal Terakhir Kerja</span><p class="text-sm font-medium text-gray-900 dark:text-white">{formatDate(detailData.last_working_date)}</p></div>
								<div><span class="text-xs text-gray-400">Alasan</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.reason || '-'}</p></div>
								{#if detailData.rejection_reason}
									<div><span class="text-xs text-gray-400">Alasan Penolakan</span><p class="text-sm text-red-600">{detailData.rejection_reason}</p></div>
								{/if}
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3 dark:text-gray-500">Exit Clearance Checklist</h3>
							{#if clearanceItems.length > 0}
								<div class="space-y-2">
									{#each clearanceItems as item}
										<!-- svelte-ignore a11y_click_events_have_key_events -->
										<!-- svelte-ignore a11y_no_static_element_interactions -->
										<div onclick={hasPermission('employee', 'update') ? () => toggleClearanceItem(item) : undefined} class="flex items-center gap-3 p-2 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700/50 transition cursor-pointer {item.is_checked ? 'bg-green-50/50 dark:bg-green-900/10' : ''}">
											<div class="w-5 h-5 rounded border-2 flex items-center justify-center shrink-0 {item.is_checked ? 'bg-green-500 border-green-500' : 'border-gray-300 dark:border-gray-600'}">
												{#if item.is_checked}
													<svg class="w-3 h-3 text-white" fill="none" viewBox="0 0 24 24" stroke-width="3" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
												{/if}
											</div>
											<div class="flex-1 min-w-0">
												<div class="text-sm font-medium {item.is_checked ? 'text-green-700 dark:text-green-300 line-through' : 'text-gray-700 dark:text-gray-300'}">{item.item_name}</div>
												{#if item.description}<div class="text-xs text-gray-400">{item.description}</div>{/if}
											</div>
											{#if item.checked_by_name}
												<span class="text-xs text-gray-400 shrink-0">oleh {item.checked_by_name}</span>
											{/if}
										</div>
									{/each}
								</div>
							{:else}
								<p class="text-sm text-gray-500 italic">Belum ada item clearance</p>
							{/if}
						</div>
					</div>

					{#if detailData.status === 'pending' && hasPermission('employee', 'update')}
						{@const dd = detailData}
						<div class="mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
							<div class="flex items-center gap-3">
								<button onclick={() => handleApprove(dd.id)} disabled={clearanceItems.some(i => !i.is_checked)} class="px-5 py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer" title={clearanceItems.some(i => !i.is_checked) ? 'Selesaikan semua clearance terlebih dahulu' : ''}>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
									Setujui & Proses Resign
								</button>
								<button onclick={() => openReject(dd.id)} class="px-5 py-2.5 border border-red-200 text-red-600 rounded-lg text-sm font-semibold hover:bg-red-50 transition cursor-pointer">
									Tolak
								</button>
							</div>
							{#if clearanceItems.some(i => !i.is_checked)}
								<p class="text-xs text-amber-600 mt-2">Semua item exit clearance harus dicentang sebelum resign dapat disetujui.</p>
							{/if}
						</div>
					{/if}
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<PullToRefresh onRefresh={load}>
			{#if isLoading}
				<div class="p-6 animate-pulse space-y-3">{#each [1,2,3] as _}<div class="h-16 bg-gray-100 rounded-lg dark:bg-gray-700"></div>{/each}</div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<div class="py-8 text-center">
					<EmptyState variant="empty" title="Belum ada pengajuan resign" description="Belum ada pengajuan resign." />
					{#if hasPermission('employee', 'create')}
						<button onclick={openCreateForm} class="mt-4 px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Ajukan Resign</button>
					{/if}
				</div>
			{:else}
				<div class="space-y-2">
					{#each items as item}
						<MobileCard
							title={item.employee_name || 'Saya'}
							subtitle={`${typeLabels[item.resign_type] || item.resign_type} • ${formatDate(item.last_working_date)}`}
							avatar={getInitials(item.employee_name || 'Saya')}
							avatarColor={getAvatarTheme('resign').gradientClasses}
							onclick={() => openDetail(item.id)}
						>
							{#snippet children()}
								<p class="text-xs text-gray-600 dark:text-gray-300 line-clamp-2">{item.reason}</p>
							{/snippet}
							{#snippet footer()}
								{@html getStatusBadge(item.status)}
							{/snippet}
						</MobileCard>
					{/each}
				</div>
				<div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-700">
					<div class="text-xs text-gray-400">{(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari {total}</div>
					<div class="flex items-center gap-2">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-30 cursor-pointer dark:border-gray-600 dark:text-gray-400">Sebelumnya</button>
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-30 cursor-pointer dark:border-gray-600 dark:text-gray-400">Selanjutnya</button>
					</div>
				</div>
			{/if}
			</PullToRefresh>
		</div>
	{/if}
</div>

<!-- Reject Modal -->
<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<div onkeydown={(e) => { if (e.key === 'Escape') cancelReject(); }} class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Pengajuan Resign</h3>
				<div class="space-y-3">
					<label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer dark:border-gray-600 dark:text-gray-300">Batal</button>
					<button onclick={handleReject} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition cursor-pointer">Tolak</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
