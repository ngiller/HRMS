<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { leaveRequests as api, auth } from '$lib/api.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import SwipeActions from '$lib/components/SwipeActions.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import StaggerList from '$lib/components/StaggerList.svelte';
	import { getAvatarTheme } from '$lib/avatar-theme.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { LeaveRequest, LeaveType, LeaveBalance, LeaveForm, ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

	let items = $state<LeaveRequest[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let form = $state<LeaveForm>({
		leave_type_id: '', start_date: '', end_date: '', total_days: 1,
		is_half_day: false, reason: '', document_url: '', contact_during_leave: '',
	});
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailData = $state<LeaveRequest | null>(null);
	let isDetailLoading = $state(false);

	let leaveTypes = $state<LeaveType[]>([]);
	let myBalances = $state<LeaveBalance[]>([]);
	let showBalances = $state(true);

	// Filter balances based on user's gender & marital status
	let filteredBalances = $derived.by(() => {
		const user = auth.getUser() as Record<string, any> | null;
		const gender = user?.gender || '';
		const marital = user?.marital_status || '';
		
		if (!myBalances || myBalances.length === 0) return [];
		
		return myBalances.filter(b => {
			const code = (b as any).leave_type_code || '';
			
			// Cuti Hamil/Melahirkan: hanya untuk perempuan
			if (code === 'melahirkan' && gender !== 'perempuan') return false;
			// Cuti Keguguran: hanya untuk perempuan
			if (code === 'keguguran' && gender !== 'perempuan') return false;
			// Cuti Menikah: hanya untuk lajang (belum menikah)
			if (code === 'menikah' && marital !== 'lajang') return false;
			
			return true;
		});
	});

	let showMobileForm = $state(false);
	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	// AG Grid
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	// Valid leave statuses only (matching PostgreSQL leave_status enum)
const LEAVE_STATUSES = ['pending', 'approved', 'rejected', 'cancelled', 'paid'] as const;

const statusColors: Record<string, string> = {
    pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
    approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
    rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
    cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
    paid: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
};

const statusLabels: Record<string, string> = {
    pending: 'Menunggu',
    approved: 'Disetujui',
    rejected: 'Ditolak',
    cancelled: 'Dibatalkan',
    paid: 'Dibayar'
};

function getStatusBadge(status: string) {
    return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${statusLabels[status] || status}</span>`;
}

	function iconView(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>';
	}
	function iconApprove(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>';
	}
	function iconReject(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>';
	}
	function iconCancel(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>';
	}

	function createActionButton(html: string, className: string, ariaLabel: string, onClick: () => void): HTMLButtonElement {
		const btn = document.createElement('button');
		btn.innerHTML = html;
		btn.className = className;
		btn.setAttribute('aria-label', ariaLabel);
		btn.onclick = (e) => { e.stopPropagation(); onClick(); };
		return btn;
	}

	const columnDefs: ColDef[] = [
		{
			field: 'employee_name', headerName: 'Karyawan', minWidth: 200, flex: 1,
			cellRenderer: (params: AgGridCellParams<LeaveRequest>) => {
				if (!params.value) return '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-sky-50 to-sky-100 flex items-center justify-center text-xs font-semibold text-sky-600 shrink-0 ring-1 ring-sky-200">${getInitials(params.value as string)}</div>
					<span class="text-sm font-medium text-gray-900">${params.value}</span>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'leave_type_name', headerName: 'Jenis Cuti', minWidth: 140,
			cellRenderer: (params: AgGridCellParams<LeaveRequest>) => {
				if (!params.value) return '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-indigo-50 text-indigo-700">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'start_date', headerName: 'Tanggal', minWidth: 140,
			cellRenderer: (params: AgGridCellParams<LeaveRequest>) => {
				if (!params.value) return '';
				const end = params.data?.end_date || '';
				return `<span class="text-sm text-gray-700">${formatDate(params.value as string)}${end && end !== params.value ? ` — ${formatDate(end)}` : ''}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm',
		},
		{
			field: 'total_days', headerName: 'Hari', minWidth: 80, maxWidth: 90,
			valueFormatter: (params: AgGridValueParams) => params.value != null ? `${params.value} hr` : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-700 text-center',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<LeaveRequest>) => {
				const status = (params.value as string) || '';
				return getStatusBadge(status);
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Diajukan', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 180, maxWidth: 180,
			cellRenderer: (params: AgGridCellParams<LeaveRequest>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('leave', 'approve')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-green-500 hover:text-green-700 hover:bg-green-50 transition cursor-pointer',
						'Setujui', () => handleApprove(item.id));
					container.appendChild(approveBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
				}

				if ((item.status === 'pending') && hasPermission('leave', 'create')) {
					const cancelBtn = createActionButton(iconCancel(),
						'p-1.5 rounded-lg text-gray-400 hover:text-orange-600 hover:bg-orange-50 transition cursor-pointer',
						'Batalkan', () => handleCancel(item.id));
					container.appendChild(cancelBtn);
				}

				return container;
			},
			sortable: false, filter: false, resizable: false,
		},
	];

	const gridOptions: GridOptions = {
		columnDefs, defaultColDef,
		rowHeight: 56, headerHeight: 44,
		animateRows: true, domLayout: 'autoHeight',
		suppressDragLeaveHidesColumns: true, suppressRowHoverHighlight: false,
		enableCellTextSelection: true, pagination: false, theme: 'legacy',
		onGridReady: (params) => { gridApi = params.api; },
	};

	$effect(() => {
		if (gridContainer && !showForm && !showDetail) {
			if (!gridApi && agGridModule) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			if (gridApi) { gridApi.updateGridOptions({ rowData: items }); }
		}
	});

	$effect(() => {
		if ((showForm || showDetail) && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
		loadTypes();
		loadMyBalances();
	});

	onDestroy(() => { gridApi?.destroy(); gridApi = null; });

	async function loadTypes() {
		try {
			const res = await api.getTypes() as ApiResponse<LeaveType[]>;
			leaveTypes = res.data || [];
		} catch { leaveTypes = []; }
	}

	async function loadMyBalances() {
		try {
			const res = await api.getMyBalances() as ApiResponse<LeaveBalance[]>;
			myBalances = res.data || [];
		} catch { myBalances = []; }
	}

	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true; errorMessage = '';
		try {
			const response = await api.list(page, perPage, statusFilter) as ApiResponse<LeaveRequest[]>;
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal memuat data'; }
		finally { isLoading = false; }
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	function calcTotalDays() {
		if (form.start_date && form.end_date) {
			const start = new Date(form.start_date);
			const end = new Date(form.end_date);
			const diff = Math.floor((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24)) + 1;
			form.total_days = diff > 0 ? diff : 1;
		}
	}

	function openCreateForm() {
		formTitle = 'Ajukan Cuti';
		form = { leave_type_id: '', start_date: '', end_date: '', total_days: 1, is_half_day: false, reason: '', document_url: '', contact_during_leave: '' };
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	let userLeaveValidation = $derived.by(() => {
		const user = auth.getUser() as Record<string, any> | null;
		const gender = user?.gender || '';
		const marital = user?.marital_status || '';
		
		// Build a map: leave_type_id -> { code, errorMessage }
		const validationMap = new Map<string, string>();
		for (const t of leaveTypes) {
			if (t.code === 'melahirkan' && gender !== 'perempuan') {
				validationMap.set(t.id, 'Cuti hamil/melahirkan hanya untuk karyawan perempuan');
			} else if (t.code === 'keguguran' && gender !== 'perempuan') {
				validationMap.set(t.id, 'Cuti keguguran hanya untuk karyawan perempuan');
			} else if (t.code === 'menikah' && marital !== 'lajang') {
				validationMap.set(t.id, 'Cuti menikah hanya untuk karyawan yang belum menikah (lajang)');
			}
		}
		return validationMap;
	});

	async function handleSave() {
		if (!form.leave_type_id) { formError = 'Jenis cuti harus dipilih'; return; }
		if (!form.start_date) { formError = 'Tanggal mulai harus diisi'; return; }
		if (!form.end_date) { formError = 'Tanggal selesai harus diisi'; return; }
		if (form.total_days <= 0) { formError = 'Jumlah hari harus lebih dari 0'; return; }
		if (!form.reason.trim()) { formError = 'Alasan cuti harus diisi'; return; }

		// Frontend validation for leave type eligibility
		const leaveError = userLeaveValidation.get(form.leave_type_id);
		if (leaveError) {
			formError = leaveError;
			return;
		}

		isSaving = true; formError = '';
		try {
			await api.create({
				leave_type_id: form.leave_type_id,
				start_date: form.start_date,
				end_date: form.end_date,
				total_days: form.total_days,
				is_half_day: form.is_half_day,
				reason: form.reason.trim(),
				contact_during_leave: form.contact_during_leave,
			});
			cancelForm();
			load();
			loadMyBalances();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true; isDetailLoading = true; detailData = null;
		try {
			const response = await api.get(id) as ApiResponse<LeaveRequest>;
			detailData = response.data ?? null;
		} catch { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailData = null; }

	async function handleApprove(id: string) {
		isSaving = true;
		try { await api.approve(id); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return; isSaving = true;
		try { await api.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; rejectId = null; rejectReason = ''; load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	async function handleCancel(id: string) {
		isSaving = true;
		try { await api.cancel(id, {}); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal membatalkan'; }
		finally { isSaving = false; }
	}

	function parseApprovalTrail(trail: string): any[] {
		try {
			const parsed = JSON.parse(trail);
			return Array.isArray(parsed) ? parsed : [];
		} catch {
			return [];
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function getInitials(name: string): string {
		const parts = name.split(' ');
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->

<!-- eslint-disable svelte/no-at-html-tags -->

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Cuti</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan dan kelola cuti karyawan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail}				{#if hasPermission('leave', 'create')}
					<button onclick={() => { openCreateForm(); if (window.innerWidth < 640) showMobileForm = true; }} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
						<span class="hidden sm:inline">Ajukan Cuti</span>
						<span class="sm:hidden">Ajukan</span>
					</button>
				{/if}
			{/if}
		</div>
	</div>

	{#if showBalances && filteredBalances.length > 0}
		<div class="bg-white border border-gray-200 rounded-xl p-5 mb-4 shadow-sm">
			<h3 class="text-sm font-semibold text-gray-800 mb-3">Sisa Cuti Saya</h3>
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
				{#each filteredBalances as b (b)}
					<div class="bg-gradient-to-br from-blue-50 to-indigo-50 rounded-lg p-3 border border-blue-100">
						<div class="text-xs text-gray-500 mb-1 truncate">{b.leave_type_name}</div>
						<div class="flex items-baseline gap-1">
							<span class="text-xl font-bold text-blue-700">{b.remaining}</span>
							<span class="text-xs text-gray-400">/ {b.total_quota}</span>
						</div>
						<div class="text-xs text-gray-400 mt-1">Terpakai: {b.used}</div>
					</div>
				{/each}
			</div>
		</div>
	{/if}

	{#if !showForm}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">Semua</button>
				{#each LEAVE_STATUSES as status (status)}
					{@const label = statusLabels[status] || status.charAt(0).toUpperCase() + status.slice(1)}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer tap-highlight-transparent {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-700 dark:text-gray-400 dark:hover:bg-gray-800'}">{label}</button>
				{/each}
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} pengajuan ditemukan` : ''}</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">{formTitle}</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">

					{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="leave-type" class="block text-sm font-medium text-gray-700 mb-1.5">Jenis Cuti <span class="text-red-500">*</span></label>
						<select id="leave-type" bind:value={form.leave_type_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="">-- Pilih Jenis Cuti --</option>
							{#each leaveTypes as t (t)}
								<option value={t.id}>{t.name} ({t.is_paid ? 'Bayar' : 'Tidak Bayar'})</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="leave-reason" class="block text-sm font-medium text-gray-700 mb-1.5">Alasan <span class="text-red-500">*</span></label>
						<input id="leave-reason" type="text" bind:value={form.reason} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Contoh: Cuti tahunan" />
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label for="leave-start" class="block text-sm font-medium text-gray-700 mb-1.5">Tanggal Mulai <span class="text-red-500">*</span></label>
						<input id="leave-start" type="date" bind:value={form.start_date} onchange={calcTotalDays} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
					<div>
						<label for="leave-end" class="block text-sm font-medium text-gray-700 mb-1.5">Tanggal Selesai <span class="text-red-500">*</span></label>
						<input id="leave-end" type="date" bind:value={form.end_date} onchange={calcTotalDays} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
					<div>
						<label for="leave-days" class="block text-sm font-medium text-gray-700 mb-1.5">Jumlah Hari</label>
						<input id="leave-days" type="number" bind:value={form.total_days} min="1" max="365" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
				</div>

				<div class="flex items-center gap-3">
					<label class="relative inline-flex items-center cursor-pointer">
						<input type="checkbox" bind:checked={form.is_half_day} class="sr-only peer" />
						<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[#1A56DB]/20 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#1A56DB]"></div>
						<span class="ms-2 text-sm font-medium text-gray-700">Cuti Setengah Hari</span>
					</label>
				</div>

				<div>
					<label for="leave-contact" class="block text-sm font-medium text-gray-700 mb-1.5">Kontak Selama Cuti</label>
					<input id="leave-contact" type="text" bind:value={form.contact_during_leave} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="No. HP yang bisa dihubungi (opsional)" />
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Cuti
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">Detail Cuti</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if isDetailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 rounded w-48"></div><div class="h-4 bg-gray-50 rounded w-64"></div><div class="h-4 bg-gray-50 rounded w-40"></div></div>
				{:else if detailData}
					{#if detailData.approval_trail && detailData.approval_trail !== '[]' && detailData.approval_trail !== ''}
						{@const trail = parseApprovalTrail(detailData.approval_trail)}
						<div class="bg-indigo-50/50 dark:bg-indigo-900/10 rounded-xl p-4 border border-indigo-100 dark:border-indigo-900/30 mb-6">
							<div class="flex items-center gap-2 mb-3">
								<svg class="w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								<h3 class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 uppercase tracking-wider">Progress Approval</h3>
							</div>
							<div class="space-y-2">
								{#each trail as step (step)}
									{@const isPending = step.status === 'pending'}
									{@const isApproved = step.status === 'approved'}
									{@const isRejected = step.status === 'rejected'}
									<div class="flex items-center gap-3">
										<div class="w-7 h-7 rounded-full flex items-center justify-center shrink-0 {isApproved ? 'bg-emerald-100 text-emerald-600' : isRejected ? 'bg-red-100 text-red-600' : 'bg-gray-100 dark:bg-gray-700 text-gray-400'}">
											{#if isApproved}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
											{:else if isRejected}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
											{:else}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium {isApproved ? 'text-emerald-700 dark:text-emerald-300' : isRejected ? 'text-red-700 dark:text-red-300' : 'text-gray-700 dark:text-gray-300'}">
												{step.approver_name || 'Approver'} 
												<span class="text-xs font-normal text-gray-400">Level {step.level || step.step}</span>
											</p>
											{#if step.note}
												<p class="text-xs text-gray-400 truncate">{step.note}{#if step.date} &middot; {step.date || '-'}{/if}</p>
											{/if}
										</div>
										<div>
											{#if isApproved}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200">Disetujui</span>
											{:else if isRejected}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-200">Ditolak</span>
											{:else}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-50 text-yellow-700 ring-1 ring-yellow-200 animate-pulse">Menunggu</span>
											{/if}
										</div>
									</div>
									{#if step !== trail[trail.length - 1]}
										<div class="ml-3.5 border-l-2 border-gray-200 dark:border-gray-700 h-3"></div>
									{/if}
								{/each}
							</div>
						</div>
					{/if}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Informasi Cuti</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Jenis Cuti</span><p class="text-sm font-medium text-gray-900">{detailData.leave_type_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(detailData.status)}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal</span><p class="text-sm font-medium text-gray-900">{formatDate(detailData.start_date)} — {formatDate(detailData.end_date)}</p></div>
								<div><span class="text-xs text-gray-400">Total Hari</span><p class="text-sm font-medium text-gray-900">{detailData.total_days} hari{detailData.is_half_day ? ' (setengah hari)' : ''}</p></div>
								<div><span class="text-xs text-gray-400">Alasan</span><p class="text-sm text-gray-700">{detailData.reason || '-'}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Karyawan & Kontak</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Pengaju</span><p class="text-sm font-medium text-gray-900">{detailData.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Kontak</span><p class="text-sm text-gray-700">{detailData.contact_during_leave || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Diajukan Pada</span><p class="text-sm text-gray-700">{formatDate(detailData.created_at)}</p></div>
							</div>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail cuti</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _, i (i)}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 rounded w-44"></div><div class="h-3 bg-gray-50 rounded w-28"></div></div><div class="h-6 bg-gray-100 rounded-full w-20"></div><div class="h-8 bg-gray-100 rounded w-24"></div></div>{/each}</div></div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada pengajuan cuti"
					description="Belum ada pengajuan cuti yang diajukan."
					actionLabel={hasPermission('leave', 'create') ? 'Ajukan Cuti' : ''}
					onAction={hasPermission('leave', 'create') ? openCreateForm : undefined}
				/>
			{:else}
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<PullToRefresh onRefresh={load}>
				<div class="md:hidden space-y-3">
					<StaggerList items={items}>
					{#snippet children(item)}
						<SwipeActions
							onApprove={item.status === 'pending' && hasPermission('leave', 'approve') ? () => handleApprove(item.id) : undefined}
							onReject={item.status === 'pending' && hasPermission('leave', 'approve') ? () => openReject(item.id) : undefined}
						>
						<MobileCard
							avatar={item.employee_name}
							avatarColor={getAvatarTheme('leave').gradientClasses}
							title={item.employee_name}
							subtitle={item.leave_type_name}
							badges={[{ label: item.status === 'pending' ? 'Menunggu' : item.status === 'approved' ? 'Disetujui' : item.status === 'rejected' ? 'Ditolak' : item.status === 'cancelled' ? 'Dibatalkan' : item.status, color: statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300' }]}
							onclick={() => openDetail(item.id)}
							clickable={true}
						>
							{#snippet children()}
								<div class="flex items-center gap-3 text-[11px] text-gray-500 dark:text-gray-400 mb-1">
									<div class="flex items-center gap-1.5 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded-md">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
										</svg>
										<span class="tabular-nums">{formatDate(item.start_date)}{item.end_date && item.end_date !== item.start_date ? ` - ${formatDate(item.end_date)}` : ''}</span>
									</div>
									<div class="flex items-center gap-1 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded-md">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
										</svg>
										<span>{item.total_days} hari</span>
									</div>
								</div>
							{/snippet}
							{#snippet footer()}
								{#if item.status === 'pending'}
									<div class="flex items-center gap-2 pt-2">
										<button
											onclick={(e) => { e.stopPropagation(); openDetail(item.id); }}
											class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95"
										>
											Detail
										</button>
										{#if hasPermission('leave', 'approve')}
											<button
												onclick={(e) => { e.stopPropagation(); handleApprove(item.id); }}
												class="flex-1 py-2 text-xs font-semibold text-green-700 dark:text-green-300 bg-green-50 dark:bg-green-900/30 rounded-lg hover:bg-green-100 dark:hover:bg-green-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
											>
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
												Setujui
											</button>
											<button
												onclick={(e) => { e.stopPropagation(); openReject(item.id); }}
												class="flex-1 py-2 text-xs font-semibold text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
											>
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
												Tolak
											</button>
										{/if}
										{#if hasPermission('leave', 'create')}
											<button
												onclick={(e) => { e.stopPropagation(); handleCancel(item.id); }}
												class="flex-1 py-2 text-xs font-medium text-orange-600 dark:text-orange-300 bg-orange-50 dark:bg-orange-900/30 rounded-lg hover:bg-orange-100 dark:hover:bg-orange-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
											>
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
												Batalkan
											</button>
										{/if}
									</div>
								{/if}
							{/snippet}
						</MobileCard>
						</SwipeActions>
					{/snippet}
					</StaggerList>
				</div>
				</PullToRefresh>
				<div class="flex items-center justify-between px-4 py-3 mt-1">
					<div class="text-xs text-gray-400">{(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari {total}</div>
					<div class="flex items-center gap-2">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer tap-highlight-transparent active:scale-95">
							Sebelumnya
						</button>
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer tap-highlight-transparent active:scale-95">
							Selanjutnya
						</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>	<!-- Mobile Bottom Sheet Form -->
	<BottomSheet bind:open={showMobileForm} title="Ajukan Cuti">
		{#snippet footer()}
			<div class="flex items-center gap-3">
				<button onclick={() => { showMobileForm = false; showForm = false; formError = ''; }} class="flex-1 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="flex-1 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan
				</button>
			</div>
		{/snippet}
		<div class="space-y-4">
			{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-3 py-2 rounded-lg">{formError}</div>{/if}
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Jenis Cuti <span class="text-red-500">*</span></label>
				<select bind:value={form.leave_type_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-800 dark:text-white">
					<option value="">-- Pilih --</option>
					{#each leaveTypes as t (t)}
						<option value={t.id}>{t.name} ({t.is_paid ? 'Bayar' : 'Tidak Bayar'})</option>
					{/each}
				</select>
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Tanggal <span class="text-red-500">*</span></label>
				<div class="grid grid-cols-2 gap-2">
					<input type="date" bind:value={form.start_date} onchange={calcTotalDays} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
					<input type="date" bind:value={form.end_date} onchange={calcTotalDays} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
				</div>
			</div>
			<div class="grid grid-cols-2 gap-2">
				<div>
					<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Jumlah Hari</label>
					<input type="number" bind:value={form.total_days} min="1" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
				</div>
				<div class="flex items-end pb-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={form.is_half_day} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]/30" />
						<span class="text-xs text-gray-600 dark:text-gray-400">Setengah Hari</span>
					</label>
				</div>
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Alasan <span class="text-red-500">*</span></label>
				<input type="text" bind:value={form.reason} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" placeholder="Contoh: Cuti tahunan" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Kontak (opsional)</label>
				<input type="text" bind:value={form.contact_during_leave} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" placeholder="No. HP" />
			</div>
		</div>
	</BottomSheet>

	<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<!-- svelte-ignore a11y_no_static_element_interactions -->
	<div onkeydown={(e) => { if (e.key === 'Escape') cancelReject(); }}
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} onkeydown={(e) => { if (e.key === 'Escape') cancelReject(); }}
			role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak cuti" class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 text-center mb-4">Tolak Pengajuan Cuti</h3>
				<div class="space-y-3">
					<label for="reject-reason" class="block text-sm font-medium text-gray-700">Alasan Penolakan</label>
					<textarea id="reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Tolak
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
