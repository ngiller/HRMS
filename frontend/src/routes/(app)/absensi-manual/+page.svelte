<script lang="ts">
	import { onMount } from 'svelte';
	import { manualAttendance, auth } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import { AnimatedPresence } from '$lib';
import EmptyState from '$lib/components/EmptyState.svelte';
	import StaggerList from '$lib/components/StaggerList.svelte';
import { getInitials } from '$lib/avatar-theme.js';

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
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;
	let errorMessage = $state('');

	let showForm = $state(false);
	let form = $state({ date: '', check_in_time: '', check_out_time: '', reason: '' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailData = $state<ManualAttendanceItem | null>(null);

	let showRejectModal = $state(false);
	let filteredItems = $derived.by(() => {
		if (!searchQuery.trim()) return items;
		const q = searchQuery.toLowerCase();
		return items.filter(i =>
			i.employee_name?.toLowerCase().includes(q) ||
			i.reason?.toLowerCase().includes(q)
		);
	});

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; }, 400);
	}
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	// Current user info for self-approval prevention (compare UUID with item.employee_id)
	let currentUser = $derived(auth.getUser() as { id?: string } | null);
	let currentUserId = $derived(currentUser?.id || '');

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
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${labels[status] || status}</span>`;
	}

	onMount(() => {
		const params = new URLSearchParams(window.location.search);
		// Auto-filter by status from URL (e.g. ?status=pending)
		const statusParam = params.get('status');
		if (statusParam && ['pending', 'approved', 'rejected', 'cancelled'].includes(statusParam)) {
			statusFilter = statusParam;
		}
		load();
		if (params.get('action') === 'create') {
			const dateParam = params.get('date');
			if (dateParam) {
				form.date = dateParam.split('T')[0];
			}
			const checkInParam = params.get('check_in');
			if (checkInParam) {
				form.check_in_time = checkInParam;
			}
			const checkOutParam = params.get('check_out');
			if (checkOutParam) {
				form.check_out_time = checkOutParam;
			}
			showForm = true;
		}
	});

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
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				<input type="search" value={searchQuery} placeholder="Cari karyawan..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
			</div>
			
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">Semua</button>
				{#each ['pending', 'approved', 'rejected', 'cancelled'] as status (status)}
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
					<!-- Status Hero Banner -->
					<div class="rounded-2xl p-5 mb-5 {detailData.status === 'approved' ? 'bg-gradient-to-br from-emerald-50 to-emerald-100/60 dark:from-emerald-950/30 dark:to-emerald-900/20 border border-emerald-200 dark:border-emerald-800' : detailData.status === 'rejected' ? 'bg-gradient-to-br from-red-50 to-red-100/60 dark:from-red-950/30 dark:to-red-900/20 border border-red-200 dark:border-red-800' : 'bg-gradient-to-br from-amber-50 to-amber-100/60 dark:from-amber-950/30 dark:to-amber-900/20 border border-amber-200 dark:border-amber-800'}">
						<div class="flex items-center gap-3 mb-1.5">
							{#if detailData.status === 'approved'}
								<div class="w-10 h-10 rounded-full bg-emerald-100 dark:bg-emerald-800/50 flex items-center justify-center shrink-0">
									<svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								</div>
								<div>
									<p class="text-lg font-bold text-emerald-800 dark:text-emerald-200">Disetujui</p>
									<p class="text-sm text-emerald-600 dark:text-emerald-400">Pengajuan absensi manual telah disetujui</p>
								</div>
							{:else if detailData.status === 'rejected'}
								<div class="w-10 h-10 rounded-full bg-red-100 dark:bg-red-800/50 flex items-center justify-center shrink-0">
									<svg class="w-5 h-5 text-red-600 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
								</div>
								<div>
									<p class="text-lg font-bold text-red-800 dark:text-red-200">Ditolak</p>
									<p class="text-sm text-red-600 dark:text-red-400">Pengajuan absensi manual ditolak</p>
								</div>
							{:else}
								<div class="w-10 h-10 rounded-full bg-amber-100 dark:bg-amber-800/50 flex items-center justify-center shrink-0">
									<svg class="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								</div>
								<div>
									<p class="text-lg font-bold text-amber-800 dark:text-amber-200">Menunggu Persetujuan</p>
									<p class="text-sm text-amber-600 dark:text-amber-400">Pengajuan ini masih menunggu review</p>
								</div>
							{/if}
						</div>
					</div>

					<!-- Employee Info Row -->
					<div class="flex items-center gap-4 p-4 bg-gray-50 dark:bg-gray-900/50 rounded-2xl mb-5">
						<div class="w-12 h-12 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white font-bold text-lg shrink-0 shadow-sm">
							{getInitials(detailData.employee_name || '?')}
						</div>
						<div class="flex-1 min-w-0">
							<p class="text-base font-semibold text-gray-900 dark:text-white truncate">{detailData.employee_name || '-'}</p>
							<div class="flex items-center gap-2 text-sm text-gray-500 dark:text-gray-400">
								<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
								<span>{formatDate(detailData.date)}</span>
							</div>
						</div>
						<!-- Status Badge -->
						<!-- eslint-disable-next-line svelte/no-at-html-tags -->
						<div class="shrink-0">{@html getStatusBadge(detailData.status)}</div>
					</div>

					<!-- Time Cards -->
					<div class="grid grid-cols-2 gap-3 mb-5">
						<div class="p-4 bg-white dark:bg-gray-900/50 rounded-2xl border border-gray-100 dark:border-gray-700/50 text-center">
							<div class="w-9 h-9 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center mx-auto mb-2">
								<svg class="w-4 h-4 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
							</div>
							<p class="text-[10px] font-medium text-gray-400 uppercase tracking-wider mb-1">Check-In</p>
							<p class="text-xl font-bold text-gray-900 dark:text-white tabular-nums">{formatTime(detailData.check_in_time)}</p>
						</div>
						<div class="p-4 bg-white dark:bg-gray-900/50 rounded-2xl border border-gray-100 dark:border-gray-700/50 text-center">
							<div class="w-9 h-9 rounded-xl bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center mx-auto mb-2">
								<svg class="w-4 h-4 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
							</div>
							<p class="text-[10px] font-medium text-gray-400 uppercase tracking-wider mb-1">Check-Out</p>
							<p class="text-xl font-bold text-gray-900 dark:text-white tabular-nums">{formatTime(detailData.check_out_time)}</p>
						</div>
					</div>

					<!-- Reason Card -->
					<div class="p-4 bg-white dark:bg-gray-900/50 rounded-2xl border border-gray-100 dark:border-gray-700/50 mb-5">
						<div class="flex items-start gap-3">
							<div class="w-8 h-8 rounded-lg bg-purple-50 dark:bg-purple-900/30 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.129.166 2.27.293 3.423.379.35.026.67.21.865.501L12 21l2.755-4.133a1.14 1.14 0 0 1 .865-.501 48.172 48.172 0 0 0 3.423-.379c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.394 48.394 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" /></svg>
							</div>
							<div class="min-w-0 flex-1">
								<p class="text-xs font-medium text-gray-400 uppercase tracking-wider mb-1">Alasan Pengajuan</p>
								<p class="text-sm text-gray-700 dark:text-gray-300">{detailData.reason || '-'}</p>
							</div>
						</div>
					</div>

					<!-- Approval / Rejection Info -->
					{#if detailData.status === 'approved' || detailData.status === 'rejected'}
						<div class="rounded-2xl overflow-hidden border {detailData.status === 'approved' ? 'border-emerald-200 dark:border-emerald-800' : 'border-red-200 dark:border-red-800'}">
							<div class="px-4 py-2.5 {detailData.status === 'approved' ? 'bg-emerald-50 dark:bg-emerald-900/20' : 'bg-red-50 dark:bg-red-900/20'}">
								<p class="text-xs font-semibold uppercase tracking-wider {detailData.status === 'approved' ? 'text-emerald-700 dark:text-emerald-300' : 'text-red-700 dark:text-red-300'}">
									{detailData.status === 'approved' ? 'Disetujui Oleh' : 'Ditolak Oleh'}
								</p>
							</div>
							<div class="p-4 bg-white dark:bg-gray-900/50 space-y-3">
								{#if detailData.approved_by_name}
									<div class="flex items-center gap-3">
										<div class="w-9 h-9 rounded-full {detailData.status === 'approved' ? 'bg-emerald-100 dark:bg-emerald-800/40 text-emerald-600 dark:text-emerald-400' : 'bg-red-100 dark:bg-red-800/40 text-red-600 dark:text-red-400'} flex items-center justify-center text-sm font-bold shrink-0">
											{detailData.approved_by_name?.charAt(0) || '?'}
										</div>
										<div>
											<p class="text-sm font-semibold text-gray-900 dark:text-white">{detailData.approved_by_name}</p>
											{#if detailData.approved_at}
												<p class="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1.5">
													<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
													{formatDate(detailData.approved_at)}
												</p>
											{/if}
										</div>
									</div>
								{:else}
									<p class="text-sm text-gray-500 italic">Data approver tidak tersedia</p>
								{/if}
								{#if detailData.rejection_reason}
									<div class="pt-3 border-t {detailData.status === 'rejected' ? 'border-red-100 dark:border-red-800/50' : 'border-gray-100 dark:border-gray-700/50'}">
										<p class="text-xs font-semibold text-red-600 dark:text-red-400 uppercase tracking-wider mb-1">Alasan Penolakan</p>
										<div class="flex items-start gap-2 p-3 {detailData.status === 'rejected' ? 'bg-red-50 dark:bg-red-950/30' : 'bg-gray-50 dark:bg-gray-800/50'} rounded-xl">
											<svg class="w-4 h-4 text-red-500 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
											<p class="text-sm text-gray-700 dark:text-gray-300">{detailData.rejection_reason}</p>
										</div>
									</div>
								{/if}
							</div>
						</div>
					{/if}
				{:else}
					<div class="flex flex-col items-center justify-center py-16">
						<div class="w-16 h-16 rounded-full bg-red-50 dark:bg-red-900/30 flex items-center justify-center mb-4">
							<svg class="w-8 h-8 text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
						</div>
						<p class="text-sm font-medium text-gray-500">Gagal memuat detail</p>
						<button onclick={closeDetail} class="mt-4 px-4 py-2 text-sm font-medium text-gray-600 bg-gray-100 rounded-lg hover:bg-gray-200 transition cursor-pointer">Kembali</button>
					</div>
				{/if}
			</div>
			{#if detailData && detailData.status === 'pending' && hasPermission('attendance', 'approve') && detailData.employee_id !== currentUserId}
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
				<div class="p-6 animate-pulse space-y-3">{#each [1,2,3] as _ (_)}<div class="h-16 bg-gray-100 rounded-lg dark:bg-gray-700"></div>{/each}</div>
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
					<StaggerList items={filteredItems}>
					{#snippet children(item)}
						<div onclick={() => openDetail(item.id)} class="group bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-100 dark:border-gray-700/50 hover:border-gray-200 dark:hover:border-gray-600 hover:shadow-md active:shadow-sm transition-all duration-200 cursor-pointer overflow-hidden active:scale-[0.99]">
							<!-- Status indicator top bar -->
							<div class="h-1 {item.status === 'approved' ? 'bg-emerald-500' : item.status === 'rejected' ? 'bg-red-500' : item.status === 'cancelled' ? 'bg-gray-400' : 'bg-amber-400'}"></div>
							
							<div class="p-4">
								<!-- Header: Avatar + Name + Status -->
								<div class="flex items-center gap-3 mb-3">
									<div class="w-10 h-10 rounded-full bg-gradient-to-br from-blue-500 to-indigo-600 flex items-center justify-center text-white text-sm font-bold shrink-0 shadow-sm">
										{getInitials(item.employee_name || '?')}
									</div>
									<div class="flex-1 min-w-0">
										<p class="text-sm font-semibold text-gray-900 dark:text-white truncate">{item.employee_name || 'Saya'}</p>
										<p class="text-xs text-gray-500 dark:text-gray-400">{formatDate(item.date)}</p>
									</div>
									<!-- eslint-disable-next-line svelte/no-at-html-tags -->
									<div class="shrink-0">{@html getStatusBadge(item.status)}</div>
								</div>

								<!-- Time Row -->
								<div class="flex items-center gap-4 mb-2.5 px-1">
									<div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
										<svg class="w-3.5 h-3.5 text-blue-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
										<span class="font-medium tabular-nums">{formatTime(item.check_in_time)}</span>
									</div>
									<svg class="w-3 h-3 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" /></svg>
									<div class="flex items-center gap-1.5 text-xs text-gray-600 dark:text-gray-400">
										<svg class="w-3.5 h-3.5 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
										<span class="font-medium tabular-nums">{formatTime(item.check_out_time)}</span>
									</div>
								</div>

								<!-- Reason snippet -->
								{#if item.reason}
									<p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-1 px-1 mb-3">
										<svg class="w-3 h-3 inline -mt-0.5 mr-1 text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.5 8.25h9m-9 3H12m-9.75 1.51c0 1.6 1.123 2.994 2.707 3.227 1.129.166 2.27.293 3.423.379.35.026.67.21.865.501L12 21l2.755-4.133a1.14 1.14 0 0 1 .865-.501 48.172 48.172 0 0 0 3.423-.379c1.584-.233 2.707-1.626 2.707-3.228V6.741c0-1.602-1.123-2.995-2.707-3.228A48.394 48.394 0 0 0 12 3c-2.392 0-4.744.175-7.043.513C3.373 3.746 2.25 5.14 2.25 6.741v6.018Z" /></svg>
										{item.reason}
									</p>
								{/if}

								<!-- Approver info (for non-pending) -->
								{#if item.status !== 'pending' && item.approved_by_name}
									<div class="flex items-center gap-2 px-1 mb-3 {item.status === 'rejected' ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">
										{#if item.status === 'approved'}
											<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
										{:else}
											<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
										{/if}
										<div class="flex items-center gap-1.5 min-w-0">
											<span class="text-[10px] font-medium truncate">
												{item.status === 'approved' ? 'Disetujui' : 'Ditolak'} oleh <strong>{item.approved_by_name}</strong>
											</span>
											{#if item.approved_at}
												<span class="text-[9px] text-gray-400 dark:text-gray-500 shrink-0">· {formatDate(item.approved_at)}</span>
											{/if}
										</div>
									</div>
								{/if}

								<!-- Actions (only for pending) -->
								{#if item.status === 'pending'}
									<div class="flex items-center gap-2 pt-3 border-t border-gray-50 dark:border-gray-700/50" onclick={(e) => e.stopPropagation()}>
										<button onclick={() => openDetail(item.id)} class="flex-1 py-2 text-xs font-medium text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-xl hover:bg-gray-100 dark:hover:bg-gray-700 transition active:scale-[0.97] cursor-pointer">
											Detail
										</button>
										{#if hasPermission('attendance', 'approve') && item.employee_id !== currentUserId}
											<button onclick={() => handleApprove(item.id)} class="flex-1 py-2 text-xs font-semibold text-emerald-700 dark:text-emerald-300 bg-emerald-50 dark:bg-emerald-900/30 rounded-xl hover:bg-emerald-100 dark:hover:bg-emerald-900/50 transition active:scale-[0.97] cursor-pointer">
												Setujui
											</button>
											<button onclick={() => openReject(item.id)} class="flex-1 py-2 text-xs font-semibold text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-xl hover:bg-red-100 dark:hover:bg-red-900/50 transition active:scale-[0.97] cursor-pointer">
												Tolak
											</button>
										{/if}
										{#if hasPermission('attendance', 'create')}
											<button onclick={() => handleCancel(item.id)} class="flex-1 py-2 text-xs font-medium text-orange-600 dark:text-orange-300 bg-orange-50 dark:bg-orange-900/30 rounded-xl hover:bg-orange-100 dark:hover:bg-orange-900/50 transition active:scale-[0.97] cursor-pointer">
												Batalkan
											</button>
										{/if}
									</div>
								{:else}
									<div class="flex items-center gap-2 pt-2" onclick={(e) => e.stopPropagation()}>
										<button onclick={() => openDetail(item.id)} class="w-full py-2 text-xs font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded-xl hover:bg-blue-100 dark:hover:bg-blue-900/40 transition active:scale-[0.97] cursor-pointer">
											Lihat Detail
										</button>
									</div>
								{/if}
							</div>
						</div>
					{/snippet}
					</StaggerList>
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
