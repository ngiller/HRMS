<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { leaveRequests as api } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	type LeaveRequest = {
		id: string;
		employee_id: string;
		employee_name: string;
		leave_type_name: string;
		start_date: string;
		end_date: string;
		total_days: number;
		is_half_day: boolean;
		reason: string;
		status: string;
		created_at: string;
	};

	type LeaveType = { id: string; name: string; code: string; default_quota: number; is_paid: boolean; is_active: boolean; };
	type LeaveBalance = { id: string; leave_type_name: string; total_quota: number; used: number; remaining: number; };

	type FormData = {
		leave_type_id: string;
		start_date: string;
		end_date: string;
		total_days: number;
		is_half_day: boolean;
		reason: string;
		document_url: string;
		contact_during_leave: string;
	};

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
	let form = $state<FormData>({
		leave_type_id: '', start_date: '', end_date: '', total_days: 1,
		is_half_day: false, reason: '', document_url: '', contact_during_leave: '',
	});
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailData = $state<any>(null);
	let isDetailLoading = $state(false);

	let leaveTypes = $state<LeaveType[]>([]);
	let myBalances = $state<LeaveBalance[]>([]);
	let showBalances = $state(true);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	// AG Grid
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: any = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	const statusColors: Record<string, string> = {
    pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
    approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
    active: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
    completed: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
    rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
    defaulted: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
    cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
    paid: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
};

function getStatusBadge(status: string) {
    const labels: Record<string, string> = {
        pending: 'Menunggu',
        approved: 'Disetujui',
        active: 'Aktif',
        completed: 'Lunas',
        rejected: 'Ditolak',
        defaulted: 'Macet',
        cancelled: 'Dibatalkan',
        paid: 'Dibayar'
    };
    return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${labels[status] || status}</span>`;
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
			cellRenderer: (params: any) => {
				if (!params.value) return '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-sky-50 to-sky-100 flex items-center justify-center text-xs font-semibold text-sky-600 shrink-0 ring-1 ring-sky-200">${getInitials(params.value)}</div>
					<span class="text-sm font-medium text-gray-900">${params.value}</span>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'leave_type_name', headerName: 'Jenis Cuti', minWidth: 140,
			cellRenderer: (params: any) => {
				if (!params.value) return '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-indigo-50 text-indigo-700">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'start_date', headerName: 'Tanggal', minWidth: 140,
			cellRenderer: (params: any) => {
				if (!params.value) return '';
				const end = params.data?.end_date || '';
				return `<span class="text-sm text-gray-700">${formatDate(params.value)}${end && end !== params.value ? ` — ${formatDate(end)}` : ''}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm',
		},
		{
			field: 'total_days', headerName: 'Hari', minWidth: 80, maxWidth: 90,
			valueFormatter: (params: any) => params.value != null ? `${params.value} hr` : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-700 text-center',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: any) => {
				const status = params.value || '';
				return getStatusBadge(status);
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Diajukan', minWidth: 120,
			valueFormatter: (params: any) => formatDate(params.value),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 140, maxWidth: 140,
			cellRenderer: (params: any) => {
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
		if (items.length > 0 && gridContainer && !showForm && !showDetail) {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: items as any[] });
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
			const res: any = await api.getTypes();
			leaveTypes = res.data || [];
		} catch { leaveTypes = []; }
	}

	async function loadMyBalances() {
		try {
			const res: any = await api.getMyBalances();
			myBalances = res.data || [];
		} catch { myBalances = []; }
	}

	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true; errorMessage = '';
		try {
			const response: any = await api.list(page, perPage, statusFilter);
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (error: any) { errorMessage = error.message || 'Gagal memuat data'; }
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

	async function handleSave() {
		if (!form.leave_type_id) { formError = 'Jenis cuti harus dipilih'; return; }
		if (!form.start_date) { formError = 'Tanggal mulai harus diisi'; return; }
		if (!form.end_date) { formError = 'Tanggal selesai harus diisi'; return; }
		if (form.total_days <= 0) { formError = 'Jumlah hari harus lebih dari 0'; return; }
		if (!form.reason.trim()) { formError = 'Alasan cuti harus diisi'; return; }

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
		} catch (error: any) { formError = error.message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true; isDetailLoading = true; detailData = null;
		try {
			const response: any = await api.get(id);
			detailData = response.data;
		} catch { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailData = null; }

	async function handleApprove(id: string) {
		isSaving = true;
		try { await api.approve(id); load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return; isSaving = true;
		try { await api.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; rejectId = null; rejectReason = ''; load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	async function handleCancel(id: string) {
		isSaving = true;
		try { await api.cancel(id, {}); load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal membatalkan'; }
		finally { isSaving = false; }
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

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Cuti</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan dan kelola cuti karyawan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail}

				{#if hasPermission('leave', 'create')}
					<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
						Ajukan Cuti
					</button>
				{/if}
			{/if}
		</div>
	</div>

	{#if showBalances && myBalances.length > 0}
		<div class="bg-white border border-gray-200 rounded-xl p-5 mb-4 shadow-sm">
			<h3 class="text-sm font-semibold text-gray-800 mb-3">Sisa Cuti Saya</h3>
			<div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
				{#each myBalances as b}
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
						<label for="leave-type" class="block text-sm font-medium text-gray-700 mb-1.5">Jenis Cuti <span class="text-red-500">*</span></label>
						<select id="leave-type" bind:value={form.leave_type_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="">-- Pilih Jenis Cuti --</option>
							{#each leaveTypes as t}
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
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 rounded w-44"></div><div class="h-3 bg-gray-50 rounded w-28"></div></div><div class="h-6 bg-gray-100 rounded-full w-20"></div><div class="h-8 bg-gray-100 rounded w-24"></div></div>{/each}</div></div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 flex items-center justify-center"><svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" /></svg></div>
					<h3 class="text-sm font-semibold text-gray-900 mb-1">Belum ada pengajuan cuti</h3>
					<p class="text-sm text-gray-500 mb-4">Belum ada pengajuan cuti yang diajukan.</p>
					{#if hasPermission('leave', 'create')}
						<button onclick={openCreateForm} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Ajukan Cuti</button>
					{/if}
				</div>
			{:else}
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden divide-y divide-gray-100">
					{#each items as item}
						<div class="p-4 hover:bg-blue-50/40 transition-colors">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-sky-50 to-sky-100 flex items-center justify-center text-xs font-semibold text-sky-600 ring-1 ring-sky-200">{getInitials(item.employee_name)}</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900 truncate">{item.employee_name}</div>
									<div class="text-xs text-gray-400">{item.leave_type_name} &middot; {item.total_days} hr</div>
								</div>
								{@html getStatusBadge(item.status)}
							</div>
						</div>
					{/each}
				</div>
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
</div>

{#if showRejectModal}
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
{/if}
