<script lang="ts">
	import { onMount } from 'svelte';
	import { announcements as api } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import type { ApiResponse } from '$lib/types.js';

	type Announcement = {
		id: string;
		title: string;
		content: string;
		announcement_type: string;
		target_all: boolean;
		is_pinned: boolean;
		created_by_name: string;
		published_at: string;
		expired_at: string | null;
		created_at: string;
		read_count?: number;
	};

	type FormData = {
		title: string;
		content: string;
		announcement_type: string;
		target_all: boolean;
		is_pinned: boolean;
		expired_at: string;
	};

	const announcementTypes: Record<string, string> = {
		general: 'Umum',
		important: 'Penting',
		emergency: 'Darurat',
	};

	const typeColors: Record<string, string> = {
		general: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900/20 dark:text-blue-400 dark:ring-blue-800/30',
		important: 'bg-amber-50 text-amber-700 ring-amber-200 dark:bg-amber-900/20 dark:text-amber-400 dark:ring-amber-800/30',
		emergency: 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900/20 dark:text-red-400 dark:ring-red-800/30',
	};

	let items = $state<Announcement[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let typeFilter = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Side Panel Controls
	let showForm = $state(false);
	let formTitle = $state('');
	let isEditing = $state(false);
	let editId = $state<string | null>(null);
	let form = $state<FormData>({ title: '', content: '', announcement_type: 'general', target_all: true, is_pinned: false, expired_at: '' });
	let formError = $state('');
	let isSaving = $state(false);

	let showDetail = $state(false);
	let detailId = $state<string | null>(null);
	let detailData = $state<Announcement | null>(null);
	let isDetailLoading = $state(false);

	// Card Delete Actions
	let confirmDeleteId = $state<string | null>(null);

	onMount(() => {
		load();
	});

	async function load() {
		isLoading = true;
		errorMessage = '';
		try {
			const response = await api.list(page, perPage, typeFilter) as ApiResponse<Announcement[]>;
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.max(1, Math.ceil(total / perPage));
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		load();
	}

	function openCreateForm() {
		formTitle = 'Buat Pengumuman';
		isEditing = false;
		editId = null;
		form = { title: '', content: '', announcement_type: 'general', target_all: true, is_pinned: false, expired_at: '' };
		formError = '';
		showForm = true;
		showDetail = false;
	}

	async function openEdit(id: string) {
		isEditing = true;
		editId = id;
		formTitle = 'Edit Pengumuman';
		formError = '';
		showForm = true;
		showDetail = false;
		try {
			const response = await api.get(id) as ApiResponse<Announcement>;
			const d = response.data ?? {} as Announcement;
			form = {
				title: d.title || '',
				content: d.content || '',
				announcement_type: d.announcement_type || 'general',
				target_all: d.target_all ?? true,
				is_pinned: d.is_pinned || false,
				expired_at: d.expired_at ? d.expired_at.substring(0, 16) : '',
			};
		} catch (_) {
			formError = 'Gagal memuat data';
			showForm = false;
		}
	}

	function cancelForm() {
		showForm = false;
		formError = '';
	}

	async function handleSave() {
		if (!form.title.trim()) {
			formError = 'Judul pengumuman harus diisi';
			return;
		}
		if (!form.content.trim()) {
			formError = 'Konten pengumuman harus diisi';
			return;
		}

		isSaving = true;
		formError = '';
		try {
			const payload: Record<string, unknown> = {
				title: form.title.trim(),
				content: form.content.trim(),
				announcement_type: form.announcement_type,
				target_all: form.target_all,
				is_pinned: form.is_pinned,
				expired_at: form.expired_at || '',
			};
			if (isEditing && editId) {
				await api.update(editId, payload);
			} else {
				await api.create(payload);
			}
			cancelForm();
			load();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menyimpan data';
		} finally {
			isSaving = false;
		}
	}

	async function openDetail(id: string) {
		showDetail = true;
		showForm = false;
		detailId = id;
		isDetailLoading = true;
		detailData = null;
		try {
			const response = await api.get(id) as ApiResponse<Announcement>;
			detailData = response.data ?? null;
			await api.markRead(id);
		} catch (_) {
			detailData = null;
		} finally {
			isDetailLoading = false;
		}
	}

	function closeDetail() {
		showDetail = false;
		detailId = null;
		detailData = null;
	}

	async function handleDeleteDirect(id: string) {
		isSaving = true;
		try {
			await api.remove(id);
			if (detailId === id) {
				closeDetail();
			}
			load();
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal menghapus';
		} finally {
			isSaving = false;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function getInitials(name: string): string {
		if (!name) return '?';
		const parts = name.trim().split(' ');
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100 tracking-tight">Papan Pengumuman</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Pantau berita dan pengumuman resmi perusahaan</p>
		</div>
		{#if hasPermission('announcement', 'create')}
			<button onclick={openCreateForm} class="px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2 shadow-sm shadow-blue-200 dark:shadow-none">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Buat Pengumuman
			</button>
		{/if}
	</div>

	<div class="flex flex-col lg:flex-row gap-6 items-start">
		{#if !showForm && !showDetail}
		<!-- Left Side: Announcement Feed Grid -->
		<div class="flex-1 min-w-0 w-full">
			<!-- Header Filter Tabs -->
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-5 flex flex-col sm:flex-row sm:items-center justify-between gap-3 shadow-sm">
				<div class="flex flex-wrap items-center gap-2">
					<button onclick={() => { typeFilter = ''; page = 1; load(); }} class="px-3.5 py-1.5 text-xs font-semibold rounded-lg border transition cursor-pointer {!typeFilter ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'}">Semua</button>
					{#each Object.entries(announcementTypes) as [key, label]}
						<button onclick={() => { typeFilter = key; page = 1; load(); }} class="px-3.5 py-1.5 text-xs font-semibold rounded-lg border transition cursor-pointer {typeFilter === key ? 'bg-[#1A56DB] text-white border-[#1A56DB]' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'}">{label}</button>
					{/each}
				</div>
				<div class="text-xs font-medium text-gray-400 dark:text-gray-500">{total > 0 ? `${total} pengumuman` : ''}</div>
			</div>

			<!-- Error Alert -->
			{#if errorMessage}
				<div class="bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-xl px-4 py-3 mb-5 shadow-sm">{errorMessage}</div>
			{/if}

			<!-- Announcement Cards -->
			<PullToRefresh onRefresh={load}>
			{#if isLoading}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each [1, 2, 3, 4] as _}
						<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl p-5 space-y-4 animate-pulse">
							<div class="flex justify-between items-center"><div class="h-5 bg-gray-100 dark:bg-gray-850 rounded w-16"></div><div class="w-4 h-4 bg-gray-100 dark:bg-gray-850 rounded-full"></div></div>
							<div class="h-6 bg-gray-100 dark:bg-gray-850 rounded w-3/4"></div>
							<div class="h-4 bg-gray-50 dark:bg-gray-850/50 rounded w-1/2"></div>
							<div class="space-y-1.5"><div class="h-3 bg-gray-50 dark:bg-gray-850/50 rounded w-full"></div><div class="h-3 bg-gray-50 dark:bg-gray-850/50 rounded w-5/6"></div></div>
						</div>
					{/each}
				</div>
			{:else if items.length === 0}
				<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl py-16 text-center shadow-sm">
					<div class="w-14 h-14 mx-auto mb-4 rounded-2xl bg-gray-50 dark:bg-gray-800/40 flex items-center justify-center text-gray-400 dark:text-gray-500">
						<svg class="w-8 h-8" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.34 15.84c-.688-.06-1.386-.09-2.09-.09H7.5a4.5 4.5 0 1 1 0-9h.75c.704 0 1.402-.03 2.09-.09m0 9.18c.253.962.584 1.892.985 2.783.247.55.06 1.21-.463 1.511l-.657.38a.502.502 0 0 1-.673-.223 11.645 11.645 0 0 1-1.314-4.461m2.122-2.29c-.046 1.09.176 2.136.614 3.098m0 0a10.596 10.596 0 0 0 4.06-4.06m-4.06 4.06c.654.255 1.34.427 2.047.522m-3.503-6.664a10.545 10.545 0 0 0-3.068-2.282m0 0a10.54 10.54 0 0 0-3.36-1.004m3.361 1.004a10.55 10.55 0 0 1 2.264-.704" /></svg>
					</div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">Belum ada pengumuman</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">Belum ada pengumuman yang dibagikan.</p>
				</div>
			{:else}
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					{#each items as item}
						{@const isCurrentDetail = detailId === item.id && showDetail}
						{@const borderStyle = item.is_pinned
							? 'border-amber-400 dark:border-amber-500/60 bg-amber-50/10 dark:bg-amber-950/5 ring-1 ring-amber-400/30'
							: (isCurrentDetail ? 'border-[#1A56DB] dark:border-[#3B82F6]' : 'border-gray-200 dark:border-gray-800')}
						{@const ringStyle = item.announcement_type === 'emergency' ? 'ring-red-100 dark:ring-red-900/30' : item.announcement_type === 'important' ? 'ring-amber-100 dark:ring-amber-900/30' : 'ring-blue-100 dark:ring-blue-900/30'}
						
						<div class="group bg-white dark:bg-gray-900 border {borderStyle} rounded-2xl p-5 shadow-sm hover:shadow-md hover:border-gray-300 dark:hover:border-gray-700 transition duration-200 flex flex-col justify-between min-h-[190px] relative overflow-hidden">
							{#if item.is_pinned}
								<div class="absolute top-0 right-0 bg-amber-400 dark:bg-amber-500 text-white text-[9px] font-bold px-2 py-0.5 rounded-bl-lg tracking-wider flex items-center gap-0.5">
									<svg class="w-2.5 h-2.5" fill="currentColor" viewBox="0 0 20 20"><path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" /></svg>
									SEMATAN
								</div>
							{/if}
							
							<div>
								<div class="flex items-center gap-2 mb-2.5">
									<span class="inline-flex items-center px-2 py-0.5 rounded-md text-[10px] font-bold ring-1 {ringStyle} {typeColors[item.announcement_type]} uppercase">
										{announcementTypes[item.announcement_type] || item.announcement_type}
									</span>
								</div>
								<h3 class="text-base font-bold text-gray-900 dark:text-gray-100 line-clamp-1 leading-tight group-hover:text-[#1A56DB] dark:group-hover:text-[#3B82F6] transition">{item.title}</h3>
								<p class="text-xs text-gray-400 dark:text-gray-500 mt-1 flex items-center gap-1.5">
									<span>Oleh {item.created_by_name}</span>
									<span>•</span>
									<span>{formatDate(item.published_at)}</span>
								</p>
								<p class="text-xs text-gray-500 dark:text-gray-400 mt-2 line-clamp-2 leading-relaxed">
									{item.content}
								</p>
							</div>

							<!-- Card Footer Actions -->
							<div class="mt-4 pt-3 border-t border-gray-100 dark:border-gray-800 flex items-center justify-between">
								<button onclick={() => openDetail(item.id)} class="text-xs text-[#1A56DB] dark:text-[#3B82F6] font-semibold hover:underline cursor-pointer flex items-center gap-1">
									Baca Selengkapnya
									<svg class="w-3 h-3 transition group-hover:translate-x-0.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" /></svg>
								</button>

								<!-- Admin Controls on Card Hover/Right side -->
								<div class="flex items-center gap-1 opacity-50 group-hover:opacity-100 transition duration-150">
									{#if hasPermission('announcement', 'update')}
										<button onclick={() => openEdit(item.id)} aria-label="Edit" class="p-1 hover:bg-gray-100 dark:hover:bg-gray-850 rounded-lg text-gray-400 hover:text-blue-600 transition cursor-pointer">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
										</button>
									{/if}
									{#if hasPermission('announcement', 'delete')}
										{#if confirmDeleteId === item.id}
											<div class="absolute inset-0 bg-white/98 dark:bg-gray-900/98 flex flex-col items-center justify-center p-4 text-center z-10">
												<p class="text-xs font-semibold text-gray-700 dark:text-gray-300">Hapus pengumuman ini?</p>
												<div class="flex gap-2 mt-2">
													<button onclick={() => { confirmDeleteId = null; handleDeleteDirect(item.id); }} class="px-3.5 py-1 bg-red-600 text-white rounded text-xs font-semibold hover:bg-red-700 cursor-pointer">Hapus</button>
													<button onclick={() => confirmDeleteId = null} class="px-3 py-1 border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 rounded text-xs font-semibold hover:bg-gray-100 dark:hover:bg-gray-800 cursor-pointer">Batal</button>
												</div>
											</div>
										{:else}
											<button onclick={() => confirmDeleteId = item.id} aria-label="Hapus" class="p-1 hover:bg-gray-100 dark:hover:bg-gray-850 rounded-lg text-gray-400 hover:text-red-600 transition cursor-pointer">
												<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
											</button>
										{/if}
									{/if}
								</div>
							</div>
						</div>
					{/each}
				</div>

				<!-- Pagination Row -->
				<div class="flex flex-col sm:flex-row items-center justify-center sm:justify-between px-5 py-4 border border-gray-200 dark:border-gray-800 rounded-xl bg-white dark:bg-gray-900 mt-5 shadow-sm gap-4 sm:gap-0">
					<div class="text-xs text-gray-400 dark:text-gray-500 text-center sm:text-left">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span></div>
					<div class="flex flex-wrap items-center justify-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3.5 py-1.5 text-xs font-semibold rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-40 transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-semibold rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3.5 py-1.5 text-xs font-semibold rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-800 disabled:opacity-40 transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</PullToRefresh>
		</div>
		{/if}

		<!-- Right Side: Form or Full Details -->
		{#if showForm || showDetail}
			<div class="w-full bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-2xl shadow-sm overflow-hidden">
				{#if showForm}
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between bg-gray-50 dark:bg-gray-850">
						<h2 class="font-bold text-gray-900 dark:text-gray-100 text-base">{formTitle}</h2>
						<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
						</button>
					</div>
					<form onsubmit={(e) => { e.preventDefault(); handleSave(); }} class="p-6 space-y-4">
						{#if formError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-xs rounded-lg px-3 py-2">{formError}</div>{/if}
						<div>
							<label for="ann-title" class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1.5">Judul <span class="text-red-500">*</span></label>
							<input id="ann-title" type="text" bind:value={form.title} required class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" placeholder="Judul pengumuman" />
						</div>
						<div>
							<label for="ann-type" class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1.5">Tipe</label>
							<select id="ann-type" bind:value={form.announcement_type} class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
								{#each Object.entries(announcementTypes) as [key, label]}
									<option value={key}>{label}</option>
								{/each}
							</select>
						</div>
						<div>
							<label for="ann-content" class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1.5">Konten <span class="text-red-500">*</span></label>
							<textarea id="ann-content" bind:value={form.content} required rows={7} class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none resize-none focus:ring-2 focus:ring-[#1A56DB]/20" placeholder="Tulis konten pengumuman..."></textarea>
						</div>
						<div class="flex items-center gap-6">
							<label class="flex items-center gap-2 cursor-pointer">
								<input type="checkbox" bind:checked={form.is_pinned} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]/30" />
								<span class="text-sm text-gray-700 dark:text-gray-300">Sematkan (Pin di atas)</span>
							</label>
						</div>
						<div>
							<label for="ann-expired" class="block text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-1.5">Berakhir Pada (opsional)</label>
							<input id="ann-expired" type="datetime-local" bind:value={form.expired_at} class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" />
						</div>
						<button type="submit" disabled={isSaving} class="w-full py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition disabled:opacity-50 cursor-pointer">
							{isSaving ? 'Menyimpan...' : (isEditing ? 'Simpan Perubahan' : 'Publikasikan')}
						</button>
					</form>
				{:else if showDetail}
					{#if isDetailLoading}
						<div class="flex flex-col items-center justify-center py-20">
							<div class="w-8 h-8 border-4 border-gray-200 border-t-[#1A56DB] rounded-full animate-spin mb-4"></div>
							<div class="text-sm text-gray-500">Membuka pengumuman...</div>
						</div>
					{:else if detailData}
						<div class="relative bg-white dark:bg-gray-900 overflow-hidden">
							<!-- Decorative Top Accent -->
							<div class="absolute top-0 inset-x-0 h-2 bg-gradient-to-r {detailData.announcement_type === 'emergency' ? 'from-red-500 to-rose-500' : detailData.announcement_type === 'important' ? 'from-amber-400 to-orange-500' : 'from-[#1A56DB] to-blue-400'}"></div>
							
							<!-- Header Actions -->
							<div class="px-8 py-6 flex items-center justify-between border-b border-gray-100 dark:border-gray-800">
								<button onclick={closeDetail} class="flex items-center gap-2 text-sm font-semibold text-gray-500 hover:text-gray-900 dark:text-gray-400 dark:hover:text-gray-100 transition cursor-pointer group">
									<div class="w-8 h-8 rounded-full bg-gray-50 dark:bg-gray-800 flex items-center justify-center group-hover:bg-gray-100 dark:group-hover:bg-gray-700 transition">
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" /></svg>
									</div>
									Kembali
								</button>
								<div class="flex items-center gap-2">
									<span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-bold ring-1 {typeColors[detailData.announcement_type]} bg-white dark:bg-gray-900 shadow-sm uppercase tracking-widest">
										{announcementTypes[detailData.announcement_type] || detailData.announcement_type}
									</span>
									{#if detailData.is_pinned}
										<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold bg-gradient-to-r from-amber-400 to-amber-500 text-white shadow-sm uppercase tracking-widest">
											<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 20 20"><path d="M10.894 2.553a1 1 0 00-1.788 0l-7 14a1 1 0 001.169 1.409l5-1.429A1 1 0 009 15.571V11a1 1 0 112 0v4.571a1 1 0 00.725.962l5 1.428a1 1 0 001.17-1.408l-7-14z" /></svg>
											Disematkan
										</span>
									{/if}
								</div>
							</div>

							<!-- Content Body -->
							<div class="px-8 py-10 max-w-4xl mx-auto">
								<h1 class="text-3xl md:text-4xl font-extrabold text-gray-900 dark:text-white leading-tight mb-6">{detailData.title}</h1>
								
								<div class="flex flex-wrap items-center gap-x-6 gap-y-3 mb-10 pb-6 border-b border-gray-100 dark:border-gray-800 text-sm text-gray-500 dark:text-gray-400">
									<div class="flex items-center gap-2">
										<div class="w-8 h-8 rounded-full bg-gradient-to-br from-blue-100 to-indigo-100 text-blue-700 flex items-center justify-center font-bold text-xs ring-1 ring-blue-200">
											{getInitials(detailData.created_by_name || 'Admin')}
										</div>
										<span class="font-medium text-gray-900 dark:text-gray-200">{detailData.created_by_name || '-'}</span>
									</div>
									<div class="flex items-center gap-1.5">
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
										{formatDate(detailData.published_at)}
									</div>
									{#if detailData.expired_at}
										<div class="flex items-center gap-1.5 text-amber-600 dark:text-amber-400">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3Z" /></svg>
											Berakhir: {formatDate(detailData.expired_at)}
										</div>
									{/if}
									<div class="flex items-center gap-1.5 ml-auto">
										<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>
										Dibaca {detailData.read_count || 0} kali
									</div>
								</div>

								<div class="prose prose-blue max-w-none text-gray-800 dark:text-gray-200 whitespace-pre-wrap leading-loose text-[15px] md:text-base">
									{detailData.content}
								</div>
							</div>
						</div>
					{/if}
				{/if}
			</div>
		{/if}
	</div>
</div>
