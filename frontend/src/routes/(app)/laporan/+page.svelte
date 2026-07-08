<script lang="ts">
	import { onMount } from 'svelte';
	import { reports, departments, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';

	type HeadcountReport = {
		total_employees: number;
		active_employees: number;
		new_hires_this_year: number;
		resigned_this_year: number;
		by_department: { department_name: string; count: number }[];
	};

	type PayrollReport = {
		total_periods: number;
		total_gross: number;
		total_net_salary: number;
		average_net_salary: number;
	};

	type AttendanceReport = {
		total_records: number;
		on_time: number;
		late: number;
		late_percentage: number;
	};

	type LeaveReport = {
		total_requests: number;
		approved: number;
		pending: number;
		total_days_approved: number;
	};

	type OvertimeReport = {
		total_requests: number;
		approved: number;
		total_hours: number;
		total_cost: number;
	};

	let activeTab = $state<'headcount' | 'payroll' | 'attendance' | 'leave' | 'overtime'>('headcount');

	// Filters
	let selectedYear = $state(2026);
	let selectedMonth = $state(0);
	let selectedDept = $state('');
	let deptOptions = $state<{ id: string; name: string }[]>([]);

	// Data
	let headcount = $state<HeadcountReport | null>(null);
	let payroll = $state<PayrollReport | null>(null);
	let attendance = $state<AttendanceReport | null>(null);
	let leave = $state<LeaveReport | null>(null);
	let overtime = $state<OvertimeReport | null>(null);

	let loading = $state(false);
	let error = $state('');

	onMount(() => { loadDepts(); loadCurrentTab(); });

	async function loadDepts() {
		try { const res = await departments.getAll(); if (res?.success) deptOptions = res.data; } catch {}
	}

	async function loadCurrentTab() {
		loading = true; error = '';
		try {
			switch (activeTab) {
				case 'headcount': {
					const res = await reports.headcount(selectedYear);
					if (res?.success) headcount = res.data;
					break;
				}
				case 'payroll': {
					const res = await reports.payrollSummary(selectedYear);
					if (res?.success) payroll = res.data;
					break;
				}
				case 'attendance': {
					const res = await reports.attendanceSummary(selectedYear, selectedMonth, selectedDept);
					if (res?.success) attendance = res.data;
					break;
				}
				case 'leave': {
					const res = await reports.leaveSummary(selectedYear, selectedDept);
					if (res?.success) leave = res.data;
					break;
				}
				case 'overtime': {
					const res = await reports.overtimeSummary(selectedYear, selectedMonth);
					if (res?.success) overtime = res.data;
					break;
				}
			}
		} catch (err) { error = err instanceof ApiError ? err.message : 'Gagal memuat laporan'; }
		finally { loading = false; }
	}

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val || 0);
	}

	function formatPercent(val: number) {
		return (val || 0).toFixed(1) + '%';
	}
</script>

<div class="w-full">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Laporan & Analytics</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Ringkasan data HR dalam satu tampilan</p>
	</div>

	<!-- Tabs -->
	<div class="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-800 overflow-x-auto">
		{#each ['headcount', 'payroll', 'attendance', 'leave', 'overtime'] as tab}				<button onclick={() => { activeTab = tab as typeof activeTab; loadCurrentTab(); }}
				class="px-5 py-3 text-sm font-medium whitespace-nowrap transition border-b-2 -mb-px cursor-pointer
					{activeTab === tab ? 'border-[#1A56DB] text-[#1A56DB]' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}">
				{tab === 'headcount' ? 'Headcount' : tab === 'payroll' ? 'Penggajian' : tab === 'attendance' ? 'Absensi' : tab === 'leave' ? 'Cuti' : 'Lembur'}
			</button>
		{/each}
	</div>

	{#if error}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{error}</div>{/if}

	<!-- Filters -->
	<div class="flex flex-wrap gap-3 mb-4">
		<label for="filter-year" class="text-sm text-gray-500 self-center">Tahun:</label>
		<select id="filter-year" bind:value={selectedYear} onchange={loadCurrentTab}
			class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
			{#each [2024,2025,2026,2027] as y}<option value={y}>{y}</option>{/each}
		</select>
		{#if activeTab === 'attendance' || activeTab === 'overtime'}
			<label for="filter-month" class="text-sm text-gray-500 self-center">Bulan:</label>
			<select id="filter-month" bind:value={selectedMonth} onchange={loadCurrentTab}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
				<option value={0}>Semua</option>
				{#each ['Jan','Feb','Mar','Apr','Mei','Jun','Jul','Agu','Sep','Okt','Nov','Des'] as m, i}
					<option value={i+1}>{m}</option>
				{/each}
			</select>
		{/if}
		{#if activeTab === 'attendance' || activeTab === 'leave'}
			<label for="filter-dept" class="text-sm text-gray-500 self-center">Departemen:</label>
			<select id="filter-dept" bind:value={selectedDept} onchange={loadCurrentTab}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
				<option value="">Semua</option>
				{#each deptOptions as dept}<option value={dept.id}>{dept.name}</option>{/each}
			</select>
		{/if}
	</div>

	{#if loading}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-8">
			<PulseLoader variant="card" count={2} />
		</div>
	{:else}
		<!-- ═══ HEADCOUNT ═══ -->
		{#if activeTab === 'headcount' && headcount}
			<div class="space-y-6">
				<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Total Karyawan</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">{headcount.total_employees}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Aktif</p>
						<p class="text-2xl font-bold text-green-600 mt-1">{headcount.active_employees}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Bergabung Tahun Ini</p>
						<p class="text-2xl font-bold text-blue-600 mt-1">{headcount.new_hires_this_year}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Resign Tahun Ini</p>
						<p class="text-2xl font-bold text-red-600 mt-1">{headcount.resigned_this_year}</p>
					</div>
				</div>
				<!-- By Department -->
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-5 shadow-sm">
					<h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-3">Berdasarkan Departemen</h3>
					<div class="space-y-2">
						{#each headcount.by_department || [] as dept}
							<div class="flex items-center justify-between py-1.5">
								<span class="text-sm text-gray-700 dark:text-gray-300">{dept.department_name}</span>
								<div class="flex items-center gap-3">
									<div class="w-32 bg-gray-100 dark:bg-gray-800 rounded-full h-2">
										<div class="bg-[#1A56DB] h-2 rounded-full" style="width: {(dept.count / Math.max(headcount.total_employees, 1)) * 100}%"></div>
									</div>
									<span class="text-sm font-semibold text-gray-900 dark:text-gray-100 w-8 text-right">{dept.count}</span>
								</div>
							</div>
						{/each}
					</div>
				</div>
			</div>

		<!-- ═══ PAYROLL ═══ -->
		{:else if activeTab === 'payroll' && payroll}
			<div class="space-y-6">
				<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Total Periode</p>
						<p class="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">{payroll.total_periods}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Total Gross</p>
						<p class="text-lg font-bold text-[#1A56DB] mt-1">{formatCurrency(payroll.total_gross)}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Total Net</p>
						<p class="text-lg font-bold text-green-600 mt-1">{formatCurrency(payroll.total_net_salary)}</p>
					</div>
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
						<p class="text-xs text-gray-400 uppercase tracking-wide">Rata-rata Net</p>
						<p class="text-lg font-bold text-gray-900 dark:text-gray-100 mt-1">{formatCurrency(payroll.average_net_salary)}</p>
					</div>
				</div>
			</div>

		<!-- ═══ ATTENDANCE ═══ -->
		{:else if activeTab === 'attendance' && attendance}
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Total Record</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">{attendance.total_records}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Tepat Waktu</p>
					<p class="text-2xl font-bold text-green-600 mt-1">{attendance.on_time}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Terlambat</p>
					<p class="text-2xl font-bold text-red-600 mt-1">{attendance.late}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">% Terlambat</p>
					<p class="text-2xl font-bold text-amber-600 mt-1">{formatPercent(attendance.late_percentage)}</p>
				</div>
			</div>

		<!-- ═══ LEAVE ═══ -->
		{:else if activeTab === 'leave' && leave}
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Total Pengajuan</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">{leave.total_requests}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Disetujui</p>
					<p class="text-2xl font-bold text-green-600 mt-1">{leave.approved}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Pending</p>
					<p class="text-2xl font-bold text-yellow-600 mt-1">{leave.pending}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Hari Cuti (Approved)</p>
					<p class="text-2xl font-bold text-[#1A56DB] mt-1">{leave.total_days_approved}</p>
				</div>
			</div>

		<!-- ═══ OVERTIME ═══ -->
		{:else if activeTab === 'overtime' && overtime}
			<div class="grid grid-cols-2 md:grid-cols-4 gap-4">
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Total Lembur</p>
					<p class="text-2xl font-bold text-gray-900 dark:text-gray-100 mt-1">{overtime.total_requests}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Disetujui</p>
					<p class="text-2xl font-bold text-green-600 mt-1">{overtime.approved}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Total Jam</p>
					<p class="text-2xl font-bold text-[#1A56DB] mt-1">{overtime.total_hours?.toFixed(1) || '0'}</p>
				</div>
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 shadow-sm hover:shadow-md transition-all duration-200">
					<p class="text-xs text-gray-400 uppercase tracking-wide">Biaya Lembur</p>
					<p class="text-lg font-bold text-amber-600 mt-1">{formatCurrency(overtime.total_cost)}</p>
				</div>
			</div>

		{:else}
			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-8 text-center text-sm text-gray-400">
				Pilih tab di atas untuk melihat laporan
			</div>
		{/if}
	{/if}
</div>
