<script lang="ts">
	import { onMount } from 'svelte';
	import { approvalWorkflows } from '$lib/api.js';
	import { fade, scale, slide } from 'svelte/transition';
	import { fly } from 'svelte/transition';
	import { AnimatedPresence } from '$lib/index.js';

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

	const ENTITY_BG: Record<string, string> = {
		leave: 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-300',
		overtime: 'bg-amber-50 dark:bg-amber-900/20 text-amber-700 dark:text-amber-300',
		reimbursement: 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-700 dark:text-emerald-300',
		shift_change: 'bg-purple-50 dark:bg-purple-900/20 text-purple-700 dark:text-purple-300',
		loan: 'bg-rose-50 dark:bg-rose-900/20 text-rose-700 dark:text-rose-300',
		manual_attendance: 'bg-cyan-50 dark:bg-cyan-900/20 text-cyan-700 dark:text-cyan-300',
		resign: 'bg-orange-50 dark:bg-orange-900/20 text-orange-700 dark:text-orange-300',
		mutation: 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-700 dark:text-indigo-300',
	};

	const APPROVER_LABELS: Record<string, string> = {
		approval_line: 'Atasan Langsung',
		hr_manager: 'HR Manager',
		finance: 'Finance',
		director: 'Direktur',
		department_head: 'Kepala Departemen',
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

	// Create & Edit workflow modal
	let showCreateModal = $state(false);
	let createForm = $state({ name: '', entity_type: 'leave', description: '' });
	let createError = $state('');
	let isCreating = $state(false);

	let showEditModal = $state(false);
	let editForm = $state({ id: '', name: '', entity_type: 'leave', description: '' });
	let editError = $state('');
	let isEditing = $state(false);

	// Add step modal
	let showAddStepModal = $state(false);
	let addStepWorkflowId = $state('');
	let addStepWorkflowName = $state('');
	let stepForm = $state({
		step_order: 1,
		approver_type: 'approval_line',
		condition_field: '',
		condition_operator: '',
		condition_value: 0,
		escalation_hours: 0,
	});
	let stepError = $state('');
	let isAddingStep = $state(false);

	// Delete confirmation
	let showDeleteModal = $state(false);
	let deleteTarget = $state<{ id: string; name: string } | null>(null);
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

	// ─── Create Workflow ────────────────────────────────────────

	function openCreateModal(entityType?: string) {
		createForm = { name: '', entity_type: entityType || 'leave', description: '' };
		createError = '';
		showCreateModal = true;
	}

	async function handleCreate() {
		// Validasi form
		if (!createForm.name.trim()) {
			createError = 'Nama workflow harus diisi';
			return;
		}
		if (!createForm.entity_type) {
			createError = 'Tipe entity harus dipilih';
			return;
		}

		isCreating = true;
		createError = '';
		try {
			await approvalWorkflows.create({
				name: createForm.name.trim(),
				entity_type: createForm.entity_type,
				description: createForm.description.trim(),
			});
			showCreateModal = false;
			successMessage = `Workflow "${createForm.name.trim()}" berhasil dibuat!`;
			setTimeout(() => successMessage = '', 4000);
			await loadWorkflows();
		} catch (e: any) {
			createError = e?.message || 'Gagal membuat workflow';
		} finally {
			isCreating = false;
		}
	}

	// ─── Edit Workflow ──────────────────────────────────────────

	function openEditModal(wf: WorkflowSummary) {
		editForm = {
			id: wf.id,
			name: wf.name,
			entity_type: wf.entity_type,
			description: wf.description,
		};
		editError = '';
		showEditModal = true;
	}

	async function handleEdit() {
		if (!editForm.name.trim()) {
			editError = 'Nama workflow harus diisi';
			return;
		}
		if (!editForm.entity_type) {
			editError = 'Tipe entity harus dipilih';
			return;
		}

		isEditing = true;
		editError = '';
		try {
			await approvalWorkflows.update(editForm.id, {
				name: editForm.name.trim(),
				entity_type: editForm.entity_type,
				description: editForm.description.trim(),
			});
			showEditModal = false;
			successMessage = `Workflow "${editForm.name.trim()}" berhasil diperbarui!`;
			setTimeout(() => successMessage = '', 4000);
			await loadWorkflows();
		} catch (e: any) {
			editError = e?.message || 'Gagal memperbarui workflow';
		} finally {
			isEditing = false;
		}
	}

	// ─── Add Step ───────────────────────────────────────────────

	function openAddStepModal(workflowId: string, workflowName: string) {
		// Find current max step order
		const detail = workflowDetails[workflowId];
		const steps = detail?.steps || [];
		const maxOrder = steps.reduce((max, s) => Math.max(max, s.step_order), 0);

		addStepWorkflowId = workflowId;
		addStepWorkflowName = workflowName;
		stepForm = {
			step_order: maxOrder + 1,
			approver_type: 'approval_line',
			condition_field: '',
			condition_operator: '',
			condition_value: 0,
			escalation_hours: 0,
		};
		stepError = '';
		showAddStepModal = true;
	}

	async function handleAddStep() {
		if (!stepForm.approver_type) {
			stepError = 'Tipe approver harus dipilih';
			return;
		}

		isAddingStep = true;
		stepError = '';
		try {
			await approvalWorkflows.addStep(addStepWorkflowId, {
				step_order: stepForm.step_order,
				approver_type: stepForm.approver_type,
				condition_field: stepForm.condition_field || null,
				condition_operator: stepForm.condition_operator || null,
				condition_value: stepForm.condition_value || null,
				escalation_hours: stepForm.escalation_hours || null,
			});
			showAddStepModal = false;
			successMessage = `Step ${stepForm.step_order} berhasil ditambahkan ke "${addStepWorkflowName}"!`;
			setTimeout(() => successMessage = '', 4000);
			// Refresh detail
			workflowDetails[addStepWorkflowId] = null;
			await loadWorkflowDetail(addStepWorkflowId);
			await loadWorkflows();
		} catch (e: any) {
			stepError = e?.message || 'Gagal menambahkan step';
		} finally {
			isAddingStep = false;
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

	// ─── Delete Workflow ────────────────────────────────────────

	function openDeleteModal(id: string, name: string) {
		deleteTarget = { id, name };
		showDeleteModal = true;
	}

	async function handleDeleteWorkflow() {
		if (!deleteTarget) return;
		isDeleting = true;
		try {
			await approvalWorkflows.remove(deleteTarget.id);
			showDeleteModal = false;
			successMessage = `Workflow "${deleteTarget.name}" berhasil dihapus!`;
			setTimeout(() => successMessage = '', 4000);
			delete workflowDetails[deleteTarget.id];
			await loadWorkflows();
		} catch (e: any) {
			errorMessage = e?.message || 'Gagal menghapus workflow';
		} finally {
			isDeleting = false;
			deleteTarget = null;
		}
	}

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
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
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
		<button onclick={() => openCreateModal()}
			class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
			Buat Workflow Baru
		</button>
	</div>

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
						<div class="flex-1">
							<div class="h-4 bg-gray-100 dark:bg-gray-700 rounded w-28 mb-2"></div>
							<div class="h-3 bg-gray-50 dark:bg-gray-700/50 rounded w-20"></div>
						</div>
					</div>
					<div class="h-3 bg-gray-50 dark:bg-gray-700/50 rounded w-full mb-3"></div>
					<div class="flex gap-2">
						<div class="h-8 bg-gray-50 dark:bg-gray-700/50 rounded-lg w-20"></div>
						<div class="h-8 bg-gray-50 dark:bg-gray-700/50 rounded-lg w-20"></div>
					</div>
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
		<div class="mt-4 text-center">
			<button onclick={loadWorkflows} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button>
		</div>
	{:else if workflows.length === 0}
		<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 p-16 text-center">
			<div class="w-20 h-20 mx-auto mb-5 rounded-full bg-purple-50 dark:bg-purple-900/20 flex items-center justify-center">
				<svg class="w-10 h-10 text-purple-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
			</div>
			<h3 class="text-base font-semibold text-gray-900 dark:text-gray-100 mb-2">Belum Ada Workflow</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400 max-w-md mx-auto mb-6">Buat workflow approval untuk mengatur alur persetujuan setiap modul seperti cuti, lembur, reimbursement, dan lainnya.</p>
			<button onclick={() => openCreateModal()}
				class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Buat Workflow Pertama
			</button>
		</div>
	{:else}
		<!-- Entity Type Tabs -->
		{@const typesWithWorkflows = getEntityTypeCount()}
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-5">
			{#each typesWithWorkflows as { type, count } (type)}
				{@const workflows = getWorkflowsByType(type)}
				<div class="bg-white dark:bg-gray-800/50 rounded-2xl border border-gray-200 dark:border-gray-700 overflow-hidden shadow-sm hover:shadow-md transition-shadow">
					<!-- Header -->
					<div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700/50 flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-xl bg-gradient-to-br {ENTITY_COLORS[type] || 'from-gray-400 to-gray-500'} flex items-center justify-center text-white shadow-sm">
								<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d={ENTITY_ICONS[type] || 'M12 6v12m6-6H6'} />
								</svg>
							</div>
							<div>
								<h3 class="text-sm font-bold text-gray-900 dark:text-white">{ENTITY_LABELS[type] || type}</h3>
								<p class="text-xs text-gray-500 dark:text-gray-400">{count} workflow</p>
							</div>
						</div>
						<button onclick={() => openCreateModal(type)}
							class="p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 dark:hover:bg-blue-900/20 transition cursor-pointer"
							title="Tambah workflow untuk {ENTITY_LABELS[type] || type}">
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
						</button>
					</div>

					<!-- Workflow cards -->
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
												<button onclick={(e) => { e.stopPropagation(); openEditModal(wf); }}
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

								<!-- Expanded steps -->
								{#if isExpanded(wf.id)}
									{@const detail = workflowDetails[wf.id]}
									<div class="px-5 pb-4 space-y-2 border-t border-gray-50 dark:border-gray-700/30 pt-3" transition:slide={{ duration: 200 }}>
										{#if detail && detail.steps}
											{#if detail.steps.length === 0}
												<p class="text-xs text-gray-400 italic text-center py-3">Belum ada step approval. Tambahkan step pertama.</p>
											{:else}
												{#each detail.steps as step (step.id)}
													<div class="flex items-center gap-3 px-3 py-2.5 bg-gray-50 dark:bg-gray-900/50 rounded-lg border border-gray-100 dark:border-gray-700/50 group/step">
														<div class="w-7 h-7 rounded-full bg-gradient-to-br from-blue-400 to-indigo-500 flex items-center justify-center text-white text-[10px] font-bold shrink-0 shadow-sm">
															{step.step_order}
														</div>
														<div class="flex-1 min-w-0">
															<p class="text-sm font-medium text-gray-900 dark:text-white">{getApproverLabel(step.approver_type)}</p>
															{#if step.condition_field}
																<p class="text-[10px] text-gray-500 dark:text-gray-400">
																	Syarat: {step.condition_field} {step.condition_operator} {step.condition_value}
																</p>
															{/if}
														</div>
														{#if step.escalation_hours}
															<span class="text-[10px] text-amber-600 dark:text-amber-400 font-medium">Eskalasi {step.escalation_hours}jam</span>
														{/if}
								<button onclick={() => handleDeleteStep(step.id, wf.id)} disabled={deletingStep === step.id}
									class="p-1 rounded text-gray-300 hover:text-red-500 hover:bg-red-50 dark:hover:bg-red-900/20 opacity-0 group-hover/step:opacity-100 transition-all cursor-pointer disabled:opacity-30"
									title="Hapus step">
															<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
														</button>
													</div>
												{/each}
											{/if}
										{:else}
											<div class="flex items-center justify-center py-3">
												<div class="animate-spin h-4 w-4 border-2 border-blue-600 border-t-transparent rounded-full"></div>
											</div>
										{/if}

										<!-- Add step button & delete workflow -->
										<div class="flex items-center justify-between pt-2">
											<button onclick={() => openAddStepModal(wf.id, wf.name)}
												class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-blue-600 dark:text-blue-400 bg-blue-50 dark:bg-blue-900/20 rounded-lg hover:bg-blue-100 dark:hover:bg-blue-900/40 transition cursor-pointer">
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
												Tambah Step
											</button>
											<button onclick={() => openDeleteModal(wf.id, wf.name)}
												class="inline-flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/40 transition cursor-pointer">
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
												Hapus Workflow
											</button>
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

<!-- Create Workflow Modal -->
<AnimatedPresence show={showCreateModal} type="scale" duration={200}>
	<div onclick={() => { if (!isCreating) showCreateModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') showCreateModal = false; }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Buat workflow baru" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center">
					<svg class="w-7 h-7 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-5">Buat Workflow Baru</h3>

				{#if createError}
					<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5 text-sm text-red-700 dark:text-red-300">{createError}</div>
				{/if}

				<div class="space-y-4">
					<div>
						<label for="wf-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Workflow <span class="text-red-500">*</span></label>
						<input id="wf-name" type="text" bind:value={createForm.name}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400"
							placeholder="Contoh: Workflow Cuti Tahunan" />
					</div>
					<div>
						<label for="wf-entity" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Module <span class="text-red-500">*</span></label>
						<select id="wf-entity" bind:value={createForm.entity_type}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
							{#each ENTITY_ORDER as type (type)}
								<option value={type}>{ENTITY_LABELS[type]}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="wf-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
						<textarea id="wf-desc" bind:value={createForm.description} rows="2"
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400 resize-none"
							placeholder="Deskripsi workflow (opsional)"></textarea>
					</div>
				</div>

				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={() => showCreateModal = false} disabled={isCreating}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
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
	</div>
</AnimatedPresence>

<!-- Edit Workflow Modal -->
<AnimatedPresence show={showEditModal} type="scale" duration={200}>
	<div onclick={() => { if (!isEditing) showEditModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') showEditModal = false; }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Edit workflow" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-indigo-50 dark:bg-indigo-900/30 flex items-center justify-center">
					<svg class="w-7 h-7 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-5">Edit Workflow</h3>

				{#if editError}
					<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5 text-sm text-red-700 dark:text-red-300">{editError}</div>
				{/if}

				<div class="space-y-4">
					<div>
						<label for="edit-name" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Nama Workflow <span class="text-red-500">*</span></label>
						<input id="edit-name" type="text" bind:value={editForm.name}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400" />
					</div>
					<div>
						<label for="edit-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
						<textarea id="edit-desc" bind:value={editForm.description} rows="2"
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder:text-gray-400 resize-none"
							placeholder="Deskripsi workflow (opsional)"></textarea>
					</div>
				</div>

				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={() => showEditModal = false} disabled={isEditing}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleEdit} disabled={isEditing}
						class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isEditing}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							Menyimpan...
						{:else}
							Simpan Perubahan
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>

<!-- Add Step Modal -->
<AnimatedPresence show={showAddStepModal} type="scale" duration={200}>
	<div onclick={() => { if (!isAddingStep) showAddStepModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') showAddStepModal = false; }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Tambah step workflow" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-indigo-50 dark:bg-indigo-900/30 flex items-center justify-center">
					<svg class="w-7 h-7 text-indigo-600 dark:text-indigo-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15M12 4.5l-3 3m3-3l3 3" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white text-center mb-1">Tambah Step Approval</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 text-center mb-5">Untuk: <strong class="text-gray-700 dark:text-gray-300">{addStepWorkflowName}</strong></p>

				{#if stepError}
					<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg px-4 py-2.5 text-sm text-red-700 dark:text-red-300">{stepError}</div>
				{/if}

				<div class="space-y-4">
					<div>
						<label for="step-order" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Urutan Step</label>
						<input id="step-order" type="number" bind:value={stepForm.step_order} min="1" max="10"
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white" />
					</div>
					<div>
						<label for="step-approver" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Approver <span class="text-red-500">*</span></label>
						<select id="step-approver" bind:value={stepForm.approver_type}
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white">
							{#each Object.entries(APPROVER_LABELS) as [value, label] (value)}
								<option value={value}>{label}</option>
							{/each}
						</select>
					</div>
					<div>
						<label for="step-escalation" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Eskalasi (jam)</label>
						<input id="step-escalation" type="number" bind:value={stepForm.escalation_hours} min="0" max="168"
							class="w-full px-3 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
							placeholder="0 = tidak ada eskalasi" />
					</div>
				</div>

				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={() => showAddStepModal = false} disabled={isAddingStep}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleAddStep} disabled={isAddingStep}
						class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isAddingStep}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							Menyimpan...
						{:else}
							Tambah Step
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>

<!-- Delete Confirmation Modal -->
<AnimatedPresence show={showDeleteModal} type="scale" duration={200}>
	<div onclick={() => { if (!isDeleting) showDeleteModal = false; }} onkeydown={(e) => { if (e.key === 'Escape') showDeleteModal = false; }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Hapus workflow" class="bg-white dark:bg-gray-800 rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-16 h-16 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/30 flex items-center justify-center">
					<svg class="w-8 h-8 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-2">Hapus Workflow</h3>
				<p class="text-sm text-gray-500 dark:text-gray-400 mb-1">Apakah Anda yakin ingin menghapus workflow:</p>
				<p class="text-sm font-bold text-gray-900 dark:text-white mb-4">"{deleteTarget?.name}"</p>
				<p class="text-xs text-red-600 dark:text-red-400 mb-5">Tindakan ini tidak dapat dibatalkan. Semua step workflow akan ikut terhapus.</p>

				<div class="flex items-center justify-center gap-3">
					<button onclick={() => { showDeleteModal = false; deleteTarget = null; }} disabled={isDeleting}
						class="px-4 py-2.5 border border-gray-200 dark:border-gray-600 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer">Batal</button>
					<button onclick={handleDeleteWorkflow} disabled={isDeleting}
						class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isDeleting}
							<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							Menghapus...
						{:else}
							Ya, Hapus
						{/if}
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
