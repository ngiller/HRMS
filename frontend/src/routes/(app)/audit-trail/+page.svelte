<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { activityLogs, employees, ApiError } from '$lib/api.js';

	let loading = true;
	let items: any[] = [];
	let total = 0;
	let pageNum = 1;
	const perPage = 25;
	let error = '';

	// Filters
	let filterAction = '';
	let filterEntityType = '';
	let filterEmployeeId = '';
	let filterStartDate = '';
	let filterEndDate = '';

	// Dropdown options
	let entityTypes: string[] = [];
	let actions: string[] = [];
	let employeeList: any[] = [];

	async function loadEmployees() {
		try {
			const res = await employees.list(1, 200);
			if (res.success) employeeList = res.data.employees || [];
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

	async function load() {
		loading = true;
		error = '';
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

	function resetFilters() {
		filterAction = '';
		filterEntityType = '';
		filterEmployeeId = '';
		filterStartDate = '';
		filterEndDate = '';
		pageNum = 1;
		load();
	}

	function search() { pageNum = 1; load(); }

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

	// Skeleton loader array — Svelte doesn't support [...Array(n)] in template
	const skeletonItems = [1, 2, 3, 4, 5, 6, 7, 8];

	const totalPages = Math.max(1, Math.ceil(total / perPage));

	onMount(() => {
		load();
		loadEmployees();
		loadFilterOptions();
	});
	onDestroy(() => {});
</script>

<div class="p-4 md:p-6">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Audit Trail</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Log aktivitas sistem untuk audit dan pelacakan</p>
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
				on:click={search}
				class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
			>Cari</button>
			<button
				on:click={resetFilters}
				class="px-4 py-2 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors text-sm"
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

	{#if loading}
		<div class="space-y-2">
			{#each skeletonItems as _, i}
				<div class="animate-pulse bg-gray-200 dark:bg-gray-700 rounded-lg h-14"></div>
			{/each}
		</div>
	{:else if items.length === 0}
		<div class="text-center py-16 text-gray-400 dark:text-gray-500">
			<span class="text-5xl block mb-4">📋</span>
			<p class="text-lg">Belum ada log aktivitas</p>
		</div>
	{:else}
		<div class="space-y-2">
			{#each items as item}
				<div class="p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 hover:shadow-sm transition-shadow">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<span class="px-2 py-0.5 rounded-full text-xs font-medium {getActionBadge(item.action)}">
								{item.action}
							</span>
							<div>
								<span class="text-sm font-medium text-gray-900 dark:text-white">{item.entity_name || item.entity_type}</span>
								<span class="text-xs text-gray-400 dark:text-gray-500 ml-2">({item.entity_type})</span>
							</div>
						</div>
						<div class="text-right text-xs text-gray-400 dark:text-gray-500">
							<div>{formatDate(item.created_at)}</div>
							{#if item.employee_name}
								<div class="text-gray-500 dark:text-gray-400">{item.employee_name}</div>
							{/if}
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- Pagination -->
		{#if totalPages > 1}
			<div class="flex items-center justify-center gap-2 mt-6">
				<button
					on:click={() => { if (pageNum > 1) { pageNum--; load(); } }}
					disabled={pageNum <= 1}
					class="px-3 py-1.5 rounded text-sm disabled:opacity-40 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
				>Sebelumnya</button>
				<span class="text-sm text-gray-500 dark:text-gray-400">
					Halaman {pageNum} dari {totalPages}
				</span>
				<button
					on:click={() => { if (pageNum < totalPages) { pageNum++; load(); } }}
					disabled={pageNum >= totalPages}
					class="px-3 py-1.5 rounded text-sm disabled:opacity-40 bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-200 hover:bg-gray-200 dark:hover:bg-gray-600"
				>Selanjutnya</button>
			</div>
		{/if}
	{/if}
</div>
