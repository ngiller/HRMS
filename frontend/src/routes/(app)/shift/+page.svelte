<script lang="ts">
	import { onMount } from 'svelte';
	import { shifts as api, departments as deptApi } from '$lib/api.js';

	type ShiftItem = {
		id: string;
		department_id?: string;
		name: string;
		code: string;
		start_time: string;
		end_time: string;
		color: string;
		is_active: boolean;
		description: string;
		created_at: string;
	};

	let items = $state<ShiftItem[]>([]);
	let departments = $state<Array<{ id: string; name: string }>>([]);
	let loading = $state(true);
	let error = $state('');

	// Form state
	let showForm = $state(false);
	let editingId = $state<string | null>(null);
	let formDept = $state('');
	let formName = $state('');
	let formCode = $state('');
	let formStart = $state('06:00');
	let formEnd = $state('14:00');
	let formBreakStart = $state('12:00');
	let formBreakEnd = $state('13:00');
	let formColor = $state('#3B82F6');
	let formDesc = $state('');
	let formError = $state('');
	let saving = $state(false);

	// Delete confirm
	let showDelete = $state(false);
	let deleteId = $state<string | null>(null);
	let deleteName = $state('');

	onMount(async () => {
		load();
		try {
			const res: any = await deptApi.getAll();
			departments = res.data || [];
		} catch {}
	});

	async function load() {
		loading = true; error = '';
		try {
			const res: any = await api.list(1, 100);
			items = res.data || [];
		} catch (e: any) { error = e.message || 'Gagal memuat'; }
		finally { loading = false; }
	}

	function openCreate() {
		editingId = null;
		formDept = '';
		formName = ''; formCode = ''; formStart = '06:00'; formEnd = '14:00';
		formBreakStart = '12:00'; formBreakEnd = '13:00'; formColor = '#3B82F6'; formDesc = '';
		formError = ''; showForm = true;
	}

	async function openEdit(item: ShiftItem) {
		editingId = item.id;
		formDept = item.department_id || '';
		formName = item.name; formCode = item.code; formStart = item.start_time.slice(0,5);
		formEnd = item.end_time.slice(0,5);
		formColor = item.color; formDesc = item.description;
		try {
			const res: any = await api.get(item.id);
			if (res.data) {
				formBreakStart = res.data.break_start?.slice(0,5) || '12:00';
				formBreakEnd = res.data.break_end?.slice(0,5) || '13:00';
			}
		} catch {}
		formError = ''; showForm = true;
	}

	function cancelForm() { showForm = false; formError = ''; }

	async function save() {
		if (!formName.trim()) { formError = 'Nama shift harus diisi'; return; }
		if (!formCode.trim()) { formError = 'Kode shift harus diisi'; return; }
		saving = true; formError = '';
		try {
			const payload: Record<string, any> = {
				name: formName.trim(), code: formCode.trim(),
				start_time: formStart, end_time: formEnd,
				break_start: formBreakStart, break_end: formBreakEnd,
				color: formColor, description: formDesc.trim(),
				department_id: formDept || null,
			};
			if (editingId) { await api.update(editingId, payload); }
			else { await api.create(payload); }
			showForm = false; load();
		} catch (e: any) { formError = e.message || 'Gagal menyimpan'; }
		finally { saving = false; }
	}

	function confirmDelete(id: string, name: string) { deleteId = id; deleteName = name; showDelete = true; }
	function cancelDelete() { showDelete = false; deleteId = null; deleteName = ''; }

	async function handleDelete() {
		if (!deleteId) return;
		try { await api.remove(deleteId); showDelete = false; deleteId = null; load(); }
		catch (e: any) { error = e.message || 'Gagal menghapus'; showDelete = false; }
	}

	const SHIFT_COLORS: Record<string, string> = {
		morning: '#3B82F6', afternoon: '#F59E0B', night: '#8B5CF6', regular: '#10B981'
	};
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Shift Kerja</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola definisi shift kerja (Pagi, Siang, Malam, dll)</p>
		</div>
		{#if !showForm}
			<button onclick={openCreate} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Shift
			</button>
		{/if}
	</div>

	{#if showForm}
		<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl overflow-hidden shadow-sm mb-6">
			<div class="px-6 py-4 border-b border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-900/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{editingId ? 'Edit Shift' : 'Tambah Shift Baru'}</h2>
			</div>
			<div class="p-6 space-y-5">
				{#if formError}
					<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-200 text-sm px-4 py-2.5 rounded-xl">{formError}</div>
				{/if}
				<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Departemen
							<select bind:value={formDept} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white cursor-pointer">
								<option value="">Global (Semua Dept)</option>
								{#each departments as d (d.id)}
									<option value={d.id}>{d.name}</option>
								{/each}
							</select>
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Nama Shift <span class="text-red-500">*</span>
							<input type="text" bind:value={formName} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" placeholder="Pagi" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Kode Shift <span class="text-red-500">*</span>
							<input type="text" bind:value={formCode} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" placeholder="morning" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Jam Mulai <span class="text-red-500">*</span>
							<input type="time" bind:value={formStart} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Jam Selesai <span class="text-red-500">*</span>
							<input type="time" bind:value={formEnd} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Istirahat Mulai
							<input type="time" bind:value={formBreakStart} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Istirahat Selesai
							<input type="time" bind:value={formBreakEnd} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" />
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Warna
							<div class="flex items-center gap-2 mt-1.5">
								<input type="color" bind:value={formColor} class="w-10 h-10 rounded-lg border border-gray-200 dark:border-gray-700 cursor-pointer" />
								<span class="text-xs text-gray-500 dark:text-gray-400">{formColor}</span>
							</div>
						</label>
					</div>
					<div>
						<label class="block text-sm font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
							Deskripsi
							<input type="text" bind:value={formDesc} class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900 text-gray-900 dark:text-white" placeholder="Shift Pagi: 06:00 - 14:00" />
						</label>
					</div>
				</div>
			</div>
			<div class="flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-900/50">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={save} disabled={saving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if saving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Shift'}
				</button>
			</div>
		</div>
	{/if}

	<!-- Shift Cards -->
	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-2 border-[#1A56DB] border-t-transparent rounded-full"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-4"><p class="text-sm font-medium text-red-700 dark:text-red-200">{error}</p></div>
	{:else}
		<div class="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-4">
			{#each items as item (item)}
				<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl overflow-hidden shadow-sm hover:shadow-md transition-all duration-200">
					<div class="h-2" style="background: {item.color}"></div>
					<div class="p-5">
						<div class="flex items-start justify-between mb-3">
							<div class="flex items-center gap-3">
								<div class="w-10 h-10 rounded-xl flex items-center justify-center text-white font-bold text-sm shadow-sm" style="background: {item.color}">
									{item.name.charAt(0)}
								</div>
								<div>
									<h3 class="text-sm font-bold text-gray-900 dark:text-white">{item.name}</h3>
									<div class="flex flex-wrap items-center gap-1 mt-0.5">
										<span class="text-[9px] font-mono text-gray-400 uppercase bg-gray-100 dark:bg-gray-800 px-1 py-0.5 rounded">{item.code}</span>
										<span class="text-[9px] font-semibold px-1 py-0.5 rounded {item.department_id ? 'bg-indigo-50 text-indigo-700 dark:bg-indigo-950/30 dark:text-indigo-300' : 'bg-gray-50 text-gray-600 dark:bg-gray-850 dark:text-gray-400'}">
											{departments.find(d => d.id === item.department_id)?.name || 'Global'}
										</span>
									</div>
								</div>
							</div>
							<span class="inline-flex items-center px-2 py-0.5 rounded text-[10px] font-medium {item.is_active ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
								{item.is_active ? 'Aktif' : 'Nonaktif'}
							</span>
						</div>
						<div class="space-y-2 text-sm">
							<div class="flex items-center justify-between bg-gray-50 dark:bg-gray-900/50 rounded-xl px-3 py-2">
								<span class="text-gray-500 dark:text-gray-400 text-xs">Jam</span>
								<span class="font-semibold text-gray-900 dark:text-white tabular-nums">{item.start_time.slice(0,5)} - {item.end_time.slice(0,5)}</span>
							</div>
							{#if item.description}
								<p class="text-xs text-gray-500 dark:text-gray-400">{item.description}</p>
							{/if}
						</div>
						<div class="flex items-center justify-end gap-1 mt-3 pt-3 border-t border-gray-100 dark:border-gray-700">
							<button onclick={() => openEdit(item)} class="p-1.5 rounded-lg text-gray-400 hover:text-blue-600 hover:bg-blue-50 dark:hover:bg-blue-900/20 transition cursor-pointer">
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
							</button>
							<button onclick={() => confirmDelete(item.id, item.name)} class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 dark:hover:bg-red-900/20 transition cursor-pointer">
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
							</button>
						</div>
					</div>
				</div>
			{/each}
			{#if items.length === 0}
				<div class="col-span-full py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
						<svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Belum ada shift</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400">Buat shift kerja pertama untuk memulai.</p>
				</div>
			{/if}
		</div>
	{/if}
</div>

<!-- Delete Modal -->
{#if showDelete}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onclick={cancelDelete} role="presentation">
		<div onclick={(e) => e.stopPropagation()} role="dialog" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm p-6 text-center">
			<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
				<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
			</div>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Hapus Shift</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Apakah Anda yakin ingin menghapus</p>
			<p class="text-sm font-medium text-gray-900 dark:text-white mb-4">"{deleteName}"?</p>
			<div class="flex items-center justify-center gap-3">
				<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={handleDelete} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition cursor-pointer">Ya, Hapus</button>
			</div>
		</div>
	</div>
{/if}
