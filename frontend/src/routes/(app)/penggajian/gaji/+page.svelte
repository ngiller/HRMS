<script lang="ts">
	import { onMount } from 'svelte';
	import { employees as employeesApi, salaryComponents as scApi, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';

	// Types
	type Employee = {
		id: string;
		employee_id: string;
		full_name: string;
		email: string;
		gender: string;
		employment_status: string;
		position_name: string;
		department_name: string;
		base_salary: number;
		daily_wage: number;
		is_active: boolean;
	};

	type SalaryComp = {
		id: string;
		employee_id: string;
		component_name: string;
		component_type: 'allowance' | 'deduction';
		amount: number;
		is_active: boolean;
		effective_date: string;
	};

	type SalaryCompForm = {
		component_name: string;
		component_type: 'allowance' | 'deduction';
		amount: number;
		effective_date: string;
	};

	let employees = $state<Employee[]>([]);
	let searchQuery = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Pagination state
	let page = $state(1);
	let perPage = $state(10);
	let total = $state(0);
	let totalPages = $state(0);

	// Editing basic salary
	let selectedEmployee = $state<Employee | null>(null);
	let showModal = $state(false);
	let isSavingSalary = $state(false);
	let salaryBase = $state('');
	let salaryDaily = $state('');

	// Salary Components state
	let salaryComps = $state<SalaryComp[]>([]);
	let isLoadingSC = $state(false);
	let scError = $state('');
	let scSuccess = $state('');
	
	// Component Form
	let showSCForm = $state(false);
	let scFormTitle = $state('');
	let editingSCId = $state<string | null>(null);
	let scForm = $state<SalaryCompForm>({ component_name: '', component_type: 'allowance', amount: 0, effective_date: '' });
	let scFormError = $state('');
	let isSavingSC = $state(false);

	// Delete Component confirmation
	let showSCDelete = $state(false);
	let deletingSCId = $state<string | null>(null);
	let deletingSCName = $state('');

	async function loadEmployees() {
		isLoading = true;
		errorMessage = '';
		try {
			const res: any = await employeesApi.list(page, perPage, searchQuery);
			if (res.success && res.data) {
				employees = res.data || [];
				total = res.meta?.total || 0;
				totalPages = Math.ceil(total / perPage);
			}
		} catch (err: any) {
			errorMessage = err.message || 'Gagal memuat data karyawan';
		} finally {
			isLoading = false;
		}
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadEmployees();
	}

	let searchTimeout: any;
	function handleSearch() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			page = 1;
			loadEmployees();
		}, 300);
	}

	async function openEdit(emp: Employee) {
		selectedEmployee = emp;
		salaryBase = String(emp.base_salary || '');
		salaryDaily = String(emp.daily_wage || '');
		showModal = true;
		scSuccess = '';
		scError = '';
		showSCForm = false;
		showSCDelete = false;
		await loadSalaryComponents();
	}

	async function loadSalaryComponents() {
		if (!selectedEmployee) return;
		isLoadingSC = true;
		scError = '';
		try {
			const res: any = await scApi.list(selectedEmployee.id, 1, 100);
			salaryComps = res.data || [];
		} catch (err: any) {
			scError = err.message || 'Gagal memuat komponen gaji';
		} finally {
			isLoadingSC = false;
		}
	}

	async function handleSaveSalary() {
		if (!selectedEmployee) return;
		isSavingSalary = true;
		scError = '';
		scSuccess = '';
		try {
			const payload: Record<string, any> = {};
			payload.base_salary = Number(salaryBase) || 0;
			payload.daily_wage = Number(salaryDaily) || 0;
			await employeesApi.update(selectedEmployee.id, payload);
			
			// Refresh local employee representation
			selectedEmployee.base_salary = payload.base_salary;
			selectedEmployee.daily_wage = payload.daily_wage;
			
			// Refresh list
			await loadEmployees();
			scSuccess = 'Gaji pokok/harian berhasil diperbarui';
			setTimeout(() => { scSuccess = ''; }, 3000);
		} catch (err: any) {
			scError = err.message || 'Gagal menyimpan gaji';
		} finally {
			isSavingSalary = false;
		}
	}

	function openCreateSCForm() {
		scFormTitle = 'Tambah Komponen Gaji';
		editingSCId = null;
		scForm = { component_name: '', component_type: 'allowance', amount: 0, effective_date: new Date().toISOString().split('T')[0] };
		scFormError = '';
		showSCForm = true;
	}

	function openEditSCForm(comp: SalaryComp) {
		scFormTitle = 'Edit Komponen Gaji';
		editingSCId = comp.id;
		scForm = {
			component_name: comp.component_name,
			component_type: comp.component_type,
			amount: comp.amount,
			effective_date: comp.effective_date?.split('T')[0] || '',
		};
		scFormError = '';
		showSCForm = true;
	}

	async function handleSaveSC() {
		if (!selectedEmployee) return;
		if (!scForm.component_name.trim()) { scFormError = 'Nama komponen harus diisi'; return; }
		if (scForm.amount < 0) { scFormError = 'Jumlah tidak boleh negatif'; return; }

		isSavingSC = true;
		scFormError = '';
		try {
			const payload = {
				component_name: scForm.component_name.trim(),
				component_type: scForm.component_type,
				amount: scForm.amount,
				effective_date: scForm.effective_date,
			};

			if (editingSCId) {
				await scApi.update(selectedEmployee.id, editingSCId, payload);
			} else {
				await scApi.create(selectedEmployee.id, payload);
			}
			showSCForm = false;
			await loadSalaryComponents();
		} catch (err: any) {
			scFormError = err.message || 'Gagal menyimpan komponen';
		} finally {
			isSavingSC = false;
		}
	}

	function confirmSCDelete(id: string, name: string) {
		deletingSCId = id;
		deletingSCName = name;
		showSCDelete = true;
	}

	async function handleSCDelete() {
		if (!selectedEmployee || !deletingSCId) return;
		isSavingSC = true;
		try {
			await scApi.remove(selectedEmployee.id, deletingSCId);
			showSCDelete = false;
			deletingSCId = null;
			await loadSalaryComponents();
		} catch (err: any) {
			scFormError = err.message || 'Gagal menghapus komponen';
			showSCDelete = false;
		} finally {
			isSavingSC = false;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	onMount(() => {
		loadEmployees();
	});

	let filteredEmployees = $derived(employees);
</script>

<div class="w-full min-h-screen bg-gray-50/50 p-6">
	<!-- Header Section -->
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight flex items-center gap-2">
				<span>💼 Pengaturan Gaji Karyawan</span>
			</h1>
			<p class="text-sm text-gray-500 mt-1">Kelola gaji pokok, upah harian, tunjangan, dan potongan tetap karyawan secara detail</p>
		</div>
	</div>

	<!-- Main Card Table -->
	<div class="bg-white border border-gray-200 rounded-2xl shadow-sm overflow-hidden">
		<!-- Search Bar -->
		<div class="p-5 border-b border-gray-100 bg-gray-50/20">
			<div class="relative max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
				</svg>
				<input
					type="text"
					bind:value={searchQuery}
					oninput={handleSearch}
					placeholder="Cari karyawan berdasarkan nama, email, NIK..."
					class="w-full pl-9 pr-4 py-2 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition"
				/>
			</div>
		</div>

		<!-- Table -->
		{#if isLoading}
			<div class="py-12"><PulseLoader variant="text" count={3} /></div>
		{:else if errorMessage}
			<div class="p-8 text-center text-red-500 text-sm">{errorMessage}</div>
		{:else if filteredEmployees.length === 0}
			<div class="p-12 text-center text-gray-400 text-sm">Tidak ditemukan data karyawan</div>
		{:else}
			<div class="overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50">
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Karyawan</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Gaji Pokok</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Upah Harian</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Jabatan</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Departemen</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider text-right">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-50">
						{#each filteredEmployees as emp (emp.id)}
							<tr class="hover:bg-gray-50/40 transition">
								<td class="px-6 py-4">
									<div class="flex items-center gap-3">
										<div class="w-9 h-9 rounded-full flex items-center justify-center text-white font-semibold text-sm {getAvatarTheme(emp.full_name)}">
											{getInitials(emp.full_name)}
										</div>
										<div>
											<span class="text-sm font-semibold text-gray-900 block">{emp.full_name}</span>
											<span class="text-xs text-gray-400 block mt-0.5">{emp.email}</span>
										</div>
									</div>
								</td>
								<td class="px-6 py-4 text-sm font-semibold text-gray-900">{formatCurrency(emp.base_salary)}</td>
								<td class="px-6 py-4 text-sm font-semibold text-gray-900">{formatCurrency(emp.daily_wage)}</td>
								<td class="px-6 py-4 text-sm text-gray-500">{emp.position_name || '-'}</td>
								<td class="px-6 py-4 text-sm text-gray-500">{emp.department_name || '-'}</td>
								<td class="px-6 py-4 text-right">
									<button onclick={() => openEdit(emp)}
										class="px-4 py-2 bg-indigo-50 text-indigo-600 border border-indigo-100 hover:bg-indigo-100 rounded-xl text-xs font-bold transition cursor-pointer">
										Atur Gaji
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<!-- Pagination -->
			<div class="flex items-center justify-between px-6 py-4 border-t border-gray-100 bg-gray-50/50">
				<div class="text-xs text-gray-555">
					Menampilkan {total === 0 ? 0 : (page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-bold text-gray-750">{total}</span> karyawan
				</div>
				<div class="flex items-center gap-1.5">
					<button onclick={() => goToPage(page - 1)} disabled={page <= 1}
						class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
					{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
						{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
						{#if pageNum <= totalPages}
							<button onclick={() => goToPage(pageNum)}
								class="w-8 h-8 text-xs font-bold rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-indigo-600 text-white border-indigo-600 shadow-sm shadow-indigo-100' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pageNum}</button>
						{/if}
					{/each}
					<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages}
						class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
				</div>
			</div>
		{/if}
	</div>
</div>

<!-- Slide Over Modal for Configuration -->
{#if showModal && selectedEmployee}
	<!-- Backface overlay -->
	<div class="fixed inset-0 bg-black/40 backdrop-blur-xs z-40 transition-opacity" onclick={() => showModal = false}></div>

	<!-- Drawer panel -->
	<div class="fixed inset-y-0 right-0 max-w-2xl w-full bg-white z-50 shadow-2xl flex flex-col transition-transform duration-300">
		<!-- Drawer Header -->
		<div class="bg-gradient-to-r from-indigo-600 to-indigo-800 px-6 py-5 text-white flex items-center justify-between shadow-lg">
			<div>
				<h2 class="text-lg font-bold">Atur Gaji & Komponen</h2>
				<p class="text-xs text-indigo-200 mt-1">{selectedEmployee.full_name} — {selectedEmployee.position_name}</p>
			</div>
			<button onclick={() => showModal = false} class="text-white/80 hover:text-white p-1 hover:bg-white/10 rounded-lg transition cursor-pointer">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
			</button>
		</div>

		<!-- Drawer Content -->
		<div class="flex-1 overflow-y-auto p-6 space-y-6">
			{#if scSuccess}
				<div class="p-4 bg-emerald-50 text-emerald-700 text-sm font-semibold rounded-xl border border-emerald-100">{scSuccess}</div>
			{/if}
			{#if scError}
				<div class="p-4 bg-red-50 text-red-700 text-sm font-semibold rounded-xl border border-red-100">{scError}</div>
			{/if}

			<!-- Basic Salary Configuration -->
			<div class="border border-gray-200 rounded-xl p-5 bg-gray-50/30 space-y-4">
				<h3 class="text-sm font-bold text-gray-900">💵 Gaji Pokok & Upah Harian</h3>
				
				<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
					<div>
						<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1.5">Gaji Pokok Bulanan (Base Salary)</label>
						<div class="relative">
							<span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400 font-semibold">Rp</span>
							<input type="number" bind:value={salaryBase}
								class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 font-semibold text-right" />
						</div>
					</div>
					<div>
						<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1.5">Upah Harian (Daily Wage)</label>
						<div class="relative">
							<span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400 font-semibold">Rp</span>
							<input type="number" bind:value={salaryDaily}
								class="w-full pl-9 pr-3 py-2 border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 font-semibold text-right" />
						</div>
					</div>
				</div>

				<div class="flex justify-end pt-2">
					<button onclick={handleSaveSalary} disabled={isSavingSalary}
						class="px-4 py-2 bg-indigo-600 text-white rounded-xl text-xs font-bold hover:bg-indigo-700 transition cursor-pointer flex items-center gap-1.5">
						{#if isSavingSalary}
							<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						{/if}
						Simpan Gaji Pokok/Harian
					</button>
				</div>
			</div>

			<!-- Dynamic Salary Components -->
			<div class="border border-gray-200 rounded-xl p-5 bg-white space-y-4">
				<div class="flex items-center justify-between border-b border-gray-100 pb-3">
					<h3 class="text-sm font-bold text-gray-900">💼 Komponen Gaji Kustom</h3>
					<button onclick={openCreateSCForm}
						class="px-3 py-1.5 bg-indigo-50 hover:bg-indigo-100 text-indigo-600 rounded-lg text-xs font-bold transition cursor-pointer flex items-center gap-1">
						<span>+</span> Tambah Komponen
					</button>
				</div>

				<!-- SC Form Modal overlay inside panel -->
				{#if showSCForm}
					<div class="p-4 bg-gray-50 border border-gray-200 rounded-xl space-y-3">
						<h4 class="text-xs font-bold text-gray-700">{scFormTitle}</h4>
						{#if scFormError}
							<p class="text-xs text-red-500 font-semibold">{scFormError}</p>
						{/if}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-3">
							<div>
								<label class="block text-[10px] font-bold text-gray-400 uppercase mb-1">Nama Komponen</label>
								<input type="text" bind:value={scForm.component_name} placeholder="Contoh: Tunjangan Makan"
									class="w-full px-2.5 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 font-semibold" />
							</div>
							<div>
								<label class="block text-[10px] font-bold text-gray-400 uppercase mb-1">Tipe</label>
								<select bind:value={scForm.component_type}
									class="w-full px-2.5 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 font-semibold">
									<option value="allowance">Tunjangan (Penambah Gaji)</option>
									<option value="deduction">Potongan (Pengurang Gaji)</option>
								</select>
							</div>
							<div>
								<label class="block text-[10px] font-bold text-gray-400 uppercase mb-1">Jumlah</label>
								<div class="relative">
									<span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-xs text-gray-450 font-bold">Rp</span>
									<input type="number" bind:value={scForm.amount}
										class="w-full pl-8 pr-2.5 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-bold" />
								</div>
							</div>
							<div>
								<label class="block text-[10px] font-bold text-gray-400 uppercase mb-1">Tanggal Efektif</label>
								<input type="date" bind:value={scForm.effective_date}
									class="w-full px-2.5 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 font-semibold" />
							</div>
						</div>
						<div class="flex justify-end gap-2 pt-2">
							<button onclick={() => showSCForm = false} class="px-3 py-1.5 bg-white border border-gray-200 text-gray-700 rounded-lg text-xs font-semibold hover:bg-gray-50 transition cursor-pointer">Batal</button>
							<button onclick={handleSaveSC} disabled={isSavingSC}
								class="px-3 py-1.5 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 transition disabled:opacity-50 cursor-pointer">
								{isSavingSC ? 'Menyimpan...' : 'Simpan'}
							</button>
						</div>
					</div>
				{/if}

				<!-- SC Delete confirm -->
				{#if showSCDelete}
					<div class="p-4 bg-red-50 border border-red-200 rounded-xl space-y-2">
						<p class="text-xs text-red-800 font-semibold">Apakah Anda yakin ingin menghapus komponen <strong>{deletingSCName}</strong>?</p>
						<div class="flex justify-end gap-2">
							<button onclick={() => showSCDelete = false} class="px-3 py-1 bg-white border border-gray-250 text-gray-700 rounded-lg text-[11px] font-semibold hover:bg-gray-50 transition cursor-pointer">Batal</button>
							<button onclick={handleSCDelete} class="px-3 py-1 bg-red-650 text-white rounded-lg text-[11px] font-semibold hover:bg-red-700 transition cursor-pointer">Ya, Hapus</button>
						</div>
					</div>
				{/if}

				<!-- Component List -->
				{#if isLoadingSC}
					<PulseLoader variant="text" count={2} />
				{:else if salaryComps.length === 0}
					<p class="text-xs text-gray-400 text-center py-6">Belum ada komponen tunjangan atau potongan tetap.</p>
				{:else}
					<div class="space-y-2">
						{#each salaryComps as comp (comp.id)}
							<div class="flex items-center justify-between p-3 border border-gray-100 rounded-xl bg-gray-50/20 hover:bg-gray-50/50 transition">
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<span class="text-xs font-semibold text-gray-900">{comp.component_name}</span>
										<span class="inline-flex items-center px-1.5 py-0.5 rounded-md text-[9px] font-semibold {comp.component_type === 'allowance' ? 'bg-emerald-50 text-emerald-700 ring-1 ring-inset ring-emerald-600/10' : 'bg-red-50 text-red-700 ring-1 ring-inset ring-red-600/10'}">
											{comp.component_type === 'allowance' ? 'Tunjangan' : 'Potongan'}
										</span>
									</div>
									<p class="text-[10px] text-gray-400 mt-1">Berlaku mulai: {new Date(comp.effective_date).toLocaleDateString('id-ID')}</p>
								</div>
								<div class="flex items-center gap-3">
									<span class="text-sm font-semibold {comp.component_type === 'allowance' ? 'text-emerald-600' : 'text-red-600'}">{formatCurrency(comp.amount)}</span>
									<div class="flex items-center gap-1">
										<button onclick={() => openEditSCForm(comp)} class="p-1 text-gray-400 hover:text-indigo-600 transition cursor-pointer">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" /></svg>
										</button>
										<button onclick={() => confirmSCDelete(comp.id, comp.component_name)} class="p-1 text-gray-400 hover:text-red-600 transition cursor-pointer">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" /></svg>
										</button>
									</div>
								</div>
							</div>
						{/each}
					</div>
				{/if}
			</div>
		</div>

		<!-- Drawer Footer -->
		<div class="bg-gray-50 px-6 py-4 border-t border-gray-150 flex items-center justify-end gap-3 shrink-0">
			<button onclick={() => showModal = false} class="px-5 py-2.5 bg-white border border-gray-200 text-gray-700 hover:bg-gray-50 rounded-xl text-sm font-semibold transition cursor-pointer">
				Tutup
			</button>
		</div>
	</div>
{/if}
