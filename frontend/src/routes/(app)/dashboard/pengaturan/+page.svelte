<script lang="ts">
	import { auth, ApiError } from '$lib/api.js';

	let isLoading = $state(true);

	$effect(() => {
		// Just load and show available info
		isLoading = false;
	});

	const settings = [
		{ label: 'Nama Perusahaan', value: 'PT Maju Jaya', desc: 'Perusahaan pengembang properti dan infrastruktur' },
		{ label: 'Aplikasi', value: 'HRMS v1.0.0', desc: 'Human Resource Management System' },
		{ label: 'Stack Teknologi', value: 'Go (Fiber) + PostgreSQL + SvelteKit', desc: 'Full-stack web application' },
		{ label: 'Database', value: 'PostgreSQL 16', desc: 'Dengan enkripsi data sensitif (AES-256)' },
		{ label: 'Autentikasi', value: 'JWT + Refresh Token Rotation', desc: 'Dengan RBAC & Rate Limiting' },
	];
</script>

<div class="w-full">
	<div class="flex items-center gap-4 mb-8">
		<div class="w-12 h-12 bg-[#1A56DB] rounded-2xl flex items-center justify-center text-white font-bold text-lg shrink-0">
			<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
			</svg>
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Pengaturan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">Informasi sistem dan konfigurasi aplikasi</p>
		</div>
	</div>

	<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
		<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
			<h2 class="font-semibold text-gray-900 dark:text-gray-100">Informasi Sistem</h2>
		</div>
		<div class="divide-y divide-gray-100 dark:divide-gray-800">
			{#each settings as item}
				<div class="px-6 py-4">
					<div class="flex items-center justify-between">
						<div>
							<p class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.label}</p>
							<p class="text-xs text-gray-400 dark:text-gray-500 mt-0.5">{item.desc}</p>
						</div>
						<span class="text-sm text-gray-600 dark:text-gray-400 font-mono">{item.value}</span>
					</div>
				</div>
			{/each}
		</div>
	</div>

	<div class="mt-6 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-xl p-4">
		<div class="flex items-start gap-3">
			<svg class="w-5 h-5 text-blue-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
			</svg>
			<div>
				<p class="text-sm font-medium text-blue-800 dark:text-blue-200">Pengaturan perusahaan dapat dikelola melalui database</p>
				<p class="text-xs text-blue-600 dark:text-blue-400 mt-1">Untuk mengubah konfigurasi HR (BPJS, approval, dll), hubungi administrator.</p>
			</div>
		</div>
	</div>
</div>
