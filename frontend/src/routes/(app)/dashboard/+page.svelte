<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { dashboard as dashboardApi, auth, announcements as announcementsApi, attendance, leaveRequests, approvals as approvalsApi } from '$lib/api.js';
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

	const ENTITY_LABELS: Record<string, string> = {
		leave: 'Cuti',
		overtime: 'Lembur',
		reimbursement: 'Reimbursement',
		shift_change: 'Permintaan Shift',
		loan: 'Pinjaman',
		mutation: 'Mutasi',
		manual_attendance: 'Absensi Manual',
		resign: 'Resign',
	};

	const ENTITY_COLORS: Record<string, string> = {
		leave: 'bg-blue-100 text-blue-600',
		overtime: 'bg-amber-100 text-amber-600',
		reimbursement: 'bg-emerald-100 text-emerald-600',
		shift_change: 'bg-purple-100 text-purple-600',
		loan: 'bg-rose-100 text-rose-600',
		mutation: 'bg-indigo-100 text-indigo-600',
		manual_attendance: 'bg-cyan-100 text-cyan-600',
		resign: 'bg-orange-100 text-orange-600',
	};



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
		absent_today: AbsentEmployee[];
	}

	interface AbsentEmployee {
		employee_id: string;
		full_name: string;
		department_name: string;
		absence_reason: string;
		leave_reason: string;
	}

	// ── User / Greeting ──
	let userName = $derived((auth.getUser() as any)?.full_name?.split(' ')[0] || 'Pengguna');
	let todayDate = $derived(new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' }));

	// ── State ──
	let stats = $state<Stat[]>([]);
	let realPendingApprovals = $state<any[]>([]);
	let absentToday = $state<AbsentEmployee[]>([]);
	let onLeaveToday = $derived(absentToday.filter(e => e.absence_reason !== 'tanpa_keterangan'));
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
					const [annRes, attRes, leaveRes, dashRes] = await Promise.all([
						announcementsApi.list(1, 50, '').catch(() => ({ data: [] })),
						attendance.getTodayStatus().catch(() => ({ data: null })),
						leaveRequests.getMyBalances().catch(() => ({ data: [] })),
						dashboardApi.get().catch(() => ({ data: { absent_today: [] } }))
					]);
					
					employeeAnnouncements = annRes.data || [];
					todayStatus = attRes.data;
					
					if (leaveRes.data && Array.isArray(leaveRes.data)) {
						leaveBalance = leaveRes.data.find((b: any) => b.leave_type_name === 'Tahunan' || b.leave_type_name === 'Cuti Tahunan') || leaveRes.data[0];
					}

					if (dashRes?.data?.absent_today) {
						absentToday = (dashRes.data.absent_today as AbsentEmployee[]) || [];
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

			// Fetch real pending approvals
			try {
				const pendingRes = await approvalsApi.getPending();
				if (pendingRes.success && pendingRes.data) {
					const d = pendingRes.data as { items: any[]; total: number };
					realPendingApprovals = (d.items || []).sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());
				} else {
					realPendingApprovals = [];
				}
			} catch {
				realPendingApprovals = [];
			}



			absentToday = data.absent_today || [];

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

<!-- eslint-disable svelte/no-navigation-without-resolve -->

<div class="max-w-full mx-auto">
	{#if isEmployee && !isLoading}
		<!-- Employee Dashboard (Talenta Style) -->
		<div class="w-full mx-auto py-4 px-4 sm:px-6 lg:px-8">
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
			<!-- Announcements Feed (Redesigned) -->
					<div class="w-full">
						<div class="bg-white dark:bg-gray-900 rounded-xl shadow-sm border border-gray-200 dark:border-gray-800 overflow-hidden flex flex-col">
							<div class="p-5 border-b border-gray-100 dark:border-gray-800 flex items-center justify-between">
								<div class="flex items-center gap-2">
									<svg class="w-5 h-5 text-blue-600 dark:text-blue-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.34 15.84c-.688-.06-1.386-.09-2.09-.09H7.5a4.5 4.5 0 1 1 0-9h.75c.704 0 1.402-.03 2.09-.09m0 9.18c.253.962.584 1.892.985 2.783.247.55.06 1.21-.463 1.511l-.657.38a.468.468 0 0 1-.625-.2 12.114 12.114 0 0 1-1.288-4.754m0-9.18A12.114 12.114 0 0 1 10.34 3.1a.468.468 0 0 0-.625-.2l-.657.38c-.523.3-.71.96-.463 1.51.4.892.731 1.821.985 2.783m0 0a11.16 11.16 0 0 1 3.846 2.561m0-1.335c.674.682 1.26 1.448 1.74 2.28m0 0a11.161 11.161 0 0 1-.532 5.326m.532-5.326 4.655-1.553a1.125 1.125 0 0 1 1.373.706l.383 1.194a1.125 1.125 0 0 1-.655 1.366l-2.969 1.091m0 0L16.5 20.25l-2.276-4.848" /></svg>
									<h2 class="font-semibold text-gray-900 dark:text-gray-100 text-base">Pengumuman</h2>
									{#if employeeAnnouncements.filter(a => a.is_pinned).length > 0}
										<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-100 text-amber-700">{employeeAnnouncements.filter(a => a.is_pinned).length} disematkan</span>
									{/if}
								</div>
								<a href="/pengumuman" class="text-sm font-medium text-blue-600 hover:text-blue-700 transition-colors inline-flex items-center gap-1">
									Lihat Semua
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" /></svg>
								</a>
							</div>
							
							<div class="p-0 h-[calc(100vh-280px)] min-h-[300px] overflow-y-auto custom-scrollbar">
								{#if employeeAnnouncements.length === 0}
									<div class="p-8 text-center">
										<div class="w-14 h-14 mx-auto mb-3 rounded-full bg-gray-50 flex items-center justify-center">
											<svg class="w-7 h-7 text-gray-300" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.34 15.84c-.688-.06-1.386-.09-2.09-.09H7.5a4.5 4.5 0 1 1 0-9h.75c.704 0 1.402-.03 2.09-.09m0 9.18c.253.962.584 1.892.985 2.783.247.55.06 1.21-.463 1.511l-.657.38a.468.468 0 0 1-.625-.2 12.114 12.114 0 0 1-1.288-4.754m0-9.18A12.114 12.114 0 0 1 10.34 3.1a.468.468 0 0 0-.625-.2l-.657.38c-.523.3-.71.96-.463 1.51.4.892.731 1.821.985 2.783m0 0a11.16 11.16 0 0 1 3.846 2.561" /></svg>
										</div>
										<p class="text-gray-400 text-sm font-medium">Tidak ada pengumuman</p>
										<p class="text-xs text-gray-300 mt-1">Belum ada pengumuman untuk ditampilkan</p>
									</div>
								{:else}
									{@const sortedAnnouncements = [...employeeAnnouncements].sort((a, b) => {
										// Pinned first, then by date descending
										if (a.is_pinned && !b.is_pinned) return -1;
										if (!a.is_pinned && b.is_pinned) return 1;
										return new Date(b.published_at || b.created_at).getTime() - new Date(a.published_at || a.created_at).getTime();
									})}
									<div class="divide-y divide-gray-50 dark:divide-gray-800/50">
										{#each sortedAnnouncements as ann (ann.id)}
											{@const isPinned = ann.is_pinned}
											{@const typeConfig = ann.announcement_type === 'important' ? { gradient: 'from-orange-50 to-amber-50/30 dark:from-orange-950/30 dark:to-amber-900/10', border: 'border-l-orange-400 dark:border-l-orange-500', badgeBg: 'bg-orange-100 dark:bg-orange-900/40', badgeText: 'text-orange-700 dark:text-orange-400', dotColor: 'bg-orange-400', label: 'Penting' } : ann.announcement_type === 'emergency' ? { gradient: 'from-red-50 to-rose-50/30 dark:from-red-950/30 dark:to-rose-900/10', border: 'border-l-red-400 dark:border-l-red-500', badgeBg: 'bg-red-100 dark:bg-red-900/40', badgeText: 'text-red-700 dark:text-red-400', dotColor: 'bg-red-400', label: 'Darurat' } : { gradient: 'from-blue-50 to-sky-50/30 dark:from-blue-950/30 dark:to-sky-900/10', border: 'border-l-blue-400 dark:border-l-blue-500', badgeBg: 'bg-blue-100 dark:bg-blue-900/40', badgeText: 'text-blue-700 dark:text-blue-400', dotColor: 'bg-blue-400', label: 'Informasi' }}
											<a href="/pengumuman" class="block {isPinned ? 'bg-gradient-to-r ' + typeConfig.gradient + ' border-l-4 ' + typeConfig.border : 'hover:bg-gray-50 dark:hover:bg-gray-800/50 border-l-4 border-l-transparent'} transition-all duration-200 group relative">
												{#if isPinned}
													<div class="absolute top-3 right-3 inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-bold bg-amber-100 dark:bg-amber-900/40 text-amber-700 dark:text-amber-400 shadow-sm">
														<svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24"><path d="M22.314 10.172l-1.678-1.678A2.988 2.988 0 0 0 18.52 7.5H17.5L14.5.5h-1.063L12 1.563 10.563.5H9.5l-3 7H5.48a2.988 2.988 0 0 0-2.116.994L1.686 10.172a3 3 0 0 0-.5 3.216L3.073 17.5a3 3 0 0 0 2.686 1.5H9.5l1.5 4.5h2l1.5-4.5h3.741a3 3 0 0 0 2.686-1.5l1.887-4.112a3 3 0 0 0-.5-3.216z"/></svg>
														Disematkan
													</div>
												{/if}
												<div class="p-5 {isPinned ? 'pt-3' : ''}">
													<div class="flex gap-3.5">
														<div class="shrink-0 mt-1">
															<div class="w-10 h-10 rounded-xl flex items-center justify-center {typeConfig.badgeBg} {typeConfig.badgeText} ring-1 ring-inset ring-black/5 group-hover:scale-110 transition-transform duration-200">
																<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M10.34 15.84c-.688-.06-1.386-.09-2.09-.09H7.5a4.5 4.5 0 1 1 0-9h.75c.704 0 1.402-.03 2.09-.09m0 9.18c.253.962.584 1.892.985 2.783.247.55.06 1.21-.463 1.511l-.657.38a.468.468 0 0 1-.625-.2 12.114 12.114 0 0 1-1.288-4.754m0-9.18A12.114 12.114 0 0 1 10.34 3.1a.468.468 0 0 0-.625-.2l-.657.38c-.523.3-.71.96-.463 1.51.4.892.731 1.821.985 2.783m0 0a11.16 11.16 0 0 1 3.846 2.561m0-1.335c.674.682 1.26 1.448 1.74 2.28m0 0a11.161 11.161 0 0 1-.532 5.326m.532-5.326 4.655-1.553a1.125 1.125 0 0 1 1.373.706l.383 1.194a1.125 1.125 0 0 1-.655 1.366l-2.969 1.091m0 0L16.5 20.25l-2.276-4.848" /></svg>
															</div>
														</div>
														<div class="min-w-0 flex-1">
															<div class="flex items-center gap-2 mb-1">
																<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold {typeConfig.badgeBg} {typeConfig.badgeText}">{typeConfig.label}</span>
																<span class="inline-flex items-center gap-1 text-[10px] text-gray-400 dark:text-gray-500">
																	<span class="w-1 h-1 rounded-full bg-gray-300 dark:bg-gray-600"></span>
																	{new Date(ann.published_at || ann.created_at).toLocaleDateString('id-ID', {day: 'numeric', month: 'short', year: 'numeric'})}
																</span>
															</div>
															<h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 leading-snug">{ann.title}</h3>
															<p class="text-sm text-gray-500 dark:text-gray-400 mt-1.5 line-clamp-2 leading-relaxed">{ann.content || 'Klik untuk melihat detail pengumuman ini.'}</p>
															{#if ann.created_by_name}
																<div class="flex items-center gap-1.5 mt-2.5 text-xs text-gray-400 dark:text-gray-500">
																	<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
																	{ann.created_by_name || '-'}
																</div>
															{/if}
														</div>
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
									onclick={async () => await goto('/absensi')}
										class="flex-1 py-3 px-4 rounded-lg font-bold text-white text-sm transition-colors text-center {todayStatus.has_checked_in ? 'bg-gray-100 text-gray-400 cursor-not-allowed' : 'bg-blue-600 hover:bg-blue-700'}"
										disabled={todayStatus.has_checked_in}
									>
										Check In
									</button>
								<button 
									onclick={async () => await goto('/absensi')}
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
					
					<!-- Absent Today (Employee: show only those on leave, with empty state) -->
					<div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden">
						<div class="p-5 border-b border-gray-100 flex justify-between items-center bg-gradient-to-r from-blue-50 to-indigo-50/50">
							<h2 class="font-semibold text-gray-900 flex items-center gap-2 text-sm">
								<svg class="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z" /></svg>
								Karyawan Tidak Hadir
							</h2>
							{#if onLeaveToday.length > 0}
								<span class="inline-flex items-center gap-1.5 px-2 py-1 rounded-full text-xs font-medium bg-blue-50 text-blue-700 border border-blue-200">
									{onLeaveToday.length} orang
								</span>
							{/if}
						</div>
						<div class="p-4 space-y-2">
							{#if onLeaveToday.length > 0}
								{#each onLeaveToday as emp (emp.employee_id)}
									<div class="flex items-center gap-3 py-2">
										<div class="w-9 h-9 rounded-full bg-blue-50 flex items-center justify-center text-xs font-bold text-blue-600 shrink-0">
											{emp.full_name?.charAt(0) || '?'}
										</div>
										<div class="min-w-0 flex-1">
											<div class="text-sm font-medium text-gray-900 line-clamp-1">{emp.full_name}</div>
											<div class="flex items-center gap-1.5 mt-0.5">
												<span class="text-xs text-gray-500">{emp.department_name}</span>
												<span class="text-gray-300">·</span>
												<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium bg-blue-50 text-blue-700 border border-blue-200">{emp.absence_reason}</span>
											</div>
										</div>
									</div>
								{/each}
							{:else}
								<div class="flex items-center gap-3 py-4 text-center justify-center">
									<div class="w-9 h-9 rounded-full bg-emerald-50 flex items-center justify-center shrink-0">
										<svg class="w-4 h-4 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
									</div>
									<p class="text-sm text-gray-500">Tidak ada yang cuti/sakit hari ini</p>
								</div>
							{/if}
						</div>
					</div>

					<!-- Shortcuts -->
					<div class="bg-white rounded-xl shadow-sm border border-gray-200 overflow-hidden p-5">
						<h2 class="font-semibold text-gray-900 mb-4">Akses Cepat</h2>
						<div class="grid grid-cols-3 sm:grid-cols-6 gap-2 sm:gap-4">

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

							<a href="/penggajian/slip-saya" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-teal-50 text-teal-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m3.75 9v6m3-3H9m1.5-12H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Slip Gaji</span>
							</a>

							<a href="/dokumen" class="flex flex-col items-center gap-2 group p-2 rounded-lg hover:bg-gray-50 transition-colors">
								<div class="w-12 h-12 rounded-full bg-sky-50 text-sky-600 flex items-center justify-center group-hover:scale-110 transition-transform">
									<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 12.75V12A2.25 2.25 0 0 1 4.5 9.75h15A2.25 2.25 0 0 1 21.75 12v.75m-8.69-6.44-2.12-2.12a1.5 1.5 0 0 0-1.061-.44H4.5A2.25 2.25 0 0 0 2.25 6v12a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9a2.25 2.25 0 0 0-2.25-2.25h-5.379a1.5 1.5 0 0 1-1.06-.44Z" /></svg>
								</div>
								<span class="text-xs font-medium text-gray-600 text-center">Dokumen</span>
							</a>
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
			{#each stats as stat (stat)}
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
					onclick={async () => await goto('/absensi')}
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

				<button onclick={async () => await goto('/absensi')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-blue-50 dark:bg-blue-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Absensi</span>
				</button>

				<button onclick={async () => await goto('/cuti')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-emerald-50 dark:bg-emerald-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-emerald-600 dark:text-emerald-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Cuti</span>
				</button>

				<button onclick={async () => await goto('/lembur')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-amber-50 dark:bg-amber-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Lembur</span>
				</button>

				<button onclick={async () => await goto('/persetujuan')} class="flex flex-col items-center gap-1.5 p-3 bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl active:scale-90 transition-all duration-150 cursor-pointer">
					<div class="w-10 h-10 bg-purple-50 dark:bg-purple-900/30 rounded-xl flex items-center justify-center">
						<svg class="w-5 h-5 text-purple-600 dark:text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
					</div>
					<span class="text-[10px] font-medium text-gray-600 dark:text-gray-400">Setujui</span>
				</button>
			</div>

			<!-- Stat Cards (Mobile Premium) -->
			<div class="grid grid-cols-2 gap-2.5">
				{#each stats as stat (stat)}
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
			<!-- Attendance Summary Card -->
			{#if stats.length > 0}
				{@const hadirStat = stats[1]}
				<div class="bg-gradient-to-br from-blue-50 to-indigo-50/50 dark:from-blue-950/20 dark:to-indigo-950/10 border border-blue-100 dark:border-blue-900/50 rounded-xl p-5 mb-6">
					<div class="flex items-center justify-between mb-4">
						<div class="flex items-center gap-2.5">
							<div class="w-10 h-10 rounded-xl bg-blue-100 dark:bg-blue-900/40 flex items-center justify-center">
								<svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
							</div>
							<div>
								<h3 class="text-sm font-semibold text-blue-900 dark:text-blue-200">Ringkasan Kehadiran</h3>
								<p class="text-xs text-blue-600/70 dark:text-blue-400/70">Data real-time hari ini</p>
							</div>
						</div>
						<div class="text-right">
							<div class="text-2xl font-bold text-blue-900 dark:text-blue-200">{hadirStat.value}</div>
							<div class="text-xs font-medium text-emerald-600">{hadirStat.change}</div>
						</div>
					</div>
					<div class="grid grid-cols-3 gap-3">
						<div class="bg-white/80 dark:bg-gray-900/50 rounded-xl p-3 text-center">
							<p class="text-lg font-bold text-emerald-600 dark:text-emerald-400 tabular-nums">{stats[0]?.value || '0'}</p>
							<p class="text-[10px] font-medium text-gray-500 dark:text-gray-400 mt-0.5">Total Karyawan</p>
						</div>
						<div class="bg-white/80 dark:bg-gray-900/50 rounded-xl p-3 text-center">
							<p class="text-lg font-bold text-blue-600 dark:text-blue-400 tabular-nums">{hadirStat.value}</p>
							<p class="text-[10px] font-medium text-gray-500 dark:text-gray-400 mt-0.5">Hadir Hari Ini</p>
						</div>
						<div class="bg-white/80 dark:bg-gray-900/50 rounded-xl p-3 text-center">
							<p class="text-lg font-bold text-amber-600 dark:text-amber-400 tabular-nums">{stats[2]?.value || '0'}</p>
							<p class="text-[10px] font-medium text-gray-500 dark:text-gray-400 mt-0.5">Pending</p>
						</div>
					</div>
					<!-- Attendance rate progress bar -->
					{#if true}
						{@const rate = parseInt(hadirStat.change) || 0}
						<div class="mt-4 pt-4 border-t border-blue-100 dark:border-blue-900/50">
							<div class="flex items-center justify-between text-xs mb-1.5">
								<span class="font-medium text-blue-700 dark:text-blue-300">Tingkat Kehadiran</span>
								<span class="font-bold text-blue-700 dark:text-blue-300">{rate}%</span>
							</div>
							<div class="w-full h-2.5 bg-blue-100 dark:bg-blue-900/50 rounded-full overflow-hidden">
								<div class="h-full bg-gradient-to-r from-blue-500 to-emerald-500 rounded-full transition-all duration-700 ease-out" style="width: {Math.min(rate, 100)}%"></div>
							</div>
						</div>
					{/if}
				</div>
			{/if}

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
			<div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
				<!-- Absent Today -->
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Absen Hari Ini</h3>
						{#if absentToday.length > 0}
							<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-red-50 text-red-700 border border-red-200">
								<span class="w-1.5 h-1.5 bg-red-500 rounded-full animate-pulse"></span>
								{absentToday.length} tidak hadir
							</span>
						{:else}
							<span class="text-xs text-gray-400">Karyawan tidak hadir</span>
						{/if}
					</div>
					<div class="space-y-1 max-h-[320px] overflow-y-auto">
						{#if absentToday.length === 0}
							<div class="text-center py-6">
								<div class="w-10 h-10 mx-auto mb-2 rounded-full bg-green-50 flex items-center justify-center">
									<svg class="w-5 h-5 text-green-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								</div>
								<p class="text-xs text-gray-400">Semua karyawan hadir hari ini</p>
							</div>
						{:else}
							{#each absentToday as emp (emp.employee_id)}
								{@const reasonColor = emp.absence_reason !== 'tanpa_keterangan' ? 'bg-blue-50 text-blue-700 border border-blue-200' : 'bg-red-50 text-red-700 border border-red-200'}
								{@const reasonLabel = emp.absence_reason !== 'tanpa_keterangan' ? emp.absence_reason : 'Tanpa Ket.'}
								<div class="flex items-start gap-2.5 py-2.5 px-2 rounded-lg hover:bg-gray-50 transition cursor-pointer" onclick={() => goto('/karyawan/' + emp.employee_id)}>
									<div class="w-8 h-8 rounded-full bg-gray-100 flex items-center justify-center text-[10px] font-bold text-gray-600 shrink-0 mt-0.5">
										{emp.full_name?.charAt(0) || '?'}
									</div>
									<div class="min-w-0 flex-1">
										<div class="text-xs font-semibold text-gray-900 line-clamp-1">{emp.full_name}</div>
										<div class="flex items-center gap-1.5 mt-0.5">
											<span class="text-[10px] text-gray-500">{emp.department_name || '-'}</span>
										</div>
										<div class="mt-1">
											<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-medium {reasonColor}">{reasonLabel}</span>
										</div>
									</div>
								</div>
							{/each}
						{/if}
					</div>
				</div>

				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Pending Approval</h3>
						{#if realPendingApprovals.length > 0}
							<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-amber-50 text-amber-700 border border-amber-200">
								<span class="w-1.5 h-1.5 bg-amber-500 rounded-full animate-pulse"></span>
								{realPendingApprovals.length} menunggu
							</span>
						{:else}
							<span class="text-xs text-gray-400">Menunggu persetujuan</span>
						{/if}
					</div>
					<div class="space-y-1 max-h-[320px] overflow-y-auto custom-scrollbar">
						{#if realPendingApprovals.length === 0}
							<div class="text-center py-6">
								<div class="w-10 h-10 mx-auto mb-2 rounded-full bg-emerald-50 flex items-center justify-center">
									<svg class="w-5 h-5 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
								</div>
								<p class="text-xs text-gray-400">Tidak ada pending approval</p>
							</div>
						{:else}
							{#each realPendingApprovals as item (item.tracking_id || item.entity_id)}
								{@const entityColor = ENTITY_COLORS[item.entity_type] || 'bg-gray-100 text-gray-600'}
								{@const entityLabel = ENTITY_LABELS[item.entity_type] || item.entity_type}
											<div class="flex items-start justify-between py-2.5 px-2 rounded-lg hover:bg-gray-50 transition cursor-pointer active:scale-[0.99]" onclick={() => goto('/persetujuan')}>
									<div class="flex items-start gap-2.5 min-w-0 flex-1">
										<div class="w-8 h-8 rounded-lg flex items-center justify-center text-[10px] font-bold shrink-0 mt-0.5 {entityColor}">
											{entityLabel.substring(0, 2).toUpperCase()}
										</div>
										<div class="min-w-0 flex-1">
											<div class="flex items-center gap-1.5 flex-wrap">
												<span class="text-xs font-semibold text-gray-900 line-clamp-1">{item.title || entityLabel}</span>
												<span class="text-[10px] px-1.5 py-0.5 rounded font-medium bg-gray-100 text-gray-600">{entityLabel}</span>
											</div>
											<div class="flex items-center gap-1.5 text-[11px] text-gray-500 mt-1">
												<span class="font-medium text-gray-700">{item.requestor_name || '-'}</span>
												<span>&middot;</span>
												<span>{new Date(item.created_at).toLocaleDateString('id-ID', {day:'numeric', month:'short'})}</span>
											</div>
											<div class="flex items-center gap-1.5 mt-1.5">
												<div class="flex-1 h-1.5 bg-gray-100 rounded-full overflow-hidden flex gap-0.5 max-w-[80px]">
													{#each Array(item.total_steps) as _, i (i)}
														<div class="h-full flex-1 rounded-full {i < item.current_step ? 'bg-blue-500' : 'bg-gray-200'}"></div>
													{/each}
												</div>
												<span class="text-[10px] text-gray-400 font-medium">{item.current_step}/{item.total_steps}</span>
											</div>
										</div>
									</div>
									<svg class="w-4 h-4 text-gray-300 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" /></svg>
								</div>
							{/each}
						{/if}
					</div>
				</div>
				<div class="bg-white border border-gray-200 rounded-xl p-5">
					<div class="flex items-center justify-between mb-4">
						<h3 class="text-sm font-semibold text-gray-900">Karyawan Baru</h3>
						<span class="text-xs text-gray-400">Terbaru</span>
					</div>
					<div class="space-y-1">
						{#each newEmployees as item (item)}
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
