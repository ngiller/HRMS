<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { goto } from '$app/navigation';
	import { employees as employeesApi, departments as deptApi, roles as rolesApi, positions as positionsApi, positionGrades as gradesApi, ApiError } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import { hasPermission } from '$lib/permissions.js';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams, Position, PositionGrade } from '$lib/types.js';
	type Employee = {
		id: string;
		employee_id: string;
		full_name: string;
		email: string;
		gender: string;
		employment_status: string;
		is_active: boolean;
		deleted_at: string | null;
		role_name: string;
		position_name: string;
		department_name: string;
		base_salary: number;
		join_date: string;
		phone: string;
		place_of_birth: string;
		date_of_birth: string;
		religion: string;
		marital_status: string;
		address: string;
		nik: string;
		npwp: string;
		role_id: string;
		position_id: string;
		department_id: string;
		bank_name: string;
		bank_account: string;
		address_ktp: string;
	};

	type EmployeeForm = {
		employee_id: string;
		full_name: string;
		email: string;
		password: string;
		gender: string;
		place_of_birth: string;
		date_of_birth: string;
		religion: string;
		marital_status: string;
		join_date: string;
		employment_status: string;
		phone: string;
		address: string;
		nik: string;
		npwp: string;
		bank_name: string;
		bank_account: string;
		address_ktp: string;
		role_id: string;
		position_id: string;
		department_id: string;
	};

	type Department = {
		id: string;
		name: string;
		code: string;
	};

	type Role = {
		id: string;
		name: string;
		slug: string;
	};

	let employees = $state<Employee[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let searchQuery = $state('');
	let filterDepartment = $state('');
	let filterStatus = $state('');
	let showDeleted = $state(false);
	let isLoading = $state(true);
	let errorMessage = $state('');

	// Summary stats
	let totalActive = $state(0);
	let totalKontrak = $state(0);
	let totalPercobaan = $state(0);

	// Form state — inline gantikan tabel
	let showForm = $state(false);
	let formTitle = $state('');
	let editingId = $state<string | null>(null);
	let form = $state<EmployeeForm>({
		employee_id: '', full_name: '', email: '', password: '',
		gender: 'laki_laki', place_of_birth: '', date_of_birth: '', religion: 'islam', marital_status: 'belum_menikah',
		join_date: '', employment_status: 'percobaan',
		phone: '', address: '',
		nik: '', npwp: '', bank_name: '', bank_account: '', address_ktp: '',
		role_id: '', position_id: '', department_id: '',
	});
	let formError = $state('');
	let isSaving = $state(false);

	// Delete confirm
	let showDeleteConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');

	// Dropdown data
	let departments = $state<Department[]>([]);
	let roles = $state<Role[]>([]);
	let positions = $state<Position[]>([]);
	let positionGrades = $state<PositionGrade[]>([]);

	let selectedPositionGradeInfo = $derived.by(() => {
		if (!form.position_id || positionGrades.length === 0) return null;
		const pos = positions.find(p => p.id === form.position_id);
		if (!pos || !(pos as any).grade_id) return null;
		return positionGrades.find((g: any) => g.id === (pos as any).grade_id) || null;
	});

	onMount(async () => { const m = await getAgGrid(); agGridModule = m;
		loadEmployees();
		loadDropdowns();
	});

	async function loadDropdowns() {
		try {
			const [deptResp, roleResp, posResp, gradeResp] = await Promise.all([
				deptApi.getAll(),
				rolesApi.list(1, 100),
				positionsApi.getAll(),
				gradesApi.getAll(),
			]);
			departments = deptResp.data || [];
			roles = (roleResp.data || []).filter((r: any) => r.slug !== 'super_admin');
			positions = posResp.data || [];
			positionGrades = gradeResp.data || [];
		} catch {
			// silent
		}
	}

	async function loadEmployees() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response: ApiResponse<Employee[]> = await employeesApi.list(page, perPage, searchQuery, filterDepartment, filterStatus, showDeleted) as ApiResponse<Employee[]>;
			const data = response.data || [];
			employees = data;
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);

			totalActive = data.filter((e: Employee) => e.is_active && !e.deleted_at).length;
			totalKontrak = data.filter((e: Employee) => e.employment_status?.toLowerCase() === 'kontrak').length;
			totalPercobaan = data.filter((e: Employee) => e.employment_status?.toLowerCase() === 'percobaan').length;
		} catch (error: unknown) {
			errorMessage = (error as { message?: string }).message || 'Gagal memuat data karyawan';
			console.error('Employee list error:', error);
		} finally {
			isLoading = false;
		}
	}

	function goToPage(p: number) {
		if (p < 1 || p > totalPages) return;
		page = p;
		loadEmployees();
	}

	function viewDetail(id: string) {
		goto('/karyawan/detail', { state: { employeeId: id } });
	}

	function openCreateForm() {
		formTitle = 'Tambah Karyawan';
		editingId = null;
		form = {
			employee_id: '', full_name: '', email: '', password: '',
			gender: 'laki_laki', place_of_birth: '', date_of_birth: '',
			religion: 'islam', marital_status: 'lajang',
			join_date: new Date().toISOString().split('T')[0],
			employment_status: 'percobaan', phone: '', address: '',
			nik: '', npwp: '', bank_name: '', bank_account: '', address_ktp: '',
			role_id: '', position_id: '', department_id: '',
		};
		formError = '';
		showForm = true;
	}

	function openEditForm(emp: Employee) {
		formTitle = 'Edit Karyawan';
		editingId = emp.id;
		form = {
			employee_id: emp.employee_id, full_name: emp.full_name, email: emp.email,
			password: '', gender: emp.gender, place_of_birth: '', date_of_birth: '',
			religion: 'islam', marital_status: 'belum_menikah',
			join_date: typeof emp.join_date === 'string' ? emp.join_date.split('T')[0] : '',
			employment_status: emp.employment_status, phone: emp.phone,
			address: '', nik: '', npwp: '', bank_name: '', bank_account: '', address_ktp: '',
			role_id: '', position_id: '', department_id: '',
		};
		formError = '';
		showForm = true;

		// Load full detail
		employeesApi.get(emp.id).then((resp: any) => {
			if (resp.data) {
				const d = resp.data;
				form.employee_id = d.employee_id || '';
				form.gender = d.gender || 'laki_laki';
				form.place_of_birth = d.place_of_birth || '';
				form.date_of_birth = d.date_of_birth ? d.date_of_birth.split('T')[0] : '';
				form.religion = d.religion || 'islam';
				form.marital_status = d.marital_status || 'belum_menikah';
				form.join_date = d.join_date ? d.join_date.split('T')[0] : '';
				form.phone = d.phone || '';
				form.address = d.address || '';
				form.nik = d.nik || '';
				form.npwp = d.npwp || '';
				form.bank_name = d.bank_name || '';
				form.bank_account = d.bank_account || '';
				form.address_ktp = d.address_ktp || '';
				form.role_id = d.role_id || '';
				form.position_id = d.position_id || '';
				form.department_id = d.department_id || '';
			}
		}).catch((err: unknown) => {
			console.error('Gagal memuat detail karyawan:', err);
		});
	}

	function cancelForm() {
		showForm = false;
		formError = '';
	}

	async function handleSave() {
		if (!form.full_name.trim()) { formError = 'Nama lengkap harus diisi'; return; }
		if (!form.email.trim()) { formError = 'Email harus diisi'; return; }
		if (!editingId && !form.password.trim()) { formError = 'Password harus diisi'; return; }

		isSaving = true;
		formError = '';
		try {
			if (editingId) {
				const payload: Record<string, unknown> = {
					full_name: form.full_name.trim(),
					email: form.email.trim(),
					gender: form.gender,
					join_date: form.join_date,
					employment_status: form.employment_status,
				};
				if (form.place_of_birth) payload.place_of_birth = form.place_of_birth;
				if (form.date_of_birth) payload.date_of_birth = form.date_of_birth;
				if (form.religion) payload.religion = form.religion;
				if (form.marital_status) payload.marital_status = form.marital_status;

				if (form.phone) payload.phone = form.phone;
				if (form.address) payload.address = form.address;
				if (form.nik) payload.nik = form.nik;
				if (form.npwp) payload.npwp = form.npwp;
				if (form.bank_name) payload.bank_name = form.bank_name;
				if (form.bank_account) payload.bank_account = form.bank_account;
				if (form.address_ktp) payload.address_ktp = form.address_ktp;
				if (form.role_id) payload.role_id = form.role_id;
				if (form.position_id) payload.position_id = form.position_id;
				if (form.department_id) payload.department_id = form.department_id;
				await employeesApi.update(editingId, payload);
			} else {
				await employeesApi.create({
					employee_id: form.employee_id.trim() || undefined,
					full_name: form.full_name.trim(),
					email: form.email.trim(),
					password: form.password,
					gender: form.gender,
					place_of_birth: form.place_of_birth || undefined,
					date_of_birth: form.date_of_birth || undefined,
					religion: form.religion || undefined,
					marital_status: form.marital_status || undefined,
					join_date: form.join_date,
					employment_status: form.employment_status,
					phone: form.phone || undefined,
					address: form.address || undefined,
					nik: form.nik || undefined,
					npwp: form.npwp || undefined,
					bank_name: form.bank_name || undefined,
					bank_account: form.bank_account || undefined,
					address_ktp: form.address_ktp || undefined,
					role_id: form.role_id || undefined,
					position_id: form.position_id || undefined,
					department_id: form.department_id || undefined,
				});
			}
			cancelForm();
			loadEmployees();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menyimpan karyawan';
		} finally {
			isSaving = false;
		}
	}

	function confirmDelete(id: string, name: string) {
		deletingId = id;
		deletingName = name;
		showDeleteConfirm = true;
	}

	function cancelDelete() {
		showDeleteConfirm = false;
		deletingId = null;
		deletingName = '';
	}

	async function handleDelete() {
		if (!deletingId) return;
		isSaving = true;
		try {
			await employeesApi.remove(deletingId);
			showDeleteConfirm = false;
			deletingId = null;
			deletingName = '';
			loadEmployees();
		} catch (error: unknown) {
			formError = (error as { message?: string }).message || 'Gagal menghapus karyawan';
			showDeleteConfirm = false;
		} finally {
			isSaving = false;
		}
	}

const employmentStatusColors: Record<string, string> = {
    tetap: 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800',
    kontrak: 'bg-blue-50 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-200 dark:ring-blue-800',
    percobaan: 'bg-amber-50 text-amber-700 ring-amber-200 dark:bg-amber-900 dark:text-amber-200 dark:ring-amber-800',
    magang: 'bg-purple-50 text-purple-700 ring-purple-200 dark:bg-purple-900 dark:text-purple-200 dark:ring-purple-800',
    harian: 'bg-cyan-50 text-cyan-700 ring-cyan-200 dark:bg-cyan-900 dark:text-cyan-200 dark:ring-cyan-800',
};


function getStatusBadge(status: string): string {
    const cls = employmentStatusColors[status?.toLowerCase()] || 'bg-gray-50 dark:bg-gray-800 text-gray-600 dark:text-gray-400 dark:bg-gray-900 dark:text-gray-300';
    const label = status?.charAt(0).toUpperCase() + status?.slice(1);
    return `<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 ${cls}">${label}</span>`;
};

	function isDeleted(emp: Employee): boolean {
		return !!emp.deleted_at;
	}

	// Restore
	let restoringId = $state<string | null>(null);

	async function handleRestore(id: string) {
		restoringId = id;
		try {
			await employeesApi.restore(id);
			loadEmployees();
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal mengaktifkan kembali karyawan';
		} finally {
			restoringId = null;
		}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	// Export/Import
	let isExporting = $state(false);
	let isImporting = $state(false);
	let importResult = $state<{ success: number; errors: string[]; message: string } | null>(null);

	async function handleExport() {
		isExporting = true;
		try {
			const blob = await employeesApi.exportExcel();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'karyawan.xlsx';
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal export data';
		} finally {
			isExporting = false;
		}
	}

	async function handleImport(e: Event) {
		const target = e.target as HTMLInputElement;
		const file = target.files?.[0];
		if (!file) return;

		isImporting = true;
		importResult = null;
		try {
			const res: any = await employeesApi.importExcel(file);
			importResult = res.data || { success: 0, message: 'Selesai' };
			loadEmployees();
		} catch (err: unknown) {
			importResult = { success: 0, errors: [(err as { message?: string }).message || 'Terjadi kesalahan'], message: 'Gagal import' };
		} finally {
			isImporting = false;
			target.value = '';
		}
	}

	// ── AG Grid ──
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	const defaultColDef: ColDef = {
		sortable: true,
		resizable: true,
		filter: true,
		floatingFilter: false,
	};

	// SVG icon helpers
	function iconView(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.036 12.322a1.012 1.012 0 0 1 0-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /></svg>';
	}
	function iconEdit(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>';
	}
	function iconDelete(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>';
	}
	function iconRestore(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>';
	}

	// DOM-based button creator for AG Grid cell renderer
	function createActionButton(html: string, className: string, ariaLabel: string, onClick: () => void): HTMLButtonElement {
		const btn = document.createElement('button');
		btn.innerHTML = html;
		btn.className = className;
		btn.setAttribute('aria-label', ariaLabel);
		btn.onclick = (e) => {
			e.stopPropagation();
			onClick();
		};
		return btn;
	}

	const columnDefs: ColDef[] = [
		{
			field: 'full_name', headerName: 'Karyawan', minWidth: 220, flex: 1,
			valueGetter: (params) => params.data?.full_name || '',
			cellRenderer: (params: AgGridCellParams<Employee>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const email = params.data?.email || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 flex items-center justify-center text-xs font-semibold text-gray-600 dark:text-gray-400 shrink-0 ring-1 ring-gray-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900 dark:text-white">${params.value}</div><div class="text-xs text-gray-400">${email}</div></div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{ field: 'position_name', headerName: 'Posisi', minWidth: 150, headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider', cellClass: 'text-sm text-gray-600 dark:text-gray-400' },
		{ field: 'department_name', headerName: 'Departemen', minWidth: 150, headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider', cellClass: 'text-sm text-gray-600 dark:text-gray-400' },		{
			field: 'base_salary', headerName: 'Gaji Pokok', minWidth: 140,
			hide: !hasPermission('payroll', 'read'),
			valueFormatter: (params: AgGridValueParams) => (params.value as number) > 0 ? new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(params.value as number) : '-',
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-700 dark:text-gray-300 tabular-nums text-right',
			type: 'rightAligned',
		},
		{ field: 'join_date', headerName: 'Bergabung', minWidth: 130,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500 dark:text-gray-400',
		},
		{
			field: 'employment_status', headerName: 'Status', minWidth: 120,
			cellRenderer: (params: AgGridCellParams<Employee>) => {
				const emp = params.data;
				if (!emp) return '';
				if (emp.deleted_at) {
					return '<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ring-1 ring-inset bg-red-50 text-red-700 ring-red-600/20">Dihapus</span>';
				}
				return getStatusBadge(emp.employment_status);
			},
			headerClass: 'text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider',
		},
		{
			field: 'id', headerName: '', minWidth: 140, maxWidth: 140,
			cellRenderer: (params: AgGridCellParams<Employee>) => {
				const emp = params.data;
				if (!emp) return '';

				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';

				const isDel = !!emp.deleted_at;

				// View button — always visible
				const viewBtn = createActionButton(iconView(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Lihat detail',
					() => viewDetail(emp.employee_id)
				);
				container.appendChild(viewBtn);

				if (isDel && hasPermission('employee', 'update')) {
					// Restore button for deleted employees
					const restoreBtn = createActionButton(iconRestore(),
						'p-1.5 rounded-lg text-emerald-500 hover:text-emerald-700 hover:bg-emerald-50 transition cursor-pointer',
						'Aktifkan kembali',
						() => handleRestore(emp.id)
					);
					container.appendChild(restoreBtn);
				} else {
					if (hasPermission('employee', 'update')) {
						const editBtn = createActionButton(iconEdit(),
							'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
							'Edit karyawan',
							() => openEditForm(emp)
						);
						container.appendChild(editBtn);
					}
					if (hasPermission('employee', 'delete')) {
						const deleteBtn = createActionButton(iconDelete(),
							'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
							'Nonaktifkan karyawan',
							() => confirmDelete(emp.id, emp.full_name)
						);
						container.appendChild(deleteBtn);
					}
				}

				return container;
			},
			sortable: false, filter: false, resizable: false,
		}
	];

	const gridOptions: GridOptions = {
		columnDefs,
		defaultColDef,
		rowHeight: 56,
		headerHeight: 44,
		animateRows: true,
		domLayout: 'autoHeight',
		suppressDragLeaveHidesColumns: true,
		suppressRowHoverHighlight: false,
		enableCellTextSelection: true,
		pagination: false, // we use our own pagination
		theme: 'legacy',
		onGridReady: (params) => {
			gridApi = params.api;
		},
	};

	$effect(() => {
		if (showForm && gridApi) {
			gridApi.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (employees.length > 0 && gridContainer && !showForm) {
			if (!gridApi && agGridModule) {
				gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi;
			}
			gridApi?.updateGridOptions({ rowData: employees });
		}
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
	});

	// Debounce search
	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			searchQuery = target.value;
			page = 1;
			loadEmployees();
		}, 400);
	}
</script>

<div class="w-full">
	<!-- Header Section -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Karyawan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Kelola seluruh data karyawan perusahaan</p>
		</div>
		{#if !showForm}
			<div class="flex items-center gap-2 flex-wrap">
				{#if hasPermission('employee', 'create')}
					<button
						onclick={openCreateForm}
						class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer"
					>
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
						</svg>
						Tambah Karyawan
					</button>
				{/if}
				{#if hasPermission('employee', 'read')}
					<button onclick={handleExport} disabled={isExporting}
						class="inline-flex items-center gap-2 px-4 py-2.5 border border-gray-200 dark:border-gray-800 text-gray-700 dark:text-gray-300 rounded-xl text-sm font-semibold hover:bg-gray-50 dark:bg-gray-800 transition shadow-sm cursor-pointer disabled:opacity-50">
						{#if isExporting}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						{:else}
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
						{/if}
						{isExporting ? 'Mengexport...' : 'Export Excel'}
					</button>
					<label class="inline-flex items-center gap-2 px-4 py-2.5 border border-gray-200 dark:border-gray-800 text-gray-700 dark:text-gray-300 rounded-xl text-sm font-semibold hover:bg-gray-50 dark:bg-gray-800 transition shadow-sm cursor-pointer disabled:opacity-50">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5m-13.5-9L12 3m0 0 4.5 4.5M12 3v13.5" /></svg>
						<span>{isImporting ? 'Mengimport...' : 'Import Excel'}</span>
						<input type="file" accept=".xlsx,.xls" onchange={handleImport} class="hidden" disabled={isImporting} />
					</label>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Import Result Notification -->
	{#if importResult}
		<div class="border rounded-xl px-5 py-3.5 mb-4 flex items-start justify-between gap-4 {importResult.success > 0 ? 'bg-emerald-50 border-emerald-200' : 'bg-red-50 border-red-200'}">
			<div class="flex items-start gap-3">
				{#if importResult.success > 0}
					<svg class="w-5 h-5 text-emerald-500 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
				{:else}
					<svg class="w-5 h-5 text-red-500 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				{/if}
				<div>
					<p class="text-sm font-medium {importResult.success > 0 ? 'text-emerald-800' : 'text-red-800'}">{importResult.message}</p>
					{#if importResult.errors && importResult.errors.length > 0}
						<ul class="mt-1 space-y-0.5">
							{#each importResult.errors as err}
								<li class="text-xs text-red-600">{err}</li>
							{/each}
						</ul>
					{/if}
				</div>
			</div>
			<button onclick={() => { importResult = null; }}
				class="p-1 rounded-lg text-gray-400 hover:text-gray-600 dark:text-gray-400 hover:bg-white/50 dark:bg-gray-900/50 transition cursor-pointer"
				aria-label="Tutup hasil import">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
			</button>
		</div>
	{/if}

	<!-- Summary Stats (sembunyikan saat form aktif) -->
	<div class="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-6" class:hidden={showForm}>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Total</span>
					<div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{isLoading ? '-' : total}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Aktif</span>
					<div class="w-7 h-7 rounded-lg bg-emerald-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{isLoading ? '-' : totalActive}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Kontrak</span>
					<div class="w-7 h-7 rounded-lg bg-blue-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{isLoading ? '-' : totalKontrak}</p>
			</div>
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-4 py-3.5">
				<div class="flex items-center justify-between">
					<span class="text-xs font-medium text-gray-500 dark:text-gray-400">Percobaan</span>
					<div class="w-7 h-7 rounded-lg bg-amber-50 flex items-center justify-center">
						<svg class="w-3.5 h-3.5 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					</div>
				</div>
				<p class="text-xl font-bold text-gray-900 dark:text-white mt-1 tabular-nums">{isLoading ? '-' : totalPercobaan}</p>
			</div>
		</div>

		<!-- Search & Filter Bar -->
		{#if !showForm}
		<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center gap-3">
			<div class="relative flex-1 min-w-0 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="search"
					value={searchQuery}
					placeholder="Cari berdasarkan nama atau email..."
					oninput={onSearchInput}
					class="w-full pl-9 pr-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white dark:bg-gray-900 transition placeholder:text-gray-400"
					aria-label="Cari karyawan"
				/>
			</div>
			<select bind:value={filterDepartment} onchange={() => { page = 1; loadEmployees(); }}
				class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition">
				<option value="">Semua Departemen</option>
				{#each departments as d}
					<option value={d.id}>{d.name}</option>
				{/each}
			</select>
			<select bind:value={filterStatus} onchange={() => { page = 1; loadEmployees(); }}
			class="px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition">
				<option value="">Semua Status</option>
				{#each Object.keys(employmentStatusColors) as st}
					<option value={st}>{st.charAt(0).toUpperCase() + st.slice(1)}</option>
				{/each}
			</select>
			<label class="inline-flex items-center gap-2 text-sm text-gray-600 dark:text-gray-400 cursor-pointer select-none">
				<input type="checkbox" bind:checked={showDeleted} onchange={() => { page = 1; loadEmployees(); }}
					class="w-4 h-4 rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]" />
				<span class="text-xs">Tampilkan nonaktif</span>
			</label>
			<div class="flex items-center gap-2 text-xs text-gray-400 whitespace-nowrap">
				{total > 0 ? `${total} karyawan ditemukan` : ''}
			</div>
		</div>
	{/if}

	<!-- Inline Form (gantikan tabel + stats + search) -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl shadow-sm transition-all duration-300 transform {showForm ? 'opacity-100 translate-y-0 relative' : 'opacity-0 translate-y-4 hidden'}" class:hidden={!showForm}>
			<!-- Form Header -->
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
				<h2 class="text-lg font-semibold text-gray-900 dark:text-white">{formTitle}</h2>
				<button
					onclick={cancelForm}
					class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:bg-gray-800 transition cursor-pointer"
					aria-label="Tutup"
				>
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
					</svg>
				</button>
			</div>

			<!-- Form Body -->
			<div class="px-6 py-5 space-y-8">
				{#if formError}
					<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>
				{/if}

				<!-- Seksi 1: Informasi Pribadi -->
				<section>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 pb-2 border-b border-gray-100">Informasi Pribadi</h3>
					<div class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									NIK / NIP
									<input type="text" bind:value={form.employee_id}
										disabled={!!editingId}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400 disabled:bg-gray-50 dark:bg-gray-800 disabled:text-gray-400"
										placeholder="Auto-generate jika kosong" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Nama Lengkap <span class="text-red-500">*</span>
									<input type="text" bind:value={form.full_name}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Contoh: John Doe" />
								</label>
							</div>
						</div>
						
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Email <span class="text-red-500">*</span>
									<input type="email" bind:value={form.email}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Contoh: john@company.com" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Password {#if !editingId}<span class="text-red-500">*</span>{/if}
									<input type="password" bind:value={form.password}
										disabled={!!editingId}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400 disabled:bg-gray-50 dark:bg-gray-800 disabled:text-gray-400"
										placeholder={editingId ? '(tidak diubah)' : 'Min. 6 karakter'} />
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									No. Telepon
									<input type="tel" bind:value={form.phone}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Contoh: 08123456789" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Jenis Kelamin <span class="text-red-500">*</span>
									<select bind:value={form.gender}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="laki_laki">Laki-laki</option>
										<option value="perempuan">Perempuan</option>
									</select>
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Tempat Lahir
									<input type="text" bind:value={form.place_of_birth}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Contoh: Jakarta" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Tanggal Lahir
									<input type="date" bind:value={form.date_of_birth}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Agama
									<select bind:value={form.religion}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="islam">Islam</option>
										<option value="kristen">Kristen</option>
										<option value="katolik">Katolik</option>
										<option value="hindu">Hindu</option>
										<option value="buddha">Buddha</option>
										<option value="konghucu">Konghucu</option>
										<option value="lainnya">Lainnya</option>
									</select>
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Status Pernikahan
									<select bind:value={form.marital_status}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="belum_menikah">Belum Menikah</option>
										<option value="menikah">Menikah</option>
										<option value="cerai_hidup">Cerai Hidup</option>
										<option value="cerai_mati">Cerai Mati</option>
									</select>
								</label>
							</div>
						</div>
					</div>
				</section>

				<!-- Seksi 2: Data Pekerjaan -->
				<section>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 pb-2 border-b border-gray-100">Data Pekerjaan</h3>
					<div class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Tanggal Bergabung <span class="text-red-500">*</span>
									<input type="date" bind:value={form.join_date}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Status Karyawan <span class="text-red-500">*</span>
									<select bind:value={form.employment_status}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="percobaan">Percobaan</option>
										<option value="kontrak">Kontrak</option>
										<option value="tetap">Tetap</option>
										<option value="magang">Magang</option>
										<option value="harian">Harian</option>
									</select>
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Departemen
									<select bind:value={form.department_id}
										onchange={() => { form.position_id = ''; }}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="">Pilih departemen (opsional)</option>
										{#each departments as d}
											<option value={d.id}>{d.name} ({d.code})</option>
										{/each}
									</select>
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Posisi
									<select bind:value={form.position_id}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="">Pilih posisi (opsional)</option>
										{#each positions.filter(p => !form.department_id || p.department_id === form.department_id) as p}
											<option value={p.id}>{p.name}</option>
										{/each}
									</select>
								</label>
							</div>
						</div>

						{#if selectedPositionGradeInfo}
							<div class="bg-gradient-to-r from-purple-50 to-blue-50 dark:from-purple-900/20 dark:to-blue-900/20 border border-purple-200 dark:border-purple-800 rounded-lg px-4 py-3 flex items-center gap-3">
								<svg class="w-5 h-5 text-purple-600 dark:text-purple-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33" />
								</svg>
								<div class="text-xs text-gray-700 dark:text-gray-300">
									<span class="font-semibold">{selectedPositionGradeInfo.name}</span>
									<span class="text-gray-400 mx-1.5">—</span>
									Rentang Gaji Golongan:
									<span class="font-semibold tabular-nums">
										{selectedPositionGradeInfo.min_salary ? new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(selectedPositionGradeInfo.min_salary) : '—'}
										<span class="text-gray-400 mx-1">—</span>
										{selectedPositionGradeInfo.max_salary ? new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(selectedPositionGradeInfo.max_salary) : '—'}
									</span>
								</div>
							</div>
						{/if}

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Role Sistem
									<select bind:value={form.role_id}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-900">
										<option value="">Pilih role (opsional)</option>
										{#each roles as r}
											<option value={r.id}>{r.name}</option>
										{/each}
									</select>
								</label>
							</div>
							<div></div>
						</div>
					</div>
				</section>

				<!-- Seksi 3: Administratif & Keuangan -->
				<section>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-4 pb-2 border-b border-gray-100">Administratif & Keuangan</h3>
					<div class="space-y-4">
						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									NIK KTP
									<input type="text" bind:value={form.nik}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Nomor Induk Kependudukan" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									NPWP
									<input type="text" bind:value={form.npwp}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Nomor Pokok Wajib Pajak" />
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Nama Bank
									<input type="text" bind:value={form.bank_name}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Contoh: BCA" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									No. Rekening
									<input type="text" bind:value={form.bank_account}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Nomor rekening bank" />
								</label>
							</div>
						</div>

						<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Alamat KTP
									<input type="text" bind:value={form.address_ktp}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Alamat sesuai KTP" />
								</label>
							</div>
							<div>
								<label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">
									Alamat Domisili
									<input type="text" bind:value={form.address}
										class="mt-1.5 w-full px-3 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400"
										placeholder="Alamat saat ini" />
								</label>
							</div>
						</div>
					</div>
				</section>
			</div>

			<!-- Form Footer -->
			<div class="sticky bottom-0 z-10 flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-b-xl">
				<button onclick={cancelForm}
					class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:bg-gray-800 dark:hover:bg-gray-800 transition cursor-pointer">
					Batal
				</button>
				<button onclick={handleSave} disabled={isSaving}
					class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
						</svg>
					{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Karyawan'}
				</button>
			</div>
		</div>
	<!-- Table Card -->
	<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm" class:hidden={showForm}>
			{#if isLoading}
				<PulseLoader variant="table-row" count={5} />
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center">
						<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
						</svg>
					</div>
					<p class="text-sm font-medium text-gray-900 dark:text-white mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">{errorMessage}</p>
					<button onclick={loadEmployees} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if employees.length === 0}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
						<svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
						</svg>
					</div>
					<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Belum ada data karyawan</h3>
					<p class="text-sm text-gray-500 dark:text-gray-400 max-w-xs mx-auto">
						{searchQuery
							? `Tidak ditemukan karyawan dengan kata kunci "${searchQuery}"`
							: 'Data karyawan akan muncul di sini setelah ditambahkan.'}
					</p>
				</div>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>

				<!-- Mobile Cards -->
				<PullToRefresh onRefresh={loadEmployees}>
					<div class="md:hidden space-y-3">
						{#each employees as emp}
							<MobileCard
								title={emp.full_name}
								subtitle={emp.position_name || emp.email || '-'}
								avatar={getInitials(emp.full_name)}
								avatarColor={getAvatarTheme('employee').gradientClasses}
								badges={[{ label: emp.employment_status ? emp.employment_status.charAt(0).toUpperCase() + emp.employment_status.slice(1) : '-', color: employmentStatusColors[emp.employment_status?.toLowerCase()] || 'bg-gray-50 text-gray-600 dark:bg-gray-800 dark:text-gray-400' }]}
							>
								{#snippet children()}
									<div class="flex items-center gap-1.5 text-[11px] text-gray-500 dark:text-gray-400">
										<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
										<span>{emp.department_name || 'Tanpa departemen'}</span>
									</div>
								{/snippet}
								{#snippet footer()}
									<div class="flex items-center gap-2">
										<button onclick={() => viewDetail(emp.employee_id)} class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95">Detail</button>
										{#if isDeleted(emp)}
											{#if hasPermission('employee', 'update')}
												<button onclick={() => handleRestore(emp.id)} disabled={restoringId === emp.id} class="flex-1 py-2 text-xs font-medium text-emerald-600 dark:text-emerald-300 bg-emerald-50 dark:bg-emerald-900/30 rounded-lg hover:bg-emerald-100 dark:hover:bg-emerald-900/50 transition cursor-pointer active:scale-95 disabled:opacity-50">{restoringId === emp.id ? '...' : 'Aktifkan'}</button>
											{/if}
										{:else}
											{#if hasPermission('employee', 'update')}
												<button onclick={() => openEditForm(emp)} class="flex-1 py-2 text-xs font-medium text-[#1A56DB] dark:text-blue-300 bg-blue-50 dark:bg-blue-900/30 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/50 transition cursor-pointer active:scale-95">Edit</button>
											{/if}
											{#if hasPermission('employee', 'delete')}
												<button onclick={() => confirmDelete(emp.id, emp.full_name)} class="flex-1 py-2 text-xs font-medium text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95">Nonaktifkan</button>
											{/if}
										{/if}
									</div>
								{/snippet}
							</MobileCard>
						{/each}
					</div>
				</PullToRefresh>

				<!-- Pagination -->
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50 dark:bg-gray-800/30">
					<div class="text-xs text-gray-500 dark:text-gray-400">
						Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700 dark:text-gray-300">{total}</span>
					</div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:bg-gray-800 hover:text-gray-900 dark:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)}
									class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:bg-gray-800'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages}
							class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 dark:border-gray-800 text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:bg-gray-800 hover:text-gray-900 dark:text-white disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
	</div>
</div>

<!-- Delete Confirmation Modal -->
<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<div onclick={cancelDelete} onkeydown={(e) => e.key === 'Enter' && cancelDelete()} role="button" tabindex="0" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} onkeydown={(e) => e.key === 'Enter' && e.stopPropagation()} role="dialog" tabindex="-1" class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
					</svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Nonaktifkan Karyawan</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Apakah Anda yakin ingin menonaktifkan karyawan</p>
				<p class="text-sm font-medium text-gray-900 dark:text-white mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 mb-6">Karyawan yang dinonaktifkan tidak akan muncul di daftar aktif.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 dark:border-gray-800 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={isSaving}
						class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
							</svg>
						{/if}
						Ya, Nonaktifkan
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
