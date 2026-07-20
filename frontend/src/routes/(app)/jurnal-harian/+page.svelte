<script lang="ts">
	import { onMount } from 'svelte';
	import { dailyJournals, employees, departments, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme } from '$lib/avatar-theme.js';

	interface JournalItem {
		id: string; employee_name: string; department_name: string;
		journal_date: string; work_description: string; achievements: string;
		challenges: string; plan_tomorrow: string; status: string;
		acknowledged_by_name: string; submitted_at: string;
	}

	interface DeptOption { id: string; name: string; }

	let data = $state<JournalItem[]>([]);
	let total = $state(0);
	let currentPage = $state(1);
	let perPage = $state(25);
	let isLoading = $state(true);
	let deptFilter = $state('');
	let dateFrom = $state('');
	let dateTo = $state('');
	let errorMessage = $state('');
	let searchQuery = $state('');
	let searchTimeout: ReturnType<typeof setTimeout>;

	let showForm = $state(false);
	let createForm = $state({ journal_date: new Date().toISOString().split('T')[0], work_description: '', achievements: '', challenges: '', plan_tomorrow: '' });
	let createLoading = $state(false);
	let deptOptions = $state<DeptOption[]>([]);

	let showDetail = $state(false);
	let filteredData = $derived.by(() => {
		if (!searchQuery.trim()) return data;
		const q = searchQuery.toLowerCase();
		return data.filter(i =>
			i.employee_name?.toLowerCase().includes(q) ||
			i.work_description?.toLowerCase().includes(q) ||
			i.department_name?.toLowerCase().includes(q)
		);
	});

	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; }, 400);
	}
	let detailItem = $state<JournalItem | null>(null);
	let detailLoading = $state(false);

	let actionLoading = $state(false);

	onMount(() => { loadData(); loadDepts(); });

	async function loadDepts() {
		try { const res = await departments.getAll(); if (res?.success) deptOptions = res.data; } catch { /* silent fail */ }
	}

	async function loadData() {
		isLoading = true; errorMessage = '';
		try {
			const res = await dailyJournals.list(currentPage, perPage, deptFilter, '', dateFrom, dateTo) as { success: boolean; data: JournalItem[]; meta?: { total: number } };
			if (res?.success) { data = res.data || []; total = res.meta?.total || 0; }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal memuat data'; }
		finally { isLoading = false; }
	}

	function onPageChange(page: number) { currentPage = page; loadData(); }

	async function loadDetail(id: string) {
		showForm = false;
		if (showDetail) { showDetail = false; await new Promise(r => setTimeout(r, 50)); }
		detailLoading = true; showDetail = true;
		try { const res = await dailyJournals.get(id); if (res?.success) detailItem = res.data; } catch { detailItem = null; }
		finally { detailLoading = false; }
	}

	function closeDetail() { showDetail = false; detailItem = null; }

	function openForm() {
		showDetail = false;
		createForm = { journal_date: new Date().toISOString().split('T')[0], work_description: '', achievements: '', challenges: '', plan_tomorrow: '' };
		errorMessage = ''; showForm = true;
	}

	function cancelForm() { showForm = false; errorMessage = ''; }

	async function handleCreate() {
		createLoading = true; errorMessage = '';
		try {
			const res = await dailyJournals.create(createForm);
			if (res?.success) { showForm = false; loadData(); }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal membuat jurnal' }
		finally { createLoading = false; }
	}

	async function handleAcknowledge(id: string) {
		actionLoading = true;
		try {
			const res = await dailyJournals.acknowledge(id);
			if (res?.success) { loadData(); loadDetail(id); }
		} catch (err) { errorMessage = err instanceof ApiError ? err.message : 'Gagal mengakui jurnal'; }
		finally { actionLoading = false; }
	}

	const totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	const statusColors: Record<string, string> = {
		draft: 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300',
		submitted: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
		acknowledged: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
	};

	function getStatusBadge(status: string) {
		const labels: Record<string, string> = { draft: 'Draft', submitted: 'Terkirim', acknowledged: 'Diketahui' };
		return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${statusColors[status] || 'bg-gray-100 text-gray-600'}">${labels[status] || status}</span>`;
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
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Jurnal Harian</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Catat dan pantau aktivitas kerja harian karyawan</p>
		</div>
		{#if !showForm}
			<button onclick={openForm} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Isi Jurnal
			</button>
		{/if}
	</div>

	{#if showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm mb-6">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Jurnal Harian Baru</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<form onsubmit={(e) => { e.preventDefault(); handleCreate(); }} class="px-6 py-5 space-y-4">
				{#if errorMessage}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-2.5">{errorMessage}</div>{/if}
				<div>
					<label for="journal-date" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tanggal <span class="text-red-500">*</span></label>
					<input id="journal-date" type="date" bind:value={createForm.journal_date} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20 max-w-xs" />
				</div>
				<div>
					<label for="journal-work" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi Pekerjaan <span class="text-red-500">*</span></label>
					<textarea id="journal-work" bind:value={createForm.work_description} required rows={4} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20 resize-none" placeholder="Apa yang dikerjakan hari ini?"></textarea>
				</div>
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label for="journal-achieve" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Capaian</label>
						<textarea id="journal-achieve" bind:value={createForm.achievements} rows={3} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none resize-none" placeholder="Capaian hari ini"></textarea>
					</div>
					<div>
						<label for="journal-challenge" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Kendala</label>
						<textarea id="journal-challenge" bind:value={createForm.challenges} rows={3} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none resize-none" placeholder="Kendala yang dihadapi"></textarea>
					</div>
				</div>
				<div>
					<label for="journal-plan" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Rencana Besok</label>
					<textarea id="journal-plan" bind:value={createForm.plan_tomorrow} rows={2} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none resize-none" placeholder="Rencana kerja untuk besok"></textarea>
				</div>
				<div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
					<button type="button" onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button type="submit" disabled={createLoading} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
						{#if createLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Submit Jurnal
					</button>
				</div>
			</form>
		</div>
	{:else if showDetail}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Detail Jurnal Harian</h2>
				<button onclick={closeDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5">
				{#if detailLoading}
					<PulseLoader variant="text" count={1} />
				{:else if detailItem}
					{@const item = detailItem}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi</h3>
							<div class="space-y-3">
								<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.employee_name}</p></div>
								<div><span class="text-xs text-gray-400">Tanggal</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{formatDate(item.journal_date)}</p></div>
								<div><span class="text-xs text-gray-400">Departemen</span><p class="text-sm text-gray-700 dark:text-gray-300">{item.department_name || '-'}</p></div>
								<div><span class="text-xs text-gray-400">Status</span><p>{@html getStatusBadge(item.status)}</p></div>
							</div>
						</div>
						<div>
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Detail</h3>
							<div class="space-y-3">
								{#if item.acknowledged_by_name}<div><span class="text-xs text-gray-400">Diketahui Oleh</span><p class="text-sm font-medium text-green-600">{item.acknowledged_by_name}</p></div>{/if}
								{#if item.submitted_at}<div><span class="text-xs text-gray-400">Dikirim</span><p class="text-sm text-gray-500">{formatDate(item.submitted_at)}</p></div>{/if}
							</div>
						</div>
					</div>
					<div class="mt-6 space-y-4">
						<div class="p-4 bg-gray-50 dark:bg-gray-800/30 rounded-lg">
							<h4 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">Deskripsi Pekerjaan</h4>
							<p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{item.work_description}</p>
						</div>
						{#if item.achievements}
							<div class="p-4 bg-green-50 dark:bg-green-900/10 rounded-lg">
								<h4 class="text-sm font-semibold text-green-700 dark:text-green-400 mb-2">🏆 Capaian</h4>
								<p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{item.achievements}</p>
							</div>
						{/if}
						{#if item.challenges}
							<div class="p-4 bg-yellow-50 dark:bg-yellow-900/10 rounded-lg">
								<h4 class="text-sm font-semibold text-yellow-700 dark:text-yellow-400 mb-2">⚠️ Kendala</h4>
								<p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{item.challenges}</p>
							</div>
						{/if}
						{#if item.plan_tomorrow}
							<div class="p-4 bg-blue-50 dark:bg-blue-900/10 rounded-lg">
								<h4 class="text-sm font-semibold text-blue-700 dark:text-blue-400 mb-2">📋 Rencana Besok</h4>
								<p class="text-sm text-gray-600 dark:text-gray-400 whitespace-pre-wrap">{item.plan_tomorrow}</p>
							</div>
						{/if}
					</div>
					<div class="mt-6 pt-4 border-t border-gray-100 dark:border-gray-800">
						{#if item.status === 'submitted'}
							<button onclick={() => handleAcknowledge(item.id)} disabled={actionLoading} class="w-full py-2.5 bg-green-600 text-white rounded-lg text-sm font-semibold hover:bg-green-700 transition disabled:opacity-50 cursor-pointer">
								{actionLoading ? 'Memproses...' : '✓ Tandai Diketahui'}
							</button>
						{/if}
					</div>
				{/if}
			</div>
		</div>
	{:else}
		<div class="flex flex-wrap gap-3 mb-4" role="group" aria-label="Filter jurnal harian">
		<div class="relative flex-1 max-w-md">
			<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
			<input type="search" value={searchQuery} placeholder="Cari jurnal..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:focus:bg-gray-900 transition placeholder:text-gray-400" />
		</div>
		
			<select name="dept_filter" bind:value={deptFilter} onchange={() => { currentPage = 1; loadData(); }}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none"
				aria-label="Filter departemen">
				<option value="">Semua Departemen</option>
				{#each deptOptions as dept (dept.id)}<option value={dept.id}>{dept.name}</option>{/each}
			</select>
			<input name="date_from" type="date" bind:value={dateFrom} onchange={() => { currentPage = 1; loadData(); }} class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" placeholder="Dari tanggal" aria-label="Dari tanggal" />
			<input name="date_to" type="date" bind:value={dateTo} onchange={() => { currentPage = 1; loadData(); }} class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" placeholder="Sampai tanggal" aria-label="Sampai tanggal" />
		</div>
		{#if errorMessage}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{errorMessage}</div>{/if}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
			<!-- Desktop Table -->
			<div class="hidden md:block">
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
							<tr>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tanggal</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Karyawan</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Departemen</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Pekerjaan</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Aksi</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#if isLoading}
									<PulseLoader variant="table-row" count={5} />
							{:else if data.length === 0}
								<tr><td colspan="6" class="px-4 py-8 text-center text-sm text-gray-400">Belum ada jurnal harian</td></tr>
							{:else}
								{#each filteredData as item (item)}
									<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition">
										<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{formatDate(item.journal_date)}</td>
										<td class="px-4 py-3 text-gray-700 dark:text-gray-300">{item.employee_name}</td>
										<td class="px-4 py-3 text-gray-500 dark:text-gray-400">{item.department_name || '-'}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400 max-w-xs truncate">{item.work_description}</td>
										<td class="px-4 py-3">{@html getStatusBadge(item.status)}</td>
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

			<!-- Mobile Card List (Talenta Style) -->
			<div class="md:hidden p-3 space-y-3">
				{#if isLoading}
					<PulseLoader variant="card" count={3} />
				{:else if data.length === 0}
					<EmptyState
						variant="empty"
						title="Belum ada jurnal harian"
						description="Belum ada jurnal harian yang dicatat."
					/>
				{:else}
					{#each filteredData as item (item)}
						{@const theme = getAvatarTheme('dailyJournal')}
						<MobileCard
							avatar={item.employee_name}
							avatarColor={theme.gradientClasses}
							title={item.employee_name}
							subtitle={item.department_name || 'Tanpa departemen'}
							badges={[{ label: item.status === 'draft' ? 'Draft' : item.status === 'submitted' ? 'Terkirim' : item.status === 'acknowledged' ? 'Diketahui' : item.status, color: statusColors[item.status] || 'bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-300' }]}
							onclick={() => loadDetail(item.id)}
							clickable={true}
						>
							{#snippet children()}
								<div class="flex items-center gap-2 text-[11px] text-gray-500 dark:text-gray-400 mb-2">
									<div class="flex items-center gap-1.5 px-2 py-1 bg-gray-50 dark:bg-gray-800 rounded-md">
										<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
										</svg>
										<span class="tabular-nums">{formatDate(item.journal_date)}</span>
									</div>
								</div>
								<div class="text-xs text-gray-600 dark:text-gray-400 line-clamp-2 leading-relaxed">
									{item.work_description}
								</div>
							{/snippet}
							{#snippet footer()}
								<div class="flex justify-end pt-2">
									<span class="text-xs font-medium text-[#1A56DB] dark:text-blue-400">Lihat Detail →</span>
								</div>
							{/snippet}
						</MobileCard>
					{/each}

					<!-- Mobile Pagination -->
					<div class="flex items-center justify-between px-1 py-2">
						<span class="text-xs text-gray-400">{(currentPage - 1) * perPage + 1}-{Math.min(currentPage * perPage, total)} dari {total}</span>
						<div class="flex gap-2">
							<button onclick={() => onPageChange(currentPage - 1)} disabled={currentPage <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer active:scale-95">Prev</button>
							<button onclick={() => onPageChange(currentPage + 1)} disabled={currentPage >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-30 disabled:cursor-not-allowed transition cursor-pointer active:scale-95">Next</button>
						</div>
					</div>
				{/if}
			</div>
		</div>
	{/if}
</div>
