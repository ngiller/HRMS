<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { shiftChangeRequests as api, workSchedules, employees } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	type ShiftChangeRequest = {
		id: string;
		request_type: string;
		employee_id: string;
		employee_name: string;
		target_date: string;
		current_schedule_name: string;
		requested_schedule_name: string;
		swap_partner_name: string;
		reason: string;
		status: string;
		created_at: string;
	};

	type FormData = {
		request_type: string;
		target_date: string;
		current_schedule_id: string;
		requested_schedule_id: string;
		swap_partner_id: string;
		swap_partner_date: string;
		swap_partner_schedule_id: string;
		reason: string;
	};

	type WorkScheduleOption = { id: string; name: string };

	let items = $state<ShiftChangeRequest[]>([]);
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
		request_type: 'individual', target_date: '', current_schedule_id: '',
		requested_schedule_id: '', swap_partner_id: '', swap_partner_date: '',
		swap_partner_schedule_id: '', reason: ''
	});
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<any>(null);
	let isDetailLoading = $state(false);

	let scheduleOptions = $state<WorkScheduleOption[]>([]);
	let scheduleSwapOptions = $state<WorkScheduleOption[]>([]);
	let employeeOptions = $state<any[]>([]);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu',
		partner_pending: 'Tunggu Partner',
		approved: 'Disetujui',
		rejected: 'Ditolak',
		cancelled: 'Dibatalkan',
	};

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		partner_pending: 'bg-orange-50 text-orange-700 ring-orange-200 dark:bg-orange-900 dark:text-orange-200 dark:ring-orange-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700',
	};

	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: any = null;

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
	function iconSwap(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5" /></svg>';
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
			field: 'employee_name', headerName: 'Karyawan', minWidth: 240, flex: 1,
			cellRenderer: (params: any) => {
				if (!params.value) return '';
				const initials = getInitials(params.value);
				const swapName = params.data?.swap_partner_name || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 flex items-center justify-center text-xs font-semibold text-purple-600 shrink-0 ring-1 ring-purple-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div>${swapName ? `<div class="text-xs text-gray-400">Swap: ${swapName}</div>` : ''}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'request_type', headerName: 'Tipe', minWidth: 100, maxWidth: 120,
			valueFormatter: (params: any) => params.value === 'individual' ? 'Individual' : params.value === 'swap' ? 'Swap' : params.value || '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'target_date', headerName: 'Tanggal', minWidth: 120,
			valueFormatter: (params: any) => formatDate(params.value),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'requested_schedule_name', headerName: 'Jadwal Diminta', minWidth: 160,
			cellRenderer: (params: any) => {
				const requested = params.value || '';
				const current = params.data?.current_schedule_name || '';
				return `<div class="text-sm text-gray-700">${requested}</div>${current ? `<div class="text-xs text-gray-400">Dari: ${current}</div>` : ''}`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 130, maxWidth: 150,
			cellRenderer: (params: any) => {
				const status = params.value || '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${statusLabels[status] || status}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 120,
			valueFormatter: (params: any) => formatDate(params.value),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 150, maxWidth: 150,
			cellRenderer: (params: any) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('shift_change', 'update')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-green-500 hover:text-green-700 hover:bg-green-50 transition cursor-pointer',
						'Setujui', () => handleApprove(item.id));
					container.appendChild(approveBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
				}

				if (item.status === 'partner_pending' && hasPermission('shift_change', 'update')) {
					const swapBtn = createActionButton(iconSwap(),
						'p-1.5 rounded-lg text-blue-500 hover:text-blue-700 hover:bg-blue-50 transition cursor-pointer',
						'Konfirmasi Swap', () => handleConfirmSwap(item.id));
					container.appendChild(swapBtn);
				}

				if ((item.status === 'pending' || item.status === 'partner_pending') && hasPermission('shift_change', 'create')) {
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
		if ((showForm || showDetail) && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (items.length > 0 && gridContainer && !showForm && !showDetail) {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: items as any[] });
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
		loadOptions();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});
	async function loadOptions() {
		try {
			const [schedRes, empRes]: any[] = await Promise.all([
				workSchedules.getAll(),
				employees.list(1, 9999),
			]);
			scheduleOptions = schedRes.data || [];
			employeeOptions = empRes.data || [];
		} catch (_) {}
	}

	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
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

	function openCreateForm() {
		formTitle = 'Ajukan Permintaan Shift';
		form = {
			request_type: 'individual', target_date: '', current_schedule_id: '',
			requested_schedule_id: '', swap_partner_id: '', swap_partner_date: '',
			swap_partner_schedule_id: '', reason: ''
		};
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.target_date) { formError = 'Tanggal target harus diisi'; return; }
		if (!form.requested_schedule_id) { formError = 'Jadwal yang diminta harus dipilih'; return; }
		if (!form.reason.trim()) { formError = 'Alasan permintaan harus diisi'; return; }
		if (form.request_type === 'swap' && !form.swap_partner_id) { formError = 'Partner swap harus dipilih'; return; }

		isSaving = true;
		formError = '';
		try {
			const payload: any = {
				request_type: form.request_type,
				target_date: form.target_date,
				current_schedule_id: form.current_schedule_id || '',
				requested_schedule_id: form.requested_schedule_id,
				swap_partner_id: form.swap_partner_id || '',
				swap_partner_date: form.swap_partner_date || '',
				swap_partner_schedule_id: form.swap_partner_schedule_id || '',
				reason: form.reason.trim(),
			};
			await api.create(payload);
			cancelForm();
			load();
		} catch (error: any) { formError = error.message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true;
		detailId = id;
		isDetailLoading = true;
		detailData = null;
		try {
			const response: any = await api.get(id);
			detailData = response.data;
		} catch (_) { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; }

	async function handleApprove(id: string) {
		isSaving = true;
		try { await api.approve(id); load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		isSaving = true;
		try { await api.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; rejectId = null; rejectReason = ''; load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	async function handleConfirmSwap(id: string) {
		isSaving = true;
		try { await api.confirmSwap(id); load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal konfirmasi swap'; }
		finally { isSaving = false; }
	}

	async function handleCancel(id: string) {
		isSaving = true;
		try { await api.cancel(id); load(); }
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
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Permintaan Shift</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan dan kelola perubahan jadwal shift karyawan</p>
		</div>
		{#if !showForm && hasPermission('shift_change', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Ajukan Permintaan
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">Semua</button>
				{#each Object.entries(statusLabels) as [key, label]}
					<button onclick={() => { statusFilter = key; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{label}</button>
				{/each}
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} permintaan ditemukan` : ''}</div>
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
						<label for="shift-type" class="block text-sm font-medium text-gray-700 mb-1.5">Tipe Permintaan <span class="text-red-500">*</span></label>
						<select id="shift-type" bind:value={form.request_type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="individual">Individual</option>
							<option value="swap">Tukar Shift (Swap)</option>
						</select>
					</div>
					<div>
						<label for="shift-target-date" class="block text-sm font-medium text-gray-700 mb-1.5">Tanggal Target <span class="text-red-500">*</span></label>
						<input id="shift-target-date" type="date" bind:value={form.target_date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="shift-current" class="block text-sm font-medium text-gray-700 mb-1.5">Jadwal Saat Ini</label>
						<select id="shift-current" bind:value={form.current_schedule_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="">-- Pilih (opsional) --</option>
							{#each scheduleOptions as s}<option value={s.id}>{s.name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="shift-requested" class="block text-sm font-medium text-gray-700 mb-1.5">Jadwal yang Diminta <span class="text-red-500">*</span></label>
						<select id="shift-requested" bind:value={form.requested_schedule_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="">-- Pilih --</option>
							{#each scheduleOptions as s}<option value={s.id}>{s.name}</option>{/each}
						</select>
					</div>
				</div>
				{#if form.request_type === 'swap'}
					<div class="border-t border-gray-100 pt-4">
						<h3 class="text-sm font-semibold text-gray-700 mb-3">Informasi Partner Swap</h3>
						<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
							<div>
								<label for="shift-partner" class="block text-sm font-medium text-gray-700 mb-1.5">Partner Karyawan <span class="text-red-500">*</span></label>
								<select id="shift-partner" bind:value={form.swap_partner_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
									<option value="">-- Pilih --</option>
									{#each employeeOptions.filter(e => e.full_name) as e}<option value={e.id}>{e.full_name} ({e.employee_id})</option>{/each}
								</select>
							</div>
							<div>
								<label for="shift-partner-date" class="block text-sm font-medium text-gray-700 mb-1.5">Tanggal Partner (opsional)</label>
								<input id="shift-partner-date" type="date" bind:value={form.swap_partner_date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							</div>
							<div>
								<label for="shift-partner-sched" class="block text-sm font-medium text-gray-700 mb-1.5">Jadwal Partner (opsional)</label>
								<select id="shift-partner-sched" bind:value={form.swap_partner_schedule_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
									<option value="">-- Pilih (opsional) --</option>
									{#each scheduleOptions as s}<option value={s.id}>{s.name}</option>{/each}
								</select>
							</div>
						</div>
					</div>
				{/if}
				<div>
					<label for="shift-reason" class="block text-sm font-medium text-gray-700 mb-1.5">Alasan <span class="text-red-500">*</span></label>
					<textarea id="shift-reason" bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Jelaskan alasan perubahan shift..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Permintaan
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">Detail Permintaan Shift</h2>
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
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Informasi Permintaan</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Tipe</span><p class="text-sm font-medium text-gray-900">{detailData.request_type === 'individual' ? 'Individual' : 'Tukar Shift (Swap)'}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p><span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 {statusColors[detailData.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">{statusLabels[detailData.status] || detailData.status}</span></p></div>
								<div><span class="text-xs text-gray-400">Tanggal Target</span><p class="text-sm font-medium text-gray-900">{formatDate(detailData.target_date)}</p></div>
								<div><span class="text-xs text-gray-400">Alasan</span><p class="text-sm text-gray-700">{detailData.reason || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Diajukan Pada</span><p class="text-sm text-gray-700">{formatDate(detailData.created_at)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Karyawan & Jadwal</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Pengaju</span><p class="text-sm font-medium text-gray-900">{detailData.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Jadwal Saat Ini</span><p class="text-sm text-gray-700">{detailData.current_schedule_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Jadwal Diminta</span><p class="text-sm font-medium text-gray-900">{detailData.requested_schedule_name || '-'}</p></div>
								{#if detailData.swap_partner_name}
									<div><span class="text-xs text-gray-400">Partner Swap</span><p class="text-sm font-medium text-gray-900">{detailData.swap_partner_name}</p></div>
									<div><span class="text-xs text-gray-400">Partner Konfirmasi</span><p class="text-sm">{detailData.swap_partner_confirmed ? 'Sudah dikonfirmasi' : 'Belum dikonfirmasi'}</p></div>
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail permintaan</p>
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
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 flex items-center justify-center"><svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182" /></svg></div>
					<h3 class="text-sm font-semibold text-gray-900 mb-1">Belum ada permintaan shift</h3>
					<p class="text-sm text-gray-500">Belum ada permintaan perubahan shift yang diajukan.</p>
				</div>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden divide-y divide-gray-100">
					{#each items as item}
						<div class="p-4 hover:bg-blue-50/40 transition-colors">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 flex items-center justify-center text-xs font-semibold text-purple-600 ring-1 ring-purple-200">{getInitials(item.employee_name)}</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900 truncate">{item.employee_name}</div>
									<div class="text-xs text-gray-400">{item.request_type === 'individual' ? 'Individual' : 'Swap'} &middot; {formatDate(item.target_date)}</div>
								</div>
								<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 {statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">{statusLabels[item.status] || item.status}</span>
							</div>
							<div class="flex items-center justify-end gap-1 mt-2">
								<button onclick={() => openDetail(item.id)} class="px-2.5 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 transition cursor-pointer">Detail</button>
								{#if item.status === 'pending' && hasPermission('shift_change', 'update')}
									<button onclick={() => handleApprove(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-green-50 text-green-700 hover:bg-green-100 transition cursor-pointer">Setujui</button>
									<button onclick={() => openReject(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-red-50 text-red-700 hover:bg-red-100 transition cursor-pointer">Tolak</button>
								{/if}
								{#if item.status === 'partner_pending' && hasPermission('shift_change', 'update')}
									<button onclick={() => handleConfirmSwap(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-blue-50 text-blue-700 hover:bg-blue-100 transition cursor-pointer">Konfirmasi</button>
								{/if}
								{#if (item.status === 'pending' || item.status === 'partner_pending') && hasPermission('shift_change', 'create')}
									<button onclick={() => handleCancel(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-gray-50 text-gray-600 hover:bg-gray-100 transition cursor-pointer">Batalkan</button>
								{/if}
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
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak permintaan shift" class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 text-center mb-4">Tolak Permintaan Shift</h3>
				<div class="space-y-3">
					<label for="shift-reject-reason" class="block text-sm font-medium text-gray-700">Alasan Penolakan</label>
					<textarea id="shift-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
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
