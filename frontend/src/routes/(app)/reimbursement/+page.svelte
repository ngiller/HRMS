<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount, onDestroy } from 'svelte';
	import { reimbursements as api, employees } from '$lib/api.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import SwipeActions from '$lib/components/SwipeActions.svelte';
import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme } from '$lib/avatar-theme.js';
	import { hasPermission } from '$lib/permissions.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';
	type Reimbursement = {
		id: string;
		employee_id: string;
		employee_name: string;
		type: string;
		amount: number;
		description: string;
		status: string;
		created_at: string;
		receipt_urls: string[];
		payment_method: string;
		paid_at: string | null;
		paid_by_name: string | null;
	};

	type FormData = {
		type: string;
		amount: number;
		description: string;
		receipt_urls: string[];
		receipt_files: File[];
		receipt_previews: string[];
	};

	let isUploading = $state(false);
	let dragOver = $state(false);
	let fileInput: HTMLInputElement | undefined = $state();

	const reimbursementTypes: Record<string, string> = {
		medical: 'Biaya Medis', training: 'Pelatihan',
		travel: 'Perjalanan Dinas', supplies: 'Bahan Kerja', other: 'Lain-lain',
	};

	let items = $state<Reimbursement[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let form = $state<FormData>({ type: 'medical', amount: 0, description: '', receipt_urls: [], receipt_files: [], receipt_previews: [] });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<Reimbursement | null>(null);
	let isDetailLoading = $state(false);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	let showPayModal = $state(false);
	let payId = $state<string | null>(null);
	let payMethod = $state('payroll');

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu', approved: 'Disetujui',
		rejected: 'Ditolak', paid: 'Dibayar', cancelled: 'Dibatalkan',
	};

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		active: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		completed: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
		paid: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		defaulted: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
	};

	function getStatusBadge(status: string) {
		const labels: Record<string, string> = {
			pending: 'Menunggu', approved: 'Disetujui', active: 'Aktif',
			completed: 'Lunas', rejected: 'Ditolak', defaulted: 'Macet', cancelled: 'Dibatalkan', paid: 'Dibayar'
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${labels[status] || status}</span>`;
	}

	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

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
	function iconPay(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" /></svg>';
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
			field: 'employee_name', headerName: 'Karyawan', minWidth: 220, flex: 1,
			cellRenderer: (params: AgGridCellParams<Reimbursement>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-teal-50 to-teal-100 flex items-center justify-center text-xs font-semibold text-teal-600 shrink-0 ring-1 ring-teal-200">${initials}</div>
					<div class="text-sm font-medium text-gray-900">${params.value}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'type', headerName: 'Tipe', minWidth: 140,
			valueFormatter: (params: AgGridValueParams) => reimbursementTypes[params.value as string] || (params.value as string) || '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'amount', headerName: 'Jumlah', minWidth: 140, type: 'rightAligned',
			valueFormatter: (params: AgGridValueParams) => formatCurrency((params.value as number) || 0),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm font-semibold text-gray-900 text-right tabular-nums',
		},
		{
			field: 'status', headerName: 'Status', minWidth: 120, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<Reimbursement>) => {
				const status = (params.value as string) || '';
				return getStatusBadge(status);
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 120,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 150, maxWidth: 150,
			cellRenderer: (params: AgGridCellParams<Reimbursement>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Detail', () => openDetail(item.id));
				container.appendChild(viewBtn);

				if (item.status === 'pending' && hasPermission('reimbursement', 'approve')) {
					const approveBtn = createActionButton(iconApprove(),
						'p-1.5 rounded-lg text-green-500 hover:text-green-700 hover:bg-green-50 transition cursor-pointer',
						'Setujui', () => handleApprove(item.id));
					container.appendChild(approveBtn);

					const rejectBtn = createActionButton(iconReject(),
						'p-1.5 rounded-lg text-red-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Tolak', () => openReject(item.id));
					container.appendChild(rejectBtn);
				}

				if (item.status === 'approved' && hasPermission('reimbursement', 'update')) {
					const payBtn = createActionButton(iconPay(),
						'p-1.5 rounded-lg text-blue-500 hover:text-blue-700 hover:bg-blue-50 transition cursor-pointer',
						'Bayar', () => openPay(item.id));
					container.appendChild(payBtn);
				}

				if (item.status === 'pending' && hasPermission('reimbursement', 'create')) {
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
		if (showForm || showDetail) {
			gridApi?.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (gridContainer && !showForm && !showDetail) {
			if (!gridApi && agGridModule) { gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi; }
			if (gridApi) { gridApi.updateGridOptions({ rowData: items }); }
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
			const response = await api.list(page, perPage, statusFilter) as ApiResponse<Reimbursement[]>;
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
		formTitle = 'Ajukan Reimbursement';
		form = { type: 'medical', amount: 0, description: '', receipt_urls: [], receipt_files: [], receipt_previews: [] };
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		form.receipt_previews.forEach(p => URL.revokeObjectURL(p));
		showForm = false; formError = '';
	}

	async function handleUploadFiles(files: FileList | null) {
		if (!files || files.length === 0) return;
		isUploading = true;
		for (const file of Array.from(files)) {
			if (!file.type.startsWith('image/')) {
				formError = 'Hanya file gambar yang diizinkan';
				continue;
			}
			if (file.size > 5 * 1024 * 1024) {
				formError = 'Ukuran file maksimal 5MB';
				continue;
			}
			// Create preview
			const preview = URL.createObjectURL(file);
			form.receipt_files = [...form.receipt_files, file];
			form.receipt_previews = [...form.receipt_previews, preview];
		}
		isUploading = false;
	}

	function removeReceipt(index: number) {
		URL.revokeObjectURL(form.receipt_previews[index]);
		form.receipt_files = form.receipt_files.filter((_, i) => i !== index);
		form.receipt_previews = form.receipt_previews.filter((_, i) => i !== index);
		form.receipt_urls = form.receipt_urls.filter((_, i) => i !== index);
	}

	async function handleSave() {
		if (!form.type) { formError = 'Tipe reimbursement harus diisi'; return; }
		if (form.amount <= 0) { formError = 'Jumlah reimbursement harus lebih dari 0'; return; }
		if (!form.description.trim()) { formError = 'Deskripsi reimbursement harus diisi'; return; }

		isSaving = true;
		formError = '';
		try {
			// Upload files first to get URLs
			const uploadedUrls = [...form.receipt_urls];
			for (const file of form.receipt_files) {
				try {
					const resp = await api.uploadReceipt(file) as { data?: { url: string } };
					if (resp?.data?.url) {
						uploadedUrls.push(resp.data.url);
					}
				} catch (err: any) {
					formError = err.message || 'Gagal upload file';
					isSaving = false;
					return;
				}
			}

			await api.create({
				type: form.type,
				amount: form.amount,
				description: form.description.trim(),
				receipt_urls: uploadedUrls,
			});
			cancelForm();
			load();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	async function openDetail(id: string) {
		showDetail = true;
		detailId = id;
		isDetailLoading = true;
		detailData = null;
		try {
			const response = await api.get(id) as ApiResponse<Reimbursement>;
			detailData = response.data ?? null;
		} catch (_) { detailData = null; }
		finally { isDetailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailId = null; detailData = null; }

	async function handleApprove(id: string) {
		isSaving = true;
		try { await api.approve(id); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menyetujui'; }
		finally { isSaving = false; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		isSaving = true;
		try { await api.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; rejectId = null; rejectReason = ''; load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
		finally { isSaving = false; }
	}

	function openPay(id: string) { payId = id; payMethod = 'payroll'; showPayModal = true; }
	function cancelPay() { showPayModal = false; payId = null; payMethod = 'payroll'; }

	async function handlePay() {
		if (!payId) return;
		isSaving = true;
		try { await api.pay(payId, { payment_method: payMethod }); showPayModal = false; payId = null; load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal membayar'; showPayModal = false; }
		finally { isSaving = false; }
	}

	async function handleCancel(id: string) {
		isSaving = true;
		try { await api.cancel(id); load(); }
		catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal membatalkan'; }
		finally { isSaving = false; }
	}

	function parseApprovalTrail(trail: string): any[] {
		try {
			const parsed = JSON.parse(trail);
			return Array.isArray(parsed) ? parsed : [];
		} catch {
			return [];
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function formatCurrency(amount: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0, maximumFractionDigits: 0 }).format(amount);
	}

	function getInitials(name: string): string {
		const parts = name.split(' ');
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->

<!-- eslint-disable svelte/no-at-html-tags -->

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Reimbursement</h1>
			<p class="text-sm text-gray-500 mt-0.5">Ajukan dan kelola pengajuan reimbursement karyawan</p>
		</div>
		{#if !showForm && hasPermission('reimbursement', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Ajukan Reimbursement
			</button>
		{/if}
	</div>

	{#if !showForm}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">Semua</button>
				{#each Object.entries(statusLabels) as [key, label] (key)}
					<button onclick={() => { statusFilter = key; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{label}</button>
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
						<label for="reimb-type" class="block text-sm font-medium text-gray-700 mb-1.5">Tipe <span class="text-red-500">*</span></label>
						<select id="reimb-type" bind:value={form.type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
							{#each Object.entries(reimbursementTypes) as [key, label] (key)}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="reimb-amount" class="block text-sm font-medium text-gray-700 mb-1.5">Jumlah <span class="text-red-500">*</span></label>
						<div class="relative">
							<span class="absolute inset-y-0 left-0 pl-3 flex items-center text-sm text-gray-400">Rp</span>
							<input id="reimb-amount" type="number" min="1" bind:value={form.amount} class="w-full pl-10 pr-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" placeholder="0" />
						</div>
					</div>
				</div>
				<div>
					<label for="reimb-desc" class="block text-sm font-medium text-gray-700 mb-1.5">Deskripsi <span class="text-red-500">*</span></label>
					<textarea id="reimb-desc" bind:value={form.description} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Jelaskan tujuan reimbursement..."></textarea>
				</div>
				<div>
					<label for="reimb-receipt" class="block text-sm font-medium text-gray-700 mb-1.5">Lampiran Bukti (Foto)</label>
					<div
						class="border-2 border-dashed rounded-xl p-4 text-center transition cursor-pointer {dragOver ? 'border-[#1A56DB] bg-blue-50' : 'border-gray-200 hover:border-gray-300 hover:bg-gray-50'}"
						ondragover={(e) => { e.preventDefault(); dragOver = true; }}
						ondragleave={() => dragOver = false}
						ondrop={(e) => { e.preventDefault(); dragOver = false; handleUploadFiles(e.dataTransfer?.files || null); }}
						role="button"
						tabindex="0"
						onclick={() => fileInput?.click()}
						onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') fileInput?.click(); }}
					>
						<input type="file" accept="image/*" multiple bind:this={fileInput} onchange={(e) => handleUploadFiles((e.target as HTMLInputElement).files)} class="hidden" />
						{#if isUploading}
							<div class="flex flex-col items-center gap-2">
								<svg class="w-8 h-8 text-[#1A56DB] animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								<span class="text-sm text-gray-500">Mengupload...</span>
							</div>
						{:else}
							<svg class="w-8 h-8 mx-auto mb-2 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" /></svg>
							<p class="text-sm text-gray-500">Klik atau drag & drop foto bukti di sini</p>
							<p class="text-xs text-gray-400 mt-1">JPEG/PNG, maks 5MB per file</p>
						{/if}
					</div>
					{#if form.receipt_previews.length > 0}
						<div class="mt-3 grid grid-cols-3 sm:grid-cols-4 md:grid-cols-5 gap-2">
							{#each form.receipt_previews as preview, i (i)}
								<div class="relative group aspect-square rounded-lg overflow-hidden border border-gray-200 bg-gray-50">
									<img src={preview} alt="Bukti {i+1}" class="w-full h-full object-cover" />
									<button onclick={() => removeReceipt(i)} class="absolute top-1 right-1 w-5 h-5 rounded-full bg-red-500 text-white flex items-center justify-center opacity-0 group-hover:opacity-100 transition cursor-pointer text-xs" aria-label="Hapus bukti">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
									</button>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Reimbursement
				</button>
			</div>
		</div>
	{:else if showDetail}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">Detail Reimbursement</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if isDetailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 rounded w-48"></div><div class="h-4 bg-gray-50 rounded w-64"></div><div class="h-4 bg-gray-50 rounded w-40"></div></div>
				{:else if detailData}
					{@const dd = detailData}
					{#if (detailData as any).approval_trail && (detailData as any).approval_trail !== '[]' && (detailData as any).approval_trail !== ''}
						{@const trail = parseApprovalTrail((detailData as any).approval_trail)}
						<div class="bg-indigo-50/50 dark:bg-indigo-900/10 rounded-xl p-4 border border-indigo-100 dark:border-indigo-900/30 mb-6">
							<div class="flex items-center gap-2 mb-3">
								<svg class="w-4 h-4 text-indigo-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								<h3 class="text-xs font-semibold text-indigo-700 dark:text-indigo-300 uppercase tracking-wider">Progress Approval</h3>
							</div>
							<div class="space-y-2">
								{#each trail as step (step)}
									{@const isPending = step.status === 'pending'}
									{@const isApproved = step.status === 'approved'}
									{@const isRejected = step.status === 'rejected'}
									<div class="flex items-center gap-3">
										<div class="w-7 h-7 rounded-full flex items-center justify-center shrink-0 {isApproved ? 'bg-emerald-100 text-emerald-600' : isRejected ? 'bg-red-100 text-red-600' : 'bg-gray-100 dark:bg-gray-700 text-gray-400'}">
											{#if isApproved}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
											{:else if isRejected}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
											{:else}
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
											{/if}
										</div>
										<div class="flex-1 min-w-0">
											<p class="text-sm font-medium {isApproved ? 'text-emerald-700 dark:text-emerald-300' : isRejected ? 'text-red-700 dark:text-red-300' : 'text-gray-700 dark:text-gray-300'}">
												{step.approver_name || 'Approver'} 
												<span class="text-xs font-normal text-gray-400">Level {step.level || step.step}</span>
											</p>
											{#if step.note}
												<p class="text-xs text-gray-400 truncate">{step.note}{#if step.date} &middot; {step.date || '-'}{/if}</p>
											{/if}
										</div>
										<div>
											{#if isApproved}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200">Disetujui</span>
											{:else if isRejected}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-200">Ditolak</span>
											{:else}
												<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-50 text-yellow-700 ring-1 ring-yellow-200 animate-pulse">Menunggu</span>
											{/if}
										</div>
									</div>
									{#if step !== trail[trail.length - 1]}
										<div class="ml-3.5 border-l-2 border-gray-200 dark:border-gray-700 h-3"></div>
									{/if}
								{/each}
							</div>
						</div>
					{/if}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Informasi Pengajuan</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(dd.status)}</p></div>
								<div><span class="text-xs text-gray-400">Tipe</span><p class="text-sm font-medium text-gray-900">{reimbursementTypes[dd.type] || dd.type}</p></div>
								<div><span class="text-xs text-gray-400">Jumlah</span><p class="text-sm font-bold text-gray-900">{formatCurrency(dd.amount)}</p></div>
								<div><span class="text-xs text-gray-400">Deskripsi</span><p class="text-sm text-gray-700">{dd.description || '-'}</p></div>
								{#if dd.receipt_urls && dd.receipt_urls.length > 0}
									<div>
										<span class="text-xs text-gray-400 mb-2 block">Lampiran ({dd.receipt_urls.length})</span>
										<div class="grid grid-cols-2 sm:grid-cols-3 gap-2">
											{#each dd.receipt_urls as url (url)}
												<button onclick={() => window.open(url, '_blank')} class="aspect-square rounded-lg overflow-hidden border border-gray-200 bg-gray-50 hover:ring-2 hover:ring-[#1A56DB] transition cursor-pointer group relative">
													<img src={url} alt="Bukti" class="w-full h-full object-cover" loading="lazy" />
													<div class="absolute inset-0 bg-black/0 group-hover:bg-black/10 transition flex items-center justify-center">
														<svg class="w-6 h-6 text-white opacity-0 group-hover:opacity-100 transition" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
													</div>
												</button>
											{/each}
										</div>
									</div>
								{/if}
								<div><span class="text-xs text-gray-400">Diajukan Pada</span><p class="text-sm text-gray-700">{formatDate(dd.created_at)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Karyawan & Pembayaran</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Pengaju</span><p class="text-sm font-medium text-gray-900">{dd.employee_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Metode Pembayaran</span><p class="text-sm text-gray-700">{dd.payment_method === 'payroll' ? 'Potong Gaji' : dd.payment_method === 'manual_transfer' ? 'Transfer Manual' : '-'}</p></div>
								{#if dd.paid_at}<div><span class="text-xs text-gray-400">Dibayar Pada</span><p class="text-sm text-gray-700">{formatDate(dd.paid_at)}</p></div>{/if}
								{#if dd.paid_by_name}<div><span class="text-xs text-gray-400">Dibayar Oleh</span><p class="text-sm text-gray-700">{dd.paid_by_name}</p></div>{/if}
							</div>
							{#if dd.status === 'approved' && hasPermission('reimbursement', 'update')}
								<button onclick={() => openPay(dd.id)} disabled={isSaving} class="mt-4 w-full px-4 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-semibold hover:bg-blue-700 transition disabled:opacity-50 inline-flex items-center justify-center gap-2 cursor-pointer">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" /></svg>
									Bayarkan
								</button>
							{/if}
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail reimbursement</p>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<div class="p-6 animate-pulse"><div class="space-y-3">{#each [1,2,3,4,5] as _, i (i)}<div class="flex items-center gap-4 py-2"><div class="flex-1 space-y-1.5"><div class="h-4 bg-gray-100 rounded w-44"></div><div class="h-3 bg-gray-50 rounded w-28"></div></div><div class="h-6 bg-gray-100 rounded-full w-20"></div><div class="h-8 bg-gray-100 rounded w-24"></div></div>{/each}</div></div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada pengajuan reimbursement"
					description="Belum ada pengajuan reimbursement yang diajukan."
				/>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<PullToRefresh onRefresh={load}>
				<div class="md:hidden space-y-3">
					{#each items as item (item)}
						<SwipeActions
							onApprove={item.status === 'pending' && hasPermission('reimbursement', 'approve') ? () => handleApprove(item.id) : undefined}
							onReject={item.status === 'pending' && hasPermission('reimbursement', 'approve') ? () => openReject(item.id) : undefined}
						>
						<MobileCard
							avatar={item.employee_name}
							avatarColor={getAvatarTheme('reimbursement').gradientClasses}
							title={item.employee_name}
							subtitle={reimbursementTypes[item.type] || item.type}
							badges={[{ label: item.status === 'pending' ? 'Menunggu' : item.status === 'approved' ? 'Disetujui' : item.status === 'rejected' ? 'Ditolak' : item.status === 'paid' ? 'Dibayar' : item.status === 'cancelled' ? 'Dibatalkan' : item.status, color: statusColors[item.status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300' }]}
							onclick={() => openDetail(item.id)}
							clickable={true}
						>
							{#snippet children()}
								<div class="text-sm font-bold text-gray-900 dark:text-white mb-2">{formatCurrency(item.amount)}</div>
								<div class="text-xs text-gray-500 dark:text-gray-400 line-clamp-2 mb-1">
									{item.description || 'Tidak ada deskripsi'}
								</div>
							{/snippet}
							{#snippet footer()}
								<div class="flex items-center gap-2 pt-2">
									<button
										onclick={(e) => { e.stopPropagation(); openDetail(item.id); }}
										class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95"
									>
										Detail
									</button>
									{#if item.status === 'pending' && hasPermission('reimbursement', 'approve')}
										<button
											onclick={(e) => { e.stopPropagation(); handleApprove(item.id); }}
											class="flex-1 py-2 text-xs font-semibold text-green-700 dark:text-green-300 bg-green-50 dark:bg-green-900/30 rounded-lg hover:bg-green-100 dark:hover:bg-green-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
										>
											<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
											Setujui
										</button>
										<button
											onclick={(e) => { e.stopPropagation(); openReject(item.id); }}
											class="flex-1 py-2 text-xs font-semibold text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
										>
											<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
											Tolak
										</button>
									{/if}
									{#if item.status === 'approved' && hasPermission('reimbursement', 'update')}
										<button
											onclick={(e) => { e.stopPropagation(); openPay(item.id); }}
											class="flex-1 py-2 text-xs font-semibold text-blue-700 dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition cursor-pointer active:scale-95 inline-flex items-center justify-center gap-1"
										>
											<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" /></svg>
											Bayar
										</button>
									{/if}
								</div>
							{/snippet}
						</MobileCard>
						</SwipeActions>
					{/each}
				</div>
				</PullToRefresh>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30">
					<div class="text-xs text-gray-500">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
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

<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
			<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak reimbursement" class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 text-center mb-4">Tolak Reimbursement</h3>
				<div class="space-y-3">
					<label for="reimb-reject-reason" class="block text-sm font-medium text-gray-700">Alasan Penolakan</label>
					<textarea id="reimb-reject-reason" bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
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
</AnimatedPresence>

{#if showPayModal}
			<div onclick={cancelPay} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelPay(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Bayar reimbursement" class="bg-white rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-blue-50 flex items-center justify-center"><svg class="w-7 h-7 text-blue-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 text-center mb-4">Bayar Reimbursement</h3>
				<div class="space-y-3">
					<label for="reimb-pay-method" class="block text-sm font-medium text-gray-700">Metode Pembayaran</label>
					<select id="reimb-pay-method" bind:value={payMethod} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
						<option value="payroll">Melalui Payroll (Potong Gaji)</option>
						<option value="manual_transfer">Transfer Manual</option>
					</select>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelPay} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handlePay} disabled={isSaving} class="px-5 py-2.5 bg-blue-600 text-white rounded-lg text-sm font-semibold hover:bg-blue-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Konfirmasi Pembayaran
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
