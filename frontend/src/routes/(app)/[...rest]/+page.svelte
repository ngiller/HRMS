<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';

	const moduleNames: Record<string, string> = {
		absensi: 'Absensi & Kehadiran',
		payroll: 'Payroll & Penggajian',
		cuti: 'Manajemen Cuti',
		reimbursement: 'Reimbursement',
		lembur: 'Lembur',
		pinjaman: 'Pinjaman Karyawan',
		kpi: 'KPI & Performance',
		pengaturan: 'Pengaturan',
		profile: 'Profile',
		'change-password': 'Ubah Password',
	};

	let moduleName = $derived.by(() => {
		const segments = $page.url.pathname.split('/').filter(Boolean);
		const key = segments[segments.length - 1]?.toLowerCase() || '';
		return moduleNames[key] || 'Fitur';
	});
</script>

<div class="max-w-lg mx-auto text-center py-16">
	<!-- Icon -->
	<div class="w-20 h-20 bg-blue-50 rounded-2xl flex items-center justify-center mx-auto mb-6">
		<svg class="w-10 h-10 text-[#1A56DB]" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
			<path stroke-linecap="round" stroke-linejoin="round" d="M11.42 15.17 17.25 21A2.652 2.652 0 0 0 21 17.25l-5.877-5.877M11.42 15.17l2.496-3.03c.317-.384.74-.626 1.208-.766M11.42 15.17l-4.655 5.653a2.548 2.548 0 1 1-3.586-3.586l6.837-5.63m5.108-.233c.55-.164 1.163-.188 1.743-.14a4.5 4.5 0 0 0 4.486-6.336l-3.276 3.277a3.004 3.004 0 0 1-2.25-2.25l3.276-3.276a4.5 4.5 0 0 0-6.336 4.486c.091 1.076-.071 2.264-.904 2.95l-.102.085m-1.745 1.437L5.909 7.5H4.5L2.25 3.75l1.5-1.5L7.5 4.5v1.409l4.087 4.113" />
		</svg>
	</div>

	<!-- Title -->
	<h1 class="text-xl font-bold text-gray-900 mb-2">
		{moduleName}
	</h1>
	<p class="text-sm text-gray-500 mb-8 max-w-sm mx-auto">
		Fitur ini sedang dalam tahap pengembangan dan akan segera hadir.
	</p>

	<!-- Progress bar -->
	<div class="w-full bg-gray-100 rounded-full h-2 mb-6 max-w-xs mx-auto">
		<div class="bg-[#1A56DB] h-2 rounded-full animate-pulse" style="width: 25%"></div>
	</div>

	<!-- Back button -->
	<button
		onclick={() => goto('/dashboard')}
		class="inline-flex items-center gap-2 px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition cursor-pointer"
	>
		<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
			<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
		</svg>
		Kembali ke Dashboard
	</button>
</div>
