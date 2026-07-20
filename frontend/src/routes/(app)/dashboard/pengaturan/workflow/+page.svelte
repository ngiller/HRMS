<script lang="ts">
	import { onMount } from 'svelte';
	import { approvalWorkflows, employees } from '$lib/api.js';
	import { slide, fly } from 'svelte/transition';

	const ENTITY_LABELS: Record<string, string> = {
		leave: 'Cuti',
		overtime: 'Lembur',
		reimbursement: 'Reimbursement',
		shift_change: 'Permintaan Shift',
		loan: 'Pinjaman',
		manual_attendance: 'Absensi Manual',
		resign: 'Resign',
		mutation: 'Mutasi',
	};

	const ENTITY_ICONS: Record<string, string> = {
		leave: 'M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z',
		overtime: 'M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z',
		reimbursement: 'M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z',
		shift_change: 'M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0l3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182',
		loan: 'M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Z',
		manual_attendance: 'M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z',
		resign: 'M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3',
		mutation: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Z',
	};

	const ENTITY_COLORS: Record<string, string> = {
		leave: 'from-blue-500 to-blue-600',
		overtime: 'from-amber-500 to-amber-600',
		reimbursement: 'from-emerald-500 to-emerald-600',
		shift_change: 'from-purple-500 to-purple-600',
		loan: 'from-rose-500 to-rose-600',
		manual_attendance: 'from-cyan-500 to-cyan-600',
		resign: 'from-orange-500 to-orange-600',
		mutation: 'from-indigo-500 to-indigo-600',
	};

	const APPROVER_LABELS: Record<string, string> = {
		approval_line: 'Atasan Langsung',
		hr_manager: 'HR Manager',
		finance: 'Finance',
		director: 'Direktur',
		department_head: 'Kepala Departemen',
		manager: 'Manager',
		specific_employee: 'Karyawan Tertentu',
	};

	const ENTITY_ORDER = ['leave', 'overtime', 'reimbursement', 'shift_change', 'loan', 'manual_attendance', 'resign', 'mutation'];

	type WorkflowSummary = {
		id: string;
		entity_type: string;
		name: string;
		description: string;
		is_active: boolean;
		step_count: number;
	};

	type WorkflowStep = {
		id: string;
		workflow_id: string;
		step_order: number;
		approver_type: string;
		approver_role_id: string | null;
		approver_employee_id: string | null;
		approver_employee_name: string | null;
		step_mode?: string;
		condition_field: string | null;
		condition_operator: string | null;
		condition_value: number | null;
		escalation_hours: number | null;
	};

	type WorkflowDetail = {
		workflow: {
			id: string;
			entity_type: string;
			name: string;
			description: string;
			is_active: boolean;
			steps: WorkflowStep[];
		};
		steps: WorkflowStep[];
	};

	let workflows = $state<WorkflowSummary[]>([]);
	let workflowDetails = $state<Record<string, WorkflowDetail | null>>({});
	let isLoading = $state(true);
	let errorMessage = $state('');
	let successMessage = $state('');

	// ═══ Inline Create Workflow ═══
	let showCreateForm = $state(false);
	let createForm = $state({ name: '', entity_type: 'leave', description: '' });
	let createError = $state('');
	let isCreating = $state(false);

	// ═══ Inline Edit Workflow (per workflow ID) ═══
	let editingWorkflowId = $state<string | null>(null);
	let editForm = $state({ name: '', description: '' });
	let editError = $state('');
	let isEditing = $state(false);

	// ═══ Employee list for specific_employee picker ═══
	let employeeList = $state<{ id: string; full_name: string; department_name: string }[]>([]);
	let employeeSearch = $state('');
	let filteredEmployees = $derived.by(() => {
		if (!employeeSearch.trim()) return employeeList.slice(0, 20);
		const q = employeeSearch.toLowerCase();
		return employeeList.filter(e => e.full_name.toLowerCase().includes(q) || (e.department_name || '').toLowerCase().includes(q)).slice(0, 20);
	});
	let showEmployeePicker = $state(false);

	// ═══ Inline Edit Step ═══
	let editingStepId = $state<string | null>(null);
	let editingStepWfId = $state<string | null>(null);
	let editStepForm = $state({
		step_order: 1,
		approver_type: 'approval_line',
		step_mode: 'single',
		condition_field: '',
		condition_operator: '',
		condition_value: 0,
		escalation_hours: 0,
		approver_employee_id: '',
		approver_employee_name: '',
	});
	let editStepError = $state('');
	let isEditingStep = $state(false);

	// ═══ Inline Add Step (per workflow ID) ═══
	let addingStepForId = $state<string | null>(null);
	let stepForm = $state({
		step_order: 1,
		approver_type: 'approval_line',
		step_mode: 'single',
		condition_field: '',
		condition_operator: '',
		condition_value: 0,
		escalation_hours: 0,
		approver_employee_id: '',
		approver_employee_name: '',
	});
	let stepError = $state('');
	let isAddingStep = $state(false);

	// ═══ Inline Delete Confirmation ═══
	let confirmingDeleteId = $state<string | null>(null);
	let isDeleting = $state(false);

	// Expanded workflows
	let expandedWorkflows = $state<Set<string>>(new Set());

	onMount(() => {
		loadWorkflows();
	});

	async function loadWorkflows() {
		isLoading = true;
		errorMessage = '';
		try {
			const res = await approvalWorkflows.list();
			if (res.success) {
				workflows = res.data || [];
			} else {
				workflows = [];
			}
		} catch (e: any) {
			errorMessage = e?.message || 'Gagal memuat data workflow';
		} finally {
			isLoading = false;
		}
	}

	async function loadWorkflowDetail(id: string) {
		if (workflowDetails[id]) return;
		try {
			const res = await approvalWorkflows.get(id);
			if (res.success) {
				workflowDetails[id] = res.data;
			}
		} catch {
			workflowDetails[id] = null;
		}
	}

	function getWorkflowsByType(type: string): WorkflowSummary[] {
		return workflows.filter(w => w.entity_type === type);
	}

	function getEntityTypeCount(): { type: string; count: number }[] {
		const counts: Record<string, number> = {};
		for (const w of workflows) {
			counts[w.entity_type] = (counts[w.entity_type] || 0) + 1;
		}
		return ENTITY_ORDER
			.filter(t => counts[t])
			.map(t => ({ type: t, count: counts[t] || 0 }));
	}

	// ─── Create Workflow (inline) ───────────────────────────────

	function toggleCreateForm(entityType?: string) {
		showCreateForm = !showCreateForm;
		createForm = { name: '', entity_type: entityType || (showCreateForm ? createForm.entity_type : 'leave'), description: '' };
		createError = '';
		// If clicking a different entity type while form is open, just update the type
		if (entityType && showCreateForm) {
			showCreateForm = true;
			createForm.entity_type = entityType;
		}
	}

	async function handleCreate() {
		if (!createForm.name.trim()) { createError = 'Nama workflow harus diisi'; return; }
		if (!createForm.entity_type) { createError = 'Tipe entity harus dipilih'; return; }

		isCreating = true;
		createError = '';
		try {
			await approvalWorkflows.create({
				name: createForm.name.trim(),
				entity_type: createForm.entity_type,
				description: createForm.description.trim(),
			});
			showCreateForm = false;
			successMessage = `Workflow "${createForm.name.trim()}" berhasil dibuat!`;
			setTimeout(() => successMessage = '', 4000);
			await loadWorkflows();
		} catch (e: any) {
			createError = e?.message || 'Gagal membuat workflow';
		} finally {
			isCreating = false;
		}
	}

	// ─── Edit Workflow (inline) ─────────────────────────────────

	function toggleEdit(wf: WorkflowSummary) {
		if (editingWorkflowId === wf.id) {
			editingWorkflowId = null;
		} else {
			editingWorkflowId = wf.id;
			editForm = { name: wf.name, description: wf.description };
			editError = '';
		}
	}

	async function handleEdit() {
		if (!editForm.name.trim()) { editError = 'Nama workflow harus diisi'; return; }
		if (!editingWorkflowId) return;

		isEditing = true;
		editError = '';
		try {
			const wf = workflows.find(w => w.id === editingWorkflowId);
			await approvalWorkflows.update(editingWorkflowId, {
				name: editForm.name.trim(),
				entity_type: wf?.entity_type || 'leave',
				description: editForm.description.trim(),
			});
			editingWorkflowId = null;
			successMessage = `Workflow "${editForm.name.trim()}" berhasil diperbarui!`;
			setTimeout(() => successMessage = '', 4000);
			await loadWorkflows();
		} catch (e: any) {
			editError = e?.message || 'Gagal memperbarui workflow';
		} finally {
			isEditing = false;
		}
	}

	// ─── Add Step (inline) ──────────────────────────────────────

	async function loadEmployeeList() {
		try {
			const res = await employees.list(1, 100);
			if (res.success) {
				const emps = res.data || [];
				employeeList = emps.map((e: any) => ({
					id: e.id,
					full_name: e.full_name,
					department_name: e.department_name || '',
				}));
			}
		} catch {
			// silent
		}
	}

	function toggleAddStep(workflowId: string) {
		if (addingStepForId === workflowId) {
			addingStepForId = null;
		} else {
			// Tutup edit workflow kalo lagi kebuka
			editingWorkflowId = null;
			
			const detail = workflowDetails[workflowId];
			const steps = detail?.steps || [];
			const maxOrder = steps.reduce((max, s) => Math.max(max, s.step_order), 0);

			addingStepForId = workflowId;
			stepForm = {
				step_order: maxOrder + 1,
				approver_type: 'approval_line',
				step_mode: 'single',
				condition_field: '',
				condition_operator: '',
				condition_value: 0,
				escalation_hours: 0,
				approver_employee_id: '',
				approver_employee_name: '',
			};
			employeeSearch = '';
			showEmployeePicker = false;
			stepError = '';
			loadEmployeeList();
		}
	}

	async function handleAddStep() {
		if (!stepForm.approver_type) { stepError = 'Tipe approver harus dipilih'; return; }
		if (!addingStepForId) return;
		if (stepForm.approver_type === 'specific_employee' && !stepForm.approver_employee_id) {
			stepError = 'Pilih karyawan yang akan menjadi approver';
			return;
		}

		isAddingStep = true;
		stepError = '';
		const wid = addingStepForId; // capture before async ops
		try {
			await approvalWorkflows.addStep(addingStepForId, {
				step_order: stepForm.step_order,
				approver_type: stepForm.approver_type,
				step_mode: stepForm.step_mode,
				approver_employee_id: stepForm.approver_employee_id || null,
				condition_field: stepForm.condition_field || null,
				condition_operator: stepForm.condition_operator || null,
				condition_value: stepForm.condition_value || null,
				escalation_hours: stepForm.escalation_hours || null,
			});
			addingStepForId = null;
			successMessage = `Step ${stepForm.step_order} berhasil ditambahkan!`;
			setTimeout(() => successMessage = '', 4000);
			// Refresh detail
			workflowDetails[wid] = null;
			await loadWorkflowDetail(wid);
			await loadWorkflows();
		} catch (e: any) {
			stepError = e?.message || 'Gagal menambahkan step';
		} finally {
			isAddingStep = false;
		}
	}

	// ─── Edit Step (inline) ────────────────────────────────────

	function toggleEditStep(step: WorkflowStep, wfId: string) {
		if (editingStepId === step.id) {
			editingStepId = null;
			editingStepWfId = null;
		} else {
			// Tutup add step kalo lagi kebuka
			addingStepForId = null;
			
			editingStepId = step.id;
			editingStepWfId = wfId;
			editStepForm = {
				step_order: step.step_order,
				approver_type: step.approver_type,
				step_mode: step.step_mode || 'single',
				condition_field: step.condition_field || '',
				condition_operator: step.condition_operator || '',
				condition_value: step.condition_value || 0,
				escalation_hours: step.escalation_hours || 0,
				approver_employee_id: step.approver_employee_id || '',
				approver_employee_name: step.approver_employee_name || '',
			};
			employeeSearch = '';
			showEmployeePicker = false;
			editStepError = '';
			loadEmployeeList();
		}
	}

	async function handleUpdateStep() {
		if (!editingStepId) return;
		if (!editStepForm.approver_type) { editStepError = 'Tipe approver harus dipilih'; return; }
		if (editStepForm.approver_type === 'specific_employee' && !editStepForm.approver_employee_id) {
			editStepError = 'Pilih karyawan yang akan menjadi approver';
			return;
		}

		isEditingStep = true;
		editStepError = '';
		const wfId = editingStepWfId;
		try {
			await approvalWorkflows.updateStep(editingStepId, {
				step_order: editStepForm.step_order,
				approver_type: editStepForm.approver_type,
				step_mode: editStepForm.step_mode,
				approver_employee_id: editStepForm.approver_employee_id || null,
				condition_field: editStepForm.condition_field || null,
				condition_operator: editStepForm.condition_operator || null,
				condition_value: editStepForm.condition_value || null,
				escalation_hours: editStepForm.escalation_hours || null,
			});
			editingStepId = null;
			editingStepWfId = null;
			successMessage = `Step ${editStepForm.step_order} berhasil diperbarui!`;
			setTimeout(() => successMessage = '', 4000);
			if (wfId) {
				workflowDetails[wfId] = null;
				await loadWorkflowDetail(wfId);
			}
			await loadWorkflows();
		} catch (e: any) {
			editStepError = e?.message || 'Gagal memperbarui step';
		} finally {
			isEditingStep = false;
		}
	}

	// ─── Delete Step ────────────────────────────────────────────

	let deletingStep = $state<string | null>(null);

	async function handleDeleteStep(stepId: string, workflowId: string) {
		if (deletingStep) return;
		deletingStep = stepId;
		try {
			await approvalWorkflows.removeStep(stepId);
			successMessage = 'Step berhasil dihapus!';
			setTimeout(() => successMessage = '', 4000);
			workflowDetails[workflowId] = null;
			await loadWorkflowDetail(workflowId);
			await loadWorkflows();
		} catch (e: any) {
			errorMessage = e?.message || 'Gagal menghapus step';
		} finally {
			deletingStep = null;
		}
	}

	// ─── Delete Workflow (inline) ───────────────────────────────

	function toggleDeleteConfirm(id: string) {
		confirmingDeleteId = confirmingDeleteId === id ? null : id;
	}

	async function handleDeleteWorkflow() {
		if (!confirmingDeleteId) return;
		isDeleting = true;
		try {
			const wf = workflows.find(w => w.id === confirmingDeleteId);
			await approvalWorkflows.remove(confirmingDeleteId);
			successMessage = `Workflow "${wf?.name}" berhasil dihapus!`;
			setTimeout(() => successMessage = '', 4000);
			delete workflowDetails[confirmingDeleteId];
			confirmingDeleteId = null;
			await loadWorkflows();
		} catch (e: any) {
			errorMessage = e?.message || 'Gagal menghapus workflow';
		} finally {
			isDeleting = false;
		}
	}

	// ─── Expand / Collapse ──────────────────────────────────────

	function toggleExpand(id: string) {
		if (expandedWorkflows.has(id)) {
			expandedWorkflows.delete(id);
		} else {
			expandedWorkflows.add(id);
			loadWorkflowDetail(id);
		}
		expandedWorkflows = new Set(expandedWorkflows);
	}

	function isExpanded(id: string): boolean {
		return expandedWorkflows.has(id);
	}

	function getApproverLabel(type: string): string {
		return APPROVER_LABELS[type] || type;
	}
</script>

<svelte:head>
	<title>Workflow Approval - HRMS</title>
</svelte:head>

<div class="w-full animate-in fade-in duration-500">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div class="flex items-center gap-4">
			<div class="w-14 h-14 bg-gradient-to-br from-violet-500 to-purple-600 rounded-2xl flex items-center justify-center text-white shadow-lg shadow-purple-500/30">
				<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
				</svg>
			</div>
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Workflow Approval</h1>
				<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Konfigurasi alur persetujuan untuk setiap modul</p>
			</div>
		</div>
		<button onclick={() => toggleCreateForm()}
			class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
			Buat Workflow Baru
		</button>
	</div>

	<!-- ═══ INLINE CREATE WORKFLOW FORM ═══ -->
	{#if showCreateForm}
		<div transition:slide={{ duration: 200 }} class="mb-6 bg-white dark:bg-gray-800/50 rounded-2xl border border-blue-200 dark:border-blue-800 shadow-sm overflow-hidden">
			<div class="px-6 py-5">
				<div class="flex items-center gap-3 mb-5">
					<div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
						<svg class="w-5 h-5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					</div>
					<div>
						<h3 class="text-sm font-bold text-gray-900 dark:text-white">Buat Workflow Baru</h3>
						<p class="text-xs text-gray-500 dark:text-gray-400">Isi detail workflow approval</p>
					</div>
				</div>

				{#if createError}
					<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5 text-sm text-red-700 dark:text-red-300">{createError}</div>
				{/if}

				<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Workflow <span class="text-red-500">*</span></label>
						<input type="text" bind:value={createForm.name}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400"
							placeholder="Contoh: Workflow Cuti Tahunan" />
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Module <span class="text-red-500">*</span></label>
						<select bind:value={createForm.entity_type}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
							{#each ENTITY_ORDER as type (type)}
								<option value={type}>{ENTITY_LABELS[type]}</option>
							{/each}
						</select>
					</div>
					<div>
						<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
						<input type="text" bind:value={createForm.description}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400"
							placeholder="Deskripsi (opsional)" />
					</div>
				</div>

				<div class="flex items-center justify-end gap-3 mt-5 pt-4 border-t border-gray-100 dark:border-gray-700">
					<button onclick={() => showCreateForm = false} disabled={isCreating}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">
						Batal
					</button>
					<button onclick={handleCreate} disabled={isCreating}
						class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isCreating}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							Menyimpan...
						{:else}
							Buat Workflow
						{/if}
					</button>
				</div>
			</div>
		</div>
	{/if}

	<!-- Success Message -->
	{#if successMessage}
		<div transition:fly={{ y: -10, duration: 300 }} class="mb-4 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl px-5 py-3 flex items-center gap-2.5">
			<svg class="w-5 h-5 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
			<p class="text-sm font-medium text-emerald-800 dark:text-emerald-200">{successMessage}</p>
		</div>
	{/if}

	{#if isLoading}
		<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
			{#each [1,2,3,4,5,6] as _ (_)}
				<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 p-5 animate-pulse">
					<div class="flex items-center gap-3 mb-4">
						<div class="w-10 h-10 rounded-xl bg-gray-100 dark:bg-gray-700"></div>
						<div class="flex-1"><div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-28 mb-2"></div><div class="h-3 bg-gray-50 dark:bg-gray-700/50 rounded w-20"></div></div>
					</div>
					<div class="h-3 bg-gray-50 dark:bg-gray-700/50 rounded w-full mb-3"></div>
					<div class="flex gap-2"><div class="h-8 bg-gray-50 dark:bg-gray-700/50 rounded-lg w-20"></div><div class="h-8 bg-gray-50 dark:bg-gray-700/50 rounded-lg w-20"></div></div>
				</div>
			{/each}
		</div>
	{:else if errorMessage}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-4">
			<div class="flex items-center gap-2.5">
				<svg class="w-5 h-5 text-red-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				<p class="text-sm font-medium text-red-800 dark:text-red-200">{errorMessage}</p>
			</div>
		</div>
		<div class="mt-4 text-center"><button onclick={loadWorkflows} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button></div>
	{:else if workflows.length === 0}
		<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 p-16 text-center">
			<div class="w-20 h-20 mx-auto mb-5 rounded-full bg-purple-50 dark:bg-purple-900/20 flex items-center justify-center">
				<svg class="w-10 h-10 text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
			</div>
			<h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2">Belum Ada Workflow</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400 max-w-md mx-auto mb-6">Buat workflow approval untuk mengatur alur persetujuan setiap modul.</p>
			<button onclick={() => toggleCreateForm()}
				class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Buat Workflow Pertama
			</button>
		</div>
	{:else}
		<!-- Entity Type Cards -->
		{@const typesWithWorkflows = getEntityTypeCount()}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
			{#each typesWithWorkflows as { type, count } (type)}
				{@const workflows = getWorkflowsByType(type)}
				<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm hover:shadow-md transition-shadow">
					<!-- Card Header -->
					<div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700/50 flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-xl bg-gradient-to-br {ENTITY_COLORS[type] || 'from-gray-400 to-gray-500'} flex items-center justify-center text-white shadow-sm">
								<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d={ENTITY_ICONS[type] || 'M12 6v12m6-6H6'} /></svg>
							</div>
							<div>
								<h3 class="text-sm font-bold text-gray-900 dark:text-white">{ENTITY_LABELS[type] || type}</h3>
								<p class="text-xs text-gray-500 dark:text-gray-400">{count} workflow</p>
							</div>
						</div>
						<button onclick={() => toggleCreateForm(type)}
							class="p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 dark:hover:bg-blue-900/20 transition cursor-pointer"
							title="Tambah workflow untuk {ENTITY_LABELS[type] || type}">
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
						</button>
					</div>

					<!-- Workflow items -->
					<div class="divide-y divide-gray-50 dark:divide-gray-700/30">
						{#each workflows as wf (wf.id)}
							<div class="transition-colors hover:bg-gray-50/50 dark:hover:bg-gray-800/30">
								<!-- Collapsible header -->
								<div onclick={() => toggleExpand(wf.id)} onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') toggleExpand(wf.id); }}
									role="button" tabindex="0"
									class="w-full flex items-center justify-between px-5 py-3.5 text-left cursor-pointer group">
									<div class="flex items-center gap-3 min-w-0">
										<svg class="w-4 h-4 text-gray-300 dark:text-gray-600 shrink-0 transition-transform duration-200 {isExpanded(wf.id) ? 'rotate-90' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="m8.25 4.5 7.5 7.5-7.5 7.5" />
										</svg>
										<div class="min-w-0 flex-1">
											<div class="flex items-center gap-2">
												<p class="text-sm font-semibold text-gray-900 dark:text-white truncate group-hover:text-[#1A56DB] dark:group-hover:text-blue-400 transition-colors">{wf.name}</p>
												<button onclick={(e) => { e.stopPropagation(); toggleEdit(wf); }}
													class="p-1 rounded text-gray-300 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 opacity-0 group-hover:opacity-100 transition-all cursor-pointer"
													title="Edit workflow">
													<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
												</button>
											</div>
											{#if wf.description}
												<p class="text-xs text-gray-500 dark:text-gray-400 truncate">{wf.description}</p>
											{/if}
										</div>
									</div>
									<div class="flex items-center gap-3 shrink-0">
										<span class="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium {wf.is_active ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300' : 'bg-gray-100 text-gray-500 dark:bg-gray-700 dark:text-gray-400'}">
											{wf.is_active ? 'Aktif' : 'Nonaktif'}
										</span>
										<span class="text-xs text-gray-400 tabular-nums">{wf.step_count} step</span>
									</div>
								</div>

								<!-- Expanded section -->
								{#if isExpanded(wf.id)}
									{@const detail = workflowDetails[wf.id]}
									<div class="px-5 pb-4 space-y-2 border-t border-gray-50 dark:border-gray-700/30 pt-3" transition:slide={{ duration: 200 }}>
										
										<!-- ═══ INLINE EDIT WORKFLOW FORM ═══ -->
										{#if editingWorkflowId === wf.id}
											<div class="mb-4 p-4 bg-blue-50 dark:bg-blue-900/10 rounded-xl border border-blue-200 dark:border-blue-800/50">
												{#if editError}
													<div class="mb-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-3 py-2 text-xs text-red-700 dark:text-red-300">{editError}</div>
												{/if}
												<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3">
													<div>
														<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Nama Workflow <span class="text-red-500">*</span></label>
														<input type="text" bind:value={editForm.name}
															class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
													</div>
													<div>
														<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Deskripsi</label>
														<input type="text" bind:value={editForm.description}
															class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
													</div>
												</div>
												<div class="flex items-center gap-2 justify-end">
													<button onclick={() => editingWorkflowId = null} disabled={isEditing}
														class="px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition cursor-pointer">Batal</button>
													<button onclick={handleEdit} disabled={isEditing}
														class="px-3 py-1.5 text-xs font-semibold bg-[#1A56DB] text-white rounded-lg hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
														{#if isEditing}
															<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
														{/if}
														Simpan
													</button>
												</div>
											</div>
										{/if}

										<!-- Steps list -->
										{#if detail && detail.steps}
											{#if detail.steps.length === 0}
												<p class="text-xs text-gray-400 italic text-center py-3">Belum ada step approval. Tambahkan step pertama.</p>
											{:else}
												{#each detail.steps as step (step.id)}
													{#if editingStepId === step.id}
														<!-- ═══ INLINE EDIT STEP FORM ═══ -->
														<div class="p-3 bg-blue-50 dark:bg-blue-900/10 rounded-lg border border-blue-200 dark:border-blue-800/50" transition:slide={{ duration: 200 }}>
															{#if editStepError}
																<div class="mb-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-3 py-2 text-xs text-red-700 dark:text-red-300">{editStepError}</div>
															{/if}
															<div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
																<div>
																	<label class="block text-[10px] font-medium text-gray-700 dark:text-gray-300 mb-1">Urutan</label>
																	<input type="number" bind:value={editStepForm.step_order} min="1" max="10"
																		class="w-full px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded-lg outline-none focus:ring-1 focus:ring-blue-400/20 focus:border-blue-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
																</div>
																<div>
																	<label class="block text-[10px] font-medium text-gray-700 dark:text-gray-300 mb-1">Tipe Approver</label>
																	<select bind:value={editStepForm.approver_type}
																		class="w-full px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded-lg outline-none focus:ring-1 focus:ring-blue-400/20 focus:border-blue-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
																		{#each Object.entries(APPROVER_LABELS) as [value, label] (value)}
																			<option value={value}>{label}</option>
																		{/each}
																	</select>
																</div>
																{#if editStepForm.approver_type === 'specific_employee'}
																<div class="sm:col-span-2">
																	<label class="block text-[10px] font-medium text-gray-700 dark:text-gray-300 mb-1">Pilih Karyawan</label>
																	<div class="relative">
																		<div onclick={() => { showEmployeePicker = !showEmployeePicker; employeeSearch = ''; }}
																			class="w-full px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white cursor-pointer flex items-center justify-between">
																			{#if editStepForm.approver_employee_name}
																				<span>{editStepForm.approver_employee_name}</span>
																			{:else}
																				<span class="text-gray-400">Pilih karyawan...</span>
																			{/if}
																			<svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" /></svg>
																		</div>
																		{#if showEmployeePicker}
																			<div class="absolute z-50 mt-1 w-full bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg max-h-48 overflow-hidden flex flex-col">
																				<div class="p-1.5 border-b border-gray-100 dark:border-gray-600">
																					<input type="text" bind:value={employeeSearch}
																						class="w-full px-2 py-1 text-[10px] border border-gray-200 dark:border-gray-500 rounded-md bg-gray-50 dark:bg-gray-600 text-gray-900 dark:text-white outline-none focus:ring-1 focus:ring-blue-400"
																						placeholder="Cari nama..." />
																				</div>
																				<div class="overflow-y-auto flex-1">
																					{#if filteredEmployees.length === 0}
																						<p class="text-[10px] text-gray-400 text-center py-2">Tidak ada</p>
																					{:else}
																						{#each filteredEmployees as emp (emp.id)}
																							<button onclick={() => {
																								editStepForm.approver_employee_id = emp.id;
																								editStepForm.approver_employee_name = emp.full_name;
																								showEmployeePicker = false;
																							}}
																								class="w-full text-left px-2.5 py-1.5 text-[10px] hover:bg-blue-50 dark:hover:bg-blue-900/20 transition cursor-pointer flex items-center justify-between">
																								<span class="font-medium text-gray-900 dark:text-white">{emp.full_name}</span>
																								{#if emp.department_name}
																									<span class="text-gray-400">{emp.department_name}</span>
																								{/if}
																							</button>
																						{/each}
																					{/if}
																				</div>
																			</div>
																		{/if}
																	</div>
																</div>
																{/if}
																<div>
																	<label class="block text-[10px] font-medium text-gray-700 dark:text-gray-300 mb-1">Eskalasi (jam)</label>
																	<input type="number" bind:value={editStepForm.escalation_hours} min="0" max="168"
																		class="w-full px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-600 rounded-lg outline-none focus:ring-1 focus:ring-blue-400/20 focus:border-blue-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
																</div>
																<div>
																	<label class="block text-[10px] font-medium text-gray-700 dark:text-gray-300 mb-1.5">Mode</label>
																	<div class="flex items-center gap-1.5">
																		<label class="flex items-center gap-1 px-2 py-1.5 border rounded-md cursor-pointer text-[10px] {editStepForm.step_mode === 'single' ? 'border-blue-500 bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 font-medium' : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400'}">
																			<input type="radio" name="edit_step_mode" value="single" bind:group={editStepForm.step_mode} class="sr-only" />
																			Single
																		</label>
																		<label class="flex items-center gap-1 px-2 py-1.5 border rounded-md cursor-pointer text-[10px] {editStepForm.step_mode === 'any' ? 'border-amber-400 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-300 font-medium' : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400'}">
																			<input type="radio" name="edit_step_mode" value="any" bind:group={editStepForm.step_mode} class="sr-only" />
																			Parallel
																		</label>
																	</div>
																</div>
															</div>
															<div class="flex items-center justify-end gap-2 mt-2 pt-2 border-t border-blue-200/50 dark:border-blue-800/30">
																<button onclick={() => { editingStepId = null; editingStepWfId = null; }} disabled={isEditingStep}
																	class="px-2.5 py-1 text-[10px] font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition cursor-pointer">Batal</button>
																<button onclick={handleUpdateStep} disabled={isEditingStep}
																	class="px-2.5 py-1 text-[10px] font-semibold bg-blue-500 text-white rounded-lg hover:bg-blue-600 transition disabled:opacity-50 inline-flex items-center gap-1 cursor-pointer">
																	{#if isEditingStep}
																		<svg class="w-2.5 h-2.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
																	{/if}
																	Simpan
																</button>
															</div>
														</div>
													{:else}
													<div class="flex items-center gap-3 px-3 py-2.5 bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-gray-100 dark:border-gray-700/50 group/step">
														<div class="w-7 h-7 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white text-[10px] font-bold shrink-0 shadow-sm">{step.step_order}</div>
													<div class="flex-1 min-w-0">
														<p class="text-sm font-medium text-gray-900 dark:text-white">
															{getApproverLabel(step.approver_type)}
															{#if step.approver_type === 'specific_employee' && step.approver_employee_name}
																<span class="text-xs text-amber-600 dark:text-amber-400 font-normal">— {step.approver_employee_name}</span>
															{/if}
														</p>
														<div class="flex items-center gap-2 mt-0.5">
															{#if step.step_mode === 'any'}
																<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[9px] font-bold uppercase tracking-wider bg-amber-50 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300 border border-amber-200 dark:border-amber-800">
																	<svg class="w-2.5 h-2.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M18 7.5v3m0 0v3m0-3h3m-3 0h-3m-2.25-4.125a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0ZM3 20.25a7.125 7.125 0 0 1 14.25 0" /></svg>
																	Parallel
																</span>
															{/if}
															{#if step.condition_field}
																<span class="text-[10px] text-gray-500 dark:text-gray-400">Syarat: {step.condition_field} {step.condition_operator} {step.condition_value}</span>
															{/if}
														</div>
													</div>
													{#if step.escalation_hours}
														<span class="text-[10px] text-amber-600 dark:text-amber-400 font-medium">Eskalasi {step.escalation_hours}j</span>
													{/if}
													<button onclick={() => toggleEditStep(step, wf.id)}
														class="p-1 rounded text-gray-300 hover:text-blue-500 hover:bg-blue-50 dark:hover:bg-blue-900/20 opacity-0 group-hover/step:opacity-100 transition-all cursor-pointer"
														title="Edit step">
														<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
													</button>
													<button onclick={() => handleDeleteStep(step.id, wf.id)} disabled={deletingStep === step.id}
														class="p-1 rounded text-gray-300 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 opacity-0 group-hover/step:opacity-100 transition-all cursor-pointer disabled:opacity-30"
														title="Hapus step">
														<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
													</button>
												</div>
											{/if}
											{/each}
											{/if}
										{:else}
											<div class="flex items-center justify-center py-3">
												<div class="animate-spin h-4 w-4 border-2 border-blue-600 border-t-transparent rounded-full"></div>
											</div>
										{/if}

										<!-- ═══ INLINE ADD STEP FORM ═══ -->
										{#if addingStepForId === wf.id}
											<div class="mt-3 p-4 bg-amber-50 dark:bg-amber-900/10 rounded-xl border border-amber-200 dark:border-amber-800/50" transition:slide={{ duration: 200 }}>
												{#if stepError}
													<div class="mb-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-3 py-2 text-xs text-red-700 dark:text-red-300">{stepError}</div>
												{/if}
												<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
													<div>
														<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Urutan Step</label>
														<input type="number" bind:value={stepForm.step_order} min="1" max="10"
															class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-amber-400/20 focus:border-amber-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
													</div>
											<div>
												<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Tipe Approver <span class="text-red-500">*</span></label>
												<select bind:value={stepForm.approver_type}
													class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-amber-400/20 focus:border-amber-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
													{#each Object.entries(APPROVER_LABELS) as [value, label] (value)}
														<option value={value}>{label}</option>
													{/each}
												</select>
											</div>

											<!-- Employee picker for specific_employee -->
											{#if stepForm.approver_type === 'specific_employee'}
											<div class="sm:col-span-2">
												<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Pilih Karyawan <span class="text-red-500">*</span></label>
												<div class="relative">
													<div onclick={() => { showEmployeePicker = !showEmployeePicker; employeeSearch = ''; }}
														class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-700 text-gray-900 dark:text-white cursor-pointer flex items-center justify-between">
														{#if stepForm.approver_employee_name}
															<span>{stepForm.approver_employee_name}</span>
														{:else}
															<span class="text-gray-400">Cari & pilih karyawan...</span>
														{/if}
														<svg class="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" /></svg>
													</div>
													
													{#if showEmployeePicker}
														<div class="absolute z-50 mt-1 w-full bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-lg shadow-lg max-h-60 overflow-hidden flex flex-col">
															<div class="p-2 border-b border-gray-100 dark:border-gray-600">
																<input type="text" bind:value={employeeSearch}
																	class="w-full px-2.5 py-1.5 text-xs border border-gray-200 dark:border-gray-500 rounded-md bg-gray-50 dark:bg-gray-600 text-gray-900 dark:text-white outline-none focus:ring-1 focus:ring-amber-400"
																	placeholder="Cari nama karyawan..." />
															</div>
															<div class="overflow-y-auto flex-1">
																{#if filteredEmployees.length === 0}
																	<p class="text-xs text-gray-400 text-center py-3">Tidak ada karyawan ditemukan</p>
																{:else}
																	{#each filteredEmployees as emp (emp.id)}
																		<button onclick={() => {
																			stepForm.approver_employee_id = emp.id;
																			stepForm.approver_employee_name = emp.full_name;
																			showEmployeePicker = false;
																		}}
																			class="w-full text-left px-3 py-2 text-xs hover:bg-amber-50 dark:hover:bg-amber-900/20 transition cursor-pointer flex items-center justify-between {stepForm.approver_employee_id === emp.id ? 'bg-amber-50 dark:bg-amber-900/20' : ''}">
																			<span class="font-medium text-gray-900 dark:text-white">{emp.full_name}</span>
																			{#if emp.department_name}
																				<span class="text-gray-400">{emp.department_name}</span>
																			{/if}
																		</button>
																	{/each}
																{/if}
															</div>
														</div>
													{/if}
												</div>
											</div>
											{/if}

											<div>
												<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">Eskalasi (jam)</label>
														<input type="number" bind:value={stepForm.escalation_hours} min="0" max="168"
															class="w-full px-3 py-2 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-amber-400/20 focus:border-amber-400 transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
															placeholder="0" />
													</div>
													<div>
														<label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-2">Mode</label>
														<div class="flex items-center gap-2">
															<label class="flex items-center gap-1.5 px-3 py-2 border rounded-lg cursor-pointer transition-all text-xs {stepForm.step_mode === 'single' ? 'border-[#1A56DB] bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300 font-medium' : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-500'}">
																<input type="radio" name="step_mode_{wf.id}" value="single" bind:group={stepForm.step_mode} class="sr-only" />
																<div class="w-3 h-3 rounded-full border-2 flex items-center justify-center {stepForm.step_mode === 'single' ? 'border-[#1A56DB]' : 'border-gray-300 dark:border-gray-500'}">
																	{#if stepForm.step_mode === 'single'}<div class="w-1.5 h-1.5 rounded-full bg-[#1A56DB]"></div>{/if}
																</div>
																Single
															</label>
															<label class="flex items-center gap-1.5 px-3 py-2 border rounded-lg cursor-pointer transition-all text-xs {stepForm.step_mode === 'any' ? 'border-amber-400 bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-300 font-medium' : 'border-gray-200 dark:border-gray-600 text-gray-600 dark:text-gray-400 hover:border-gray-300 dark:hover:border-gray-500'}">
																<input type="radio" name="step_mode_{wf.id}" value="any" bind:group={stepForm.step_mode} class="sr-only" />
																<div class="w-3 h-3 rounded-full border-2 flex items-center justify-center {stepForm.step_mode === 'any' ? 'border-amber-400' : 'border-gray-300 dark:border-gray-500'}">
																	{#if stepForm.step_mode === 'any'}<div class="w-1.5 h-1.5 rounded-full bg-amber-400"></div>{/if}
																</div>
																Parallel
															</label>
														</div>
													</div>
												</div>
												<div class="flex items-center justify-end gap-2 mt-3 pt-3 border-t border-amber-200/50 dark:border-amber-800/30">
													<button onclick={() => addingStepForId = null} disabled={isAddingStep}
														class="px-3 py-1.5 text-xs font-medium text-gray-600 dark:text-gray-400 hover:bg-gray-200 dark:hover:bg-gray-700 rounded-lg transition cursor-pointer">Batal</button>
													<button onclick={handleAddStep} disabled={isAddingStep}
														class="px-3 py-1.5 text-xs font-semibold bg-amber-500 text-white rounded-lg hover:bg-amber-600 transition disabled:opacity-50 inline-flex items-center gap-1.5 cursor-pointer">
														{#if isAddingStep}
															<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
														{/if}
														Tambah Step
													</button>
												</div>
											</div>
										{/if}

										<!-- Action buttons -->
										<div class="flex items-center justify-between pt-2">
											<button onclick={() => toggleAddStep(wf.id)}
												class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/20 rounded-lg hover:bg-amber-100 dark:hover:bg-amber-900/40 transition cursor-pointer">
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
												{addingStepForId === wf.id ? 'Tutup' : 'Tambah Step'}
											</button>

											<!-- ═══ INLINE DELETE CONFIRMATION ═══ -->
											{#if confirmingDeleteId === wf.id}
												<div class="flex items-center gap-2 animate-in fade-in">
													<span class="text-[10px] text-red-600 dark:text-red-400 font-medium">Yakin hapus workflow ini?</span>
													<button onclick={() => confirmingDeleteId = null} disabled={isDeleting}
														class="px-2.5 py-1 text-[10px] font-medium text-gray-600 dark:text-gray-400 bg-gray-100 dark:bg-gray-700 rounded-lg hover:bg-gray-200 dark:hover:bg-gray-600 transition cursor-pointer">Batal</button>
													<button onclick={handleDeleteWorkflow} disabled={isDeleting}
														class="px-2.5 py-1 text-[10px] font-semibold text-white bg-red-600 rounded-lg hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-1 cursor-pointer">
														{#if isDeleting}
															<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
														{/if}
														Ya, Hapus
													</button>
												</div>
											{:else}
												<button onclick={() => toggleDeleteConfirm(wf.id)}
													class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/40 transition cursor-pointer">
													<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
													Hapus Workflow
												</button>
											{/if}
										</div>
									</div>
								{/if}
							</div>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	{/if}
</div>
