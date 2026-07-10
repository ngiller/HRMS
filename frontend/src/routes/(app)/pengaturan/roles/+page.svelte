<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { roles as rolesApi } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type RoleSummary = {
		id: string;
		name: string;
		slug: string;
		description: string;
		is_system_role: boolean;
		is_active: boolean;
		employee_count: number;
		created_at: string;
	};

	type RoleDetail = {
		id: string;
		name: string;
		slug: string;
		description: string;
		permissions: Record<string, Record<string, boolean>>;
		is_system_role: boolean;
		is_active: boolean;
		created_at: string;
		updated_at: string;
	};

	type PermissionModule = {
		module: string;
		label: string;
		actions: string[];
	};

	let roles = $state<RoleSummary[]>([]);
	let permissionModules = $state<PermissionModule[]>([]);
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
	type RoleForm = {
		name: string;
		slug: string;
		description: string;
		is_active: boolean;
		permissions: Record<string, Record<string, boolean>>;
	};

	let form = $state<RoleForm>({
		name: '', slug: '', description: '', is_active: true, permissions: {}
	});
	let formError = $state('');
	let isSaving = $state(false);

	// Delete confirm (tetap modal kecil)
	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');

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
			field: 'name', headerName: 'Role', minWidth: 240, flex: 1,
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const desc = params.data?.description || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 flex items-center justify-center text-xs font-semibold text-purple-600 shrink-0 ring-1 ring-purple-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div>${desc ? `<div class="text-xs text-gray-400 truncate max-w-48">${desc}</div>` : ''}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'slug', headerName: 'Slug', minWidth: 140,
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
				if (!params.value) return '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-gray-100 text-gray-600">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'employee_count', headerName: 'Karyawan', minWidth: 100, maxWidth: 120,
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
				if (params.value == null) return '';
				return `<span class="text-sm font-medium text-gray-700 tabular-nums">${params.value}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-center',
		},
		{
			field: 'is_system_role', headerName: 'Tipe', minWidth: 110,
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
				if (params.value) {
					return '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-50 text-purple-700 ring-1 ring-purple-600/20">Sistem</span>';
				}
				return '<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700 ring-1 ring-blue-600/20">Kustom</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'is_active', headerName: 'Status', minWidth: 110,
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
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
			cellRenderer: (params: AgGridCellParams<RoleSummary>) => {
				const role = params.data;
				if (!role) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				if (hasPermission('user_management', 'update')) {
					const editBtn = createActionButton(iconEdit(),
						'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
						'Edit role',
						() => openEditForm(role)
					);
					container.appendChild(editBtn);
				}
				if (hasPermission('user_management', 'delete') && !role.is_system_role) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Hapus role',
						() => confirmDelete(role.id, role.name)
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
		if (roles.length > 0 && gridContainer && !showForm) {
			if (!gridApi) {
				gridApi = agGridModule!.createGrid(gridContainer, gridOptions) as GridApi;
			}
			gridApi.updateGridOptions({ rowData: roles });
		}
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});

	// Summary stats
	let totalSystem = $state(0);
	let totalCustom = $state(0);
	let totalEmployeesInRoles = $state(0);

	onMount(async () => { const m = await getAgGrid(); agGridModule = m;
		loadRoles();
		loadPermissionTemplate();
	});

	async function loadRoles() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response: ApiResponse<RoleSummary[]> = await rolesApi.list(page, perPage, searchQuery) as ApiResponse<RoleSummary[]>;
			const data = response.data || [];
			roles = data;
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);

			totalSystem = data.filter((r: RoleSummary) => r.is_system_role).length;
			totalCustom = data.filter((r: RoleSummary) => !r.is_system_role).length;
			totalEmployeesInRoles = data.reduce((sum: number, r: RoleSummary) => sum + (r.employee_count || 0), 0);
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal memuat data role';
			console.error('Role list error:', error);
		} finally {
			isLoading = false;
		}
	}

	async function loadPermissionTemplate() {
		try {
			const response: ApiResponse<PermissionModule[]> = await rolesApi.getPermissionTemplate() as ApiResponse<PermissionModule[]>;
			permissionModules = response.data || [];
		} catch {
			permissionModules = [];
		}
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadRoles();
	}

	function openCreateForm() {
		formTitle = 'Tambah Role';
		editingId = null;
		form = { name: '', slug: '', description: '', is_active: true, permissions: {} };
		formError = '';
		showForm = true;
	}

	function openEditForm(role: RoleSummary) {
		formTitle = 'Edit Role';
		editingId = role.id;
		form = { name: role.name, slug: role.slug, description: role.description, is_active: role.is_active, permissions: {} };
		formError = '';
		showForm = true;

		// Load detail for permissions
		rolesApi.get(role.id).then((resp: any) => {
			if (resp.data) {
				form.name = resp.data.name;
				form.slug = resp.data.slug;
				form.description = resp.data.description || '';
				form.is_active = resp.data.is_active ?? true;
				form.permissions = resp.data.permissions || {};
			}
		}).catch((err: unknown) => {
			console.error('Gagal memuat detail role:', err);
		});
	}

	function cancelForm() {
		showForm = false;
		formError = '';
	}

	function isPermissionChecked(module: string, action: string): boolean {
		return form.permissions[module]?.[action] === true;
	}

	function togglePermission(module: string, action: string) {
		if (!form.permissions[module]) {
			form.permissions[module] = {};
		}
		form.permissions[module][action] = !form.permissions[module][action];
		// Trigger reactivity
		form = { ...form, permissions: { ...form.permissions } };
	}

	function setModulePermissions(module: string, checked: boolean) {
		const modDef = permissionModules.find(m => m.module === module);
		if (!modDef) return;
		if (!form.permissions[module]) {
			form.permissions[module] = {};
		}
		for (const action of modDef.actions) {
			form.permissions[module][action] = checked;
		}
		form = { ...form, permissions: { ...form.permissions } };
	}

	function isModuleAllChecked(module: string): boolean {
		const modDef = permissionModules.find(m => m.module === module);
		if (!modDef) return false;
		return modDef.actions.every(a => form.permissions[module]?.[a] === true);
	}

	async function handleSave() {
		if (!form.name.trim()) { formError = 'Nama role harus diisi'; return; }
		if (!form.slug.trim()) { formError = 'Slug role harus diisi'; return; }

		isSaving = true;
		formError = '';
		try {
			const payload: Record<string, unknown> = {
				name: form.name.trim(),
				slug: form.slug.trim().toLowerCase().replace(/\s+/g, '_'),
				description: form.description.trim(),
				permissions: form.permissions,
			};
			if (editingId) payload.is_active = form.is_active;

			if (editingId) {
				await rolesApi.update(editingId, payload);
			} else {
				await rolesApi.create(payload);
			}
			cancelForm();
			loadRoles();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menyimpan role';
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
			await rolesApi.remove(deletingId);
			showDeleteConfirm = false;
			deletingId = null;
			deletingName = '';
			loadRoles();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menghapus role';
			showDeleteConfirm = false;
		} finally {
			isSaving = false;
		}
	}

	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			searchQuery = target.value;
			page = 1;
			loadRoles();
		}, 400);
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
	<!-- Header Section -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Role Management</h1>
			<p class="text-sm text-gray-500 mt-0.5">Atur hak akses role dan permission pengguna</p>
		</div>
		{#if !showForm && hasPermission('user_management', 'create')}
			<button
				onclick={openCreateForm}
				class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
				</svg>
				Tambah Role
			</button>
		{/if}
	</div>

	<!-- Summary Stats (sembunyikan saat form aktif) -->
	{#if !showForm && !isLoading && !errorMessage && roles.length > 0}
		<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
			<div class="bg-white border border-gray-200 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500">Total</span>
					<div class="w-7 h-7 rounded-lg bg-purple-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-purple-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 mt-1 tabular-nums">{total}</p>
			</div>
			<div class="bg-white border border-gray-200 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500">Sistem</span>
					<div class="w-7 h-7 rounded-lg bg-amber-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 mt-1 tabular-nums">{totalSystem}</p>
			</div>
			<div class="bg-white border border-gray-200 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500">Kustom</span>
					<div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 mt-1 tabular-nums">{totalCustom}</p>
			</div>
			<div class="bg-white border border-gray-200 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500">Total Karyawan</span>
					<div class="w-7 h-7 rounded-lg bg-emerald-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 mt-1 tabular-nums">{totalEmployeesInRoles}</p>
			</div>
		</div>
	{/if}

	<!-- Search & Filter Bar (sembunyikan saat form aktif) -->
	{#if !showForm}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="search"
					value={searchQuery}
					placeholder="Cari berdasarkan nama atau slug..."
					oninput={onSearchInput}
					class="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white transition placeholder:text-gray-400"
					aria-label="Cari role"
				/>
			</div>
			<div class="flex items-center gap-2 text-xs text-gray-400">
				{total > 0 ? `${total} role ditemukan` : ''}
			</div>
		</div>
	{/if}

	<!-- Inline Form (gantikan tabel) -->
	<div class:hidden={!showForm}>
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<!-- Form Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">{formTitle}</h2>
				<button
					onclick={cancelForm}
					class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer"
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
					<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>
				{/if}

				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="role-name" class="block text-sm font-medium text-gray-700 mb-1.5">Nama Role <span class="text-red-500">*</span></label>
						<input id="role-name" type="text" bind:value={form.name}
							class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
							placeholder="Contoh: HR Manager" />
					</div>
					<div>
						<label for="role-slug" class="block text-sm font-medium text-gray-700 mb-1.5">Slug <span class="text-red-500">*</span></label>
						<input id="role-slug" type="text" bind:value={form.slug}
							class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
							placeholder="Contoh: hr_manager" />
					</div>
				</div>

				<div>
					<label for="role-desc" class="block text-sm font-medium text-gray-700 mb-1.5">Deskripsi</label>
					<textarea id="role-desc" bind:value={form.description} rows="2"
						class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400 resize-none"
						placeholder="Deskripsi role (opsional)"></textarea>
				</div>

				{#if editingId}
					<div class="flex items-center gap-3">
						<label class="relative inline-flex items-center cursor-pointer">
							<input type="checkbox" bind:checked={form.is_active} class="sr-only peer" />
							<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-[#1A56DB]/20 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#1A56DB]"></div>
							<span class="ms-2 text-sm font-medium text-gray-700">Status Aktif</span>
						</label>
					</div>
				{/if}

				<!-- Permissions Editor -->
				<div>
					<div class="flex items-center justify-between mb-3">
						<h3 class="text-sm font-semibold text-gray-900">Permissions</h3>
						<span class="text-xs text-gray-400">Centang modul & action yang diizinkan</span>
					</div>
					<div class="space-y-2 max-h-80 overflow-y-auto border border-gray-200 rounded-xl p-3">
						{#each permissionModules as mod}
							<div class="border border-gray-100 rounded-lg p-3 hover:bg-gray-50/50 transition">
								<div class="flex items-center justify-between mb-2">
									<label class="flex items-center gap-2 text-sm font-medium text-gray-700 cursor-pointer">
										<input
											type="checkbox"
											checked={isModuleAllChecked(mod.module)}
											onchange={(e) => setModulePermissions(mod.module, (e.target as HTMLInputElement).checked)}
											class="w-4 h-4 rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]"
										/>
										{mod.label}
									</label>
									<span class="text-xs text-gray-400">{mod.module}</span>
								</div>
								<div class="flex flex-wrap gap-2 ml-6">
									{#each mod.actions as action}
										<label class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-md text-xs cursor-pointer transition {isPermissionChecked(mod.module, action) ? 'bg-blue-50 text-blue-700 ring-1 ring-blue-200' : 'bg-gray-50 text-gray-500 hover:bg-gray-100'}">
											<input
												type="checkbox"
												checked={isPermissionChecked(mod.module, action)}
												onchange={() => togglePermission(mod.module, action)}
												class="hidden"
											/>
											{action === 'create' ? 'Buat' : action === 'read' ? 'Lihat' : action === 'update' ? 'Ubah' : action === 'delete' ? 'Hapus' : action === 'approve' ? 'Setujui' : action}
										</label>
									{/each}
								</div>
							</div>
						{/each}
						{#if permissionModules.length === 0}
							<p class="text-sm text-gray-400 text-center py-4">Template permission tidak tersedia</p>
						{/if}
					</div>
				</div>
			</div>

			<!-- Form Footer -->
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					Batal
				</button>
				<button onclick={handleSave} disabled={isSaving}
					class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
					{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Role'}
				</button>
			</div>
		</div>
	</div>
	<div class:hidden={showForm}>
		<!-- Table Card -->
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<PulseLoader variant="table-row" count={4} />
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center">
						<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
						</svg>
					</div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={loadRoles} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if roles.length === 0}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 flex items-center justify-center">
						<svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" />
						</svg>
					</div>
					<h3 class="text-sm font-semibold text-gray-900 mb-1">Belum ada data role</h3>
					<p class="text-sm text-gray-500">Role akan muncul di sini setelah ditambahkan.</p>
				</div>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>

				<!-- Mobile Cards -->
				<div class="md:hidden divide-y divide-gray-100">
					{#each roles as role}
					{@const canEdit = hasPermission('user_management', 'update')}
					{@const canDelete = hasPermission('user_management', 'delete') && !role.is_system_role}
						<div class="p-4 hover:bg-blue-50/40 transition-colors">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-purple-50 to-purple-100 flex items-center justify-center text-xs font-semibold text-purple-600 shrink-0 ring-1 ring-purple-200">
									{getInitials(role.name)}
								</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900 truncate">{role.name}</div>
									<div class="text-xs text-gray-400">{role.slug}</div>
								</div>
								<div class="flex items-center gap-1 shrink-0">
									{#if canEdit}
										<button onclick={() => openEditForm(role)} class="p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer" aria-label="Edit role">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
											</svg>
										</button>
									{/if}
									{#if canDelete}
										<button onclick={() => confirmDelete(role.id, role.name)} class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer" aria-label="Hapus role">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" />
											</svg>
										</button>
									{/if}
								</div>
							</div>
							<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-400 ml-13">
								<span>{role.employee_count} karyawan</span>
								<span class="w-1 h-1 bg-gray-300 rounded-full shrink-0"></span>
								{#if role.is_active}
									<span class="text-emerald-600 font-medium">Aktif</span>
								{:else}
									<span class="text-red-500 font-medium">Nonaktif</span>
								{/if}
							</div>
						</div>
					{/each}
				</div>

				<!-- Pagination -->
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30">
					<div class="text-xs text-gray-500">
						Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700">{total}</span>
					</div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pn = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pn <= totalPages}
								<button onclick={() => goToPage(pn)}
									class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pn === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pn}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<!-- Delete Confirmation Modal (tetap modal, cuma buat konfirmasi) -->
<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelDelete} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelDelete(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Hapus role" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
					</svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Hapus Role</h3>
				<p class="text-sm text-gray-500 mb-1">Apakah Anda yakin ingin menghapus role</p>
				<p class="text-sm font-medium text-gray-900 mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 mb-6">Tindakan ini tidak dapat dibatalkan.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={isSaving}
						class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
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
