<script lang="ts">
/* eslint-disable svelte/no-navigation-without-resolve */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { payroll } from '$lib/api';
	import { getUser, hasPermission } from '$lib/permissions';
	import MobileCard from '$lib/components/MobileCard.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	type PayslipItem = {
		id: string;
		payroll_period_id: string;
		employee_id: string;
		employee_name: string;
		position_name: string;
		period_name: string;
		net_salary: number;
		status: string;
	};
	let payslips = $state<PayslipItem[]>([]);
	let loading = $state(true);
	let error = $state('');
	let downloadingId = $state<string | null>(null);

	async function downloadPdf(periodId: string, employeeId: string, employeeName: string, periodName: string) {
		downloadingId = `${periodId}_${employeeId}`;
		try {
			const res = await payroll.getPayslip(periodId, employeeId);
			if (!res.success || !res.data) {
				console.error('Failed to fetch payslip data');
				return;
			}
			const data = res.data;

			const { toPng } = await import('html-to-image');
			const { jsPDF } = await import('jspdf');

			const div = document.createElement('div');
			div.style.cssText = 'position:fixed;left:-9999px;top:0;width:800px;background:#ffffff;padding:40px;font-family:Arial,sans-serif;color:#000000;z-index:99999;box-sizing:border-box;';
			
			const incHtml = (data.allowances || []).map((a: any) => `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>${a.name}</span><span>${formatCurrency(a.amount)}</span></div>`).join('');
			let dedHtml = '';
			if (data.bpjs_kesehatan > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>BPJS Kesehatan</span><span>${formatCurrency(data.bpjs_kesehatan)}</span></div>`;
			if (data.bpjs_jht > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>BPJS JHT</span><span>${formatCurrency(data.bpjs_jht)}</span></div>`;
			if (data.bpjs_jp > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>BPJS JP</span><span>${formatCurrency(data.bpjs_jp)}</span></div>`;
			if (data.pph21_amount > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>PPh 21</span><span>${formatCurrency(data.pph21_amount)}</span></div>`;
			if (data.loan_deduction > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>Pinjaman</span><span>${formatCurrency(data.loan_deduction)}</span></div>`;
			const otherDeds = (data.deductions || []).filter((d: any) => !['BPJS Kesehatan','BPJS JHT','BPJS JP','PPh 21','Pinjaman','Lain-lain'].includes(d.name));
			dedHtml += otherDeds.map((d: any) => `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>${d.name}</span><span>${formatCurrency(d.amount)}</span></div>`).join('');
			if (data.other_deductions > 0) dedHtml += `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>Lain-lain</span><span>${formatCurrency(data.other_deductions)}</span></div>`;

			div.innerHTML = `
				<div style="border-bottom:2px solid #000;padding-bottom:15px;margin-bottom:25px;display:flex;justify-content:space-between;align-items:flex-end;">
					<div>
						<h1 style="font-size:24px;font-weight:bold;margin:0;letter-spacing:1px;text-transform:uppercase;">SLIP GAJI</h1>
						<p style="font-size:14px;color:#444;margin:5px 0 0 0;">Periode: ${periodName}</p>
					</div>
					<div style="text-align:right;">
						<p style="font-size:16px;font-weight:bold;margin:0;">${data.employee_name}</p>
						<p style="font-size:14px;color:#444;margin:3px 0 0 0;">NIP: ${data.employee_id_code} &bull; ${data.position_name}</p>
						<p style="font-size:14px;color:#444;margin:3px 0 0 0;">Departemen: ${data.department_name}</p>
					</div>
				</div>

				<div style="display:flex;gap:30px;margin-bottom:30px;">
					<!-- PENDAPATAN -->
					<div style="flex:1;">
						<h3 style="font-size:14px;font-weight:bold;border-bottom:1px solid #ccc;padding-bottom:5px;margin-bottom:10px;text-transform:uppercase;">Pendapatan</h3>
						<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;">
							<span>Gaji Pokok</span>
							<span>${data.base_salary > 0 ? formatCurrency(data.base_salary) : '-'}</span>
						</div>
						${data.daily_wage > 0 ? `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>Upah Harian</span><span>${formatCurrency(data.daily_wage)}</span></div>` : ''}
						${incHtml}
						${data.overtime_pay > 0 ? `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>Lembur</span><span>${formatCurrency(data.overtime_pay)}</span></div>` : ''}
						${data.thr_amount > 0 ? `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>THR</span><span>${formatCurrency(data.thr_amount)}</span></div>` : ''}
						${data.bonus_amount > 0 ? `<div style="display:flex;justify-content:space-between;padding:4px 0;font-size:14px;"><span>Bonus</span><span>${formatCurrency(data.bonus_amount)}</span></div>` : ''}
						<div style="display:flex;justify-content:space-between;padding:8px 0;margin-top:10px;border-top:1px dashed #000;font-size:14px;font-weight:bold;">
							<span>Total Pendapatan</span>
							<span>${formatCurrency(data.gross_salary)}</span>
						</div>
					</div>

					<!-- POTONGAN -->
					<div style="flex:1;">
						<h3 style="font-size:14px;font-weight:bold;border-bottom:1px solid #ccc;padding-bottom:5px;margin-bottom:10px;text-transform:uppercase;">Potongan</h3>
						${dedHtml}
						<div style="display:flex;justify-content:space-between;padding:8px 0;margin-top:10px;border-top:1px dashed #000;font-size:14px;font-weight:bold;">
							<span>Total Potongan</span>
							<span>${formatCurrency(data.total_deductions)}</span>
						</div>
					</div>
				</div>

				<div style="border:2px solid #000;padding:20px;text-align:center;background:#fafafa;">
					<p style="font-size:14px;margin:0 0 5px 0;text-transform:uppercase;font-weight:bold;color:#444;">Take Home Pay</p>
					<p style="font-size:28px;font-weight:bold;margin:0;">${formatCurrency(data.net_salary)}</p>
				</div>
			`;
			document.body.appendChild(div);

			await new Promise(r => setTimeout(r, 100));

			const dataUrl = await toPng(div, { pixelRatio: 2, quality: 1, cacheBust: true });

			const pdf = new jsPDF('portrait', 'mm', 'a4');
			const pdfWidth = pdf.internal.pageSize.getWidth();
			const margin = 15;
			const imgWidth = pdfWidth - margin * 2;
			const imgHeight = (div.offsetHeight / div.offsetWidth) * imgWidth;

			pdf.addImage(dataUrl, 'PNG', margin, margin, imgWidth, imgHeight);
			pdf.save(`Slip_Gaji_${employeeName.replace(/\\s+/g, '_')}_${periodName.replace(/\\s+/g, '_')}.pdf`);
		} catch (err) {
			console.error('Failed to generate PDF:', err);
		} finally {
			const existing = document.querySelector('div[style*="left:-9999px"]');
			if (existing && existing.parentNode) existing.parentNode.removeChild(existing);
			downloadingId = null;
		}
	}

	function renderIncomeItems(data: any): string {
		let items = '';
		items += row('Gaji Pokok', data.base_salary > 0 ? formatCurrency(data.base_salary) : '-', '#111827');
		if (data.daily_wage > 0) items += row('Upah Harian', formatCurrency(data.daily_wage), '#111827');
		for (const a of data.allowances || []) items += row(a.name, '+ ' + formatCurrency(a.amount), '#059669');
		if (data.overtime_pay > 0) items += row('Lembur', '+ ' + formatCurrency(data.overtime_pay), '#d97706');
		if (data.thr_amount > 0) items += row('THR', '+ ' + formatCurrency(data.thr_amount), '#059669');
		if (data.bonus_amount > 0) items += row('Bonus', '+ ' + formatCurrency(data.bonus_amount), '#059669');
		items += `<div style="display:flex;justify-content:space-between;padding:10px 0 0;margin-top:6px;border-top:2px solid #a7f3d0;background:#ecfdf5;margin:6px -12px -12px;padding:10px 12px">`;
		items += `<span style="font-size:13px;font-weight:700;color:#111827">Gross Salary</span>`;
		items += `<span style="font-size:13px;font-weight:700;color:#111827">${formatCurrency(data.gross_salary)}</span>`;
		items += `</div>`;
		return items;
	}

	function renderDeductionItems(data: any): string {
		let items = '';
		const dedLabel = (v: number) => v > 0 ? formatCurrency(v) : '-';
		items += row('BPJS Kesehatan (1%)', dedLabel(data.bpjs_kesehatan), '#dc2626');
		items += row('BPJS JHT (2%)', dedLabel(data.bpjs_jht), '#dc2626');
		items += row('BPJS JP (1%)', dedLabel(data.bpjs_jp), '#dc2626');
		if (data.pph21_amount > 0) items += row('PPh 21', formatCurrency(data.pph21_amount), '#dc2626');
		if (data.loan_deduction > 0) items += row('Pinjaman', formatCurrency(data.loan_deduction), '#dc2626');
		for (const d of (data.deductions || []).filter((dd: any) => 
			dd.name !== 'BPJS Kesehatan' && dd.name !== 'BPJS JHT' && dd.name !== 'BPJS JP' &&
			dd.name !== 'PPh 21' && dd.name !== 'Pinjaman' && dd.name !== 'Lain-lain'
		)) items += row(d.name, formatCurrency(d.amount), '#dc2626');
		if (data.other_deductions > 0) items += row('Lain-lain', formatCurrency(data.other_deductions), '#dc2626');
		items += `<div style="display:flex;justify-content:space-between;padding:10px 0 0;margin-top:6px;border-top:2px solid #fecaca;background:#fef2f2;margin:6px -12px -12px;padding:10px 12px">`;
		items += `<span style="font-size:13px;font-weight:700;color:#111827">Total Potongan</span>`;
		items += `<span style="font-size:13px;font-weight:700;color:#dc2626">${formatCurrency(data.total_deductions)}</span>`;
		items += `</div>`;
		return items;
	}

	function row(label: string, value: string, valueColor: string): string {
		return `<div style="display:flex;justify-content:space-between;padding:5px 0;border-bottom:1px solid #f9fafb">
			<span style="font-size:12px;color:#4b5563">${label}</span>
			<span style="font-size:12px;font-weight:600;color:${valueColor}">${value}</span>
		</div>`;
	}

	onMount(async () => {
		const user = getUser();
		if (!user) {
			goto('/login');
			return;
		}

		if (!hasPermission('payslip', 'read')) {
			error = 'Anda tidak memiliki akses ke slip gaji';
			loading = false;
			return;
		}

		try {
			const res = await payroll.listMyPayslips();
			if (res.success) {
				payslips = res.data || [];
			} else {
				error = 'Gagal memuat slip gaji';
			}
		} catch {
			error = 'Terjadi kesalahan saat memuat data';
		} finally {
			loading = false;
		}
	});

	function formatCurrency(val: number) {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function statusBadge(status: string) {
		const map: Record<string, string> = {
			draft: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
			calculated: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
			approved: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
			paid: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900 dark:text-emerald-300',
		};
		return map[status] || 'bg-gray-100 text-gray-700';
	}

	const statusLabels: Record<string, string> = {
		draft: 'Draft',
		calculated: 'Dihitung',
		approved: 'Disetujui',
		paid: 'Dibayarkan',
	};

	const statusBadgeColors: Record<string, string> = {
		draft: 'bg-gray-100 text-gray-700 ring-gray-200 dark:bg-gray-700 dark:text-gray-300 dark:ring-gray-600',
		calculated: 'bg-blue-100 text-blue-700 ring-blue-200 dark:bg-blue-900 dark:text-blue-300 dark:ring-blue-800',
		approved: 'bg-green-100 text-green-700 ring-green-200 dark:bg-green-900 dark:text-green-300 dark:ring-green-800',
		paid: 'bg-emerald-100 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-300 dark:ring-emerald-800',
	};
</script>

<!-- eslint-disable svelte/no-useless-children-snippet -->

<div class="w-full">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Slip Gaji Saya</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Riwayat slip gaji Anda</p>
	</div>

	{#if loading}
		<div class="py-16 text-center text-gray-400 text-sm">
			<div class="animate-pulse space-y-3 max-w-md mx-auto">
				{#each [1,2,3] as _, i (i)}
					<div class="h-20 bg-gray-100 dark:bg-gray-800 rounded-xl"></div>
				{/each}
			</div>
		</div>
	{:else if error}
		<div class="py-16 text-center">
			<p class="text-sm text-red-500 mb-4">{error}</p>
			<button onclick={() => location.reload()} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium cursor-pointer">Muat Ulang</button>
		</div>
	{:else if payslips.length === 0}
		<EmptyState
			variant="empty"
			title="Belum ada slip gaji"
			description="Belum ada slip gaji tersedia untuk Anda."
		/>
	{:else}
		<!-- Mobile View -->
		<div class="md:hidden space-y-3">
			{#each payslips as slip (slip.id)}
				<MobileCard
					title={slip.period_name}
					subtitle={`${slip.employee_name} — ${slip.position_name}`}
					avatar={getInitials(slip.employee_name)}
					avatarColor={getAvatarTheme('payroll').gradientClasses}
					badges={[{ label: statusLabels[slip.status] || slip.status, color: statusBadgeColors[slip.status] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300' }]}
					onclick={() => goto(`/penggajian/payslip/${slip.payroll_period_id}/${slip.employee_id}`)}
				>
					{#snippet children()}
						<div class="flex items-center justify-between">
							<span class="text-xs text-gray-400 dark:text-gray-500">Take Home Pay</span>
							<span class="text-base font-bold text-blue-600 dark:text-blue-400 tabular-nums">{formatCurrency(slip.net_salary)}</span>
						</div>
						<div class="mt-2 pt-2 border-t border-gray-100 dark:border-gray-700 flex items-center justify-end">
							<button 
								onclick={(e: MouseEvent) => { e.stopPropagation(); downloadPdf(slip.payroll_period_id, slip.employee_id, slip.employee_name, slip.period_name); }}
								disabled={downloadingId === `${slip.payroll_period_id}_${slip.employee_id}`}
								class="inline-flex items-center justify-center gap-1.5 px-2.5 py-1.5 text-xs font-medium text-emerald-600 bg-emerald-50 hover:bg-emerald-100 rounded-lg transition-colors disabled:opacity-50 cursor-pointer"
							>
								{#if downloadingId === `${slip.payroll_period_id}_${slip.employee_id}`}
									<svg class="w-3 h-3 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
								{:else}
									<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
									</svg>
								{/if}
								<span>Download PDF</span>
							</button>
						</div>
					{/snippet}
				</MobileCard>
			{/each}
		</div>

		<!-- Desktop View -->
		<div class="hidden md:block bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-800 rounded-xl overflow-hidden shadow-sm">
			<div class="overflow-x-auto">
				<table class="w-full text-sm text-left">
					<thead class="text-xs text-gray-500 bg-gray-50/50 dark:bg-gray-800/50 uppercase border-b border-gray-200 dark:border-gray-800">
						<tr>
							<th class="px-6 py-4 font-medium">Periode</th>
							<th class="px-6 py-4 font-medium">Karyawan</th>
							<th class="px-6 py-4 font-medium text-right">Take Home Pay</th>
							<th class="px-6 py-4 font-medium text-center">Status</th>
							<th class="px-6 py-4 font-medium text-right">Aksi</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-gray-200 dark:divide-gray-800">
						{#each payslips as slip (slip.id)}
							<tr class="hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
								<td class="px-6 py-4 font-medium text-gray-900 dark:text-white whitespace-nowrap">
									{slip.period_name}
								</td>
								<td class="px-6 py-4">
									<div class="font-medium text-gray-900 dark:text-white">{slip.employee_name}</div>
									<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{slip.position_name}</div>
								</td>
								<td class="px-6 py-4 text-right font-bold text-blue-600 dark:text-blue-400 tabular-nums">
									{formatCurrency(slip.net_salary)}
								</td>
								<td class="px-6 py-4 text-center">
									<span class="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border {statusBadgeColors[slip.status] || 'bg-gray-100 text-gray-700'}">
										{statusLabels[slip.status] || slip.status}
									</span>
								</td>
								<td class="px-6 py-4 text-right">
									<div class="flex items-center justify-end gap-2">
										<button 
											onclick={() => goto(`/penggajian/payslip/${slip.payroll_period_id}/${slip.employee_id}`)}
											class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors cursor-pointer"
										>
											Lihat Slip
											<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
												<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
											</svg>
										</button>
										<button 
											onclick={() => downloadPdf(slip.payroll_period_id, slip.employee_id, slip.employee_name, slip.period_name)}
											disabled={downloadingId === `${slip.payroll_period_id}_${slip.employee_id}`}
											class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium text-emerald-600 hover:text-emerald-700 bg-emerald-50 hover:bg-emerald-100 rounded-lg transition-colors disabled:opacity-50 cursor-pointer"
										>
											{#if downloadingId === `${slip.payroll_period_id}_${slip.employee_id}`}
												<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
											{:else}
												<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
													<path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" />
												</svg>
											{/if}
											PDF
										</button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>