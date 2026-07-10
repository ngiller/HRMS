<script lang="ts">
	import { onMount } from 'svelte';
	import { manualAttendance } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';

	interface ManualAttendanceItem {
		id: string;
		employee_id: string;
		employee_name: string;
		date: string;
		check_in_time: string | null;
		check_out_time: string | null;
		reason: string;
		status: string;
		approved_by: string | null;
		approved_by_name: string;
		approved_at: string | null;
		rejection_reason: string;
		created_at: string;
	}

	let items = $state<ManualAttendanceItem[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let form = $state({ date: '', check_in_time: '', check_out_time: '', reason: '' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailData = $state<ManualAttendanceItem | null>(null);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
	};

	function getStatusBadge(status: string): string {
		const labels: Record<string, string> = {
			pending: 'Menunggu', approved: 'Disetujui', rejected: 'Ditolak', cancelled: 'Dibatalkan'
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600'}">${labels[status] || status}</span>`;
	}

	onMount(() => load());

	async function load() {
		isLoading = true; errorMessage = '';
		try {
			const res = await manualAttendance.list(page, perPage, statusFilter);
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
		form = { date: '', check_in_time: '', check_out_time: '', reason: '' };
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.date) { formError = 'Tanggal harus diisi'; return; }
		if (!form.reason.trim()) { formError = 'Alasan harus diisi'; return; }
		if (!form.check_in_time && !form.check_out_time) { formError = 'Setidaknya satu waktu (check-in atau check-out) harus diisi'; return; }

		isSaving = true; formError = '';
		try {
			await manualAttendance.create(form);
			cancelForm();
			load();
		} catch (e: unknown) {
			formError = (e as { message?: string }).message || 'Gagal menyimpan';
		} finally {
			isSaving = false;
		}
	}

	async function openDetail(id: string) {
		showDetail = true; detailData = null;
		try {
			const res = await manualAttendance.get(id);
			detailData = res.data ?? null;
		} catch {
			detailData = null;
		}
	}

	function closeDetail() { showDetail = false; detailData = null; }

	async function handleApprove(id: string) {
		try { await manualAttendance.approve(id); load(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menyetujui'; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		try { await manualAttendance.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; load(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
	}

	async function handleCancel(id: string) {
		try { await manualAttendance.cancel(id); load(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal membatalkan'; }
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function formatTime(iso: string | null): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight dark:text-white">Absensi Manual</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan absensi manual jika lupa check-in/out</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail && hasPermission('attendance', 'create')}
				<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					Ajukan Absensi Manual
				</button>
			{/if}
		</div>
	</div>

	<!-- Filter -->
	{#if !showForm && !showDetail}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 dark:bg-gray-800 dark:border-gray-700">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">Semua</button>
				{#each ['pending', 'approved', 'rejected', 'cancelled'] as status}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">{status}</button>
				{/each}
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} pengajuan ditemukan` : ''}</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Ajukan Absensi Manual</h2>
				<button onclick={cancelForm} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tanggal <span class="text-red-500">*</span></label>
						<input type="date" bind:value={form.date} max={new Date().toISOString().split('T')[0]} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Alasan <span class="text-red-500">*</span></label>
						<input type="text" bind:value={form.reason} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition dark:bg-gray-700 dark:border-gray-600 dark:text-white" placeholder="Contoh: Lupa check-in karena ada keperluan mendesak" />
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jam Check-In</label>
						<input type="time" bind:value={form.check_in_time} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jam Check-Out</label>
						<input type="time" bind:value={form.check_out_time} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
				</div>
				<p class="text-xs text-gray-400">Catatan: Setidaknya salah satu waktu (check-in atau check-out) harus diisi. Maksimal 3 pengajuan per bulan.</p>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer dark:border-gray-600 dark:text-gray-300">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Absensi Manual</h2>
				<button onclick={closeDetail} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if detailData}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3 dark:text-gray-500">Informasi</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal</span><p class="text-sm font-medium text-gray-900 dark:text-white">{formatDate(detailData.date)}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(detailData.status)}</p></div>
								<div><span class="text-xs text-gray-400">Alasan</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.reason || '-'}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3 dark:text-gray-500">Waktu</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Check-In</span><p class="text-sm font-medium text-gray-900 dark:text-white">{formatTime(detailData.check_in_time)}</p></div>
								<div><span class="text-xs text-gray-400">Check-Out</span><p class="text-sm font-medium text-gray-900 dark:text-white">{formatTime(detailData.check_out_time)}</p></div>
								{#if detailData.approved_by_name}
									<div><span class="text-xs text-gray-400">Disetujui/Ditolak Oleh</span><p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.approved_by_name}</p></div>
								{/if}
								{#if detailData.rejection_reason}
									<div><span class="text-xs text-gray-400">Alasan Penolakan</span><p class="text-sm text-red-600">{detailData.rejection_reason}</p></div>
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail</p>
				{/if}
			</div>
			{#if detailData && detailData.status === 'pending' && hasPermission('attendance', 'approve')}
				{@const dataId = detailData.id}
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<button onclick={() => { openReject(dataId); closeDetail(); }} class="px-5 py-2.5 bg-red-50 text-red-700 rounded-lg text-sm font-semibold hover:bg-red-100 transition cursor-pointer dark:bg-red-900/30 dark:text-red-300 dark:hover:bg-red-900/50">Tolak</button>
				<button onclick={() => { handleApprove(dataId); closeDetail(); }} class="px-5 py-2.5 bg-green-50 text-green-700 rounded-lg text-sm font-semibold hover:bg-green-100 transition cursor-pointer dark:bg-green-900/30 dark:text-green-300 dark:hover:bg-green-900/50">Setujui</button>
			</div>
			{/if}
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
				<EmptyState
					variant="empty"
					title="Belum ada pengajuan"
					description="Belum ada pengajuan absensi manual."
					actionLabel={hasPermission('attendance', 'create') ? 'Ajukan Absensi Manual' : undefined}
					onAction={hasPermission('attendance', 'create') ? openCreateForm : undefined}
				/>
			{:else}
				<div class="space-y-2 p-4">
					{#each items as item}
						<MobileCard
							title={item.employee_name || 'Saya'}
							subtitle={formatDate(item.date)}
							avatar={getInitials(item.employee_name || 'Saya')}
							avatarColor={getAvatarTheme('attendance').gradientClasses}
							badges={[{ label: ({ pending: 'Menunggu', approved: 'Disetujui', rejected: 'Ditolak', cancelled: 'Dibatalkan' })[item.status] || item.status, color: statusColors[item.status] || 'bg-gray-50 text-gray-600' }]}
							onclick={() => openDetail(item.id)}
						>
							{#snippet children()}
								<div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
									<span>In: {formatTime(item.check_in_time)}</span>
									<span>Out: {formatTime(item.check_out_time)}</span>
								</div>
								<p class="text-xs text-gray-600 dark:text-gray-300 line-clamp-2 mt-1">{item.reason}</p>
							{/snippet}
							{#snippet footer()}
								{#if item.status === 'pending'}
									<div class="flex items-center gap-2">
										<button onclick={(e) => { e.stopPropagation(); openDetail(item.id); }} class="flex-1 py-2 text-xs font-medium text-gray-500 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer dark:text-gray-400">Detail</button>
										{#if hasPermission('attendance', 'approve')}
											<button onclick={(e) => { e.stopPropagation(); handleApprove(item.id); }} class="flex-1 py-2 text-xs font-semibold text-green-700 bg-green-50 dark:bg-green-900/30 rounded-lg hover:bg-green-100 dark:hover:bg-green-900/50 transition cursor-pointer dark:text-green-300">Setujui</button>
											<button onclick={(e) => { e.stopPropagation(); openReject(item.id); }} class="flex-1 py-2 text-xs font-semibold text-red-600 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer dark:text-red-300">Tolak</button>
										{/if}
										{#if hasPermission('attendance', 'create')}
											<button onclick={(e) => { e.stopPropagation(); handleCancel(item.id); }} class="flex-1 py-2 text-xs font-medium text-orange-600 bg-orange-50 dark:bg-orange-900/30 rounded-lg hover:bg-orange-100 dark:hover:bg-orange-900/50 transition cursor-pointer dark:text-orange-300">Batalkan</button>
										{/if}
									</div>
								{/if}
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
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Pengajuan Absensi Manual</h3>
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
