<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { page as appPage } from '$app/stores';
	import { goto } from '$app/navigation';
	import { payroll as payrollApi } from '$lib/api.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type PayrollItem = {
		id: string;
		payroll_period_id: string;
		employee_id: string;
		employee_name: string;
		employee_id_code: string;
		department_name: string;
		position_name: string;
		employment_status: string;
		base_salary: number;
		daily_wage: number;
		total_days_worked: number;
		allowances: { name: string; amount: number }[];
		overtime_pay: number;
		overtime_hours: number;
		thr_amount: number;
		bonus_amount: number;
		gross_salary: number;
		deductions: { name: string; amount: number }[];
		pph21_amount: number;
		bpjs_kesehatan: number;
		bpjs_jht: number;
		bpjs_jp: number;
		loan_deduction: number;
		other_deductions: number;
		total_deductions: number;
		net_salary: number;
		company_cost: { name: string; amount: number }[];
		status: string;
		notes: string;
	};

	type Period = {
		id: string;
		month: number;
		year: number;
		period_name: string;
		start_date: string;
		end_date: string;
		status: string;
		total_employee: number;
		total_gross: number;
		total_deductions: number;
		total_net: number;
		total_company_cost: number;
		approved_by: string;
		approved_at: string;
		paid_by: string;
		paid_at: string;
	};

	let periodId = $derived($appPage.params.id as string);
	let period = $state<Period | null>(null);
	let items = $state<PayrollItem[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let perPage = $state(50);
	let totalPages = $state(0);
	let isLoading = $state(true);
	let errorMessage = $state('');
	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	function iconView(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" /></svg>';
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
			cellRenderer: (params: AgGridCellParams<PayrollItem>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const dept = params.data?.department_name || '';
				const pos = params.data?.position_name || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xs font-semibold text-gray-600 shrink-0 ring-1 ring-gray-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div><div class="text-xs text-gray-400">${dept} • ${pos}</div></div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'base_salary', headerName: 'Gaji Pokok', minWidth: 130, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => (params.value as number) > 0 ? formatCurrency(params.value as number) : '-',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-900 text-right tabular-nums',
		},
		{
			field: 'allowances', headerName: 'Tunjangan', minWidth: 130, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => {
				if (!params.value || !Array.isArray(params.value) || params.value.length === 0) return '-';
				const total = (params.value as { name: string; amount: number }[]).reduce((s: number, a: { name: string; amount: number }) => s + a.amount, 0);
				return formatCurrency(total);
			},
			cellClass: 'text-sm text-emerald-600 text-right tabular-nums',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'overtime_pay', headerName: 'Lembur', minWidth: 120, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => (params.value as number) > 0 ? formatCurrency(params.value as number) : '-',
			cellClass: 'text-sm text-amber-600 text-right tabular-nums',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'gross_salary', headerName: 'Gross', minWidth: 130, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => formatCurrency(params.value as number),
			cellClass: 'text-sm font-medium text-gray-900 text-right tabular-nums',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'total_deductions', headerName: 'Potongan', minWidth: 130, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => formatCurrency(params.value as number),
			cellClass: 'text-sm text-red-600 text-right tabular-nums',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'net_salary', headerName: 'Net', minWidth: 130, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => formatCurrency(params.value as number),
			cellClass: 'text-sm font-bold text-blue-600 text-right tabular-nums',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'id', headerName: '', minWidth: 80, maxWidth: 80,
			cellRenderer: (params: AgGridCellParams<PayrollItem>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Slip Gaji',
					() => viewPayslip(item)
				);
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
		if (items.length > 0 && gridContainer && agGridModule) {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: items });
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		loadData();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});
	async function loadData() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const [periodRes, itemsRes] = await Promise.all([
				payrollApi.getPeriod(periodId),
				payrollApi.listItems(periodId, currentPage, perPage),
			]);
			const periodData = periodRes as ApiResponse<Period>;
			const itemsData = itemsRes as ApiResponse<PayrollItem[]>;
			period = periodData.data || null;
			items = itemsData.data || [];
			total = itemsData.meta?.total || 0;
			currentPage = itemsData.meta?.page || 1;
			perPage = itemsData.meta?.per_page || 50;
			totalPages = Math.ceil(total / perPage);
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	function getStatusColor(status: string): string {
		const map: Record<string, string> = {
			draft: 'bg-gray-50 text-gray-700 ring-gray-600/20',
			calculated: 'bg-blue-50 text-blue-700 ring-blue-600/20',
			approved: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20',
			paid: 'bg-purple-50 text-purple-700 ring-purple-600/20',
		};
		return map[status] || 'bg-gray-50 text-gray-700 ring-gray-600/20';
	}

	function statusLabel(status: string): string {
		const map: Record<string, string> = {
			draft: 'Draft', calculated: 'Dihitung', approved: 'Disetujui', paid: 'Dibayarkan',
		};
		return map[status] || status;
	}

	let allowanceTotal = $derived(items.reduce((sum, i) => {
		return sum + (i.allowances || []).reduce((s, a) => s + a.amount, 0);
	}, 0));

	let deductionTotal = $derived(items.reduce((sum, i) => sum + i.total_deductions, 0));
	let netTotal = $derived(items.reduce((sum, i) => sum + i.net_salary, 0));

	function viewPayslip(item: PayrollItem) {
		goto(`/penggajian/payslip/${periodId}/${item.employee_id}`);
	}

	// ── THR Modal ──
	let showTHRModal = $state(false);
	let isCalculatingTHR = $state(false);
	let thrResult = $state<{ message?: string; items_calculated?: number } | null>(null);
	let thrError = $state('');

	function handleOpenTHRModal() {
		thrResult = null;
		thrError = '';
		showTHRModal = true;
	}

	async function handleCalculateTHR() {
		isCalculatingTHR = true;
		thrError = '';
		thrResult = null;
		try {
			const res: any = await payrollApi.calculateTHR(periodId);
			thrResult = res.data || { message: 'THR berhasil dihitung' };
			// Refresh data
			await loadData();
		} catch (err: unknown) {
			thrError = (err as { message?: string }).message || 'Gagal menghitung THR';
		} finally {
			isCalculatingTHR = false;
		}
	}

	function closeTHRModal() {
		showTHRModal = false;
		thrResult = null;
		thrError = '';
	}
</script>

<!-- THR Calculation Modal -->
{#if showTHRModal}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<div onclick={closeTHRModal} onkeydown={(e) => { if (e.key === 'Escape') closeTHRModal(); }}
		role="presentation" tabindex="-1" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} class="bg-white rounded-2xl shadow-2xl w-full max-w-md" role="dialog" tabindex="-1" aria-modal="true" aria-label="Hitung THR">
			<div class="px-6 py-6">
				<div class="text-center mb-6">
					<div class="w-14 h-14 mx-auto mb-4 rounded-2xl bg-amber-50 flex items-center justify-center">
						<svg class="w-7 h-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
						</svg>
					</div>
					<h3 class="text-lg font-semibold text-gray-900 mb-2">Hitung THR</h3>
					<p class="text-sm text-gray-500">THR akan dihitung secara proporsional berdasarkan masa kerja setiap karyawan hingga periode ini.</p>
				</div>

				{#if thrResult}
					<div class="bg-emerald-50 border border-emerald-200 rounded-xl p-4 mb-4">
						<div class="flex items-center gap-3">
							<svg class="w-5 h-5 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
							<div>
								<p class="text-sm font-semibold text-emerald-800">THR Berhasil Dihitung</p>
								<p class="text-xs text-emerald-600 mt-0.5">{thrResult.message || 'THR berhasil dihitung untuk semua karyawan.'}</p>
							</div>
						</div>
					</div>
				{/if}

				{#if thrError}
					<div class="bg-red-50 border border-red-200 rounded-xl p-4 mb-4">
						<div class="flex items-center gap-3">
							<svg class="w-5 h-5 text-red-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
							</svg>
							<p class="text-sm text-red-700">{thrError}</p>
						</div>
					</div>
				{/if}

				<div class="flex items-center justify-center gap-3">
					<button onclick={closeTHRModal}
						class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">
						{thrResult ? 'Tutup' : 'Batal'}
					</button>
					{#if !thrResult}
						<button onclick={handleCalculateTHR} disabled={isCalculatingTHR}
							class="px-5 py-2.5 bg-gradient-to-r from-amber-500 to-amber-600 text-white rounded-lg text-sm font-semibold hover:from-amber-600 hover:to-amber-700 transition-all active:scale-[0.97] disabled:opacity-50 inline-flex items-center gap-2 shadow-sm shadow-amber-200 cursor-pointer">
							{#if isCalculatingTHR}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								Menghitung...
							{:else}
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
								Konfirmasi Hitung THR
							{/if}
						</button>
					{/if}
				</div>
			</div>
		</div>
	</div>
{/if}

<div class="w-full">
	<!-- Back Button -->
	<button onclick={() => goto('/penggajian')}
		class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 transition mb-5 cursor-pointer group">
		<svg class="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
		</svg>
		<span>Kembali ke daftar periode</span>
	</button>

	{#if isLoading}
		<div class="bg-white border border-gray-200 rounded-xl p-6 animate-pulse space-y-4">
			<div class="h-6 bg-gray-100 rounded w-48"></div>
			<div class="h-4 bg-gray-50 rounded w-72"></div>
			<div class="grid grid-cols-4 gap-4 mt-4">
				{#each [1,2,3,4] as _}<div class="h-20 bg-gray-100 rounded-xl"></div>{/each}
			</div>
		</div>
	{:else if errorMessage}
		<div class="bg-white border border-gray-200 rounded-xl py-16 text-center">
			<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
			<button onclick={loadData} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button>
		</div>
	{:else if period}
		<!-- Period Header -->
		<div class="bg-white border border-gray-200 rounded-xl p-6 mb-4 shadow-sm">
			<div class="flex items-center justify-between mb-4">
				<div>
					<h1 class="text-xl font-bold text-gray-900">{period.period_name}</h1>
					<p class="text-sm text-gray-500 mt-0.5">{formatDate(period.start_date)} — {formatDate(period.end_date)}</p>
				</div>
				<div class="flex items-center gap-2">
					{#if hasPermission('payroll', 'update') && (period.status === 'draft' || period.status === 'calculated')}
						<button onclick={handleOpenTHRModal}
							class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium rounded-lg border border-amber-200 text-amber-700 bg-amber-50 hover:bg-amber-100 hover:border-amber-300 transition-all cursor-pointer"
							title="Hitung THR untuk periode ini">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
							Hitung THR
						</button>
					{/if}
					<span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ring-1 ring-inset capitalize {getStatusColor(period.status)}">
						{statusLabel(period.status)}
					</span>
				</div>
			</div>
			<div class="grid grid-cols-2 sm:grid-cols-4 gap-3">
				<div class="bg-gray-50 rounded-lg px-4 py-3">
					<span class="text-xs font-medium text-gray-500">Karyawan</span>
					<p class="text-lg font-bold text-gray-900 mt-0.5">{period.total_employee}</p>
				</div>
				<div class="bg-gray-50 rounded-lg px-4 py-3">
					<span class="text-xs font-medium text-gray-500">Total Tunjangan</span>
					<p class="text-lg font-bold text-emerald-600 mt-0.5">{period.total_gross > 0 ? formatCurrency(allowanceTotal) : '-'}</p>
				</div>
				<div class="bg-gray-50 rounded-lg px-4 py-3">
					<span class="text-xs font-medium text-gray-500">Total Potongan</span>
					<p class="text-lg font-bold text-red-600 mt-0.5">{period.total_deductions > 0 ? formatCurrency(period.total_deductions) : '-'}</p>
				</div>
				<div class="bg-gray-50 rounded-lg px-4 py-3">
					<span class="text-xs font-medium text-gray-500">Total Net</span>
					<p class="text-lg font-bold text-blue-600 mt-0.5">{period.total_net > 0 ? formatCurrency(period.total_net) : '-'}</p>
				</div>
			</div>
		</div>

		<!-- Payroll Items Table -->
		<div class="bg-white border border-gray-200 rounded-xl shadow-sm overflow-hidden">
			<div class="px-5 py-3.5 border-b border-gray-100 bg-gray-50/50">
				<h2 class="text-sm font-semibold text-gray-900">Daftar Gaji Karyawan</h2>
			</div>

			{#if items.length === 0}
				<div class="py-12 text-center">
					<p class="text-sm text-gray-400 mb-2">Belum ada data gaji</p>
					{#if period.status === 'draft' && hasPermission('payroll', 'update')}
						<button onclick={() => payrollApi.calculatePayroll(periodId).then(() => loadData())}
							class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-xs font-medium cursor-pointer">
							Hitung Gaji Sekarang
						</button>
					{/if}
				</div>
			{:else}					<!-- Desktop Table — AG Grid -->
					<div class="hidden md:block">
						<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
					</div>

				<PullToRefresh onRefresh={loadData}>
				<div class="md:hidden space-y-3">
					{#each items as item}
						<MobileCard
							title={item.employee_name}
							subtitle={`${item.department_name} · ${item.position_name}`}
							avatar={getInitials(item.employee_name)}
							avatarColor={getAvatarTheme('payroll').gradientClasses}
							onclick={() => viewPayslip(item)}
						>
							{#snippet children()}
								<div class="grid grid-cols-3 gap-3">
									<div class="text-center">
										<div class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider">Dasar</div>
										<div class="text-xs font-semibold text-gray-900 dark:text-white mt-0.5">{item.base_salary > 0 ? formatCurrency(item.base_salary) : '-'}</div>
									</div>
									<div class="text-center">
										<div class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider">Gross</div>
										<div class="text-xs font-semibold text-gray-900 dark:text-white mt-0.5">{formatCurrency(item.gross_salary)}</div>
									</div>
									<div class="text-center">
										<div class="text-[10px] text-gray-400 dark:text-gray-500 uppercase tracking-wider">Net</div>
										<div class="text-xs font-semibold text-blue-600 dark:text-blue-400 mt-0.5">{formatCurrency(item.net_salary)}</div>
									</div>
								</div>
								{#if item.overtime_pay > 0 || item.thr_amount > 0}
									<div class="flex items-center gap-3 mt-2 text-[10px] text-gray-400 dark:text-gray-500">
										{#if item.overtime_pay > 0}<span>Lembur: {formatCurrency(item.overtime_pay)}</span>{/if}
										{#if item.thr_amount > 0}<span>THR: {formatCurrency(item.thr_amount)}</span>{/if}
									</div>
								{/if}
							{/snippet}
						</MobileCard>
					{/each}
				</div>
				</PullToRefresh>
			{/if}
		</div>
	{/if}
</div>
