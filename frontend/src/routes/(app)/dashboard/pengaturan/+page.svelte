<script lang="ts">
	import { auth, ApiError } from '$lib/api.js';

	const settings = [
		{ 
			id: 'company',
			label: 'Nama Perusahaan', 
			value: 'PT Maju Jaya', 
			desc: 'Perusahaan pengembang properti dan infrastruktur',
			icon: 'M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z',
			color: 'text-blue-600 dark:text-blue-400',
			bgColor: 'bg-blue-50 dark:bg-blue-900/30'
		},
		{ 
			id: 'app',
			label: 'Versi Aplikasi', 
			value: 'HRMS v1.0.0', 
			desc: 'Human Resource Management System Terintegrasi',
			icon: 'M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25m18 0A2.25 2.25 0 0018.75 3H5.25A2.25 2.25 0 003 5.25m18 0V12a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 12V5.25',
			color: 'text-indigo-600 dark:text-indigo-400',
			bgColor: 'bg-indigo-50 dark:bg-indigo-900/30'
		},
		{ 
			id: 'stack',
			label: 'Stack Teknologi', 
			value: 'Go + PostgreSQL + SvelteKit', 
			desc: 'Arsitektur Full-stack web modern yang cepat dan aman',
			icon: 'M14.25 9.75L16.5 12l-2.25 2.25m-4.5 0L7.5 12l2.25-2.25M6 20.25h12A2.25 2.25 0 0020.25 18V6A2.25 2.25 0 0018 3.75H6A2.25 2.25 0 003.75 6v12A2.25 2.25 0 006 20.25z',
			color: 'text-emerald-600 dark:text-emerald-400',
			bgColor: 'bg-emerald-50 dark:bg-emerald-900/30'
		},
		{ 
			id: 'db',
			label: 'Database Server', 
			value: 'PostgreSQL 16', 
			desc: 'Dilengkapi enkripsi data tingkat lanjut (AES-256)',
			icon: 'M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125',
			color: 'text-amber-600 dark:text-amber-400',
			bgColor: 'bg-amber-50 dark:bg-amber-900/30'
		},
		{ 
			id: 'auth',
			label: 'Sistem Keamanan', 
			value: 'JWT + Rate Limiting', 
			desc: 'Pengamanan Role-Based Access Control tersertifikasi',
			icon: 'M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z',
			color: 'text-rose-600 dark:text-rose-400',
			bgColor: 'bg-rose-50 dark:bg-rose-900/30'
		},
	];
</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 mb-8">
		<div class="flex items-center gap-4">
			<div class="w-14 h-14 bg-gradient-to-br from-[#1A56DB] to-indigo-600 rounded-2xl flex items-center justify-center text-white font-bold text-lg shrink-0 shadow-lg shadow-blue-500/30">
				<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
				</svg>
			</div>
			<div>
				<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Pengaturan Sistem</h1>
				<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Ringkasan informasi dan konfigurasi inti aplikasi</p>
			</div>
		</div>
	</div>

	<!-- Info Grid -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 mb-8">
		{#each settings as item}
			<div class="group bg-white/80 dark:bg-gray-900/80 backdrop-blur-md rounded-2xl p-5 border border-gray-100 dark:border-gray-800 shadow-sm hover:shadow-xl hover:border-gray-200 dark:hover:border-gray-700 transition-all duration-300 hover:-translate-y-1">
				<div class="flex items-center gap-4 mb-4">
					<div class="w-12 h-12 rounded-xl {item.bgColor} flex items-center justify-center {item.color} shadow-inner transition-transform duration-300 group-hover:scale-110">
						<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="{item.icon}" />
						</svg>
					</div>
					<div>
						<h3 class="text-xs font-semibold uppercase tracking-wider text-gray-500 dark:text-gray-400">{item.label}</h3>
						<p class="text-sm font-bold text-gray-900 dark:text-gray-100 mt-0.5">{item.value}</p>
					</div>
				</div>
				<p class="text-xs leading-relaxed text-gray-500 dark:text-gray-400">{item.desc}</p>
			</div>
		{/each}
	</div>

	<!-- Note Banner -->
	<div class="bg-gradient-to-r from-blue-50 to-indigo-50 dark:from-blue-900/20 dark:to-indigo-900/20 border border-blue-100 dark:border-blue-800/50 rounded-2xl p-5 relative overflow-hidden group">
		<!-- Decorative bg circle -->
		<div class="absolute -right-10 -top-10 w-40 h-40 bg-blue-500/10 dark:bg-blue-500/20 rounded-full blur-2xl group-hover:bg-blue-500/20 transition-all duration-500"></div>
		
		<div class="flex flex-col sm:flex-row sm:items-center gap-4 relative z-10">
			<div class="w-12 h-12 rounded-full bg-white dark:bg-gray-800 shadow-sm flex items-center justify-center shrink-0 border border-blue-100 dark:border-blue-800/30">
				<svg class="w-6 h-6 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" />
				</svg>
			</div>
			<div>
				<h3 class="text-sm font-bold text-blue-900 dark:text-blue-200">Pengaturan Master Perusahaan</h3>
				<p class="text-xs text-blue-700/80 dark:text-blue-300/80 mt-1 leading-relaxed max-w-2xl">
					Informasi ini bersifat read-only untuk pengguna biasa. Untuk mengubah konfigurasi inti HR seperti master BPJS, approval workflow, atau parameter sistem lainnya, silakan hubungi Administrator HR Anda.
				</p>
			</div>
		</div>
	</div>
</div>
