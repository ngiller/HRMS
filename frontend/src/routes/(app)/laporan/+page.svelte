<script lang="ts">
	import { onMount } from 'svelte';
	import { reports, departments, employees, ApiError, attendance as attendanceApi } from '$lib/api.js';
	import config from '$lib/config.js';
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
	let selectedEmployeeId = $state('');
	let deptOptions = $state<{ id: string; name: string }[]>([]);
	let employeeOptions = $state<{ id: string; full_name: string }[]>([]);

	// Data
	let headcount = $state<HeadcountReport | null>(null);
	let payroll = $state<PayrollReport | null>(null);
	let attendance = $state<AttendanceReport | null>(null);
	let leave = $state<LeaveReport | null>(null);
	let overtime = $state<OvertimeReport | null>(null);

	
	// Detailed Attendance List State
	let attendanceRecords = $state<any[]>([]);
	let attendanceTotal = $state(0);
	let attendancePage = $state(1);
	const attendancePerPage = 10;
	let attendanceTotalPages = $derived(Math.max(1, Math.ceil(attendanceTotal / attendancePerPage)));
	let selectedRecord = $state<any | null>(null);
	let showDetail = $state(false);
	let exportLoading = $state(false);

	function formatTime(d: string | null | undefined) {
		if (!d) return '-';
		return new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			present: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			late: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			absent: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			leave: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
			holiday: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
		};
		return map[status] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300';
	}

	function getStatusText(status: string) {
		const map: Record<string, string> = {
			present: 'Hadir',
			late: 'Terlambat',
			absent: 'Absen',
			leave: 'Cuti',
			holiday: 'Libur',
		};
		return map[status] || status;
	}

	function getPhotoUrl(url: string | null | undefined): string {
		if (!url) return '';
		if (url.startsWith('http')) return url;
		return `${config.API_BASE_URL}${url}`;
	}

	async function exportAttendance() {
		exportLoading = true;
		try {
			let dateFrom = '';
			let dateTo = '';
			if (selectedMonth > 0) {
				const monthStr = String(selectedMonth).padStart(2, '0');
				dateFrom = `${selectedYear}-${monthStr}-01`;
				const lastDay = new Date(selectedYear, selectedMonth, 0).getDate();
				dateTo = `${selectedYear}-${monthStr}-${lastDay}`;
			} else {
				dateFrom = `${selectedYear}-01-01`;
				dateTo = `${selectedYear}-12-31`;
			}
			const blob = await attendanceApi.exportReport(selectedDept, selectedEmployeeId, '', dateFrom, dateTo);
			const url = window.URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = `Laporan_Absensi_${selectedYear}_${selectedMonth || 'Semua'}.xlsx`;
			a.click();
			window.URL.revokeObjectURL(url);
		} catch (err: any) {
			alert(err.message || 'Gagal export laporan');
		} finally {
			exportLoading = false;
		}
	}

	let loading = $state(false);
	let error = $state('');

	onMount(() => { loadDepts(); loadCurrentTab(); });

	async function loadDepts() {
		try {
			const [deptRes, empRes] = await Promise.all([
				departments.getAll(),
				employees.list(1, 1000),
			]);
			if (deptRes?.success) deptOptions = deptRes.data;
			if (empRes?.success) employeeOptions = empRes.data || [];
		} catch { /* noop */ }
	}

	
	function changeAttendancePage(page: number) {
		attendancePage = page;
		loadCurrentTab();
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

					// Build date filters for the detailed list
					let dateFrom = '';
					let dateTo = '';
					if (selectedMonth > 0) {
						const monthStr = String(selectedMonth).padStart(2, '0');
						dateFrom = `${selectedYear}-${monthStr}-01`;
						const lastDay = new Date(selectedYear, selectedMonth, 0).getDate();
						dateTo = `${selectedYear}-${monthStr}-${lastDay}`;
					} else {
						dateFrom = `${selectedYear}-01-01`;
						dateTo = `${selectedYear}-12-31`;
					}

					const listRes: any = await attendanceApi.report(attendancePage, attendancePerPage, selectedDept, selectedEmployeeId, '', dateFrom, dateTo);
					if (listRes?.success) {
						attendanceRecords = listRes.data || [];
						attendanceTotal = listRes.meta?.total || 0;
					}
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
		{#each ['headcount', 'payroll', 'attendance', 'leave', 'overtime'] as tab (tab)}				<button onclick={() => { activeTab = tab as typeof activeTab; loadCurrentTab(); }}
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
			{#each [2024,2025,2026,2027] as y (y)}<option value={y}>{y}</option>{/each}
		</select>
		{#if activeTab === 'attendance' || activeTab === 'overtime'}
			<label for="filter-month" class="text-sm text-gray-500 self-center">Bulan:</label>
			<select id="filter-month" bind:value={selectedMonth} onchange={loadCurrentTab}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
				<option value={0}>Semua</option>
				{#each ['Jan','Feb','Mar','Apr','Mei','Jun','Jul','Agu','Sep','Okt','Nov','Des'] as m, i (i)}
					<option value={i+1}>{m}</option>
				{/each}
			</select>
		{/if}
		{#if activeTab === 'attendance' || activeTab === 'leave'}
			<label for="filter-dept" class="text-sm text-gray-500 self-center">Departemen:</label>
			<select id="filter-dept" bind:value={selectedDept} onchange={loadCurrentTab}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
				<option value="">Semua</option>
				{#each deptOptions as dept (dept)}<option value={dept.id}>{dept.name}</option>{/each}
			</select>
		{/if}
		{#if activeTab === 'attendance'}
			<label for="filter-employee" class="text-sm text-gray-500 self-center">Karyawan:</label>
			<select id="filter-employee" bind:value={selectedEmployeeId} onchange={loadCurrentTab}
				class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none max-w-[220px]">
				<option value="">Semua Karyawan</option>
				{#each employeeOptions as emp (emp)}
					<option value={emp.id}>{emp.full_name}</option>
				{/each}
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
						{#each headcount.by_department || [] as dept (dept)}
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
			<div class="space-y-6">
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

				<!-- Detailed Table -->
				<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden overflow-x-auto">
					<div class="px-5 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between flex-wrap gap-4">
						<div>
							<h3 class="text-base font-bold text-gray-900 dark:text-gray-100">Daftar Kehadiran Karyawan</h3>
							<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Menampilkan detail check-in dan check-out seluruh staf</p>
						</div>
						
						<button onclick={exportAttendance} disabled={exportLoading || attendanceRecords.length === 0}
							class="px-4 py-2 bg-emerald-600 text-white rounded-lg text-sm font-semibold hover:bg-emerald-700 disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer flex items-center gap-1.5 shadow-sm">
							{#if exportLoading}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
								Exporting...
							{:else}
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v6m0 0l-3-3m3 3l3-3m-12 6h18" /></svg>
								Export Excel
							{/if}
						</button>
					</div>

					<div class="overflow-x-auto">
						<table class="w-full text-left text-sm">
							<thead>
								<tr class="bg-gray-50 dark:bg-gray-900/50 border-b border-gray-200 dark:border-gray-700 text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">
									<th class="p-4">Karyawan</th>
									<th class="p-4">Tanggal</th>
									<th class="p-4">Check In</th>
									<th class="p-4">Check Out</th>
									<th class="p-4">Status</th>
									<th class="p-4 text-center">Aksi</th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-100 dark:divide-gray-700/50 text-sm">
								{#each attendanceRecords as record (record.id)}
									<tr class="hover:bg-gray-50 dark:hover:bg-gray-700/50 transition-colors">
										<td class="p-4">
											<div class="font-medium text-gray-900 dark:text-white">{record.employee_name}</div>
											<div class="text-xs text-gray-500 mt-0.5">{record.department_name}</div>
										</td>
										<td class="p-4 whitespace-nowrap text-gray-700 dark:text-gray-300">
											{formatDate(record.date)}
										</td>
										<td class="p-4">
											<div class="font-semibold text-gray-900 dark:text-white">{formatTime(record.check_in_time)}</div>
											{#if record.check_in_location_name}
												<div class="text-xs text-gray-500 truncate max-w-[200px] mt-0.5" title={record.check_in_location_name}>{record.check_in_location_name}</div>
											{/if}
										</td>
										<td class="p-4">
											<div class="font-semibold text-gray-900 dark:text-white">{formatTime(record.check_out_time)}</div>
											{#if record.check_out_location_name}
												<div class="text-xs text-gray-500 truncate max-w-[200px] mt-0.5" title={record.check_out_location_name}>{record.check_out_location_name}</div>
											{/if}
										</td>
										<td class="p-4 whitespace-nowrap">
											<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium {getStatusBadge(record.status)}">
												{getStatusText(record.status)}
											</span>
										</td>
										<td class="p-4 text-center whitespace-nowrap">
											<button onclick={() => { selectedRecord = record; showDetail = true; }}
												class="inline-flex items-center px-2.5 py-1.5 text-xs font-medium text-[#1A56DB] dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 hover:bg-blue-100 dark:hover:bg-blue-900/40 rounded-lg transition cursor-pointer">
												Lihat Detail
											</button>
										</td>
									</tr>
								{:else}
									<tr>
										<td colspan="6" class="p-4 py-12 text-center text-gray-400">
											Tidak ada data absensi untuk filter terpilih.
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>

					<!-- Pagination -->
					{#if attendanceTotal > attendancePerPage}
						<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 rounded-b-xl">
							<div class="text-xs text-gray-500 dark:text-gray-400">
								Menampilkan {attendanceTotal === 0 ? 0 : (attendancePage - 1) * attendancePerPage + 1}-{Math.min(attendancePage * attendancePerPage, attendanceTotal)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{attendanceTotal}</span>
							</div>
							<div class="flex items-center gap-1.5">
								<button onclick={() => changeAttendancePage(attendancePage - 1)} disabled={attendancePage <= 1}
									class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
								{#each Array.from({ length: Math.min(5, attendanceTotalPages) }) as _, i (i)}
									{@const pageNum = Math.max(1, Math.min(attendancePage - 2, attendanceTotalPages - 4)) + i}
									{#if pageNum <= attendanceTotalPages}
										<button onclick={() => changeAttendancePage(pageNum)}
											class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === attendancePage ? 'bg-blue-600 text-white border-blue-600 shadow-sm' : 'border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'}">{pageNum}</button>
									{/if}
								{/each}
								<button onclick={() => changeAttendancePage(attendancePage + 1)} disabled={attendancePage >= attendanceTotalPages}
									class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-700 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-900 dark:hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
							</div>
						</div>
					{/if}
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


{#if showDetail && selectedRecord}
	<!-- Backdrop -->
	<div class="fixed inset-0 z-[100] bg-black/60 backdrop-blur-sm flex items-center justify-center p-4 animate-in fade-in duration-200">
		<!-- Modal Content Box -->
		<div class="bg-white dark:bg-gray-900 rounded-2xl max-w-2xl w-full max-h-[90vh] overflow-y-auto shadow-2xl border border-gray-150 dark:border-gray-800 flex flex-col animate-in zoom-in-95 duration-200">
			<!-- Header -->
			<div class="sticky top-0 z-10 bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border-b border-gray-200 dark:border-gray-800 h-14 flex items-center justify-between px-6">
				<h2 class="font-bold text-gray-900 dark:text-white text-base">Detail Absensi - {selectedRecord.employee_name}</h2>
				<button onclick={() => { showDetail = false; setTimeout(() => selectedRecord = null, 300); }} 
					class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
				</button>
			</div>
			
			<div class="p-6 space-y-5">
				<!-- Info Karyawan / Tanggal -->
				<div class="flex items-center justify-between bg-gray-50 dark:bg-gray-800/50 rounded-xl p-4 border border-gray-100 dark:border-gray-850">
					<div>
						<div class="text-sm font-bold text-gray-900 dark:text-white">{formatDate(selectedRecord.date)}</div>
						<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{selectedRecord.department_name}</div>
					</div>
					<div class="flex items-center gap-2">
						{#if selectedRecord.is_late && selectedRecord.late_minutes && selectedRecord.late_minutes > 0}
							<span class="text-xs text-amber-600 font-bold bg-amber-50 dark:bg-amber-900/30 px-2.5 py-1 rounded-full">Telat {selectedRecord.late_minutes} mnt</span>
						{/if}
						<span class="px-2.5 py-1 rounded-full text-xs font-semibold {getStatusBadge(selectedRecord.status)}">{getStatusText(selectedRecord.status)}</span>
					</div>
				</div>

				<div class="grid grid-cols-1 md:grid-cols-2 gap-5">
					<!-- Check In -->
					<div class="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col">
						<div class="bg-blue-50/50 dark:bg-blue-900/10 px-4 py-2.5 border-b border-gray-250 dark:border-gray-800 flex items-center justify-between">
							<span class="text-xs font-bold text-blue-600 dark:text-blue-400 uppercase tracking-wider">Check In</span>
							<span class="text-base font-bold text-gray-900 dark:text-white">{formatTime(selectedRecord.check_in_time)}</span>
						</div>
						
						<div class="p-4 space-y-4 flex-1 flex flex-col justify-between">
							<div class="grid grid-cols-2 gap-3">
								<!-- Foto -->
								<div>
									<p class="text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-1.5">Foto Selfie</p>
									{#if selectedRecord.check_in_photo_url}
										<div class="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-850 aspect-square shadow-sm">
											<img src={getPhotoUrl(selectedRecord.check_in_photo_url)} alt="Check In" class="w-full h-full object-cover" />
										</div>
									{:else}
										<div class="rounded-lg bg-gray-50 dark:bg-gray-850 border border-gray-200 dark:border-gray-800 aspect-square flex items-center justify-center">
											<span class="text-[10px] text-gray-400">No Photo</span>
										</div>
									{/if}
								</div>
								
								<!-- Map -->
								<div>
									<p class="text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-1.5">Peta Lokasi</p>
									{#if selectedRecord.check_in_lat && selectedRecord.check_in_lng}
										<div class="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-850 aspect-square shadow-sm">
											<iframe title="Lokasi Check In" width="100%" height="100%" frameborder="0" style="border:0"
												src={`https://maps.google.com/maps?q=${selectedRecord.check_in_lat},${selectedRecord.check_in_lng}&t=&z=15&ie=UTF8&iwloc=&output=embed`}
												allowfullscreen></iframe>
										</div>
									{:else}
										<div class="rounded-lg bg-gray-50 dark:bg-gray-850 border border-gray-200 dark:border-gray-800 aspect-square flex items-center justify-center">
											<span class="text-[10px] text-gray-400">No Map</span>
										</div>
									{/if}
								</div>
							</div>
							
							{#if selectedRecord.check_in_location_name}
								<div class="text-[11px] text-gray-500 leading-tight bg-gray-50 dark:bg-gray-900/50 p-2.5 rounded-lg border border-gray-100 dark:border-gray-850">
									{selectedRecord.check_in_location_name}
								</div>
							{/if}
						</div>
					</div>

					<!-- Check Out -->
					<div class="bg-white dark:bg-gray-800 rounded-xl overflow-hidden border border-gray-200 dark:border-gray-800 shadow-sm flex flex-col">
						<div class="bg-amber-50/50 dark:bg-amber-900/10 px-4 py-2.5 border-b border-gray-250 dark:border-gray-800 flex items-center justify-between">
							<span class="text-xs font-bold text-amber-600 dark:text-amber-400 uppercase tracking-wider">Check Out</span>
							<span class="text-base font-bold text-gray-900 dark:text-white">{formatTime(selectedRecord.check_out_time)}</span>
						</div>
						
						<div class="p-4 space-y-4 flex-1 flex flex-col justify-between">
							<div class="grid grid-cols-2 gap-3">
								<!-- Foto -->
								<div>
									<p class="text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-1.5">Foto Selfie</p>
									{#if selectedRecord.check_out_photo_url}
										<div class="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-850 aspect-square shadow-sm">
											<img src={getPhotoUrl(selectedRecord.check_out_photo_url)} alt="Check Out" class="w-full h-full object-cover" />
										</div>
									{:else}
										<div class="rounded-lg bg-gray-50 dark:bg-gray-850 border border-gray-200 dark:border-gray-800 aspect-square flex items-center justify-center">
											<span class="text-[10px] text-gray-400">No Photo</span>
										</div>
									{/if}
								</div>
								
								<!-- Map -->
								<div>
									<p class="text-[10px] font-semibold text-gray-400 uppercase tracking-wider mb-1.5">Peta Lokasi</p>
									{#if selectedRecord.check_out_lat && selectedRecord.check_out_lng}
										<div class="rounded-lg overflow-hidden border border-gray-200 dark:border-gray-850 aspect-square shadow-sm">
											<iframe title="Lokasi Check Out" width="100%" height="100%" frameborder="0" style="border:0"
												src={`https://maps.google.com/maps?q=${selectedRecord.check_out_lat},${selectedRecord.check_out_lng}&t=&z=15&ie=UTF8&iwloc=&output=embed`}
												allowfullscreen></iframe>
										</div>
									{:else}
										<div class="rounded-lg bg-gray-50 dark:bg-gray-850 border border-gray-200 dark:border-gray-800 aspect-square flex items-center justify-center">
											<span class="text-[10px] text-gray-400">No Map</span>
										</div>
									{/if}
								</div>
							</div>
							
							{#if selectedRecord.check_out_location_name}
								<div class="text-[11px] text-gray-500 leading-tight bg-gray-50 dark:bg-gray-900/50 p-2.5 rounded-lg border border-gray-100 dark:border-gray-850">
									{selectedRecord.check_out_location_name}
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
{/if}
