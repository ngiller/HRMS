<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { payroll as payrollApi } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions';

	type PayslipData = {
		id: string;
		payroll_period_id: string;
		employee_id: string;
		employee_name: string;
		employee_id_code: string;
		department_name: string;
		position_name: string;
		employment_status: string;
		base_salary: number;
		daily_wage: number;
		total_days_worked: number;
		allowances: { name: string; amount: number }[];
		overtime_pay: number;
		overtime_hours: number;
		thr_amount: number;
		bonus_amount: number;
		gross_salary: number;
		deductions: { name: string; amount: number }[];
		pph21_amount: number;
		bpjs_kesehatan: number;
		bpjs_jht: number;
		bpjs_jp: number;
		loan_deduction: number;
		other_deductions: number;
		total_deductions: number;
		net_salary: number;
		company_cost: { name: string; amount: number }[];
		status: string;
		notes: string;
		period_name: string;
		period_status: string;
	};

	let periodId = $derived($page.params.periodId as string);
	let employeeId = $derived($page.params.employeeId as string);
	let payslip = $state<PayslipData | null>(null);
	let isLoading = $state(true);
	let errorMessage = $state('');
	let hasPayrollRead = $state(false);
	let isDownloading = $state(false);
	let payslipRef = $state<HTMLDivElement | null>(null);

	onMount(async () => {
		hasPayrollRead = hasPermission('payroll', 'read');
		try {
			interface PayslipResponse { success: boolean; data?: PayslipData; }
			let res: PayslipResponse;
			if (hasPayrollRead) {
				res = await payrollApi.getPayslip(periodId, employeeId) as PayslipResponse;
			} else {
				res = await payrollApi.getMyPayslip(periodId) as PayslipResponse;
			}
			payslip = res.data || null;
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memuat slip gaji';
		} finally {
			isLoading = false;
		}
	});

	async function downloadPDF() {
		if (!payslipRef || !payslip) return;
		isDownloading = true;
		try {
			const { toPng } = await import('html-to-image');
			const { jsPDF } = await import('jspdf');

			const dataUrl = await toPng(payslipRef, {
				pixelRatio: 2,
				quality: 1,
				cacheBust: true,
			});

			const pdf = new jsPDF('portrait', 'mm', 'a4');
			const pdfWidth = pdf.internal.pageSize.getWidth();

			const margin = 10;
			const imgWidth = pdfWidth - margin * 2;
			const imgHeight = (payslipRef.offsetHeight / payslipRef.offsetWidth) * imgWidth;

			pdf.addImage(dataUrl, 'PNG', margin, margin, imgWidth, imgHeight);
			pdf.save(`Slip Gaji - ${payslip.employee_name} - ${payslip.period_name}.pdf`);
		} catch (err) {
			console.error('Failed to generate PDF:', err);
		} finally {
			isDownloading = false;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function statusBadge(status: string): string {
		const map: Record<string, string> = {
			draft: 'bg-gray-100 text-gray-600',
			calculated: 'bg-blue-100 text-blue-700',
			approved: 'bg-emerald-100 text-emerald-700',
			paid: 'bg-purple-100 text-purple-700',
		};
		return map[status] || 'bg-gray-100 text-gray-600';
	}

	function statusLabel(status: string): string {
		const map: Record<string, string> = {
			draft: 'Draft', calculated: 'Dihitung', approved: 'Disetujui', paid: 'Dibayarkan',
		};
		return map[status] || status;
	}

	function getEmploymentIcon(status: string): string {
		const icons: Record<string, string> = {
			tetap: '\u25CF',
			kontrak: '\u25CF',
			magang: '\u25CB',
			harian: '\u25B3',
		};
		return icons[status] || '\u25CF';
	}
	function statusLabelEmployment(status: string): string {
		const map: Record<string, string> = { tetap: 'Tetap', kontrak: 'Kontrak', magang: 'Magang', harian: 'Harian' };
		return map[status] || status || '-';
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

	{#if isLoading}
		<div class="bg-white border border-gray-200 rounded-xl p-6 animate-pulse space-y-4">
			<div class="h-6 bg-gray-100 rounded w-48"></div>
			<div class="h-4 bg-gray-50 rounded w-36"></div>
			<div class="grid grid-cols-2 gap-4 mt-4">
				<div class="h-40 bg-gray-50 rounded-xl"></div>
				<div class="h-40 bg-gray-50 rounded-xl"></div>
			</div>
		</div>
	{:else if errorMessage}
		<div class="bg-white border border-gray-200 rounded-xl py-16 text-center">
			<p class="text-sm text-gray-500">{errorMessage}</p>
		</div>
	{:else if payslip}
		<!-- ═══════ PAYSLIP CONTENT FOR PDF ═══════ -->
		<div bind:this={payslipRef} class="print-area">
			<!-- ═══════ HEADER ═══════ -->
			<div class="border-b border-gray-200 px-6 py-5 bg-gradient-to-r from-gray-50 to-white">
				<div class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<div class="w-12 h-12 rounded-xl bg-[#1A56DB] flex items-center justify-center text-white shadow-sm">
							<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18" />
							</svg>
						</div>
						<div>
							<h1 class="text-xl font-bold text-gray-900">Slip Gaji</h1>
							<p class="text-sm text-gray-500 mt-0.5">{payslip.period_name}</p>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<span class="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium ring-1 ring-inset {statusBadge(payslip.period_status)}">
							<span class="w-1.5 h-1.5 rounded-full bg-current"></span>
							{statusLabel(payslip.period_status)}
						</span>					<button onclick={() => window.print()}
						class="inline-flex items-center gap-1.5 px-3 py-1.5 border border-gray-200 rounded-lg text-xs font-medium text-gray-600 hover:bg-gray-50 transition cursor-pointer">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M6.72 13.829c-.24.03-.48.062-.72.096m.72-.096a42.415 42.415 0 0 1 10.56 0m-10.56 0L6.34 18m10.94-4.171c.24.03.48.062.72.096m-.72-.096L17.66 18m0 0 .229 2.523a1.125 1.125 0 0 1-1.12 1.227H7.231c-.662 0-1.18-.568-1.12-1.227L6.34 18m11.318 0h1.091A2.25 2.25 0 0 0 21 15.75V9.456c0-1.081-.768-2.015-1.837-2.175a48.055 48.055 0 0 0-1.913-.247M6.34 18H5.25A2.25 2.25 0 0 1 3 15.75V9.456c0-1.081.768-2.015 1.837-2.175a48.041 48.041 0 0 1 1.913-.247m10.5 0a48.536 48.536 0 0 0-10.5 0m10.5 0V3.375c0-.621-.504-1.125-1.125-1.125h-8.25c-.621 0-1.125.504-1.125 1.125v3.659" />
							</svg>
							Cetak
						</button>
						<button onclick={downloadPDF} disabled={isDownloading}
							class="inline-flex items-center gap-1.5 px-3 py-1.5 border border-[#1A56DB] rounded-lg text-xs font-medium text-[#1A56DB] hover:bg-blue-50 transition disabled:opacity-50 cursor-pointer">
							{#if isDownloading}
								<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							{:else}
								<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
								</svg>
							{/if}
							Download PDF
						</button>
					</div>
				</div>
			</div>

			<div class="p-6 space-y-6">
				<!-- ═══════ EMPLOYEE INFO CARD ═══════ -->
				<div class="bg-gradient-to-r from-gray-50 to-white rounded-xl border border-gray-200 px-5 py-4">
					<div class="flex items-center justify-between">
						<div class="flex items-center gap-3">
							<div class="w-10 h-10 rounded-full bg-[#1A56DB]/10 flex items-center justify-center text-[#1A56DB]">
								<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" />
								</svg>
							</div>
							<div>
								<p class="text-sm font-semibold text-gray-900">{payslip.employee_name}</p>
								<p class="text-xs text-gray-500">{payslip.department_name} • {payslip.position_name}</p>
							</div>
						</div>
						<div class="text-right">
							<p class="text-xs text-gray-400">NIP: <span class="font-medium text-gray-600">{payslip.employee_id_code}</span></p>
							<p class="text-xs text-gray-400 capitalize mt-0.5">
								<span class="inline-flex items-center gap-1">
									{getEmploymentIcon(payslip.employment_status)}
									<span>{statusLabelEmployment(payslip.employment_status)}</span>
								</span>
							</p>
						</div>
					</div>
				</div>

				<!-- ═══════ TWO-COLUMN: INCOME vs DEDUCTIONS ═══════ -->
				<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
					<!-- LEFT: INCOME -->
					<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
						<div class="bg-emerald-50/80 border-b border-emerald-100 px-4 py-3">
							<div class="flex items-center gap-2">
								<div class="w-7 h-7 rounded-lg bg-emerald-500 flex items-center justify-center text-white">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
									</svg>
								</div>
								<h3 class="text-sm font-bold text-emerald-800 uppercase tracking-wider">Pendapatan</h3>
							</div>
						</div>
						<div class="px-4 py-3 space-y-0 divide-y divide-gray-50">
							<!-- Base Salary -->
							<div class="flex justify-between items-center py-2.5">
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-600">Gaji Pokok</span>
								</div>
								<span class="text-sm font-semibold text-gray-900 tabular-nums">{payslip.base_salary > 0 ? formatCurrency(payslip.base_salary) : '-'}</span>
							</div>

							<!-- Daily Wage -->
							{#if payslip.daily_wage > 0}
								{@const dailyRate = payslip.daily_wage / payslip.total_days_worked}
								<div class="flex justify-between items-start py-2.5">
									<div class="flex flex-col gap-1">
										<span class="text-sm text-gray-600">Upah Harian</span>
										<div class="flex items-center gap-1.5 text-xs text-gray-400">
											<span class="inline-flex items-center px-1.5 py-0.5 rounded bg-gray-100 font-medium text-gray-500">{formatCurrency(dailyRate)}</span>
											<span class="text-gray-300">×</span>
											<span class="font-medium text-gray-500">{payslip.total_days_worked} hari</span>
											<span class="text-gray-300">=</span>
											<span class="font-semibold text-gray-600">{formatCurrency(payslip.daily_wage)}</span>
										</div>
									</div>
									<span class="text-sm font-semibold text-gray-900 tabular-nums">{formatCurrency(payslip.daily_wage)}</span>
								</div>
							{/if}

							<!-- Allowances -->
							{#each payslip.allowances || [] as allowance}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">{allowance.name}</span>
									<span class="text-sm font-medium text-emerald-600 tabular-nums">+ {formatCurrency(allowance.amount)}</span>
								</div>
							{/each}

							<!-- Overtime -->
							{#if payslip.overtime_pay > 0}
								{@const overtimeRate = payslip.overtime_pay / payslip.overtime_hours}
								<div class="flex justify-between items-center py-2.5">
									<div class="flex flex-col">
										<span class="text-sm text-gray-600">Lembur</span>
										<span class="text-xs text-gray-400 mt-0.5">{formatCurrency(overtimeRate)}/jam × {payslip.overtime_hours} jam</span>
									</div>
									<span class="text-sm font-medium text-amber-600 tabular-nums">+ {formatCurrency(payslip.overtime_pay)}</span>
								</div>
							{/if}

							<!-- THR -->
							{#if payslip.thr_amount > 0}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">THR</span>
									<span class="text-sm font-medium text-emerald-600 tabular-nums">+ {formatCurrency(payslip.thr_amount)}</span>
								</div>
							{/if}

							<!-- Bonus -->
							{#if payslip.bonus_amount > 0}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">Bonus</span>
									<span class="text-sm font-medium text-emerald-600 tabular-nums">+ {formatCurrency(payslip.bonus_amount)}</span>
								</div>
							{/if}

							<!-- Gross -->
							<div class="flex justify-between items-center py-3 mt-0.5 bg-emerald-50/50 -mx-4 px-4 border-t-2 border-emerald-200">
								<span class="text-sm font-bold text-gray-900">Gross Salary</span>
								<span class="text-sm font-bold text-gray-900 tabular-nums">{formatCurrency(payslip.gross_salary)}</span>
							</div>
						</div>
					</div>

					<!-- RIGHT: DEDUCTIONS -->
					<div class="bg-white rounded-xl border border-gray-200 overflow-hidden">
						<div class="bg-red-50/80 border-b border-red-100 px-4 py-3">
							<div class="flex items-center gap-2">
								<div class="w-7 h-7 rounded-lg bg-red-500 flex items-center justify-center text-white">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
									</svg>
								</div>
								<h3 class="text-sm font-bold text-red-800 uppercase tracking-wider">Potongan</h3>
							</div>
						</div>
						<div class="px-4 py-3 space-y-0 divide-y divide-gray-50">
							<!-- BPJS Kesehatan -->
							<div class="flex justify-between items-center py-2.5">
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-600">BPJS Kesehatan</span>
									<span class="text-xs text-gray-400">(1%)</span>
								</div>
								<span class="text-sm font-medium text-red-600 tabular-nums">{payslip.bpjs_kesehatan > 0 ? formatCurrency(payslip.bpjs_kesehatan) : '-'}</span>
							</div>

							<!-- BPJS JHT -->
							<div class="flex justify-between items-center py-2.5">
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-600">BPJS JHT</span>
									<span class="text-xs text-gray-400">(2%)</span>
								</div>
								<span class="text-sm font-medium text-red-600 tabular-nums">{payslip.bpjs_jht > 0 ? formatCurrency(payslip.bpjs_jht) : '-'}</span>
							</div>

							<!-- BPJS JP -->
							<div class="flex justify-between items-center py-2.5">
								<div class="flex items-center gap-2">
									<span class="text-sm text-gray-600">BPJS JP</span>
									<span class="text-xs text-gray-400">(1%)</span>
								</div>
								<span class="text-sm font-medium text-red-600 tabular-nums">{payslip.bpjs_jp > 0 ? formatCurrency(payslip.bpjs_jp) : '-'}</span>
							</div>

							<!-- PPh 21 -->
							{#if payslip.pph21_amount > 0}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">PPh 21</span>
									<span class="text-sm font-medium text-red-600 tabular-nums">{formatCurrency(payslip.pph21_amount)}</span>
								</div>
							{/if}

							<!-- Loan -->
							{#if payslip.loan_deduction > 0}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">Pinjaman</span>
									<span class="text-sm font-medium text-red-600 tabular-nums">{formatCurrency(payslip.loan_deduction)}</span>
								</div>
							{/if}

							<!-- Other Deductions from array (selain yang sudah ditampilkan khusus + Lain-lain karena dari other_deductions) -->
							{#each (payslip.deductions || []).filter(d => 
								d.name !== 'BPJS Kesehatan' && 
								d.name !== 'BPJS JHT' && 
								d.name !== 'BPJS JP' && 
								d.name !== 'PPh 21' && 
								d.name !== 'Pinjaman' &&
								d.name !== 'Lain-lain'
							) as deduction}
								<div class="flex justify-between items-center py-2.5">
									<span class="text-sm text-gray-600">{deduction.name}</span>
									<span class="text-sm font-medium text-red-600 tabular-nums">{formatCurrency(deduction.amount)}</span>
								</div>
							{/each}

							<!-- Other Deductions -->
							{#if payslip.other_deductions > 0}
								<div class="flex justify-between items-start py-2.5">
									<div class="flex flex-col">
										<span class="text-sm text-gray-600">Lain-lain</span>
										<span class="text-xs text-gray-400 mt-0.5">Potongan lainnya</span>
									</div>
									<span class="text-sm font-medium text-red-600 tabular-nums">{formatCurrency(payslip.other_deductions)}</span>
								</div>
							{/if}

							<!-- Total -->
							<div class="flex justify-between items-center py-3 mt-0.5 bg-red-50/50 -mx-4 px-4 border-t-2 border-red-200">
								<span class="text-sm font-bold text-gray-900">Total Potongan</span>
								<span class="text-sm font-bold text-red-600 tabular-nums">{formatCurrency(payslip.total_deductions)}</span>
							</div>
						</div>
					</div>
				</div>

				<!-- ═══════ TAKE HOME PAY ═══════ -->
				<div class="relative overflow-hidden bg-gradient-to-br from-blue-600 via-indigo-600 to-purple-700 rounded-2xl p-6 md:p-8 flex flex-col md:flex-row md:items-center justify-between shadow-xl shadow-indigo-900/20 text-white gap-6 group hover:shadow-2xl hover:shadow-indigo-900/30 transition-all duration-300 border border-white/10">
					<!-- decorative shapes -->
					<div class="absolute top-0 right-0 -mr-12 -mt-12 w-40 h-40 rounded-full bg-white/10 blur-3xl group-hover:bg-white/20 transition-all duration-500"></div>
					<div class="absolute bottom-0 left-0 -ml-8 -mb-8 w-32 h-32 rounded-full bg-white/10 blur-2xl"></div>
					
					<div class="relative z-10 flex flex-col gap-2 w-full">
						<div class="flex items-center gap-2 text-blue-100 mb-1">
							<svg class="w-5 h-5 opacity-80" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v12m-3-2.818.879.659c1.171.879 3.07.879 4.242 0 1.172-.879 1.172-2.303 0-3.182C13.536 12.219 12.768 12 12 12c-.725 0-1.45-.22-2.003-.659-1.106-.879-1.106-2.303 0-3.182s2.9-.879 4.006 0l.415.33M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
							</svg>
							<span class="text-sm font-semibold tracking-widest uppercase opacity-90">Take Home Pay</span>
						</div>
						<div class="flex items-baseline">
							<span class="text-4xl md:text-5xl font-extrabold tabular-nums tracking-tight drop-shadow-md">{formatCurrency(payslip.net_salary)}</span>
						</div>
						
						<!-- Calculation Breakdown -->
						<div class="flex flex-col sm:flex-row sm:items-center gap-2 mt-3 text-xs sm:text-sm text-blue-50 font-medium bg-white/10 w-fit px-3.5 py-2 rounded-xl backdrop-blur-md border border-white/10 shadow-inner">
							<div class="flex items-center gap-1.5 opacity-90">
								<span>Kotor:</span>
								<span>{formatCurrency(payslip.gross_salary)}</span>
							</div>
							<span class="hidden sm:inline text-white/40 px-1">•</span>
							<div class="flex items-center gap-1.5 opacity-90">
								<span>Potongan:</span>
								<span class="text-red-200">-{formatCurrency(payslip.total_deductions)}</span>
							</div>
						</div>
					</div>
					
					<div class="relative z-10 hidden md:flex items-center justify-center w-16 h-16 rounded-2xl bg-white/10 backdrop-blur-md border border-white/20 shadow-inner shrink-0 group-hover:scale-110 group-hover:rotate-3 transition-transform duration-300">
						<span class="text-3xl drop-shadow-sm">💰</span>
					</div>
				</div>


				<!-- ═══════ NOTES ═══════ -->
				{#if payslip.notes}
					<div class="bg-gray-50 rounded-xl border border-gray-200 px-5 py-3.5">
						<div class="flex items-start gap-2">
							<svg class="w-4 h-4 text-gray-400 mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L6.832 19.82a4.5 4.5 0 0 1-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 0 1 1.13-1.897L16.863 4.487Zm0 0L19.5 7.125" />
							</svg>
							<p class="text-sm text-gray-600">{payslip.notes}</p>
						</div>
					</div>
				{/if}
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

		/* Sembunyiin semua elemen, tampilkan hanya area payslip */
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

		/* Sembunyiin tombol */
		button, .cursor-pointer, .no-print {
			display: none !important;
		}

		/* Hilangin shadow & rounding untuk cetak */
		.shadow-sm, .shadow-2xl, .shadow {
			box-shadow: none !important;
		}
		.rounded-xl, .rounded-lg, .rounded-2xl, .rounded-full {
			border-radius: 0 !important;
		}

		/* Border tetep tipis */
		.border {
			border-width: 1px !important;
		}

		/* Pastiin teks hitam solid biar jelas */
		.text-gray-900, .text-gray-800, .text-gray-700, .text-gray-600, .text-gray-500, .text-gray-400 {
			color: #000 !important;
		}
		.text-emerald-600, .text-emerald-700, .text-emerald-800 {
			color: #059669 !important;
		}
		.text-red-600, .text-red-700, .text-red-800 {
			color: #dc2626 !important;
		}
		.text-blue-600, .text-blue-700, .text-blue-500, .text-blue-400 {
			color: #2563eb !important;
		}
		.text-amber-600 {
			color: #d97706 !important;
		}

		/* Gradient jadi solid biar gak error print */
		.bg-gradient-to-r {
			background: #f9fafb !important;
		}

		/* Background cards */
		.bg-emerald-50\/80 {
			background: #ecfdf5 !important;
		}
		.bg-red-50\/80 {
			background: #fef2f2 !important;
		}
		.bg-emerald-50\/50 {
			background: #ecfdf5 !important;
		}
		.bg-red-50\/50 {
			background: #fef2f2 !important;
		}
		.bg-blue-50, .bg-gradient-to-r.from-blue-50 {
			background: #eff6ff !important;
		}
		.bg-gray-50 {
			background: #f9fafb !important;
		}

		/* Pastiin layout dua kolom */
		.grid.grid-cols-1.lg\:grid-cols-2 {
			display: grid !important;
			grid-template-columns: 1fr 1fr !important;
			gap: 12px !important;
		}

		/* Ukuran font pas buat cetak */
		.text-sm {
			font-size: 11px !important;
		}
		.text-xs {
			font-size: 9px !important;
		}
		.text-xl {
			font-size: 18px !important;
		}
		.text-3xl {
			font-size: 22px !important;
		}
		.text-base {
			font-size: 13px !important;
		}

		/* Padding/gap lebih hemat */
		.space-y-6 {
			margin-top: 0 !important;
		}
		.p-6 {
			padding: 16px !important;
		}
		.px-6 {
			padding-left: 16px !important;
			padding-right: 16px !important;
		}
		.py-5 {
			padding-top: 12px !important;
			padding-bottom: 12px !important;
		}
		.px-5 {
			padding-left: 16px !important;
			padding-right: 16px !important;
		}
		.py-4 {
			padding-top: 10px !important;
			padding-bottom: 10px !important;
		}
		.px-4 {
			padding-left: 12px !important;
			padding-right: 12px !important;
		}
		.py-3 {
			padding-top: 8px !important;
			padding-bottom: 8px !important;
		}
		.py-2\.5 {
			padding-top: 6px !important;
			padding-bottom: 6px !important;
		}
		.gap-6 {
			gap: 12px !important;
		}
		.gap-4 {
			gap: 10px !important;
		}
		.gap-3 {
			gap: 8px !important;
		}
		.gap-2 {
			gap: 6px !important;
		}
		.gap-1\.5 {
			gap: 4px !important;
		}

		/* SVG ikon disembunyiin */
		svg {
			display: none !important;
		}

		/* W-10, h-10 dll gak perlu di print */
		.w-12, .h-12, .w-10, .h-10, .w-7, .h-7 {
			display: none !important;
		}

		/* Fix border untuk print */
		.border-gray-200 {
			border-color: #e5e7eb !important;
		}
		.border-emerald-200 {
			border-color: #a7f3d0 !important;
		}
		.border-red-200 {
			border-color: #fecaca !important;
		}
		.border-blue-200 {
			border-color: #bfdbfe !important;
		}
		.border-emerald-100 {
			border-color: #d1fae5 !important;
		}
		.border-red-100 {
			border-color: #fee2e2 !important;
		}
	}
</style>
