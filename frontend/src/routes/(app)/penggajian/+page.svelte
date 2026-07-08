<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { payroll as payrollApi } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type PayrollPeriod = {
		id: string;
		month: number;
		year: number;
		period_name: string;
		start_date: string;
		end_date: string;
		status: string;
		total_employee: number;
		total_gross: number;
		total_net: number;
		approved_by_name: string;
		paid_by_name: string;
		created_at: string;
	};

	let periods = $state<PayrollPeriod[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let isLoading = $state(true);
	let errorMessage = $state('');

	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true,
		resizable: true,
		filter: true,
		floatingFilter: false,
	};

	// Create form
	let showForm = $state(false);
	let formData = $state({ month: new Date().getMonth() + 1, year: new Date().getFullYear(), period_name: '', start_date: '', end_date: '' });
	let formError = $state('');
	let isSaving = $state(false);

	// Confirm
	let showConfirm = $state(false);
	let confirmAction = $state<'calculate' | 'approve' | 'pay'>('calculate');
	let confirmPeriodId = $state('');
	let confirmPeriodName = $state('');
	async function loadPeriods() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const res: ApiResponse<PayrollPeriod[]> = await payrollApi.listPeriods(currentPage, perPage) as ApiResponse<PayrollPeriod[]>;
			periods = res.data || [];
			total = res.meta?.total || 0;
			currentPage = res.meta?.page || 1;
			perPage = res.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memuat data penggajian';
		} finally {
			isLoading = false;
		}
	}

	function openCreateForm() {
		const now = new Date();
		formData = { month: now.getMonth() + 1, year: now.getFullYear(), period_name: '', start_date: '', end_date: '' };
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		showForm = false;
		formError = '';
	}

	async function handleCreate() {
		if (!formData.period_name.trim()) { formError = 'Nama periode harus diisi'; return; }
		isSaving = true;
		formError = '';
		try {
			await payrollApi.createPeriod(formData);
			cancelForm();
			loadPeriods();
		} catch (err: unknown) {
			formError = (err as { message?: string }).message || 'Gagal membuat periode';
		} finally {
			isSaving = false;
		}
	}

	function openConfirm(action: 'calculate' | 'approve' | 'pay', period: PayrollPeriod) {
		confirmAction = action;
		confirmPeriodId = period.id;
		confirmPeriodName = period.period_name;
		showConfirm = true;
	}

	function cancelConfirm() { showConfirm = false; }

	async function handleConfirm() {
		isSaving = true;
		try {
			if (confirmAction === 'calculate') {
				await payrollApi.calculatePayroll(confirmPeriodId);
			} else if (confirmAction === 'approve') {
				await payrollApi.approvePeriod(confirmPeriodId);
			} else if (confirmAction === 'pay') {
				await payrollApi.payPeriod(confirmPeriodId);
			}
			showConfirm = false;
			loadPeriods();
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memproses';
			showConfirm = false;
		} finally {
			isSaving = false;
		}
	}

	function viewDetail(id: string) {
		goto(`/penggajian/${id}`);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function monthName(m: number): string {
		const names = ['', 'Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
			'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];
		return names[m] || '';
	}

	// SVG icon helpers
	function iconView(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" /></svg>';
	}
	function iconCalculate(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z" /></svg>';
	}
	function iconApprove(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>';
	}
	function iconPay(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>';
	}

	function createActionButton(html: string, className: string, ariaLabel: string, onClick: () => void): HTMLButtonElement {
		const btn = document.createElement('button');
		btn.innerHTML = html;
		btn.className = className;
		btn.setAttribute('aria-label', ariaLabel);
		btn.onclick = (e) => { e.stopPropagation(); onClick(); };
		return btn;
	}

	// ── AG Grid Column Definitions ──
	const columnDefs: ColDef[] = [
		{
			field: 'period_name', headerName: 'Periode', minWidth: 200, flex: 1,
			cellRenderer: (params: AgGridCellParams<PayrollPeriod>) => {
				if (!params.value) return '';
				const initials = (params.value as string).substring(0, 2).toUpperCase();
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-blue-50 to-blue-100 dark:from-blue-900/50 dark:to-blue-800/50 flex items-center justify-center text-xs font-semibold text-blue-600 dark:text-blue-300 shrink-0 ring-1 ring-blue-200 dark:ring-blue-800">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900 dark:text-white">${params.value}</div><div class="text-xs text-gray-400 dark:text-gray-500">${monthName((params.data as PayrollPeriod)?.month)} ${(params.data as PayrollPeriod)?.year || ''}</div></div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'start_date', headerName: 'Tanggal', minWidth: 200,
			valueFormatter: (params: AgGridValueParams) => {
				const d = params.data as PayrollPeriod;
				return d?.start_date && d?.end_date ? `${formatDate(d.start_date)} — ${formatDate(d.end_date)}` : '-';
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'total_employee', headerName: 'Karyawan', minWidth: 100, maxWidth: 120,
			valueFormatter: (params: AgGridValueParams) => params.value != null ? String(params.value) : '',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-900 dark:text-white text-center tabular-nums',
		},
		{
			field: 'total_gross', headerName: 'Total Gross', minWidth: 140, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => (params.value as number) > 0 ? formatCurrency(params.value as number) : '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-900 dark:text-white text-right tabular-nums',
		},
		{
			field: 'total_net', headerName: 'Total Net', minWidth: 140, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => (params.value as number) > 0 ? formatCurrency(params.value as number) : '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm font-semibold text-gray-900 dark:text-white text-right tabular-nums',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 150,
			cellRenderer: (params: AgGridCellParams<PayrollPeriod>) => {
				const status = (params.value as string) || '';
				return `<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ring-1 ring-inset capitalize ${getStatusColor(status)}">${statusLabel(status)}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
		},
		{
			field: 'approved_by_name', headerName: 'Disetujui', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => (params.value as string) || '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'id', headerName: '', minWidth: 140, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<PayrollPeriod>) => {
				const p = params.data;
				if (!p) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				if (p.status === 'draft' && hasPermission('payroll', 'update')) {
					const calcBtn = createActionButton(iconCalculate(),
						'p-1.5 rounded-lg text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition cursor-pointer',
						'Hitung Gaji', () => openConfirm('calculate', p));
					container.appendChild(calcBtn);
				}
				if (p.status === 'calculated' && hasPermission('payroll', 'update')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-emerald-600 hover:bg-emerald-50 dark:hover:bg-emerald-900/30 transition cursor-pointer',
						'Setujui', () => openConfirm('approve', p));
					container.appendChild(approveBtn);
				}
				if (p.status === 'approved' && hasPermission('payroll', 'update')) {
					const payBtn = createActionButton(iconPay(),
						'p-1.5 rounded-lg text-purple-600 hover:bg-purple-50 dark:hover:bg-purple-900/30 transition cursor-pointer',
						'Bayarkan', () => openConfirm('pay', p));
					container.appendChild(payBtn);
				}

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-[#1A56DB] hover:bg-blue-50 dark:hover:bg-blue-900/30 transition cursor-pointer',
					'Detail', () => viewDetail(p.id));
				container.appendChild(viewBtn);

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
		if (showForm && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (periods.length > 0 && gridContainer && agGridModule && !showForm) {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: periods });
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		loadPeriods();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function getStatusColor(status: string): string {
		const map: Record<string, string> = {
			draft: 'bg-gray-50 text-gray-700 ring-gray-600/20 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-500',
			calculated: 'bg-blue-50 text-blue-700 ring-blue-600/20 dark:bg-blue-900/30 dark:text-blue-200 dark:ring-blue-800',
			approved: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20 dark:bg-emerald-900/30 dark:text-emerald-200 dark:ring-emerald-800',
			paid: 'bg-purple-50 text-purple-700 ring-purple-600/20 dark:bg-purple-900/30 dark:text-purple-200 dark:ring-purple-800',
		};
		return map[status] || 'bg-gray-50 text-gray-700 ring-gray-600/20 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-500';
	}

	function statusLabel(status: string): string {
		const map: Record<string, string> = {
			draft: 'Draft',
			calculated: 'Telah Dihitung',
			approved: 'Disetujui',
			paid: 'Dibayarkan',
		};
		return map[status] || status;
	}
</script>

<div class="w-full">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Payroll & Penggajian</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola periode penggajian dan hitung gaji karyawan</p>
		</div>
		{#if hasPermission('payroll', 'create')}
			<button onclick={openCreateForm}
				class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Periode Baru
			</button>
		{/if}
	</div>

	<!-- Create Form -->
	{#if showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl p-5 mb-6 shadow-sm">
			<h2 class="text-sm font-semibold text-gray-900 dark:text-white mb-4">Buat Periode Penggajian Baru</h2>
			{#if formError}
				<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm px-4 py-2.5 rounded-lg mb-4">{formError}</div>
			{/if}
			<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-4">
				<div>
					<label for="payroll-period-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Periode <span class="text-red-500">*</span></label>
					<input id="payroll-period-name" type="text" bind:value={formData.period_name}
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition"
						placeholder="Contoh: Juni 2026 atau Mg-1 Jun 26" />
				</div>
				<div>
					<label for="payroll-month" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Bulan Acuan</label>
					<select id="payroll-month" bind:value={formData.month}
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
						{#each Array.from({ length: 12 }, (_, i) => i + 1) as m}
							<option value={m}>{monthName(m)}</option>
						{/each}
					</select>
				</div>
				<div>
					<label for="payroll-year" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tahun</label>
					<input id="payroll-year" type="number" min="2020" max="2099" bind:value={formData.year}
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
				</div>
			</div>
			
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-4">
				<div>
					<label for="payroll-start-date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal Mulai (Opsional)</label>
					<input id="payroll-start-date" type="date" bind:value={formData.start_date}
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">Kosongkan untuk menghitung dari awal bulan</p>
				</div>
				<div>
					<label for="payroll-end-date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal Selesai (Opsional)</label>
					<input id="payroll-end-date" type="date" bind:value={formData.end_date}
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					<p class="text-xs text-gray-400 dark:text-gray-500 mt-1">Kosongkan untuk menghitung sampai akhir bulan</p>
				</div>
			</div>
			<div class="flex items-center justify-end gap-2">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleCreate} disabled={isSaving}
					class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Buat Periode
				</button>
			</div>
		</div>
	{/if}

	<!-- Summary Stats -->		<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6">
		<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5 shadow-sm hover:shadow-md transition-all duration-200">
			<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Periode</span>
			<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{isLoading ? '-' : total}</p>
		</div>
		<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5 shadow-sm hover:shadow-md transition-all duration-200">
			<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Draft</span>
			<p class="text-xl font-bold text-gray-400 dark:text-gray-500 mt-1 tabular-nums">{isLoading ? '-' : periods.filter(p => p.status === 'draft').length}</p>
		</div>
		<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5 shadow-sm hover:shadow-md transition-all duration-200">
			<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Telah Dihitung</span>
			<p class="text-xl font-bold text-blue-600 mt-1 tabular-nums">{isLoading ? '-' : periods.filter(p => p.status === 'calculated').length}</p>
		</div>
		<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5 shadow-sm hover:shadow-md transition-all duration-200">
			<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Dibayarkan</span>
			<p class="text-xl font-bold text-purple-600 mt-1 tabular-nums">{isLoading ? '-' : periods.filter(p => p.status === 'paid').length}</p>
		</div>
	</div>

	<!-- Periods Table -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
		{#if isLoading}
			<div class="p-6"><PulseLoader variant="table-row" count={3} /></div>
		{:else if errorMessage}
			<div class="py-16 text-center">
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
				<button onclick={loadPeriods} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
			</div>
		{:else if periods.length === 0}
			<div class="py-16 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
					<svg class="w-7 h-7 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" />
					</svg>
				</div>
				<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Belum ada periode penggajian</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 max-w-xs mx-auto">Buat periode penggajian baru untuk mulai menghitung gaji karyawan.</p>
			</div>
		{:else}
			<!-- Desktop Table — AG Grid -->
			<div class="hidden md:block">
				<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
			</div>

			<PullToRefresh onRefresh={loadPeriods}>
			<div class="md:hidden space-y-3">
				{#each periods as p}
					<MobileCard
						title={p.period_name}
						subtitle={`${monthName(p.month)} ${p.year}`}
						avatar={getInitials(p.period_name)}
						avatarColor={getAvatarTheme('payroll').gradientClasses}
						badges={[{ label: statusLabel(p.status), color: getStatusColor(p.status) }]}
						onclick={() => viewDetail(p.id)}
					>
						{#snippet children()}
							<div class="flex items-center gap-3 text-xs text-gray-400 dark:text-gray-500">
								<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
								</svg>
								{formatDate(p.start_date)} — {formatDate(p.end_date)}
							</div>
						{/snippet}
						{#snippet footer()}
							<div class="flex items-center justify-between">
								<span class="text-xs text-gray-500 dark:text-gray-400">{p.total_employee} karyawan</span>
								{#if p.total_net > 0}
									<span class="text-xs font-semibold text-gray-900 dark:text-white">{formatCurrency(p.total_net)}</span>
								{/if}
							</div>
						{/snippet}
					</MobileCard>
				{/each}
			</div>
			</PullToRefresh>

			<!-- Pagination -->
			<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-800 bg-gray-50/30 dark:bg-gray-900/30">
				<div class="text-xs text-gray-500 dark:text-gray-400">{total} periode</div>
				<div class="flex items-center gap-1.5">
					<button onclick={() => { currentPage--; loadPeriods(); }} disabled={currentPage <= 1}
						class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
					<span class="text-xs text-gray-400 dark:text-gray-500 px-2">{currentPage}/{totalPages}</span>
					<button onclick={() => { currentPage++; loadPeriods(); }} disabled={currentPage >= totalPages}
						class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Confirm Modal -->
{#if showConfirm}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelConfirm} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelConfirm(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Konfirmasi" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-blue-50 dark:bg-blue-900/20 flex items-center justify-center">
					<svg class="w-7 h-7 text-blue-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
					</svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">
					{confirmAction === 'calculate' ? 'Hitung Gaji' : confirmAction === 'approve' ? 'Setujui Periode' : 'Bayarkan Periode'}
				</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">
					{confirmAction === 'calculate' ? `Hitung gaji untuk periode "${confirmPeriodName}"?` :
					 confirmAction === 'approve' ? `Setujui periode "${confirmPeriodName}"? Tindakan ini tidak bisa dibatalkan.` :
					 `Bayarkan periode "${confirmPeriodName}"?`}
				</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelConfirm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleConfirm} disabled={isSaving}
						class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						{confirmAction === 'calculate' ? 'Ya, Hitung' : confirmAction === 'approve' ? 'Ya, Setujui' : 'Ya, Bayarkan'}
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
