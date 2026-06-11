<script lang="ts">
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import { auth } from '$lib/api.js';

	let { children } = $props();
	let sidebarOpen = $state(false);

	// Auth guard: redirect to login if not authenticated
	$effect(() => {
		if (!auth.isAuthenticated()) {
			goto('/login', { replaceState: true });
		}
	});

	function handleLogout() {
		auth.clearSession();
		goto('/login', { replaceState: true });
	}

	type NavItem = {
		label: string;
		path: string;
		icon: string;
		badge?: string;
	};

	const navItems: NavItem[] = [
		{ label: 'Dashboard', path: '/dashboard', icon: 'M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6Zm0 9.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6Zm0 9.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z' },
		{ label: 'Karyawan', path: '/dashboard/karyawan', icon: 'M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z' },
		{ label: 'Absensi', path: '/dashboard/absensi', icon: 'M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5' },
		{ label: 'Payroll', path: '/dashboard/payroll', icon: 'M2.25 18.75a60.07 60.07 0 0 1 15.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 0 0 4.5 6h.75m13.5 0h.75a.75.75 0 0 0 .75-.75V4.5M12 3v18m-9-4.5h18' },
		{ label: 'Cuti', path: '/dashboard/cuti', icon: 'M21.752 15.002A9.72 9.72 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z' },
		{ label: 'Reimbursement', path: '/dashboard/reimbursement', icon: 'M9 12h3.75M9 15h3.75M9 18h3.75m3 .75H18a2.25 2.25 0 0 0 2.25-2.25V6.108c0-1.135-.845-2.098-1.976-2.192a48.424 48.424 0 0 0-1.123-.08m-5.801 0c-.065.21-.1.433-.1.664 0 .414.336.75.75.75h4.5a.75.75 0 0 0 .75-.75 2.25 2.25 0 0 0-.1-.664m-5.8 0A2.251 2.251 0 0 1 13.5 2.25H15c1.012 0 1.867.668 2.15 1.586m-5.8 0c-.376.023-.75.05-1.124.08C9.095 4.01 8.25 4.973 8.25 6.108V8.25m0 0H4.875c-.621 0-1.125.504-1.125 1.125v11.25c0 .621.504 1.125 1.125 1.125h9.75c.621 0 1.125-.504 1.125-1.125V9.375c0-.621-.504-1.125-1.125-1.125H8.25ZM6.75 12h.008v.008H6.75V12Zm0 3h.008v.008H6.75V15Zm0 3h.008v.008H6.75V18Z' },
		{ label: 'Lembur', path: '/dashboard/lembur', icon: 'M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z' },
		{ label: 'Pinjaman', path: '/dashboard/pinjaman', icon: 'M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Z' },
		{ label: 'KPI', path: '/dashboard/kpi', icon: 'M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z' },
	];

	function isActive(path: string): boolean {
		return $page.url.pathname === path;
	}

	let dropdownOpen = $state(false);
	let searchQuery = $state('');

	function toggleDropdown() {
		dropdownOpen = !dropdownOpen;
	}

	function closeDropdown() {
		dropdownOpen = false;
	}

	const menuItems = [
		{
			label: 'Profile',
			icon: 'M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z',
			href: '/dashboard/profile'
		},
		{
			label: 'Change Password',
			icon: 'M16.5 10.5V6.75a4.5 4.5 0 1 0-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 0 0 2.25-2.25v-6.75a2.25 2.25 0 0 0-2.25-2.25H6.75a2.25 2.25 0 0 0-2.25 2.25v6.75a2.25 2.25 0 0 0 2.25 2.25Z',
			href: '/dashboard/change-password'
		},
		{
			label: 'Logout',
			icon: 'M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15',
			href: '#',
			onclick: true
		}
	];
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<div onclick={closeDropdown} class="hidden md:flex h-screen overflow-hidden bg-gray-50">
	<!-- Sidebar -->
	<aside class="w-64 bg-white border-r border-gray-200 flex flex-col shrink-0" onclick={(e) => e.stopPropagation()}>
		<!-- Logo -->
		<div class="h-16 flex items-center gap-3 px-5 border-b border-gray-200 shrink-0">
			<div class="w-9 h-9 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-sm">HR</div>
			<div>
				<div class="font-bold text-gray-900 text-sm">HRMS</div>
				<div class="text-xs text-gray-400">PT Maju Jaya</div>
			</div>
		</div>

		<!-- Navigation -->
		<nav class="flex-1 overflow-y-auto p-3 space-y-0.5">
			{#each navItems as item}
				<a
					href={item.path}
					class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition {isActive(item.path) ? 'bg-blue-50 text-[#1A56DB] font-semibold' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
					aria-current={isActive(item.path) ? 'page' : undefined}
				>
					<svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
					</svg>
					<span>{item.label}</span>
					{#if item.badge}
						<span class="ml-auto bg-red-100 text-red-600 text-xs font-medium px-2 py-0.5 rounded-full">{item.badge}</span>
					{/if}
				</a>
			{/each}
		</nav>

		<!-- Bottom section -->
		<div class="p-3 border-t border-gray-200">
			<a
				href="/dashboard/pengaturan"
				class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm text-gray-600 hover:bg-gray-100 hover:text-gray-900 transition"
			>
				<svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
					<path stroke-linecap="round" stroke-linejoin="round" d="M15 12a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" />
				</svg>
				<span>Pengaturan</span>
			</a>
		</div>
	</aside>

	<!-- Right Side: Topbar + Content -->
	<div class="flex-1 flex flex-col min-w-0">
		<!-- Topbar -->
		<header class="h-16 bg-white border-b border-gray-200 flex items-center justify-between px-6 shrink-0">
			<!-- Left: Search -->
			<div class="relative">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" />
				</svg>
				<input
					type="search"
					bind:value={searchQuery}
					class="w-72 pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white transition placeholder:text-gray-400"
					placeholder="Cari karyawan..."
				/>
			</div>

			<!-- Right: Notification + User -->
			<div class="flex items-center gap-4">
				<!-- Notification Bell -->
				<button class="relative p-2 rounded-lg hover:bg-gray-100 transition cursor-pointer" aria-label="Notifikasi">
					<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 0 0 5.454-1.31A8.967 8.967 0 0 1 18 9.75V9A6 6 0 0 0 6 9v.75a8.967 8.967 0 0 1-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 0 1-5.714 0m5.714 0a3 3 0 1 1-5.714 0" />
					</svg>
					<span class="absolute top-1.5 right-1.5 w-2 h-2 bg-red-500 rounded-full"></span>
				</button>

				<!-- User Avatar + Dropdown -->
				<div class="relative">
					<button
						onclick={(e) => { e.stopPropagation(); toggleDropdown(); }}
						class="flex items-center gap-3 pl-4 border-l border-gray-200 cursor-pointer group"
						aria-label="Menu pengguna"
						aria-expanded={dropdownOpen}
					>
						<div class="w-9 h-9 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-full flex items-center justify-center text-white text-xs font-semibold shrink-0">
							BH
						</div>
						<div class="hidden sm:block text-left">
							<div class="text-sm font-medium text-gray-900 group-hover:text-[#1A56DB] transition">Budi Hartono</div>
							<div class="text-xs text-gray-400">Staff IT</div>
						</div>
						<svg class="w-4 h-4 text-gray-400 hidden sm:block transition {dropdownOpen ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="m19.5 8.25-7.5 7.5-7.5-7.5" />
						</svg>
					</button>

					<!-- Dropdown Menu -->
					{#if dropdownOpen}
						<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
						<div
							onclick={(e) => e.stopPropagation()}
							class="absolute right-0 top-full mt-2 w-56 bg-white rounded-xl shadow-lg border border-gray-200 py-1.5 z-50"
							role="menu"
						>
				{#each menuItems as item}
					{#if item.label === 'Logout'}
						<button
							onclick={handleLogout}
							class="flex items-center gap-3 w-full text-left px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition"
							role="menuitem"
						>
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
								<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
							</svg>
							{item.label}
						</button>
					{:else}
						<a
							href={item.href}
							class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition"
							role="menuitem"
						>
							<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
								<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
							</svg>
							{item.label}
						</a>
					{/if}
				{/each}
						</div>
					{/if}
				</div>
			</div>
		</header>

		<!-- Main Content -->
		<main class="flex-1 overflow-y-auto">
			<div class="p-6">
				{@render children()}
			</div>
		</main>
	</div>
</div>

<!-- Mobile Layout -->
<div class="md:hidden min-h-screen bg-gray-50 flex flex-col">
	<!-- Mobile Top Bar -->
	<div class="bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between shrink-0">
		<button onclick={() => sidebarOpen = !sidebarOpen} class="p-2 -ml-2 rounded-lg hover:bg-gray-100 transition cursor-pointer" aria-label="Buka menu">
			<svg class="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6.75h16.5M3.75 12h16.5m-16.5 5.25h16.5" />
			</svg>
		</button>
		<div class="flex items-center gap-2">
			<div class="w-7 h-7 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-xs">HR</div>
			<span class="font-bold text-sm text-gray-900">HRMS</span>
		</div>
		<!-- Mobile User Avatar -->
		<div class="relative">
			<button onclick={(e) => { e.stopPropagation(); toggleDropdown(); }} class="w-8 h-8 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-full flex items-center justify-center text-xs font-semibold text-white cursor-pointer" aria-label="Menu pengguna">
				BH
			</button>
			{#if dropdownOpen}
				<div onclick={(e) => e.stopPropagation()} class="absolute right-0 top-full mt-2 w-48 bg-white rounded-xl shadow-lg border border-gray-200 py-1.5 z-50" role="menu">
					{#each menuItems as item}
						{#if item.label === 'Logout'}
							<button
								onclick={handleLogout}
								class="flex items-center gap-3 w-full text-left px-4 py-2.5 text-sm text-red-600 hover:bg-red-50 transition"
								role="menuitem"
							>
								<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
									<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
								</svg>
								{item.label}
							</button>
						{:else}
							<a href={item.href} class="flex items-center gap-3 px-4 py-2.5 text-sm text-gray-700 hover:bg-gray-50 transition" role="menuitem">
								<svg class="w-4 h-4 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
									<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
								</svg>
								{item.label}
							</a>
						{/if}
					{/each}
				</div>
			{/if}
		</div>
	</div>

	<!-- Mobile Sidebar Overlay -->
	{#if sidebarOpen}
		<div class="fixed inset-0 z-50 flex">
			<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
			<div onclick={() => sidebarOpen = false} onkeydown={(e) => e.key === 'Escape' && (sidebarOpen = false)} class="absolute inset-0 bg-black/40 transition-opacity" role="presentation"></div>
			<aside class="relative w-72 bg-white h-full shadow-xl flex flex-col">
				<div class="h-16 flex items-center gap-3 px-5 border-b border-gray-200">
					<div class="w-9 h-9 bg-[#1A56DB] rounded-lg flex items-center justify-center text-white font-bold text-sm">HR</div>
					<div>
						<div class="font-bold text-gray-900 text-sm">HRMS</div>
						<div class="text-xs text-gray-400">PT Maju Jaya</div>
					</div>
				</div>
				<nav class="flex-1 overflow-y-auto p-3 space-y-0.5">
					{#each navItems as item}
						<a
							href={item.path}
							onclick={() => sidebarOpen = false}
							class="flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition {isActive(item.path) ? 'bg-blue-50 text-[#1A56DB] font-semibold' : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'}"
						>
							<svg class="w-5 h-5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
								<path stroke-linecap="round" stroke-linejoin="round" d={item.icon} />
							</svg>
							{item.label}
						</a>
					{/each}
				</nav>
			</aside>
		</div>
	{/if}

	<!-- Mobile Content -->
	<main class="flex-1 overflow-y-auto p-4">
		{@render children()}
	</main>
</div>
