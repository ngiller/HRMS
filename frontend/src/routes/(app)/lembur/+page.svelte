<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { overtime as api, employees } from '$lib/api.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import SwipeActions from '$lib/components/SwipeActions.svelte';
	import BottomSheet from '$lib/components/BottomSheet.svelte';
import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme } from '$lib/avatar-theme.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type OvertimeRequest = {
		id: string;
		employee_id: string;
		employee_name: string;
		date: string;
		total_hours: number;
		overtime_type: string;
		reason: string;
		status: string;
		created_at: string;
		start_time: string;
		end_time: string;
		is_mandatory: boolean;
		overtime_pay: number | null;
		hourly_rate: number;
	};

	type FormData = {
		date: string;
		start_time: string;
		end_time: string;
		total_hours: number;
		overtime_type: string;
		reason: string;
		is_mandatory: boolean;
	};

	let items = $state<OvertimeRequest[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let form = $state<FormData>({
		date: '', start_time: '', end_time: '', total_hours: 0,
		overtime_type: 'weekday', reason: '', is_mandatory: false
	});
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<OvertimeRequest | null>(null);
	let isDetailLoading = $state(false);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	let showCancelRequestConfirm = $state(false);
	let cancelRequestId = $state<string | null>(null);

	let showCancelFormConfirm = $state(false);
	let showMobileForm = $state(false);

	const statusColors: Record<string, string> = {
    pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
    approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
    active: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
    completed: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
    rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
    defaulted: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
    cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
};

function getStatusBadge(status: string) {
    const labels: Record<string, string> = {
        pending: 'Menunggu',
        approved: 'Disetujui',
        active: 'Aktif',
        completed: 'Lunas',
        rejected: 'Ditolak',
        defaulted: 'Macet',
        cancelled: 'Dibatalkan'
    };
    return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${labels[status] || status}</span>`;
}

	const overtimeTypeLabels: Record<string, string> = {
		weekday: 'Hari Kerja', weekend: 'Akhir Pekan', holiday: 'Hari Libur',
	};

	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

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
			field: 'employee_name', headerName: 'Karyawan', minWidth: 220, flex: 1,
			cellRenderer: (params: AgGridCellParams<OvertimeRequest>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-amber-50 to-amber-100 flex items-center justify-center text-xs font-semibold text-amber-600 shrink-0 ring-1 ring-amber-200">${initials}</div>
					<div class="text-sm font-medium text-gray-900">${params.value}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'date', headerName: 'Tanggal', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'overtime_type', headerName: 'Tipe', minWidth: 120,
			cellRenderer: (params: AgGridCellParams<OvertimeRequest>) => {
				if (!params.value) return '';
				const label = overtimeTypeLabels[params.value as string] || params.value as string;
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-50 text-blue-700">${label}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'total_hours', headerName: 'Jam', minWidth: 80, maxWidth: 100,
			valueFormatter: (params: AgGridValueParams) => params.value != null ? `${params.value} jam` : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-700 tabular-nums',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<OvertimeRequest>) => {
				const status = (params.value as string) || '';
				return getStatusBadge(status);
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 140, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<OvertimeRequest>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('overtime', 'approve')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-green-500 hover:text-green-700 hover:bg-green-50 transition cursor-pointer',
						'Setujui', () => handleApprove(item.id));
					container.appendChild(approveBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
				}

				if (item.status === 'pending' && hasPermission('overtime', 'create')) {
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
		columnDefs,
		defaultColDef,
		rowHeight: 56,
		headerHeight: 44,
		animateRows: true,
		domLayout: 'autoHeight',
		suppressDragLeaveHidesColumns: true,
		suppressRowHoverHighlight: false,
		enableCellTextSelection: true,
		pagination: false,
		theme: 'legacy',
		onGridReady: (params) => { gridApi = params.api; },
	};

	$effect(() => {
		if (showForm || showDetail) {
			gridApi?.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (gridContainer && !showForm && !showDetail) {
			if (!gridApi && agGridModule) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			if (gridApi) { gridApi.updateGridOptions({ rowData: items }); }
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});
	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response = await api.list(page, perPage, statusFilter) as ApiResponse<OvertimeRequest[]>;
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal memuat data'; }
		finally { isLoading = false; }
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	function openCreateForm() {
		formTitle = 'Ajukan Lembur';
		const now = new Date();
		form = {
			date: '', start_time: '', end_time: '', total_hours: 0,
			overtime_type: 'weekday', reason: '', is_mandatory: false
		};
		formError = '';
		showForm = true;
	}

	function cancelForm() { 
		showCancelFormConfirm = true;
	}

	function confirmCancelForm() {
		showCancelFormConfirm = false;
		showForm = false; 
		formError = ''; 
	}

	function abortCancelForm() {
		showCancelFormConfirm = false;
	}

	function calcHours() {
		if (form.start_time && form.end_time) {
			const start = new Date(form.start_time);
			const end = new Date(form.end_time);
			if (end > start) {
				form.total_hours = parseFloat(((end.getTime() - start.getTime()) / 3600000).toFixed(1));
			}
		}
	}

	async function handleSave() {
		if (!form.date) { formError = 'Tanggal lembur harus diisi'; return; }
		if (!form.start_time) { formError = 'Waktu mulai harus diisi'; return; }
		if (!form.end_time) { formError = 'Waktu selesai harus diisi'; return; }
		if (form.total_hours <= 0) { formError = 'Total jam lembur harus lebih dari 0'; return; }
		if (!form.reason.trim()) { formError = 'Alasan lembur harus diisi'; return; }

		isSaving = true;
		formError = '';
		try {
			const startISO = new Date(form.start_time).toISOString();
			const endISO = new Date(form.end_time).toISOString();
			await api.create({
				date: form.date,
				start_time: startISO,
				end_time: endISO,
				total_hours: form.total_hours,
				overtime_type: form.overtime_type,
				reason: form.reason.trim(),
				is_mandatory: form.is_mandatory,
			});
			confirmCancelForm();
			load();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true;
		detailId = id;
		isDetailLoading = true;
		detailData = null;
		try {
			const response = await api.get(id) as ApiResponse<OvertimeRequest>;
			detailData = response.data ?? null;
		} catch (_) { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; }

	async function handleApprove(id: string) {
		isSaving = true;
		try { await api.approve(id); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		isSaving = true;
		try { await api.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; rejectId = null; rejectReason = ''; load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	function handleCancel(id: string) {
		cancelRequestId = id;
		showCancelRequestConfirm = true;
	}

	function abortCancelRequest() {
		showCancelRequestConfirm = false;
		cancelRequestId = null;
	}

	async function confirmCancelRequest() {
		if (!cancelRequestId) return;
		isSaving = true;
		try {
			await api.cancel(cancelRequestId);
			showCancelRequestConfirm = false;
			cancelRequestId = null;
			load();
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal membatalkan';
		} finally {
			isSaving = false;
		}
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

	function formatDateTime(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(amount);
	}

	function getInitials(name: string): string {
		const parts = name.split(' ');
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Lembur</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan dan kelola pengajuan lembur karyawan</p>
		</div>
		{#if !showForm && hasPermission('overtime', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer hidden sm:inline-flex">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Ajukan Lembur
			</button>
			<button onclick={() => { openCreateForm(); showMobileForm = true; }} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer sm:hidden">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Ajukan
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">Semua</button>
				{#each Object.keys(statusColors) as status}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{status.charAt(0).toUpperCase() + status.slice(1)}</button>
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
						<label for="overtime-date" class="block text-sm font-medium text-gray-700 mb-1.5">Tanggal <span class="text-red-500">*</span></label>
						<input id="overtime-date" type="date" bind:value={form.date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
					<div>
						<label for="overtime-type" class="block text-sm font-medium text-gray-700 mb-1.5">Tipe Lembur <span class="text-red-500">*</span></label>
						<select id="overtime-type" bind:value={form.overtime_type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="weekday">Hari Kerja (1.5x - 2x)</option>
							<option value="weekend">Akhir Pekan (2x - 3x)</option>
							<option value="holiday">Hari Libur (2x - 3x)</option>
						</select>
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="overtime-start" class="block text-sm font-medium text-gray-700 mb-1.5">Waktu Mulai <span class="text-red-500">*</span></label>
						<input id="overtime-start" type="datetime-local" bind:value={form.start_time} oninput={calcHours} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
					<div>
						<label for="overtime-end" class="block text-sm font-medium text-gray-700 mb-1.5">Waktu Selesai <span class="text-red-500">*</span></label>
						<input id="overtime-end" type="datetime-local" bind:value={form.end_time} oninput={calcHours} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="overtime-hours" class="block text-sm font-medium text-gray-700 mb-1.5">Total Jam</label>
						<div class="flex items-center gap-3">
							<input id="overtime-hours" type="number" step="0.5" min="0.5" max="24" bind:value={form.total_hours} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							<span class="text-sm text-gray-500">jam</span>
						</div>
					</div>
					<div>
						<label for="overtime-type-label" class="block text-sm font-medium text-gray-700 mb-1.5">Tipe</label>
						<label for="overtime-mandatory" class="flex items-center gap-2 mt-2.5">
							<input id="overtime-mandatory" type="checkbox" bind:checked={form.is_mandatory} class="w-4 h-4 rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]" />
							<span class="text-sm text-gray-600">Lembur wajib</span>
						</label>
					</div>
				</div>
				<div>
					<label for="overtime-reason" class="block text-sm font-medium text-gray-700 mb-1.5">Alasan <span class="text-red-500">*</span></label>
					<textarea id="overtime-reason" bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Jelaskan alasan lembur..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Lembur
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">Detail Lembur</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if isDetailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 rounded w-48"></div><div class="h-4 bg-gray-50 rounded w-64"></div><div class="h-4 bg-gray-50 rounded w-40"></div></div>
				{:else if detailData}
					{#if (detailData as any).approval_trail && (detailData as any).approval_trail !== '[]' && (detailData as any).approval_trail !== ''}
						{@const trail = parseApprovalTrail((detailData as any).approval_trail)}
						<div class="bg-indigo-50/50 dark:bg-indigo-900/10 rounded-xl p-4 border border-indigo-100 dark:border-indigo-900/30 mb-6">
							<div class="flex items-center gap-2 mb-3">
								<svg class="w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								<h3 class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 uppercase tracking-wider">Progress Approval</h3>
							</div>
							<div class="space-y-2">
								{#each trail as step}
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
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Informasi Pengajuan</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(detailData.status)}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal</span><p class="text-sm font-medium text-gray-900">{formatDate(detailData.date)}</p></div>
								<div><span class="text-xs text-gray-400">Tipe Lembur</span><p class="text-sm font-medium text-gray-900">{overtimeTypeLabels[detailData.overtime_type] || detailData.overtime_type}</p></div>
								<div><span class="text-xs text-gray-400">Waktu</span><p class="text-sm font-medium text-gray-700">{formatDateTime(detailData.start_time)} - {formatDateTime(detailData.end_time)}</p></div>
								<div><span class="text-xs text-gray-400">Total Jam</span><p class="text-sm font-medium text-gray-900">{detailData.total_hours} jam</p></div>
								<div><span class="text-xs text-gray-400">Lembur Wajib</span><p class="text-sm font-medium text-gray-700">{detailData.is_mandatory ? 'Ya' : 'Tidak'}</p></div>
								<div><span class="text-xs text-gray-400">Alasan</span><p class="text-sm font-medium text-gray-700">{detailData.reason || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Diajukan Pada</span><p class="text-sm font-medium text-gray-700">{formatDate(detailData.created_at)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Karyawan & Perhitungan</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Pengaju</span><p class="text-sm font-medium text-gray-900">{detailData.employee_name || '-'}</p></div>
								{#if detailData.overtime_pay != null && detailData.overtime_pay > 0}
									<div class="border-t border-gray-100 pt-3 mt-3">
										<span class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Perhitungan Lembur</span>
										<div class="mt-2 p-3 bg-blue-50 rounded-lg">
											<div class="flex items-center justify-between"><span class="text-xs text-gray-500">Rate/jam</span><span class="text-sm font-medium text-gray-900">{formatCurrency(detailData.hourly_rate || 0)}</span></div>
											<div class="flex items-center justify-between mt-1"><span class="text-xs text-gray-500">Total Jam</span><span class="text-sm font-medium text-gray-900">{detailData.total_hours} jam</span></div>
											<div class="border-t border-blue-200 mt-2 pt-2 flex items-center justify-between"><span class="text-sm font-semibold text-gray-700">Total Lembur</span><span class="text-sm font-bold text-blue-700">{formatCurrency(detailData.overtime_pay)}</span></div>
										</div>
									</div>
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail lembur</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 rounded w-44"></div><div class="h-3 bg-gray-50 rounded w-28"></div></div><div class="h-6 bg-gray-100 rounded-full w-20"></div><div class="h-8 bg-gray-100 rounded w-24"></div></div>{/each}</div></div>
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
					title="Belum ada pengajuan lembur"
					description="Belum ada pengajuan lembur yang diajukan."
				/>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<PullToRefresh onRefresh={load}>
				<div class="md:hidden space-y-3">
					{#each items as item}
						<SwipeActions
							onApprove={item.status === 'pending' && hasPermission('overtime', 'approve') ? () => handleApprove(item.id) : undefined}
							onReject={item.status === 'pending' && hasPermission('overtime', 'approve') ? () => openReject(item.id) : undefined}
						>
						<MobileCard
							avatar={item.employee_name}
							avatarColor={getAvatarTheme('overtime').gradientClasses}
							title={item.employee_name}
							subtitle={overtimeTypeLabels[item.overtime_type] || item.overtime_type}
							badges={[{ label: item.status === 'pending' ? 'Menunggu' : item.status === 'approved' ? 'Disetujui' : item.status === 'rejected' ? 'Ditolak' : item.status === 'cancelled' ? 'Dibatalkan' : item.status, color: statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300' }]}
							onclick={() => openDetail(item.id)}
							clickable={true}
						>
							{#snippet children()}
								<div class="flex items-center gap-2 text-[11px] text-gray-500 dark:text-gray-400 mb-1">
									<div class="flex items-center gap-1.5 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded-md">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
										</svg>
										<span class="tabular-nums">{formatDate(item.date)}</span>
									</div>
									<div class="flex items-center gap-1 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded-md">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
										</svg>
										<span>{item.total_hours} jam</span>
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
										{#if hasPermission('overtime', 'approve')}
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
										{#if hasPermission('overtime', 'create')}
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
					{/each}
				</div>
				</PullToRefresh>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30">
					<div class="text-xs text-gray-500">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>	<!-- Mobile Bottom Sheet Form -->
	<BottomSheet bind:open={showMobileForm} title="Ajukan Lembur">
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
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Tanggal <span class="text-red-500">*</span></label>
				<input type="date" bind:value={form.date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Tipe Lembur <span class="text-red-500">*</span></label>
				<select bind:value={form.overtime_type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white bg-white dark:bg-gray-800">
					<option value="weekday">Hari Kerja</option>
					<option value="weekend">Akhir Pekan</option>
					<option value="holiday">Hari Libur</option>
				</select>
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Waktu <span class="text-red-500">*</span></label>
				<div class="grid grid-cols-2 gap-2">
					<input type="datetime-local" bind:value={form.start_time} oninput={calcHours} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
					<input type="datetime-local" bind:value={form.end_time} oninput={calcHours} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
				</div>
			</div>
			<div class="grid grid-cols-2 gap-2">
				<div>
					<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Total Jam</label>
					<div class="flex items-center gap-2">
						<input type="number" step="0.5" min="0.5" bind:value={form.total_hours} class="flex-1 px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white" />
						<span class="text-xs text-gray-500">jam</span>
					</div>
				</div>
				<div class="flex items-end pb-2">
					<label class="flex items-center gap-2 cursor-pointer">
						<input type="checkbox" bind:checked={form.is_mandatory} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]/30" />
						<span class="text-xs text-gray-600 dark:text-gray-400">Wajib</span>
					</label>
				</div>
			</div>
			<div>
				<label class="block text-xs font-semibold text-gray-500 mb-1.5 uppercase tracking-wider">Alasan <span class="text-red-500">*</span></label>
				<textarea bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-800 dark:text-white resize-none" placeholder="Jelaskan alasan lembur..."></textarea>
			</div>
		</div>
	</BottomSheet>

	<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak pengajuan lembur" class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 text-center mb-4">Tolak Pengajuan Lembur</h3>
				<div class="space-y-3">
					<label for="lembur-reject-reason" class="block text-sm font-medium text-gray-700">Alasan Penolakan</label>
					<textarea id="lembur-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
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

{#if showCancelRequestConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={abortCancelRequest} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') abortCancelRequest(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Batalkan Pengajuan Lembur" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm transform transition-all">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-orange-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-orange-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Batalkan Pengajuan?</h3>
				<p class="text-sm text-gray-500 mb-6">Apakah Anda yakin ingin membatalkan pengajuan lembur ini? Aksi ini tidak dapat diubah kembali.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={abortCancelRequest} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Tidak, Kembali</button>
					<button onclick={confirmCancelRequest} disabled={isSaving} class="px-5 py-2.5 bg-orange-500 text-white rounded-lg text-sm font-semibold hover:bg-orange-600 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Ya, Batalkan
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

{#if showCancelFormConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={abortCancelForm} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') abortCancelForm(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Konfirmasi Batal" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm transform transition-all">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-amber-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Batalkan Pengisian?</h3>
				<p class="text-sm text-gray-500 mb-6">Semua data yang sudah Anda isi pada form pengajuan lembur ini akan hilang.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={abortCancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Lanjutkan Mengisi</button>
					<button onclick={confirmCancelForm} class="px-5 py-2.5 bg-amber-500 text-white rounded-lg text-sm font-semibold hover:bg-amber-600 transition inline-flex items-center gap-2 cursor-pointer">
						Ya, Batalkan
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
