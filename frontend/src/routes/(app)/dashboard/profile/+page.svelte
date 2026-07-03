<script lang="ts">
	import { onMount } from 'svelte';
	import { auth } from '$lib/api.js';

	interface UserData {
		id: string;
		employee_id: string;
		full_name: string;
		email: string;
		role_id: string;
		role_slug: string;
		role_name: string;
		position_name: string;
		department_name: string;
		avatar_initials: string;
	}

	let user = $state(auth.getUser() as UserData | null);
	let userData = $state<any>(null);
	let isLoading = $state(true);

	onMount(() => {
		loadProfile();
	});

	async function loadProfile() {
		isLoading = true;
		try {
			const res = await auth.getMe();
			if (res?.success) {
				userData = res.data;
			}
		} catch {
			// Fallback ke localStorage
			userData = user;
		} finally {
			isLoading = false;
		}
	}

	function getInitials(name: string): string {
		if (!name) return 'NA';
		return name.substring(0, 2).toUpperCase();
	}

	const infoItems = $derived([
		{ label: 'Nama Lengkap', value: userData?.full_name || user?.full_name || '-' },
		{ label: 'Email', value: userData?.email || user?.email || '-' },
		{ label: 'Jabatan', value: userData?.position_name || user?.position_name || '-' },
		{ label: 'Departemen', value: userData?.department_name || user?.department_name || '-' },
		{ label: 'Role', value: userData?.role_name || user?.role_name || '-' },
		{ label: 'ID Karyawan', value: userData?.employee_id || user?.employee_id || '-' },
	]);
</script>

<div class="w-full">
	<div class="flex items-center gap-4 mb-8">
		<div class="w-16 h-16 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-2xl flex items-center justify-center text-white font-bold text-xl shrink-0">
			{getInitials(userData?.full_name || user?.full_name || '')}
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Profil Saya</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">Informasi akun dan data diri</p>
		</div>
	</div>

	{#if isLoading}
		<div class="text-center py-12">
			<div class="animate-spin h-8 w-8 border-4 border-[#1A56DB] border-t-transparent rounded-full mx-auto"></div>
			<p class="text-sm text-gray-400 mt-3">Memuat data...</p>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
				<h2 class="font-semibold text-gray-900 dark:text-gray-100">Data Diri</h2>
			</div>
			<div class="divide-y divide-gray-100 dark:divide-gray-800">
				{#each infoItems as item}
					<div class="px-6 py-3.5 flex items-center justify-between">
						<span class="text-sm text-gray-500 dark:text-gray-400">{item.label}</span>
						<span class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.value}</span>
					</div>
				{/each}
			</div>
		</div>

		<div class="mt-6 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl p-4">
			<div class="flex items-start gap-3">
				<svg class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
				</svg>
				<div>
					<p class="text-sm font-medium text-amber-800 dark:text-amber-200">Data profil bersumber dari data karyawan</p>
					<p class="text-xs text-amber-600 dark:text-amber-400 mt-1">Untuk perubahan data, hubungi HR atau buka halaman Karyawan.</p>
				</div>
			</div>
		</div>
	{/if}
</div>
