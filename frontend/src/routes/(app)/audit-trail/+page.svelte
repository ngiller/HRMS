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
	const perPage = 25;
	let error = $state('');
	let hasMore = $state(true);
	let sentinelEl = $state<HTMLDivElement>();
	let observer: IntersectionObserver | null = null;

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
		pageNum = 1;
		hasMore = true;
		try {
			const res = await activityLogs.list({
				page: 1,
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
				hasMore = items.length < total;
			}
		} catch (e) {
			if (e instanceof ApiError) error = e.message;
			else error = 'Gagal memuat log aktivitas';
		} finally {
			loading = false;
		}
	}

	async function loadMore() {
		if (loadingMore || !hasMore) return;
		loadingMore = true;
		try {
			const nextPage = pageNum + 1;
			const res = await activityLogs.list({
				page: nextPage,
				perPage,
				action: filterAction,
				entityType: filterEntityType,
				userId: filterEmployeeId,
				startDate: filterStartDate || undefined,
				endDate: filterEndDate || undefined,
			});
			if (res.success) {
				const newItems = res.data.logs || [];
				items = [...items, ...newItems];
				pageNum = nextPage;
				hasMore = items.length < total;
			}
		} catch {
			// Silent fail on load more
		} finally {
			loadingMore = false;
		}
	}

	function resetFilters() {
		filterAction = '';
		filterEntityType = '';
		filterEmployeeId = '';
		filterStartDate = '';
		filterEndDate = '';
		load();
	}

	function search() { load(); }

	$effect(() => {
		if (sentinelEl && !loading) {
			observer?.disconnect();
			observer = new IntersectionObserver((entries) => {
				if (entries[0].isIntersecting && hasMore && !loadingMore) {
					loadMore();
				}
			}, { rootMargin: '200px' });
			observer.observe(sentinelEl);
		}
		return () => observer?.disconnect();
	});

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
		<div class="space-y-2">
			{#each items as item, i}
				<AnimatedPresence show={true} delay={Math.min(i * 20, 300)}>
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
				</AnimatedPresence>
			{/each}
		</div>

		<!-- Infinite scroll sentinel -->
		{#if hasMore}
			<div bind:this={sentinelEl} class="flex items-center justify-center py-4">
				{#if loadingMore}
					<div class="flex items-center gap-2 text-sm text-gray-400">
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
						<span>Memuat data lainnya...</span>
					</div>
				{:else}
					<div class="w-6 h-6 rounded-full border-2 border-gray-200 dark:border-gray-700 border-dashed"></div>
				{/if}
			</div>
		{:else if items.length > 0}
			<div class="text-center py-4 text-xs text-gray-400 dark:text-gray-500">
				<span>Menampilkan semua {total} log</span>
			</div>
		{/if}
	{/if}
	</PullToRefresh>
</div>
