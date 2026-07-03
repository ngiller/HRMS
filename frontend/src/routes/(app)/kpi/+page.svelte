<script lang="ts">
	import { onMount } from 'svelte';
	import { kpi, employees, positions, departments, ApiError } from '$lib/api.js';

	// ─── Tab State ───
	let activeTab = $state<'reviews' | 'templates'>('reviews');

	// ─── Reviews State ───
	let reviewData = $state<any[]>([]);
	let reviewTotal = $state(0);
	let reviewPage = $state(1);
	let reviewPerPage = $state(25);
	let reviewLoading = $state(true);
	let reviewStatusFilter = $state('');
	let reviewError = $state('');

	let showReviewForm = $state(false);
	let reviewForm = $state({ employee_id: '', kpi_template_id: '', period: 'Q1', year: 2026 });
	let reviewCreateLoading = $state(false);
	let employeeOptions = $state<any[]>([]);
	let templateOptions = $state<any[]>([]);

	let showReviewDetail = $state(false);
	let reviewDetail = $state<any>(null);
	let reviewDetailLoading = $state(false);

	// ─── Templates State ───
	let templateData = $state<any[]>([]);
	let templateTotal = $state(0);
	let templatePage = $state(1);
	let templatePerPage = $state(25);
	let templateLoading = $state(true);
	let templateYearFilter = $state(0);
	let templateError = $state('');

	let showTemplateForm = $state(false);
	let templateFormTitle = $state('');
	let templateForm = $state<any>({
		title: '', position_id: '', department_id: '', period_type: 'yearly', year: 2026, description: '', indicators: []
	});
	let templateFormError = $state('');
	let templateSaving = $state(false);
	let editingTemplateId = $state<string | null>(null);

	let showTemplateDetail = $state(false);
	let templateDetail = $state<any>(null);
	let templateDetailLoading = $state(false);

	let showDeleteModal = $state(false);
	let deleteId = $state<string | null>(null);
	let deleteLoading = $state(false);

	let positionOptions = $state<any[]>([]);
	let departmentAllOptions = $state<any[]>([]);

	// ─── Mount ───
	onMount(() => {
		loadReviews();
		loadReviewOptions();
		loadTemplates();
		loadTemplateOptions();
	});

	// ─── Shared Helpers ───
	async function loadReviewOptions() {
		try {
			const [empRes, tmplRes] = await Promise.all([
				employees.list(1, 100),
				kpi.listTemplates(1, 50),
			]);
			if (empRes?.success) employeeOptions = empRes.data;
			if (tmplRes?.success) templateOptions = tmplRes.data;
		} catch {}
	}

	async function loadTemplateOptions() {
		try {
			const [posRes, deptRes] = await Promise.all([
				positions.getAll(),
				departments.getAll(),
			]);
			if (posRes?.success) positionOptions = posRes.data;
			if (deptRes?.success) departmentAllOptions = deptRes.data;
		} catch {}
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	// ═══════════════════════════════════════════════
	//  REVIEWS
	// ═══════════════════════════════════════════════

	async function loadReviews() {
		reviewLoading = true;
		reviewError = '';
		try {
			const res: any = await kpi.listReviews(reviewPage, reviewPerPage, reviewStatusFilter);
			if (res?.success) {
				reviewData = res.data || [];
				reviewTotal = res.meta?.total || 0;
			}
		} catch (err) {
			reviewError = err instanceof ApiError ? err.message : 'Gagal memuat data';
		} finally {
			reviewLoading = false;
		}
	}

	function onReviewPageChange(p: number) { reviewPage = p; loadReviews(); }

	async function loadReviewDetail(id: string) {
		showReviewForm = false;
		if (showReviewDetail) { showReviewDetail = false; await new Promise(r => setTimeout(r, 50)); }
		reviewDetailLoading = true;
		showReviewDetail = true;
		try {
			const res: any = await kpi.getReview(id);
			if (res?.success) reviewDetail = res.data;
		} catch { reviewDetail = null; }
		finally { reviewDetailLoading = false; }
	}

	function closeReviewDetail() { showReviewDetail = false; reviewDetail = null; }

	function openReviewForm() {
		showReviewDetail = false;
		reviewForm = { employee_id: '', kpi_template_id: '', period: 'Q1', year: 2026 };
		reviewError = '';
		showReviewForm = true;
	}

	function cancelReviewForm() { showReviewForm = false; reviewError = ''; }

	async function handleCreateReview() {
		reviewCreateLoading = true;
		reviewError = '';
		try {
			const res: any = await kpi.createReview(reviewForm);
			if (res?.success) {
				showReviewForm = false;
				reviewForm = { employee_id: '', kpi_template_id: '', period: 'Q1', year: 2026 };
				loadReviews();
			}
		} catch (err) {
			reviewError = err instanceof ApiError ? err.message : 'Gagal membuat review KPI';
		} finally {
			reviewCreateLoading = false;
		}
	}

	const reviewTotalPages = $derived(Math.max(1, Math.ceil(reviewTotal / reviewPerPage)));

	const reviewStatusColors: Record<string, string> = {
		draft: 'bg-gray-100 text-gray-700 dark:bg-gray-900/30 dark:text-gray-400',
		self_review: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
		manager_review: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
		hr_review: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-400',
		completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
	};

	const reviewCategoryColors: Record<string, string> = {
		outstanding: 'bg-green-100 text-green-700',
		above: 'bg-blue-100 text-blue-700',
		meets: 'bg-gray-100 text-gray-700',
		needs_improvement: 'bg-yellow-100 text-yellow-700',
		underperform: 'bg-red-100 text-red-700',
	};

	function getReviewStatusBadge(status: string) {
		const labels: Record<string, string> = {
			draft: 'Draft', self_review: 'Self Review', manager_review: 'Manager Review',
			hr_review: 'HR Review', completed: 'Selesai',
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${reviewStatusColors[status] || 'bg-gray-100 text-gray-700'}">${labels[status] || status}</span>`;
	}

	function getReviewCategoryBadge(cat: string) {
		if (!cat) return '-';
		const labels: Record<string, string> = {
			outstanding: 'Outstanding', above: 'Above Expectation', meets: 'Meets Expectation',
			needs_improvement: 'Needs Improvement', underperform: 'Underperform',
		};
		return `<span class="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${reviewCategoryColors[cat] || 'bg-gray-100 text-gray-700'}">${labels[cat] || cat}</span>`;
	}

	// ═══════════════════════════════════════════════
	//  TEMPLATES
	// ═══════════════════════════════════════════════

	async function loadTemplates() {
		templateLoading = true;
		templateError = '';
		try {
			const res: any = await kpi.listTemplates(templatePage, templatePerPage, templateYearFilter);
			if (res?.success) {
				templateData = res.data || [];
				templateTotal = res.meta?.total || 0;
			}
		} catch (err) {
			templateError = err instanceof ApiError ? err.message : 'Gagal memuat data';
		} finally {
			templateLoading = false;
		}
	}

	function onTemplatePageChange(p: number) { templatePage = p; loadTemplates(); }

	async function loadTemplateDetail(id: string) {
		showTemplateForm = false;
		if (showTemplateDetail) { showTemplateDetail = false; await new Promise(r => setTimeout(r, 50)); }
		templateDetailLoading = true;
		showTemplateDetail = true;
		try {
			const res: any = await kpi.getTemplate(id);
			if (res?.success) templateDetail = res.data;
		} catch { templateDetail = null; }
		finally { templateDetailLoading = false; }
	}

	function closeTemplateDetail() { showTemplateDetail = false; templateDetail = null; }

	function openCreateTemplate() {
		showTemplateDetail = false;
		editingTemplateId = null;
		templateForm = {
			title: '', position_id: '', department_id: '', period_type: 'yearly',
			year: 2026, description: '', indicators: [
				{ name: '', description: '', target: 0, weight: 0, measurement_unit: '', sort_order: 0 }
			]
		};
		templateFormTitle = 'Buat Template KPI Baru';
		templateFormError = '';
		showTemplateForm = true;
	}

	function openEditTemplate(id: string) {
		showTemplateDetail = false;
		showTemplateForm = true;
		editingTemplateId = id;
		templateFormTitle = 'Edit Template KPI';
		templateFormError = '';
		templateSaving = true;

		kpi.getTemplate(id).then((res: any) => {
			if (res?.success) {
				const t = res.data;
				templateForm = {
					title: t.title || '',
					position_id: t.position_id || '',
					department_id: t.department_id || '',
					period_type: t.period_type || 'yearly',
					year: t.year || 2026,
					description: t.description || '',
					is_active: t.is_active,
					indicators: (t.indicators || []).map((ind: any, i: number) => ({
						name: ind.name || '',
						description: ind.description || '',
						target: ind.target || 0,
						weight: ind.weight || 0,
						measurement_unit: ind.measurement_unit || '',
						sort_order: i
					})),
				};
				if (templateForm.indicators.length === 0) {
					templateForm.indicators = [{ name: '', description: '', target: 0, weight: 0, measurement_unit: '', sort_order: 0 }];
				}
			}
		}).catch(() => {
			templateFormError = 'Gagal memuat data template';
		}).finally(() => {
			templateSaving = false;
		});
	}

	function cancelTemplateForm() {
		showTemplateForm = false;
		templateFormError = '';
		editingTemplateId = null;
	}

	function addIndicator() {
		templateForm.indicators = [...templateForm.indicators, {
			name: '', description: '', target: 0, weight: 0, measurement_unit: '', sort_order: templateForm.indicators.length
		}];
	}

	function removeIndicator(index: number) {
		if (templateForm.indicators.length <= 1) return;
		templateForm.indicators = templateForm.indicators.filter((_: any, i: number) => i !== index)
			.map((ind: any, i: number) => ({ ...ind, sort_order: i }));
	}

	function calcTotalWeight(): number {
		return templateForm.indicators.reduce((sum: number, ind: any) => sum + (parseFloat(ind.weight) || 0), 0);
	}

	async function handleSaveTemplate() {
		if (!templateForm.title.trim()) { templateFormError = 'Judul template harus diisi'; return; }
		if (!templateForm.period_type) { templateFormError = 'Tipe periode harus diisi'; return; }
		if (!templateForm.year || templateForm.year < 2024) { templateFormError = 'Tahun harus >= 2024'; return; }
		if (templateForm.indicators.length === 0) { templateFormError = 'Minimal 1 indikator harus ditambahkan'; return; }

		const totalW = calcTotalWeight();
		if (totalW < 99.5 || totalW > 100.5) { templateFormError = `Total bobot indikator harus 100% (saat ini ${totalW.toFixed(1)}%)`; return; }

		for (let i = 0; i < templateForm.indicators.length; i++) {
			const ind = templateForm.indicators[i];
			if (!ind.name.trim()) { templateFormError = `Nama indikator #${i + 1} harus diisi`; return; }
			if (parseFloat(ind.target) <= 0) { templateFormError = `Target indikator #${i + 1} harus lebih dari 0`; return; }
			if (parseFloat(ind.weight) <= 0) { templateFormError = `Bobot indikator #${i + 1} harus lebih dari 0`; return; }
		}

		templateSaving = true;
		templateFormError = '';
		try {
			const payload = {
				title: templateForm.title.trim(),
				position_id: templateForm.position_id || null,
				department_id: templateForm.department_id || null,
				period_type: templateForm.period_type,
				year: templateForm.year,
				description: templateForm.description.trim(),
				is_active: templateForm.is_active,
				indicators: templateForm.indicators.map((ind: any) => ({
					name: ind.name.trim(),
					description: ind.description.trim(),
					target: parseFloat(ind.target),
					weight: parseFloat(ind.weight),
					measurement_unit: ind.measurement_unit,
					sort_order: ind.sort_order,
				})),
			};

			if (editingTemplateId) {
				await kpi.updateTemplate(editingTemplateId, payload);
			} else {
				await kpi.createTemplate(payload);
			}
			cancelTemplateForm();
			loadTemplates();
			loadReviewOptions();
		} catch (err) {
			templateFormError = err instanceof ApiError ? err.message : 'Gagal menyimpan template';
		} finally {
			templateSaving = false;
		}
	}

	function openDeleteModal(id: string) {
		deleteId = id;
		deleteLoading = false;
		showDeleteModal = true;
	}

	function cancelDelete() { showDeleteModal = false; deleteId = null; }

	async function handleDelete() {
		if (!deleteId) return;
		deleteLoading = true;
		try {
			await kpi.deleteTemplate(deleteId);
			showDeleteModal = false;
			deleteId = null;
			if (showTemplateDetail) { showTemplateDetail = false; templateDetail = null; }
			loadTemplates();
			loadReviewOptions();
		} catch (err) {
			templateError = err instanceof ApiError ? err.message : 'Gagal menghapus template';
			showDeleteModal = false;
		} finally {
			deleteLoading = false;
		}
	}

	const templateTotalPages = $derived(Math.max(1, Math.ceil(templateTotal / templatePerPage)));
	const indicatorTotalWeight = $derived(calcTotalWeight());

	const templatePeriodLabels: Record<string, string> = {
		yearly: 'Tahunan', quarterly: 'Kuartal',
	};
</script>

<div class="w-full">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">KPI & Performance</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Kelola template dan review kinerja karyawan</p>
	</div>

	<!-- ─── Tabs ─── -->
	<div class="flex gap-1 mb-6 border-b border-gray-200 dark:border-gray-800">
		<button onclick={() => { activeTab = 'reviews'; showReviewForm = false; showReviewDetail = false; }}
			class="px-5 py-3 text-sm font-medium transition border-b-2 -mb-px cursor-pointer
				{activeTab === 'reviews' ? 'border-[#1A56DB] text-[#1A56DB]' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}">
			Review KPI
		</button>
		<button onclick={() => { activeTab = 'templates'; showTemplateForm = false; showTemplateDetail = false; }}
			class="px-5 py-3 text-sm font-medium transition border-b-2 -mb-px cursor-pointer
				{activeTab === 'templates' ? 'border-[#1A56DB] text-[#1A56DB]' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'}">
			Template KPI
		</button>
	</div>

	{#if activeTab === 'reviews'}
		<!-- ════════════════════════════════════════ -->
		<!--  REVIEWS TAB -->
		<!-- ════════════════════════════════════════ -->
		<div class="flex items-center justify-between mb-4">
			<div class="text-sm text-gray-500 dark:text-gray-400">Total {reviewTotal} review</div>
			{#if !showReviewForm}
				<button onclick={openReviewForm} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					Buat Review Baru
				</button>
			{/if}
		</div>

		{#if showReviewForm}
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm mb-6">
				<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Buat Review KPI Baru</h2>
					<button onclick={cancelReviewForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>
				</div>
				<form onsubmit={(e) => { e.preventDefault(); handleCreateReview(); }} class="px-6 py-5 space-y-4">
					{#if reviewError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-2.5">{reviewError}</div>{/if}
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="review-employee" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Karyawan <span class="text-red-500">*</span></label>
							<select id="review-employee" bind:value={reviewForm.employee_id} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20">
								<option value="">Pilih Karyawan</option>
								{#each employeeOptions as emp}<option value={emp.id}>{emp.full_name}</option>{/each}
							</select>
						</div>
						<div>
							<label for="review-template" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Template KPI <span class="text-red-500">*</span></label>
							<select id="review-template" bind:value={reviewForm.kpi_template_id} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20">
								<option value="">Pilih Template</option>
								{#each templateOptions as tmpl}<option value={tmpl.id}>{tmpl.title} ({tmpl.year})</option>{/each}
							</select>
						</div>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="review-period" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Periode <span class="text-red-500">*</span></label>
							<select id="review-period" bind:value={reviewForm.period} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
								<option value="Q1">Q1 (Jan-Mar)</option><option value="Q2">Q2 (Apr-Jun)</option>
								<option value="Q3">Q3 (Jul-Sep)</option><option value="Q4">Q4 (Okt-Des)</option>
								<option value="yearly">Tahunan</option>
							</select>
						</div>
						<div>
							<label for="review-year" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tahun <span class="text-red-500">*</span></label>
							<input id="review-year" type="number" bind:value={reviewForm.year} min="2024" max="2030" class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" />
						</div>
					</div>
					<div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
						<button type="button" onclick={cancelReviewForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
						<button type="submit" disabled={reviewCreateLoading} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
							{#if reviewCreateLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
							Buat Review
						</button>
					</div>
				</form>
			</div>
		{:else if showReviewDetail}
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
				<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Detail Review KPI</h2>
					<button onclick={closeReviewDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>
				</div>
				<div class="px-6 py-5">
					{#if reviewDetailLoading}
						<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-48"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-64"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-40"></div></div>
					{:else if reviewDetail}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
							<div>
								<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Review</h3>
								<div class="space-y-3">
									<div><span class="text-xs text-gray-400">Karyawan</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{reviewDetail.employee_name}</p></div>
									<div><span class="text-xs text-gray-400">Template</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{reviewDetail.template_title}</p></div>
									<div><span class="text-xs text-gray-400">Periode</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{reviewDetail.period} {reviewDetail.year}</p></div>
									<div><span class="text-xs text-gray-400">Status</span><p>{@html getReviewStatusBadge(reviewDetail.status)}</p></div>
								</div>
							</div>
							<div>
								<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Hasil Penilaian</h3>
								<div class="space-y-3">
									<div><span class="text-xs text-gray-400">Skor Akhir</span><p class="text-2xl font-bold text-[#1A56DB]">{reviewDetail.final_score ?? '-'}</p></div>
									<div><span class="text-xs text-gray-400">Kategori</span><p>{@html getReviewCategoryBadge(reviewDetail.final_category)}</p></div>
									{#if reviewDetail.salary_increase}<div><span class="text-xs text-gray-400">Kenaikan Gaji</span><p class="text-sm font-semibold text-green-600">{reviewDetail.salary_increase}%</p></div>{/if}
									{#if reviewDetail.bonus_amount}<div><span class="text-xs text-gray-400">Bonus</span><p class="text-sm font-semibold text-green-600">{reviewDetail.bonus_amount.toLocaleString('id-ID')}</p></div>{/if}
								</div>
							</div>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<!-- Filter -->
			<div class="flex gap-3 mb-4">
				<select bind:value={reviewStatusFilter} onchange={() => { reviewPage = 1; loadReviews(); }}
					class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
					<option value="">Semua Status</option>
					<option value="draft">Draft</option><option value="self_review">Self Review</option>
					<option value="manager_review">Manager Review</option><option value="hr_review">HR Review</option>
					<option value="completed">Selesai</option>
				</select>
			</div>

			{#if reviewError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{reviewError}</div>{/if}

			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
							<tr>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Karyawan</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Template</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Periode</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Skor</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Kategori</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Tanggal</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Aksi</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#if reviewLoading}
								{#each [1,2,3,4,5] as _}
									<tr class="animate-pulse">
										<td class="px-4 py-3"><div class="flex items-center gap-3"><div class="w-8 h-8 bg-gray-100 dark:bg-gray-800 rounded-full"></div><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-36"></div></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-28"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-20"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-12"></div></td>
										<td class="px-4 py-3"><div class="h-5 bg-gray-100 dark:bg-gray-800 rounded-full w-20"></div></td>
										<td class="px-4 py-3"><div class="h-5 bg-gray-100 dark:bg-gray-800 rounded-full w-24"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-16"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-12"></div></td>
									</tr>
								{/each}
							{:else if reviewData.length === 0}
								<tr><td colspan="8" class="px-4 py-8 text-center text-sm text-gray-400">Belum ada review KPI</td></tr>
							{:else}
								{#each reviewData as item}
									<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition {reviewDetail?.id === item.id && showReviewDetail ? 'bg-blue-50/50 dark:bg-blue-900/10' : ''}">
										<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{item.employee_name}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.template_title}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.period} {item.year}</td>
										<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{item.final_score ?? '-'}</td>
										<td class="px-4 py-3">{@html getReviewCategoryBadge(item.final_category)}</td>
										<td class="px-4 py-3">{@html getReviewStatusBadge(item.status)}</td>
										<td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{new Date(item.created_at).toLocaleDateString('id-ID')}</td>
										<td class="px-4 py-3"><button onclick={() => loadReviewDetail(item.id)} class="text-xs text-[#1A56DB] hover:underline font-medium cursor-pointer">Detail</button></td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
				<div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
					<span class="text-xs text-gray-400">Total {reviewTotal} data</span>
					<div class="flex gap-1">
						<button onclick={() => onReviewPageChange(reviewPage - 1)} disabled={reviewPage <= 1} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Prev</button>
						<span class="px-3 py-1 text-sm text-gray-500">{(reviewPage - 1) * reviewPerPage + 1} - {Math.min(reviewPage * reviewPerPage, reviewTotal)}</span>
						<button onclick={() => onReviewPageChange(reviewPage + 1)} disabled={reviewPage >= reviewTotalPages} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Next</button>
					</div>
				</div>
			</div>
		{/if}

	{:else}
		<!-- ════════════════════════════════════════ -->
		<!--  TEMPLATES TAB -->
		<!-- ════════════════════════════════════════ -->
		<div class="flex items-center justify-between mb-4">
			<div class="flex gap-3">
				<select bind:value={templateYearFilter} onchange={() => { templatePage = 1; loadTemplates(); }}
					class="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
					<option value={0}>Semua Tahun</option>
					{#each [2024,2025,2026,2027,2028] as y}<option value={y}>{y}</option>{/each}
				</select>
			</div>
			{#if !showTemplateForm}
				<button onclick={openCreateTemplate} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-2">
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
					Buat Template
				</button>
			{/if}
		</div>

		{#if showTemplateForm}
			<!-- ─── Template Form (Create/Edit) ─── -->
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm mb-6">
				<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">{templateFormTitle}</h2>
					<button onclick={cancelTemplateForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>
				</div>
				<form onsubmit={(e) => { e.preventDefault(); handleSaveTemplate(); }} class="px-6 py-5 space-y-4">
					{#if templateFormError}
						<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-2.5">{templateFormError}</div>
					{/if}

					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="tmpl-title" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Judul Template <span class="text-red-500">*</span></label>
							<input id="tmpl-title" type="text" bind:value={templateForm.title} required class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" placeholder="Contoh: KPI Marketing 2026" />
						</div>
						<div>
							<label for="tmpl-period" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tipe Periode <span class="text-red-500">*</span></label>
							<select id="tmpl-period" bind:value={templateForm.period_type} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
								<option value="yearly">Tahunan</option>
								<option value="quarterly">Kuartal</option>
							</select>
						</div>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="tmpl-pos" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Posisi Jabatan</label>
							<select id="tmpl-pos" bind:value={templateForm.position_id} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
								<option value="">Semua Posisi</option>
								{#each positionOptions as pos}<option value={pos.id}>{pos.name}</option>{/each}
							</select>
						</div>
						<div>
							<label for="tmpl-dept" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Departemen</label>
							<select id="tmpl-dept" bind:value={templateForm.department_id} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none">
								<option value="">Semua Departemen</option>
								{#each departmentAllOptions as dept}<option value={dept.id}>{dept.name}</option>{/each}
							</select>
						</div>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4">
						<div>
							<label for="tmpl-year" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Tahun <span class="text-red-500">*</span></label>
							<input id="tmpl-year" type="number" bind:value={templateForm.year} min="2024" max="2030" class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20" />
						</div>
						<div>
							<label for="tmpl-desc" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Deskripsi</label>
							<input id="tmpl-desc" type="text" bind:value={templateForm.description} class="w-full px-3 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" placeholder="Deskripsi template (opsional)" />
						</div>
					</div>

					<!-- ─── Indicators ─── -->
					<div class="border-t border-gray-200 dark:border-gray-700 pt-4">
						<div class="flex items-center justify-between mb-3">
							<h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Indikator Penilaian</h3>
							<div class="flex items-center gap-3">
								<span class="text-xs text-gray-400">Total Bobot: <span class="font-medium {Math.abs(indicatorTotalWeight - 100) < 0.5 ? 'text-green-600' : 'text-red-500'}">{indicatorTotalWeight.toFixed(1)}%</span></span>
								<button type="button" onclick={addIndicator} class="px-3 py-1.5 bg-[#1A56DB] text-white rounded-lg text-xs font-medium hover:bg-[#1e40af] transition cursor-pointer flex items-center gap-1">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
									Tambah
								</button>
							</div>
						</div>
						<div class="space-y-3">
							{#each templateForm.indicators as indicator, i}
								<div class="p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-gray-800/30">
									<div class="flex items-center justify-between mb-2">
										<span class="text-xs font-semibold text-gray-500 uppercase tracking-wider">Indikator #{i + 1}</span>
										<button type="button" onclick={() => removeIndicator(i)} disabled={templateForm.indicators.length <= 1} class="text-xs text-red-500 hover:text-red-700 disabled:opacity-30 disabled:cursor-not-allowed cursor-pointer">Hapus</button>
									</div>
									<div class="grid grid-cols-2 md:grid-cols-5 gap-2">
										<div class="col-span-2 md:col-span-2">
											<label for="indicator-name-{i}" class="block text-xs text-gray-500 mb-0.5">Nama Indikator <span class="text-red-500">*</span></label>
											<input id="indicator-name-{i}" type="text" bind:value={indicator.name} class="w-full px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" placeholder="Contoh: Target Penjualan" />
										</div>
										<div>
											<label for="indicator-target-{i}" class="block text-xs text-gray-500 mb-0.5">Target <span class="text-red-500">*</span></label>
											<input id="indicator-target-{i}" type="number" bind:value={indicator.target} step="0.01" min="0.01" class="w-full px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" />
										</div>
										<div>
											<label for="indicator-weight-{i}" class="block text-xs text-gray-500 mb-0.5">Bobot % <span class="text-red-500">*</span></label>
											<input id="indicator-weight-{i}" type="number" bind:value={indicator.weight} step="0.1" min="0.1" max="100" class="w-full px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" />
										</div>
										<div>
											<label for="indicator-unit-{i}" class="block text-xs text-gray-500 mb-0.5">Satuan</label>
											<input id="indicator-unit-{i}" type="text" bind:value={indicator.measurement_unit} class="w-full px-2 py-1.5 border border-gray-300 dark:border-gray-600 rounded text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none" placeholder="%, Rp, unit" />
										</div>
									</div>
								</div>
							{/each}
						</div>
					</div>

					<div class="flex items-center justify-end gap-3 pt-2 border-t border-gray-100 dark:border-gray-800">
						<button type="button" onclick={cancelTemplateForm} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
						<button type="submit" disabled={templateSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
							{#if templateSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
							{editingTemplateId ? 'Simpan Perubahan' : 'Buat Template'}
						</button>
					</div>
				</form>
			</div>
		{:else if showTemplateDetail}
			<!-- ─── Template Detail ─── -->
			<div class="bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
				<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
					<h2 class="text-lg font-semibold text-gray-900 dark:text-gray-100">Detail Template KPI</h2>
					<button onclick={closeTemplateDetail} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>
				</div>
				<div class="px-6 py-5">
					{#if templateDetailLoading}
						<div class="animate-pulse space-y-3 p-4"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-48"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-64"></div><div class="h-4 bg-gray-50 dark:bg-gray-800 rounded w-40"></div></div>
					{:else if templateDetail}
						<div class="grid grid-cols-1 md:grid-cols-2 gap-6">
							<div>
								<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Informasi Template</h3>
								<div class="space-y-3">
									<div><span class="text-xs text-gray-400">Judul</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{templateDetail.title}</p></div>
									<div><span class="text-xs text-gray-400">Status</span><p><span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {templateDetail.is_active ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-600'}">{templateDetail.is_active ? 'Aktif' : 'Nonaktif'}</span></p></div>
									<div><span class="text-xs text-gray-400">Periode</span><p class="text-sm font-medium text-gray-900 dark:text-gray-100">{templatePeriodLabels[templateDetail.period_type] || templateDetail.period_type} - {templateDetail.year}</p></div>
									<div><span class="text-xs text-gray-400">Posisi</span><p class="text-sm text-gray-700 dark:text-gray-300">{templateDetail.position_name || 'Semua Posisi'}</p></div>
									<div><span class="text-xs text-gray-400">Departemen</span><p class="text-sm text-gray-700 dark:text-gray-300">{templateDetail.dept_name || 'Semua Departemen'}</p></div>
									{#if templateDetail.description}<div><span class="text-xs text-gray-400">Deskripsi</span><p class="text-sm text-gray-700 dark:text-gray-300">{templateDetail.description}</p></div>{/if}
								</div>
							</div>
						</div>

						<!-- Indicators Table -->
						<div class="mt-6">
							<h3 class="text-xs font-semibold text-gray-400 dark:text-gray-500 uppercase tracking-wider mb-3">Indikator Penilaian</h3>
							<div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
								<table class="w-full text-sm">
									<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
										<tr>
											<th class="px-3 py-2 font-medium text-gray-500 dark:text-gray-400">#</th>
											<th class="px-3 py-2 font-medium text-gray-500 dark:text-gray-400">Nama Indikator</th>
											<th class="px-3 py-2 font-medium text-gray-500 dark:text-gray-400">Target</th>
											<th class="px-3 py-2 font-medium text-gray-500 dark:text-gray-400">Satuan</th>
											<th class="px-3 py-2 font-medium text-gray-500 dark:text-gray-400">Bobot</th>
										</tr>
									</thead>
									<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
										{#each templateDetail.indicators || [] as ind, i}
											<tr>
												<td class="px-3 py-2 text-gray-500">{i + 1}</td>
												<td class="px-3 py-2 font-medium text-gray-900 dark:text-gray-100">{ind.name}</td>
												<td class="px-3 py-2 text-gray-700 dark:text-gray-300">{ind.target}</td>
												<td class="px-3 py-2 text-gray-500">{ind.measurement_unit || '-'}</td>
												<td class="px-3 py-2 font-medium text-[#1A56DB]">{ind.weight}%</td>
											</tr>
										{/each}
									</tbody>
									<tfoot class="bg-gray-50 dark:bg-gray-800/50">
										<tr>
											<td colspan="4" class="px-3 py-2 text-xs font-semibold text-gray-500 text-right">Total Bobot</td>
											<td class="px-3 py-2 font-bold text-[#1A56DB]">
												{(templateDetail.indicators || []).reduce((s: number, ind: any) => s + (parseFloat(ind.weight) || 0), 0).toFixed(1)}%
											</td>
										</tr>
									</tfoot>
								</table>
							</div>
						</div>

						<div class="mt-6 flex gap-2">
							<button onclick={() => openEditTemplate(templateDetail.id)} class="px-4 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer">Edit Template</button>
							<button onclick={() => { closeTemplateDetail(); openDeleteModal(templateDetail.id); }} class="px-4 py-2 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition cursor-pointer">Hapus</button>
						</div>
					{/if}
				</div>
			</div>
		{:else}
			<!-- ─── Template Table ─── -->
			{#if templateError}<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3 mb-4">{templateError}</div>{/if}

			<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
				<div class="overflow-x-auto">
					<table class="w-full text-sm">
						<thead class="bg-gray-50 dark:bg-gray-800/50 text-left">
							<tr>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Judul</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Periode</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Posisi</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Departemen</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Status</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Dibuat</th>
								<th class="px-4 py-3 font-medium text-gray-500 dark:text-gray-400">Aksi</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-gray-100 dark:divide-gray-800">
							{#if templateLoading}
								{#each [1,2,3] as _}
									<tr class="animate-pulse">
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-40"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-24"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-28"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-28"></div></td>
										<td class="px-4 py-3"><div class="h-5 bg-gray-100 dark:bg-gray-800 rounded-full w-16"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-20"></div></td>
										<td class="px-4 py-3"><div class="h-4 bg-gray-100 dark:bg-gray-800 rounded w-16"></div></td>
									</tr>
								{/each}
							{:else if templateData.length === 0}
								<tr><td colspan="7" class="px-4 py-8 text-center text-sm text-gray-400">Belum ada template KPI. Klik "Buat Template" untuk membuat baru.</td></tr>
							{:else}
								{#each templateData as item}
									<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/30 transition">
										<td class="px-4 py-3 font-medium text-gray-900 dark:text-gray-100">{item.title}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{templatePeriodLabels[item.period_type] || item.period_type} {item.year}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.position_name || '-'}</td>
										<td class="px-4 py-3 text-gray-600 dark:text-gray-400">{item.dept_name || '-'}</td>
										<td class="px-4 py-3"><span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium {item.is_active ? 'bg-green-50 text-green-700' : 'bg-gray-100 text-gray-600'}">{item.is_active ? 'Aktif' : 'Nonaktif'}</span></td>
										<td class="px-4 py-3 text-gray-500 dark:text-gray-400 text-xs">{formatDate(item.created_at)}</td>
										<td class="px-4 py-3 flex gap-1">
											<button onclick={() => loadTemplateDetail(item.id)} class="text-xs text-[#1A56DB] hover:underline font-medium cursor-pointer">Detail</button>
											<button onclick={() => { showTemplateDetail = false; openEditTemplate(item.id); }} class="text-xs text-amber-600 hover:underline font-medium cursor-pointer">Edit</button>
										</td>
									</tr>
								{/each}
							{/if}
						</tbody>
					</table>
				</div>
				<div class="flex items-center justify-between px-4 py-3 border-t border-gray-100 dark:border-gray-800">
					<span class="text-xs text-gray-400">Total {templateTotal} data</span>
					<div class="flex gap-1">
						<button onclick={() => onTemplatePageChange(templatePage - 1)} disabled={templatePage <= 1} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Prev</button>
						<span class="px-3 py-1 text-sm text-gray-500">{(templatePage - 1) * templatePerPage + 1} - {Math.min(templatePage * templatePerPage, templateTotal)}</span>
						<button onclick={() => onTemplatePageChange(templatePage + 1)} disabled={templatePage >= templateTotalPages} class="px-3 py-1 text-sm rounded border border-gray-200 dark:border-gray-700 disabled:opacity-30 hover:bg-gray-50 dark:hover:bg-gray-800 transition cursor-pointer">Next</button>
					</div>
				</div>
			</div>
		{/if}
	{/if}
</div>

<!-- ─── Delete Confirm Modal ─── -->
{#if showDeleteModal}
	<!-- svelte-ignore a11y_interactive_supports_focus -->
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelDelete} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelDelete(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Hapus template KPI" class="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-md">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 dark:bg-red-900/20 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">Hapus Template KPI</h3>
				<p class="text-sm text-gray-500">Apakah Anda yakin ingin menghapus template ini? Tindakan ini tidak dapat dibatalkan.</p>
				<div class="flex items-center justify-center gap-3 mt-6">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 dark:border-gray-700 rounded-lg text-sm font-medium text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={deleteLoading} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if deleteLoading}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Ya, Hapus
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
