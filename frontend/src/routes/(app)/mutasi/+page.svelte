<script lang="ts">
	import { onMount } from 'svelte';
	import { mutations, employees, departments, positions, positionGrades } from '$lib/api.js';
	import { hasPermission, getUser } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';

	interface MutationItem {
		id: string;
		employee_id: string;
		employee_name: string;
		mutation_type: string;
		old_department_name: string;
		new_department_name: string;
		old_position_name: string;
		new_position_name: string;
		old_base_salary: number | null;
		new_base_salary: number | null;
		reason: string;
		notes: string;
		effective_date: string;
		status: string;
		approved_by_name: string;
		rejection_reason: string;
		created_at: string;
	}

	let items = $state<MutationItem[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let statusFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Form state
	let showForm = $state(false);
	let form = $state({
		employee_id: '',
		mutation_type: 'promotion',
		new_department_id: '',
		new_position_id: '',
		new_position_grade_id: '',
		new_employment_status: '',
		new_base_salary: null as number | null,
		reason: '',
		effective_date: '',
		notes: '',
	});
	let formError = $state('');
	let isSaving = $state(false);

	// Dropdown data
	let employeeList: any[] = $state([]);
	let deptList: any[] = $state([]);
	let posList: any[] = $state([]);
	let gradeList: any[] = $state([]);

	let showDetail = $state(false);
	let detailData = $state<MutationItem | null>(null);

	let showRejectModal = $state(false);
	let rejectId = $state<string | null>(null);
	let rejectReason = $state('');

	const typeLabels: Record<string, string> = {
		promotion: 'Promosi',
		demotion: 'Demosi',
		transfer: 'Mutasi Departemen',
		position_change: 'Perubahan Jabatan',
		status_change: 'Perubahan Status',
		salary_change: 'Perubahan Gaji',
	};

	const statusLabels: Record<string, string> = {
		pending: 'Menunggu',
		approved: 'Disetujui',
		rejected: 'Ditolak',
		cancelled: 'Dibatalkan',
	};

	const statusColors: Record<string, string> = {
		pending: 'bg-yellow-50 text-yellow-700 ring-yellow-200 dark:bg-yellow-900 dark:text-yellow-200 dark:ring-yellow-800',
		approved: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		rejected: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		cancelled: 'bg-gray-50 text-gray-600 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700',
	};

	function getStatusBadge(status: string): string {
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600'}">${statusLabels[status] || status}</span>`;
	}

	onMount(() => load());

	async function load() {
		isLoading = true; errorMessage = '';
		try {
			const res = await mutations.list(page, perPage, statusFilter);
			if (res.success) {
				items = res.data || [];
				total = (res as any).meta?.total || 0;
				page = (res as any).meta?.page || 1;
				perPage = (res as any).meta?.per_page || 25;
				totalPages = Math.ceil(total / perPage);
			}
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	async function openForm() {
		formError = '';
		form = {
			employee_id: '',
			mutation_type: 'promotion',
			new_department_id: '',
			new_position_id: '',
			new_position_grade_id: '',
			new_employment_status: '',
			new_base_salary: null,
			reason: '',
			effective_date: '',
			notes: '',
		};
		showForm = true;
		try {
			const [empRes, deptRes, posRes, gradeRes] = await Promise.all([
				employees.list(1, 200),
				departments.getAll(),
				positions.getAll(),
				positionGrades.getAll(),
			]);
			if (empRes.success) employeeList = empRes.data || [];
			if (deptRes.success) deptList = deptRes.data || [];
			if (posRes.success) posList = posRes.data || [];
			if (gradeRes.success) gradeList = gradeRes.data || [];
		} catch { /* Silently fail dropdowns */ }
	}

	function cancelForm() { showForm = false; formError = ''; }

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	async function handleSave() {
		if (!form.employee_id) { formError = 'Karyawan harus diisi'; return; }
		if (!form.reason.trim()) { formError = 'Alasan mutasi harus diisi'; return; }
		if (!form.effective_date) { formError = 'Tanggal berlaku harus diisi'; return; }
		if (!form.new_department_id && !form.new_position_id && !form.new_position_grade_id && !form.new_employment_status && form.new_base_salary === null) {
			formError = 'Minimal satu perubahan harus diisi'; return;
		}

		isSaving = true; formError = '';
		try {
			await mutations.create(form);
			cancelForm();
			load();
		} catch (e: unknown) {
			formError = (e as { message?: string }).message || 'Gagal menyimpan';
		} finally {
			isSaving = false;
		}
	}

	async function openDetail(id: string) {
		showDetail = true; detailData = null;
		try {
			const res = await mutations.get(id);
			if (res.success) detailData = res.data as MutationItem;
		} catch { /* */ }
	}

	function closeDetail() { showDetail = false; detailData = null; }

	async function handleApprove(id: string) {
		try { await mutations.approve(id); load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menyetujui'; }
	}

	function openReject(id: string) { rejectId = id; rejectReason = ''; showRejectModal = true; }
	function cancelReject() { showRejectModal = false; rejectId = null; rejectReason = ''; }

	async function handleReject() {
		if (!rejectId) return;
		try { await mutations.reject(rejectId, { rejection_reason: rejectReason }); showRejectModal = false; load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal menolak'; showRejectModal = false; }
	}

	async function handleCancel(id: string) {
		try { await mutations.cancel(id); load(); if (detailData) closeDetail(); }
		catch (e: unknown) { errorMessage = (e as { message?: string }).message || 'Gagal membatalkan'; }
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight dark:text-white">Mutasi & Promosi</h1>
			<p class="text-sm text-gray-500 mt-0.5">Kelola promosi, demosi, mutasi departemen, dan perubahan jabatan</p>
		</div>
		<div class="flex items-center gap-2">
			{#if !showForm && !showDetail && hasPermission('employee', 'create')}
				<button onclick={openForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					Buat Mutasi
				</button>
			{/if}
		</div>
	</div>

	{#if !showForm && !showDetail}
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 dark:bg-gray-800 dark:border-gray-700">
			<div class="flex flex-wrap items-center gap-2">
				<button onclick={() => { statusFilter = ''; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {!statusFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">Semua</button>
				{#each ['pending', 'approved', 'rejected', 'cancelled'] as status}
					<button onclick={() => { statusFilter = status; page = 1; load(); }} class="px-3 py-1.5 text-xs font-medium rounded-lg border transition cursor-pointer {statusFilter === status ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 text-gray-600 hover:bg-gray-100 dark:border-gray-600 dark:text-gray-400'}">{statusLabels[status] || status}</button>
				{/each}
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} mutasi ditemukan` : ''}</div>
		</div>
	{/if}

	{#if showForm}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Buat Mutasi Baru</h2>
				<button onclick={cancelForm} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-4">
				{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Karyawan <span class="text-red-500">*</span></label>
						<select bind:value={form.employee_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="">-- Pilih Karyawan --</option>
							{#each employeeList as emp}
								<option value={emp.id}>{emp.full_name} ({emp.employee_id})</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tipe Mutasi <span class="text-red-500">*</span></label>
						<select bind:value={form.mutation_type} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							{#each Object.entries(typeLabels) as [key, label]}
								<option value={key}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Departemen Baru</label>
						<select bind:value={form.new_department_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each deptList as dept}
								<option value={dept.id}>{dept.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Jabatan Baru</label>
						<select bind:value={form.new_position_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each posList as pos}
								<option value={pos.id}>{pos.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Grade Baru</label>
						<select bind:value={form.new_position_grade_id} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							{#each gradeList as grade}
								<option value={grade.id}>{grade.name}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Status Kepegawaian Baru</label>
						<select bind:value={form.new_employment_status} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white">
							<option value="">-- Tidak Berubah --</option>
							<option value="tetap">Tetap</option>
							<option value="kontrak">Kontrak</option>
							<option value="percobaan">Percobaan</option>
							<option value="harian">Harian</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Gaji Pokok Baru</label>
						<input type="number" bind:value={form.new_base_salary} placeholder="Kosongkan jika tidak berubah" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Tanggal Berlaku <span class="text-red-500">*</span></label>
						<input type="date" bind:value={form.effective_date} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white" />
					</div>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Alasan Mutasi <span class="text-red-500">*</span></label>
					<textarea bind:value={form.reason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white resize-none" placeholder="Jelaskan alasan mutasi/promosi..."></textarea>
				</div>
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1.5 dark:text-gray-300">Catatan (Opsional)</label>
					<textarea bind:value={form.notes} rows="2" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white resize-none" placeholder="Catatan tambahan..."></textarea>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 bg-gray-50/50 dark:border-gray-700 dark:bg-gray-800/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer dark:border-gray-600 dark:text-gray-300">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					Ajukan Mutasi
				</button>
			</div>
		</div>
	{:else if showDetail && detailData}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Detail Mutasi</h2>
				<button onclick={closeDetail} class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 space-y-6">
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<span class="text-xs text-gray-400">Karyawan</span>
						<p class="text-sm font-medium text-gray-900 dark:text-white">{detailData.employee_name || '-'}</p>
					</div>
					<div>
						<span class="text-xs text-gray-400">Tipe Mutasi</span>
						<p class="text-sm font-medium text-gray-900 dark:text-white">{typeLabels[detailData.mutation_type] || detailData.mutation_type}</p>
					</div>
					<div>
						<span class="text-xs text-gray-400">Tanggal Berlaku</span>
						<p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(detailData.effective_date)}</p>
					</div>
					<div>
						<span class="text-xs text-gray-400">Status</span>
						<p class="mt-0.5">{@html getStatusBadge(detailData.status)}</p>
					</div>
				</div>

				<div class="border-t border-gray-100 dark:border-gray-700 pt-4">
					<h3 class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-3">Perubahan</h3>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						{#if detailData.old_department_name || detailData.new_department_name}
							<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
								<span class="text-xs text-gray-400">Departemen</span>
								<p class="text-sm"><span class="text-gray-500 line-through">{detailData.old_department_name || '-'}</span> → <span class="font-medium text-gray-900 dark:text-white">{detailData.new_department_name || '-'}</span></p>
							</div>
						{/if}
						{#if detailData.old_position_name || detailData.new_position_name}
							<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
								<span class="text-xs text-gray-400">Jabatan</span>
								<p class="text-sm"><span class="text-gray-500 line-through">{detailData.old_position_name || '-'}</span> → <span class="font-medium text-gray-900 dark:text-white">{detailData.new_position_name || '-'}</span></p>
							</div>
						{/if}
						{#if detailData.old_base_salary !== null || detailData.new_base_salary !== null}
							<div class="p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg">
								<span class="text-xs text-gray-400">Gaji Pokok</span>
								<p class="text-sm"><span class="text-gray-500 line-through">{detailData.old_base_salary ? formatCurrency(detailData.old_base_salary) : '-'}</span> → <span class="font-medium text-emerald-600 dark:text-emerald-400">{detailData.new_base_salary ? formatCurrency(detailData.new_base_salary) : '-'}</span></p>
							</div>
						{/if}
					</div>
				</div>

				<div class="border-t border-gray-100 dark:border-gray-700 pt-4">
					<span class="text-xs text-gray-400">Alasan</span>
					<p class="text-sm text-gray-700 dark:text-gray-300 mt-1">{detailData.reason || '-'}</p>
					{#if detailData.notes}
						<span class="text-xs text-gray-400 mt-3 block">Catatan</span>
						<p class="text-sm text-gray-700 dark:text-gray-300 mt-1">{detailData.notes}</p>
					{/if}
					{#if detailData.rejection_reason}
						<div class="mt-3 p-3 bg-red-50 dark:bg-red-900/20 rounded-lg">
							<span class="text-xs text-red-500 font-semibold">Alasan Penolakan</span>
							<p class="text-sm text-red-600 dark:text-red-400 mt-0.5">{detailData.rejection_reason}</p>
						</div>
					{/if}
				</div>

				{#if detailData.status === 'pending' && hasPermission('employee', 'update')}
					<div class="border-t border-gray-100 dark:border-gray-700 pt-4">
						<div class="flex items-center gap-3">
							<button onclick={() => handleApprove(detailData!.id)} class="px-5 py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition inline-flex items-center gap-2 cursor-pointer">
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
								Setujui & Terapkan
							</button>
							<button onclick={() => openReject(detailData!.id)} class="px-5 py-2.5 border border-red-200 text-red-600 rounded-lg text-sm font-semibold hover:bg-red-50 transition cursor-pointer">Tolak</button>
							<button onclick={() => handleCancel(detailData!.id)} class="px-5 py-2.5 border border-gray-200 text-gray-600 rounded-lg text-sm font-semibold hover:bg-gray-50 transition cursor-pointer dark:border-gray-600 dark:text-gray-400">Batalkan</button>
						</div>
					</div>
				{/if}

				{#if detailData.status === 'approved'}
					<div class="border-t border-gray-100 dark:border-gray-700 pt-4">
						<a href={`/mutasi/${detailData.id}/sk`}
							class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer">
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
							Cetak SK Mutasi
						</a>
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm dark:bg-gray-800 dark:border-gray-700">
			<PullToRefresh onRefresh={load}>
			{#if isLoading}
				<div class="p-6 animate-pulse space-y-3">{#each [1,2,3] as _}<div class="h-16 bg-gray-100 rounded-lg dark:bg-gray-700"></div>{/each}</div>
			{:else if errorMessage}
				<div class="py-16 text-center">
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<div class="py-8 text-center">
					<EmptyState variant="empty" title="Belum ada mutasi" description="Belum ada riwayat mutasi atau promosi." />
					{#if hasPermission('employee', 'create')}
						<button onclick={openForm} class="mt-4 px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Buat Mutasi Baru</button>
					{/if}
				</div>
			{:else}
				<div class="space-y-2">
					{#each items as item}
						<MobileCard
							title={item.employee_name || '-'}
							subtitle={`${typeLabels[item.mutation_type] || item.mutation_type} • ${formatDate(item.effective_date)}`}
							avatar={getInitials(item.employee_name || '-')}
							avatarColor={getAvatarTheme('mutation').gradientClasses}
							onclick={() => openDetail(item.id)}
						>
							{#snippet children()}
								<p class="text-xs text-gray-600 dark:text-gray-300 line-clamp-2">{item.reason}</p>
							{/snippet}
							{#snippet footer()}
								{@html getStatusBadge(item.status)}
							{/snippet}
						</MobileCard>
					{/each}
				</div>
				<div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-700">
					<div class="text-xs text-gray-400">{(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari {total}</div>
					<div class="flex items-center gap-2">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-30 cursor-pointer dark:border-gray-600 dark:text-gray-400">Sebelumnya</button>
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-50 disabled:opacity-30 cursor-pointer dark:border-gray-600 dark:text-gray-400">Selanjutnya</button>
					</div>
				</div>
			{/if}
			</PullToRefresh>
		</div>
	{/if}
</div>

<!-- Reject Modal -->
<AnimatedPresence show={showRejectModal} type="scale" duration={200}>
	<div onkeydown={(e) => { if (e.key === 'Escape') cancelReject(); }} class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-4">Tolak Mutasi</h3>
				<div class="space-y-3">
					<label class="block text-sm font-medium text-gray-700 dark:text-gray-300">Alasan Penolakan</label>
					<textarea bind:value={rejectReason} rows="3" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 dark:bg-gray-700 dark:border-gray-600 dark:text-white resize-none" placeholder="Masukkan alasan penolakan"></textarea>
				</div>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelReject} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer dark:border-gray-600 dark:text-gray-300">Batal</button>
					<button onclick={handleReject} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition cursor-pointer">Tolak</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
