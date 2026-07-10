<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { dashboard as dashboardApi, auth, announcements as announcementsApi, attendance, leaveRequests } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';

	let Chart: any;
	let chartLoaded = false;

	// ── Types ──
	interface Stat {
		label: string;
		value: string;
		change: string;
		changeClass: string;
		iconBg: string;
		iconColor: string;
		icon: string;
	}

	interface PendingItem {
		initials: string;
		name: string;
		type: string;
		detail: string;
		bg: string;
		text: string;
	}

	interface NewEmployee {
		initials: string;
		name: string;
		position: string;
		date: string;
		status: string;
	}

	interface AttendanceDay { day: string; count: number }
	interface Composition { status: string; count: number }
	interface MonthlyTrend { month: string; count: number }
	interface GenderBreakdown { gender: string; count: number }
	interface DepartmentStat { department_name: string; employee_count: number }

	interface DashboardData {
		total_employees: number;
		active_employees: number;
		present_today: number;
		attendance_rate: number;
		pending_approvals: number;
		payroll_this_month: string;
		attendance_by_day: AttendanceDay[];
		monthly_trend: MonthlyTrend[];
		composition: Composition[];
		gender_breakdown: GenderBreakdown[];
		department_stats: DepartmentStat[];
		recent_employees: NewEmployee[];
	}

	// ── User / Greeting ──
	let userName = $derived((auth.getUser() as any)?.full_name?.split(' ')[0] || 'Pengguna');
	let todayDate = $derived(new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }));

	// ── State ──
	let stats = $state<Stat[]>([]);
	let pendingApprovals = $state<PendingItem[]>([]);
	let newEmployees = $state<NewEmployee[]>([]);
	let isLoading = $state(true);
	let errorMessage = $state('');
	let activeTab = $state<'overview' | 'departments' | 'analytics'>('overview');
	
	let isEmployee = $derived.by(() => {
		const user = auth.getUser() as any;
		if (!user) return true;
		const slug = (user.role_slug || '').toLowerCase();
		const name = (user.role_name || '').toLowerCase();
		const adminRoles = ['admin', 'hr', 'owner', 'superadmin', 'super_admin', 'super admin', 'hr_manager', 'hr_staff', 'hr manager', 'hr staff'];
		return !adminRoles.includes(slug) && !adminRoles.includes(name);
	});
	let employeeAnnouncements = $state<any[]>([]);
	let todayStatus = $state<any>(null);
	let leaveBalance = $state<any>(null);
	let currentTime = $state(new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' }));

	onMount(() => {
		const timer = setInterval(() => {
			currentTime = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
		}, 1000);
		return () => clearInterval(timer);
	});

	// Unique canvas refs per tab to avoid bind:this conflicts
	let overviewAttendanceCanvas = $state<HTMLCanvasElement>(undefined!);
	let overviewCompositionCanvas = $state<HTMLCanvasElement>(undefined!);
	let overviewTrendCanvas = $state<HTMLCanvasElement>(undefined!);
	let overviewGenderCanvas = $state<HTMLCanvasElement>(undefined!);
	let departmentsDeptCanvas = $state<HTMLCanvasElement>(undefined!);
	let analyticsTrendCanvas = $state<HTMLCanvasElement>(undefined!);
	let analyticsCompositionCanvas = $state<HTMLCanvasElement>(undefined!);
	let analyticsGenderCanvas = $state<HTMLCanvasElement>(undefined!);
	let analyticsDeptCanvas = $state<HTMLCanvasElement>(undefined!);

	let overviewAttChart: any = null;
	let overviewCompChart: any = null;
	let overviewTrendChart: any = null;
	let overviewGenderChart: any = null;
	let deptBarChart: any = null;
	let analyticsTrendChart: any = null;
	let analyticsCompChart: any = null;
	let analyticsGenderChart: any = null;
	let analyticsDeptChart: any = null;

	// ── Helpers ──
	function getInitials(name: string): string {
		if (!name) return '--';
		const parts = name.split(' ');
		return parts.length > 1
			? (parts[0][0] || '') + (parts[1][0] || '')
			: name.substring(0, 2).toUpperCase();
	}

	function getTypeIcon(type: string): string {
		const icons: Record<string, string> = {
			'Cuti Sakit': '🤒', 'Reimbursement': '💰', 'Lembur': '⏰',
			'Cuti Tahunan': '🏖️', 'Cuti': '🏖️',
		};
		return icons[type] || '📋';
	}

	const chartColors = ['#1A56DB', '#10B981', '#F59E0B', '#8B5CF6', '#0EA5E9', '#F43F5E', '#14B8A6', '#F97316'];
	const genderLabels: Record<string, string> = { laki_laki: 'Laki-laki', perempuan: 'Perempuan' };
	const statusColors: Record<string, string> = { tetap: '#10B981', kontrak: '#F59E0B', percobaan: '#0EA5E9', magang: '#8B5CF6' };

	// ── Generic Chart Renderers ──
	function renderBarChart(canvas: HTMLCanvasElement, chart: any, setChart: (c: any) => void, labels: string[], values: number[], label: string, horizontal = false) {
		if (!canvas || !canvas.getContext) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		chart?.destroy();

		// For horizontal bar, sort descending
		const sortedLabels = horizontal ? [...labels].reverse() : labels;
		const sortedValues = horizontal ? [...values].reverse() : values;

		const newChart = new Chart(ctx, {
			type: 'bar',
			data: {
				labels: sortedLabels,
				datasets: [{
					label,
					data: sortedValues,
					backgroundColor: sortedValues.map((v, i) => chartColors[i % chartColors.length] + '33'),
					borderColor: sortedValues.map((_, i) => chartColors[i % chartColors.length]),
					borderWidth: 2,
					borderRadius: 4,
				}]
			},
			options: {
				...(horizontal ? { indexAxis: 'y' as const } : {}),
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: { backgroundColor: '#1F2937', padding: 10, cornerRadius: 8 },
				},
				scales: {
					x: {
						beginAtZero: true,
						ticks: { stepSize: 1, font: { size: 11 }, color: '#9CA3AF' },
						grid: { color: 'rgba(0,0,0,0.04)' }
					},
					y: {
						ticks: { font: { size: 11 }, color: horizontal ? '#6B7280' : '#9CA3AF' },
						grid: horizontal ? { display: false } : { color: 'rgba(0,0,0,0.04)' }
					}
				}
			}
		});
		setChart(newChart);
	}

	function renderDoughnutChart(canvas: HTMLCanvasElement, chart: any, setChart: (c: any) => void, labels: string[], values: number[], bgColors: string[]) {
		if (!canvas || !canvas.getContext) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		chart?.destroy();

		const newChart = new Chart(ctx, {
			type: 'doughnut',
			data: {
				labels,
				datasets: [{
					data: values,
					backgroundColor: bgColors,
					borderColor: '#ffffff',
					borderWidth: 2,
					hoverOffset: 8,
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				cutout: '70%',
				plugins: {
					legend: {
						position: 'bottom',
						labels: { padding: 12, usePointStyle: true, pointStyle: 'circle', font: { size: 11 }, color: '#6B7280' }
					},
					tooltip: {
						backgroundColor: '#1F2937', padding: 10, cornerRadius: 8,
						callbacks: {
							label: (ctx: any) => {
								const total = (ctx.dataset.data as number[]).reduce((a: number, b: number) => a + b, 0);
								const pct = ((ctx.parsed as number) / total * 100).toFixed(1);
								return ` ${ctx.label}: ${ctx.parsed} (${pct}%)`;
							}
						}
					}
				}
			}
		});
		setChart(newChart);
	}

	function renderLineChart(canvas: HTMLCanvasElement, chart: any, setChart: (c: any) => void, labels: string[], values: number[]) {
		if (!canvas || !canvas.getContext) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		chart?.destroy();

		const newChart = new Chart(ctx, {
			type: 'line',
			data: {
				labels,
				datasets: [{
					label: 'Karyawan Baru',
					data: values,
					borderColor: '#1A56DB',
					backgroundColor: 'rgba(26, 86, 219, 0.08)',
					fill: true,
					tension: 0.4,
					pointBackgroundColor: '#1A56DB',
					pointBorderColor: '#ffffff',
					pointBorderWidth: 2,
					pointRadius: 4,
					pointHoverRadius: 6,
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: { backgroundColor: '#1F2937', padding: 10, cornerRadius: 8 },
				},
				scales: {
					y: {
						beginAtZero: true,
						ticks: { stepSize: 1, font: { size: 11 }, color: '#9CA3AF' },
						grid: { color: 'rgba(0,0,0,0.04)' }
					},
					x: {
						ticks: { font: { size: 11 }, color: '#9CA3AF' },
						grid: { display: false }
					}
				}
			}
		});
		setChart(newChart);
	}

	function renderAttendanceChart(canvas: HTMLCanvasElement, chart: any, setChart: (c: any) => void, days: AttendanceDay[]) {
		if (!canvas || !canvas.getContext) return;
		const ctx = canvas.getContext('2d');
		if (!ctx) return;
		chart?.destroy();

		const labels = days.map(d => d.day?.trim().substring(0, 3) || '-');
		const values = days.map(d => d.count);

		const newChart = new Chart(ctx, {
			type: 'bar',
			data: {
				labels,
				datasets: [{
					label: 'Kehadiran',
					data: values,
					backgroundColor: values.map(v => v > 0 ? 'rgba(16, 185, 129, 0.7)' : 'rgba(239, 68, 68, 0.5)'),
					borderColor: values.map(v => v > 0 ? 'rgb(16, 185, 129)' : 'rgb(239, 68, 68)'),
					borderWidth: 1,
					borderRadius: 4,
				}]
			},
			options: {
				responsive: true,
				maintainAspectRatio: false,
				plugins: {
					legend: { display: false },
					tooltip: {
						backgroundColor: '#1F2937', titleFont: { size: 12 }, bodyFont: { size: 13 },
						padding: 10, cornerRadius: 8,
						callbacks: { label: (ctx: any) => `${ctx.raw} karyawan hadir` }
					}
				},
				scales: {
					y: { beginAtZero: true, ticks: { stepSize: 1, font: { size: 11 }, color: '#9CA3AF' }, grid: { color: 'rgba(0,0,0,0.04)' } },
					x: { ticks: { font: { size: 11 }, color: '#9CA3AF' }, grid: { display: false } }
				}
			}
		});
		setChart(newChart);
	}

	// ── Dashboard Load ──
	async function loadDashboard() {
		isLoading = true;
		errorMessage = '';
		try {
			// Dynamic import Chart.js — reduces initial bundle by ~400 kB
			if (!chartLoaded && !isEmployee) {
				const chartModule = await import('chart.js');
				Chart = chartModule.Chart;
				Chart.register(...chartModule.registerables);
				chartLoaded = true;
			}
			
			if (isEmployee) {
				// Fetch data for employee dashboard
				try {
					const [annRes, attRes, leaveRes] = await Promise.all([
						announcementsApi.list(1, 5, '').catch(() => ({ data: [] })),
						attendance.getTodayStatus().catch(() => ({ data: null })),
						leaveRequests.getMyBalances().catch(() => ({ data: [] }))
					]);
					
					employeeAnnouncements = annRes.data || [];
					todayStatus = attRes.data;
					
					if (leaveRes.data && Array.isArray(leaveRes.data)) {
						leaveBalance = leaveRes.data.find((b: any) => b.leave_type_name === 'Tahunan' || b.leave_type_name === 'Cuti Tahunan') || leaveRes.data[0];
					}
				} catch (e) {
					console.error("Error loading employee dashboard", e);
				}
				return;
			}
			
			const response = await dashboardApi.get();
			const data: DashboardData = response.data;

			stats = [
				{
					label: 'TOTAL KARYAWAN', value: String(data.total_employees || 0),
					change: `${data.active_employees || 0} aktif`, changeClass: 'text-blue-600',
					iconBg: 'bg-blue-50', iconColor: 'text-blue-600',
					icon: 'M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z'
				},
				{
					label: 'HADIR HARI INI', value: String(data.present_today || 0),
					change: `${Math.round(data.attendance_rate || 0)}% kehadiran`, changeClass: 'text-green-600',
					iconBg: 'bg-green-50', iconColor: 'text-green-600',
					icon: 'M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'
				},
				{
					label: 'PENDING APPROVAL', value: String(data.pending_approvals || 0),
					change: data.pending_approvals > 0 ? 'Perlu ditindaklanjuti' : 'Tidak ada',
					changeClass: data.pending_approvals > 0 ? 'text-amber-600' : 'text-green-600',
					iconBg: 'bg-amber-50', iconColor: 'text-amber-600',
					icon: 'M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z'
				},
				{
					label: 'PAYROLL BULAN INI', value: data.payroll_this_month && data.payroll_this_month !== 'Rp0' ? data.payroll_this_month : 'Rp 0',
					change: 'Take home pay', changeClass: 'text-purple-600',
					iconBg: 'bg-purple-50', iconColor: 'text-purple-600',
					icon: 'M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18'
				}
			];

			pendingApprovals = data.pending_approvals > 0
				? [{ initials: '--', name: 'Data approval', type: 'Menunggu', detail: `${data.pending_approvals} permintaan`, bg: 'bg-amber-100', text: 'text-amber-600' }]
				: [{ initials: '✓', name: 'Semua', type: 'Tidak ada', detail: 'Tidak ada pending approval', bg: 'bg-green-100', text: 'text-green-600' }];

			newEmployees = (data.recent_employees || []).slice(0, 5).map((e: any) => ({
				initials: getInitials(e.full_name),
				name: e.full_name || '-',
				position: e.position_name || '-',
				date: e.join_date ? new Date(e.join_date).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' }) : '-',
				status: e.employment_status === 'percobaan' ? 'Percobaan' : e.employment_status === 'kontrak' ? 'Kontrak' : 'Tetap'
			}));
			if (newEmployees.length === 0) {
				newEmployees = [{ initials: '--', name: 'Belum ada data', position: '', date: '', status: '' }];
			}

			// Render charts after DOM is ready
			requestAnimationFrame(() => {
				destroyAllCharts();
				renderAllCharts(data);
			});
		} catch (error) {
			errorMessage = 'Gagal memuat data dashboard';
			console.error('Dashboard error:', error);
		} finally {
			isLoading = false;
		}
	}

	function renderAllCharts(data: DashboardData) {
		const att = data.attendance_by_day || [];
		const comp = data.composition || [];
		const trend = data.monthly_trend || [];
		const gender = data.gender_breakdown || [];
		const dept = data.department_stats || [];

		// Overview tab
		renderAttendanceChart(overviewAttendanceCanvas, overviewAttChart, (c) => { overviewAttChart = c; }, att);
		renderDoughnutChart(overviewCompositionCanvas, overviewCompChart, (c) => { overviewCompChart = c; },
			comp.map(c => c.status === 'tetap' ? 'Tetap' : c.status === 'kontrak' ? 'Kontrak' : c.status === 'percobaan' ? 'Percobaan' : c.status === 'magang' ? 'Magang' : c.status),
			comp.map(c => c.count),
			comp.map(c => statusColors[c.status] || '#9CA3AF')
		);
		renderLineChart(overviewTrendCanvas, overviewTrendChart, (c) => { overviewTrendChart = c; }, trend.map(t => t.month), trend.map(t => t.count));
		renderDoughnutChart(overviewGenderCanvas, overviewGenderChart, (c) => { overviewGenderChart = c; },
			gender.map(g => genderLabels[g.gender] || g.gender),
			gender.map(g => g.count),
			['#0EA5E9', '#F43F5E']
		);

		// Departments tab
		renderBarChart(departmentsDeptCanvas, deptBarChart, (c) => { deptBarChart = c; },
			dept.map(d => d.department_name), dept.map(d => d.employee_count), 'Jumlah Karyawan', true
		);

		// Analytics tab
		renderLineChart(analyticsTrendCanvas, analyticsTrendChart, (c) => { analyticsTrendChart = c; }, trend.map(t => t.month), trend.map(t => t.count));
		renderDoughnutChart(analyticsCompositionCanvas, analyticsCompChart, (c) => { analyticsCompChart = c; },
			comp.map(c => c.status === 'tetap' ? 'Tetap' : c.status === 'kontrak' ? 'Kontrak' : c.status === 'percobaan' ? 'Percobaan' : c.status === 'magang' ? 'Magang' : c.status),
			comp.map(c => c.count),
			comp.map(c => statusColors[c.status] || '#9CA3AF')
		);
		renderDoughnutChart(analyticsGenderCanvas, analyticsGenderChart, (c) => { analyticsGenderChart = c; },
			gender.map(g => genderLabels[g.gender] || g.gender),
			gender.map(g => g.count),
			['#0EA5E9', '#F43F5E']
		);
		renderBarChart(analyticsDeptCanvas, analyticsDeptChart, (c) => { analyticsDeptChart = c; },
			dept.map(d => d.department_name), dept.map(d => d.employee_count), 'Jumlah Karyawan', true
		);
	}

	function destroyAllCharts() {
		const all = [overviewAttChart, overviewCompChart, overviewTrendChart, overviewGenderChart,
			deptBarChart, analyticsTrendChart, analyticsCompChart, analyticsGenderChart, analyticsDeptChart];
		all.forEach(c => c?.destroy());
		overviewAttChart = overviewCompChart = overviewTrendChart = overviewGenderChart = null;
		deptBarChart = analyticsTrendChart = analyticsCompChart = analyticsGenderChart = analyticsDeptChart = null;
	}

	onMount(loadDashboard);
	onDestroy(() => destroyAllCharts());
</script>

<div class="max-w-full mx-auto">
	{#if isEmployee && !isLoading}
		<!-- Employee Dashboard (Talenta Style) -->
		<div class="max-w-6xl mx-auto py-4 px-4 sm:px-6 lg:px-8">
			<!-- Header -->
			<div class="flex items-center justify-between mb-6">
				<div>
					<h1 class="text-2xl font-bold text-gray-900">Halo, {userName}!</h1>
					<p class="text-sm text-gray-500 mt-1">{todayDate}</p>
				</div>
				<div class="hidden sm:block">
					<div class="px-4 py-2 bg-blue-50 text-blue-700 rounded-lg font-medium text-sm border border-blue-100 flex items-center gap-2">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
						{currentTime}
					</div>
				</div>
			</div>
			
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
				<!-- Left Column: Attendance & Quick Links -->
				<div class="lg:col-span-2 space-y-6">
					<!-- Live Attendance Card -->
					<div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
						<div class="p-5 border-b border-gray-100 flex justify-between items-center bg-gray-50/50">
							<h2 class="font-semibold text-gray-900 flex items-center gap-2">
								<svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
								Live Attendance
							</h2>
							<span class="text-xs font-medium text-gray-500">{todayStatus?.schedule_name || 'Tidak ada jadwal'}</span>
						</div>
						
						<div class="p-6">
							{#if todayStatus}
								<div class="grid grid-cols-2 divide-x divide-gray-200 gap-4 mb-6">
									<div class="text-center sm:text-left px-2 sm:px-4">
										<p class="text-xs text-gray-500 font-medium uppercase tracking-wider mb-1">Check In</p>
										<div class="text-2xl sm:text-3xl font-bold text-gray-900 tabular-nums">
											{todayStatus.has_checked_in && todayStatus.record?.check_in_time ? new Date(todayStatus.record.check_in_time).toLocaleTimeString('id-ID', {hour:'2-digit', minute:'2-digit'}) : '--:--'}
										</div>
										<p class="text-[10px] sm:text-xs text-gray-400 mt-1">Jadwal: {todayStatus.schedule_start || '--:--'}</p>
									</div>
									
									<div class="text-center sm:text-left px-2 sm:px-4">
										<p class="text-xs text-gray-500 font-medium uppercase tracking-wider mb-1">Check Out</p>
										<div class="text-2xl sm:text-3xl font-bold text-gray-900 tabular-nums">
											{todayStatus.has_checked_out && todayStatus.record?.check_out_time ? new Date(todayStatus.record.check_out_time).toLocaleTimeString('id-ID', {hour:'2-digit', minute:'2-digit'}) : '--:--'}
										</div>
										<p class="text-[10px] sm:text-xs text-gray-400 mt-1">Jadwal: {todayStatus.schedule_end || '--:--'}</p>
									</div>
								</div>
								
								<div class="flex gap-3">
									<button 
										onclick={() => goto('/absensi')}
										class="flex-1 py-3 px-4 rounded-lg font-bold text-white text-sm transition-colors text-center {todayStatus.has_checked_in ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'}"
										disabled={todayStatus.has_checked_in}
									>
										Check In
									</button>
									<button 
										onclick={() => goto('/absensi')}
										class="flex-1 py-3 px-4 rounded-lg font-bold text-white text-sm transition-colors text-center {todayStatus.has_checked_out || !todayStatus.has_checked_in ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-orange-500 hover:bg-orange-600'}"
										disabled={todayStatus.has_checked_out || !todayStatus.has_checked_in}
									>
										Check Out
									</button>
								</div>
							{:else}
								<div class="text-center py-6 text-gray-500 text-sm">Gagal memuat status absensi.</div>
							{/if}
						</div>
					</div>
					
					<!-- Shortcuts -->
					<div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden p-5">
						<h2 class="font-semibold text-gray-900 mb-4">Akses Cepat</h2>
						<div class="grid grid-cols-4 gap-2 sm:gap-4">
							<a href="/cuti" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-indigo-50 text-indigo-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 012.25-2.25h13.5A2.25 2.25 0 0121 7.5v11.25m-18 0A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75m-18 0v-7.5A2.25 2.25 0 015.25 9h13.5A2.25 2.25 0 0121 11.25v7.5" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Ajukan Cuti</span>
							</a>
							<a href="/lembur" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-amber-50 text-amber-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Lembur</span>
							</a>
							<a href="/absensi-manual" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-emerald-50 text-emerald-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Revisi Absen</span>
							</a>
							<a href="/reimbursement" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-pink-50 text-pink-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 004.5 6h.75m13.5 0h.75a.75.75 0 00.75-.75V4.5m-15 15h15" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Klaim Dana</span>
							</a>
						</div>
					</div>
				</div>
				
				<!-- Right Column: Leave Balances & Announcements -->
				<div class="space-y-6">
					<!-- Leave Balance -->
					<a href="/cuti" class="block bg-gradient-to-br from-indigo-500 to-blue-600 rounded-xl shadow-sm overflow-hidden text-white p-5 hover:shadow-md transition-shadow active:scale-[0.98]">
						<div class="flex items-center justify-between mb-4 opacity-90">
							<h2 class="font-semibold text-sm">Sisa Cuti Tahunan</h2>
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" /></svg>
						</div>
						
						<div class="flex items-end gap-2">
							<span class="text-4xl font-bold">{leaveBalance?.remaining ?? '-'}</span>
							<span class="text-blue-100 font-medium mb-1">Hari</span>
						</div>
						
						{#if leaveBalance?.expired_at}
							<div class="mt-4 text-xs text-blue-200">
								Berlaku sampai {new Date(leaveBalance.expired_at).toLocaleDateString('id-ID', {day: 'numeric', month: 'long', year: 'numeric'})}
							</div>
						{/if}
						<div class="mt-4 block text-center py-2 bg-white/20 hover:bg-white/30 rounded-lg text-sm font-medium transition-colors">Lihat Detail</div>
					</a>
					
					<!-- Announcements Mini Feed -->
					<div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden flex flex-col">
						<div class="p-4 border-b border-gray-100 flex items-center justify-between">
							<h2 class="font-semibold text-gray-900 text-sm">Pengumuman</h2>
							<a href="/pengumuman" class="text-xs font-medium text-blue-600 hover:underline">Lihat Semua</a>
						</div>
						
						<div class="p-0 max-h-[300px] overflow-y-auto custom-scrollbar">
							{#if employeeAnnouncements.length === 0}
								<div class="p-6 text-center">
									<p class="text-gray-400 text-xs font-medium">Tidak ada pengumuman</p>
								</div>
							{:else}
								<div class="divide-y divide-gray-50">
									{#each employeeAnnouncements as ann}
										<a href="/pengumuman" class="block p-4 hover:bg-gray-50 transition-colors">
											<div class="flex gap-3">
												<div class="shrink-0 mt-0.5">
													<div class="w-8 h-8 rounded-full flex items-center justify-center {ann.announcement_type === 'important' ? 'bg-orange-100 text-orange-600' : ann.announcement_type === 'emergency' ? 'bg-red-100 text-red-600' : 'bg-blue-100 text-blue-600'}">
														<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z" /></svg>
													</div>
												</div>
												<div class="min-w-0">
													<h3 class="text-sm font-semibold text-gray-900 truncate">{ann.title}</h3>
													<p class="text-xs text-gray-500 mt-1 line-clamp-2 leading-relaxed">{ann.content || 'Klik untuk melihat detail'}</p>
													<p class="text-[10px] text-gray-400 mt-2 font-medium">{new Date(ann.published_at || ann.created_at).toLocaleDateString('id-ID', {day: 'numeric', month: 'short'})}</p>
												</div>
											</div>
										</a>
									{/each}
								</div>
							{/if}
						</div>
					</div>
				</div>
			</div>
		</div>
		
		<style>
			.custom-scrollbar::-webkit-scrollbar { width: 4px; }
			.custom-scrollbar::-webkit-scrollbar-track { background: transparent; }
			.custom-scrollbar::-webkit-scrollbar-thumb { background-color: #E5E7EB; border-radius: 10px; }
			.custom-scrollbar:hover::-webkit-scrollbar-thumb { background-color: #D1D5DB; }
		</style>
	{:else}
		<!-- Page Header (Admin/HR Dashboard) -->
	<div class="flex items-center justify-between mb-6">
		<div>
			<h1 class="text-xl font-bold text-gray-900">Dashboard</h1>
			<p class="text-sm text-gray-500 mt-0.5">Overview sumber daya manusia perusahaan</p>
		</div>
		<div class="flex items-center gap-2 bg-white border border-gray-200 rounded-lg p-1">
			<button
				class="px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer"
				class:bg-gray-900={activeTab === 'overview'}
				class:text-white={activeTab === 'overview'}
				class:text-gray-500={activeTab !== 'overview'}
				onclick={() => activeTab = 'overview'}
			>Overview</button>
			<button
				class="px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer"
				class:bg-gray-900={activeTab === 'departments'}
				class:text-white={activeTab === 'departments'}
				class:text-gray-500={activeTab !== 'departments'}
				onclick={() => activeTab = 'departments'}
			>Departemen</button>
			<button
				class="px-3 py-1.5 text-xs font-medium rounded-md transition cursor-pointer"
				class:bg-gray-900={activeTab === 'analytics'}
				class:text-white={activeTab === 'analytics'}
				class:text-gray-500={activeTab !== 'analytics'}
				onclick={() => activeTab = 'analytics'}
			>Analytics</button>
		</div>
	</div>

	{#if isLoading}
		<!-- Loading Skeleton -->
		<div class="space-y-6">
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				<PulseLoader variant="card" count={4} />
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<PulseLoader variant="card" count={2} />
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
				<PulseLoader variant="card" count={3} />
			</div>
		</div>
	{:else if errorMessage}
		<!-- Error State -->
		<div class="bg-red-50 border border-red-200 rounded-xl p-8 text-center">
			<svg class="w-12 h-12 text-red-400 mx-auto mb-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
			</svg>
			<p class="text-red-700 text-sm font-medium">{errorMessage}</p>
			<button onclick={loadDashboard} class="mt-4 text-sm text-white bg-red-600 hover:bg-red-700 px-4 py-2 rounded-lg font-medium transition cursor-pointer">Coba lagi</button>
		</div>
	{:else}
		<!-- ── Stat Cards (Desktop) ── -->
		<div class="hidden md:grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
			{#each stats as stat}
				<div class="bg-white border border-gray-200 rounded-xl p-5 hover:shadow-lg hover:border-gray-300 transition-all duration-200 group">
					<div class="flex items-center justify-between">
						<div class="flex-1 min-w-0">
							<div class="text-xs font-semibold text-gray-400 tracking-wider uppercase">{stat.label}</div>
							<div class="text-2xl font-bold text-gray-900 mt-1.5 tabular-nums">{stat.value}</div>
							<div class="text-xs mt-1 font-medium {stat.changeClass}">{stat.change}</div>
						</div>
						<div class="w-12 h-12 rounded-xl flex items-center justify-center shrink-0 {stat.iconBg} group-hover:scale-110 transition-transform duration-200">
							<svg class="w-6 h-6 {stat.iconColor}" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d={stat.icon} />
							</svg>
						</div>
					</div>
				</div>
			{/each}
		</div>

		<!-- ── Mobile: Greeting + Quick Actions (Talenta Style) ── -->
		<div class="md:hidden mb-4">
			<!-- Greeting -->
			<div class="flex items-center justify-between mb-4">
				<div>
					<h2 class="text-lg font-bold text-gray-900 dark:text-white">Hai, {userName}! 👋</h2>
					<p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{todayDate}</p>
				</div>
				<button
					onclick={() => goto('/absensi')}
					class="w-10 h-10 bg-[#1A56DB] text-white rounded-xl flex items-center justify-center shadow-sm shadow-blue-200 dark:shadow-blue-900/30 active:scale-90 transition-all duration-150 cursor-pointer"
					aria-label="Absensi"
				>
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
					</svg>
				</button>
			</div>

			<!-- Quick Actions -->
			<div class="grid grid-cols-4 gap-2 mb-4">
				<button onclick={() => goto('/absensi')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-blue-50 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Absensi</span>
				</button>
				<button onclick={() => goto('/cuti')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-emerald-50 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Cuti</span>
				</button>
				<button onclick={() => goto('/lembur')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-amber-50 dark:bg-amber-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Lembur</span>
				</button>
				<button onclick={() => goto('/persetujuan')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-purple-50 dark:bg-purple-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Setujui</span>
				</button>
			</div>

			<!-- Stat Cards (Mobile Premium) -->
			<div class="grid grid-cols-2 gap-2.5">
				{#each stats as stat}
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm border border-gray-200 dark:border-gray-800 rounded-xl p-3.5 active:scale-[0.95] transition-all duration-150 shadow-sm hover:shadow-md">
						<div class="flex items-start gap-2.5">
							<div class="w-9 h-9 rounded-lg flex items-center justify-center shrink-0 {stat.iconBg} dark:bg-opacity-20">
								<svg class="w-5 h-5 {stat.iconColor}" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d={stat.icon} />
								</svg>
							</div>
							<div class="min-w-0 flex-1">
								<div class="text-[10px] font-semibold text-gray-400 dark:text-gray-500 tracking-wider uppercase truncate">{stat.label}</div>
								<div class="text-lg font-bold text-gray-900 dark:text-white tabular-nums mt-0.5">{stat.value}</div>
								<div class="text-[11px] mt-0.5 font-medium {stat.changeClass} truncate">{stat.change}</div>
							</div>
						</div>
					</div>
				{/each}
			</div>
		</div>

		<!-- OVERVIEW TAB -->
		<div class:hidden={activeTab !== 'overview'}>
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-4 mb-6">
				<div class="bg-white border border-gray-200 rounded-xl p-5 lg:col-span-2">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Kehadiran 7 Hari Terakhir</h3>
						<span class="text-xs text-gray-400">Jumlah karyawan hadir per hari</span>
					</div>
					<div class="h-52"><canvas bind:this={overviewAttendanceCanvas}></canvas></div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<h3 class="text-sm font-semibold text-gray-900 mb-4">Komposisi Karyawan</h3>
					<div class="h-52 flex items-center justify-center"><canvas bind:this={overviewCompositionCanvas}></canvas></div>
				</div>
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Tren Rekrutmen</h3>
						<span class="text-xs text-gray-400">6 bulan terakhir</span>
					</div>
					<div class="h-44"><canvas bind:this={overviewTrendCanvas}></canvas></div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<h3 class="text-sm font-semibold text-gray-900 mb-4">Komposisi Gender</h3>
					<div class="h-44 flex items-center justify-center"><canvas bind:this={overviewGenderCanvas}></canvas></div>
				</div>
			</div>
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Pending Approval</h3>
						<span class="text-xs text-gray-400">Menunggu persetujuan</span>
					</div>
					<div class="space-y-1">
						{#each pendingApprovals as item}
							<div class="flex items-center justify-between py-2.5 px-2 rounded-lg hover:bg-gray-50 transition">
								<div class="flex items-center gap-3">
									<div class="w-9 h-9 {item.bg} rounded-full flex items-center justify-center text-xs font-semibold {item.text}">{item.initials}</div>
									<div>
										<div class="text-sm font-medium text-gray-900"><span class="mr-1">{getTypeIcon(item.type)}</span>{item.type}</div>
										<div class="text-xs text-gray-400">{item.name} &mdash; {item.detail}</div>
									</div>
								</div>
								<span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium {item.initials === '✓' ? 'bg-green-50 text-green-700 border border-green-200' : 'bg-amber-50 text-amber-700 border border-amber-200'}">{item.initials === '✓' ? 'Selesai' : 'Menunggu'}</span>
							</div>
						{/each}
					</div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Karyawan Baru</h3>
						<span class="text-xs text-gray-400">Terbaru</span>
					</div>
					<div class="space-y-1">
						{#each newEmployees as item}
							<div class="flex items-center justify-between py-2.5 px-2 rounded-lg hover:bg-gray-50 transition cursor-pointer">
								<div class="flex items-center gap-3">
									<div class="w-9 h-9 bg-gradient-to-br from-gray-200 to-gray-300 rounded-full flex items-center justify-center text-xs font-semibold text-gray-600">{item.initials}</div>
									<div>
										<div class="text-sm font-medium text-gray-900">{item.name}</div>
										<div class="text-xs text-gray-400">{item.position}{item.date ? ` &mdash; Bergabung ${item.date}` : ''}</div>
									</div>
								</div>
								{#if item.status}<span class="inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium bg-green-50 text-green-700 border border-green-200">{item.status}</span>{/if}
							</div>
						{/each}
					</div>
				</div>
			</div>
		</div>

		<!-- DEPARTMENTS TAB -->
		<div class:hidden={activeTab !== 'departments'}>
			<div class="bg-white border border-gray-200 rounded-xl p-5">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-semibold text-gray-900">Distribusi Karyawan per Departemen</h3>
					<span class="text-xs text-gray-400">Top 10 departemen</span>
				</div>
				<div class="h-80"><canvas bind:this={departmentsDeptCanvas}></canvas></div>
			</div>
		</div>

		<!-- ANALYTICS TAB -->
		<div class:hidden={activeTab !== 'analytics'}>
			<div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mb-6">
				<div class="bg-white border border-gray-200 rounded-xl p-5 lg:col-span-2">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Tren Rekrutmen Bulanan</h3>
						<span class="text-xs text-gray-400">6 bulan terakhir</span>
					</div>
					<div class="h-64"><canvas bind:this={analyticsTrendCanvas}></canvas></div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<h3 class="text-sm font-semibold text-gray-900 mb-4">Status Karyawan</h3>
					<div class="h-56 flex items-center justify-center"><canvas bind:this={analyticsCompositionCanvas}></canvas></div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<h3 class="text-sm font-semibold text-gray-900 mb-4">Gender</h3>
					<div class="h-56 flex items-center justify-center"><canvas bind:this={analyticsGenderCanvas}></canvas></div>
				</div>
			</div>
			<div class="bg-white border border-gray-200 rounded-xl p-5">
				<div class="flex items-center justify-between mb-4">
					<h3 class="text-sm font-semibold text-gray-900">Distribusi Karyawan per Departemen</h3>
					<span class="text-xs text-gray-400">Top 10</span>
				</div>
				<div class="h-72"><canvas bind:this={analyticsDeptCanvas}></canvas></div>
			</div>
		</div>						<p class="text-xs text-gray-400 text-center mt-8 pb-4 hidden md:block">HRMS &mdash; Sistem Informasi Sumber Daya Manusia &copy; 2026</p>
	{/if}
	{/if}
</div>
