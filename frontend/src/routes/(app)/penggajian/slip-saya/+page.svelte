<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { payroll } from '$lib/api';
	import { getUser, hasPermission } from '$lib/permissions';

	let payslips = $state<any[]>([]);
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
</script>

<div class="w-full max-w-full px-4 xl:px-8 py-6">
	<div class="mb-6">
		<h1 class="text-2xl font-semibold text-gray-900 dark:text-white">Slip Gaji Saya</h1>
		<p class="text-sm text-gray-500 dark:text-gray-400 mt-1">Riwayat slip gaji Anda</p>
	</div>

	{#if loading}
		<div class="py-16 text-center text-gray-400 text-sm">Memuat data...</div>
	{:else if error}
		<div class="py-16 text-center text-red-500 text-sm">{error}</div>
	{:else if payslips.length === 0}
		<div class="py-16 text-center text-gray-400 text-sm">Belum ada slip gaji tersedia</div>
	{:else}
		<div class="space-y-3">
			{#each payslips as slip}
				<button
					onclick={() => goto(`/penggajian/payslip/${slip.payroll_period_id}/${slip.employee_id}`)}
					class="w-full text-left bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4 hover:shadow-md transition-shadow cursor-pointer"
				>
					<div class="flex items-center justify-between">
						<div>
							<p class="font-medium text-gray-900 dark:text-white">{slip.period_name}</p>
							<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">{slip.employee_name} — {slip.position_name}</p>
						</div>
						<div class="text-right">
							<p class="font-semibold text-gray-900 dark:text-white">{formatCurrency(slip.net_salary)}</p>
							<span class="inline-block px-2 py-0.5 text-xs rounded-full mt-1 {statusBadge(slip.status)}">
								{slip.status}
							</span>
						</div>
					</div>
				</button>
			{/each}
		</div>
	{/if}
</div>