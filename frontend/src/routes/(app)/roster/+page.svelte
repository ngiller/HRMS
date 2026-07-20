<script lang="ts">
	import { onMount } from 'svelte';
	import { fade, fly } from 'svelte/transition';
	import { rosters as rosterApi, shifts as shiftApi, departments as deptApi, employees as empApi, rosterEntries as entryApi } from '$lib/api.js';

	type RosterSummary = {
		id: string; department_id: string; department_name: string;
		name: string; month: number; year: number;
		is_published: boolean; entry_count: number; created_at: string;
	};

	type ShiftItem = { id: string; name: string; code: string; start_time: string; end_time: string; color: string; department_id?: string; };
	type DeptItem = { id: string; name: string; };
	type EmpItem = { id: string; full_name: string; employee_id: string; };

	type DayInfo = {
		shift_id?: string; shift_name?: string; shift_code?: string;
		shift_color?: string; start_time?: string; end_time?: string;
		entry_id?: string; notes?: string;
	};

	type CalendarEmp = {
		employee_id: string; employee_name: string; days: Record<string, DayInfo>;
	};

	let rosters = $state<RosterSummary[]>([]);
	let shifts = $state<ShiftItem[]>([]);
	let departments = $state<DeptItem[]>([]);
	let employees = $state<EmpItem[]>([]);

	let loading = $state(true);
	let error = $state('');

	// View state
	let selectedRosterId = $state<string>('');
	let calendarData = $state<CalendarEmp[]>([]);
	let calendarLoading = $state(false);
	let showAddEmployeeMenu = $state(false);
	let availableEmployeesToAdd = $state<EmpItem[]>([]);
	let showDeleteConfirm = $state(false);
	let deleteEmployeeId = $state('');
	let deleteEmployeeName = $state('');

	// Filter
	let filterDept = $state('');
	let filterMonth = $state(new Date().getMonth() + 1);
	let filterYear = $state(new Date().getFullYear());

	// Current month for calendar display
	let currentMonth = $state(new Date().getMonth() + 1);
	let currentYear = $state(new Date().getFullYear());

	// Selected roster detail
	let selectedRoster = $state<RosterSummary | null>(null);

	// Assign mode
	let assignMode = $state(false);
	let assignEmpId = $state('');
	let assignDate = $state('');
	let assignShiftId = $state('');
	let assignError = $state('');
	let assignNotes = $state('');
	let assignEntryId = $state('');

	// Auto schedule settings
	let showAutoSettings = $state(false);
	let autoBlockSize = $state(3);
	let autoOffSize = $state(3);

	// Notifications / feedback
	let successMsg = $state('');

	const MONTHS = ['Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'];
	const DAYS = ['Min','Sen','Sel','Rab','Kam','Jum','Sab'];

	onMount(async () => {
		await loadAll();
	});

	async function loadAll() {
		loading = true; error = '';
		try {
			const [shiftRes, deptRes, empRes] = await Promise.all([
				shiftApi.getAll() as any,
				deptApi.getAll() as any,
				empApi.list(1, 500, '', '', 'active') as any,
			]);
			shifts = shiftRes.data || [];
			departments = deptRes.data || [];
			employees = empRes.data || [];
			await loadRosters();
		} catch (e: any) { error = e.message || 'Gagal memuat data'; }
		finally { loading = false; }
	}

	async function loadRosters() {
		try {
			const res: any = await rosterApi.list(1, 50, filterDept);
			rosters = res.data || [];
			// Auto-select latest roster for current month
			const match = rosters.find(r => r.month === filterMonth && r.year === filterYear);
			if (match) selectRoster(match);
		} catch {}
	}

	function selectRoster(r: RosterSummary) {
		selectedRosterId = r.id;
		selectedRoster = r;
		currentMonth = r.month;
		currentYear = r.year;
		loadCalendar();
	}

	async function loadCalendar() {
		if (!selectedRosterId) return;
		calendarLoading = true;
		try {
			const res: any = await rosterApi.getCalendar(selectedRosterId);
			calendarData = res.data || [];
			if (selectedRoster?.department_id) {
				const shiftRes: any = await shiftApi.getAll(selectedRoster.department_id);
				shifts = shiftRes.data || [];
			}
		} catch (e: any) {
			console.error('Gagal memuat kalender:', e);
			calendarData = [];
		}
		finally { calendarLoading = false; }
	}

	async function createRoster() {
		if (!filterDept) { error = 'Pilih departemen dulu'; return; }
		const name = `${departments.find(d => d.id === filterDept)?.name || ''} ${MONTHS[filterMonth-1]} ${filterYear}`;
		try {
			await rosterApi.create({ department_id: filterDept, name, month: filterMonth, year: filterYear });
			await loadRosters();
		} catch (e: any) { error = e.message || 'Gagal membuat roster'; }
	}

	async function togglePublish() {
		if (!selectedRoster) return;
		try {
			await rosterApi.update(selectedRoster.id, { is_published: !selectedRoster.is_published });
			selectedRoster = { ...selectedRoster, is_published: !selectedRoster.is_published };
			rosters = rosters.map(r => r.id === selectedRoster!.id ? { ...r, is_published: !r.is_published } : r);
		} catch {}
	}

	async function deleteRoster() {
		if (!selectedRoster) return;
		try {
			await rosterApi.remove(selectedRoster.id);
			selectedRosterId = '';
			selectedRoster = null;
			calendarData = [];
			await loadRosters();
		} catch (e: any) { error = e.message || 'Gagal menghapus'; }
	}

	function removeEmployeeFromRoster(empId: string, empName: string) {
		deleteEmployeeId = empId;
		deleteEmployeeName = empName;
		showDeleteConfirm = true;
	}

	async function confirmRemoveEmployee() {
		if (!selectedRosterId || !deleteEmployeeId) return;
		showDeleteConfirm = false;
		error = '';
		successMsg = '';
		try {
			await rosterApi.removeEmployee(selectedRosterId, deleteEmployeeId);
			successMsg = `✓ ${deleteEmployeeName} berhasil dihapus dari roster`;
			deleteEmployeeId = '';
			deleteEmployeeName = '';
			await loadCalendar();
		} catch (e: any) {
			error = e.message || 'Gagal menghapus karyawan dari roster';
		}
	}

	function cancelRemoveEmployee() {
		showDeleteConfirm = false;
		deleteEmployeeId = '';
		deleteEmployeeName = '';
	}

	// Get days in month
	function getDaysInMonth() {
		const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();
		const firstDay = new Date(currentYear, currentMonth - 1, 1).getDay();
		const result: (number | null)[] = [];
		for (let i = 0; i < firstDay; i++) result.push(null);
		for (let i = 1; i <= daysInMonth; i++) result.push(i);
		return result;
	}

	function formatDate(day: number): string {
		return `${currentYear}-${String(currentMonth).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
	}

	function getShiftForDay(emp: CalendarEmp, day: number): DayInfo | undefined {
		const dateKey = formatDate(day);
		return emp.days[dateKey];
	}

	function getShiftBadge(dayInfo: DayInfo | undefined): string {
		if (!dayInfo?.shift_name) return '';
		return dayInfo.shift_name;
	}

	function getShiftColor(dayInfo: DayInfo | undefined): string {
		return dayInfo?.shift_color || '#E5E7EB';
	}

	// Quick assign
	function openAssign(empId: string, date: string) {
		assignEmpId = empId;
		assignDate = date;
		assignShiftId = '';
		assignError = '';
		assignNotes = '';
		assignEntryId = '';
		assignMode = true;

		// Pre-populate if there is an existing shift
		const currentEmpData = calendarData.find(c => c.employee_id === empId);
		if (currentEmpData) {
			const dayInfo = currentEmpData.days[date];
			if (dayInfo) {
				assignShiftId = dayInfo.shift_id || '';
				assignNotes = dayInfo.notes || '';
				assignEntryId = dayInfo.entry_id || '';
			}
		}
	}
	function closeAssign() {
		assignMode = false;
		assignEmpId = '';
		assignDate = '';
		assignShiftId = '';
		assignNotes = '';
		assignEntryId = '';
	}

	async function saveAssign() {
		if (!assignShiftId || !selectedRosterId) { assignError = 'Pilih shift'; return; }
		try {
			await rosterApi.bulkCreateEntries(selectedRosterId, {
				entries: [{ roster_id: selectedRosterId, employee_id: assignEmpId, date: assignDate, shift_id: assignShiftId, notes: assignNotes }]
			});
			closeAssign();
			await loadCalendar();
		} catch (e: any) { assignError = e.message || 'Gagal menyimpan'; }
	}

	async function deleteAssign() {
		if (!assignEntryId) return;
		try {
			await entryApi.remove(assignEntryId);
			closeAssign();
			await loadCalendar();
		} catch (e: any) { assignError = e.message || 'Gagal menghapus entri'; }
	}

	function formatDisplayDate(dateStr: string): string {
		if (!dateStr) return '';
		const date = new Date(dateStr);
		if (isNaN(date.getTime())) return dateStr;
		const options: Intl.DateTimeFormatOptions = { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' };
		return date.toLocaleDateString('id-ID', options);
	}

	function getEmployeeForRoster(emp: CalendarEmp): EmpItem | undefined {
		return employees.find(e => e.id === emp.employee_id);
	}

	// Shared helper: fetch active employees from the roster's department
	async function fetchDeptEmployees(): Promise<EmpItem[]> {
		if (!selectedRoster?.department_id) return [];
		try {
			const res: any = await empApi.list(1, 999, '', selectedRoster.department_id, 'active');
			return res.data || [];
		} catch {
			return [];
		}
	}

	async function toggleAddEmployeeMenu() {
		if (showAddEmployeeMenu) {
			showAddEmployeeMenu = false;
			return;
		}

		error = '';
		successMsg = '';
		if (!selectedRoster || !selectedRosterId) { error = 'Pilih roster terlebih dahulu'; return; }

		try {
			const deptEmps = await fetchDeptEmployees();
			const existingIds = new Set(calendarData.map(c => c.employee_id));
			availableEmployeesToAdd = deptEmps.filter(e => !existingIds.has(e.id));
			showAddEmployeeMenu = true;
		} catch (e: any) {
			error = 'Gagal memuat daftar karyawan';
		}
	}

	async function addSingleEmployee(empId: string) {
		showAddEmployeeMenu = false;
		error = '';
		successMsg = '';
		
		const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();
		const firstShift = shifts[0]?.id || '';
		if (!firstShift) { error = 'Buat shift kerja terlebih dahulu di menu Shift Kerja'; return; }

		calendarLoading = true;
		try {
			const entries: Array<{ roster_id: string; employee_id: string; date: string; shift_id: string; notes: string }> = [];
			for (let day = 1; day <= daysInMonth; day++) {
				entries.push({
					roster_id: selectedRosterId,
					employee_id: empId,
					date: formatDate(day),
					shift_id: firstShift,
					notes: ''
				});
			}

			await rosterApi.bulkCreateEntries(selectedRosterId, { entries });
			
			const empName = availableEmployeesToAdd.find(e => e.id === empId)?.full_name || '';
			successMsg = `✓ Karyawan ${empName} berhasil ditambahkan ke roster`;
			
			await loadCalendar();
		} catch (e: any) {
			console.error('Gagal menambahkan karyawan:', e);
			error = e.message || 'Gagal menambahkan karyawan ke roster.';
		} finally {
			calendarLoading = false;
		}
	}

	async function addAllDeptEmployeesToRoster() {
		error = '';
		successMsg = '';
		if (!selectedRoster || !selectedRosterId) { error = 'Pilih roster terlebih dahulu'; return; }
		if (!selectedRoster.department_id) { error = 'Roster tidak memiliki departemen'; return; }

		const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();
		const firstShift = shifts[0]?.id || '';
		if (!firstShift) { error = 'Buat shift kerja terlebih dahulu di menu Shift Kerja'; return; }

		calendarLoading = true;

		try {
			// Fetch employees filtered by roster's department
			const deptEmps = await fetchDeptEmployees();

			if (deptEmps.length === 0) {
				error = 'Tidak ada karyawan aktif di departemen ini';
				return;
			}

			// Remove already-added employees
			const existingIds = new Set(calendarData.map(c => c.employee_id));
			const newEmps = deptEmps.filter(e => !existingIds.has(e.id));

			const addedNames = newEmps.map(e => e.full_name);

			if (newEmps.length === 0) {
				error = 'Semua karyawan di departemen ini sudah ditambahkan ke roster';
				return;
			}

			// Build entries for each employee × each day
			const entries: Array<{ roster_id: string; employee_id: string; date: string; shift_id: string; notes: string }> = [];
			for (const emp of newEmps) {
				for (let day = 1; day <= daysInMonth; day++) {
					entries.push({ roster_id: selectedRosterId, employee_id: emp.id, date: formatDate(day), shift_id: firstShift, notes: '' });
				}
			}

			if (entries.length === 0) { error = 'Tidak ada entri yang perlu ditambahkan'; return; }

			const count = newEmps.length;
			const totalEntries = entries.length;
			successMsg = `✓ ${count} karyawan ditambahkan (${totalEntries} entri jadwal)`;
			if (count <= 3) {
				successMsg += ': ' + addedNames.join(', ');
			}

			await rosterApi.bulkCreateEntries(selectedRosterId, { entries });
			await loadCalendar();
		} catch (e: any) {
			console.error('Gagal menambahkan karyawan:', e);
			error = e.message || 'Gagal menambahkan karyawan ke roster. Cek koneksi dan izin akses.';
			successMsg = '';
		} finally {
			calendarLoading = false;
		}
	}

	async function handleAutoSchedule() {
		error = '';
		successMsg = '';
		showAutoSettings = false;
		if (!selectedRoster || !selectedRosterId) return;

		const targetEmps = calendarData.length > 0
			? calendarData.map(c => ({ id: c.employee_id, name: c.employee_name }))
			: (await fetchDeptEmployees()).map(e => ({ id: e.id, name: e.full_name }));

		if (shifts.length === 0) {
			error = 'Belum ada shift kerja yang terdaftar. Buat shift di menu Shift Kerja terlebih dahulu.';
			return;
		}

		// Sort shifts by start_time ASC
		const sortedShifts = [...shifts].sort((a, b) => a.start_time.localeCompare(b.start_time));

		// Number of unique shifts
		const N = sortedShifts.length;
		
		// Define the rotation cycle dynamically
		let cycle: Array<{ shiftId: string | null; duration: number }> = [];
		
		if (N === 1) {
			// For a single shift (e.g. office hours), use a standard work/off cycle
			const offDays = autoOffSize > 0 ? autoOffSize : 2;
			const workDays = autoBlockSize > 0 ? autoBlockSize : 5;
			cycle = [
				{ shiftId: sortedShifts[0].id, duration: workDays },
				{ shiftId: null, duration: offDays }
			];
		} else {
			// For N >= 2 shifts, use a flexible rotation
			const blockSize = autoBlockSize > 0 ? autoBlockSize : 3;
			const offDays = autoOffSize > 0 ? autoOffSize : 3;
			for (let i = 0; i < N; i++) {
				cycle.push({ shiftId: sortedShifts[i].id, duration: blockSize });
			}
			cycle.push({ shiftId: null, duration: offDays });
		}

		const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();
		const entries: Array<{ roster_id: string; employee_id: string; date: string; shift_id: string; notes: string }> = [];

		calendarLoading = true;
		error = '';
		try {
			const cycleDays = cycle.reduce((sum, item) => sum + item.duration, 0);

			targetEmps.forEach((emp, empIdx) => {
				// Calculate cohort offset to distribute shifts evenly across employees
				let offset = 0;
				if (N === 1) {
					// Offset by 1 day per employee to distribute off days throughout the week
					offset = (empIdx % cycleDays) * 1;
				} else {
					// Offset by block size to distribute shifts
					const numCohorts = N + 1;
					const cohort = empIdx % numCohorts;
					const blockSize = autoBlockSize > 0 ? autoBlockSize : 3;
					offset = cohort * blockSize;
				}

				for (let day = 1; day <= daysInMonth; day++) {
					const cyclePos = (day - 1 + offset) % cycleDays;
					
					// Determine shiftId by traversing the cycle
					let accum = 0;
					let shiftId = null;
					for (const step of cycle) {
						accum += step.duration;
						if (cyclePos < accum) {
							shiftId = step.shiftId;
							break;
						}
					}

					if (shiftId) {
						const date = formatDate(day);
						entries.push({
							roster_id: selectedRosterId,
							employee_id: emp.id,
							date,
							shift_id: shiftId,
							notes: 'Jadwal Otomatis (Rotasi Fleksibel)'
						});
					}
				}
			});

			if (entries.length > 0) {
				await rosterApi.bulkCreateEntries(selectedRosterId, { entries, clear_existing: true });
			}
			await loadCalendar();
			successMsg = `✓ Jadwal otomatis berhasil dibuat untuk ${targetEmps.length} karyawan`;
		} catch (e: any) {
			console.error('Gagal membuat jadwal otomatis:', e);
			error = e.message || 'Gagal membuat jadwal otomatis. Cek koneksi dan izin akses.';
		} finally {
			calendarLoading = false;
		}
	}
</script>

<svelte:window onclick={(e) => {
	const target = e.target as HTMLElement;
	if (showAddEmployeeMenu && target && !target.closest('.add-emp-container')) {
		showAddEmployeeMenu = false;
	}
	if (showAutoSettings && target && !target.closest('.auto-schedule-container')) {
		showAutoSettings = false;
	}
}} />

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Roster Bulanan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Atur jadwal shift karyawan per bulan per departemen</p>
		</div>
	</div>

	<!-- Controls -->
	<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl p-4 mb-5 shadow-sm">
		<div class="flex flex-col sm:flex-row items-start sm:items-center gap-3 flex-wrap">
			<div class="flex items-center gap-2">
				<label class="text-xs font-semibold text-gray-500 dark:text-gray-400 whitespace-nowrap">Departemen:</label>
				<select bind:value={filterDept} onchange={() => loadRosters()} class="px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
					<option value="">Semua Dept</option>
					{#each departments as d (d)}
						<option value={d.id}>{d.name}</option>
					{/each}
				</select>
			</div>
			<div class="flex items-center gap-2">
				<label class="text-xs font-semibold text-gray-500 dark:text-gray-400 whitespace-nowrap">Bulan:</label>
				<select bind:value={filterMonth} class="px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
					{#each MONTHS as m, i}
						<option value={i+1}>{m}</option>
					{/each}
				</select>
				<select bind:value={filterYear} class="px-3 py-2 border border-gray-200 dark:border-gray-700 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] bg-white dark:bg-gray-900 text-gray-900 dark:text-white">
					{#each [currentYear-1, currentYear, currentYear+1] as y}
						<option value={y}>{y}</option>
					{/each}
				</select>
			</div>
			<button onclick={createRoster} disabled={!filterDept} class="px-4 py-2 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-40 disabled:cursor-not-allowed cursor-pointer active:scale-[0.97] inline-flex items-center gap-2">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Buat Roster {filterDept ? MONTHS[filterMonth-1] : ''} {filterYear}
			</button>
		</div>

		<!-- Roster tabs -->
		{#if rosters.length > 0}
			<div class="flex flex-wrap gap-2 mt-4 pt-4 border-t border-gray-100 dark:border-gray-700">
				{#each rosters as r (r)}
					<button onclick={() => selectRoster(r)} class="px-3 py-1.5 rounded-xl text-xs font-semibold transition cursor-pointer {selectedRosterId === r.id ? 'bg-[#1A56DB] text-white shadow-sm' : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'}">
						{r.name}
						{#if r.is_published}
							<span class="ml-1.5 text-[10px] opacity-70">✓</span>
						{/if}
					</button>
				{/each}
			</div>
		{/if}
	</div>

	{#if selectedRoster}
		<!-- Roster Toolbar -->
		<div class="flex items-center justify-between mb-4">
			<div class="flex items-center gap-2">
				<h2 class="text-lg font-bold text-gray-900 dark:text-white">{selectedRoster.name}</h2>
				<span class="px-2 py-0.5 rounded text-[10px] font-medium {selectedRoster.is_published ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300'}">
					{selectedRoster.is_published ? 'Published' : 'Draft'}
				</span>
			</div>
			<div class="flex items-center gap-2">
				<div class="relative inline-block text-left add-emp-container">
					<button onclick={toggleAddEmployeeMenu} class="px-3 py-1.5 border border-gray-200 dark:border-gray-700 rounded-xl text-xs font-semibold text-gray-650 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer inline-flex items-center gap-1">
						<span>➕ Tambah Karyawan</span>
						<svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
					</button>
					{#if showAddEmployeeMenu}
						<div class="absolute right-0 mt-2 w-64 origin-top-right rounded-xl bg-white dark:bg-gray-800 shadow-xl ring-1 ring-black ring-opacity-5 focus:outline-none z-50 border border-gray-200 dark:border-gray-700 p-2.5">
							<div class="px-2 py-1 mb-1.5 border-b border-gray-100 dark:border-gray-700">
								<span class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Pilih Karyawan</span>
							</div>
							<div class="max-h-60 overflow-y-auto space-y-0.5">
								{#if availableEmployeesToAdd.length === 0}
									<div class="text-[11px] text-gray-500 dark:text-gray-400 py-3 text-center">Semua karyawan sudah ditambahkan</div>
								{:else}
									{#each availableEmployeesToAdd as emp (emp.id)}
										<button onclick={() => addSingleEmployee(emp.id)} class="w-full text-left px-2.5 py-1.5 text-xs text-gray-750 dark:text-gray-350 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-[#1A56DB] dark:hover:text-white rounded-lg transition-colors truncate cursor-pointer font-medium">
											{emp.full_name}
										</button>
									{/each}
								{/if}
							</div>
						</div>
					{/if}
				</div>
				<button onclick={addAllDeptEmployeesToRoster} class="px-3 py-1.5 border border-gray-200 dark:border-gray-700 rounded-xl text-xs font-semibold text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">
					Tambah Semua Karyawan
				</button>
				<div class="relative inline-block text-left auto-schedule-container">
					<div class="flex items-center">
						<button onclick={handleAutoSchedule} class="px-3 py-1.5 bg-indigo-50 border border-r-0 border-indigo-200 dark:border-indigo-850 dark:bg-indigo-900/30 rounded-l-xl text-xs font-bold text-indigo-650 dark:text-indigo-400 hover:bg-indigo-100 transition cursor-pointer flex items-center gap-1">
							✨ Jadwal Otomatis
						</button>
						<button onclick={() => showAutoSettings = !showAutoSettings} class="px-2 py-1.5 bg-indigo-50 border border-indigo-200 dark:border-indigo-850 dark:bg-indigo-900/30 rounded-r-xl text-xs text-indigo-650 dark:text-indigo-400 hover:bg-indigo-100 transition cursor-pointer border-l-0" title="Pengaturan Jadwal Otomatis">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 6h9.75M10.5 6a1.5 1.5 0 1 1-3 0m3 0a1.5 1.5 0 1 0-3 0M3.75 6H7.5m3 12h9.75m-9.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-3.75 0H7.5m9-6h3.75m-3.75 0a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m-9.75 0h9.75" /></svg>
						</button>
					</div>
					{#if showAutoSettings}
						<div class="absolute right-0 mt-2 w-64 origin-top-right rounded-xl bg-white dark:bg-gray-800 shadow-xl ring-1 ring-black ring-opacity-5 focus:outline-none z-50 border border-gray-200 dark:border-gray-700 p-4 space-y-3">
							<div class="border-b border-gray-100 dark:border-gray-700 pb-2">
								<span class="text-xs font-bold text-gray-900 dark:text-white uppercase tracking-wider">Pengaturan Siklus</span>
							</div>
							<div class="space-y-3">
								<div>
									<label class="block text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1">Durasi Kerja (Hari)</label>
									<input type="number" min="1" max="10" bind:value={autoBlockSize} class="w-full px-2.5 py-1.5 text-xs bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-indigo-500" />
									<span class="text-[9px] text-gray-400 mt-1 block">Hari kerja berurutan tiap shift.</span>
								</div>
								<div>
									<label class="block text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-1">Durasi Libur (Hari)</label>
									<input type="number" min="1" max="10" bind:value={autoOffSize} class="w-full px-2.5 py-1.5 text-xs bg-gray-50 dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg text-gray-900 dark:text-white focus:outline-none focus:ring-1 focus:ring-indigo-500" />
									<span class="text-[9px] text-gray-400 mt-1 block">Hari libur di akhir siklus.</span>
								</div>
							</div>
							<div class="flex justify-end gap-2 pt-1 border-t border-gray-100 dark:border-gray-700">
								<button onclick={() => { autoBlockSize = 3; autoOffSize = 3; showAutoSettings = false; }} class="px-2 py-1 text-[10px] border border-gray-200 dark:border-gray-700 rounded-lg text-gray-600 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Reset</button>
								<button onclick={() => showAutoSettings = false} class="px-2 py-1 text-[10px] bg-[#1A56DB] text-white rounded-lg hover:bg-[#1e40af] transition cursor-pointer font-bold">Tutup</button>
							</div>
						</div>
					{/if}
				</div>
				<button onclick={togglePublish} class="px-3 py-1.5 border border-gray-200 dark:border-gray-700 rounded-xl text-xs font-semibold {selectedRoster.is_published ? 'text-amber-600 border-amber-200 dark:border-amber-800' : 'text-emerald-600 border-emerald-200 dark:border-emerald-800'} hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">
					{selectedRoster.is_published ? 'Unpublish' : 'Publish'}
				</button>
				<button onclick={deleteRoster} class="px-3 py-1.5 border border-red-200 dark:border-red-800 rounded-xl text-xs font-semibold text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20 transition cursor-pointer">
					Hapus
				</button>
			</div>
		</div>

		{#if error || successMsg}
			<div class="text-xs px-4 py-2.5 rounded-xl mb-4 font-semibold flex items-center justify-between {error ? 'bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-200' : 'bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 text-emerald-700 dark:text-emerald-200'}">
				<span>{error ? '⚠️ ' + error : successMsg}</span>
				<button onclick={() => { error = ''; successMsg = ''; }} class="{error ? 'text-red-500 hover:text-red-700' : 'text-emerald-500 hover:text-emerald-700'} font-bold ml-2 cursor-pointer">×</button>
			</div>
		{/if}

		<!-- Calendar Grid -->
		{#if calendarLoading}
			<div class="flex items-center justify-center py-20"><div class="animate-spin h-8 w-8 border-2 border-[#1A56DB] border-t-transparent rounded-full"></div></div>
		{:else}
			<div class="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-2xl overflow-hidden shadow-sm">
				<div class="overflow-x-auto">
					<table class="w-full text-sm border-collapse min-w-[800px]">
						<thead>
							<tr class="bg-gray-50 dark:bg-gray-900/50">
								<th class="sticky left-0 bg-gray-50 dark:bg-gray-900/50 z-10 px-3 py-2.5 text-left text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider border-r border-b border-gray-100 dark:border-gray-700 min-w-[160px]">
									Karyawan
								</th>
								{#each getDaysInMonth() as day, i}
									{#if day !== null}
										<th class="px-1 py-2 text-center text-[10px] font-semibold text-gray-500 dark:text-gray-400 border-b border-gray-100 dark:border-gray-700 min-w-[32px]">
											<div class="text-[9px] uppercase">{DAYS[new Date(currentYear, currentMonth - 1, day).getDay()]}</div>
											<div class="text-xs font-bold text-gray-800 dark:text-gray-200">{day}</div>
										</th>
									{/if}
								{/each}
							</tr>
						</thead>
						<tbody>
							{#each calendarData as emp (emp)}
								<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition-colors">
									<td class="sticky left-0 bg-white dark:bg-gray-800 z-10 px-3 py-2 text-sm font-medium text-gray-900 dark:text-white border-r border-b border-gray-100 dark:border-gray-700 whitespace-nowrap group/row">
										<div class="flex items-center justify-between gap-2">
											<div class="flex items-center gap-2 min-w-0">
												<div class="w-6 h-6 rounded-full bg-gradient-to-br from-[#1A56DB] to-blue-300 flex items-center justify-center text-[9px] font-bold text-white shrink-0">
													{emp.employee_name.charAt(0)}
												</div>
												<span class="truncate max-w-[100px]" title={emp.employee_name}>{emp.employee_name}</span>
											</div>
											<button onclick={() => removeEmployeeFromRoster(emp.employee_id, emp.employee_name)} class="opacity-0 group-hover/row:opacity-100 transition-opacity p-1 text-red-500 hover:text-red-700 hover:bg-red-50 dark:hover:bg-red-900/30 rounded-lg cursor-pointer" title="Hapus dari roster">
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.34 9m-4.78 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
											</button>
										</div>
									</td>
									{#each getDaysInMonth() as day, i}
										{#if day !== null}
											{@const dayInfo = getShiftForDay(emp, day)}
											{@const dateKey = formatDate(day)}
											<td class="px-0.5 py-1 text-center border-b border-gray-100 dark:border-gray-700 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors relative group" onclick={() => openAssign(emp.employee_id, dateKey)}>
												{#if dayInfo?.shift_name}
													<div class="mx-auto w-full rounded-md px-0.5 py-1 text-[9px] font-semibold text-white truncate shadow-sm" style="background: {dayInfo.shift_color}">
														{dayInfo.shift_name}
													</div>
												{:else}
													<div class="text-[9px] text-gray-300 dark:text-gray-600 opacity-0 group-hover:opacity-100 transition-opacity">+</div>
												{/if}
											</td>
										{/if}
									{/each}
								</tr>
							{/each}
						</tbody>
					</table>
				</div>

				{#if calendarData.length === 0}
					<div class="py-16 text-center">
						<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
							<svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25Z" /></svg>
						</div>
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Roster masih kosong</h3>
						<p class="text-sm text-gray-500 dark:text-gray-400 mb-4">Klik "Tambah Semua Karyawan" atau klik sel kosong untuk assign shift.</p>
					</div>
				{/if}

				<!-- Legend -->
				<div class="px-4 py-3 border-t border-gray-100 dark:border-gray-700 flex flex-wrap items-center gap-4">
					<span class="text-xs text-gray-500 dark:text-gray-400 font-semibold">Legenda:</span>
					{#each shifts as s (s)}
						<div class="flex items-center gap-1.5">
							<div class="w-3 h-3 rounded" style="background: {s.color}"></div>
							<span class="text-xs text-gray-600 dark:text-gray-400">{s.name} ({s.start_time.slice(0,5)}-{s.end_time.slice(0,5)})</span>
						</div>
					{/each}
				</div>
			</div>
		{/if}
	{:else if !selectedRoster && !loading}
		<div class="py-16 text-center">
			<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center">
				<svg class="w-7 h-7 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
			</div>
			<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">Pilih atau buat roster</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400">Pilih departemen dan bulan, lalu klik "Buat Roster" untuk memulai.</p>
		</div>
	{/if}
</div>

<!-- Assign Slide-over Overlay -->
{#if assignMode}
	<div class="fixed inset-0 z-50 bg-black/40 backdrop-blur-xs transition-opacity" transition:fade={{ duration: 150 }} onclick={closeAssign} role="presentation"></div>
{/if}

<!-- Assign Slide-over Panel -->
{#if assignMode}
	<div onclick={(e) => e.stopPropagation()} role="dialog" 
		 class="fixed right-0 top-0 bottom-0 z-50 bg-white dark:bg-gray-800 shadow-2xl w-full max-w-md h-full flex flex-col border-l border-gray-100 dark:border-gray-700"
		 transition:fly={{ x: 400, duration: 250 }}>
		
		<!-- Header -->
		<div class="p-6 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
			<div>
				<h3 class="text-lg font-bold text-gray-900 dark:text-white">Atur Jadwal Shift</h3>
				<p class="text-xs text-gray-500 dark:text-gray-400 mt-1">Assign atau ubah shift kerja harian karyawan.</p>
			</div>
			<button onclick={closeAssign} class="p-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 hover:bg-gray-50 dark:hover:bg-gray-800 rounded-xl transition cursor-pointer">
				<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
			</button>
		</div>

		<!-- Body -->
		<div class="flex-1 overflow-y-auto p-6 space-y-6">
			<!-- Employee Section -->
			<div class="flex items-center gap-3 bg-gray-50 dark:bg-gray-900/50 p-4 rounded-2xl">
				<div class="w-10 h-10 rounded-full bg-gradient-to-br from-indigo-500 to-blue-400 flex items-center justify-center text-sm font-bold text-white shadow-sm shrink-0">
					{(employees.find(e => e.id === assignEmpId)?.full_name || 'K').charAt(0)}
				</div>
				<div>
					<label class="block text-[10px] font-bold text-gray-400 uppercase tracking-wider">Karyawan</label>
					<p class="text-sm font-bold text-gray-900 dark:text-white mt-0.5">{employees.find(e => e.id === assignEmpId)?.full_name || assignEmpId}</p>
				</div>
			</div>

			<!-- Date Section -->
			<div class="flex items-center gap-3 bg-gray-50 dark:bg-gray-900/50 p-4 rounded-2xl">
				<div class="w-10 h-10 rounded-full bg-emerald-50 dark:bg-emerald-950/30 flex items-center justify-center text-emerald-600 shrink-0">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
				</div>
				<div>
					<label class="block text-[10px] font-bold text-gray-400 uppercase tracking-wider">Tanggal Kerja</label>
					<p class="text-sm font-bold text-gray-900 dark:text-white mt-0.5">{formatDisplayDate(assignDate)}</p>
				</div>
			</div>

			<!-- Shift List Section -->
			<div class="space-y-2.5">
				<label class="block text-xs font-bold text-gray-500 dark:text-gray-400">Pilih Shift Kerja</label>
				<div class="grid grid-cols-1 gap-2">
					{#each shifts as s (s)}
						<button onclick={() => assignShiftId = s.id} 
								class="w-full flex items-center justify-between p-3.5 rounded-2xl border text-left transition cursor-pointer group {assignShiftId === s.id ? 'ring-2 ring-offset-1 dark:ring-offset-gray-800' : 'hover:bg-gray-50 dark:hover:bg-gray-800 border-gray-100 dark:border-gray-700'}" 
								style="border-color: {s.color}30; {assignShiftId === s.id ? `background: ${s.color}08; border-color: ${s.color}; ring-color: ${s.color};` : ''}">
							<div class="flex items-center gap-3">
								<div class="w-4 h-4 rounded-full flex items-center justify-center shrink-0" style="border: 2px solid {s.color}; background: {assignShiftId === s.id ? s.color : 'transparent'}">
									{#if assignShiftId === s.id}
										<div class="w-1.5 h-1.5 rounded-full bg-white"></div>
									{/if}
								</div>
								<div>
									<span class="text-xs font-bold text-gray-900 dark:text-white group-hover:text-indigo-600 dark:group-hover:text-indigo-400 transition-colors">{s.name}</span>
									<span class="block text-[10px] text-gray-400 mt-0.5">Shift Code: {s.code}</span>
								</div>
							</div>
							<span class="text-xs font-bold text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-800/80 px-2.5 py-1 rounded-lg">{s.start_time.slice(0,5)} - {s.end_time.slice(0,5)}</span>
						</button>
					{/each}
				</div>
			</div>

			<!-- Notes Section -->
			<div class="space-y-2">
				<label for="assign_notes" class="block text-xs font-bold text-gray-500 dark:text-gray-400">Catatan / Keterangan (Opsional)</label>
				<textarea id="assign_notes" bind:value={assignNotes} rows="3" placeholder="Masukkan keterangan tambahan jika diperlukan..." class="w-full p-3.5 text-xs bg-gray-50 dark:bg-gray-900 border border-gray-100 dark:border-gray-700 rounded-2xl focus:outline-none focus:ring-2 focus:ring-indigo-500/20 focus:border-indigo-500 transition resize-none"></textarea>
			</div>

			{#if assignError}
				<div class="p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 text-xs rounded-xl font-semibold flex items-center gap-2">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
					<span>{assignError}</span>
				</div>
			{/if}
		</div>

		<!-- Footer -->
		<div class="p-6 border-t border-gray-100 dark:border-gray-700 flex items-center gap-3">
			{#if assignEntryId}
				<button onclick={deleteAssign} class="px-4 py-2.5 border border-red-200 dark:border-red-800 text-red-600 dark:text-red-400 rounded-xl text-xs font-bold hover:bg-red-50 dark:hover:bg-red-900/10 transition cursor-pointer flex items-center gap-1.5 shrink-0" title="Kosongkan jadwal hari ini (Libur)">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.34 9m-4.78 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
					Libur
				</button>
			{/if}
			<button onclick={closeAssign} class="flex-1 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-xs font-semibold text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer text-center">Batal</button>
			<button onclick={saveAssign} class="flex-1 py-2.5 bg-[#1A56DB] hover:bg-[#1e40af] text-white rounded-xl text-xs font-bold transition cursor-pointer shadow-md text-center">Simpan Jadwal</button>
		</div>
	</div>
{/if}

<!-- Delete Employee Confirmation Modal -->
{#if showDeleteConfirm}
	<div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4" onclick={cancelRemoveEmployee} role="presentation">
		<div onclick={(e) => e.stopPropagation()} role="dialog" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm p-6 text-center">
			<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center">
				<svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
			</div>
			<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Hapus dari Roster</h3>
			<p class="text-xs text-gray-500 dark:text-gray-400 mb-1">Apakah Anda yakin ingin menghapus</p>
			<p class="text-sm font-semibold text-gray-900 dark:text-white mb-2">"{deleteEmployeeName}"?</p>
			<p class="text-xs text-red-500 dark:text-red-400 mb-4 bg-red-50 dark:bg-red-900/25 p-2 rounded-lg">Semua entri jadwal kerja karyawan ini di roster bulan ini akan dihapus permanen.</p>
			<div class="flex items-center justify-center gap-3">
				<button onclick={cancelRemoveEmployee} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-xl text-xs font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
				<button onclick={confirmRemoveEmployee} class="px-5 py-2.5 bg-red-600 text-white rounded-xl text-xs font-semibold hover:bg-red-700 transition cursor-pointer">Ya, Hapus</button>
			</div>
		</div>
	</div>
{/if}

