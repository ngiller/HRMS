<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { mutations, company as companyApi } from '$lib/api.js';

	type MutationDetail = {
		id: string;
		employee_id: string;
		employee_name: string;
		employee_id_code?: string;
		mutation_type: string;
		old_department_name: string;
		new_department_name: string;
		old_position_name: string;
		new_position_name: string;
		old_employment_status: string;
		new_employment_status: string;
		old_base_salary: number | null;
		new_base_salary: number | null;
		old_position_grade_name: string;
		new_position_grade_name: string;
		reason: string;
		notes: string;
		effective_date: string;
		status: string;
		approved_by_name: string;
		approved_at: string;
		created_at: string;
	};

	type CompanyData = {
		id: string;
		name: string;
		address: string;
		city: string;
		npwp: string;
	};

	let mutationId = $derived($page.params.id);
	let mutation = $state<MutationDetail | null>(null);
	let company = $state<CompanyData | null>(null);
	let isLoading = $state(true);
	let errorMessage = $state('');
	let isDownloading = $state(false);
	let skRef = $state<HTMLDivElement | null>(null);

	onMount(async () => {
		if (!mutationId) {
			errorMessage = 'ID mutasi tidak ditemukan';
			isLoading = false;
			return;
		}
		try {
			const [mutRes, compRes] = await Promise.all([
				mutations.get(mutationId as string),
				companyApi.getSettings(),
			]);
			if (mutRes.success) mutation = mutRes.data as MutationDetail;
			if (compRes.success) company = compRes.data as CompanyData;
		} catch (e: unknown) {
			errorMessage = (e as { message?: string }).message || 'Gagal memuat data';
		} finally {
			isLoading = false;
		}
	});

	const typeLabels: Record<string, string> = {
		promotion: 'Promosi Jabatan',
		demotion: 'Demosi Jabatan',
		transfer: 'Mutasi Departemen',
		position_change: 'Perubahan Jabatan',
		status_change: 'Perubahan Status Kepegawaian',
		salary_change: 'Penyesuaian Gaji',
	};

	const statusTypeLabels: Record<string, string> = {
		tetap: 'Karyawan Tetap',
		kontrak: 'Karyawan Kontrak',
		percobaan: 'Karyawan Percobaan',
		harian: 'Karyawan Harian',
	};

	function formatCurrency(val: number | null): string {
		if (val === null) return '-';
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', {
			day: 'numeric', month: 'long', year: 'numeric',
		});
	}

	function getTypeLabel(type: string): string {
		return typeLabels[type] || type;
	}

	async function downloadPDF() {
		if (!skRef) return;
		isDownloading = true;
		try {
			const { toPng } = await import('html-to-image');
			const { jsPDF } = await import('jspdf');

			// Temporarily expand for capture
			skRef.style.width = '800px';
			skRef.style.padding = '40px';

			const dataUrl = await toPng(skRef, {
				pixelRatio: 2,
				quality: 1,
				cacheBust: true,
			});

			// Reset
			skRef.style.width = '';
			skRef.style.padding = '';

			const pdf = new jsPDF('portrait', 'mm', 'a4');
			const pdfWidth = pdf.internal.pageSize.getWidth();
			const margin = 15;
			const imgWidth = pdfWidth - margin * 2;
			const imgHeight = (skRef.offsetHeight / skRef.offsetWidth) * imgWidth;

			pdf.addImage(dataUrl, 'PNG', margin, margin, imgWidth, imgHeight);
			pdf.save(`SK Mutasi - ${mutation?.employee_name || '-'}.pdf`);
		} catch (err) {
			console.error('Failed to generate PDF:', err);
		} finally {
			isDownloading = false;
		}
	}
</script>

<div class="w-full max-w-full mx-auto px-4 xl:px-8">
	<button onclick={() => history.back()}
		class="inline-flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-900 transition mb-5 cursor-pointer group">
		<svg class="w-4 h-4 group-hover:-translate-x-0.5 transition-transform" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
		</svg>
		<span>Kembali</span>
	</button>

	<!-- Action Buttons -->
	<div class="flex items-center justify-end gap-3 mb-4">
		<button onclick={() => window.print()}
			class="inline-flex items-center gap-1.5 px-4 py-2 border border-gray-200 rounded-lg text-sm font-medium text-gray-600 hover:bg-gray-50 transition cursor-pointer">
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M6.72 13.829c-.24.03-.48.062-.72.096m.72-.096a42.415 42.415 0 0 1 10.56 0m-10.56 0L6.34 18m10.94-4.171c.24.03.48.062.72.096m-.72-.096L17.66 18m0 0 .229 2.523a1.125 1.125 0 0 1-1.12 1.227H7.231c-.662 0-1.18-.568-1.12-1.227L6.34 18m11.318 0h1.091A2.25 2.25 0 0 0 21 15.75V9.456c0-1.081-.768-2.015-1.837-2.175a48.055 48.055 0 0 0-1.913-.247M6.34 18H5.25A2.25 2.25 0 0 1 3 15.75V9.456c0-1.081.768-2.015 1.837-2.175a48.041 48.041 0 0 1 1.913-.247m10.5 0a48.536 48.536 0 0 0-10.5 0m10.5 0V3.375c0-.621-.504-1.125-1.125-1.125h-8.25c-.621 0-1.125.504-1.125 1.125v3.659" />
			</svg>
			Cetak
		</button>
		<button onclick={downloadPDF} disabled={isDownloading}
			class="inline-flex items-center gap-1.5 px-4 py-2 border border-[#1A56DB] rounded-lg text-sm font-medium text-[#1A56DB] hover:bg-blue-50 transition disabled:opacity-50 cursor-pointer">
			{#if isDownloading}
				<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
			{:else}
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
				</svg>
			{/if}
			Download PDF
		</button>
	</div>

	{#if isLoading}
		<div class="bg-white border border-gray-200 rounded-xl p-8 animate-pulse space-y-4">
			<div class="h-8 bg-gray-100 rounded w-64"></div>
			<div class="h-4 bg-gray-50 rounded w-48"></div>
			<div class="h-40 bg-gray-50 rounded-xl"></div>
		</div>
	{:else if errorMessage}
		<div class="bg-white border border-gray-200 rounded-xl py-16 text-center">
			<p class="text-sm text-gray-500">{errorMessage}</p>
		</div>
	{:else if mutation}
		<div bind:this={skRef} class="bg-white border border-gray-200 rounded-xl overflow-hidden print-area">
			<!-- Kop Surat -->
			<div class="border-b-2 border-gray-900 px-10 py-8 text-center">
				<h1 class="text-xl font-bold text-gray-900 uppercase tracking-wide">{company?.name || 'PERUSAHAAN'}</h1>
				<p class="text-sm text-gray-500 mt-1 max-w-xl mx-auto">{company?.address || ''}</p>
				<div class="mt-4 pt-4 border-t border-gray-300">
					<h2 class="text-lg font-bold text-gray-900 uppercase tracking-wider">SURAT KEPUTUSAN MUTASI</h2>
					<p class="text-sm text-gray-500 mt-0.5">Nomor: HRM/SK-MUT/{mutation!.id.slice(0,8).toUpperCase()}/{new Date().getFullYear()}</p>
				</div>
			</div>

			<!-- Body -->
			<div class="px-10 py-8 space-y-6 text-sm text-gray-800 leading-relaxed">
				<p class="text-justify">
					Yang bertanda tangan di bawah ini, <span class="font-semibold">{company?.name || 'Perusahaan'}</span>, menerangkan bahwa:
				</p>

				<div class="bg-gray-50 rounded-xl border border-gray-200 px-6 py-4">
					<table class="w-full text-sm">
					<tbody>
						<tr>
							<td class="py-1.5 text-gray-500 w-40 align-top">Nama Karyawan</td>
							<td class="py-1.5 font-semibold text-gray-900">{mutation.employee_name}</td>
						</tr>
						<tr>
							<td class="py-1.5 text-gray-500 align-top">NIP</td>
							<td class="py-1.5 text-gray-900">{mutation.employee_id_code || mutation.employee_id || '-'}</td>
						</tr>
						<tr>
							<td class="py-1.5 text-gray-500 align-top">Jenis Mutasi</td>
							<td class="py-1.5"><span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-50 text-blue-700">{getTypeLabel(mutation.mutation_type)}</span></td>
						</tr>
						<tr>
							<td class="py-1.5 text-gray-500 align-top">Tanggal Berlaku</td>
							<td class="py-1.5 text-gray-900">{formatDate(mutation.effective_date)}</td>
						</tr>
					</tbody>
				</table>
				</div>

				<p class="text-justify">Berdasarkan hasil evaluasi dan pertimbangan manajemen, dengan ini diputuskan untuk melakukan perubahan data karyawan sebagai berikut:</p>

				<!-- Perubahan Detail -->
				<div class="space-y-3">
					{#if mutation.old_department_name || mutation.new_department_name}
						<div class="flex items-start gap-3 p-3 border border-gray-200 rounded-lg">
							<div class="w-8 h-8 rounded-lg bg-blue-50 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21" /></svg>
							</div>
							<div>
								<p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Departemen</p>
								<p class="text-sm mt-0.5">
									<span class="text-gray-500 line-through">{mutation.old_department_name || '-'}</span>
									<span class="mx-2 text-gray-300">→</span>
									<span class="font-semibold">{mutation.new_department_name || '-'}</span>
								</p>
							</div>
						</div>
					{/if}

					{#if mutation.old_position_name || mutation.new_position_name}
						<div class="flex items-start gap-3 p-3 border border-gray-200 rounded-lg">
							<div class="w-8 h-8 rounded-lg bg-emerald-50 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-emerald-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 14.15v4.25c0 1.094-.787 2.036-1.872 2.18-2.087.277-4.216.42-6.378.42s-4.291-.143-6.378-.42c-1.085-.144-1.872-1.086-1.872-2.18v-4.25m16.5 0a2.18 2.18 0 0 0 .75-1.661V8.706c0-1.081-.768-2.015-1.837-2.175a48.114 48.114 0 0 0-3.413-.387m4.5 8.006c-.194.165-.42.295-.673.38A23.978 23.978 0 0 1 12 15.75c-2.648 0-5.195-.429-7.577-1.22a2.016 2.016 0 0 1-.673-.38m0 0A2.18 2.18 0 0 1 3 12.489V8.706c0-1.081.768-2.015 1.837-2.175a48.111 48.111 0 0 1 3.413-.387m7.5 0V5.25A2.25 2.25 0 0 0 13.5 3h-3a2.25 2.25 0 0 0-2.25 2.25v.894m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>
							</div>
							<div>
								<p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Jabatan</p>
								<p class="text-sm mt-0.5">
									<span class="text-gray-500 line-through">{mutation.old_position_name || '-'}</span>
									<span class="mx-2 text-gray-300">→</span>
									<span class="font-semibold">{mutation.new_position_name || '-'}</span>
								</p>
							</div>
						</div>
					{/if}

					{#if mutation.old_employment_status || mutation.new_employment_status}
						<div class="flex items-start gap-3 p-3 border border-gray-200 rounded-lg">
							<div class="w-8 h-8 rounded-lg bg-purple-50 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-purple-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75m-3-7.036A11.959 11.959 0 0 1 3.598 6 11.99 11.99 0 0 0 3 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285Z" /></svg>
							</div>
							<div>
								<p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Status Kepegawaian</p>
								<p class="text-sm mt-0.5">
									<span class="text-gray-500 line-through">{statusTypeLabels[mutation.old_employment_status] || mutation.old_employment_status || '-'}</span>
									<span class="mx-2 text-gray-300">→</span>
									<span class="font-semibold">{statusTypeLabels[mutation.new_employment_status] || mutation.new_employment_status || '-'}</span>
								</p>
							</div>
						</div>
					{/if}

					{#if mutation.old_base_salary !== null || mutation.new_base_salary !== null}
						<div class="flex items-start gap-3 p-3 border border-gray-200 rounded-lg">
							<div class="w-8 h-8 rounded-lg bg-amber-50 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
							</div>
							<div>
								<p class="text-xs font-semibold text-gray-400 uppercase tracking-wider">Gaji Pokok</p>
								<p class="text-sm mt-0.5">
									<span class="text-gray-500 line-through">{mutation.old_base_salary !== null ? formatCurrency(mutation.old_base_salary) : '-'}</span>
									<span class="mx-2 text-gray-300">→</span>
									<span class="font-semibold text-emerald-600">{mutation.new_base_salary !== null ? formatCurrency(mutation.new_base_salary) : '-'}</span>
								</p>
							</div>
						</div>
					{/if}
				</div>

				<!-- Alasan -->
				<div class="border-t border-gray-200 pt-4">
					<p class="text-xs font-semibold text-gray-400 uppercase tracking-wider mb-2">Alasan Mutasi</p>
					<p class="text-justify text-gray-700 italic">"{mutation.reason || '-'}"</p>
				</div>

				<!-- Penutup -->
				<div class="border-t border-gray-200 pt-4 text-justify">
					<p>Demikian surat keputusan ini dibuat untuk diketahui dan dilaksanakan sebagaimana mestinya.</p>
					<div class="mt-8 text-right">
						<p class="text-sm text-gray-500">{company?.city || 'Jakarta'}, {formatDate(mutation.approved_at || mutation.created_at || mutation.effective_date)}</p>
						<div class="mt-8">
							<p class="text-sm font-semibold text-gray-900">{company?.name || 'Perusahaan'}</p>
							<div class="mt-12">
								<p class="text-sm font-semibold text-gray-900 underline underline-offset-4 decoration-2 decoration-gray-300">( ______________________________ )</p>
								<p class="text-xs text-gray-500 mt-1">HR Manager / Direktur</p>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>

<style>
	@media print {
		@page {
			size: A4;
			margin: 15mm 15mm 15mm 15mm;
		}
		:global(body *) {
			visibility: hidden !important;
		}
		.print-area, .print-area * {
			visibility: visible !important;
		}
		.print-area {
			position: absolute !important;
			left: 0 !important;
			top: 0 !important;
			width: 100% !important;
			max-width: 100% !important;
			padding: 0 !important;
			margin: 0 !important;
		}
		button, .cursor-pointer {
			display: none !important;
		}
		.rounded-xl, .rounded-lg {
			border-radius: 0 !important;
		}
		.border {
			border-width: 1px !important;
		}
		.text-gray-900, .text-gray-800, .text-gray-700, .text-gray-600 {
			color: #000 !important;
		}
		svg { display: none !important; }
	}
</style>
