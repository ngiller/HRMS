<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { documents as api, employees } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
import ConfirmModal from '$lib/components/ConfirmModal.svelte';
	type Document = {
		id: string;
		employee_id: string;
		employee_name: string;
		doc_type: string;
		file_name: string;
		title: string;
		status: string;
		expiry_date: string;
		created_at: string;
	};

	type FormData = {
		employee_id: string;
		doc_type: string;
		title: string;
		description: string;
		expiry_date: string;
		file: File | null;
		file_name: string;
		file_url: string;
	};

	const docTypes: Record<string, string> = {
		ktp: 'KTP', kk: 'Kartu Keluarga', ijazah: 'Ijazah',
		sertifikat: 'Sertifikat', kontrak: 'Kontrak Kerja',
		npwp: 'NPWP', bpjs: 'BPJS', photo: 'Foto', other: 'Lainnya',
	};

	let items = $state<Document[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let typeFilter = $state('');
	let employeeFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let form = $state<FormData>({ employee_id: '', doc_type: 'ktp', title: '', description: '', expiry_date: '', file: null, file_name: '', file_url: '' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<any>(null);
let showDeleteConfirm = $state(false);
let deleteId: string | null = null;
	let isDetailLoading = $state(false);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	let allEmployees = $state<any[]>([]);

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu', verified: 'Terverifikasi',
		rejected: 'Ditolak',
	};
	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		verified: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
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
	function iconVerify(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>';
	}
	function iconReject(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>';
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
			field: 'employee_name', headerName: 'Karyawan', minWidth: 220, flex: 1,
			cellRenderer: (params: any) => {
				if (!params.value) return '';
				const initials = getInitials(params.value);
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-indigo-50 to-indigo-100 flex items-center justify-center text-xs font-semibold text-indigo-600 shrink-0 ring-1 ring-indigo-200">${initials}</div>
					<div class="text-sm font-medium text-gray-900 dark:text-white">${params.value}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'title', headerName: 'Judul', minWidth: 180, flex: 1,
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-900 dark:text-white font-medium',
		},
		{
			field: 'doc_type', headerName: 'Tipe', minWidth: 120,
			valueFormatter: (params: any) => docTypes[params.value] || params.value || '',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600 dark:text-gray-400',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: any) => {
				const status = params.value || '';
				return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${statusLabels[status] || status}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
		},
		{
			field: 'expiry_date', headerName: 'Berakhir', minWidth: 120,
			valueFormatter: (params: any) => params.value ? formatDate(params.value) : '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 120,
			valueFormatter: (params: any) => formatDate(params.value),
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'id', headerName: '', minWidth: 150, maxWidth: 150,
			cellRenderer: (params: any) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-[#1A56DB] hover:bg-blue-50 dark:hover:bg-blue-900/30 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('document', 'update')) {
					const verifyBtn = createActionButton(iconVerify(),
						'p-1.5 rounded-lg text-green-500 dark:text-green-400 hover:text-green-700 hover:bg-green-50 dark:hover:bg-green-900/30 transition cursor-pointer',
						'Verifikasi', () => handleVerify(item.id));
					container.appendChild(verifyBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 dark:text-red-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
				}

				if (hasPermission('document', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/30 transition cursor-pointer',
						'Hapus', () => requestDelete(item.id));
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
		if (showForm && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (items.length > 0 && gridContainer && !showForm) {
			if (!gridApi) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			gridApi.updateGridOptions({ rowData: items as any[] });
		}
	});

	onDestroy(() => { gridApi?.destroy(); gridApi = null; });

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		try {
			const empResp: any = await employees.list(1, 200);
			allEmployees = empResp.data || [];
		} catch (_) {}
		load();
	});

	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true; errorMessage = '';
		try {
			const response: any = await api.list(page, perPage, statusFilter, employeeFilter, typeFilter);
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
		formTitle = 'Tambah Dokumen';
		form = { employee_id: '', doc_type: 'ktp', title: '', description: '', expiry_date: '', file: null, file_name: '', file_url: '' };
		formError = '';
		showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function handleSave() {
		if (!form.employee_id) { formError = 'Karyawan harus dipilih'; return; }
		if (!form.title.trim()) { formError = 'Judul dokumen harus diisi'; return; }
		if (!form.doc_type) { formError = 'Tipe dokumen harus diisi'; return; }

		isSaving = true; formError = '';
		try {
			await api.create({
				employee_id: form.employee_id,
				doc_type: form.doc_type,
				title: form.title.trim(),
				description: form.description.trim(),
				file_name: form.file?.name || '',
				file_url: form.file_url || '',
				expiry_date: form.expiry_date || '',
			});
			cancelForm(); load();
		} catch (error: any) { formError = error.message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true; detailId = id; isDetailLoading = true; detailData = null;
		try { const response: any = await api.get(id); detailData = response.data; }
		catch (_) { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; }

	async function handleVerify(id: string) {
		isSaving = true;
		try { await api.verify(id); load(); }
		catch (error: any) { errorMessage = error.message || 'Gagal memverifikasi'; }
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

	function requestDelete(id: string) {
  deleteId = id;
  showDeleteConfirm = true;
}
async function confirmDelete() {
  if (!deleteId) return;
  isSaving = true;
  try {
    await api.remove(deleteId);
    load();
  } catch (error: any) {
    errorMessage = error.message || 'Gagal menghapus';
  } finally {
    isSaving = false;
    showDeleteConfirm = false;
    deleteId = null;
  }
}
function cancelDelete() {
  showDeleteConfirm = false;
  deleteId = null;
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
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Dokumen Karyawan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola dokumen karyawan dengan verifikasi</p>
		</div>
		{#if !showForm && hasPermission('document', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Dokumen
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">Semua</button>
				{#each Object.entries(statusLabels) as [key, label]}
					<button onclick={() => { statusFilter = key; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800'}">{label}</button>
				{/each}
			</div>
			<div class="flex flex-wrap items-center gap-2">
				<select bind:value={typeFilter} onchange={() => { page = 1; load(); }} class="px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-800 rounded-lg outline-none bg-white dark:bg-gray-900">
					<option value="">Semua Tipe</option>
					{#each Object.entries(docTypes) as [key, label]}
						<option value={key}>{label}</option>
					{/each}
				</select>
				<span class="text-xs text-gray-400 dark:text-gray-500">{total > 0 ? `${total} dokumen` : ''}</span>
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
						<label for="doc-employee" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Karyawan <span class="text-red-500">*</span></label>
						<select id="doc-employee" bind:value={form.employee_id} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
							<option value="">Pilih Karyawan</option>
							{#each allEmployees as emp}
								<option value={emp.id}>{emp.full_name || emp.employee_name || '-'}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="doc-type" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Dokumen <span class="text-red-500">*</span></label>
						<select id="doc-type" bind:value={form.doc_type} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
							{#each Object.entries(docTypes) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
				</div>
				<div>
					<label for="doc-title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Judul Dokumen <span class="text-red-500">*</span></label>
					<input id="doc-title" type="text" bind:value={form.title} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="Misal: Scan KTP" />
				</div>
				<div>
					<label for="doc-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
					<textarea id="doc-desc" bind:value={form.description} rows="2" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Keterangan tambahan..."></textarea>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="doc-expiry" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal Berakhir</label>
						<input id="doc-expiry" type="date" bind:value={form.expiry_date} class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
					</div>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Simpan
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Dokumen</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if isDetailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-48"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-64"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-40"></div></div>
				{:else if detailData}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Dokumen</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Status</span><p><span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 {statusColors[detailData.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">{statusLabels[detailData.status] || detailData.status}</span></p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Tipe</span><p class="text-sm font-medium text-gray-900 dark:text-white">{docTypes[detailData.doc_type] || detailData.doc_type}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Judul</span><p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.title || '-'}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Deskripsi</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.description || '-'}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Nama File</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.file_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Masa Berakhir</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.expiry_date ? formatDate(detailData.expiry_date) : '-'}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Karyawan & Verifikasi</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400 dark:text-gray-500">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.employee_name || '-'}</p></div>
								{#if detailData.verified_by_name}<div><span class="text-xs text-gray-400 dark:text-gray-500">Diverifikasi Oleh</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailData.verified_by_name}</p></div>{/if}
								{#if detailData.verified_at}<div><span class="text-xs text-gray-400 dark:text-gray-500">Diverifikasi Pada</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(detailData.verified_at)}</p></div>{/if}
								{#if detailData.rejection_reason}<div><span class="text-xs text-gray-400 dark:text-gray-500">Alasan Ditolak</span><p class="text-sm text-red-600 dark:text-red-400">{detailData.rejection_reason}</p></div>{/if}
							</div>
							<div class="mt-4 flex gap-2">
								{#if detailData.file_url}
									<a href={detailData.file_url} target="_blank" class="inline-flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg text-sm font-medium hover:bg-gray-200 dark:hover:bg-gray-700 transition">
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
										Download
									</a>
								{/if}
								{#if detailData.status === 'pending' && hasPermission('document', 'update')}
									<button onclick={() => handleVerify(detailData.id)} disabled={isSaving} class="px-3 py-2 bg-green-600 text-white rounded-lg text-sm font-medium hover:bg-green-700 transition disabled:opacity-50 cursor-pointer">Verifikasi</button>
									<button onclick={() => openReject(detailData.id)} disabled={isSaving} class="px-3 py-2 bg-red-600 text-white rounded-lg text-sm font-medium hover:bg-red-700 transition disabled:opacity-50 cursor-pointer">Tolak</button>
								{/if}
							</div>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 dark:text-gray-400 text-center py-8">Gagal memuat detail dokumen</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-44"></div><div class="h-3 bg-gray-50 dark:bg-gray-800 rounded w-28"></div></div><div class="h-6 bg-gray-100 dark:bg-gray-800 rounded-full w-20"></div><div class="h-8 bg-gray-100 dark:bg-gray-800 rounded w-24"></div></div>{/each}</div></div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center"><svg class="w-7 h-7 text-gray-400 dark:text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" /></svg></div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Belum ada dokumen</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">Belum ada dokumen yang diupload.</p>
				</div>
			{:else}
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden divide-y divide-gray-100 dark:divide-gray-800">
					{#each items as item}
						<div class="p-4 hover:bg-blue-50/40 dark:hover:bg-blue-900/20 transition-colors">
							<div class="flex items-center gap-3 mb-2">
								<div class="w-10 h-10 rounded-lg bg-gradient-to-br from-indigo-50 to-indigo-100 dark:from-indigo-900/50 dark:to-indigo-800/50 flex items-center justify-center text-xs font-semibold text-indigo-600 dark:text-indigo-300 ring-1 ring-indigo-200 dark:ring-indigo-800">{getInitials(item.employee_name)}</div>
								<div class="flex-1 min-w-0">
									<div class="text-sm font-medium text-gray-900 dark:text-white truncate">{item.title}</div>
									<div class="text-xs text-gray-400 dark:text-gray-500">{docTypes[item.doc_type] || item.doc_type}</div>
								</div>
								<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 {statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">{statusLabels[item.status] || item.status}</span>
							</div>
							<div class="flex items-center gap-1 mt-2">
								<button onclick={() => openDetail(item.id)} class="px-2.5 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Detail</button>
								{#if item.status === 'pending' && hasPermission('document', 'update')}
									<button onclick={() => handleVerify(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 hover:bg-green-100 dark:hover:bg-green-900/50 transition cursor-pointer">Verifikasi</button>
									<button onclick={() => openReject(item.id)} disabled={isSaving} class="px-2.5 py-1.5 text-xs font-medium rounded-lg bg-red-50 dark:bg-red-900/30 text-red-700 dark:text-red-300 hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer">Tolak</button>
								{/if}
							</div>
						</div>
					{/each}
				</div>
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

{#if showRejectModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak dokumen" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Dokumen</h3>
				<div class="space-y-3">
					<label for="doc-reject-reason" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea id="doc-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Tolak
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}

{#if showDeleteConfirm}
	<ConfirmModal
		title="Hapus Dokumen"
		message="Apakah Anda yakin ingin menghapus dokumen ini? Tindakan ini tidak dapat dibatalkan."
		confirmText="Hapus"
		confirmColor="red"
		onConfirm={confirmDelete}
		onCancel={cancelDelete}
		{isSaving}
	/>
{/if}
