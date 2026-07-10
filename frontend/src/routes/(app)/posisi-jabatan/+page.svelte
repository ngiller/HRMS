<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { positions as api, departments as deptApi, positionGrades as pgApi } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams, Department, PositionGrade } from '$lib/types.js';
	type Position = {
		id: string;
		name: string;
		department_name: string;
		department_id: string;
		grade_name: string;
		grade_id: string;
		description: string;
		is_active: boolean;
		created_at: string;
	};

	type Form = {
		name: string;
		department_id: string;
		grade_id: string;
		description: string;
	};

	let items = $state<Position[]>([]);
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
	let form = $state<Form>({ name: '', department_id: '', grade_id: '', description: '' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');

	let departments = $state<Department[]>([]);
	let positionGrades = $state<PositionGrade[]>([]);
	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response: ApiResponse<Position[]> = await api.list(page, perPage, searchQuery) as ApiResponse<Position[]>;
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal memuat data';
		} finally { isLoading = false; }
	}

	async function loadDepartments() {
		try { const r: ApiResponse<Department[]> = await deptApi.getAll() as ApiResponse<Department[]>; departments = r.data || []; }
		catch { departments = []; }
	}

	async function loadGrades() {
		try { const r: ApiResponse<PositionGrade[]> = await pgApi.getAll() as ApiResponse<PositionGrade[]>; positionGrades = r.data || []; }
		catch { positionGrades = []; }
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	function openCreateForm() {
		formTitle = 'Tambah Posisi Jabatan';
		editingId = null;
		form = { name: '', department_id: '', grade_id: '', description: '' };
		formError = '';
		showForm = true;
	}

	function openEditForm(item: Position) {
		formTitle = 'Edit Posisi Jabatan';
		editingId = item.id;
		form = { name: item.name, department_id: item.department_id, grade_id: item.grade_id, description: item.description };
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.name.trim()) { formError = 'Nama posisi jabatan harus diisi'; return; }
		if (!form.department_id) { formError = 'Departemen harus dipilih'; return; }

		isSaving = true;
		formError = '';
		try {
			const payload: Record<string, unknown> = { name: form.name.trim(), department_id: form.department_id, description: form.description.trim() };
			if (form.grade_id) payload.grade_id = form.grade_id;
			if (editingId) { await api.update(editingId, payload); }
			else { await api.create(payload); }
			cancelForm();
			load();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	function confirmDelete(id: string, name: string) { deletingId = id; deletingName = name; showDeleteConfirm = true; }
	function cancelDelete() { showDeleteConfirm = false; deletingId = null; deletingName = ''; }

	async function handleDelete() {
		if (!deletingId) return;
		isSaving = true;
		try { await api.remove(deletingId); showDeleteConfirm = false; deletingId = null; deletingName = ''; load(); }
		catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menghapus data'; showDeleteConfirm = false; }
		finally { isSaving = false; }
	}

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
		btn.onclick = (e) => {
			e.stopPropagation();
			onClick();
		};
		return btn;
	}

	const columnDefs: ColDef[] = [
		{
			field: 'name', headerName: 'Posisi', minWidth: 220, flex: 1,
			valueGetter: (params) => params.data?.name || '',
			cellRenderer: (params: AgGridCellParams<Position>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const desc = params.data?.description || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-amber-50 to-amber-100 flex items-center justify-center text-xs font-semibold text-amber-600 ring-1 ring-amber-200 shrink-0">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div>${desc ? `<div class="text-xs text-gray-400 truncate max-w-48">${desc}</div>` : ''}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'department_name', headerName: 'Departemen', minWidth: 150,
			valueGetter: (params) => params.data?.department_name || '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600 dark:text-gray-400',
		},
		{
			field: 'grade_name', headerName: 'Golongan', minWidth: 150,
			valueGetter: (params) => params.data?.grade_name || '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600 dark:text-gray-400',
		},
		{
			field: 'is_active', headerName: 'Status', minWidth: 100, maxWidth: 130,
			cellRenderer: (params: AgGridCellParams<Position>) => {
				const active = params.value;
				return active
					? '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ring-1 ring-inset bg-green-50 text-green-700 ring-green-600/20">Aktif</span>'
					: '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ring-1 ring-inset bg-gray-100 text-gray-500 ring-gray-600/20">Nonaktif</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 130,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'id', headerName: '', minWidth: 100, maxWidth: 100,
			cellRenderer: (params: AgGridCellParams<Position>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				if (hasPermission('position', 'update')) {
					const editBtn = createActionButton(iconEdit(),
						'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
						'Edit posisi',
						() => openEditForm(item)
					);
					container.appendChild(editBtn);
				}
				if (hasPermission('position', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Hapus posisi',
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
		onGridReady: (params) => {
			gridApi = params.api;
		},
	};

	$effect(() => {
		if (!gridContainer && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (gridContainer && !showForm) {
			if (!gridApi) {
				gridApi = agGridModule!.createGrid(gridContainer, gridOptions) as GridApi;
			}
			gridApi.updateGridOptions({ rowData: items });
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
		loadDepartments();
		loadGrades();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Posisi Jabatan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola posisi/jabatan karyawan di setiap departemen</p>
		</div>
		{#if !showForm && hasPermission('position', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Posisi
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				<input type="search" value={searchQuery} placeholder="Cari posisi jabatan..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
			</div>
			<div class="text-xs text-gray-400 dark:text-gray-500">{total > 0 ? `${total} posisi ditemukan` : ''}</div>
		</div>
	{/if}

	<div class:hidden={!showForm}>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{formTitle}</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="pos-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Posisi <span class="text-red-500">*</span></label>
						<input id="pos-name" type="text" bind:value={form.name} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Contoh: Frontend Developer" />
					</div>
					<div>
						<label for="pos-dept" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Departemen <span class="text-red-500">*</span></label>
						<select id="pos-dept" bind:value={form.department_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
							<option value="">Pilih departemen</option>
							{#each departments as d}<option value={d.id}>{d.name}</option>{/each}
						</select>
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="pos-grade" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Golongan Jabatan</label>
						<select id="pos-grade" bind:value={form.grade_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
							<option value="">Pilih golongan (opsional)</option>
							{#each positionGrades as g}<option value={g.id}>{g.name} (Level {g.level})</option>{/each}
						</select>
					</div>
				</div>
				<div>
					<label for="pos-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
					<textarea id="pos-desc" bind:value={form.description} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Deskripsi posisi (opsional)"></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Posisi'}
				</button>
			</div>
		</div>
	</div>
	<div class:hidden={showForm}>
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<PulseLoader variant="table-row" count={5} />
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada data posisi jabatan"
					description={searchQuery ? `Tidak ditemukan dengan kata kunci "${searchQuery}"` : 'Data posisi jabatan akan muncul setelah ditambahkan.'}
				/>
			{:else}
				<div class="hidden md:block">
					<div class="ag-theme-quartz" style="width:100%;" bind:this={gridContainer}></div>
				</div>
				<!-- Mobile cards -->
				<PullToRefresh onRefresh={load}>
				<div class="md:hidden space-y-3">
					{#each items as item}
						<MobileCard
							title={item.name}
							subtitle={item.description || ''}
							avatar={getInitials(item.name)}
							avatarColor={getAvatarTheme('position').gradientClasses}
							badges={[{ label: item.is_active ? 'Aktif' : 'Nonaktif', color: item.is_active ? 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800' : 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700' }]}
						>
							{#snippet children()}
								<div class="flex items-center gap-3 text-xs text-gray-500 dark:text-gray-400">
									<span class="flex items-center gap-1.5">{item.department_name || '-'}</span>
									{#if item.grade_name}
										<span>· {item.grade_name}</span>
									{/if}
								</div>
							{/snippet}
							{#snippet footer()}
								<div class="flex items-center gap-2">
									{#if hasPermission('position', 'update')}
										<button onclick={() => openEditForm(item)} class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95">Edit</button>
									{/if}
									{#if hasPermission('position', 'delete')}
										<button onclick={() => confirmDelete(item.id, item.name)} class="flex-1 py-2 text-xs font-medium text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95">Hapus</button>
									{/if}
								</div>
							{/snippet}
						</MobileCard>
					{/each}
				</div>
				</PullToRefresh>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-800 bg-gray-50/30 dark:bg-gray-900/30">
					<div class="text-xs text-gray-500 dark:text-gray-400">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelDelete} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelDelete(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Hapus posisi jabatan" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Hapus Posisi Jabatan</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Apakah Anda yakin ingin menghapus</p>
				<p class="text-sm font-medium text-gray-900 dark:text-white mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 dark:text-gray-500 mb-6">Data yang sudah dihapus tidak dapat dikembalikan.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Ya, Hapus
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
