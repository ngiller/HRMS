<script lang="ts">
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

<div class="w-full">
	<div class="mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Slip Gaji Saya</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Riwayat slip gaji Anda</p>
	</div>

	{#if loading}
		<div class="py-16 text-center text-gray-400 text-sm">
			<div class="animate-pulse space-y-3 max-w-md mx-auto">
				{#each [1,2,3] as _}
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
			{#each payslips as slip}
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
						{#each payslips as slip}
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
									<button 
										onclick={() => goto(`/penggajian/payslip/${slip.payroll_period_id}/${slip.employee_id}`)}
										class="inline-flex items-center justify-center gap-1.5 px-3 py-1.5 text-xs font-medium text-blue-600 hover:text-blue-700 bg-blue-50 hover:bg-blue-100 rounded-lg transition-colors cursor-pointer"
									>
										Lihat Slip
										<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
											<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5 21 12m0 0-7.5 7.5M21 12H3" />
										</svg>
									</button>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</div>
	{/if}
</div>