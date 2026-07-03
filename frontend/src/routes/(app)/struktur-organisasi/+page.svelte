<script lang="ts">
	import { onMount } from 'svelte';
	import { organization as api } from '$lib/api.js';

	type OrgNode = {
		id: string;
		label: string;
		type: 'department' | 'position' | 'employee';
		meta?: Record<string, any>;
		children?: OrgNode[];
	};

	let tree = $state<OrgNode[]>([]);
	let isLoading = $state(true);
	let errorMessage = $state('');
	let searchQuery = $state('');
	let expandedIds = $state<Set<string>>(new Set());

	let stats = $state({ departments: 0, positions: 0, employees: 0 });

	onMount(() => { load(); });

	async function load() {
		isLoading = true;
		errorMessage = '';
		try {
			const response: any = await api.getTree();
			const data: OrgNode[] = response.data || [];
			tree = data;
			expandAllNodes(data);
			countStats(data);
		} catch (error: any) {
			errorMessage = error.message || 'Gagal memuat data struktur organisasi';
		} finally {
			isLoading = false;
		}
	}

	function expandAllNodes(nodes: OrgNode[]) {
		const next = new Set(expandedIds);
		function walk(list: OrgNode[]) {
			for (const node of list) {
				next.add(node.id);
				if (node.children) walk(node.children);
			}
		}
		walk(nodes);
		expandedIds = next;
	}

	function countStats(nodes: OrgNode[]) {
		let depts = 0, pos = 0, emps = 0;
		function walk(nodes: OrgNode[]) {
			for (const n of nodes) {
				if (n.type === 'department') depts++;
				else if (n.type === 'position') pos++;
				else if (n.type === 'employee') emps++;
				if (n.children) walk(n.children);
			}
		}
		walk(nodes);
		stats = { departments: depts, positions: pos, employees: emps };
	}

	function toggle(id: string) {
		const next = new Set(expandedIds);
		if (next.has(id)) next.delete(id);
		else next.add(id);
		expandedIds = next;
	}

	function expandAll() {
		const all = new Set<string>();
		function walk(nodes: OrgNode[]) {
			for (const n of nodes) {
				all.add(n.id);
				if (n.children) walk(n.children);
			}
		}
		walk(tree);
		expandedIds = all;
	}

	function collapseAll() {
		expandedIds = new Set();
	}

	function matchesSearch(node: OrgNode, q: string): boolean {
		if (!q) return true;
		const query = q.toLowerCase();
		if (node.label.toLowerCase().includes(query)) return true;
		if (node.meta?.head_name && String(node.meta.head_name).toLowerCase().includes(query)) return true;
		if (node.meta?.grade_name && String(node.meta.grade_name).toLowerCase().includes(query)) return true;
		if (node.meta?.employee_id && String(node.meta.employee_id).toLowerCase().includes(query)) return true;
		if (node.children) return node.children.some(c => matchesSearch(c, query));
		return false;
	}

	function getFilteredTree(nodes: OrgNode[]): OrgNode[] {
		if (!searchQuery) return nodes;
		return nodes
			.filter(n => matchesSearch(n, searchQuery))
			.map(n => ({
				...n,
				children: n.children ? getFilteredTree(n.children) : undefined,
			}));
	}

	function nodeColor(type: string): string {
		if (type === 'department') return 'bg-blue-50 text-blue-700 ring-blue-150 border-blue-200 dark:bg-blue-900/20 dark:text-blue-400 dark:border-blue-900/50';
		if (type === 'position') return 'bg-amber-50 text-amber-700 ring-amber-150 border-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:border-amber-900/50';
		return 'bg-emerald-50 text-emerald-700 ring-emerald-150 border-emerald-200 dark:bg-emerald-900/20 dark:text-emerald-400 dark:border-emerald-900/50';
	}

	function getInitials(name: string): string {
		if (!name) return 'EMP';
		const parts = name.trim().split(/\s+/);
		if (parts.length >= 2) {
			return (parts[0][0] + parts[1][0]).toUpperCase();
		}
		return parts[0][0].toUpperCase();
	}

	function getAvatarColor(name: string): string {
		let hash = 0;
		for (let i = 0; i < name.length; i++) {
			hash = name.charCodeAt(i) + ((hash << 5) - hash);
		}
		const colors = [
			'bg-blue-500 text-white',
			'bg-emerald-500 text-white',
			'bg-violet-500 text-white',
			'bg-indigo-500 text-white',
			'bg-rose-500 text-white',
			'bg-amber-500 text-white',
			'bg-teal-500 text-white',
			'bg-cyan-500 text-white',
		];
		const index = Math.abs(hash) % colors.length;
		return colors[index];
	}

	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			searchQuery = target.value;
		}, 300);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 tracking-tight">Struktur Organisasi</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Bagan hierarki departemen, posisi jabatan, dan karyawan</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={expandAll} class="px-3.5 py-2 text-xs font-semibold border border-gray-200 dark:border-gray-700 rounded-xl text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Perluas Semua</button>
			<button onclick={collapseAll} class="px-3.5 py-2 text-xs font-semibold border border-gray-200 dark:border-gray-700 rounded-xl text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Ciutkan Semua</button>
		</div>
	</div>

	<!-- Stats Widget Cards -->
	{#if !isLoading && !errorMessage}
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 flex items-center gap-4 shadow-sm">
				<div class="w-11 h-11 rounded-xl bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 flex items-center justify-center shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
				</div>
				<div>
					<p class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Total Departemen</p>
					<p class="text-xl font-bold text-gray-900 dark:text-gray-100 mt-0.5">{stats.departments}</p>
				</div>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 flex items-center gap-4 shadow-sm">
				<div class="w-11 h-11 rounded-xl bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 flex items-center justify-center shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
				</div>
				<div>
					<p class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Posisi Jabatan</p>
					<p class="text-xl font-bold text-gray-900 dark:text-gray-100 mt-0.5">{stats.positions}</p>
				</div>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-4 flex items-center gap-4 shadow-sm">
				<div class="w-11 h-11 rounded-xl bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 flex items-center justify-center shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
				</div>
				<div>
					<p class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wide">Karyawan Aktif</p>
					<p class="text-xl font-bold text-gray-900 dark:text-gray-100 mt-0.5">{stats.employees}</p>
				</div>
			</div>
		</div>
	{/if}

	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl overflow-hidden shadow-sm">
		<!-- Search Header -->
		<div class="px-5 py-4 border-b border-gray-150 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/30">
			<div class="relative max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				<input type="search" value={searchQuery} placeholder="Cari departemen, posisi, atau karyawan..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-white dark:bg-gray-850 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400" />
			</div>
		</div>

		{#if isLoading}
			<div class="p-6 animate-pulse space-y-4">
				{#each [1, 2, 3] as _}
					<div class="flex items-center gap-3">
						<div class="h-6 w-6 bg-gray-100 dark:bg-gray-800 rounded-lg"></div>
						<div class="h-5 bg-gray-100 dark:bg-gray-800 rounded w-48"></div>
					</div>
					<div class="ml-8 space-y-3">
						{#each [1, 2] as _}
							<div class="flex items-center gap-3">
								<div class="h-5 w-5 bg-gray-50 dark:bg-gray-850 rounded"></div>
								<div class="h-4 bg-gray-50 dark:bg-gray-850 rounded w-36"></div>
							</div>
						{/each}
					</div>
				{/each}
			</div>
		{:else if errorMessage}
			<div class="py-16 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 dark:bg-red-900/10 flex items-center justify-center">
					<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
				</div>
				<p class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">Gagal memuat data</p>
				<p class="text-xs text-gray-500 mb-4">{errorMessage}</p>
				<button onclick={load} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-xs font-semibold hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
			</div>
		{:else}
			<div class="p-6">
				{#each getFilteredTree(tree) as node (node.id)}
					{@render treeNode(node)}
				{/each}
				{#if getFilteredTree(tree).length === 0}
					<div class="py-16 text-center">
						<svg class="w-12 h-12 mx-auto mb-3 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
						<p class="text-sm text-gray-500 dark:text-gray-400">{searchQuery ? `Tidak ditemukan dengan kata kunci "${searchQuery}"` : 'Belum ada data struktur organisasi'}</p>
					</div>
				{/if}
			</div>
		{/if}
	</div>
</div>

{#snippet treeNode(node: OrgNode)}
	{@const isExpanded = expandedIds.has(node.id)}
	{@const hasChildren = node.children && node.children.length > 0}
	
	<div class="mb-2">
		<!-- svelte-ignore a11y_no_static_element_interactions -->
		<div
			class="flex items-center gap-3 px-4 py-2.5 rounded-xl border border-transparent hover:border-gray-200 dark:hover:border-gray-800 cursor-pointer transition-all duration-150 {node.type === 'employee' ? 'bg-gray-50/50 dark:bg-gray-850/30' : ''}"
			role="button"
			tabindex="0"
			onclick={() => toggle(node.id)}
			onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); toggle(node.id); } }}
		>
			<!-- Expand/Collapse Chevron Indicator -->
			{#if hasChildren}
				<svg class="w-3.5 h-3.5 text-gray-400 transition-transform duration-200 {isExpanded ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="3" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M8.25 4.5l7.5 7.5-7.5 7.5" />
				</svg>
			{:else}
				<span class="w-3.5"></span>
			{/if}

			<!-- Node Icon or Letter Avatar -->
			{#if node.type === 'employee'}
				<div class="w-8 h-8 rounded-full flex items-center justify-center text-xs font-bold shrink-0 {getAvatarColor(node.label)}">
					{getInitials(node.label)}
				</div>
			{:else}
				<div class="w-8 h-8 rounded-xl flex items-center justify-center border shrink-0 {nodeColor(node.type)}">
					{#if node.type === 'department'}
						<svg class="w-4.5 h-4.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
					{:else}
						<svg class="w-4.5 h-4.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
					{/if}
				</div>
			{/if}

			<!-- Node Details -->
			<div class="flex-1 min-w-0 flex items-center justify-between">
				<div>
					<span class="text-sm font-semibold text-gray-900 dark:text-gray-100">{node.label}</span>
					
					<!-- Metadata Badges -->
					{#if node.type === 'department' && node.meta?.head_name}
						<span class="text-xs text-gray-400 dark:text-gray-500 ml-2 hidden sm:inline">— Kepala: <strong class="text-gray-600 dark:text-gray-300 font-medium">{node.meta.head_name}</strong></span>
					{/if}
					{#if node.type === 'position' && node.meta?.grade_name}
						<span class="ml-2.5 inline-flex items-center px-1.5 py-0.5 rounded-md text-[10px] font-bold bg-amber-50 dark:bg-amber-900/10 text-amber-700 dark:text-amber-400 border border-amber-200/50 dark:border-amber-900/30 uppercase tracking-wide">
							{node.meta.grade_name}
						</span>
					{/if}
					{#if node.type === 'employee'}
						{#if node.meta?.employee_id}
							<span class="text-xs text-gray-400 dark:text-gray-500 ml-2">#{node.meta.employee_id}</span>
						{/if}
						{#if node.meta?.join_date}
							<span class="text-[10px] text-gray-400 dark:text-gray-500 ml-3 hidden sm:inline-flex items-center gap-1">
								<svg class="w-3 h-3 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
								Bergabung: {formatDate(node.meta.join_date)}
							</span>
						{/if}
					{/if}
				</div>

				<!-- Right indicators -->
				<div class="flex items-center gap-2">
					{#if node.type === 'employee'}
						<!-- Status indicator -->
						<span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-bold bg-emerald-50 dark:bg-emerald-950/30 text-emerald-700 dark:text-emerald-400 border border-emerald-200/50 dark:border-emerald-900/30">
							<span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
							Aktif
						</span>
					{/if}
					{#if hasChildren}
						<span class="text-[10px] font-bold px-2 py-0.5 bg-gray-100 dark:bg-gray-800 text-gray-500 dark:text-gray-400 rounded-md tabular-nums">
							{node.children!.length} cabang
						</span>
					{/if}
				</div>
			</div>
		</div>

		<!-- Expandable Child Nodes Tree Guide -->
		{#if isExpanded && hasChildren}
			<div class="ml-5 border-l-2 border-dashed border-gray-200 dark:border-gray-800 pl-4 mt-0.5">
				{#each node.children as child (child.id)}
					{@render treeNode(child)}
				{/each}
			</div>
		{/if}
	</div>
{/snippet}
