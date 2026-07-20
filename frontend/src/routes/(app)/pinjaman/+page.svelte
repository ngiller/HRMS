<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount } from 'svelte';
	import { loans, employees, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import StaggerList from '$lib/components/StaggerList.svelte';
	import { getAvatarTheme } from '$lib/avatar-theme.js';

	interface LoanItem {
		id: string; employee_name: string; loan_type: string; amount: number;
		installment_count: number; installment_amount: number; remaining_balance: number;
		status: string; created_at: string; employee_id: string; interest_rate: number;
		total_interest: number; total_amount: number; payment_method: string;
		purpose: string; disbursed_at: string;
		approval_trail?: string;
	}

	interface EmployeeOption { id: string; full_name: string; }

	interface LoanStats {
		total_loans: number; active_loans: number;
		total_disbursed: number; total_outstanding: number;
	}

	let loanData = $state<LoanItem[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let perPage = $state(25);
	let isLoading = $state(true);
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;
	let statusFilter = $state('');
	let errorMessage = $state('');

	// Inline views
	let showForm = $state(false);
	let createForm = $state({ employee_id: '', loan_type: 'regular', amount: 0, interest_rate: 0, installment_count: 6, payment_method: 'payroll_deduction', purpose: '' });
	let createLoading = $state(false);
	let employeeOptions = $state<EmployeeOption[]>([]);

	let showDetail = $state(false);
	let detailItem = $state<LoanItem | null>(null);
	let detailLoading = $state(false);

	let actionLoading = $state(false);
	let rejectionReason = $state('');
	let filteredLoans = $derived.by(() => {
		if (!searchQuery.trim()) return loanData;
		const q = searchQuery.toLowerCase();
		return loanData.filter(i =>
			i.employee_name?.toLowerCase().includes(q) ||
			i.loan_type?.toLowerCase().includes(q) ||
			i.purpose?.toLowerCase().includes(q)
		);
	});

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; }, 400);
	}

	// Reject Modal
	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);

	// Stats
	let stats = $state<LoanStats | null>(null);

	onMount(() => {
		loadData();
		loadStats();
		loadEmployeeOptions();
	});

	async function loadStats() {
		try {
			const res = await loans.getStats();
			if (res?.success) stats = res.data;
		} catch { /* silent fail */ }
	}

	async function loadEmployeeOptions() {
		try {
			const res = await employees.list(1, 100);
			if (res?.success) employeeOptions = res.data;
		} catch { /* silent */ }
	}

	async function loadData() {
		isLoading = true;
		errorMessage = '';
		try {
			const res = await loans.list(currentPage, perPage, statusFilter) as { success: boolean; data: LoanItem[]; meta?: { total: number } };
			if (res?.success) {
				loanData = res.data || [];
				total = res.meta?.total || 0;
			}
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	}

	function onPageChange(page: number) {
		currentPage = page;
		loadData();
	}

	async function loadDetail(id: string) {
		showForm = false;
		rejectionReason = '';
		if (showDetail) {
			showDetail = false;
			await new Promise(r => setTimeout(r, 50));
		}
		detailLoading = true;
		showDetail = true;
		try {
			const res = await loans.get(id);
			if (res?.success) detailItem = res.data;
		} catch {
			detailItem = null;
		} finally {
			detailLoading = false;
		}
	}

	function closeDetail() {
		showDetail = false;
		detailItem = null;
	}

	function openForm() {
		showDetail = false;
		createForm = { employee_id: '', loan_type: 'regular', amount: 0, interest_rate: 0, installment_count: 6, payment_method: 'payroll_deduction', purpose: '' };
		errorMessage = '';
		showForm = true;
	}

	function cancelForm() {
		showForm = false;
		errorMessage = '';
	}

	async function handleCreate() {
		createLoading = true;
		errorMessage = '';
		try {
			const res = await loans.create(createForm);
			if (res?.success) {
				showForm = false;
				createForm = { employee_id: '', loan_type: 'regular', amount: 0, interest_rate: 0, installment_count: 6, payment_method: 'payroll_deduction', purpose: '' };
				loadData();
				loadStats();
			}
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal membuat pinjaman';
		} finally {
			createLoading = false;
		}
	}

	async function handleApprove(id: string) {
		actionLoading = true;
		try {
			const res = await loans.approve(id);
			if (res?.success) {
				loadData();
				loadDetail(id);
				loadStats();
			}
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal menyetujui pinjaman';
		} finally {
			actionLoading = false;
		}
	}

	function openReject(id: string) {
		rejectId = id;
		rejectionReason = '';
		showRejectModal = true;
	}

	function cancelReject() {
		showRejectModal = false;
		rejectId = null;
		rejectionReason = '';
	}

	async function handleReject() {
		if (!rejectId) return;
		actionLoading = true;
		try {
			const res = await loans.reject(rejectId, { rejection_reason: rejectionReason });
			if (res?.success) {
				showRejectModal = false;
				const id = rejectId;
				rejectId = null;
				rejectionReason = '';
				loadData();
				if (showDetail && detailItem?.id === id) loadDetail(id);
				loadStats();
			}
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal menolak pinjaman';
		} finally {
			actionLoading = false;
		}
	}

	async function handleDisburse(id: string) {
		actionLoading = true;
		try {
			const res = await loans.disburse(id, { disburse_date: new Date().toISOString().split('T')[0] });
			if (res?.success) {
				loadData();
				loadDetail(id);
				loadStats();
			}
		} catch (err) {
			errorMessage = err instanceof ApiError ? err.message : 'Gagal mencairkan pinjaman';
		} finally {
			actionLoading = false;
		}
	}

	const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		active: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		completed: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		defaulted: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
	};

	function parseApprovalTrail(trail: string): any[] {
		try {
			const parsed = JSON.parse(trail);
			return Array.isArray(parsed) ? parsed : [];
		} catch {
			return [];
		}
	}

	function getStatusBadge(status: string) {
		const labels: Record<string, string> = {
			pending: 'Menunggu', approved: 'Disetujui', active: 'Aktif',
			completed: 'Lunas', rejected: 'Ditolak', defaulted: 'Macet', cancelled: 'Dibatalkan',
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600 dark:bg-gray-900 dark:text-gray-300'}">${labels[status] || status}</span>`;
	}

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->

<!-- eslint-disable svelte/no-at-html-tags -->

<div class="w-full">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Pinjaman Karyawan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Kelola pengajuan dan pencairan pinjaman</p>
		</div>
		{#if !showForm}
			<button onclick={openForm} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Ajukan Pinjaman
			</button>
		{/if}
	</div>

	<!-- Stats Cards -->
	{#if stats}
		<div class="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
				<p class="text-xs text-gray-400 uppercase tracking-wide">Total Pinjaman</p>
				<p class="text-xl font-bold text-gray-900 dark:text-gray-100 mt-1">{stats.total_loans || 0}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
				<p class="text-xs text-gray-400 uppercase tracking-wide">Pinjaman Aktif</p>
				<p class="text-xl font-bold text-gray-900 dark:text-gray-100 mt-1">{stats.active_loans || 0}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
				<p class="text-xs text-gray-400 uppercase tracking-wide">Total Tercairkan</p>
				<p class="text-xl font-bold text-[#1A56DB] mt-1">{formatCurrency(stats.total_disbursed || 0)}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4">
				<p class="text-xs text-gray-400 uppercase tracking-wide">Outstanding</p>
				<p class="text-xl font-bold text-amber-600 mt-1">{formatCurrency(stats.total_outstanding || 0)}</p>
			</div>
		</div>
	{/if}

	{#if showForm}
		<!-- ─── Inline Form ─── -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Ajukan Pinjaman Baru</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); handleCreate(); }} class="px-6 py-5 space-y-4">
				{#if errorMessage}
					<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-2.5">{errorMessage}</div>
				{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="loan-employee" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Karyawan <span class="text-red-500">*</span></label>
						<select id="loan-employee" bind:value={createForm.employee_id} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20">
							<option value="">Pilih Karyawan</option>
							{#each employeeOptions as emp (emp.id)}
								<option value={emp.id}>{emp.full_name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="loan-type" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Pinjaman <span class="text-red-500">*</span></label>
						<select id="loan-type" bind:value={createForm.loan_type} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
							<option value="regular">Regular</option>
							<option value="emergency">Darurat</option>
							<option value="education">Pendidikan</option>
						</select>
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
					<div>
						<label for="loan-amount" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Jumlah Pinjaman (Rp) <span class="text-red-500">*</span></label>
						<input id="loan-amount" type="number" bind:value={createForm.amount} min="1" required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" />
					</div>
					<div>
						<label for="loan-interest" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Bunga (%)</label>
						<input id="loan-interest" type="number" bind:value={createForm.interest_rate} min="0" step="0.1" class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" />
					</div>
					<div>
						<label for="loan-tenor" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tenor (bulan) <span class="text-red-500">*</span></label>
						<select id="loan-tenor" bind:value={createForm.installment_count} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
							{#each Array.from({ length: 24 }, (_, i) => i + 1) as n (n)}
								<option value={n}>{n} bulan</option>
							{/each}
						</select>
					</div>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="loan-payment" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Metode Pembayaran</label>
						<select id="loan-payment" bind:value={createForm.payment_method} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
							<option value="payroll_deduction">Potong Gaji</option>
							<option value="manual_transfer">Transfer Manual</option>
						</select>
					</div>
					<div>
						<label for="loan-purpose" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tujuan Pinjaman <span class="text-red-500">*</span></label>
						<textarea id="loan-purpose" bind:value={createForm.purpose} required rows={2} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20 resize-none" placeholder="Deskripsi tujuan pinjaman"></textarea>
					</div>
				</div>
				<div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
					<button type="button" onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button type="submit" disabled={createLoading} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
						{#if createLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Ajukan Pinjaman
					</button>
				</div>
			</form>
		</div>
	{:else if showDetail}
		<!-- ─── Inline Detail ─── -->
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Detail Pinjaman</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">					{#if detailLoading}
					<PulseLoader variant="text" count={3} />
				{:else if detailItem}
					{@const item = detailItem}
					{#if (detailItem as any).approval_trail && (detailItem as any).approval_trail !== '[]' && (detailItem as any).approval_trail !== ''}
						{@const trail = parseApprovalTrail((detailItem as any).approval_trail)}
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
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Pinjaman</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.employee_name}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(item.status)}</p></div>
								<div><span class="text-xs text-gray-400">Tipe</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.loan_type}</p></div>
								<div><span class="text-xs text-gray-400">Jumlah Pinjaman</span><p class="text-sm font-bold text-[#1A56DB]">{formatCurrency(item.amount)}</p></div>
								<div><span class="text-xs text-gray-400">Bunga</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.interest_rate}% ({formatCurrency(item.total_interest)})</p></div>
								<div><span class="text-xs text-gray-400">Total</span><p class="text-sm font-semibold text-gray-900 dark:text-gray-100">{formatCurrency(item.total_amount)}</p></div>
								<div><span class="text-xs text-gray-400">Metode Pembayaran</span><p class="text-sm text-gray-700 dark:text-gray-300">{item.payment_method === 'payroll_deduction' ? 'Potong Gaji' : 'Transfer Manual'}</p></div>
								<div><span class="text-xs text-gray-400">Tujuan</span><p class="text-sm text-gray-700 dark:text-gray-300">{item.purpose}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Detail Pembayaran</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Tenor</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.installment_count} bulan</p></div>
								<div><span class="text-xs text-gray-400">Angsuran/Bulan</span><p class="text-sm font-semibold text-gray-900 dark:text-gray-100">{formatCurrency(item.installment_amount)}</p></div>
								<div><span class="text-xs text-gray-400">Sisa</span><p class="text-sm font-semibold text-amber-600">{formatCurrency(item.remaining_balance)}</p></div>
								{#if item.disbursed_at}
									<div><span class="text-xs text-gray-400">Dicairkan</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{formatDate(item.disbursed_at)}</p></div>
								{/if}
							</div>
						</div>
					</div>
					<div class="mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
						{#if item.status === 'pending'}
							<div class="flex gap-2">
								<button onclick={() => handleApprove(item.id)} disabled={actionLoading} class="flex-1 py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition disabled:opacity-50 cursor-pointer">Setujui</button>
								<button onclick={() => openReject(item.id)} disabled={actionLoading} class="flex-1 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 cursor-pointer">Tolak</button>
							</div>
						{/if}
						{#if item.status === 'approved'}
							<button onclick={() => handleDisburse(item.id)} disabled={actionLoading} class="w-full py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 cursor-pointer">
								{actionLoading ? 'Memproses...' : 'Cairkan Dana'}
							</button>
						{/if}
					</div>
				{:else}
					<p class="text-sm text-gray-500 text-center py-8">Gagal memuat detail pinjaman</p>
				{/if}
			</div>
		</div>
	{:else}
		<!-- ─── Table ─── -->
		<div class="flex flex-col sm:flex-row sm:items-center gap-3 mb-4">
		<div class="relative flex-1 max-w-md">
			<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
			<input type="search" value={searchQuery} placeholder="Cari pinjaman..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
		</div>
				<div class="flex gap-3">
			<select
				bind:value={statusFilter}
				onchange={() => { currentPage = 1; loadData(); }}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none"
			>
				<option value="">Semua Status</option>
				<option value="pending">Menunggu</option>
				<option value="approved">Disetujui</option>
				<option value="active">Aktif</option>
				<option value="completed">Lunas</option>
				<option value="rejected">Ditolak</option>
			</select>
			</div>
		</div>

		{#if errorMessage}
			<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{errorMessage}</div>
		{/if}

		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
			<!-- Desktop Table -->
			<div class="hidden md:block overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
						<tr>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Karyawan</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tipe</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Jumlah</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tenor</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Angsuran</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Sisa</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tanggal</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if isLoading}
							<PulseLoader variant="table-row" count={5} />
						{:else if loanData.length === 0}
							<tr><td colspan="9" class="px-4 py-8 text-center text-sm text-gray-400">Belum ada data pinjaman</td></tr>
						{:else}
							{#each filteredLoans as item (item)}
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition {detailItem?.id === item.id && showDetail ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''}">
									<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{item.employee_name}</td>
									<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.loan_type}</td>
									<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{formatCurrency(item.amount)}</td>
									<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.installment_count} bulan</td>
									<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{formatCurrency(item.installment_amount)}</td>
									<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{formatCurrency(item.remaining_balance)}</td>
									<td class="px-4 py-3">{@html getStatusBadge(item.status)}</td>
									<td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{new Date(item.created_at).toLocaleDateString('id-ID')}</td>
									<td class="px-4 py-3">
										<button onclick={() => loadDetail(item.id)} class="text-xs text-[#1A56DB] hover:underline font-medium cursor-pointer">Detail</button>
									</td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>

			<!-- Mobile Card List -->
			<div class="md:hidden p-3 space-y-3">
				{#if isLoading}
					<PulseLoader variant="card" count={3} />
				{:else if loanData.length === 0}					<EmptyState
						variant="empty"
						title="Belum ada data pinjaman"
						description="Belum ada data pinjaman yang diajukan."
					/>
				{:else}
					<StaggerList items={filteredLoans}>
						{#snippet children(item)}
							{@const theme = getAvatarTheme('loan')}
							<MobileCard
								avatar={item.employee_name}
								avatarColor={theme.gradientClasses}
								title={item.employee_name}
								subtitle={item.loan_type.charAt(0).toUpperCase() + item.loan_type.slice(1)}
								badges={[{ label: item.status === 'pending' ? 'Menunggu' : item.status === 'approved' ? 'Disetujui' : item.status === 'active' ? 'Aktif' : item.status === 'completed' ? 'Lunas' : item.status === 'rejected' ? 'Ditolak' : item.status, color: item.status === 'pending' ? 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800' : item.status === 'approved' || item.status === 'active' ? 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800' : item.status === 'completed' ? 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800' : item.status === 'rejected' ? 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800' : 'bg-gray-50 text-gray-600 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700' }]}
								onclick={() => loadDetail(item.id)}
								clickable={true}
							>
								{#snippet children()}
									<div class="grid grid-cols-3 gap-2 text-center mb-2">
										<div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5 px-2">
											<div class="text-[10px] text-gray-400 dark:text-gray-500">Jumlah</div>
											<div class="text-xs font-bold text-[#1A56DB]">{formatCurrency(item.amount)}</div>
										</div>
										<div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5 px-2">
											<div class="text-[10px] text-gray-400 dark:text-gray-500">Angsuran/bln</div>
											<div class="text-xs font-semibold text-gray-900 dark:text-white">{formatCurrency(item.installment_amount)}</div>
										</div>
										<div class="bg-gray-50 dark:bg-gray-800 rounded-lg py-1.5 px-2">
											<div class="text-[10px] text-gray-400 dark:text-gray-500">Sisa</div>
											<div class="text-xs font-bold text-amber-600 dark:text-amber-400">{formatCurrency(item.remaining_balance)}</div>
										</div>
									</div>
								{/snippet}
								{#snippet footer()}
									<div class="flex items-center justify-between text-xs text-gray-400 dark:text-gray-500">
										<span>{item.installment_count} bulan &middot; {new Date(item.created_at).toLocaleDateString('id-ID')}</span>
										<span class="text-[#1A56DB] font-medium">Detail →</span>
									</div>
								{/snippet}
							</MobileCard>
						{/snippet}
					</StaggerList>
				{/if}
			</div>

			<div class="hidden md:flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
				<span class="text-xs text-gray-400">Total {total} data</span>
				<div class="flex gap-1">
					<button onclick={() => onPageChange(currentPage - 1)} disabled={currentPage <= 1} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Prev</button>
					<span class="px-3 py-1 text-sm text-gray-500">{(currentPage - 1) * perPage + 1} - {Math.min(currentPage * perPage, total)}</span>
					<button onclick={() => onPageChange(currentPage + 1)} disabled={currentPage >= totalPages} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Next</button>
				</div>
			</div>
			
			<!-- Mobile Pagination -->
			<div class="md:hidden flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
				<span class="text-xs text-gray-400">{(currentPage - 1) * perPage + 1}-{Math.min(currentPage * perPage, total)} dari {total}</span>
				<div class="flex gap-2">
					<button onclick={() => onPageChange(currentPage - 1)} disabled={currentPage <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer active:scale-95">Prev</button>
					<button onclick={() => onPageChange(currentPage + 1)} disabled={currentPage >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer active:scale-95">Next</button>
				</div>
			</div>
		</div>
	{/if}
</div>

<!-- ─── Reject Modal ─── -->
<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
			<div onclick={cancelReject} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelReject(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tolak pengajuan pinjaman" class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 text-center mb-4">Tolak Pengajuan Pinjaman</h3>
				<div class="space-y-3">
					<label for="loan-reject-reason" class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea id="loan-reject-reason" bind:value={rejectionReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 transition resize-none" placeholder="Masukkan alasan penolakan (opsional)"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleReject} disabled={actionLoading} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if actionLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Tolak
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
