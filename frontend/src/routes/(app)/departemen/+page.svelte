<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { departments as deptApi } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import { AnimatedPresence } from '$lib';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { Department, WorkSchedule, DepartmentForm, ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

	let departments = $state<Department[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let searchQuery = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Form state — gantikan modal
	let showForm = $state(false);
	let formTitle = $state('');
	let editingId = $state<string | null>(null);
	let form = $state<DepartmentForm>({ name: '', code: '', parent_id: '', head_id: '', work_schedule_id: '', description: '', is_active: true });
	let formError = $state('');
	let isSaving = $state(false);

	// Delete confirm (tetap modal kecil)
	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');

	// Schedule picker inline
	// AG Grid
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

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
			field: 'name', headerName: 'Departemen', minWidth: 240, flex: 1,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const desc = params.data?.description || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-indigo-50 to-indigo-100 flex items-center justify-center text-xs font-semibold text-indigo-600 shrink-0 ring-1 ring-indigo-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div>${desc ? `<div class="text-xs text-gray-400 truncate max-w-48">${desc}</div>` : ''}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'code', headerName: 'Kode', minWidth: 100,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				if (!params.value) return '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-gray-100 text-gray-600">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{ field: 'head_name', headerName: 'Kepala', minWidth: 140, headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider', cellClass: 'text-sm text-gray-600' },
		{ field: 'parent_name', headerName: 'Induk', minWidth: 140, headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider', cellClass: 'text-sm text-gray-500' },
		{
			field: 'work_schedule_name', headerName: 'Jadwal', minWidth: 160,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				const dept = params.data;
				if (!dept) return '';
				if (schedulePickerDeptId === dept.id) {
					const container = document.createElement('div');
					container.className = 'flex items-center gap-1.5';
					const select = document.createElement('select');
					select.className = 'w-40 px-2 py-1 border border-[#1A56DB] rounded text-xs outline-none bg-white';
					const emptyOpt = document.createElement('option');
					emptyOpt.value = '';
					emptyOpt.textContent = '(tanpa jadwal)';
					select.appendChild(emptyOpt);
					for (const ws of workSchedules) {
						const opt = document.createElement('option');
						opt.value = ws.id;
						opt.textContent = ws.name;
						if (ws.id === selectedScheduleId) opt.selected = true;
						select.appendChild(opt);
					}
					select.onchange = () => { selectedScheduleId = select.value; };
					container.appendChild(select);
					const saveBtn = document.createElement('button');
					saveBtn.className = 'px-2 py-1 bg-[#1A56DB] text-white rounded text-xs font-medium hover:bg-[#1e40af] transition cursor-pointer';
					saveBtn.textContent = isSavingSchedule ? '...' : 'Simpan';
					saveBtn.disabled = isSavingSchedule;
					saveBtn.onclick = () => handleScheduleChange();
					container.appendChild(saveBtn);
					const cancelBtn = document.createElement('button');
					cancelBtn.className = 'px-2 py-1 border border-gray-200 rounded text-xs text-gray-600 hover:bg-gray-100 transition cursor-pointer';
					cancelBtn.textContent = 'Batal';
					cancelBtn.onclick = () => closeSchedulePicker();
					container.appendChild(cancelBtn);
					return container;
				}
				const wrapper = document.createElement('div');
				wrapper.className = 'flex items-center gap-1.5 group';
				const span = document.createElement('span');
				span.className = 'text-sm';
				span.textContent = dept.work_schedule_name || '-';
				wrapper.appendChild(span);
				const editBtn = document.createElement('button');
				editBtn.innerHTML = '<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>';
				editBtn.className = 'p-1 rounded text-gray-300 hover:text-[#1A56DB] hover:bg-blue-50 opacity-0 group-hover:opacity-100 transition-all cursor-pointer';
				editBtn.setAttribute('aria-label', 'Ubah jadwal');
				editBtn.onclick = (e) => {
					e.stopPropagation();
					const wsId = dept.work_schedule_name ? (workSchedules.find((w: WorkSchedule) => w.name === dept.work_schedule_name)?.id || '') : '';
					openSchedulePicker(dept.id, wsId);
				};
				if (hasPermission('department', 'update')) {
					wrapper.appendChild(editBtn);
				}
				return wrapper;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'employee_count', headerName: 'Karyawan', minWidth: 100, maxWidth: 120,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				if (params.value == null) return '';
				return `<span class="text-sm font-medium text-gray-700 tabular-nums">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-center',
		},
		{
			field: 'is_active', headerName: 'Status', minWidth: 110,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				if (params.value == null) return '';
				if (params.value) {
					return '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20">Aktif</span>';
				}
				return '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-600/20">Nonaktif</span>';
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
			field: 'id', headerName: '', minWidth: 110, maxWidth: 110,
			cellRenderer: (params: AgGridCellParams<Department>) => {
				const dept = params.data;
				if (!dept) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				if (hasPermission('department', 'update')) {
					const editBtn = createActionButton(iconEdit(),
						'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
						'Edit departemen',
						() => openEditForm(dept)
					);
					container.appendChild(editBtn);
				}
				if (hasPermission('department', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Hapus departemen',
						() => confirmDelete(dept.id, dept.name)
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
		if (!gridContainer && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (departments.length > 0 && gridContainer && !showForm) {
			if (!gridApi && agGridModule) {
				gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi;
			}
			if (gridApi) {
				gridApi.updateGridOptions({ rowData: departments });
			}
		}
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});

	let schedulePickerDeptId = $state<string | null>(null);
	let selectedScheduleId = $state('');
	let isSavingSchedule = $state(false);

	// All departments for parent dropdown
	let allDepts = $state<Department[]>([]);

	// Work schedules for dropdown
	let workSchedules = $state<WorkSchedule[]>([]);

	// Summary stats (computed from current page)
	let totalActiveDepts = $state(0);
	let totalSubDepts = $state(0);
	let totalEmployeesAll = $state(0);

	onMount(async () => { const m = await getAgGrid(); agGridModule = m;
		loadDepartments();
		loadAllDepts();
		loadWorkSchedules();
	});

	async function loadDepartments() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response: ApiResponse<Department[]> = await deptApi.list(page, perPage, searchQuery) as ApiResponse<Department[]>;
			const data = response.data || [];
			departments = data;
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);

			totalActiveDepts = data.filter((d: Department) => d.is_active).length;
			totalSubDepts = data.filter((d: Department) => d.parent_name).length;
			totalEmployeesAll = data.reduce((sum: number, d: Department) => sum + (d.employee_count || 0), 0);
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal memuat data departemen';
			console.error('Department list error:', error);
		} finally {
			isLoading = false;
		}
	}

	async function loadAllDepts() {
		try {
			const response: ApiResponse<Department[]> = await deptApi.getAll() as ApiResponse<Department[]>;
			allDepts = response.data || [];
		} catch {
			allDepts = [];
		}
	}

	async function loadWorkSchedules() {
		try {
			const response: ApiResponse<WorkSchedule[]> = await deptApi.getWorkSchedules() as ApiResponse<WorkSchedule[]>;
			workSchedules = response.data || [];
		} catch {
			workSchedules = [];
		}
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadDepartments();
	}

	function openCreateForm() {
		formTitle = 'Tambah Departemen';
		editingId = null;
		form = { name: '', code: '', parent_id: '', head_id: '', work_schedule_id: '', description: '', is_active: true };
		formError = '';
		showForm = true;
	}

	function openEditForm(dept: Department) {
		formTitle = 'Edit Departemen';
		editingId = dept.id;
		form = {
			name: dept.name,
			code: dept.code,
			parent_id: '',
			head_id: '',
			work_schedule_id: '',
			description: dept.description,
			is_active: dept.is_active,
		};
		formError = '';
		showForm = true;

		// Load full detail for parent_id, head_id, work_schedule_id
		deptApi.get(dept.id).then((resp) => {
			if (resp.data) {
				const d = resp.data as Department;
				form.head_id = d.head_id || '';
				form.work_schedule_id = d.work_schedule_id || '';
				form.is_active = d.is_active ?? true;
			}
		}).catch((err: unknown) => {
			console.error('Gagal memuat detail departemen:', err);
		});
	}

	function cancelForm() {
		showForm = false;
		formError = '';
	}

	async function handleSave() {
		// Validasi
		if (!form.name.trim()) {
			formError = 'Nama departemen harus diisi';
			return;
		}
		if (!form.code.trim()) {
			formError = 'Kode departemen harus diisi';
			return;
		}

		isSaving = true;
		formError = '';
		try {
			const payload: Record<string, unknown> = {
				name: form.name.trim(),
				code: form.code.trim().toUpperCase(),
				description: form.description.trim(),
			};
			if (form.parent_id) payload.parent_id = form.parent_id;
			if (form.head_id) payload.head_id = form.head_id;
			if (form.work_schedule_id) payload.work_schedule_id = form.work_schedule_id;
			if (editingId) payload.is_active = form.is_active;

			if (editingId) {
				await deptApi.update(editingId, payload);
			} else {
				await deptApi.create(payload);
			}
			cancelForm();
			loadDepartments();
			loadAllDepts();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menyimpan departemen';
		} finally {
			isSaving = false;
		}
	}

	function confirmDelete(id: string, name: string) {
		deletingId = id;
		deletingName = name;
		showDeleteConfirm = true;
	}

	function cancelDelete() {
		showDeleteConfirm = false;
		deletingId = null;
		deletingName = '';
	}

	async function handleDelete() {
		if (!deletingId) return;
		isSaving = true;
		try {
			await deptApi.remove(deletingId);
			showDeleteConfirm = false;
			deletingId = null;
			deletingName = '';
			loadDepartments();
			loadAllDepts();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menghapus departemen';
			showDeleteConfirm = false;
		} finally {
			isSaving = false;
		}
	}

	// Debounce search
	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			searchQuery = target.value;
			page = 1;
			loadDepartments();
		}, 400);
	}

	function openSchedulePicker(deptId: string, currentScheduleId: string) {
		schedulePickerDeptId = deptId;
		selectedScheduleId = currentScheduleId || '';
		gridApi?.refreshCells();
	}

	function closeSchedulePicker() {
		schedulePickerDeptId = null;
		selectedScheduleId = '';
		gridApi?.refreshCells();
	}

	async function handleScheduleChange() {
		if (!schedulePickerDeptId) return;
		isSavingSchedule = true;
		try {
			await deptApi.updateWorkSchedule(schedulePickerDeptId, selectedScheduleId);
			closeSchedulePicker();
			loadDepartments();
		} catch (err: unknown) {
			formError = (err as { message?: string }).message || 'Gagal mengupdate jadwal';
		} finally {
			isSavingSchedule = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

</script>

<div class="w-full">
	<!-- Header Section -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Departemen</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola seluruh departemen dan sub-departemen perusahaan</p>
		</div>
		{#if !showForm && hasPermission('department', 'create')}
			<button
				onclick={openCreateForm}
				class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Tambah Departemen
			</button>
		{/if}
	</div>

	<!-- Summary Stats (sembunyikan saat form aktif) -->
	{#if !showForm && !isLoading && !errorMessage && departments.length > 0}
		<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total</span>
					<div class="w-7 h-7 rounded-lg bg-indigo-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{total}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Aktif</span>
					<div class="w-7 h-7 rounded-lg bg-emerald-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{totalActiveDepts}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Sub-Dept</span>
					<div class="w-7 h-7 rounded-lg bg-amber-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{totalSubDepts}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total Karyawan</span>
					<div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{totalEmployeesAll}</p>
			</div>
		</div>
	{/if}

	<!-- Search & Filter Bar (sembunyikan saat form aktif) -->
	{#if !showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="search"
					value={searchQuery}
					placeholder="Cari berdasarkan nama atau kode..."
					oninput={onSearchInput}
					class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400"
					aria-label="Cari departemen"
				/>
			</div>
			<div class="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
				{total > 0 ? `${total} departemen ditemukan` : ''}
			</div>
		</div>
	{/if}

	<!-- Inline Form (gantikan tabel) -->
	<div class:hidden={!showForm}>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<!-- Form Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{formTitle}</h2>
				<button
					onclick={cancelForm}
					class="p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer"
					aria-label="Tutup"
				>
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Form Body -->
			<div class="px-6 py-5 space-y-4">
				{#if formError}
					<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm px-4 py-2.5 rounded-lg">
						{formError}
					</div>
				{/if}

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="dept-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Departemen <span class="text-red-500">*</span></label>
						<input id="dept-name"
							type="text"
							bind:value={form.name}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
							placeholder="Contoh: Teknologi Informasi"
						/>
					</div>

					<div>
						<label for="dept-code" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Kode Departemen <span class="text-red-500">*</span></label>
						<input id="dept-code"
							type="text"
							bind:value={form.code}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400 uppercase"
							placeholder="Contoh: IT"
						/>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="dept-parent" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Departemen Induk</label>
						<select id="dept-parent"
							bind:value={form.parent_id}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900"
						>
							<option value="">Tidak ada (departemen utama)</option>
							{#each allDepts as d (d)}
								{#if d.id !== editingId}
									<option value={d.id}>{d.name} ({d.code})</option>
								{/if}
							{/each}
						</select>
					</div>

					<div>
						<label for="dept-schedule" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Jadwal Kerja</label>
						<select id="dept-schedule"
							bind:value={form.work_schedule_id}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900"
						>
							<option value="">Pilih jadwal (opsional)</option>
							{#each workSchedules as ws (ws)}
								<option value={ws.id}>{ws.name}</option>
							{/each}
						</select>
					</div>
				</div>

				<div>
					<label for="dept-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
					<textarea id="dept-desc"
						bind:value={form.description}
						rows="3"
						class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400 resize-none"
						placeholder="Deskripsi departemen (opsional)"
					></textarea>
				</div>

				{#if editingId}
					<div class="flex items-center gap-3">
						<label class="relative inline-flex items-center cursor-pointer">
							<input type="checkbox" bind:checked={form.is_active} class="sr-only peer" />
							<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[#1A56DB]/20 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#1A56DB]"></div>
							<span class="ms-2 text-sm font-medium text-gray-700 dark:text-gray-300">Status Aktif</span>
						</label>
					</div>
				{/if}
			</div>

			<!-- Form Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<button
					onclick={cancelForm}
					class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer"
				>
					Batal
				</button>
				<button
					onclick={handleSave}
					disabled={isSaving}
					class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer"
				>
					{#if isSaving}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
					{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Departemen'}
				</button>
			</div>
		</div>
	</div>
	<div class:hidden={showForm}>
		<!-- Table Card -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<PulseLoader variant="table-row" count={5} />
			{:else if errorMessage}
				<!-- Error State -->
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
						<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
						</svg>
					</div>
					<p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
					<button onclick={loadDepartments} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">
						Muat Ulang
					</button>
				</div>
			{:else if departments.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada data departemen"
					description={searchQuery ? `Tidak ditemukan departemen dengan kata kunci "${searchQuery}"` : 'Data departemen akan muncul di sini setelah ditambahkan.'}
				/>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>

			<!-- Mobile Cards -->
			<div class="md:hidden space-y-3">
				<PullToRefresh onRefresh={loadDepartments}>
				{#each departments as dept (dept)}
					<MobileCard
						title={dept.name}
						subtitle={dept.code}
						avatar={getInitials(dept.name)}
						avatarColor={getAvatarTheme('department').gradientClasses}
						badges={[{ label: dept.is_active ? 'Aktif' : 'Nonaktif', color: dept.is_active ? 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800' : 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800' }]}
					>
						{#snippet footer()}
							<div class="flex items-center gap-2">
								{#if hasPermission('department', 'update')}
									<button onclick={() => openEditForm(dept)} class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95">Edit</button>
								{/if}
								{#if hasPermission('department', 'delete')}
									<button onclick={() => confirmDelete(dept.id, dept.name)} class="flex-1 py-2 text-xs font-medium text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95">Hapus</button>
								{/if}
							</div>
						{/snippet}
					</MobileCard>
				{/each}
				</PullToRefresh>
			</div>

				<!-- Pagination -->
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-800 bg-gray-50/30 dark:bg-gray-900/30">
					<div class="text-xs text-gray-500 dark:text-gray-400">
						Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span>
					</div>
					<div class="flex items-center gap-1.5">
						<button
							onclick={() => goToPage(page - 1)}
							disabled={page <= 1}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer"
						>
							Sebelumnya
						</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button
									onclick={() => goToPage(pageNum)}
									class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}"
								>
									{pageNum}
								</button>
							{/if}
						{/each}
						<button
							onclick={() => goToPage(page + 1)}
							disabled={page >= totalPages}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer"
						>
							Selanjutnya
						</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Delete Confirmation Modal (tetap modal, cuma buat konfirmasi) -->
<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div
		onclick={cancelDelete}
		onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelDelete(); }}
		role="presentation"
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4"
	>
		<div
			onclick={(e) => e.stopPropagation()}
			role="dialog"
			aria-modal="true"
			aria-label="Hapus departemen"
			tabindex="-1"
			class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm"
		>
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
					<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
					</svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Hapus Departemen</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-1">
					Apakah Anda yakin ingin menghapus departemen
				</p>
				<p class="text-sm font-medium text-gray-900 dark:text-white mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 dark:text-gray-500 mb-6">Tindakan ini tidak dapat dibatalkan.</p>
				<div class="flex items-center justify-center gap-3">
					<button
						onclick={cancelDelete}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer"
					>
						Batal
					</button>
					<button
						onclick={handleDelete}
						disabled={isSaving}
						class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer"
					>
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
