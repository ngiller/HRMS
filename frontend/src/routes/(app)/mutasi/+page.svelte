<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount, onDestroy } from 'svelte';
	import { mutations, employees, departments, positions, positionGrades } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import SwipeActions from '$lib/components/SwipeActions.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

	interface MutationItem {
		id: string;
		employee_id: string;
		employee_name: string;
		mutation_type: string;
		old_department_name: string;
		new_department_name: string;
		old_position_name: string;
		new_position_name: string;
		old_base_salary: number | null;
		new_base_salary: number | null;
		reason: string;
		notes: string;
		effective_date: string;
		status: string;
		approved_by_name: string;
		rejection_reason: string;
		created_at: string;
		approval_trail?: string;
	}

	interface FormData {
		employee_id: string;
		mutation_type: string;
		new_department_id: string;
		new_position_id: string;
		new_position_grade_id: string;
		new_employment_status: string;
		new_base_salary: number | null;
		reason: string;
		effective_date: string;
		notes: string;
	}

	let items = $state<MutationItem[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Form state
	let showForm = $state(false);
	let form = $state<FormData>({
		employee_id: '',
		mutation_type: 'promotion',
		new_department_id: '',
		new_position_id: '',
		new_position_grade_id: '',
		new_employment_status: '',
		new_base_salary: null,
		reason: '',
		effective_date: '',
		notes: '',
	});
	let formError = $state('');
	let isSaving = $state(false);

	// Dropdown data
	let employeeList: any[] = $state([]);
	let deptList: any[] = $state([]);
	let posList: any[] = $state([]);
	let gradeList: any[] = $state([]);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<MutationItem | null>(null);
	let isDetailLoading = $state(false);

	let processingId = $state<string | null>(null);
	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	const typeLabels: Record<string, string> = {
		promotion: 'Promosi',
		demotion: 'Demosi',
		transfer: 'Mutasi Departemen',
		position_change: 'Perubahan Jabatan',
		status_change: 'Perubahan Status',
		salary_change: 'Perubahan Gaji',
	};

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu',
		approved: 'Disetujui',
		rejected: 'Ditolak',
		cancelled: 'Dibatalkan',
	};

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-50 text-gray-600 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700',
	};

	function getStatusBadge(status: string): string {
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${statusLabels[status] || status}</span>`;
	}

	let filteredItems = $derived.by(() => {
		if (!searchQuery.trim()) return items;
		const q = searchQuery.toLowerCase();
		return items.filter(i =>
			i.employee_name?.toLowerCase().includes(q) ||
			typeLabels[i.mutation_type]?.toLowerCase().includes(q) ||
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
	function iconCancel(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>';
	}
	function iconSk(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>';
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
			cellRenderer: (params: AgGridCellParams<MutationItem>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-indigo-50 to-indigo-100 flex items-center justify-center text-xs font-semibold text-indigo-600 shrink-0 ring-1 ring-indigo-200">${initials}</div>
					<div class="text-sm font-medium text-gray-900">${params.value}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'mutation_type', headerName: 'Tipe', minWidth: 140,
			valueFormatter: (params: AgGridValueParams) => typeLabels[params.value as string] || (params.value as string) || '',
			cellRenderer: (params: AgGridCellParams<MutationItem>) => {
				if (!params.value) return '';
				const label = typeLabels[params.value as string] || params.value as string;
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-indigo-50 text-indigo-700 dark:bg-indigo-900/30 dark:text-indigo-300">${label}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'effective_date', headerName: 'Tanggal Berlaku', minWidth: 140,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<MutationItem>) => {
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
			field: 'id', headerName: '', minWidth: 180, maxWidth: 180,
			cellRenderer: (params: AgGridCellParams<MutationItem>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'approved') {
					const skBtn = createActionButton(iconSk(),
						'p-1.5 rounded-lg text-gray-400 hover:text-emerald-600 hover:bg-emerald-50 transition cursor-pointer',
						'Cetak SK', () => window.open(`/mutasi/${item.id}/sk`, '_blank'));
					container.appendChild(skBtn);
				}

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

				if (item.status === 'pending' && hasPermission('employee', 'create')) {
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
		isLoading = true; errorMessage = '';
		try {
			const res = await mutations.list(page, perPage, statusFilter) as ApiResponse<MutationItem[]>;
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

	async function openForm() {
		formError = '';
		form = {
			employee_id: '',
			mutation_type: 'promotion',
			new_department_id: '',
			new_position_id: '',
			new_position_grade_id: '',
			new_employment_status: '',
			new_base_salary: null,
			reason: '',
			effective_date: '',
			notes: '',
		};
		showForm = true;
		try {
			const [empRes, deptRes, posRes, gradeRes] = await Promise.all([
				employees.list(1, 200),
				departments.getAll(),
				positions.getAll(),
				positionGrades.getAll(),
			]);
			if (empRes.success) employeeList = empRes.data || [];
			if (deptRes.success) deptList = deptRes.data || [];
			if (posRes.success) posList = posRes.data || [];
			if (gradeRes.success) gradeList = gradeRes.data || [];
		} catch { /* Silently fail dropdowns */ }
	}

	function cancelForm() { showForm = false; formError = ''; }

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	async function handleSave() {
		if (!form.employee_id) { formError = 'Karyawan harus diisi'; return; }
		if (!form.reason.trim()) { formError = 'Alasan mutasi harus diisi'; return; }
		if (!form.effective_date) { formError = 'Tanggal berlaku harus diisi'; return; }
		if (!form.new_department_id && !form.new_position_id && !form.new_position_grade_id && !form.new_employment_status && form.new_base_salary === null) {
			formError = 'Minimal satu perubahan harus diisi'; return;
		}

		isSaving = true; formError = '';
		try {
			await mutations.create(form);
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
		try {
			const res = await mutations.get(id) as ApiResponse<MutationItem>;
			detailData = res.data ?? null;
		} catch { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; }

	async function handleApprove(id: string) {
		processingId = id;
		try { await mutations.approve(id); load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menyetujui'; }
		finally { processingId = null; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		processingId = rejectId;
		try { await mutations.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
		finally { processingId = null; }
	}

	async function handleCancel(id: string) {
		try { await mutations.cancel(id); load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal membatalkan'; }
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

	async function exportExcel() {
		try {
			const blob = await mutations.exportExcel(statusFilter);
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'riwayat-mutasi.xlsx';
			document.body.appendChild(a);
			a.click();
			a.remove();
			URL.revokeObjectURL(url);
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal export Excel';
		}
	}

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; }, 400);
	}

	let selectedEmployee = $derived(employeeList.find((emp: any) => emp.id === form.employee_id));
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->
<!-- eslint-disable svelte/no-at-html-tags -->
<!-- eslint-disable svelte/no-navigation-without-resolve -->

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight dark:text-white">Mutasi & Promosi</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola promosi, demosi, mutasi departemen, dan perubahan jabatan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail && hasPermission('employee', 'read')}
				<button onclick={exportExcel} class="inline-flex items-center gap-2 px-4 py-2.5 border border-gray-200 text-gray-700 rounded-xl text-sm font-semibold hover:bg-gray-50 transition-all cursor-pointer dark:border-gray-600 dark:text-gray-300 dark:hover:bg-gray-800">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
					Export Excel
				</button>
			{/if}
			{#if !showForm && !showDetail && hasPermission('employee', 'create')}
				<button onclick={openForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					<span class="hidden sm:inline">Buat Mutasi</span>
					<span class="sm:hidden">Buat</span>
				</button>
			{/if}
		</div>
	</div>

	{#if !showForm && !showDetail}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400 dark:hover:bg-gray-700'}">Semua</button>
				{#each ['pending', 'approved', 'rejected', 'cancelled'] as status (status)}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400 dark:hover:bg-gray-700'}">{statusLabels[status] || status}</button>
				{/each}
			</div>
			<div class="flex items-center gap-3">
				<div class="relative">
					<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
					<input type="search" value={searchQuery} placeholder="Cari mutasi..." oninput={onSearchInput} class="w-40 lg:w-56 pl-9 pr-3 py-1.5 bg-gray-50 dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg text-xs outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
				</div>
				<div class="text-xs text-gray-400 dark:text-gray-500">{total > 0 ? `${total} mutasi ditemukan` : ''}</div>
			</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Buat Mutasi Baru</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div class="md:col-span-2">
						<label for="mut-employee" class="block text-sm font-semibold text-gray-700 mb-1.5 dark:text-gray-300">Karyawan <span class="text-red-500">*</span></label>
						<select id="mut-employee" bind:value={form.employee_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							<option value="">-- Pilih Karyawan --</option>
							{#each employeeList as emp (emp.id)}
								<option value={emp.id}>{emp.full_name} ({emp.employee_id})</option>
							{/each}
						</select>
					</div>

					{#if selectedEmployee}
						<div class="md:col-span-2 bg-blue-50/50 dark:bg-blue-950/20 border border-blue-100 dark:border-blue-900/30 rounded-xl p-3.5 flex items-start gap-3 animate-in fade-in slide-in-from-top-1 duration-300">
							<svg class="w-4.5 h-4.5 text-blue-600 dark:text-blue-400 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 1 1 1.063 1.06l-.041.02a.75.75 0 1 1-1.063-1.06zm0-3.75.041-.02a.75.75 0 1 1 1.063 1.06l-.041.02a.75.75 0 1 1-1.063-1.06zM12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18z" /></svg>
							<div class="text-xs space-y-1.5 w-full">
								<p class="font-bold text-blue-950 dark:text-blue-300">Posisi & Jabatan Saat Ini:</p>
								<div class="grid grid-cols-2 md:grid-cols-4 gap-2 text-blue-900 dark:text-blue-400">
									<div><span class="text-gray-400 dark:text-gray-500 font-medium">Departemen:</span><br/><span class="font-semibold">{selectedEmployee.department_name || '-'}</span></div>
									<div><span class="text-gray-400 dark:text-gray-500 font-medium">Jabatan:</span><br/><span class="font-semibold">{selectedEmployee.position_name || '-'}</span></div>
									<div><span class="text-gray-400 dark:text-gray-500 font-medium">Status:</span><br/><span class="font-semibold capitalize">{selectedEmployee.employment_status || '-'}</span></div>
									<div><span class="text-gray-400 dark:text-gray-500 font-medium">Gaji Pokok:</span><br/><span class="font-semibold text-emerald-600 dark:text-emerald-400">{selectedEmployee.base_salary ? formatCurrency(selectedEmployee.base_salary) : '-'}</span></div>
								</div>
							</div>
						</div>
					{/if}
					<div>
						<label for="mut-type" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tipe Mutasi <span class="text-red-500">*</span></label>
						<select id="mut-type" bind:value={form.mutation_type} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							{#each Object.entries(typeLabels) as [key, label] (key)}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="mut-dept" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Departemen Baru</label>
						<select id="mut-dept" bind:value={form.new_department_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each deptList as dept (dept.id)}
								<option value={dept.id}>{dept.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="mut-pos" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jabatan Baru</label>
						<select id="mut-pos" bind:value={form.new_position_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each posList as pos (pos.id)}
								<option value={pos.id}>{pos.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="mut-grade" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Grade Baru</label>
						<select id="mut-grade" bind:value={form.new_position_grade_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each gradeList as grade (grade.id)}
								<option value={grade.id}>{grade.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="mut-status" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Status Kepegawaian Baru</label>
						<select id="mut-status" bind:value={form.new_employment_status} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							<option value="tetap">Tetap</option>
							<option value="kontrak">Kontrak</option>
							<option value="percobaan">Percobaan</option>
							<option value="harian">Harian</option>
						</select>
					</div>
					<div>
						<label for="mut-salary" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Gaji Pokok Baru</label>
						<div class="relative">
							<span class="absolute inset-y-0 left-0 pl-3 flex items-center text-sm text-gray-400">Rp</span>
							<input id="mut-salary" type="number" bind:value={form.new_base_salary} placeholder="Kosongkan jika tidak berubah" class="w-full pl-10 pr-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white" />
						</div>
					</div>
					<div>
						<label for="mut-date" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tanggal Berlaku <span class="text-red-500">*</span></label>
						<input id="mut-date" type="date" bind:value={form.effective_date} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 dark:text-white" />
					</div>
				</div>
				<div>
					<label for="mut-reason" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Alasan Mutasi <span class="text-red-500">*</span></label>
					<textarea id="mut-reason" bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white dark:bg-gray-700 dark:text-white" placeholder="Jelaskan alasan mutasi/promosi..."></textarea>
				</div>
				<div>
					<label for="mut-notes" class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Catatan (Opsional)</label>
					<textarea id="mut-notes" bind:value={form.notes} rows="2" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white dark:bg-gray-700 dark:text-white" placeholder="Catatan tambahan..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Mutasi
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Mutasi</h2>
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
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Mutasi</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-white">{dd.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Tipe Mutasi</span><p class="text-sm font-medium text-gray-900 dark:text-white">{typeLabels[dd.mutation_type] || dd.mutation_type}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Tanggal Berlaku</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(dd.effective_date)}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Status</span><p>{@html getStatusBadge(dd.status)}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Diajukan Pada</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(dd.created_at)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Perubahan</h3>
							<div class="space-y-3">
								{#if dd.old_department_name || dd.new_department_name}
									<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
										<span class="text-xs text-gray-400 dark:text-gray-500">Departemen</span>
										<p class="text-sm"><span class="text-gray-500 line-through dark:text-gray-400">{dd.old_department_name || '-'}</span> <span class="text-gray-300 dark:text-gray-600">→</span> <span class="font-medium text-gray-900 dark:text-white">{dd.new_department_name || '-'}</span></p>
									</div>
								{/if}
								{#if dd.old_position_name || dd.new_position_name}
									<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
										<span class="text-xs text-gray-400 dark:text-gray-500">Jabatan</span>
										<p class="text-sm"><span class="text-gray-500 line-through dark:text-gray-400">{dd.old_position_name || '-'}</span> <span class="text-gray-300 dark:text-gray-600">→</span> <span class="font-medium text-gray-900 dark:text-white">{dd.new_position_name || '-'}</span></p>
									</div>
								{/if}
								{#if dd.old_base_salary !== null || dd.new_base_salary !== null}
									<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
										<span class="text-xs text-gray-400 dark:text-gray-500">Gaji Pokok</span>
										<p class="text-sm"><span class="text-gray-500 line-through dark:text-gray-400">{dd.old_base_salary ? formatCurrency(dd.old_base_salary) : '-'}</span> <span class="text-gray-300 dark:text-gray-600">→</span> <span class="font-medium text-emerald-600 dark:text-emerald-400">{dd.new_base_salary ? formatCurrency(dd.new_base_salary) : '-'}</span></p>
									</div>
								{/if}
							</div>
						</div>
					</div>

					<div class="border-t border-gray-100 dark:border-gray-700 mt-5 pt-4">
						<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-2">Alasan</h3>
						<p class="text-sm text-gray-700 dark:text-gray-300">{dd.reason || '-'}</p>
						{#if dd.notes}
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mt-4 mb-2">Catatan</h3>
							<p class="text-sm text-gray-700 dark:text-gray-300">{dd.notes}</p>
						{/if}
						{#if dd.rejection_reason}
							<div class="mt-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
								<span class="text-xs text-red-500 dark:text-red-400 font-semibold">Alasan Penolakan</span>
								<p class="text-sm text-red-600 dark:text-red-400 mt-0.5">{dd.rejection_reason}</p>
							</div>
						{/if}
					</div>

					<!-- Action Buttons -->
					{#if dd.status === 'pending' && hasPermission('employee', 'update')}
						<div class="border-t border-gray-100 dark:border-gray-700 mt-5 pt-4">
							<div class="flex items-center gap-3">
								<button onclick={() => handleApprove(dd.id)} disabled={processingId !== null} class="px-5 py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition inline-flex items-center gap-2 disabled:opacity-50 cursor-pointer">
									{#if processingId === dd.id}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{:else}<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>{/if}
									Setujui & Terapkan
								</button>
								<button onclick={() => openReject(dd.id)} disabled={processingId !== null} class="px-5 py-2.5 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-lg text-sm font-semibold hover:bg-red-50 dark:hover:bg-red-900/20 transition disabled:opacity-50 cursor-pointer">Tolak</button>
								<button onclick={() => handleCancel(dd.id)} disabled={processingId !== null} class="px-5 py-2.5 border border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 rounded-lg text-sm font-semibold hover:bg-gray-50 dark:hover:bg-gray-700 transition disabled:opacity-50 cursor-pointer">Batalkan</button>
							</div>
						</div>
					{/if}

					{#if dd.status === 'approved'}
						<div class="border-t border-gray-100 dark:border-gray-700 mt-5 pt-4">
							<a href={`/mutasi/${dd.id}/sk`}
								class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer">
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
								Cetak SK Mutasi
							</a>
						</div>
					{/if}
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail mutasi</p>
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
					<EmptyState variant="empty" title="Belum ada mutasi" description="Belum ada riwayat mutasi atau promosi." />
					{#if hasPermission('employee', 'create')}
						<button onclick={openForm} class="mt-4 px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Buat Mutasi Baru</button>
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
							avatarColor={getAvatarTheme('mutation').gradientClasses}
							title={item.employee_name}
							subtitle={`${typeLabels[item.mutation_type] || item.mutation_type} • ${formatDate(item.effective_date)}`}
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
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak mutasi" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Mutasi</h3>
				<div class="space-y-3">
					<label for="mut-reject-reason" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea id="mut-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white dark:bg-gray-700 dark:text-white" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={processingId !== null} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if processingId === rejectId}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Tolak
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
