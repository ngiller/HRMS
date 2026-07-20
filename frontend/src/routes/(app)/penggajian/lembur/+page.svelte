<script lang="ts">
	import { onMount } from 'svelte';
	import { employees as employeesApi, company as companyApi, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import MobileCard from '$lib/components/MobileCard.svelte';

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
		is_active: boolean;
	};

	type EmployeeOvertimeConfig = {
		override_type: 'hourly_rate' | 'divisor' | 'percentage' | 'none';
		hourly_rate?: number;
		divisor?: number;
		rate_percentage?: number;
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

	// Selected Employee for Editing
	let selectedEmployee = $state<Employee | null>(null);
	let showModal = $state(false);
	let isLoadingConfig = $state(false);
	let isSavingConfig = $state(false);
	let configError = $state('');
	let configSuccess = $state('');

	let employeeOvertimeConfig = $state<EmployeeOvertimeConfig>({ override_type: 'none' });

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
		showModal = true;
		isLoadingConfig = true;
		configError = '';
		configSuccess = '';
		try {
			const res: any = await companyApi.getEmployeeOvertimeConfig(emp.id);
			const config = res?.data?.overtime_config || { override_type: 'none' };
			employeeOvertimeConfig = {
				override_type: config.override_type || 'none',
				hourly_rate: config.hourly_rate,
				divisor: config.divisor,
				rate_percentage: config.rate_percentage !== undefined ? config.rate_percentage * 100 : undefined,
			};
		} catch {
			employeeOvertimeConfig = { override_type: 'none' };
		} finally {
			isLoadingConfig = false;
		}
	}

	async function handleSave() {
		if (!selectedEmployee) return;
		isSavingConfig = true;
		configError = '';
		configSuccess = '';
		try {
			const payload: any = {
				override_type: employeeOvertimeConfig.override_type,
				hourly_rate: employeeOvertimeConfig.override_type === 'hourly_rate' && employeeOvertimeConfig.hourly_rate !== undefined ? employeeOvertimeConfig.hourly_rate : null,
				divisor: employeeOvertimeConfig.override_type === 'divisor' && employeeOvertimeConfig.divisor !== undefined ? employeeOvertimeConfig.divisor : null,
				rate_percentage: employeeOvertimeConfig.override_type === 'percentage' && employeeOvertimeConfig.rate_percentage !== undefined ? employeeOvertimeConfig.rate_percentage / 100 : null,
			};
			await companyApi.updateEmployeeOvertimeConfig(selectedEmployee.id, payload);
			configSuccess = 'Konfigurasi lembur berhasil disimpan';
			setTimeout(() => {
				showModal = false;
				selectedEmployee = null;
			}, 1000);
		} catch (err: any) {
			configError = err.message || 'Gagal menyimpan konfigurasi';
		} finally {
			isSavingConfig = false;
		}
	}

	onMount(() => {
		loadEmployees();
	});

	// Filtered employees list
	let filteredEmployees = $derived(employees);
</script>

<div class="w-full min-h-screen bg-gray-50/50 p-6">
	<!-- Header Section -->
	<div class="flex flex-col md:flex-row md:items-center justify-between gap-4 mb-8">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight flex items-center gap-2">
				<span>⏰ Pengaturan Upah Lembur Karyawan</span>
			</h1>
			<p class="text-sm text-gray-500 mt-1">Kelola basic rate upah lembur per jam karyawan, divisor kustom, nominal kustom, atau kustom persentase gaji pokok</p>
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
			<div class="hidden md:block overflow-x-auto">
				<table class="w-full text-left border-collapse">
					<thead>
						<tr class="border-b border-gray-100 bg-gray-50/50">
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">Karyawan</th>
							<th class="px-6 py-4 text-xs font-bold text-gray-400 uppercase tracking-wider">ID Karyawan</th>
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
								<td class="px-6 py-4 text-sm text-gray-600 font-medium">{emp.employee_id || '-'}</td>
								<td class="px-6 py-4 text-sm text-gray-500">{emp.position_name || '-'}</td>
								<td class="px-6 py-4 text-sm text-gray-500">{emp.department_name || '-'}</td>
								<td class="px-6 py-4 text-right">
									<button onclick={() => openEdit(emp)}
										class="px-4 py-2 bg-indigo-50 text-indigo-600 border border-indigo-100 hover:bg-indigo-100 rounded-xl text-xs font-bold transition cursor-pointer">
										Atur Lembur
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>

			<div class="md:hidden space-y-3 px-4 py-3">
				{#each filteredEmployees as emp (emp.id)}
					<MobileCard
						title={emp.full_name}
						subtitle={`${emp.position_name || '-'} · ${emp.department_name || '-'}`}
						avatar={getInitials(emp.full_name)}
						avatarColor={getAvatarTheme(emp.full_name).gradientClasses}
					>
						{#snippet footer()}
							<div class="flex items-center justify-between">
								<span class="text-xs text-gray-400">ID: {emp.employee_id || '-'}</span>
								<button onclick={() => openEdit(emp)}
									class="px-4 py-2 bg-indigo-50 text-indigo-600 border border-indigo-100 hover:bg-indigo-100 rounded-xl text-xs font-bold transition cursor-pointer active:scale-95">
									Atur Lembur
								</button>
							</div>
						{/snippet}
					</MobileCard>
				{/each}
			</div>

			<!-- Pagination -->
			<div class="flex items-center justify-between px-6 py-4 border-t border-gray-100 bg-gray-50/50">
				<div class="text-xs text-gray-550">
					Menampilkan {total === 0 ? 0 : (page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-bold text-gray-700">{total}</span> karyawan
				</div>
				<div class="flex items-center gap-1.5">
					<button onclick={() => goToPage(page - 1)} disabled={page <= 1}
						class="px-3 py-1.5 text-xs font-semibold rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
					{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
						{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
						{#if pageNum <= totalPages}
							<button onclick={() => goToPage(pageNum)}
								class="w-8 h-8 text-xs font-bold rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-indigo-650 text-white border-indigo-650 shadow-sm shadow-indigo-105' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pageNum}</button>
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
				<h2 class="text-lg font-bold">Atur Override Rate Lembur</h2>
				<p class="text-xs text-indigo-200 mt-1">{selectedEmployee.full_name} — {selectedEmployee.position_name}</p>
			</div>
			<button onclick={() => showModal = false} class="text-white/80 hover:text-white p-1 hover:bg-white/10 rounded-lg transition cursor-pointer">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" /></svg>
			</button>
		</div>

		<!-- Drawer Content -->
		<div class="flex-1 overflow-y-auto p-6 space-y-6">
			{#if configSuccess}
				<div class="p-4 bg-emerald-50 text-emerald-700 text-sm font-semibold rounded-xl border border-emerald-100">{configSuccess}</div>
			{/if}
			{#if configError}
				<div class="p-4 bg-red-50 text-red-700 text-sm font-semibold rounded-xl border border-red-100">{configError}</div>
			{/if}

			{#if isLoadingConfig}
				<div class="py-12"><PulseLoader variant="text" count={3} /></div>
			{:else}
				<div class="space-y-4">
					<div class="border border-gray-200 rounded-xl p-5 bg-gray-50/30 space-y-4">
						<div>
							<span class="text-sm font-bold text-gray-900">Perhitungan Rate per Jam</span>
							<p class="text-xs text-gray-400 mt-1">Secara default, upah lembur per jam dihitung dengan formula <strong>Gaji Pokok / 173</strong>. Anda dapat menetapkan nominal tetap per jam, mengubah pembagi (divisor), atau menetapkan kustom persentase gaji pokok.</p>
						</div>

						<div class="flex flex-col md:flex-row gap-4 md:items-center">
							<div class="w-full md:w-64">
								<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">Pilih Metode</label>
								<select bind:value={employeeOvertimeConfig.override_type}
									class="w-full px-3 py-2 bg-white border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 font-semibold cursor-pointer">
									<option value="none">Default (Gaji Pokok / 173)</option>
									<option value="hourly_rate">Kustom Nominal per Jam (Rp)</option>
									<option value="divisor">Kustom Pembagi Gaji Pokok (Divisor)</option>
									<option value="percentage">Kustom Persentase Gaji Pokok (%)</option>
								</select>
							</div>

							{#if employeeOvertimeConfig.override_type === 'hourly_rate'}
								<div class="w-full md:w-56">
									<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">Nominal per Jam</label>
									<div class="relative">
										<span class="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-gray-400 font-semibold">Rp</span>
										<input type="number" step="1000" min="0" bind:value={employeeOvertimeConfig.hourly_rate}
											class="w-full pl-9 pr-3 py-2 bg-white border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 text-right font-bold" />
									</div>
								</div>
							{:else if employeeOvertimeConfig.override_type === 'divisor'}
								<div class="w-full md:w-40">
									<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">Nilai Pembagi (Divisor)</label>
									<input type="number" step="1" min="1" bind:value={employeeOvertimeConfig.divisor}
										class="w-full px-3 py-2 bg-white border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 text-right font-bold" />
								</div>
							{:else if employeeOvertimeConfig.override_type === 'percentage'}
								<div class="w-full md:w-40">
									<label class="block text-xs font-bold text-gray-400 uppercase tracking-wider mb-1">Persentase (%)</label>
									<div class="relative">
										<input type="number" step="0.01" min="0" max="100" bind:value={employeeOvertimeConfig.rate_percentage}
											class="w-full pl-3 pr-8 py-2 bg-white border border-gray-200 rounded-xl text-sm outline-none focus:ring-1 focus:ring-indigo-500 text-right font-bold" />
										<span class="absolute right-3 top-1/2 -translate-y-1/2 text-sm text-gray-400 font-semibold">%</span>
									</div>
								</div>
							{/if}
						</div>
					</div>
				</div>
			{/if}
		</div>

		<!-- Drawer Footer -->
		<div class="bg-gray-50 px-6 py-4 border-t border-gray-150 flex items-center justify-end gap-3 shrink-0">
			<button onclick={() => showModal = false} class="px-5 py-2.5 bg-white border border-gray-200 text-gray-700 hover:bg-gray-50 rounded-xl text-sm font-semibold transition cursor-pointer">
				Batal
			</button>
			<button onclick={handleSave} disabled={isSavingConfig || isLoadingConfig}
				class="px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-semibold hover:bg-indigo-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer shadow-sm shadow-indigo-100">
				{#if isSavingConfig}
					<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
				{/if}
				Simpan Override
			</button>
		</div>
	</div>
{/if}
