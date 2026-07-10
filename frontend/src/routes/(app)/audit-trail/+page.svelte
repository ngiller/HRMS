<script lang="ts">
	import { onMount } from 'svelte';
	import { activityLogs, employees, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';

	let loading = $state(true);
	let loadingMore = $state(false);
	let items = $state<any[]>([]);
	let total = $state(0);
	let pageNum = $state(1);
	const perPage = 10;
	let error = $state('');
	let totalPages = $derived(Math.max(1, Math.ceil(total / perPage)));

	// Filters
	let filterAction = $state('');
	let filterEntityType = $state('');
	let filterEmployeeId = $state('');
	let filterStartDate = $state('');
	let filterEndDate = $state('');

	// Dropdown options
	let entityTypes = $state<string[]>([]);
	let actions = $state<string[]>([]);
	let employeeList = $state<any[]>([]);

	async function loadEmployees() {
		try {
			const res = await employees.list(1, 200);
			if (res.success) employeeList = res.data || [];
		} catch {}
	}

	async function loadFilterOptions() {
		try {
			const [typesRes, actionsRes] = await Promise.all([
				activityLogs.getEntityTypes(),
				activityLogs.getActions()
			]);
			if (typesRes.success) entityTypes = typesRes.data?.entity_types || [];
			if (actionsRes.success) actions = actionsRes.data?.actions || [];
		} catch {}
	}

	onMount(() => {
		loadEmployees();
		loadFilterOptions();
		load();
	});

	async function load(resetPage = false) {
		loading = true;
		error = '';
		if (resetPage) pageNum = 1;

		try {
			const res = await activityLogs.list({
				page: pageNum,
				perPage,
				action: filterAction,
				entityType: filterEntityType,
				userId: filterEmployeeId,
				startDate: filterStartDate || undefined,
				endDate: filterEndDate || undefined,
			});
			if (res.success) {
				items = res.data.logs || [];
				total = res.data.total || 0;

			}
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
			else error = 'Gagal memuat log aktivitas';
		} finally {
			loading = false;
		}
	}

	async function goToPage(page: number) {
		if (page < 1 || page > totalPages || page === pageNum) return;
		pageNum = page;
		load();
	}
	function resetFilters() {
		filterAction = '';
		filterEntityType = '';
		filterEmployeeId = '';
		filterStartDate = '';
		filterEndDate = '';
		load(true);
	}

	function search() { load(true); }


	function formatDate(dateStr: string) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('id-ID', {
			day: 'numeric', month: 'short', year: 'numeric',
			hour: '2-digit', minute: '2-digit'
		});
	}

	function getActionBadge(action: string): string {
		const badges: Record<string, string> = {
			create: 'bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300',
			update: 'bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
			delete: 'bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
			approve: 'bg-emerald-100 text-emerald-800 dark:bg-emerald-900/40 dark:text-emerald-300',
			reject: 'bg-rose-100 text-rose-800 dark:bg-rose-900/40 dark:text-rose-300',
			pay: 'bg-purple-100 text-purple-800 dark:bg-purple-900/40 dark:text-purple-300',
			cancel: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300',
			login: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/40 dark:text-indigo-300',
			logout: 'bg-orange-100 text-orange-800 dark:bg-orange-900/40 dark:text-orange-300',
		};
		return badges[action] || 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-300';
	}

</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Audit Trail</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Log aktivitas sistem untuk audit dan pelacakan</p>
		</div>
	</div>

	<!-- Filters -->
	<div class="mb-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-lg border border-gray-200 dark:border-gray-700">
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-5 gap-3">
			<div>
				<label for="filter-action" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Aksi</label>
				<select id="filter-action"
					bind:value={filterAction}
					class="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
				>
					<option value="">Semua Aksi</option>
					{#each actions as a}
						<option value={a}>{a}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="filter-entity" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Entitas</label>
				<select id="filter-entity"
					bind:value={filterEntityType}
					class="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
				>
					<option value="">Semua Entitas</option>
					{#each entityTypes as et}
						<option value={et}>{et}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="filter-employee" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Karyawan</label>
				<select id="filter-employee"
					bind:value={filterEmployeeId}
					class="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
				>
					<option value="">Semua Karyawan</option>
					{#each employeeList as emp}
						<option value={emp.id}>{emp.full_name}</option>
					{/each}
				</select>
			</div>
			<div>
				<label for="filter-start-date" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Dari Tanggal</label>
				<input id="filter-start-date"
					type="date"
					bind:value={filterStartDate}
					class="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
				/>
			</div>
			<div>
				<label for="filter-end-date" class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">Sampai Tanggal</label>
				<input id="filter-end-date"
					type="date"
					bind:value={filterEndDate}
					class="w-full px-3 py-2 rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
				/>
			</div>
		</div>
		<div class="flex gap-2 mt-3">
			<button
				onclick={search}
				class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium cursor-pointer"
			>Cari</button>
			<button
				onclick={resetFilters}
				class="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm cursor-pointer"
			>Reset</button>
		</div>
	</div>

	{#if error}
		<div class="mb-4 p-3 bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-300 rounded-lg text-sm">{error}</div>
	{/if}

	<!-- Stats -->
	<div class="mb-4 flex items-center gap-4 text-sm text-gray-500 dark:text-gray-400">
		<span>Total: <strong>{total}</strong> log</span>
	</div>

	<PullToRefresh onRefresh={load}>
	{#if loading}
		<PulseLoader variant="text" count={8} />
	{:else if items.length === 0}
		<div class="text-center py-16 text-gray-400 dark:text-gray-500">
			<span class="text-5xl block mb-4">📋</span>
			<p class="text-lg">Belum ada log aktivitas</p>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden overflow-x-auto">
			<table class="w-full text-left border-collapse whitespace-nowrap">
				<thead>
					<tr class="bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
						<th class="p-4">Waktu</th>
						<th class="p-4">Karyawan</th>
						<th class="p-4">Aksi</th>
						<th class="p-4">Entitas</th>
						<th class="p-4">Nama Entitas</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-gray-100 dark:divide-gray-700/50 text-sm">
					{#each items as item}
						<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
							<td class="p-4 text-gray-500 dark:text-gray-400">{formatDate(item.created_at)}</td>
							<td class="p-4 font-medium text-gray-900 dark:text-white">{item.employee_name || '-'}</td>
							<td class="p-4">
								<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {getActionBadge(item.action)}">
									{item.action}
								</span>
							</td>
							<td class="p-4 text-gray-500 dark:text-gray-300">{item.entity_type}</td>
							<td class="p-4 text-gray-900 dark:text-white font-medium">{item.entity_name || '-'}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>

		<!-- Pagination -->
		<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 rounded-b-xl border border-gray-200 dark:border-gray-700 shadow-sm mt-[-4px]">
			<div class="text-xs text-gray-500 dark:text-gray-400">
				Menampilkan {total === 0 ? 0 : (pageNum - 1) * perPage + 1}-{Math.min(pageNum * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span>
			</div>
			<div class="flex items-center gap-1.5">
				<button onclick={() => goToPage(pageNum - 1)} disabled={pageNum <= 1}
					class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
				{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
					{@const pageIdx = Math.max(1, Math.min(pageNum - 2, totalPages - 4)) + i}
					{#if pageIdx <= totalPages}
						<button onclick={() => goToPage(pageIdx)}
							class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageIdx === pageNum ? 'bg-blue-600 text-white border-blue-600 shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}">{pageIdx}</button>
					{/if}
				{/each}
				<button onclick={() => goToPage(pageNum + 1)} disabled={pageNum >= totalPages}
					class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
			</div>
		</div>
	{/if}
	</PullToRefresh>
</div>
