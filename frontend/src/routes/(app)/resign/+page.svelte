<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount, onDestroy } from 'svelte';
	import { resign } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import SwipeActions from '$lib/components/SwipeActions.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

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
		approval_trail?: string;
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
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let form = $state({ last_working_date: '', reason: '', resign_type: 'voluntary' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<ResignItem | null>(null);
	let clearanceItems = $state<ClearanceItem[]>([]);
	let isDetailLoading = $state(false);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu',
		approved: 'Disetujui',
		rejected: 'Ditolak',
		processed: 'Diproses',
	};

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
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${statusLabels[status] || status}</span>`;
	}

	let filteredItems = $derived.by(() => {
		if (!searchQuery.trim()) return items;
		const q = searchQuery.toLowerCase();
		return items.filter(i =>
			i.employee_name?.toLowerCase().includes(q) ||
			typeLabels[i.resign_type]?.toLowerCase().includes(q) ||
			i.reason?.toLowerCase().includes(q)
		);
	});

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
			cellRenderer: (params: AgGridCellParams<ResignItem>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-rose-50 to-rose-100 flex items-center justify-center text-xs font-semibold text-rose-600 shrink-0 ring-1 ring-rose-200">${initials}</div>
					<div class="text-sm font-medium text-gray-900">${params.value}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'resign_type', headerName: 'Jenis', minWidth: 160,
			valueFormatter: (params: AgGridValueParams) => typeLabels[params.value as string] || (params.value as string) || '',
			cellRenderer: (params: AgGridCellParams<ResignItem>) => {
				if (!params.value) return '';
				const label = typeLabels[params.value as string] || params.value as string;
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-rose-50 text-rose-700 dark:bg-rose-900/30 dark:text-rose-300">${label}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'last_working_date', headerName: 'Terakhir Kerja', minWidth: 140,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<ResignItem>) => {
				return getStatusBadge((params.value as string) || '');
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
			cellRenderer: (params: AgGridCellParams<ResignItem>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('employee', 'update')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-green-500 hover:text-green-700 hover:bg-green-50 transition cursor-pointer',
						'Setujui', () => handleApprove(item.id));
					container.appendChild(approveBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
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
		isLoading = true; errorMessage = '';
		try {
			const res = await resign.list(page, perPage, statusFilter) as ApiResponse<ResignItem[]>;
			items = res.data || [];
			total = res.meta?.total || 0;
			page = res.meta?.page || 1;
			perPage = res.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
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

	async function openDetail(id: string) {
		showDetail = true;
		detailId = id;
		isDetailLoading = true;
		detailData = null;
		clearanceItems = [];
		try {
			const res = await resign.get(id) as ApiResponse<ResignItem>;
			detailData = res.data ?? null;
			if (res.data?.id) {
				const clearRes = await resign.listClearance(res.data.id);
				if (clearRes.success && clearRes.data) clearanceItems = clearRes.data as ClearanceItem[] || [];
			}
		} catch { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; clearanceItems = []; }

	async function handleApprove(id: string) {
		isSaving = true;
		try {
			await resign.approve(id);
			load();
		} catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		isSaving = true;
		try {
			await resign.reject(rejectId, { rejection_reason: rejectReason });
			showRejectModal = false;
			rejectId = null;
			rejectReason = '';
			load();
		} catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	async function toggleClearanceItem(item: ClearanceItem) {
		isSaving = true;
		try {
			await resign.updateClearance(item.id, { is_checked: !item.is_checked });
			if (detailData) {
				const id = detailData.id;
				const clearRes = await resign.listClearance(id);
				if (clearRes.success && clearRes.data) clearanceItems = clearRes.data as ClearanceItem[] || [];
			}
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal update clearance';
		} finally {
			isSaving = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function parseApprovalTrail(trail: string): any[] {
		try {
			const parsed = JSON.parse(trail);
			return Array.isArray(parsed) ? parsed : [];
		} catch {
			return [];
		}
	}

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; }, 400);
	}
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->
<!-- eslint-disable svelte/no-at-html-tags -->

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight dark:text-white">Resign & Exit</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola pengajuan resign dan exit clearance karyawan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail && hasPermission('employee', 'create')}
				<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					<span class="hidden sm:inline">Ajukan Resign</span>
					<span class="sm:hidden">Ajukan</span>
				</button>
			{/if}
		</div>
	</div>

	{#if !showForm && !showDetail}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400 dark:hover:bg-gray-700'}">Semua</button>
				{#each Object.entries(statusLabels) as [key, label] (key)}
					<button onclick={() => { statusFilter = key; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400 dark:hover:bg-gray-700'}">{label}</button>
				{/each}
			</div>
			<div class="flex items-center gap-3">
				<div class="relative">
					<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
					<input type="search" value={searchQuery} placeholder="Cari resign..." oninput={onSearchInput} class="w-40 lg:w-56 pl-9 pr-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-xs outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
				</div>
				<div class="text-xs text-gray-400 dark:text-gray-500">{total > 0 ? `${total} pengajuan ditemukan` : ''}</div>
			</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Ajukan Resign</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="resign-type" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jenis Resign</label>
						<select id="resign-type" bind:value={form.resign_type} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							{#each Object.entries(typeLabels) as [key, label] (key)}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="resign-date" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tanggal Terakhir Kerja <span class="text-red-500">*</span></label>
						<input id="resign-date" type="date" bind:value={form.last_working_date} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white" />
					</div>
				</div>
				<div>
					<label for="resign-reason" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Alasan Resign <span class="text-red-500">*</span></label>
					<textarea id="resign-reason" bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white dark:bg-gray-700 dark:text-white" placeholder="Jelaskan alasan resign..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Resign
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Resign</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if isDetailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-48"></div><div class="h-4 bg-gray-50 dark:bg-gray-600 rounded w-64"></div><div class="h-4 bg-gray-50 dark:bg-gray-600 rounded w-40"></div></div>
				{:else if detailData}
					{@const dd = detailData}
					<!-- Approval Trail -->
					{#if (detailData as any).approval_trail && (detailData as any).approval_trail !== '[]' && (detailData as any).approval_trail !== ''}
						{@const trail = parseApprovalTrail((detailData as any).approval_trail)}
						<div class="bg-indigo-50/50 dark:bg-indigo-900/10 rounded-xl p-4 border border-indigo-100 dark:border-indigo-900/30 mb-6">
							<div class="flex items-center gap-2 mb-3">
								<svg class="w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								<h3 class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 uppercase tracking-wider">Progress Approval</h3>
							</div>
							<div class="space-y-2">
								{#each trail as step (step)}
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
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900/30 dark:text-emerald-300 dark:ring-emerald-800">Disetujui</span>
											{:else if isRejected}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-200 dark:bg-red-900/30 dark:text-red-300 dark:ring-red-800">Ditolak</span>
											{:else}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-50 text-yellow-700 ring-1 ring-yellow-200 animate-pulse dark:bg-yellow-900/30 dark:text-yellow-300 dark:ring-yellow-800">Menunggu</span>
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
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Resign</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-white">{dd.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Jenis</span><p class="text-sm text-gray-700 dark:text-gray-300">{typeLabels[dd.resign_type] || dd.resign_type}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Status</span><p>{@html getStatusBadge(dd.status)}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Tanggal Terakhir Kerja</span><p class="text-sm font-medium text-gray-900 dark:text-white">{formatDate(dd.last_working_date)}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Diajukan Pada</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(dd.created_at)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Exit Clearance Checklist</h3>
							{#if clearanceItems.length > 0}
								<div class="space-y-2">
									{#each clearanceItems as item (item)}
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
								<p class="text-sm text-gray-500 italic dark:text-gray-400">Belum ada item clearance</p>
							{/if}
						</div>
					</div>

					<div class="border-t border-gray-100 dark:border-gray-700 mt-5 pt-4">
						<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">Alasan</h3>
						<p class="text-sm text-gray-700 dark:text-gray-300">{dd.reason || '-'}</p>
						{#if dd.rejection_reason}
							<div class="mt-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
								<span class="text-xs text-red-500 dark:text-red-400 font-semibold">Alasan Penolakan</span>
								<p class="text-sm text-red-600 dark:text-red-400 mt-0.5">{dd.rejection_reason}</p>
							</div>
						{/if}
					</div>

					{#if dd.status === 'pending' && hasPermission('employee', 'update')}
						<div class="border-t border-gray-100 dark:border-gray-700 mt-5 pt-4">
							<div class="flex items-center gap-3">
								<button onclick={() => handleApprove(dd.id)} disabled={isSaving || clearanceItems.some(i => !i.is_checked)} class="px-5 py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer" title={clearanceItems.some(i => !i.is_checked) ? 'Selesaikan semua clearance terlebih dahulu' : ''}>
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
									Setujui & Proses Resign
								</button>
								<button onclick={() => openReject(dd.id)} disabled={isSaving} class="px-5 py-2.5 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-lg text-sm font-semibold hover:bg-red-50 dark:hover:bg-red-900/20 transition disabled:opacity-50 cursor-pointer">
									Tolak
								</button>
							</div>
							{#if clearanceItems.some(i => !i.is_checked)}
								<p class="text-xs text-amber-600 dark:text-amber-400 mt-2">Semua item exit clearance harus dicentang sebelum resign dapat disetujui.</p>
							{/if}
						</div>
					{/if}
				{:else}
					<p class="text-sm text-gray-500 text-center py-8 dark:text-gray-400">Gagal memuat detail</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _, i (i)}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-44"></div><div class="h-3 bg-gray-50 dark:bg-gray-600 rounded w-28"></div></div><div class="h-6 bg-gray-100 dark:bg-gray-700 rounded-full w-20"></div><div class="h-8 bg-gray-100 dark:bg-gray-700 rounded w-24"></div></div>{/each}</div></div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<div class="py-8 text-center">
					<EmptyState variant="empty" title="Belum ada pengajuan resign" description="Belum ada pengajuan resign." />
					{#if hasPermission('employee', 'create')}
						<button onclick={openCreateForm} class="mt-4 px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Ajukan Resign</button>
					{/if}
				</div>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<PullToRefresh onRefresh={load}>
				<div class="md:hidden space-y-3 p-3">
					{#each filteredItems as item (item)}
						<SwipeActions
							onApprove={item.status === 'pending' && hasPermission('employee', 'update') ? () => handleApprove(item.id) : undefined}
							onReject={item.status === 'pending' && hasPermission('employee', 'update') ? () => openReject(item.id) : undefined}
						>
						<MobileCard
							avatar={getInitials(item.employee_name || '-')}
							avatarColor={getAvatarTheme('resign').gradientClasses}
							title={item.employee_name || 'Saya'}
							subtitle={`${typeLabels[item.resign_type] || item.resign_type} • ${formatDate(item.last_working_date)}`}
							badges={[{ label: statusLabels[item.status] || item.status, color: statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300' }]}
							onclick={() => openDetail(item.id)}
							clickable={true}
						>
							{#snippet children()}
								<p class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mb-1">{item.reason}</p>
							{/snippet}
							{#snippet footer()}
								<div class="flex items-center gap-2 pt-2">
									<button
										onclick={(e) => { e.stopPropagation(); openDetail(item.id); }}
										class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95"
									>
										Detail
									</button>
									{#if item.status === 'pending' && hasPermission('employee', 'update')}
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
								</div>
							{/snippet}
						</MobileCard>
						</SwipeActions>
					{/each}
				</div>
				</PullToRefresh>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-700 bg-gray-50/30 dark:bg-gray-800/30">
					<div class="text-xs text-gray-500 dark:text-gray-400">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Reject Modal -->
<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak pengajuan resign" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Pengajuan Resign</h3>
				<div class="space-y-3">
					<label for="resign-reject-reason" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea id="resign-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white dark:bg-gray-700 dark:text-white" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Tolak
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
