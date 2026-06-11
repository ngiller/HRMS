<script lang="ts">
	import { onMount } from 'svelte';
	import { dashboard as dashboardApi } from '$lib/api.js';

	interface Stat {
		label: string;
		value: string;
		change: string;
		changeClass: string;
		iconBg: string;
		iconColor: string;
		icon: string;
	}

	type PendingItem = {
		initials: string;
		name: string;
		type: string;
		detail: string;
		bg: string;
		text: string;
	};

	type NewEmployee = {
		initials: string;
		name: string;
		position: string;
		date: string;
		status: string;
	};

	type AttendanceDay = {
		day: string;
		count: number;
		height: number;
	};

	type Composition = {
		status: string;
		count: number;
	};

	let stats = $state<Stat[]>([]);
	let pendingApprovals = $state<PendingItem[]>([]);
	let newEmployees = $state<NewEmployee[]>([]);
	let attendanceData = $state<AttendanceDay[]>([]);
	let composition = $state<Composition[]>([]);
	let isLoading = $state(true);
	let errorMessage = $state('');

	async function loadDashboard() {
		isLoading = true;
		errorMessage = '';
		try {
			const response = await dashboardApi.get();
			const data = response.data;

			stats = [
				{
					label: 'TOTAL KARYAWAN',
					value: String(data.total_employees || 150),
					change: 'Aktif',
					changeClass: 'text-blue-600',
					iconBg: 'bg-blue-50',
					iconColor: 'text-blue-600',
					icon: 'M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z'
				},
				{
					label: 'HADIR HARI INI',
					value: String(data.present_today || 0),
					change: `dari ${data.total_employees || 150} karyawan`,
					changeClass: 'text-green-600',
					iconBg: 'bg-green-50',
					iconColor: 'text-green-600',
					icon: 'M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'
				},
				{
					label: 'PENDING APPROVAL',
					value: String(data.pending_approvals || 0),
					change: data.pending_approvals > 0 ? 'Perlu ditindaklanjuti' : 'Tidak ada',
					changeClass: data.pending_approvals > 0 ? 'text-amber-600' : 'text-green-600',
					iconBg: 'bg-amber-50',
					iconColor: 'text-amber-600',
					icon: 'M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'
				},
				{
					label: 'PAYROLL BULAN INI',
					value: data.payroll_this_month || 'Rp0',
					change: 'Take home pay',
					changeClass: 'text-purple-600',
					iconBg: 'bg-purple-50',
					iconColor: 'text-purple-600',
					icon: 'M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18'
				}
			];

			// Attendance data from API
			attendanceData = (data.attendance_by_day || []).map((d: AttendanceDay) => ({
				...d,
				height: Math.max(20, Math.min(96, Math.round((d.count / 25) * 96)))
			}));

			// Composition from API
			composition = data.composition || [];

			// Pending approvals (shown as placeholder if no real data)
			if (data.pending_approvals > 0) {
				pendingApprovals = [
					{ initials: '--', name: 'Data approval', type: 'Menunggu', detail: `${data.pending_approvals} permintaan`, bg: 'bg-gray-100', text: 'text-gray-600' }
				];
			} else {
				pendingApprovals = [
					{ initials: '✓', name: 'Semua', type: 'Tidak ada', detail: 'Tidak ada pending approval', bg: 'bg-green-100', text: 'text-green-600' }
				];
			}

			// Recent employees from API
			newEmployees = (data.recent_employees || []).slice(0, 5).map((e: any) => {
				const nameParts = (e.full_name || '').split(' ');
				const initials = nameParts.length > 1
					? (nameParts[0][0] || '') + (nameParts[1][0] || '')
					: (e.full_name || '').substring(0, 2).toUpperCase();
				return {
					initials: initials || '--',
					name: e.full_name || '-',
					position: e.position_name || '-',
					date: e.join_date ? new Date(e.join_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-',
					status: e.employment_status === 'percobaan' ? 'Percobaan' : e.employment_status === 'kontrak' ? 'Kontrak' : 'Tetap'
				};
			});

			if (newEmployees.length === 0) {
				newEmployees = [
					{ initials: '--', name: 'Belum ada data', position: '', date: '', status: '' }
				];
			}
		} catch (error) {
			errorMessage = 'Gagal memuat data dashboard';
			console.error('Dashboard error:', error);
		} finally {
			isLoading = false;
		}
	}

	onMount(loadDashboard);

	function getTypeIcon(type: string): string {
		const map: Record<string, string> = {
			'Cuti Sakit': '🤒',
			'Reimbursement': '💰',
			'Lembur': '⏰',
			'Cuti Tahunan': '🏖️'
		};
		return map[type] || '📋';
	}
</script>

{#if isLoading}
	<!-- Loading Skeleton -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
		{#each [1,2,3,4] as _}
			<div class="bg-white border border-gray-200 rounded-xl p-5 animate-pulse">
				<div class="h-3 bg-gray-100 rounded w-24 mb-3"></div>
				<div class="h-7 bg-gray-100 rounded w-16 mb-2"></div>
				<div class="h-3 bg-gray-100 rounded w-32"></div>
			</div>
		{/each}
	</div>
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
		<div class="bg-white border border-gray-200 rounded-xl p-5 lg:col-span-2 animate-pulse">
			<div class="h-4 bg-gray-100 rounded w-40 mb-6"></div>
			<div class="h-40 bg-gray-50 rounded"></div>
		</div>
		<div class="bg-white border border-gray-200 rounded-xl p-5 animate-pulse">
			<div class="h-4 bg-gray-100 rounded w-36 mb-6"></div>
			{#each [1,2,3,4] as _}
				<div class="flex justify-between mb-3">
					<div class="h-3 bg-gray-100 rounded w-16"></div>
					<div class="h-3 bg-gray-100 rounded w-8"></div>
				</div>
				<div class="h-2.5 bg-gray-100 rounded-full mb-4"></div>
			{/each}
		</div>
	</div>
{:else if errorMessage}
	<div class="bg-red-50 border border-red-200 rounded-xl p-8 text-center">
		<p class="text-red-700 text-sm">{errorMessage}</p>
		<button
			onclick={loadDashboard}
			class="mt-3 text-sm text-[#1A56DB] hover:underline font-medium cursor-pointer"
		>
			Coba lagi
		</button>
	</div>
{:else}
	<!-- Stats Cards -->
	<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
		{#each stats as stat}
			<div class="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-md transition-shadow">
				<div class="flex items-center justify-between">
					<div class="flex-1 min-w-0">
						<div class="text-xs font-semibold text-gray-400 tracking-wider uppercase">{stat.label}</div>
						<div class="text-2xl font-bold text-gray-900 mt-1.5 tabular-nums">{stat.value}</div>
						<div class="text-xs mt-1 {stat.changeClass}">{stat.change}</div>
					</div>
					<div class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 {stat.iconBg}">
						<svg class="w-6 h-6 {stat.iconColor}" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d={stat.icon} />
						</svg>
					</div>
				</div>
			</div>
		{/each}
	</div>

	<!-- Charts & Composition -->
	<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
		<!-- Attendance Chart -->
		<div class="bg-white border border-gray-200 rounded-xl p-5 lg:col-span-2">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-gray-900">Kehadiran 7 Hari Terakhir</h3>
				<select class="text-xs border border-gray-200 rounded-lg px-2 py-1.5 text-gray-500 bg-white outline-none" aria-label="Pilih periode">
					<option>7 Hari</option>
				</select>
			</div>
			<div class="h-48 flex items-end justify-around gap-2 px-2 pb-1">
				{#if attendanceData.length > 0}
					{#each attendanceData as item}
						<div class="flex flex-col items-center gap-1 flex-1">
							<span class="text-xs text-gray-400 tabular-nums">{item.count}</span>
							<div
								class="w-full max-w-[40px] rounded-t transition-all duration-500 hover:opacity-80 cursor-pointer"
								style="height: {item.height}px; background: #4ade80;"
							></div>
							<span class="text-xs text-gray-500">{item.day?.trim() || '-'}</span>
						</div>
					{/each}
				{:else}
					<div class="w-full h-full flex items-center justify-center">
						<p class="text-sm text-gray-400">Belum ada data kehadiran</p>
					</div>
				{/if}
			</div>
		</div>

		<!-- Employee Composition -->
		<div class="bg-white border border-gray-200 rounded-xl p-5">
			<h3 class="text-sm font-semibold text-gray-900 mb-4">Komposisi Karyawan</h3>
			<div class="space-y-4">
				{#if composition.length > 0}
					{@const total = composition.reduce((sum, c) => sum + c.count, 0) || 1}
					{#each composition as item}
						{@const pct = Math.round((item.count / total) * 100)}
						<div>
							<div class="flex justify-between text-sm mb-1.5">
								<span class="text-gray-600 capitalize">{item.status}</span>
								<span class="font-medium text-gray-900 tabular-nums">{item.count}</span>
							</div>
							<div class="w-full bg-gray-100 rounded-full h-2.5" role="progressbar" aria-valuenow={pct} aria-valuemin="0" aria-valuemax="100">
								<div class="bg-[#1A56DB] h-2.5 rounded-full" style="width: {pct}%"></div>
							</div>
						</div>
					{/each}
					<div class="mt-4 pt-4 border-t border-gray-100">
						<div class="flex justify-between text-sm">
							<span class="text-gray-600 font-medium">Total</span>
							<span class="font-bold text-gray-900 tabular-nums">{total}</span>
						</div>
					</div>
				{:else}
					<p class="text-sm text-gray-400 text-center py-8">Belum ada data</p>
				{/if}
			</div>
		</div>
	</div>

	<!-- Tables Row -->
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
		<!-- Pending Approval -->
		<div class="bg-white border border-gray-200 rounded-xl p-5">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-gray-900">Pending Approval</h3>
			</div>
			<div class="space-y-1">
				{#each pendingApprovals as item}
					<div class="flex items-center justify-between py-2.5 px-2 rounded-lg hover:bg-gray-50 transition">
						<div class="flex items-center gap-3">
							<div class="w-9 h-9 {item.bg} rounded-full flex items-center justify-center text-xs font-semibold {item.text}">
								{item.initials}
							</div>
							<div>
								<div class="text-sm font-medium text-gray-900">
									<span class="mr-1">{getTypeIcon(item.type)}</span>
									{item.type}
								</div>
								<div class="text-xs text-gray-400">{item.name} &mdash; {item.detail}</div>
							</div>
						</div>
						<span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200">
							{item.initials === '✓' ? 'Selesai' : 'Menunggu'}
						</span>
					</div>
				{/each}
			</div>
		</div>

		<!-- New Employees -->
		<div class="bg-white border border-gray-200 rounded-xl p-5">
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-gray-900">Karyawan Baru</h3>
			</div>
			<div class="space-y-1">
				{#each newEmployees as item}
					<div class="flex items-center justify-between py-2.5 px-2 rounded-lg hover:bg-gray-50 transition cursor-pointer">
						<div class="flex items-center gap-3">
							<div class="w-9 h-9 bg-gradient-to-br from-gray-200 to-gray-300 rounded-full flex items-center justify-center text-xs font-semibold text-gray-600">
								{item.initials}
							</div>
							<div>
								<div class="text-sm font-medium text-gray-900">{item.name}</div>
								<div class="text-xs text-gray-400">{item.position}{item.date ? ` &mdash; Bergabung ${item.date}` : ''}</div>
							</div>
						</div>
						{#if item.status}
							<span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-50 text-green-700 border border-green-200">
								{item.status}
							</span>
						{/if}
					</div>
				{/each}
			</div>
		</div>
	</div>

	<p class="text-xs text-gray-400 text-center mt-8 pb-4">
		HRMS &mdash; Sistem Informasi Sumber Daya Manusia &copy; 2026
	</p>
{/if}
