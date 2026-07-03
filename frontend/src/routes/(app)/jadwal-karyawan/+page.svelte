<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { employeeSchedules as api, scheduleTemplates as templateApi, employees, attendanceLocations } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	type EmployeeSummary = { id: string; full_name: string; employee_id: string; department_name: string; };
	type TemplateSummary = { id: string; name: string; description: string; schedule_type: string; day_count: number; };
	type LocationSummary = { id: string; name: string; address: string; };
	type ScheduleEntry = {
		id: string; employee_id: string; employee_name: string;
		template_name: string; day_of_week: number | null;
		specific_date: string | null; start_time: string | null;
		end_time: string | null; is_remote: boolean;
		effective_from: string; effective_until: string | null;
		priority: number; is_active: boolean;
		location_names: string; created_at: string;
	};
	type ResolvedSchedule = {
		schedule_id: string; source: string;
		start_time: string; end_time: string;
		break_start: string; break_end: string;
		is_remote: boolean;
		location?: { name: string; address: string; latitude: number; longitude: number; radius_meters: number; };
	};

	// Tab state
	type Tab = 'list' | 'assign' | 'resolve';
	let activeTab = $state<Tab>('list');
	let tabCounts = $state({ list: 0, assign: 0, resolve: 0 });

	// === Tab 1: List Employee Schedules ===
	let schedules = $state<ScheduleEntry[]>([]);
	let totalSchedules = $state(0);
	let schedPage = $state(1);
	let schedPerPage = $state(25);
	let schedTotalPages = $state(0);
	let schedEmployeeId = $state('');
	let isLoadingSchedules = $state(false);
	let schedError = $state('');

	// Delete
	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);

	// === Tab 2: Assign Schedule ===
	let employees_list = $state<EmployeeSummary[]>([]);
	let templates = $state<TemplateSummary[]>([]);
	let locations = $state<LocationSummary[]>([]);

	let assignForm = $state({
		employee_id: '',
		template_id: '',
		day_of_week: '',
		specific_date: '',
		start_time: '',
		end_time: '',
		break_start: '12:00',
		break_end: '13:00',
		is_remote: false,
		effective_from: new Date().toISOString().split('T')[0],
		effective_until: '',
		priority: 0,
		reason: '',
		use_template: true,
	});

	let assignFormError = $state('');
	let isAssigning = $state(false);
	let assignedLocations = $state<{ attendance_location_id: string; day_of_week: string | null; }[]>([]);

	// === Tab 3: Resolve Schedule ===
	let resolveEmployeeId = $state('');
	let resolveDate = $state(new Date().toISOString().split('T')[0]);
	let resolvedSchedule = $state<ResolvedSchedule | null>(null);
	let isResolving = $state(false);
	let resolveError = $state('');

	// AG Grid
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: any = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	const DAY_LABELS: Record<number, string> = { 0: 'Senin', 1: 'Selasa', 2: 'Rabu', 3: 'Kamis', 4: 'Jumat', 5: 'Sabtu', 6: 'Minggu' };

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
				return `<div class="flex items-center gap-2"><div class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center text-xs font-semibold text-blue-600">${params.value.substring(0, 2).toUpperCase()}</div><span class="text-sm font-medium text-gray-900">${params.value}</span></div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'specific_date', headerName: 'Tanggal', minWidth: 120,
			cellRenderer: (params: any) => {
				if (params.value) return `<span class="text-sm font-medium text-gray-900">${formatDate(params.value)}</span>`;
				if (params.data?.day_of_week != null) return `<span class="text-sm text-gray-600">Setiap ${DAY_LABELS[params.data.day_of_week] || ''}</span>`;
				return '<span class="text-xs text-gray-400">Periodik</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'template_name', headerName: 'Template', minWidth: 140,
			cellRenderer: (params: any) => params.value ? `<span class="text-sm text-gray-700">${params.value}</span>` : '<span class="text-xs text-gray-400">Manual</span>',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'start_time', headerName: 'Jam', minWidth: 120,
			valueFormatter: (params: any) => params.data?.start_time && params.data?.end_time ? `${params.data.start_time.slice(0, 5)}-${params.data.end_time.slice(0, 5)}` : '-',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'is_remote', headerName: 'Remote', minWidth: 90,
			cellRenderer: (params: any) => params.value ? '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-50 text-purple-700">WFH</span>' : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'effective_from', headerName: 'Berlaku', minWidth: 160,
			valueFormatter: (params: any) => {
				if (!params.data) return '';
				const from = formatDate(params.data.effective_from);
				const until = params.data.effective_until ? formatDate(params.data.effective_until) : '∞';
				return `${from} - ${until}`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'location_names', headerName: 'Lokasi', minWidth: 130,
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 80, maxWidth: 80,
			cellRenderer: (params: any) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				const deleteBtn = createActionButton(
					'<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>',
					'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer', 'Hapus', () => confirmDelete(item.id));
				container.appendChild(deleteBtn);
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
		if (activeTab !== 'list' && gridApi) { gridApi.destroy(); gridApi = null; }
	});

	$effect(() => {
		if (schedules.length > 0 && gridContainer && activeTab === 'list') {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: schedules as any[] });
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		loadSchedules();
		loadAllEmployees();
		loadTemplates();
		loadLocations();
	});

	onDestroy(() => { gridApi?.destroy(); gridApi = null; });

	async function loadSchedules() {
		gridApi?.destroy();
		gridApi = null;
		isLoadingSchedules = true; schedError = '';
		try {
			const response: any = await api.list(schedPage, schedPerPage, schedEmployeeId);
			schedules = response.data || [];
			totalSchedules = response.meta?.total || 0;
			schedTotalPages = Math.ceil(totalSchedules / schedPerPage);
			tabCounts.list = totalSchedules;
		} catch (error: any) { schedError = error.message || 'Gagal memuat data'; }
		finally { isLoadingSchedules = false; }
	}

	async function loadAllEmployees() {
		try {
			const response: any = await employees.list(1, 500);
			employees_list = response.data || [];
		} catch { employees_list = []; }
	}

	async function loadTemplates() {
		try {
			const response: any = await templateApi.getAll();
			templates = response.data || [];
		} catch { templates = []; }
	}

	async function loadLocations() {
		try {
			const response: any = await attendanceLocations.getAll();
			locations = response.data || [];
		} catch { locations = []; }
	}

	function confirmDelete(id: string) { deletingId = id; showDeleteConfirm = true; }
	function cancelDelete() { showDeleteConfirm = false; deletingId = null; }

	async function handleDelete() {
		if (!deletingId) return;
		try { await api.remove(deletingId); showDeleteConfirm = false; deletingId = null; loadSchedules(); }
		catch (error: any) { schedError = error.message || 'Gagal menghapus'; showDeleteConfirm = false; }
	}

	function openAssignTab() {
		activeTab = 'assign';
		resetForm();
	}

	function resetForm() {
		assignForm = {
			employee_id: '', template_id: '', day_of_week: '', specific_date: '',
			start_time: '', end_time: '', break_start: '12:00', break_end: '13:00',
			is_remote: false, effective_from: new Date().toISOString().split('T')[0],
			effective_until: '', priority: 0, reason: '', use_template: true,
		};
		assignedLocations = [];
		assignFormError = '';
	}

	async function handleAssign() {
		if (!assignForm.employee_id) { assignFormError = 'Pilih karyawan'; return; }
		if (!assignForm.effective_from) { assignFormError = 'Tanggal berlaku harus diisi'; return; }

		const isPeriodic = assignForm.day_of_week !== '';
		const isSpecific = assignForm.specific_date !== '';

		if (!isPeriodic && !isSpecific) {
			assignFormError = 'Pilih hari dalam seminggu (periodik) atau tanggal spesifik';
			return;
		}

		isAssigning = true; assignFormError = '';
		try {
			const payload: any = {
				employee_id: assignForm.employee_id,
				effective_from: assignForm.effective_from,
				reason: assignForm.reason || '',
				locations: assignedLocations.map(l => ({
					attendance_location_id: l.attendance_location_id,
					day_of_week: l.day_of_week ? parseInt(l.day_of_week) : null,
				})),
			};

			if (isPeriodic) payload.day_of_week = parseInt(assignForm.day_of_week);
			if (isSpecific) payload.specific_date = assignForm.specific_date;
			if (assignForm.effective_until) payload.effective_until = assignForm.effective_until;
			if (assignForm.priority) payload.priority = assignForm.priority;
			if (assignForm.is_remote) payload.is_remote = true;

			if (assignForm.use_template && assignForm.template_id) {
				payload.template_id = assignForm.template_id;
			} else {
				payload.start_time = assignForm.start_time || '08:00';
				payload.end_time = assignForm.end_time || '17:00';
				payload.break_start = assignForm.break_start || '12:00';
				payload.break_end = assignForm.break_end || '13:00';
			}

			await api.create(payload);
			resetForm();
			activeTab = 'list';
			loadSchedules();
		} catch (error: any) { assignFormError = error.message || 'Gagal menyimpan jadwal'; }
		finally { isAssigning = false; }
	}

	function addLocation() {
		assignedLocations = [...assignedLocations, { attendance_location_id: '', day_of_week: null }];
	}

	function removeLocation(index: number) {
		assignedLocations = assignedLocations.filter((_, i) => i !== index);
	}

	function updateLocation(index: number, field: string, value: any) {
		assignedLocations = assignedLocations.map((loc, i) => i === index ? { ...loc, [field]: value } : loc);
	}

	async function handleResolve() {
		if (!resolveEmployeeId) { resolveError = 'Pilih karyawan'; return; }
		if (!resolveDate) { resolveError = 'Pilih tanggal'; return; }

		isResolving = true; resolveError = ''; resolvedSchedule = null;
		try {
			const response: any = await api.resolve(resolveEmployeeId, resolveDate);
			resolvedSchedule = response.data;
		} catch (error: any) { resolveError = error.message || 'Gagal menentukan jadwal'; }
		finally { isResolving = false; }
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function goToSchedPage(p: number) {
		if (p < 1) return;
		schedPage = p;
		loadSchedules();
	}

	const sourceLabels: Record<string, string> = {
		date_override: 'Override Tanggal',
		weekly_schedule: 'Jadwal Periodik',
		work_schedule: 'Jadwal Individu',
		department_schedule: 'Jadwal Departemen',
	};

	const sourceColors: Record<string, string> = {
		date_override: 'bg-amber-50 text-amber-700 border-amber-200',
		weekly_schedule: 'bg-blue-50 text-blue-700 border-blue-200',
		work_schedule: 'bg-gray-50 text-gray-700 border-gray-200',
		department_schedule: 'bg-purple-50 text-purple-700 border-purple-200',
	};
</script>

<div class="w-full">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Atur Jadwal Karyawan</h1>
			<p class="text-sm text-gray-500 mt-0.5">Atur jadwal kerja fleksibel per karyawan — periodik atau tanggal spesifik</p>
		</div>
		{#if activeTab === 'list'}
		<div>
			<button onclick={() => { activeTab = 'resolve'; resolvedSchedule = null; resolveError = ''; }} class="px-4 py-2 border border-gray-200 bg-white text-gray-700 rounded-lg text-sm font-medium hover:bg-gray-50 transition cursor-pointer flex items-center gap-2">
				<svg class="w-4 h-4 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
				Alat Cek Jadwal
			</button>
		</div>
		{/if}
	</div>

	<!-- Tabs Removed for uniformity -->
	<!-- Tab 1: List -->
	{#if activeTab === 'list'}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="px-5 py-3.5 border-b border-gray-100 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
				<div class="flex items-center gap-3">
					<div class="relative">
						<select bind:value={schedEmployeeId} onchange={() => { schedPage = 1; loadSchedules(); }} class="w-full sm:w-56 px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							<option value="">Semua Karyawan</option>
							{#each employees_list as emp}
								<option value={emp.id}>{emp.full_name} ({emp.employee_id})</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="flex items-center justify-between sm:justify-end gap-3 w-full sm:w-auto">
					<div class="text-xs text-gray-400">{totalSchedules > 0 ? `${totalSchedules} jadwal ditemukan` : ''}</div>
					<button onclick={openAssignTab} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
						Tambah Jadwal
					</button>
				</div>
			</div>

			{#if isLoadingSchedules}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 rounded w-44"></div><div class="h-3 bg-gray-50 rounded w-28"></div></div><div class="h-8 bg-gray-100 rounded w-20"></div></div>{/each}</div></div>
			{/if}
			{#if !isLoadingSchedules && schedError}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{schedError}</p>
					<button onclick={loadSchedules} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{/if}
			{#if !isLoadingSchedules && !schedError && schedules.length === 0}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 flex items-center justify-center"><svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg></div>
					<h3 class="text-sm font-semibold text-gray-900 mb-1">Belum ada jadwal khusus</h3>
					<p class="text-sm text-gray-500 mb-4">Karyawan akan mengikuti jadwal default dari departemen masing-masing.</p>
					<button onclick={openAssignTab} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Tambah Jadwal</button>
				</div>
			{/if}
			{#if schedules.length > 0}
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden divide-y divide-gray-100">
					{#each schedules as s}
						<div class="p-4 hover:bg-blue-50/40 transition-colors">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-10 h-10 rounded-full bg-gradient-to-br from-blue-50 to-blue-100 flex items-center justify-center text-xs font-semibold text-blue-600 shrink-0">{s.employee_name.substring(0, 2).toUpperCase()}</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900 truncate">{s.employee_name}</div>
									<div class="text-xs text-gray-400">
										{s.specific_date ? formatDate(s.specific_date) : (s.day_of_week != null ? `Setiap ${DAY_LABELS[s.day_of_week]}` : 'Periodik')}
										· {s.start_time?.slice(0,5)}-{s.end_time?.slice(0,5)}
									</div>
								</div>
								<button onclick={() => confirmDelete(s.id)} class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer" aria-label="Hapus jadwal">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
								</button>
							</div>
						</div>
					{/each}
				</div>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30">
					<div class="text-xs text-gray-500">Menampilkan {(schedPage - 1) * schedPerPage + 1}-{Math.min(schedPage * schedPerPage, totalSchedules)} dari <span class="font-medium text-gray-700">{totalSchedules}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToSchedPage(schedPage - 1)} disabled={schedPage <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, schedTotalPages) }) as _, i}
							{@const pageNum = Math.max(1, Math.min(schedPage - 2, schedTotalPages - 4)) + i}
							{#if pageNum <= schedTotalPages}
								<button onclick={() => goToSchedPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === schedPage ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pageNum}</button>
							{/if}
						{/each}						<button onclick={() => goToSchedPage(schedPage + 1)} disabled={schedPage >= schedTotalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<div style="display: {activeTab === 'assign' ? 'block' : 'none'}">
		<!-- Tab 2: Assign Schedule -->
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">Tambah Jadwal Karyawan</h2>
				<button onclick={() => { activeTab = 'list'; }} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer" aria-label="Tutup">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if assignFormError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{assignFormError}</div>{/if}

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Karyawan <span class="text-red-500">*</span>
							<select bind:value={assignForm.employee_id} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">Pilih karyawan</option>
								{#each employees_list as emp}
									<option value={emp.id}>{emp.full_name} ({emp.employee_id})</option>
								{/each}
							</select>
						</label>
					</div>
					<div>
						<div>
						<span class="block text-sm font-medium text-gray-700 mb-1.5">Tipe Jadwal <span class="text-red-500">*</span></span>
						<div class="flex items-center gap-4">
							<label class="inline-flex items-center gap-2 cursor-pointer">
								<input type="radio" bind:group={assignForm.use_template} value={true} class="text-[#1A56DB] focus:ring-[#1A56DB]/20" />
								<span class="text-sm text-gray-700">Pakai Template</span>
							</label>
							<label class="inline-flex items-center gap-2 cursor-pointer">
								<input type="radio" bind:group={assignForm.use_template} value={false} class="text-[#1A56DB] focus:ring-[#1A56DB]/20" />
								<span class="text-sm text-gray-700">Manual</span>
							</label>
						</div>						</div>
					</div>

				{#if assignForm.use_template}
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Pilih Template
							<select bind:value={assignForm.template_id} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">Pilih template (opsional)</option>
								{#each templates as t}
									<option value={t.id}>{t.name} ({t.day_count} hari)</option>
								{/each}
							</select>
						</label>
					</div>
				{:else}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
								Jam Mulai
								<input type="time" bind:value={assignForm.start_time} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							</label>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
								Jam Selesai
								<input type="time" bind:value={assignForm.end_time} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							</label>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
								Istirahat Mulai
								<input type="time" bind:value={assignForm.break_start} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							</label>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
								Istirahat Selesai
								<input type="time" bind:value={assignForm.break_end} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
							</label>
						</div>
					</div>
				{/if}

				<div class="border-t border-gray-100 pt-4">
					<h3 class="text-sm font-semibold text-gray-800 mb-3">Penjadwalan</h3>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Hari dalam Seminggu (periodik)
							<select bind:value={assignForm.day_of_week} onchange={() => { if (assignForm.day_of_week !== '') assignForm.specific_date = ''; }} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">Pilih hari (opsional)</option>
								<option value="0">Senin</option>
								<option value="1">Selasa</option>
								<option value="2">Rabu</option>
								<option value="3">Kamis</option>
								<option value="4">Jumat</option>
								<option value="5">Sabtu</option>
								<option value="6">Minggu</option>						</select>
							<p class="text-xs text-gray-400 mt-1">Atau pilih tanggal spesifik di bawah</p>
						</label>
						</div>
						<div>
							<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Tanggal Spesifik
							<input type="date" bind:value={assignForm.specific_date} onchange={() => { if (assignForm.specific_date !== '') assignForm.day_of_week = ''; }} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</label>
							<p class="text-xs text-gray-400 mt-1">Untuk jadwal satu hari tertentu (libur nasional, dll)</p>
						</div>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
						Mulai Berlaku <span class="text-red-500">*</span>
						<input type="date" bind:value={assignForm.effective_from} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</label>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
						Berlaku Sampai
						<input type="date" bind:value={assignForm.effective_until} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</label>
						<p class="text-xs text-gray-400 mt-1">Kosongkan jika berlaku terus</p>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
						Prioritas
						<input type="number" bind:value={assignForm.priority} min="0" max="100" class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</label>
						<p class="text-xs text-gray-400 mt-1">Makin tinggi, makin diutamakan</p>
					</div>
				</div>

				<div class="flex items-center gap-3">
					<label class="relative inline-flex items-center cursor-pointer">
						<input type="checkbox" bind:checked={assignForm.is_remote} class="sr-only peer" />
						<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[#1A56DB]/20 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#1A56DB]"></div>
						<span class="ms-2 text-sm font-medium text-gray-700">Remote / WFH (tanpa validasi GPS)</span>
					</label>
				</div>

				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1.5">
					Alasan
					<input type="text" bind:value={assignForm.reason} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Contoh: Jadwal Ramadhan, Lembur Proyek" />
				</label>
				</div>

				<!-- Locations -->
				<div class="border-t border-gray-100 pt-4">
					<div class="flex items-center justify-between mb-3">
						<h3 class="text-sm font-semibold text-gray-800">Lokasi Absensi</h3>
						<button onclick={addLocation} class="inline-flex items-center gap-1.5 px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-100 transition cursor-pointer">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
							Tambah Lokasi
						</button>
					</div>
					{#if assignedLocations.length === 0}
						<p class="text-xs text-gray-400">Tidak ada lokasi khusus — karyawan bisa absen dari lokasi mana pun yang aktif.</p>
					{:else}
						<div class="space-y-2">
							{#each assignedLocations as loc, i}
								<div class="flex items-center gap-2">
									<select bind:value={loc.attendance_location_id} onchange={(e) => updateLocation(i, 'attendance_location_id', (e.target as HTMLSelectElement).value)} class="flex-1 px-3 py-2 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
										<option value="">Pilih lokasi</option>
										{#each locations as l}
											<option value={l.id}>{l.name}</option>
										{/each}
									</select>
									<button onclick={() => removeLocation(i)} class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer" aria-label="Hapus lokasi">
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={() => { activeTab = 'list'; }} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleAssign} disabled={isAssigning} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isAssigning}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Simpan Jadwal
				</button>
			</div>
		</div>
	</div>

	{#if activeTab === 'resolve'}
		<!-- Tab 3: Resolve Schedule -->
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<div>
					<h2 class="text-lg font-semibold text-gray-900">Cek Jadwal Karyawan</h2>
					<p class="text-xs text-gray-400">Lihat jadwal efektif untuk tanggal tertentu</p>
				</div>
				<button onclick={() => { activeTab = 'list'; }} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer" aria-label="Tutup">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
							Karyawan <span class="text-red-500">*</span>
							<select bind:value={resolveEmployeeId} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">Pilih karyawan</option>
								{#each employees_list as emp}
									<option value={emp.id}>{emp.full_name} ({emp.employee_id})</option>
								{/each}
							</select>
						</label>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5">
						Tanggal <span class="text-red-500">*</span>
						<input type="date" bind:value={resolveDate} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</label>
					</div>
					<div class="flex items-end">
						<button onclick={handleResolve} disabled={isResolving} class="w-full px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center justify-center gap-2 cursor-pointer">
							{#if isResolving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
							Cek Jadwal
						</button>
					</div>
				</div>

				{#if resolveError}
					<div class="bg-amber-50 border border-amber-200 text-amber-700 text-sm px-4 py-3 rounded-lg flex items-center gap-2">
						<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
						{resolveError}
					</div>
				{/if}

				{#if resolvedSchedule}
					<div class="bg-gradient-to-br from-blue-50 to-indigo-50 border border-blue-200 rounded-xl p-5">
						<h3 class="text-sm font-semibold text-blue-900 mb-4 flex items-center gap-2">
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
							Jadwal Efektif — {formatDate(resolveDate)}
						</h3>
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
							<div class="bg-white rounded-lg p-3 border border-blue-100">
								<div class="text-xs text-gray-500 mb-1">Sumber</div>
								<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border {sourceColors[resolvedSchedule.source] || 'bg-gray-50 text-gray-600 border-gray-200'}">{sourceLabels[resolvedSchedule.source] || resolvedSchedule.source}</span>
							</div>
							<div class="bg-white rounded-lg p-3 border border-blue-100">
								<div class="text-xs text-gray-500 mb-1">Jam Kerja</div>
								<div class="text-base font-semibold text-gray-900 tabular-nums">{resolvedSchedule.start_time.slice(0,5)} - {resolvedSchedule.end_time.slice(0,5)}</div>
							</div>
							<div class="bg-white rounded-lg p-3 border border-blue-100">
								<div class="text-xs text-gray-500 mb-1">Istirahat</div>
								<div class="text-base font-semibold text-gray-900 tabular-nums">{resolvedSchedule.break_start.slice(0,5)} - {resolvedSchedule.break_end.slice(0,5)}</div>
							</div>
							<div class="bg-white rounded-lg p-3 border border-blue-100">
								<div class="text-xs text-gray-500 mb-1">Status</div>
								{#if resolvedSchedule.is_remote}
									<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-50 text-purple-700">Remote / WFH</span>
								{:else}
									<span class="text-base font-semibold text-gray-900">Non-Remote</span>
								{/if}
							</div>
						</div>
						{#if resolvedSchedule.location}
							<div class="mt-4 bg-white rounded-lg p-3 border border-blue-100">
								<div class="text-xs text-gray-500 mb-2">Lokasi Absensi</div>
								<div class="flex items-center gap-2">
									<svg class="w-4 h-4 text-emerald-600 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
									<span class="text-sm font-medium text-gray-900">{resolvedSchedule.location.name}</span>
								</div>
								{#if resolvedSchedule.location.address}
									<p class="text-xs text-gray-500 mt-0.5 ml-6">{resolvedSchedule.location.address}</p>
								{/if}
								<div class="text-xs text-gray-400 mt-1 ml-6">
									Koordinat: {resolvedSchedule.location.latitude.toFixed(6)}, {resolvedSchedule.location.longitude.toFixed(6)}
									· Radius {resolvedSchedule.location.radius_meters}m
								</div>
							</div>
						{/if}
						<div class="mt-3 text-xs text-gray-400">
							<span class="font-medium">Scheduling ID:</span> {resolvedSchedule.schedule_id}
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div></div>

	<!-- Delete Confirmation Modal -->
	{#if showDeleteConfirm}
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelDelete} onkeydown={(e) => { if (e.key === 'Escape') cancelDelete(); }} role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" aria-modal="true" aria-label="Hapus jadwal" tabindex="-1" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Hapus Jadwal Karyawan</h3>
				<p class="text-sm text-gray-500 mb-6">Apakah Anda yakin ingin menghapus jadwal ini?</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition cursor-pointer">Ya, Hapus</button>
				</div>
			</div>
		</div>
	</div>
{/if}
