<script lang="ts">
	import { employees as employeesApi, salaryComponents as scApi, workSchedules as wsApi, company as companyApi } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';

	let { employeeId, onclose }: { employeeId: string; onclose: () => void } = $props();

	// --------------- Salary editing state ---------------
	let salaryBase = $state('');
	let salaryDaily = $state('');
	let isSavingSalary = $state(false);

	async function handleSaveSalary() {
		const emp = employee;
		if (!emp) return;
		isSavingSalary = true;
		try {
			const payload: Record<string, any> = {};
			if (salaryBase) payload.base_salary = Number(salaryBase);
			if (salaryDaily) payload.daily_wage = Number(salaryDaily);
			await employeesApi.update(emp.id, payload);
			// Refresh data
			const empRes: any = await employeesApi.get(emp.id);
			if (empRes.data) {
				const empData = empRes.data as EmployeeDetail;
				employee = empData;
				salaryBase = String(empData.base_salary || '');
				salaryDaily = String(empData.daily_wage || '');
			}
		} catch (err: any) {
			scFormError = err.message || 'Gagal menyimpan gaji';
		} finally {
			isSavingSalary = false;
		}
	}

	$effect(() => {
		const emp = employee;
		if (emp && !isLoading) {
			salaryBase = String(emp.base_salary || '');
			salaryDaily = String(emp.daily_wage || '');
		}
	});

	// --------------- Salary Component Types ---------------
	type SalaryComp = {
		id: string;
		employee_id: string;
		component_name: string;
		component_type: 'allowance' | 'deduction';
		amount: number;
		is_active: boolean;
		effective_date: string;
		created_at: string;
	};

	type SalaryCompForm = {
		component_name: string;
		component_type: 'allowance' | 'deduction';
		amount: number;
		effective_date: string;
	};

	let salaryComps = $state<SalaryComp[]>([]);
	let scTotal = $state(0);
	let scPage = $state(1);
	let scPerPage = $state(25);

	// --------------- Schedule Override State ---------------
	let allWorkSchedules = $state<{ id: string; name: string; schedule_type: string; }[]>([]);
	let schedulePickerOpen = $state(false);
	let selectedScheduleId = $state('');
	let isSavingSchedule = $state(false);
	let scTotalPages = $state(0);
	let isLoadingSC = $state(false);
	let scError = $state('');

	// Form state
	let showSCForm = $state(false);
	let scFormTitle = $state('');
	let editingSCId = $state<string | null>(null);
	let scForm = $state<SalaryCompForm>({ component_name: '', component_type: 'allowance', amount: 0, effective_date: '' });
	let scFormError = $state('');
	let isSavingSC = $state(false);

	// Delete confirm
	let showSCDelete = $state(false);
	let deletingSCId = $state<string | null>(null);
	let deletingSCName = $state('');

	// Photo upload
	let isUploadingPhoto = $state(false);
	let photoPreview = $state<string | null>(null);

	async function handlePhotoUpload(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		const emp = employee;
		if (!file || !emp) return;

		const reader = new FileReader();
		reader.onload = (ev) => { photoPreview = ev.target?.result as string; };
		reader.readAsDataURL(file);

		isUploadingPhoto = true;
		try {
			const res: any = await employeesApi.uploadPhoto(emp.id, file);
			employee!.photo_url = res.data?.photo_url || '';
			photoPreview = null;
		} catch (err: any) {
			errorMessage = err.message || 'Gagal upload foto';
			photoPreview = null;
		} finally {
			isUploadingPhoto = false;
			target.value = '';
		}
	}

	type HistoryItem = {
		id: string;
		employee_id: string;
		change_type: string;
		old_value: Record<string, any> | null;
		new_value: Record<string, any> | null;
		reason: string;
		changed_by_name: string;
		changed_at: string;
	};

	type EmployeeDetail = {
		id: string;
		employee_id: string;
		full_name: string;
		email: string;
		gender: string;
		place_of_birth: string;
		date_of_birth: string | null;
		religion: string;
		marital_status: string;
		join_date: string;
		employment_status: string;
		is_active: boolean;
		role_slug: string;
		role_name: string;
		position_name: string;
		department_id: string | null;
		department_name: string;
		phone: string;
		address: string;
		photo_url: string;
		nik: string;
		npwp: string;
		bank_name: string;
		bank_account: string;
		address_ktp: string;
		base_salary: number;
		daily_wage: number;
		work_schedule_id: string | null;
		work_schedule_name: string;
		last_login_at: string | null;
		is_locked: boolean;
		locked_until: string | null;
		created_at: string;
		updated_at: string;
	};

	let employee = $state<EmployeeDetail | null>(null);
	let isLoading = $state(true);
	let errorMessage = $state('');

	let history = $state<HistoryItem[]>([]);
	let historyTotal = $state(0);
	let historyPage = $state(1);
	let historyPerPage = $state(10);
	let historyTotalPages = $state(0);
	let isLoadingHistory = $state(false);

	async function loadEmployee(id: string) {
		isLoading = true;
		errorMessage = '';
		try {
			if (!id) throw new Error('ID karyawan tidak ditemukan');
			const empRes: any = await employeesApi.get(id);
			employee = empRes.data || null;

			if (employee) {
				const histRes: any = await employeesApi.getHistory(employee.id, { page: historyPage, per_page: historyPerPage });
				history = histRes.data || [];
				historyTotal = histRes.meta?.total || 0;
				historyTotalPages = Math.ceil(historyTotal / historyPerPage);
			}
		} catch (error: any) {
			errorMessage = error.message || 'Gagal memuat detail karyawan';
			console.error('Employee detail error:', error);
		} finally {
			isLoading = false;
		}
	}

	async function loadHistoryPage(p: number) {
		if (p < 1 || p > historyTotalPages) return;
		historyPage = p;
		isLoadingHistory = true;
		try {
			const empId = employee?.id || '';
			if (!empId) return;
			const res: any = await employeesApi.getHistory(empId, { page: p, per_page: historyPerPage });
			history = res.data || [];
		} catch { /* ignore */ }
		finally { isLoadingHistory = false; }
	}

	$effect(() => {
		if (employeeId) {
			loadEmployee(employeeId);
			loadWorkSchedules();
		}
	});

	function retryLoad() {
		const id = employeeId;
		if (id) loadEmployee(id);
	}

	// --------------- Schedule Override Functions ---------------

	async function loadWorkSchedules() {
		try {
			const res: any = await wsApi.getAll();
			allWorkSchedules = res.data || [];
		} catch { /* ignore */ }
	}

	function openSchedulePicker() {
		selectedScheduleId = employee?.work_schedule_id || '';
		schedulePickerOpen = true;
	}

	function closeSchedulePicker() {
		schedulePickerOpen = false;
		selectedScheduleId = '';
	}

	async function handleSaveSchedule() {
		const emp = employee;
		if (!emp) return;
		isSavingSchedule = true;
		try {
			const res: any = await employeesApi.updateWorkSchedule(emp.id, selectedScheduleId);
			employee!.work_schedule_id = res.data?.work_schedule_id || null;
			employee!.work_schedule_name = res.data?.work_schedule_name || '';
			closeSchedulePicker();
			const empRes: any = await employeesApi.get(emp.id);
			employee = empRes.data || null;
		} catch (err: any) {
			errorMessage = err.message || 'Gagal menyimpan jadwal';
		} finally {
			isSavingSchedule = false;
		}
	}

	function getScheduleTypeLabel(t: string): string {
		const map: Record<string, string> = { five_day: '5 Hari', six_day: '6 Hari', shift: 'Shift' };
		return map[t] || t || '-';
	}

	function formatDate(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	function formatDateTime(dateStr: string | null): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	function getStatusBadge(): string {
		const emp = employee;
		if (!emp) return '';
		if (!emp.is_active) return 'bg-red-50 text-red-700 ring-red-600/20';
		const map: Record<string, string> = {
			tetap: 'bg-emerald-50 text-emerald-700 ring-emerald-600/20',
			kontrak: 'bg-blue-50 text-blue-700 ring-blue-600/20',
			percobaan: 'bg-amber-50 text-amber-700 ring-amber-600/20',
			magang: 'bg-purple-50 text-purple-700 ring-purple-600/20',
		};
		return map[emp.employment_status?.toLowerCase()] || 'bg-gray-50 text-gray-700 ring-gray-600/20';
	}

	function getGenderLabel(g: string): string {
		const map: Record<string, string> = { laki_laki: 'Laki-laki', perempuan: 'Perempuan' };
		return map[g] || g || '-';
	}

	function getMaritalStatusLabel(s: string): string {
		const map: Record<string, string> = {
			belum_menikah: 'Belum Menikah', menikah: 'Menikah', cerai: 'Cerai', cerai_mati: 'Cerai (Meninggal)'
		};
		return map[s] || s || '-';
	}

	function getInitials(name: string): string {
		const parts = name.split(' ');
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}

	function changeTypeLabel(type: string): string {
		const map: Record<string, string> = {
			created: 'Karyawan Baru',
			promotion: 'Promosi',
			mutation: 'Mutasi',
			salary_change: 'Perubahan Gaji',
			status_change: 'Perubahan Status',
			department_change: 'Perubahan Departemen',
			updated: 'Pembaruan Data',
			deactivated: 'Non-Aktif',
		};
		return map[type] || type;
	}

	function changeTypeColor(type: string): string {
		if (type === 'created') return 'bg-emerald-50 text-emerald-700 ring-emerald-200';
		if (type === 'promotion') return 'bg-blue-50 text-blue-700 ring-blue-200';
		if (type === 'mutation' || type === 'department_change') return 'bg-amber-50 text-amber-700 ring-amber-200';
		if (type === 'salary_change') return 'bg-purple-50 text-purple-700 ring-purple-200';
		if (type === 'deactivated') return 'bg-red-50 text-red-700 ring-red-200';
		return 'bg-gray-50 text-gray-700 ring-gray-200';
	}

	function historyValueSummary(val: Record<string, any> | null): string {
		if (!val) return '-';
		const parts: string[] = [];
		if (val.position_name) parts.push(`Posisi: ${val.position_name}`);
		if (val.department_name) parts.push(`Dept: ${val.department_name}`);
		if (val.employment_status) parts.push(`Status: ${val.employment_status}`);
		if (val.role_name) parts.push(`Role: ${val.role_name}`);
		if (val.base_salary) parts.push(`Gaji: ${val.base_salary}`);
		if (val.is_active !== undefined) parts.push(val.is_active ? 'Aktif' : 'Non-Aktif');
		if (val.full_name) parts.push(`Nama: ${val.full_name}`);
		return parts.join(', ') || JSON.stringify(val);
	}

	// ======================= Salary Component Functions =======================

	async function loadSalaryComponents() {
		const emp = employee;
		if (!emp) return;
		isLoadingSC = true;
		scError = '';
		try {
			const res: any = await scApi.list(emp.id, scPage, scPerPage);
			salaryComps = res.data || [];
			scTotal = res.meta?.total || 0;
			scTotalPages = Math.ceil(scTotal / scPerPage);
		} catch (err: any) {
			scError = err.message || 'Gagal memuat komponen gaji';
		} finally {
			isLoadingSC = false;
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

	function cancelSCForm() {
		showSCForm = false;
		scFormError = '';
	}

	async function handleSaveSC() {
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
				await scApi.update(employee!.id, editingSCId, payload);
			} else {
				await scApi.create(employee!.id, payload);
			}
			cancelSCForm();
			loadSalaryComponents();
		} catch (err: any) {
			scFormError = err.message || 'Gagal menyimpan komponen gaji';
		} finally {
			isSavingSC = false;
		}
	}

	function confirmSCDelete(id: string, name: string) {
		deletingSCId = id;
		deletingSCName = name;
		showSCDelete = true;
	}

	function cancelSCDelete() {
		showSCDelete = false;
		deletingSCId = null;
		deletingSCName = '';
	}

	async function handleSCDelete() {
		if (!deletingSCId) return;
		const emp = employee;
		if (!emp) return;
		isSavingSC = true;
		try {
			await scApi.remove(emp.id, deletingSCId);
			showSCDelete = false;
			deletingSCId = null;
			deletingSCName = '';
			loadSalaryComponents();
		} catch (err: any) {
			scFormError = err.message || 'Gagal menghapus komponen gaji';
			showSCDelete = false;
		} finally {
			isSavingSC = false;
		}
	}

	function scTotalAllowance(): number {
		return salaryComps.filter(c => c.component_type === 'allowance' && c.is_active).reduce((sum, c) => sum + c.amount, 0);
	}

	function scTotalDeduction(): number {
		return salaryComps.filter(c => c.component_type === 'deduction' && c.is_active).reduce((sum, c) => sum + c.amount, 0);
	}

	function scNetSalary(): number {
		return scTotalAllowance() - scTotalDeduction();
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

		// ======================= BPJS Config =======================

	type BPJSComponentConfig = {
		enabled: boolean;
		rate_employee?: number;
		rate_company?: number;
	};

	type EmployeeBPJSConfig = {
		kesehatan?: BPJSComponentConfig;
		jht?: BPJSComponentConfig;
		jp?: BPJSComponentConfig;
		jkk?: BPJSComponentConfig;
		jkm?: BPJSComponentConfig;
	};

	const BPJS_LABELS: Record<string, { label: string; description: string }> = {
		kesehatan: { label: 'BPJS Kesehatan', description: 'Premi jaminan kesehatan (4% perusahaan, 1% karyawan)' },
		jht: { label: 'BPJS JHT', description: 'Jaminan Hari Tua (3.7% perusahaan, 2% karyawan)' },
		jp: { label: 'BPJS JP', description: 'Jaminan Pensiun (2% perusahaan, 1% karyawan)' },
		jkk: { label: 'BPJS JKK', description: 'Jaminan Kecelakaan Kerja (0.24-1.74% perusahaan)' },
		jkm: { label: 'BPJS JKM', description: 'Jaminan Kematian (0.3% perusahaan)' },
	};

	let employeeBPJSConfig = $state<EmployeeBPJSConfig>({});
	let isLoadingBPJS = $state(false);
	let isSavingBPJS = $state(false);
	let bpjsError = $state('');
	let bpjsSuccess = $state('');

	async function loadEmployeeBPJSConfig() {
		const emp = employee;
		if (!emp) return;
		isLoadingBPJS = true;
		bpjsError = '';
		try {
			const res: any = await companyApi.getEmployeeBPJSConfig(emp.id);
			const config = res?.data?.bpjs_config || {};
			// Initialize all components with defaults if not set
			employeeBPJSConfig = {
				kesehatan: { enabled: config.kesehatan?.enabled ?? true },
				jht: { enabled: config.jht?.enabled ?? true },
				jp: { enabled: config.jp?.enabled ?? true },
				jkk: { enabled: config.jkk?.enabled ?? true },
				jkm: { enabled: config.jkm?.enabled ?? true },
			};
		} catch {
			// Default to all enabled if error
			employeeBPJSConfig = {
				kesehatan: { enabled: true },
				jht: { enabled: true },
				jp: { enabled: true },
				jkk: { enabled: true },
				jkm: { enabled: true },
			};
		} finally {
			isLoadingBPJS = false;
		}
	}

	async function handleSaveBPJSConfig() {
		const emp = employee;
		if (!emp) return;
		isSavingBPJS = true;
		bpjsError = '';
		bpjsSuccess = '';
		try {
			await companyApi.updateEmployeeBPJSConfig(emp.id, employeeBPJSConfig);
			bpjsSuccess = 'Konfigurasi BPJS berhasil disimpan';
			setTimeout(() => { bpjsSuccess = ''; }, 3000);
		} catch (err: any) {
			bpjsError = err.message || 'Gagal menyimpan konfigurasi BPJS';
		} finally {
			isSavingBPJS = false;
		}
	}

	$effect(() => {
		const emp = employee;
		if (emp && !isLoading && hasPermission('payroll', 'read')) {
			loadSalaryComponents();
			loadEmployeeBPJSConfig();
		}
	});
</script>

<div class="w-full">
	<!-- Close Button -->
	<button
		onclick={onclose}
		class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 transition mb-5 cursor-pointer group"
	>
		<svg class="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
		</svg>
		<span>Kembali ke daftar karyawan</span>
	</button>

	{#if isLoading}
		<PulseLoader variant="card" count={1} />
	{:else if errorMessage}
		<!-- Error State -->
		<div class="bg-white border border-gray-200 rounded-xl py-16 text-center">
			<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center">
				<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
				</svg>
			</div>
			<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
			<p class="text-sm text-gray-500 mb-5">{errorMessage}</p>
			<div class="flex items-center justify-center gap-3">
				<button onclick={onclose} class="px-4 py-2 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-50 transition cursor-pointer">Kembali</button>
				<button onclick={retryLoad} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
			</div>
		</div>
	{:else if employee}
		<!-- Employee Profile Header -->
		<div class="bg-white border border-gray-200 rounded-xl p-6 mb-4 shadow-sm">
			<div class="flex items-center gap-5">
				<div class="relative group shrink-0">
					{#if photoPreview}
						<img src={photoPreview} alt="Preview" class="w-16 h-16 rounded-full object-cover ring-4 ring-blue-50" />
					{:else if employee.photo_url}
						<img src={employee.photo_url} alt={employee.full_name} class="w-16 h-16 rounded-full object-cover ring-4 ring-blue-50" />
					{:else}
						<div class="w-16 h-16 rounded-full bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] flex items-center justify-center text-white font-bold text-xl shrink-0 ring-4 ring-blue-50">
							{getInitials(employee.full_name)}
						</div>
					{/if}
					<label class="absolute inset-0 flex items-center justify-center bg-black/40 rounded-full opacity-0 group-hover:opacity-100 transition-opacity cursor-pointer">
						{#if isUploadingPhoto}
							<svg class="w-6 h-6 text-white animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						{:else}
							<svg class="w-6 h-6 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z" /></svg>
						{/if}
						<input type="file" accept="image/jpeg,image/png" onchange={handlePhotoUpload} class="hidden" disabled={isUploadingPhoto} />
					</label>
				</div>
				<div class="flex-1 min-w-0">
					<div class="flex items-center gap-3 flex-wrap">
						<h1 class="text-xl font-bold text-gray-900">{employee.full_name}</h1>
						{#if !employee.is_active}
							<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ring-1 ring-inset bg-red-50 text-red-700 ring-red-600/20">Non-Aktif</span>
						{/if}
					</div>
					<p class="text-sm text-gray-500 mt-0.5">{employee.position_name || '-'}</p>
					<div class="flex items-center gap-2 mt-1.5 text-xs text-gray-400">
						<span>{employee.department_name || '-'}</span>
					</div>
				</div>
				<span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ring-1 ring-inset capitalize shrink-0 {getStatusBadge()}">
					{employee.employment_status}
				</span>
			</div>
		</div>

		<!-- Info Grid -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
			<!-- Personal Info -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Informasi Pribadi</h2>
				</div>
				<div class="space-y-3.5">
				{#each [
					{label:'Email',val:employee.email},
					{label:'NIK',val:employee.nik||'-'},
					{label:'NPWP',val:employee.npwp||'-'},
					{label:'Jenis Kelamin',val:getGenderLabel(employee.gender)},
					{label:'Tempat, Tgl Lahir',val:`${employee.place_of_birth || '-'}${employee.date_of_birth ? `, ${formatDate(employee.date_of_birth)}` : ''}`},
					{label:'Agama',val:employee.religion||'-'},
					{label:'Status Pernikahan',val:getMaritalStatusLabel(employee.marital_status)},
					{label:'No. Telepon',val:employee.phone||'-'},
				] as item}
					<div class="flex justify-between items-start">
						<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
						<span class="text-sm text-gray-900 text-right">{item.val}</span>
					</div>
				{/each}
				</div>
			</div>

			<!-- Employment Info -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Informasi Pekerjaan</h2>
				</div>
				<div class="space-y-3.5">
					{#each [
						{label:'Posisi',val:employee.position_name||'-'},
						{label:'Departemen',val:employee.department_name||'-'},
						{label:'Role',val:employee.role_name||'-'},
						{label:'Status',val:employee.employment_status||'-'},
						{label:'Bergabung',val:formatDate(employee.join_date)},
						{label:'Terakhir Login',val:formatDateTime(employee.last_login_at)},
					] as item}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right capitalize">{item.val}</span>
						</div>
					{/each}
				</div>
			</div>

			<!-- Bank Info -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 21v-8.25M15.75 21v-8.25M8.25 21v-8.25M3 9l9-6 9 6m-1.5 12V10.332A48.36 48.36 0 0 0 12 9.75c-2.551 0-5.056.2-7.5.582V21M3 21h18M12 6.75h.008v.008H12V6.75Z" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Informasi Bank</h2>
				</div>
				<div class="space-y-3.5">
					{#each [
						{label:'Nama Bank',val:employee.bank_name||'-'},
						{label:'No. Rekening',val:employee.bank_account||'-'},
					] as item}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right">{item.val}</span>
						</div>
					{/each}
				</div>
			</div>

			<!-- Address -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
							<path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Alamat</h2>
				</div>
				<div class="space-y-3">
					<div>
						<span class="text-xs text-gray-400">Alamat KTP</span>
						<p class="text-sm text-gray-700 leading-relaxed mt-0.5">{employee.address_ktp || 'Tidak ada'}</p>
					</div>
					<div class="border-t border-gray-100 pt-3">
						<span class="text-xs text-gray-400">Alamat Domisili</span>
						<p class="text-sm text-gray-700 leading-relaxed mt-0.5">{employee.address || 'Tidak ada'}</p>
					</div>
				</div>
			</div>

			<!-- Schedule Info -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-violet-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-violet-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Jadwal Kerja</h2>
				</div>

				{#if schedulePickerOpen}
					<div class="bg-gray-50/50 border border-gray-200 rounded-xl p-4">
						<div class="mb-3">
							<label for="schedule-select" class="block text-xs font-medium text-gray-600 mb-1.5">Pilih Jadwal Kerja</label>
							<select id="schedule-select" bind:value={selectedScheduleId}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">— Ikuti Jadwal Departemen —</option>
								{#each allWorkSchedules as ws}
									<option value={ws.id}>{ws.name} ({getScheduleTypeLabel(ws.schedule_type)})</option>
								{/each}
							</select>
						</div>
						<div class="flex items-center justify-end gap-2">
							<button onclick={closeSchedulePicker} onkeydown={(e) => e.key === 'Enter' && closeSchedulePicker()} class="px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-100 transition cursor-pointer">Batal</button>
							<button onclick={handleSaveSchedule} disabled={isSavingSchedule}
								class="px-4 py-1.5 bg-[#1A56DB] text-white rounded-lg text-xs font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
								{#if isSavingSchedule}
									<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								{/if}
								Simpan
							</button>
						</div>
					</div>
				{:else}
					<div class="space-y-3.5">
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">Jadwal Saat Ini</span>
							<span class="text-sm text-gray-900 text-right">
								{#if employee.work_schedule_name}
									{employee.work_schedule_name}
								{:else}
									<span class="text-gray-400 italic">Mengikuti jadwal departemen</span>
								{/if}
							</span>
						</div>
						{#if employee.department_name}
							<div class="flex justify-between items-start">
								<span class="text-xs text-gray-400 shrink-0 w-32">Departemen</span>
								<span class="text-sm text-gray-900 text-right">{employee.department_name}</span>
							</div>
						{/if}
						<button onclick={openSchedulePicker}
							class="mt-2 inline-flex items-center gap-1.5 px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-100 transition cursor-pointer">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
							</svg>
							{employee.work_schedule_name ? 'Ubah Jadwal' : 'Atur Jadwal Individual'}
						</button>
					</div>
				{/if}
			</div>

			<!-- System Info -->
			<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
				<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
					<div class="w-8 h-8 rounded-lg bg-purple-50 flex items-center justify-center">
						<svg class="w-4 h-4 text-purple-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" />
						</svg>
					</div>
					<h2 class="text-sm font-semibold text-gray-900">Informasi Sistem</h2>
				</div>
				<div class="space-y-3.5">
					{#each [
						{label:'Dibuat Pada',val:formatDateTime(employee.created_at)},
						{label:'Diperbarui Pada',val:formatDateTime(employee.updated_at)},
						{label:'Akun Terkunci',val:employee.is_locked ? 'Ya' : 'Tidak'},
					] as item}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right">{item.val}</span>
						</div>
					{/each}
				</div>
			</div>
		</div>

		<!-- Riwayat Karyawan -->
		<div class="bg-white border border-gray-200 rounded-xl p-5 shadow-sm">
			<div class="flex items-center gap-2.5 mb-5 pb-3 border-b border-gray-100">
				<div class="w-8 h-8 rounded-lg bg-gray-50 flex items-center justify-center">
					<svg class="w-4 h-4 text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
					</svg>
				</div>
				<h2 class="text-sm font-semibold text-gray-900">Riwayat</h2>
				<span class="text-xs text-gray-400 ml-auto">{historyTotal} catatan</span>
			</div>

			{#if isLoadingHistory}
				<PulseLoader variant="text" count={3} />
			{:else if history.length === 0}
				<p class="text-sm text-gray-400 text-center py-6">Belum ada riwayat perubahan</p>
			{:else}
				<div class="relative">
					<div class="absolute left-4 top-2 bottom-2 w-0.5 bg-gray-100"></div>
					<div class="space-y-0">
						{#each history as item}
							<div class="relative pl-10 pb-4">
								<div class="absolute left-2.5 top-1.5 w-3 h-3 rounded-full ring-2 ring-white {changeTypeColor(item.change_type).split(' ')[0]}"></div>
								<div class="flex items-start justify-between gap-2">
									<div>
										<div class="flex items-center gap-2 flex-wrap">
											<span class="text-sm font-medium text-gray-900">{changeTypeLabel(item.change_type)}</span>
											{#if item.reason}<span class="text-xs text-gray-400">— {item.reason}</span>{/if}
										</div>
										<div class="text-xs text-gray-400 mt-0.5">
											<span>{formatDateTime(item.changed_at)}</span>
											{#if item.changed_by_name}<span> — oleh {item.changed_by_name}</span>{/if}
										</div>
										{#if item.new_value}
											<div class="mt-1.5 text-xs text-gray-500 bg-gray-50 rounded-md px-2.5 py-1.5 inline-block">
												{historyValueSummary(item.new_value)}
											</div>
										{/if}
									</div>
								</div>
							</div>
						{/each}
					</div>
				</div>
				{#if historyTotalPages > 1}
					<div class="flex items-center justify-center gap-1.5 pt-2 border-t border-gray-100 mt-2">
						<button onclick={() => loadHistoryPage(historyPage - 1)} disabled={historyPage <= 1}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						<span class="text-xs text-gray-400 px-2">Halaman {historyPage} dari {historyTotalPages}</span>
						<button onclick={() => loadHistoryPage(historyPage + 1)} disabled={historyPage >= historyTotalPages}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				{/if}
			{/if}
		</div>

		<!-- Komponen Gaji — hanya tampil untuk user dengan akses payroll -->
		{#if hasPermission('payroll', 'read')}
		<div class="bg-white border border-gray-200 rounded-xl shadow-sm mt-4 overflow-hidden">
			<!-- Header dengan gradient -->
			<div class="bg-gradient-to-r from-[#1A56DB]/5 to-[#1e3a8a]/5 px-6 py-4 border-b border-gray-200">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] flex items-center justify-center shadow-lg shadow-blue-200">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
						</div>
						<div>
							<h2 class="text-base font-bold text-gray-900">Pengaturan Gaji</h2>
							<p class="text-xs text-gray-500 mt-0.5">Kelola gaji pokok, tunjangan, dan potongan karyawan</p>
						</div>
					</div>
					{#if !showSCForm}
						<button
							onclick={openCreateSCForm}
							class="inline-flex items-center gap-2 px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-xs font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer"
						>
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
							Tambah Komponen
						</button>
					{/if}
				</div>
			</div>

			<div class="px-6 py-5">
			<!-- Gaji Pokok & Upah Harian — Salary Breakdown Card -->
			<div class="bg-gradient-to-br from-gray-50 to-white border border-gray-200 rounded-xl p-5 mb-5">
				<div class="flex items-center gap-2 mb-4">
					<svg class="w-4 h-4 text-[#1A56DB]" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 0 0 6 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0 1 18 16.5h-2.25m-7.5 0h7.5m-7.5 0l-1 3m8.5-3l1 3m0 0l.5 1.5m-.5-1.5h-9.5m0 0l-.5 1.5m.75-9l3-3 2.148 2.148A12.061 12.061 0 0 1 16.5 7.605" />
					</svg>
					<h3 class="text-sm font-semibold text-gray-800">Gaji Pokok & Upah Harian</h3>
					<span class="ml-auto text-[10px] text-gray-400">Nilai dalam Rupiah (Rp)</span>
				</div>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
					<div class="sm:col-span-1">
						<label for="salary-base" class="block text-xs font-semibold text-gray-600 mb-1.5">Gaji Pokok Bulanan</label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
								<span class="text-gray-400 text-sm font-medium">Rp</span>
							</div>
							<input id="salary-base" type="number" min="0" step="100000" bind:value={salaryBase}
								class="w-full pl-11 pr-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm font-medium outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-300"
								placeholder="0" />
						</div>
					</div>
					<div class="sm:col-span-1">
						<label for="salary-daily" class="block text-xs font-semibold text-gray-600 mb-1.5">Upah Harian</label>
						<div class="relative">
							<div class="absolute inset-y-0 left-0 pl-3.5 flex items-center pointer-events-none">
								<span class="text-gray-400 text-sm font-medium">Rp</span>
							</div>
							<input id="salary-daily" type="number" min="0" step="1000" bind:value={salaryDaily}
								class="w-full pl-11 pr-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm font-medium outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-300"
								placeholder="0" />
						</div>
						<p class="text-[11px] text-gray-400 mt-1.5">Untuk status karyawan <span class="font-medium text-gray-500">Harian</span></p>
					</div>
					<div class="flex items-end sm:col-span-2 lg:col-span-2">
						<div class="w-full">
							<div class="flex items-center justify-between mb-2">
								<span class="text-xs font-semibold text-gray-600">Ringkasan Gaji</span>
								<button onclick={handleSaveSalary} disabled={isSavingSalary}
									class="inline-flex items-center gap-1.5 px-5 py-2.5 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] text-white rounded-lg text-xs font-semibold hover:from-[#1e40af] hover:to-[#1e3a8a] transition-all active:scale-[0.97] disabled:opacity-50 shadow-sm shadow-blue-200 cursor-pointer">
									{#if isSavingSalary}
										<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
									{:else}
										<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
									{/if}
									Simpan Gaji
								</button>
							</div>
							<div class="bg-white border border-gray-200 rounded-lg p-3 grid grid-cols-2 gap-3">
								<div>
									<span class="text-[10px] text-gray-400 uppercase tracking-wider font-medium">Gaji Pokok</span>
									<p class="text-sm font-bold text-gray-900 tabular-nums mt-0.5">
								{#if Number(salaryBase) > 0}
									{formatCurrency(Number(salaryBase))}
								{:else}
									<span class="text-gray-300 font-normal">Belum diatur</span>
								{/if}
							</p>
								</div>
								<div>
									<span class="text-[10px] text-gray-400 uppercase tracking-wider font-medium">Upah Harian</span>
									<p class="text-sm font-bold text-gray-900 tabular-nums mt-0.5">
								{#if Number(salaryDaily) > 0}
									{formatCurrency(Number(salaryDaily))}/hari
								{:else}
									<span class="text-gray-300 font-normal">Belum diatur</span>
								{/if}
							</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>

			<!-- Inline Form — Komponen Baru/Edsal -->
			{#if showSCForm}
				<div class="bg-gradient-to-br from-gray-50 to-white border-2 border-[#1A56DB]/10 rounded-xl p-5 mb-5 relative">
					<div class="absolute -top-2.5 left-4 px-2 bg-gradient-to-br from-gray-50 to-white text-[11px] font-semibold text-[#1A56DB] uppercase tracking-wider">{scFormTitle}</div>
					{#if scFormError}
						<div class="flex items-center gap-2 bg-red-50 border border-red-200 text-red-700 text-xs px-4 py-2.5 rounded-lg mb-4">
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
							{scFormError}
						</div>
					{/if}
					<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
						<div>
							<label for="sc-name" class="block text-[11px] font-semibold text-gray-600 mb-1.5 uppercase tracking-wider">Nama Komponen <span class="text-red-500">*</span></label>
							<input id="sc-name" type="text" bind:value={scForm.component_name}
								class="w-full px-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-300"
								placeholder="Contoh: Tunj. Jabatan" />
						</div>
						<div>
							<label for="sc-type" class="block text-[11px] font-semibold text-gray-600 mb-1.5 uppercase tracking-wider">Tipe <span class="text-red-500">*</span></label>
							<div class="relative">
								<select id="sc-type" bind:value={scForm.component_type}
									class="w-full px-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition appearance-none cursor-pointer">
									<option value="allowance">➕ Tunjangan</option>
									<option value="deduction">➖ Potongan</option>
								</select>
								<svg class="w-4 h-4 text-gray-400 absolute right-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" /></svg>
							</div>
						</div>
						<div>
							<label for="sc-amount" class="block text-[11px] font-semibold text-gray-600 mb-1.5 uppercase tracking-wider">Jumlah <span class="text-red-500">*</span></label>
							<div class="relative">
								<div class="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
									<span class="text-gray-400 text-xs">Rp</span>
								</div>
								<input id="sc-amount" type="number" bind:value={scForm.amount} min="0" step="50000"
									class="w-full pl-9 pr-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition"
									placeholder="0" />
							</div>
						</div>
						<div>
							<label for="sc-effective-date" class="block text-[11px] font-semibold text-gray-600 mb-1.5 uppercase tracking-wider">Tgl Efektif</label>
							<input id="sc-effective-date" type="date" bind:value={scForm.effective_date}
								class="w-full px-3 py-2.5 bg-white border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
						</div>
					</div>
					<div class="flex items-center justify-end gap-2.5 mt-4 pt-4 border-t border-gray-200">
						<button onclick={cancelSCForm} class="px-4 py-2 border border-gray-200 rounded-lg text-xs font-semibold text-gray-600 hover:bg-gray-100 hover:text-gray-800 transition-all cursor-pointer">Batal</button>
						<button onclick={handleSaveSC} disabled={isSavingSC}
							class="px-5 py-2 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] text-white rounded-lg text-xs font-semibold hover:from-[#1e40af] hover:to-[#1e3a8a] transition-all active:scale-[0.97] disabled:opacity-50 inline-flex items-center gap-2 shadow-sm shadow-blue-200 cursor-pointer">
							{#if isSavingSC}
								<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							{:else}
								<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
							{/if}
							{editingSCId ? 'Simpan Perubahan' : 'Tambah Komponen'}
						</button>
					</div>
				</div>
			{/if}

			<!-- Summary KPI Cards -->
			{#if salaryComps.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-6">
					<div class="relative overflow-hidden bg-gradient-to-br from-emerald-50 to-emerald-100/60 border border-emerald-200 rounded-xl px-5 py-4">
						<div class="absolute top-0 right-0 w-20 h-20 bg-emerald-200/20 rounded-full -translate-y-1/2 translate-x-1/2"></div>
						<div class="flex items-center gap-3 mb-2">
							<div class="w-9 h-9 rounded-lg bg-emerald-500/10 flex items-center justify-center">
								<svg class="w-5 h-5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33" />
								</svg>
							</div>
							<span class="text-xs font-semibold text-emerald-700 uppercase tracking-wider">Total Tunjangan</span>
						</div>
						<p class="text-xl font-bold text-emerald-800 tabular-nums">{formatCurrency(scTotalAllowance())}</p>
						<p class="text-[10px] text-emerald-500 mt-0.5">Dari {salaryComps.filter(c => c.component_type === 'allowance').length} komponen</p>
					</div>
					<div class="relative overflow-hidden bg-gradient-to-br from-red-50 to-red-100/60 border border-red-200 rounded-xl px-5 py-4">
						<div class="absolute top-0 right-0 w-20 h-20 bg-red-200/20 rounded-full -translate-y-1/2 translate-x-1/2"></div>
						<div class="flex items-center gap-3 mb-2">
							<div class="w-9 h-9 rounded-lg bg-red-500/10 flex items-center justify-center">
								<svg class="w-5 h-5 text-red-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
								</svg>
							</div>
							<span class="text-xs font-semibold text-red-700 uppercase tracking-wider">Total Potongan</span>
						</div>
						<p class="text-xl font-bold text-red-800 tabular-nums">{formatCurrency(scTotalDeduction())}</p>
						<p class="text-[10px] text-red-500 mt-0.5">Dari {salaryComps.filter(c => c.component_type === 'deduction').length} komponen</p>
					</div>
					<div class="relative overflow-hidden bg-gradient-to-br from-blue-50 to-blue-100/60 border border-blue-200 rounded-xl px-5 py-4">
						<div class="absolute top-0 right-0 w-20 h-20 bg-blue-200/20 rounded-full -translate-y-1/2 translate-x-1/2"></div>
						<div class="flex items-center gap-3 mb-2">
							<div class="w-9 h-9 rounded-lg bg-blue-500/10 flex items-center justify-center">
								<svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5m8.25 3v6.75m0 0l-3-3m3 3l3-3M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z" />
								</svg>
							</div>
							<span class="text-xs font-semibold text-blue-700 uppercase tracking-wider">Gaji Bersih</span>
						</div>
						<p class="text-xl font-bold text-blue-800 tabular-nums">{formatCurrency(scNetSalary())}</p>
						<p class="text-[10px] text-blue-500 mt-0.5">Tunjangan - Potongan</p>
					</div>
				</div>
			{/if}

			<!-- Loading State -->
			{#if isLoadingSC}
				<PulseLoader variant="table-row" count={3} />
			{:else if salaryComps.length === 0}
				<!-- Empty State -->
				<div class="text-center py-12 px-6">
					<div class="w-16 h-16 mx-auto mb-4 rounded-2xl bg-gray-50 flex items-center justify-center">
						<svg class="w-8 h-8 text-gray-300" fill="none" viewBox="0 0 24 24" stroke-width="1" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
						</svg>
					</div>
					<h3 class="text-sm font-semibold text-gray-900 mb-1">Belum Ada Komponen Gaji</h3>
					<p class="text-sm text-gray-400 max-w-xs mx-auto">Tambah komponen seperti tunjangan jabatan, BPJS, atau potongan lainnya untuk menghitung gaji bersih karyawan.</p>
				</div>
			{:else}
				<!-- Desktop Table -->
				<div class="hidden md:block">
					<div class="overflow-x-auto">
						<table class="w-full">
							<thead>
								<tr class="border-b border-gray-200 bg-gradient-to-r from-gray-50 to-white">
									<th class="text-left px-5 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Nama Komponen</th>
									<th class="text-left px-5 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Tipe</th>
									<th class="text-right px-5 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Jumlah</th>
									<th class="text-left px-5 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Efektif</th>
									<th class="text-left px-5 py-3 text-[11px] font-semibold text-gray-500 uppercase tracking-wider">Status</th>
									<th class="px-5 py-3 w-20"></th>
								</tr>
							</thead>
							<tbody class="divide-y divide-gray-50">
								{#each salaryComps as comp, i}
									<tr class="{i % 2 === 0 ? 'bg-white' : 'bg-gray-50/30'} hover:bg-blue-50/40 transition-colors group">
										<td class="px-5 py-3.5">
											<div class="flex items-center gap-3">
												<div class="w-8 h-8 rounded-lg {comp.component_type === 'allowance' ? 'bg-emerald-50' : 'bg-red-50'} flex items-center justify-center">
													<svg class="w-4 h-4 {comp.component_type === 'allowance' ? 'text-emerald-600' : 'text-red-600'}" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
														{#if comp.component_type === 'allowance'}
															<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33" />
														{:else}
															<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
														{/if}
													</svg>
												</div>
												<span class="text-sm font-medium text-gray-900">{comp.component_name}</span>
											</div>
										</td>
										<td class="px-5 py-3.5">
											{#if comp.component_type === 'allowance'}
												<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20">Tunjangan</span>
											{:else}
												<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-[11px] font-semibold bg-red-50 text-red-700 ring-1 ring-red-600/20">Potongan</span>
											{/if}
										</td>
										<td class="px-5 py-3.5 text-sm font-semibold text-right tabular-nums {comp.component_type === 'allowance' ? 'text-emerald-700' : 'text-red-700'}">{formatCurrency(comp.amount)}</td>
										<td class="px-5 py-3.5 text-sm text-gray-500">{formatDate(comp.effective_date)}</td>
										<td class="px-5 py-3.5">
											<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[11px] font-medium {comp.is_active ? 'bg-emerald-50 text-emerald-700 ring-1 ring-emerald-600/20' : 'bg-gray-100 text-gray-500 ring-1 ring-gray-200'}">
												<span class="w-1.5 h-1.5 rounded-full {comp.is_active ? 'bg-emerald-500' : 'bg-gray-400'}"></span>
												{comp.is_active ? 'Aktif' : 'Nonaktif'}
											</span>
										</td>
										<td class="px-5 py-3.5 text-right">
											<div class="flex items-center justify-end gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
												<button onclick={() => openEditSCForm(comp)}
													class="p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer"
													title="Edit komponen">
													<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
												</button>
												<button onclick={() => confirmSCDelete(comp.id, comp.component_name)}
													class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer"
													title="Hapus komponen">
													<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
												</button>
											</div>
										</td>
									</tr>
								{/each}
							</tbody>
						</table>
					</div>
				</div>

				<!-- Mobile Cards -->
				<div class="md:hidden divide-y divide-gray-100">
					{#each salaryComps as comp}
						<div class="py-3.5 px-1 flex items-center justify-between gap-3 hover:bg-blue-50/20 transition-colors rounded-lg">
							<div class="min-w-0 flex-1">
								<div class="text-sm font-medium text-gray-900 truncate">{comp.component_name}</div>
								<div class="flex items-center gap-2 text-xs text-gray-400 mt-0.5">
									<span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-semibold {comp.component_type === 'allowance' ? 'text-emerald-700 bg-emerald-50' : 'text-red-700 bg-red-50'}">
										{comp.component_type === 'allowance' ? 'Tnj' : 'Ptn'}
									</span>
									<span class="font-medium {comp.component_type === 'allowance' ? 'text-emerald-600' : 'text-red-600'}">{formatCurrency(comp.amount)}</span>
								</div>
							</div>
							<div class="flex items-center gap-0.5 shrink-0">
								<button onclick={() => openEditSCForm(comp)} class="p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer" title="Edit">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125" /></svg>
								</button>
								<button onclick={() => confirmSCDelete(comp.id, comp.component_name)} class="p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer" title="Hapus">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79" /></svg>
								</button>
							</div>
						</div>
					{/each}
				</div>

				<!-- Pagination -->
				{#if scTotalPages > 1}
					<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30 mt-2 rounded-b-xl">
						<div class="text-xs text-gray-400">{scTotal} komponen</div>
						<div class="flex items-center gap-1.5">
							<button onclick={() => { scPage--; loadSalaryComponents(); }} disabled={scPage <= 1}
								class="px-2.5 py-1.5 text-[11px] font-medium rounded-lg border border-gray-200 text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
							<span class="text-xs text-gray-400 px-1.5 font-medium">{scPage}/{scTotalPages}</span>
							<button onclick={() => { scPage++; loadSalaryComponents(); }} disabled={scPage >= scTotalPages}
								class="px-2.5 py-1.5 text-[11px] font-medium rounded-lg border border-gray-200 text-gray-500 hover:bg-gray-100 hover:text-gray-700 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
						</div>
					</div>
				{/if}
			{/if}
		</div>
	</div>
		{/if}

		<!-- Delete Confirmation Modal for Salary Component -->
		<AnimatedPresence show={showSCDelete} type="scale" duration={200}>
	<!-- svelte-ignore a11y_interactive_supports_focus -->
			<!-- svelte-ignore a11y_click_events_have_key_events -->
			<div onclick={cancelSCDelete} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelSCDelete(); }}
				role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
				<div onclick={(e) => e.stopPropagation()} class="bg-white rounded-2xl shadow-2xl w-full max-w-sm" role="dialog" tabindex="-1" aria-modal="true" aria-label="Konfirmasi hapus">
					<div class="px-6 py-6 text-center">
						<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center">
							<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
							</svg>
						</div>
						<h3 class="text-lg font-semibold text-gray-900 mb-2">Hapus Komponen Gaji</h3>
						<p class="text-sm text-gray-500 mb-1">Apakah Anda yakin ingin menghapus</p>
						<p class="text-sm font-medium text-gray-900 mb-4">"{deletingSCName}"?</p>
						<div class="flex items-center justify-center gap-3">
							<button onclick={cancelSCDelete} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
							<button onclick={handleSCDelete} disabled={isSavingSC}
								class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
								{#if isSavingSC}
									<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								{/if}
								Ya, Hapus
							</button>
						</div>
					</div>
				</div>
			</div>
		</AnimatedPresence>
	{/if}
</div>
