<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { holidays as api } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type Holiday = {
		id: string;
		date: string;
		name: string;
		holiday_type: string;
		is_recurring_yearly: boolean;
		description: string;
		is_active: boolean;
	};

	type FormData = {
		date: string;
		name: string;
		holiday_type: string;
		is_recurring_yearly: boolean;
		description: string;
	};

	const holidayTypes: Record<string, string> = {
		national: 'Libur Nasional', joint: 'Cuti Bersama', company: 'Libur Perusahaan',
	};
	const typeColors: Record<string, string> = {
		national: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900/30 dark:text-blue-200 dark:ring-blue-800',
		joint: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900/30 dark:text-green-200 dark:ring-green-800',
		company: 'bg-purple-50 text-purple-700 ring-purple-200 dark:bg-purple-900/30 dark:text-purple-200 dark:ring-purple-800',
	};

	const monthNames = ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni',
		'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'];

	let items = $state<Holiday[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let yearFilter = $state(new Date().getFullYear());
	let typeFilter = $state('');
	let monthFilter = $state<number | null>(null);
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let isEditing = $state(false);
	let editId = $state<string | null>(null);
	let form = $state<FormData>({ date: '', name: '', holiday_type: 'national', is_recurring_yearly: false, description: '' });
	let formError = $state('');
	let isSaving = $state(false);

	// ── AG Grid ──
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
			field: 'date', headerName: 'Tanggal', minWidth: 130,
			valueFormatter: (params: AgGridValueParams) => params.value ? formatDate(params.value as string) : '',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm font-medium text-gray-900 dark:text-white tabular-nums',
		},
		{
			field: 'name', headerName: 'Nama Hari Libur', minWidth: 250, flex: 1,
			cellRenderer: (params: AgGridCellParams<Holiday>) => {
				const name = params.value as string || '';
				const desc = params.data?.description || '';
				return `<div class="text-sm font-medium text-gray-900 dark:text-white">${name}${desc ? `<span class="text-xs text-gray-400 dark:text-gray-500 ml-2">${desc}</span>` : ''}</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
		},
		{
			field: 'holiday_type', headerName: 'Tipe', minWidth: 140,
			cellRenderer: (params: AgGridCellParams<Holiday>) => {
				const t = params.value as string || '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${typeColors[t] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${holidayTypes[t] || t}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
		},
		{
			field: 'is_recurring_yearly', headerName: 'Tahunan', minWidth: 90, maxWidth: 100,
			cellRenderer: (params: any) => params.value
				? '<span class="inline-flex items-center text-xs text-green-600 dark:text-green-400 font-medium"><svg class="w-3.5 h-3.5 mr-1" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>Ya</span>'
				: '<span class="text-xs text-gray-400 dark:text-gray-500">Tidak</span>',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			sortable: false, filter: false,
		},
		{
			field: 'id', headerName: '', minWidth: 100, maxWidth: 100,
			cellRenderer: (params: AgGridCellParams<Holiday>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				if (hasPermission('announcement', 'update')) {
					const editBtn = createActionButton(iconEdit(),
						'p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/30 transition cursor-pointer',
						'Edit', () => openEdit(item.id));
					container.appendChild(editBtn);
				}

				if (hasPermission('announcement', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition cursor-pointer',
						'Hapus', () => handleDelete(item.id));
					container.appendChild(deleteBtn);
				}

				return container;
			},
			sortable: false, filter: false, resizable: false,
		},
	];

	const gridOptions: GridOptions = {
		columnDefs, defaultColDef, rowHeight: 56, headerHeight: 44,
		animateRows: true, domLayout: 'autoHeight',
		suppressDragLeaveHidesColumns: true, suppressRowHoverHighlight: false,
		enableCellTextSelection: true, pagination: false, theme: 'legacy',
		onGridReady: (params) => { gridApi = params.api; },
	};

	$effect(() => {
		if (!gridContainer && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});


	$effect(() => {
		if (gridContainer && !showForm) {
			if (!gridApi && agGridModule) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			if (gridApi) {
				const filteredItems = monthFilter !== null ? items.filter(i => new Date(i.date + 'T00:00:00').getMonth() === monthFilter) : items;
				gridApi.updateGridOptions({ rowData: filteredItems });
			}
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
	});

	onDestroy(() => { gridApi?.destroy(); gridApi = null; });
	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true; errorMessage = '';
		try {
			const response = await api.list(page, perPage, yearFilter, typeFilter) as ApiResponse<Holiday[]>;
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
		formTitle = 'Tambah Hari Libur';
		isEditing = false; editId = null;
		form = { date: '', name: '', holiday_type: 'national', is_recurring_yearly: false, description: '' };
		formError = ''; showForm = true;
	}

	async function openEdit(id: string) {
		isEditing = true; editId = id; formTitle = 'Edit Hari Libur';
		formError = ''; showForm = true;
		try {
			const response = await api.get(id) as ApiResponse<Holiday>;
			const d = response.data || {} as Holiday;
			form = {
				date: d.date || '',
				name: d.name || '',
				holiday_type: d.holiday_type || 'national',
				is_recurring_yearly: d.is_recurring_yearly || false,
				description: d.description || '',
			};
		} catch (_) { formError = 'Gagal memuat data'; showForm = false; }
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.date) { formError = 'Tanggal harus diisi'; return; }
		if (!form.name.trim()) { formError = 'Nama hari libur harus diisi'; return; }

		isSaving = true; formError = '';
		try {
			const payload: Record<string, unknown> = {
				date: form.date,
				name: form.name.trim(),
				holiday_type: form.holiday_type,
				is_recurring_yearly: form.is_recurring_yearly,
				description: form.description.trim(),
			};
			if (isEditing && editId) {
				await api.update(editId, payload);
			} else {
				await api.create(payload);
			}
			cancelForm(); load();
		} catch (error: any) { formError = error.message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function handleDelete(id: string) {
		if (!confirm('Hapus hari libur ini?')) return; isSaving = true;
		try { await api.remove(id); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menghapus'; }
		finally { isSaving = false; }
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		const d = new Date(dateStr + 'T00:00:00');
		return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	function formatShortDate(dateStr: string): string {
		if (!dateStr) return '-';
		const d = new Date(dateStr + 'T00:00:00');
		return d.toLocaleDateString('id-ID', { weekday: 'short', day: 'numeric', month: 'short' });
	}

	function getMonth(dateStr: string): string {
		if (!dateStr) return '';
		const m = new Date(dateStr + 'T00:00:00').getMonth();
		return monthNames[m];
	}
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Kalender Hari Libur</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola hari libur nasional, cuti bersama, dan libur perusahaan</p>
		</div>
		{#if !showForm && hasPermission('announcement', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Hari Libur
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { typeFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!typeFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">Semua</button>
				{#each Object.entries(holidayTypes) as [key, label]}
					<button onclick={() => { typeFilter = key; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {typeFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">{label}</button>
				{/each}
			</div>
			<div class="flex items-center gap-2">
				<select bind:value={yearFilter} onchange={() => { page = 1; load(); }} class="px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-800 rounded-lg outline-none bg-white dark:bg-gray-900">
					{#each [2026, 2027, 2025] as year}
						<option value={year}>{year}</option>
					{/each}
				</select>
				<span class="text-xs text-gray-400 dark:text-gray-500">{total > 0 ? `${total} hari libur` : ''}</span>
			</div>
		</div>
	{/if}

	{#if showForm}
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
						<label for="hol-date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal <span class="text-red-500">*</span></label>
						<input id="hol-date" type="date" bind:value={form.date} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
					<div>
						<label for="hol-type" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe</label>
						<select id="hol-type" bind:value={form.holiday_type} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
							{#each Object.entries(holidayTypes) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
				</div>
				<div>
					<label for="hol-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Hari Libur <span class="text-red-500">*</span></label>
					<input id="hol-name" type="text" bind:value={form.name} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Misal: Hari Raya Natal" />
				</div>
				<div>
					<label for="hol-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
					<input id="hol-desc" type="text" bind:value={form.description} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Keterangan tambahan..." />
				</div>
				<label class="flex items-center gap-2 cursor-pointer">
					<input type="checkbox" bind:checked={form.is_recurring_yearly} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]/30" />
					<span class="text-sm text-gray-700 dark:text-gray-300">Berulang setiap tahun (libur nasional tetap)</span>
				</label>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					{isEditing ? 'Simpan' : 'Tambah'}
				</button>
			</div>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6"><PulseLoader variant="table-row" count={5} /></div>
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
					title="Belum ada hari libur"
					description="Belum ada hari libur yang ditambahkan."
				/>
			{:else}
				<!-- Calendar-like header -->
				<div class="hidden md:grid grid-cols-12 gap-3 px-5 py-4 border-b border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
					{#each Array.from({ length: 12 }, (_, i) => i) as monthIdx}
						<button onclick={() => { monthFilter = monthFilter === monthIdx ? null : monthIdx; }}
							class="text-center text-xs font-medium py-2 px-1 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer {monthFilter === monthIdx ? 'bg-[#1A56DB] text-white shadow-sm' : monthIdx === new Date().getMonth() && monthFilter === null ? 'bg-[#1A56DB]/10 text-[#1A56DB]' : 'text-gray-600 dark:text-gray-400'}">
							{monthNames[monthIdx].substring(0, 3)}
						</button>
					{/each}
				</div>
				<!-- Desktop: AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<!-- Mobile: grouped by month -->
				<PullToRefresh onRefresh={load}>
					<div class="md:hidden space-y-3">
						{#each [...new Set(items.map(i => getMonth(i.date)))] as month}
							<div class="px-4 py-2">
								<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider">{month}</h3>
							</div>
							{#each items.filter(i => getMonth(i.date) === month) as item}
								<MobileCard
									title={item.name}
									subtitle={`${formatShortDate(item.date)}${item.is_recurring_yearly ? ' · Tahunan' : ''}`}
									avatar={getInitials(item.name)}
									avatarColor={getAvatarTheme('holiday').gradientClasses}
									badges={[{ label: holidayTypes[item.holiday_type] || item.holiday_type, color: typeColors[item.holiday_type] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300' }]}
								>
									{#snippet children()}
										{#if item.description}
											<div class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2">{item.description}</div>
										{/if}
									{/snippet}
									{#snippet footer()}
										<div class="flex items-center gap-2">
											{#if hasPermission('announcement', 'update')}
												<button onclick={() => openEdit(item.id)} class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95">Edit</button>
											{/if}
											{#if hasPermission('announcement', 'delete')}
												<button onclick={() => handleDelete(item.id)} class="flex-1 py-2 text-xs font-medium text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95">Hapus</button>
											{/if}
										</div>
									{/snippet}
								</MobileCard>
							{/each}
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
	{/if}
</div>
