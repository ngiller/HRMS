<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { workSchedules as api } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

	type WorkSchedule = {
		id: string;
		name: string;
		schedule_type: string;
		description: string;
		weekly_hours: number;
		is_active: boolean;
		created_at: string;
		monday_start: string;
		monday_end: string;
		tuesday_start: string;
		tuesday_end: string;
		wednesday_start: string;
		wednesday_end: string;
		thursday_start: string;
		thursday_end: string;
		friday_start: string;
		friday_end: string;
		saturday_start: string;
		saturday_end: string;
		sunday_start: string;
		sunday_end: string;
		break_start: string;
		break_end: string;
		late_tolerance_minutes: number;
		early_leave_tolerance: number;
	};

	type Form = {
		[key: string]: any;
		name: string;
		schedule_type: string;
		description: string;
		monday_start: string;
		monday_end: string;
		tuesday_start: string;
		tuesday_end: string;
		wednesday_start: string;
		wednesday_end: string;
		thursday_start: string;
		thursday_end: string;
		friday_start: string;
		friday_end: string;
		saturday_start: string;
		saturday_end: string;
		sunday_start: string;
		sunday_end: string;
		break_start: string;
		break_end: string;
		late_tolerance_minutes: number;
		early_leave_tolerance: number;
		weekly_hours: number;
	};

	let items = $state<WorkSchedule[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let searchQuery = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let editingId = $state<string | null>(null);
	let form = $state<Form>({
		name: '', schedule_type: '5_day', description: '',
		monday_start: '08:00', monday_end: '17:00',
		tuesday_start: '08:00', tuesday_end: '17:00',
		wednesday_start: '08:00', wednesday_end: '17:00',
		thursday_start: '08:00', thursday_end: '17:00',
		friday_start: '08:00', friday_end: '17:00',
		saturday_start: '', saturday_end: '',
		sunday_start: '', sunday_end: '',
		break_start: '12:00', break_end: '13:00',
		late_tolerance_minutes: 15, early_leave_tolerance: 15, weekly_hours: 40,
	});
	let formError = $state('');
	let isSaving = $state(false);

	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');
	let showDetail = $state(false);
	let detailItem = $state<WorkSchedule | null>(null);

	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	const DAY_NAMES = ['Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat'];
	const START_KEYS = ['monday_start', 'tuesday_start', 'wednesday_start', 'thursday_start', 'friday_start'];
	const END_KEYS = ['monday_end', 'tuesday_end', 'wednesday_end', 'thursday_end', 'friday_end'];

	function iconView(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>';
	}
	function iconEdit(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>';
	}
	function iconDelete(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>';
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
			field: 'name', headerName: 'Jadwal Kerja', minWidth: 240, flex: 1,
			cellRenderer: (params: AgGridCellParams<WorkSchedule>) => {
				if (!params.value) return '';
				const desc = params.data?.description || '';
				return '<div class="flex items-center gap-3">'
					+ '<div class="w-9 h-9 rounded-full bg-gradient-to-br from-cyan-50 to-cyan-100 flex items-center justify-center text-xs font-semibold text-cyan-600 shrink-0 ring-1 ring-cyan-200">' + (params.value as string).substring(0, 2).toUpperCase() + '</div>'
					+ '<div><div class="text-sm font-medium text-gray-900">' + params.value + '</div>' + (desc ? '<div class="text-xs text-gray-400 truncate max-w-48">' + desc + '</div>' : '') + '</div>'
					+ '</div>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'schedule_type', headerName: 'Tipe', minWidth: 130,
			cellRenderer: (params: AgGridCellParams<WorkSchedule>) => {
				if (!params.value) return '';
				const label = scheduleTypeLabels[params.value as string] || params.value;
				return '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium ring-1 bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800">' + label + '</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'weekly_hours', headerName: 'Jam/Minggu', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => params.value != null ? params.value + ' jam' : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-700 tabular-nums',
		},
		{
			field: 'is_active', headerName: 'Status', minWidth: 110,
			cellRenderer: (params: AgGridCellParams<WorkSchedule>) => {
				if (params.value == null) return '';
				return params.value
					? '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800">Aktif</span>'
					: '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800">Nonaktif</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 130,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: 'Aksi', minWidth: 120,
			cellRenderer: (params: AgGridCellParams<WorkSchedule>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center gap-2';
				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail jadwal',
					() => viewDetail(item)
				);
				container.appendChild(viewBtn);
				const editBtn = createActionButton(iconEdit(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Edit jadwal',
					() => openEditForm(item)
				);
				container.appendChild(editBtn);
				if (hasPermission('work_schedule', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Hapus jadwal',
						() => confirmDelete(item.id, item.name)
					);
					container.appendChild(deleteBtn);
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
		if (gridContainer && agGridModule && !showForm && !showDetail) {
			if (!gridApi) {
				gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi;
			}
			gridApi.updateGridOptions({ rowData: items });
			gridApi.sizeColumnsToFit();
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
			const response: ApiResponse<WorkSchedule[]> = await api.list(page, perPage, searchQuery) as ApiResponse<WorkSchedule[]>;
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
		formTitle = 'Tambah Jadwal Kerja';
		editingId = null;
		form = { name: '', schedule_type: '5_day', description: '', monday_start: '08:00', monday_end: '17:00', tuesday_start: '08:00', tuesday_end: '17:00', wednesday_start: '08:00', wednesday_end: '17:00', thursday_start: '08:00', thursday_end: '17:00', friday_start: '08:00', friday_end: '17:00', saturday_start: '', saturday_end: '', sunday_start: '', sunday_end: '', break_start: '12:00', break_end: '13:00', late_tolerance_minutes: 15, early_leave_tolerance: 15, weekly_hours: 40 };
		formError = '';
		showForm = true;
	}

	async function openEditForm(item: WorkSchedule) {
		formTitle = 'Edit Jadwal Kerja';
		editingId = item.id;
		try {
			const resp: ApiResponse<WorkSchedule> = await api.get(item.id) as ApiResponse<WorkSchedule>;
			const d = resp.data;
			if (d) {
				form = {
					name: d.name, schedule_type: d.schedule_type, description: d.description || '',
					monday_start: d.monday_start || '08:00', monday_end: d.monday_end || '17:00',
					tuesday_start: d.tuesday_start || '08:00', tuesday_end: d.tuesday_end || '17:00',
					wednesday_start: d.wednesday_start || '08:00', wednesday_end: d.wednesday_end || '17:00',
					thursday_start: d.thursday_start || '08:00', thursday_end: d.thursday_end || '17:00',
					friday_start: d.friday_start || '08:00', friday_end: d.friday_end || '17:00',
					saturday_start: d.saturday_start || '', saturday_end: d.saturday_end || '',
					sunday_start: d.sunday_start || '', sunday_end: d.sunday_end || '',
					break_start: d.break_start || '12:00', break_end: d.break_end || '13:00',
					late_tolerance_minutes: d.late_tolerance_minutes || 15, early_leave_tolerance: d.early_leave_tolerance || 15,
					weekly_hours: d.weekly_hours || 40,
				};
			}
		} catch {}
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.name.trim()) { formError = 'Nama jadwal kerja harus diisi'; return; }
		if (!form.schedule_type) { formError = 'Tipe jadwal harus dipilih'; return; }

		isSaving = true;
		formError = '';
		try {
			const p: Record<string, unknown> = { name: form.name.trim(), schedule_type: form.schedule_type, description: form.description.trim(), monday_start: form.monday_start, monday_end: form.monday_end, tuesday_start: form.tuesday_start, tuesday_end: form.tuesday_end, wednesday_start: form.wednesday_start, wednesday_end: form.wednesday_end, thursday_start: form.thursday_start, thursday_end: form.thursday_end, friday_start: form.friday_start, friday_end: form.friday_end, break_start: form.break_start, break_end: form.break_end, late_tolerance_minutes: form.late_tolerance_minutes, early_leave_tolerance: form.early_leave_tolerance, weekly_hours: form.weekly_hours };
			if (form.saturday_start) p.saturday_start = form.saturday_start;
			if (form.saturday_end) p.saturday_end = form.saturday_end;
			if (form.sunday_start) p.sunday_start = form.sunday_start;
			if (form.sunday_end) p.sunday_end = form.sunday_end;
			if (editingId) { await api.update(editingId, p); }
			else { await api.create(p); }
			cancelForm();
			load();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	function confirmDelete(id: string, name: string) { deletingId = id; deletingName = name; showDeleteConfirm = true; }
	function cancelDelete() { showDeleteConfirm = false; deletingId = null; deletingName = ''; }

	async function handleDelete() {
		if (!deletingId) return; isSaving = true;
		try { await api.remove(deletingId); showDeleteConfirm = false; deletingId = null; deletingName = ''; load(); }
		catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menghapus data'; showDeleteConfirm = false; }
		finally { isSaving = false; }
	}

	function viewDetail(item: WorkSchedule) { detailItem = item; showDetail = true; }

	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; page = 1; load(); }, 400);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function getPageNum(i: number): number {
		return Math.max(1, Math.min(page - 2, totalPages - 4)) + i;
	}

	const scheduleTypeLabels: Record<string, string> = { '5_day': '5 Hari Kerja', '6_day': '6 Hari Kerja', shift: 'Shift' };
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Jadwal Kerja</h1>
			<p class="text-sm text-gray-500 mt-0.5">Kelola jadwal dan jam kerja karyawan</p>
		</div>
		{#if !showForm && !showDetail && hasPermission('work_schedule', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Jadwal
			</button>
		{/if}
	</div>

	{#if !showForm && !showDetail}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				<input type="search" value={searchQuery} placeholder="Cari jadwal kerja..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white transition placeholder:text-gray-400" />
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? total + ' jadwal ditemukan' : ''}</div>
		</div>
	{/if}

	<div class:hidden={!showForm}>
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">{formTitle}</h2>
				<button onclick={cancelForm} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer" aria-label="Tutup">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>
				{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Nama Jadwal <span class="text-red-500">*</span>
							<input type="text" bind:value={form.name} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Contoh: 5 Hari Kerja" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Tipe Jadwal <span class="text-red-500">*</span>
							<select bind:value={form.schedule_type} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="5_day">5 Hari Kerja (Senin-Jumat)</option>
								<option value="6_day">6 Hari Kerja (Senin-Sabtu)</option>
								<option value="shift">Shift</option>
							</select>
						</label>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Deskripsi
							<textarea bind:value={form.description} rows="2" class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Deskripsi (opsional)"></textarea>
						</label>
					</div>
				</div>

				<h3 class="text-sm font-semibold text-gray-800 pt-2 border-t border-gray-100">Jam Kerja</h3>
				<div class="grid grid-cols-1 md:grid-cols-5 gap-4">
					{#each DAY_NAMES as day, i}
						<div>
							<div>
								<span class="block text-xs font-medium text-gray-600 mb-1">{day}</span>
								<div class="flex items-center gap-1">
									<input type="time" bind:value={form[START_KEYS[i]]} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" aria-label={day + ' mulai'} />
									<span class="text-xs text-gray-400">-</span>
									<input type="time" bind:value={form[END_KEYS[i]]} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
								</div>
							</div>
						</div>
					{/each}
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<div>
							<span class="block text-xs font-medium text-gray-600 mb-1" id="sabtu-label">Sabtu Mulai - Selesai</span>
							<div class="flex items-center gap-1">
								<input type="time" bind:value={form.saturday_start} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" aria-labelledby="sabtu-label" />
								<span class="text-xs text-gray-400">-</span>
								<input type="time" bind:value={form.saturday_end} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" aria-label="Sabtu selesai" />
							</div>
						</div>
					</div>
					<div>
						<span class="block text-xs font-medium text-gray-600 mb-1" id="minggu-label">Minggu Mulai - Selesai</span>
						<div class="flex items-center gap-1">
							<input type="time" bind:value={form.sunday_start} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" aria-labelledby="minggu-label" />
							<span class="text-xs text-gray-400">-</span>
							<input type="time" bind:value={form.sunday_end} class="w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" aria-label="Minggu selesai" />
						</div>
					</div>
				</div>

				<h3 class="text-sm font-semibold text-gray-800 pt-2 border-t border-gray-100">Pengaturan</h3>
				<div class="grid grid-cols-1 md:grid-cols-4 gap-4">
					<div>
						<label class="block text-xs font-medium text-gray-600 mb-1">
							Istirahat Mulai
							<input type="time" bind:value={form.break_start} class="mt-1 w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</label>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-600 mb-1">
							Istirahat Selesai
							<input type="time" bind:value={form.break_end} class="mt-1 w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</label>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-600 mb-1">
							Toleransi Telat (menit)
							<input type="number" bind:value={form.late_tolerance_minutes} min="0" max="120" class="mt-1 w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</label>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-600 mb-1">
							Jam Kerja/Minggu
							<input type="number" bind:value={form.weekly_hours} min="0" max="60" step="0.5" class="mt-1 w-full px-2 py-1.5 border border-gray-200 rounded text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</label>
					</div>
				</div>
			</div>
			<div class="sticky bottom-0 z-10 flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/95 dark:bg-gray-900/95 backdrop-blur-sm">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
					{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Jadwal'}
				</button>
			</div>
		</div>
	</div>

	{#if showDetail && detailItem}
		<div>
			<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm hover:shadow-md transition-all duration-200">
				<div class="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{detailItem.name}</h2>
					<button onclick={() => { showDetail = false; detailItem = null; }} class="p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer" aria-label="Tutup">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>
				</div>
				<div class="px-5 py-5 space-y-4">
					<div class="grid grid-cols-2 gap-4 text-sm">
						<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl px-4 py-3">
							<div class="text-xs text-gray-400 dark:text-gray-500">Tipe</div>
							<div class="text-sm font-semibold text-gray-900 dark:text-white mt-0.5">{scheduleTypeLabels[detailItem.schedule_type] || detailItem.schedule_type}</div>
						</div>
						<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl px-4 py-3">
							<div class="text-xs text-gray-400 dark:text-gray-500">Jam/Minggu</div>
							<div class="text-sm font-semibold text-gray-900 dark:text-white mt-0.5">{detailItem.weekly_hours} jam</div>
						</div>
						<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl px-4 py-3">
							<div class="text-xs text-gray-400 dark:text-gray-500">Status</div>
							<div class="mt-0.5"><span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {detailItem.is_active ? 'bg-green-50 text-green-700 dark:bg-green-900 dark:text-green-200' : 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400'}">{detailItem.is_active ? 'Aktif' : 'Nonaktif'}</span></div>
						</div>
						<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl px-4 py-3">
							<div class="text-xs text-gray-400 dark:text-gray-500">Dibuat</div>
							<div class="text-sm font-semibold text-gray-900 dark:text-white mt-0.5">{formatDate(detailItem.created_at)}</div>
						</div>
					</div>
					{#if detailItem.description}
						<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl px-4 py-3 text-sm">
							<div class="text-xs text-gray-400 dark:text-gray-500 mb-1">Deskripsi</div>
							<div class="text-gray-700 dark:text-gray-300">{detailItem.description}</div>
						</div>
					{/if}
					<div class="flex items-center justify-end gap-2 pt-4 border-t border-gray-100 dark:border-gray-800">
						<button onclick={() => { const item = detailItem; showDetail = false; detailItem = null; if (item) openEditForm(item); }} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Edit</button>
					</div>
				</div>
			</div>
		</div>
	{/if}

	{#if !showForm && !showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<PulseLoader variant="table-row" count={5} />
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center">
						<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
					</div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada data jadwal kerja"
					description={searchQuery ? `Tidak ditemukan dengan kata kunci "${searchQuery}"` : 'Data jadwal kerja akan muncul setelah ditambahkan.'}
				/>
			{:else}
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden space-y-3">
					<PullToRefresh onRefresh={load}>
					{#each items as item}
						<MobileCard
							title={item.name}
							subtitle={scheduleTypeLabels[item.schedule_type] || item.schedule_type}
							avatar={getInitials(item.name)}
							avatarColor={getAvatarTheme('schedule').gradientClasses}
							badges={[{ label: item.is_active ? 'Aktif' : 'Nonaktif', color: item.is_active ? 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800' : 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800' }]}
							onclick={() => viewDetail(item)}
						>
							{#snippet children()}
								<div class="flex items-center gap-4 text-xs text-gray-500 dark:text-gray-400">
									<span>{item.weekly_hours} jam/minggu</span>
									<span>· {formatDate(item.created_at)}</span>
								</div>
							{/snippet}
						</MobileCard>
					{/each}
					</PullToRefresh>
				</div>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-800 bg-gray-50/30 dark:bg-gray-900/30">
					<div class="text-xs text-gray-500 dark:text-gray-400">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pageNum = getPageNum(i)}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}
</div>

<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<button type="button" onclick={cancelDelete} aria-label="Tutup" class="absolute inset-0 cursor-default border-none bg-transparent"></button>
		<div onclick={(e) => e.stopPropagation()} onkeydown={(e) => { if (e.key === 'Escape') cancelDelete(); }} role="dialog" aria-modal="true" aria-label="Hapus jadwal kerja" tabindex="-1" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm relative">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Hapus Jadwal Kerja</h3>
				<p class="text-sm text-gray-500 mb-1">Apakah Anda yakin ingin menghapus</p>
				<p class="text-sm font-medium text-gray-900 mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 mb-6">Data yang sudah dihapus tidak dapat dikembalikan.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
						{/if}
						Ya, Hapus
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
