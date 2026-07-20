<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { employees as employeesApi, salaryComponents as scApi, workSchedules as wsApi, company as companyApi, auth } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';

	let { 
		employeeId, 
		initialTab = 'profile', 
		onclose 
	}: { 
		employeeId: string; 
		initialTab?: 'profile' | 'salary' | 'bpjs' | 'tax' | 'overtime'; 
		onclose: () => void 
	} = $props();

	// --------------- is_pregnant toggle state ---------------
	let isPregnant = $state(false);
	let isSavingPregnant = $state(false);
	
	async function handleTogglePregnant() {
		const emp = employee;
		if (!emp) return;
		isSavingPregnant = true;
		try {
			const newVal = !isPregnant;
			await employeesApi.update(emp.id, { is_pregnant: newVal });
			isPregnant = newVal;
			employee!.is_pregnant = newVal;
		} catch (err: any) {
			errorMessage = err.message || 'Gagal menyimpan status kehamilan';
		} finally {
			isSavingPregnant = false;
		}
	}

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

	// --------------- Approval Line State ---------------
	let allEmployees = $state<{ id: string; full_name: string; position_name: string; department_name: string; }[]>([]);
	let approvalLinePickerOpen = $state(false);
	let selectedApprovalLineId = $state('');
	let isSavingApprovalLine = $state(false);

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
		approval_line_id: string | null;
		approval_line_name: string;
		is_pregnant: boolean;
		last_login_at: string | null;
		is_locked: boolean;
		locked_until: string | null;
		ptkp_status: string;
		created_at: string;
		updated_at: string;
	};

	let employee = $state<EmployeeDetail | null>(null);

	const currentUser = $derived(auth.getUser() as any);
	const isEmployeeRole = $derived(currentUser?.role_slug === 'employee');
	const isViewingOther = $derived(currentUser && employee && currentUser.id !== employee.id);
	const shouldRestrict = $derived(isEmployeeRole && isViewingOther);
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
			loadAllEmployees();
		}
	});

	$effect(() => {
		if (employee) {
			isPregnant = employee.is_pregnant;
		}
	});

	function retryLoad() {
		const id = employeeId;
		if (id) loadEmployee(id);
	}

	// --------------- Approval Line Functions ---------------

	async function loadAllEmployees() {
		try {
			const res: any = await employeesApi.list(1, 1000, '', '', '', true);
			allEmployees = (res.data || []).filter((e: any) => e.id !== employeeId && e.is_active !== false);
		} catch { /* ignore */ }
	}

	function openApprovalLinePicker() {
		selectedApprovalLineId = employee?.approval_line_id || '';
		approvalLinePickerOpen = true;
	}

	function closeApprovalLinePicker() {
		approvalLinePickerOpen = false;
		selectedApprovalLineId = '';
	}

	async function handleSaveApprovalLine() {
		const emp = employee;
		if (!emp) return;
		isSavingApprovalLine = true;
		try {
			// Send empty string to clear approval_line_id (backend NULLIF('', '')::uuid handles it)
			await employeesApi.update(emp.id, { approval_line_id: selectedApprovalLineId || '' });
			// Refresh employee data
			const empRes: any = await employeesApi.get(emp.id);
			if (empRes.data) {
				employee = empRes.data;
			}
			closeApprovalLinePicker();
		} catch (err: any) {
			errorMessage = err.message || 'Gagal menyimpan atasan';
		} finally {
			isSavingApprovalLine = false;
		}
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
		employee_rate?: number;
		company_rate?: number;
		employee_nominal?: number;
		company_nominal?: number;
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

	// Helper to track override methods in Svelte
	let bpjsOverrideMethods = $state<Record<string, { emp: 'default'|'rate'|'nominal', comp: 'default'|'rate'|'nominal' }>>({
		kesehatan: { emp: 'default', comp: 'default' },
		jht: { emp: 'default', comp: 'default' },
		jp: { emp: 'default', comp: 'default' },
		jkk: { emp: 'default', comp: 'default' },
		jkm: { emp: 'default', comp: 'default' },
	});

	async function loadEmployeeBPJSConfig() {
		const emp = employee;
		if (!emp) return;
		isLoadingBPJS = true;
		bpjsError = '';
		try {
			const res: any = await companyApi.getEmployeeBPJSConfig(emp.id);
			const config = res?.data?.bpjs_config || {};
			
			employeeBPJSConfig = {
				kesehatan: {
					enabled: config.kesehatan?.enabled ?? true,
					employee_rate: config.kesehatan?.employee_rate !== undefined ? config.kesehatan.employee_rate * 100 : undefined,
					company_rate: config.kesehatan?.company_rate !== undefined ? config.kesehatan.company_rate * 100 : undefined,
					employee_nominal: config.kesehatan?.employee_nominal,
					company_nominal: config.kesehatan?.company_nominal,
				},
				jht: {
					enabled: config.jht?.enabled ?? true,
					employee_rate: config.jht?.employee_rate !== undefined ? config.jht.employee_rate * 100 : undefined,
					company_rate: config.jht?.company_rate !== undefined ? config.jht.company_rate * 100 : undefined,
					employee_nominal: config.jht?.employee_nominal,
					company_nominal: config.jht?.company_nominal,
				},
				jp: {
					enabled: config.jp?.enabled ?? true,
					employee_rate: config.jp?.employee_rate !== undefined ? config.jp.employee_rate * 100 : undefined,
					company_rate: config.jp?.company_rate !== undefined ? config.jp.company_rate * 100 : undefined,
					employee_nominal: config.jp?.employee_nominal,
					company_nominal: config.jp?.company_nominal,
				},
				jkk: {
					enabled: config.jkk?.enabled ?? true,
					company_rate: config.jkk?.company_rate !== undefined ? config.jkk.company_rate * 100 : undefined,
					company_nominal: config.jkk?.company_nominal,
				},
				jkm: {
					enabled: config.jkm?.enabled ?? true,
					company_rate: config.jkm?.company_rate !== undefined ? config.jkm.company_rate * 100 : undefined,
					company_nominal: config.jkm?.company_nominal,
				},
			};

			// Determine override methods
			for (const key of ['kesehatan', 'jht', 'jp', 'jkk', 'jkm'] as const) {
				const c = employeeBPJSConfig[key];
				if (c) {
					bpjsOverrideMethods[key] = {
						emp: c.employee_nominal !== undefined ? 'nominal' : (c.employee_rate !== undefined ? 'rate' : 'default'),
						comp: c.company_nominal !== undefined ? 'nominal' : (c.company_rate !== undefined ? 'rate' : 'default'),
					};
				}
			}
		} catch {
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
			const payload: any = {};
			for (const key of ['kesehatan', 'jht', 'jp', 'jkk', 'jkm'] as const) {
				const comp = employeeBPJSConfig[key];
				const method = bpjsOverrideMethods[key];
				if (comp && method) {
					payload[key] = {
						enabled: comp.enabled,
						employee_rate: comp.enabled && method.emp === 'rate' && comp.employee_rate !== undefined ? comp.employee_rate / 100 : null,
						company_rate: comp.enabled && method.comp === 'rate' && comp.company_rate !== undefined ? comp.company_rate / 100 : null,
						employee_nominal: comp.enabled && method.emp === 'nominal' && comp.employee_nominal !== undefined ? comp.employee_nominal : null,
						company_nominal: comp.enabled && method.comp === 'nominal' && comp.company_nominal !== undefined ? comp.company_nominal : null,
					};
				}
			}
			await companyApi.updateEmployeeBPJSConfig(emp.id, payload);
			bpjsSuccess = 'Konfigurasi BPJS berhasil disimpan';
			setTimeout(() => { bpjsSuccess = ''; }, 3000);
		} catch (err: any) {
			bpjsError = err.message || 'Gagal menyimpan konfigurasi BPJS';
		} finally {
			isSavingBPJS = false;
		}
	}

	// ======================= Tax (PPh 21) Config =======================
	type EmployeeTaxConfig = {
		override_type: 'rate' | 'nominal' | 'none' | 'free';
		override_rate?: number;
		override_nominal?: number;
	};

	let employeeTaxConfig = $state<EmployeeTaxConfig>({ override_type: 'none' });
	let isLoadingTax = $state(false);
	let isSavingTax = $state(false);
	let taxError = $state('');
	let taxSuccess = $state('');

	async function loadEmployeeTaxConfig() {
		const emp = employee;
		if (!emp) return;
		isLoadingTax = true;
		taxError = '';
		try {
			const res: any = await companyApi.getEmployeeTaxConfig(emp.id);
			const config = res?.data?.tax_config || { override_type: 'none' };
			employeeTaxConfig = {
				override_type: config.override_type || 'none',
				override_rate: config.override_rate !== undefined ? config.override_rate * 100 : undefined,
				override_nominal: config.override_nominal,
			};
		} catch {
			employeeTaxConfig = { override_type: 'none' };
		} finally {
			isLoadingTax = false;
		}
	}

	async function handleSaveTaxConfig() {
		const emp = employee;
		if (!emp) return;
		isSavingTax = true;
		taxError = '';
		taxSuccess = '';
		try {
			const payload: any = {
				override_type: employeeTaxConfig.override_type,
				override_rate: employeeTaxConfig.override_type === 'rate' && employeeTaxConfig.override_rate !== undefined ? employeeTaxConfig.override_rate / 100 : null,
				override_nominal: employeeTaxConfig.override_type === 'nominal' && employeeTaxConfig.override_nominal !== undefined ? employeeTaxConfig.override_nominal : null,
			};
			await companyApi.updateEmployeeTaxConfig(emp.id, payload);
			taxSuccess = 'Konfigurasi pajak berhasil disimpan';
			setTimeout(() => { taxSuccess = ''; }, 3000);
		} catch (err: any) {
			taxError = err.message || 'Gagal menyimpan konfigurasi pajak';
		} finally {
			isSavingTax = false;
		}
	}

	// ======================= Overtime Config =======================
	type EmployeeOvertimeConfig = {
		override_type: 'hourly_rate' | 'divisor' | 'percentage' | 'none';
		hourly_rate?: number;
		divisor?: number;
		rate_percentage?: number;
	};

	let employeeOvertimeConfig = $state<EmployeeOvertimeConfig>({ override_type: 'none' });
	let isLoadingOvertime = $state(false);
	let isSavingOvertime = $state(false);
	let overtimeError = $state('');
	let overtimeSuccess = $state('');
	let activeDetailTab = $state<'profile' | 'salary' | 'bpjs' | 'tax' | 'overtime'>(initialTab);

	$effect(() => {
		if (initialTab) {
			activeDetailTab = initialTab;
		}
	});

	async function loadEmployeeOvertimeConfig() {
		const emp = employee;
		if (!emp) return;
		isLoadingOvertime = true;
		overtimeError = '';
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
			isLoadingOvertime = false;
		}
	}

	async function handleSaveOvertimeConfig() {
		const emp = employee;
		if (!emp) return;
		isSavingOvertime = true;
		overtimeError = '';
		overtimeSuccess = '';
		try {
			const payload: any = {
				override_type: employeeOvertimeConfig.override_type,
				hourly_rate: employeeOvertimeConfig.override_type === 'hourly_rate' && employeeOvertimeConfig.hourly_rate !== undefined ? employeeOvertimeConfig.hourly_rate : null,
				divisor: employeeOvertimeConfig.override_type === 'divisor' && employeeOvertimeConfig.divisor !== undefined ? employeeOvertimeConfig.divisor : null,
				rate_percentage: employeeOvertimeConfig.override_type === 'percentage' && employeeOvertimeConfig.rate_percentage !== undefined ? employeeOvertimeConfig.rate_percentage / 100 : null,
			};
			await companyApi.updateEmployeeOvertimeConfig(emp.id, payload);
			overtimeSuccess = 'Konfigurasi lembur berhasil disimpan';
			setTimeout(() => { overtimeSuccess = ''; }, 3000);
		} catch (err: any) {
			overtimeError = err.message || 'Gagal menyimpan konfigurasi lembur';
		} finally {
			isSavingOvertime = false;
		}
	}

	$effect(() => {
		const emp = employee;
		if (emp && !isLoading && hasPermission('payroll', 'read')) {
			loadSalaryComponents();
			loadEmployeeBPJSConfig();
			loadEmployeeTaxConfig();
			loadEmployeeOvertimeConfig();
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
				{#if !shouldRestrict}
				<span class="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium ring-1 ring-inset capitalize shrink-0 {getStatusBadge()}">
					{employee.employment_status}
				</span>
				{/if}
			</div>
		</div>

		<!-- Main Navigation Tabs -->
		<div class="flex border-b border-gray-200 mb-6 gap-1 overflow-x-auto pb-1.5">
			<button onclick={() => activeDetailTab = 'profile'} 
				class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold transition whitespace-nowrap cursor-pointer {activeDetailTab === 'profile' ? 'bg-indigo-50 text-indigo-600 border border-indigo-100 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100/50' }">
				👤 Profil & Pekerjaan
			</button>
			{#if hasPermission('payroll', 'read')}
				<button onclick={() => activeDetailTab = 'salary'} 
					class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold transition whitespace-nowrap cursor-pointer {activeDetailTab === 'salary' ? 'bg-indigo-50 text-indigo-600 border border-indigo-100 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100/50' }">
					💼 Komponen Gaji
				</button>
				<button onclick={() => activeDetailTab = 'bpjs'} 
					class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold transition whitespace-nowrap cursor-pointer {activeDetailTab === 'bpjs' ? 'bg-indigo-50 text-indigo-600 border border-indigo-100 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100/50' }">
					🛡️ BPJS Karyawan
				</button>
				<button onclick={() => activeDetailTab = 'tax'} 
					class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold transition whitespace-nowrap cursor-pointer {activeDetailTab === 'tax' ? 'bg-indigo-50 text-indigo-600 border border-indigo-100 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100/50' }">
					📄 Pajak PPh 21
				</button>
				<button onclick={() => activeDetailTab = 'overtime'} 
					class="flex items-center gap-2 px-4 py-2.5 rounded-lg text-xs font-bold transition whitespace-nowrap cursor-pointer {activeDetailTab === 'overtime' ? 'bg-indigo-50 text-indigo-600 border border-indigo-100 shadow-sm' : 'text-gray-500 hover:text-gray-900 hover:bg-gray-100/50' }">
					⏰ Upah Lembur
				</button>
			{/if}
		</div>

		{#if activeDetailTab === 'profile'}
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
				<div class="space-y-3.5">				{#each [
					{label:'Email',val:employee.email, show: true},
					{label:'NIK',val:employee.nik||'-', show: !shouldRestrict},
					{label:'NPWP',val:employee.npwp||'-', show: !shouldRestrict},
					{label:'Jenis Kelamin',val:getGenderLabel(employee.gender), show: true},
					{label:'Tempat, Tgl Lahir',val:`${employee.place_of_birth || '-'}${employee.date_of_birth ? `, ${formatDate(employee.date_of_birth)}` : ''}`, show: !shouldRestrict},
					{label:'Agama',val:employee.religion||'-', show: !shouldRestrict},
					{label:'Status Pernikahan',val:getMaritalStatusLabel(employee.marital_status), show: !shouldRestrict},
					{label:'No. Telepon',val:employee.phone||'-', show: true},
				].filter(i => i.show) as item (item.label)}
					<div class="flex justify-between items-start">
						<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
						<span class="text-sm text-gray-900 text-right">{item.val}</span>
					</div>
				{/each}
				{#if !shouldRestrict && employee.gender === 'perempuan' && employee.marital_status === 'menikah'}
					<div class="flex justify-between items-center pt-2 border-t border-gray-100 mt-3">
						<span class="text-xs text-gray-400 shrink-0 w-32">Sedang Hamil</span>
						<label class="relative inline-flex items-center cursor-pointer">
							<input type="checkbox" checked={isPregnant} onchange={handleTogglePregnant} disabled={isSavingPregnant} class="sr-only peer" />
							<div class="w-10 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-pink-500"></div>
							{#if isSavingPregnant}
								<svg class="w-4 h-4 ml-2 animate-spin text-gray-400" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							{/if}
						</label>
					</div>
				{/if}
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
						{label:'Posisi',val:employee.position_name||'-', show: true},
						{label:'Departemen',val:employee.department_name||'-', show: true},
						{label:'Role',val:employee.role_name||'-', show: true},
						{label:'Atasan',val:employee.approval_line_name||'-', show: true},
						{label:'Status',val:employee.employment_status||'-', show: !shouldRestrict},
						{label:'Bergabung',val:formatDate(employee.join_date), show: !shouldRestrict},
						{label:'Terakhir Login',val:formatDateTime(employee.last_login_at), show: !shouldRestrict}
					].filter(i => i.show) as item (item.label)}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right capitalize">{item.val}</span>
						</div>
					{/each}
								<!-- Approval Line Picker -->
				{#if approvalLinePickerOpen}
					<div class="bg-gray-50/50 border border-gray-200 rounded-xl p-4 mt-3">
						<div class="mb-3">
							<label for="approval-line-select" class="block text-xs font-medium text-gray-600 mb-1.5">Pilih Atasan</label>
							<select id="approval-line-select" bind:value={selectedApprovalLineId}
								class="w-full px-3 py-2 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white">
								<option value="">— Tanpa Atasan —</option>
								{#each allEmployees as emp (emp.id)}
									<option value={emp.id}>{emp.full_name} {emp.department_name ? `(${emp.department_name})` : ''}</option>
								{/each}
							</select>
						</div>
						<div class="flex items-center justify-end gap-2">
							<button onclick={closeApprovalLinePicker} class="px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-100 transition cursor-pointer">Batal</button>
							<button onclick={handleSaveApprovalLine} disabled={isSavingApprovalLine}
								class="px-4 py-1.5 bg-[#1A56DB] text-white rounded-lg text-xs font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
								{#if isSavingApprovalLine}
									<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								{/if}
								Simpan
							</button>
						</div>
					</div>
				{:else if !shouldRestrict}
					<button onclick={openApprovalLinePicker}
						class="mt-2 inline-flex items-center gap-1.5 px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-100 transition cursor-pointer">
						<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
						</svg>
						{employee.approval_line_name ? 'Ubah Atasan' : 'Atur Atasan'}
					</button>
				{/if}
				</div>
			</div>

			{#if !shouldRestrict}
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
					] as item (item)}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right">{item.val}</span>
						</div>
					{/each}
				</div>
			</div>
			{/if}

			{#if !shouldRestrict}
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
			{/if}

			{#if !shouldRestrict}
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
								{#each allWorkSchedules as ws (ws.id)}
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
			{/if}

			{#if !shouldRestrict}
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
					] as item (item)}
						<div class="flex justify-between items-start">
							<span class="text-xs text-gray-400 shrink-0 w-32">{item.label}</span>
							<span class="text-sm text-gray-900 text-right">{item.val}</span>
						</div>
					{/each}
				</div>
			</div>
			{/if}
		</div>

		{#if !shouldRestrict}
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
						{#each history as item (item)}
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
		{/if}
		{/if}

		<!-- Komponen Gaji — hanya tampil untuk user dengan akses payroll -->
		{#if hasPermission('payroll', 'read') && activeDetailTab === 'salary'}
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
			<div class="bg-white border border-gray-200 rounded-2xl shadow-sm mb-6 overflow-hidden">
				<div class="bg-white border-b border-gray-100 px-6 py-5 flex items-center justify-between">
					<div class="flex items-center gap-3.5">
						<div class="w-10 h-10 rounded-xl bg-blue-50 text-blue-600 flex items-center justify-center">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
						</div>
						<div>
							<h3 class="text-[15px] font-bold text-gray-900">Gaji Pokok & Upah Harian</h3>
							<p class="text-[13px] text-gray-500 mt-0.5">Atur nominal gaji dasar yang akan diterima karyawan.</p>
						</div>
					</div>
				</div>

				<div class="p-6">
					<div class="grid grid-cols-1 md:grid-cols-2 gap-x-8 gap-y-6">
						<div>
							<label for="salary-base" class="block text-[13px] font-semibold text-gray-700 mb-2">Gaji Pokok Bulanan</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<span class="text-gray-500 font-medium group-focus-within:text-blue-600 transition-colors">Rp</span>
								</div>
								<input id="salary-base" type="number" min="0" step="100000" bind:value={salaryBase}
									class="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm font-semibold text-gray-900 outline-none focus:ring-4 focus:ring-blue-500/15 focus:border-blue-500 focus:bg-white transition-all placeholder:text-gray-400"
									placeholder="0" />
							</div>
						</div>

						<div>
							<label for="salary-daily" class="block text-[13px] font-semibold text-gray-700 mb-2">Upah Harian</label>
							<div class="relative group">
								<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
									<span class="text-gray-500 font-medium group-focus-within:text-blue-600 transition-colors">Rp</span>
								</div>
								<input id="salary-daily" type="number" min="0" step="1000" bind:value={salaryDaily}
									class="w-full pl-12 pr-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm font-semibold text-gray-900 outline-none focus:ring-4 focus:ring-blue-500/15 focus:border-blue-500 focus:bg-white transition-all placeholder:text-gray-400"
									placeholder="0" />
							</div>
							<p class="text-[12px] text-gray-500 mt-2 flex items-center gap-1.5">
								<svg class="w-3.5 h-3.5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" /></svg>
								Hanya berlaku untuk tipe karyawan harian.
							</p>
						</div>
					</div>
				</div>

				<div class="bg-gray-50/50 px-6 py-4 flex items-center justify-between border-t border-gray-100">
					<p class="text-[13px] text-gray-500">Pastikan nominal sudah sesuai sebelum menyimpan.</p>
					<button onclick={handleSaveSalary} disabled={isSavingSalary}
						class="inline-flex items-center gap-2 px-6 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-semibold hover:bg-blue-700 transition-all active:scale-[0.98] disabled:opacity-50 shadow-sm shadow-blue-200 cursor-pointer">
						{#if isSavingSalary}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						{:else}
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
						{/if}
						Simpan Perubahan
					</button>
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
								{#each salaryComps as comp, i (comp.id)}
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
					{#each salaryComps as comp (comp)}
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

			<!-- STANDALONE: Konfigurasi BPJS Karyawan -->
			{#if hasPermission('payroll', 'read') && activeDetailTab === 'bpjs'}
			<div class="bg-white border border-gray-200 rounded-xl shadow-sm mt-6 overflow-hidden">
				<!-- Header dengan gradient -->
				<div class="bg-gradient-to-r from-indigo-50 to-indigo-100/30 px-6 py-4 border-b border-gray-200">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-600 to-indigo-800 flex items-center justify-center shadow-lg shadow-indigo-100">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.03 0 1.9.693 2.166 1.638m-7.377 2.24a.75.75 0 0 1-.75.75H7.217a.75.75 0 0 1-.75-.75 48.887 48.887 0 0 0-1.123-.08M3.181 12.008a9.75 9.75 0 1 1 19.338 0 9.75 9.75 0 0 1-19.338 0Z" />
							</svg>
						</div>
						<div>
							<h2 class="text-base font-bold text-gray-900">Konfigurasi BPJS</h2>
							<p class="text-xs text-gray-500 mt-0.5">Kelola override iuran BPJS Kesehatan dan Ketenagakerjaan karyawan ini</p>
						</div>
					</div>
				</div>

				<div class="p-6 space-y-6">
					<h3 class="text-sm font-bold text-gray-900 flex items-center gap-2">
						<span>🛡️ Iuran BPJS Karyawan</span>
						{#if bpjsSuccess}<span class="text-xs font-normal text-emerald-600 bg-emerald-50 px-2.5 py-0.5 rounded-full border border-emerald-100">{bpjsSuccess}</span>{/if}
						{#if bpjsError}<span class="text-xs font-normal text-red-600 bg-red-50 px-2.5 py-0.5 rounded-full border border-red-100">{bpjsError}</span>{/if}
					</h3>

					{#if isLoadingBPJS}
						<div class="py-4"><PulseLoader variant="text" count={2} /></div>
					{:else}
						<div class="space-y-4">
							{#each ['kesehatan', 'jht', 'jp', 'jkk', 'jkm'] as key (key)}
								{@const compKey = key as keyof EmployeeBPJSConfig}
								{#if employeeBPJSConfig[compKey]}
									<div class="border border-gray-100 rounded-xl p-4 bg-gray-50/30 flex flex-col md:flex-row md:items-center justify-between gap-4">
										<div class="min-w-0 flex-1">
											<div class="flex items-center gap-2">
												<input type="checkbox" id="bpjs-chk-{key}" bind:checked={employeeBPJSConfig[compKey]!.enabled}
													class="w-4.5 h-4.5 text-indigo-600 border-indigo-500 rounded focus:ring-indigo-500 cursor-pointer" />
												<label for="bpjs-chk-{key}" class="text-sm font-semibold text-gray-900 cursor-pointer">{BPJS_LABELS[key].label}</label>
											</div>
											<p class="text-xs text-gray-400 mt-1 pl-6.5">{BPJS_LABELS[key].description}</p>
										</div>

										{#if employeeBPJSConfig[compKey]!.enabled}
											<div class="flex flex-wrap items-center gap-4 pl-6.5 md:pl-0">
												<!-- Pekerja Override (jika ada) -->
												{#if key === 'kesehatan' || key === 'jht' || key === 'jp'}
													<div class="flex flex-col gap-1">
														<span class="text-[11px] font-bold text-gray-500 uppercase tracking-wider">Potongan Pekerja</span>
														<div class="flex items-center gap-1.5">
															<select bind:value={bpjsOverrideMethods[key].emp}
																class="px-2 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 cursor-pointer font-medium">
																<option value="default">Default</option>
																<option value="rate">Kustom %</option>
																<option value="nominal">Kustom Rp</option>
															</select>
															{#if bpjsOverrideMethods[key].emp === 'rate'}
																<div class="relative w-20">
																	<input type="number" step="0.1" min="0" max="100" bind:value={employeeBPJSConfig[compKey]!.employee_rate}
																		class="w-full pl-2 pr-5 py-1 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
																	<span class="absolute right-1.5 top-1/2 -translate-y-1/2 text-[10px] text-gray-400">%</span>
																</div>
															{:else if bpjsOverrideMethods[key].emp === 'nominal'}
																<div class="relative w-28">
																	<span class="absolute left-1.5 top-1/2 -translate-y-1/2 text-[10px] text-gray-400">Rp</span>
																	<input type="number" step="1000" min="0" bind:value={employeeBPJSConfig[compKey]!.employee_nominal}
																		class="w-full pl-6 pr-2 py-1 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
																</div>
															{/if}
														</div>
													</div>
												{/if}

												<!-- Perusahaan Override -->
												<div class="flex flex-col gap-1">
													<span class="text-[11px] font-bold text-gray-500 uppercase tracking-wider">Beban Perusahaan</span>
													<div class="flex items-center gap-1.5">
														<select bind:value={bpjsOverrideMethods[key].comp}
															class="px-2 py-1.5 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 cursor-pointer font-medium">
															<option value="default">Default</option>
															<option value="rate">Kustom %</option>
															<option value="nominal">Kustom Rp</option>
														</select>
														{#if bpjsOverrideMethods[key].comp === 'rate'}
															<div class="relative w-20">
																<input type="number" step="0.1" min="0" max="100" bind:value={employeeBPJSConfig[compKey]!.company_rate}
																	class="w-full pl-2 pr-5 py-1 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
																<span class="absolute right-1.5 top-1/2 -translate-y-1/2 text-[10px] text-gray-400">%</span>
															</div>
														{:else if bpjsOverrideMethods[key].comp === 'nominal'}
															<div class="relative w-28">
																<span class="absolute left-1.5 top-1/2 -translate-y-1/2 text-[10px] text-gray-400">Rp</span>
																<input type="number" step="1000" min="0" bind:value={employeeBPJSConfig[compKey]!.company_nominal}
																	class="w-full pl-6 pr-2 py-1 bg-white border border-gray-200 rounded-lg text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
															</div>
														{/if}
													</div>
												</div>
											</div>
										{/if}
									</div>
								{/if}
							{/each}
							<div class="flex justify-end pt-2">
								<button onclick={handleSaveBPJSConfig} disabled={isSavingBPJS}
									class="px-5 py-2 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
									{#if isSavingBPJS}
										<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
									{/if}
									Simpan Override BPJS
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
			{/if}

			<!-- STANDALONE: Konfigurasi Pajak PPh 21 Karyawan -->
			{#if hasPermission('payroll', 'read') && activeDetailTab === 'tax'}
			<div class="bg-white border border-gray-200 rounded-xl shadow-sm mt-6 overflow-hidden">
				<!-- Header dengan gradient -->
				<div class="bg-gradient-to-r from-indigo-50 to-indigo-100/30 px-6 py-4 border-b border-gray-200">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-600 to-indigo-800 flex items-center justify-center shadow-lg shadow-indigo-100">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.03 0 1.9.693 2.166 1.638m-7.377 2.24a.75.75 0 0 1-.75.75H7.217a.75.75 0 0 1-.75-.75 48.887 48.887 0 0 0-1.123-.08M3.181 12.008a9.75 9.75 0 1 1 19.338 0 9.75 9.75 0 0 1-19.338 0Z" />
							</svg>
						</div>
						<div>
							<h2 class="text-base font-bold text-gray-900">Konfigurasi Pajak PPh 21</h2>
							<p class="text-xs text-gray-500 mt-0.5">Kelola override potongan Pajak Penghasilan (PPh 21) karyawan ini</p>
						</div>
					</div>
				</div>

				<div class="p-6 space-y-6">
					<h3 class="text-sm font-bold text-gray-900 flex items-center gap-2">
						<span>📄 Potongan Pajak PPh 21</span>
						{#if taxSuccess}<span class="text-xs font-normal text-emerald-600 bg-emerald-50 px-2.5 py-0.5 rounded-full border border-emerald-100">{taxSuccess}</span>{/if}
						{#if taxError}<span class="text-xs font-normal text-red-600 bg-red-50 px-2.5 py-0.5 rounded-full border border-red-100">{taxError}</span>{/if}
					</h3>

					{#if isLoadingTax}
						<div class="py-4"><PulseLoader variant="text" count={1} /></div>
					{:else}
						<div class="space-y-4">
							<div class="border border-gray-100 rounded-xl p-4 bg-gray-50/30 flex flex-col md:flex-row md:items-center justify-between gap-4">
								<div class="min-w-0 flex-1">
									<span class="text-sm font-semibold text-gray-900">Metode Pengenaan Pajak</span>
									<p class="text-xs text-gray-400 mt-1">Secara default, pajak dihitung otomatis menggunakan Kategori TER berdasarkan status PTKP ({employee.ptkp_status || 'TK0'}). Anda dapat memaksa rate persen tertentu atau nilai nominal tetap.</p>
								</div>

								<div class="flex items-center gap-3 self-start md:self-center">
									<select bind:value={employeeTaxConfig.override_type}
										class="px-3 py-2 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 font-medium cursor-pointer font-semibold">
										<option value="none">Otomatis TER (PTKP)</option>
										<option value="rate">Kustom Persentase (%)</option>
										<option value="nominal">Kustom Nominal Tetap (Rp)</option>
										<option value="free">Bebas Pajak (0%)</option>
									</select>

									{#if employeeTaxConfig.override_type === 'rate'}
										<div class="relative w-24">
											<input type="number" step="0.1" min="0" max="100" bind:value={employeeTaxConfig.override_rate}
												class="w-full pl-3 pr-6 py-1.5 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
											<span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">%</span>
										</div>
									{:else if employeeTaxConfig.override_type === 'nominal'}
										<div class="relative w-36">
											<span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-xs text-gray-400">Rp</span>
											<input type="number" step="5000" min="0" bind:value={employeeTaxConfig.override_nominal}
												class="w-full pl-8 pr-3 py-1.5 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
										</div>
									{/if}
								</div>
							</div>
							<div class="flex justify-end pt-3">
								<button onclick={handleSaveTaxConfig} disabled={isSavingTax}
									class="px-5 py-2 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
									{#if isSavingTax}
										<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
									{/if}
									Simpan Override Pajak
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
			{/if}

			<!-- STANDALONE: Konfigurasi Upah Lembur Karyawan -->
			{#if hasPermission('payroll', 'read') && activeDetailTab === 'overtime'}
			<div class="bg-white border border-gray-200 rounded-xl shadow-sm mt-6 overflow-hidden">
				<!-- Header dengan gradient -->
				<div class="bg-gradient-to-r from-indigo-50 to-indigo-100/30 px-6 py-4 border-b border-gray-200">
					<div class="flex items-center gap-3">
						<div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-600 to-indigo-800 flex items-center justify-center shadow-lg shadow-indigo-100">
							<svg class="w-5 h-5 text-white" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.03 0 1.9.693 2.166 1.638m-7.377 2.24a.75.75 0 0 1-.75.75H7.217a.75.75 0 0 1-.75-.75 48.887 48.887 0 0 0-1.123-.08M3.181 12.008a9.75 9.75 0 1 1 19.338 0 9.75 9.75 0 0 1-19.338 0Z" />
							</svg>
						</div>
						<div>
							<h2 class="text-base font-bold text-gray-900">Konfigurasi Upah Lembur</h2>
							<p class="text-xs text-gray-500 mt-0.5">Kelola rate upah lembur dasar per jam karyawan ini</p>
						</div>
					</div>
				</div>

				<div class="p-6 space-y-6">
					<h3 class="text-sm font-bold text-gray-900 flex items-center gap-2">
						<span>⏰ Konfigurasi Upah Lembur</span>
						{#if overtimeSuccess}<span class="text-xs font-normal text-emerald-600 bg-emerald-50 px-2.5 py-0.5 rounded-full border border-emerald-100">{overtimeSuccess}</span>{/if}
						{#if overtimeError}<span class="text-xs font-normal text-red-600 bg-red-50 px-2.5 py-0.5 rounded-full border border-red-100">{overtimeError}</span>{/if}
					</h3>

					{#if isLoadingOvertime}
						<div class="py-4"><PulseLoader variant="text" count={1} /></div>
					{:else}
						<div class="space-y-4">
							<div class="border border-gray-100 rounded-xl p-4 bg-gray-50/30 flex flex-col md:flex-row md:items-center justify-between gap-4">
								<div class="min-w-0 flex-1">
									<span class="text-sm font-semibold text-gray-900">Perhitungan Rate per Jam</span>
									<p class="text-xs text-gray-400 mt-1">
										Secara default, upah lembur per jam dihitung dengan formula <strong>Gaji Pokok / 173</strong>. Anda dapat menetapkan nominal tetap langsung per jam, mengubah pembagi (divisor) gaji pokok, atau menetapkan persentase dari gaji pokok.
									</p>
								</div>

								<div class="flex items-center gap-3 self-start md:self-center">
									<select bind:value={employeeOvertimeConfig.override_type}
										class="px-3 py-2 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 font-medium cursor-pointer font-semibold">
										<option value="none">Default (Gaji Pokok / 173)</option>
										<option value="hourly_rate">Kustom Nominal per Jam (Rp)</option>
										<option value="divisor">Kustom Pembagi Gaji Pokok (Divisor)</option>
										<option value="percentage">Kustom Persentase Gaji Pokok (%)</option>
									</select>

									{#if employeeOvertimeConfig.override_type === 'hourly_rate'}
										<div class="relative w-36">
											<span class="absolute left-2.5 top-1/2 -translate-y-1/2 text-xs text-gray-400">Rp</span>
											<input type="number" step="1000" min="0" bind:value={employeeOvertimeConfig.hourly_rate}
												class="w-full pl-8 pr-3 py-1.5 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
										</div>
									{:else if employeeOvertimeConfig.override_type === 'divisor'}
										<div class="relative w-24">
											<input type="number" step="1" min="1" bind:value={employeeOvertimeConfig.divisor}
												class="w-full pl-3 pr-3 py-1.5 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
										</div>
									{:else if employeeOvertimeConfig.override_type === 'percentage'}
										<div class="relative w-24">
											<input type="number" step="0.01" min="0" max="100" bind:value={employeeOvertimeConfig.rate_percentage}
												class="w-full pl-3 pr-6 py-1.5 bg-white border border-gray-200 rounded-xl text-xs outline-none focus:ring-1 focus:ring-indigo-500 text-right font-semibold" />
											<span class="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-gray-400">%</span>
										</div>
									{/if}
								</div>
							</div>
							<div class="flex justify-end pt-3">
								<button onclick={handleSaveOvertimeConfig} disabled={isSavingOvertime}
									class="px-5 py-2 bg-indigo-600 text-white rounded-lg text-xs font-semibold hover:bg-indigo-700 transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
									{#if isSavingOvertime}
										<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
									{/if}
									Simpan Override Lembur
								</button>
							</div>
						</div>
					{/if}
				</div>
			</div>
			{/if}

		<!-- Delete Confirmation Modal for Salary Component -->
		<AnimatedPresence show={showSCDelete} type="scale" duration={200}>
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
