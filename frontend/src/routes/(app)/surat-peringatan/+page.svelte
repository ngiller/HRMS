<script lang="ts">
	import { onMount } from 'svelte';
	import { reprimands, employees, ApiError } from '$lib/api.js';

	let data = $state<any[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let perPage = $state(25);
	let isLoading = $state(true);
	let statusFilter = $state('');
	let errorMessage = $state('');

	let showForm = $state(false);
	let createForm = $state({ employee_id: '', reprimand_type: 'sp1', title: '', description: '', violation_date: '', violation_details: '', effective_period_months: 6 });
	let createLoading = $state(false);
	let employeeOptions = $state<any[]>([]);

	let showDetail = $state(false);
	let detailItem = $state<any>(null);
	let detailLoading = $state(false);

	let actionLoading = $state(false);

	onMount(() => { loadData(); loadEmployees(); });

	async function loadEmployees() {
		try { const res = await employees.list(1, 100); if (res?.success) employeeOptions = res.data; } catch {}
	}

	async function loadData() {
		isLoading = true; errorMessage = '';
		try {
			const res: any = await reprimands.list(currentPage, perPage, statusFilter);
			if (res?.success) { data = res.data || []; total = res.meta?.total || 0; }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal memuat data'; }
		finally { isLoading = false; }
	}

	function onPageChange(page: number) { currentPage = page; loadData(); }

	async function loadDetail(id: string) {
		showForm = false;
		if (showDetail) { showDetail = false; await new Promise(r => setTimeout(r, 50)); }
		detailLoading = true; showDetail = true;
		try { const res = await reprimands.get(id); if (res?.success) detailItem = res.data; } catch { detailItem = null; }
		finally { detailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailItem = null; }

	function openForm() {
		showDetail = false;
		createForm = { employee_id: '', reprimand_type: 'sp1', title: '', description: '', violation_date: '', violation_details: '', effective_period_months: 6 };
		errorMessage = ''; showForm = true;
	}

	function cancelForm() { showForm = false; errorMessage = ''; }

	async function handleCreate() {
		createLoading = true; errorMessage = '';
		try {
			const res = await reprimands.create(createForm);
			if (res?.success) { showForm = false; loadData(); }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal membuat surat peringatan'; }
		finally { createLoading = false; }
	}

	async function handleAcknowledge(id: string) {
		actionLoading = true;
		try {
			const res = await reprimands.acknowledge(id, { acknowledgment_note: '' });
			if (res?.success) { loadData(); loadDetail(id); }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal mengakui surat peringatan'; }
		finally { actionLoading = false; }
	}

	const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	const typeLabels: Record<string, string> = { verbal: 'Verbal', sp1: 'SP1', sp2: 'SP2', sp3: 'SP3' };
	const typeColors: Record<string, string> = {
		verbal: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
		sp1: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
		sp2: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
		sp3: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
	};
	const statusColors: Record<string, string> = {
		issued: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800',
		acknowledged: 'bg-green-50 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-200 dark:ring-green-800',
		expired: 'bg-gray-100 text-gray-500 ring-gray-200 dark:bg-gray-800 dark:text-gray-300 dark:ring-gray-600',
	};

	function getTypeBadge(type: string) {
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${typeColors[type] || 'bg-gray-100 text-gray-600'}">${typeLabels[type] || type}</span>`;
	}
	function getStatusBadge(status: string) {
		const labels: Record<string, string> = { issued: 'Diterbitkan', acknowledged: 'Diakui', expired: 'Kadaluarsa' };
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${statusColors[status] || 'bg-gray-50 text-gray-600'}">${labels[status] || status}</span>`;
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}
</script>

<div class="w-full">
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Surat Peringatan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Kelola surat peringatan karyawan (SP1/SP2/SP3)</p>
		</div>
		{#if !showForm}
			<button onclick={openForm} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Terbitkan SP
			</button>
		{/if}
	</div>

	{#if showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm mb-6">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Terbitkan Surat Peringatan Baru</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); handleCreate(); }} class="px-6 py-5 space-y-4">
				{#if errorMessage}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-2.5">{errorMessage}</div>{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="sp-employee" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Karyawan <span class="text-red-500">*</span></label>
						<select id="sp-employee" bind:value={createForm.employee_id} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20">
							<option value="">Pilih Karyawan</option>
							{#each employeeOptions as emp}<option value={emp.id}>{emp.full_name}</option>{/each}
						</select>
					</div>
					<div>
						<label for="sp-type" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe SP <span class="text-red-500">*</span></label>
						<select id="sp-type" bind:value={createForm.reprimand_type} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
							<option value="verbal">Verbal</option>
							<option value="sp1">SP1</option>
							<option value="sp2">SP2</option>
							<option value="sp3">SP3</option>
						</select>
					</div>
				</div>
				<div>
					<label for="sp-title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Judul <span class="text-red-500">*</span></label>
					<input id="sp-title" type="text" bind:value={createForm.title} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" placeholder="Contoh: Keterlambatan Berulang" />
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="sp-date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal Pelanggaran</label>
						<input id="sp-date" type="date" bind:value={createForm.violation_date} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" />
					</div>
					<div>
						<label for="sp-period" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Masa Berlaku (bulan)</label>
						<select id="sp-period" bind:value={createForm.effective_period_months} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
							<option value={3}>3 bulan</option>
							<option value={6}>6 bulan</option>
							<option value={12}>12 bulan</option>
						</select>
					</div>
				</div>
				<div>
					<label for="sp-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi Pelanggaran</label>
					<textarea id="sp-desc" bind:value={createForm.description} rows={3} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20 resize-none" placeholder="Deskripsi pelanggaran yang dilakukan"></textarea>
				</div>
				<div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
					<button type="button" onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button type="submit" disabled={createLoading} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
						{#if createLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Terbitkan
					</button>
				</div>
			</form>
		</div>
	{:else if showDetail}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Detail Surat Peringatan</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if detailLoading}
					<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-48"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-64"></div></div>
				{:else if detailItem}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi SP</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{detailItem.employee_name}</p></div>
								<div><span class="text-xs text-gray-400">Tipe</span><p>{@html getTypeBadge(detailItem.reprimand_type)}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(detailItem.status)}</p></div>
								<div><span class="text-xs text-gray-400">Judul</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{detailItem.title}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal Diterbitkan</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(detailItem.issued_date)}</p></div>
								{#if detailItem.expired_at}<div><span class="text-xs text-gray-400">Kadaluarsa</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(detailItem.expired_at)}</p></div>{/if}
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Detail</h3>
							<div class="space-y-3">
								{#if detailItem.description}<div><span class="text-xs text-gray-400">Deskripsi</span><p class="text-sm text-gray-700 dark:text-gray-300">{detailItem.description}</p></div>{/if}
								{#if detailItem.issued_by_name}<div><span class="text-xs text-gray-400">Diterbitkan Oleh</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{detailItem.issued_by_name}</p></div>{/if}
								{#if detailItem.acknowledgment_date}<div><span class="text-xs text-gray-400">Diakui Pada</span><p class="text-sm text-gray-700 dark:text-gray-300">{formatDate(detailItem.acknowledgment_date)}</p></div>{/if}
							</div>
						</div>
					</div>
					<div class="mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
						{#if detailItem.status === 'issued'}
							<button onclick={() => handleAcknowledge(detailItem.id)} disabled={actionLoading} class="w-full py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition disabled:opacity-50 cursor-pointer">
								{actionLoading ? 'Memproses...' : 'Akui (Tanda Terima)'}
							</button>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div class="flex gap-3 mb-4">
			<select bind:value={statusFilter} onchange={() => { currentPage = 1; loadData(); }}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
				<option value="">Semua Status</option>
				<option value="issued">Diterbitkan</option>
				<option value="acknowledged">Diakui</option>
				<option value="expired">Kadaluarsa</option>
			</select>
		</div>
		{#if errorMessage}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{errorMessage}</div>{/if}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
						<tr>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Karyawan</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tipe</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Judul</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tanggal</th>
							<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
						{#if isLoading}
							{#each [1,2,3,4,5] as _}
								<tr class="animate-pulse">
									<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-36"></div></td>
									<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-14"></div></td>
									<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-40"></div></td>
									<td class="px-4 py-3"><div class="h-5 bg-gray-100 dark:bg-gray-800 rounded-full w-20"></div></td>
									<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-16"></div></td>
									<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-12"></div></td>
								</tr>
							{/each}
						{:else if data.length === 0}
							<tr><td colspan="6" class="px-4 py-8 text-center text-sm text-gray-400">Belum ada surat peringatan</td></tr>
						{:else}
							{#each data as item}
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition {detailItem?.id === item.id && showDetail ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''}">
									<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{item.employee_name}</td>
									<td class="px-4 py-3">{@html getTypeBadge(item.reprimand_type)}</td>
									<td class="px-4 py-3 text-gray-700 dark:text-gray-300">{item.title}</td>
									<td class="px-4 py-3">{@html getStatusBadge(item.status)}</td>
									<td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{new Date(item.created_at).toLocaleDateString('id-ID')}</td>
									<td class="px-4 py-3"><button onclick={() => loadDetail(item.id)} class="text-xs text-[#1A56DB] hover:underline font-medium cursor-pointer">Detail</button></td>
								</tr>
							{/each}
						{/if}
					</tbody>
				</table>
			</div>
			<div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
				<span class="text-xs text-gray-400">Total {total} data</span>
				<div class="flex gap-1">
					<button onclick={() => onPageChange(currentPage - 1)} disabled={currentPage <= 1} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Prev</button>
					<span class="px-3 py-1 text-sm text-gray-500">{(currentPage - 1) * perPage + 1} - {Math.min(currentPage * perPage, total)}</span>
					<button onclick={() => onPageChange(currentPage + 1)} disabled={currentPage >= totalPages} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Next</button>
				</div>
			</div>
		</div>
	{/if}
</div>
