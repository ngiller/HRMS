<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api.js';

	let email = $state('admin@company.com');
	let password = $state('admin123');
	let remember = $state(true);
	let isLoading = $state(false);
	let errorMessage = $state('');
	let showPassword = $state(false);

	async function handleLogin(e: Event) {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';

		try {
			const response = await auth.login(email, password);
			auth.saveSession(response);
			goto('/dashboard', { replaceState: true });
		} catch (error) {
			if (error instanceof ApiError) {
				errorMessage = error.message;
			} else {
				errorMessage = 'Terjadi kesalahan. Silakan coba lagi.';
			}
		} finally {
			isLoading = false;
		}
	}

	function goToForgotPassword() {
		goto('/forgot-password');
	}
</script>

<!-- Desktop View -->
<div class="hidden md:flex min-h-screen">
	<!-- Left Panel - Branding -->
	<div class="w-2/5 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] p-12 flex flex-col justify-between text-white relative overflow-hidden">
		<div class="relative z-10">
			<div class="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center font-bold text-xl mb-10">
				HR
			</div>
			<h2 class="text-3xl font-bold leading-tight">
				Sistem Informasi<br />
				Sumber Daya Manusia
			</h2>
			<p class="text-white/70 text-sm mt-4 max-w-xs">
				HRMS terintegrasi untuk perusahaan Indonesia — kelola karyawan, absensi, payroll, dan HR operations dalam satu platform.
			</p>
		</div>

		<!-- Decorative circles -->
		<div class="absolute -bottom-20 -right-20 w-80 h-80 rounded-full bg-white/5"></div>
		<div class="absolute -bottom-10 -right-10 w-40 h-40 rounded-full bg-white/5"></div>

		<div class="relative z-10 text-white/40 text-xs">
			&copy; 2026 HRMS Application
		</div>
	</div>

	<!-- Right Panel - Login Form -->
	<div class="w-3/5 flex items-center justify-center p-8">
		<div class="w-full max-w-md">
			<div class="mb-8">
				<h1 class="text-2xl font-bold text-gray-900">Masuk</h1>
				<p class="text-gray-500 mt-2">Silakan login dengan akun Anda</p>
			</div>

			<form onsubmit={handleLogin} class="space-y-5">
				{#if errorMessage}
					<div class="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3" role="alert">
						{errorMessage}
					</div>
				{/if}

				<div>
					<label for="email" class="block text-sm font-medium text-gray-700 mb-1.5">Email</label>
					<input
						id="email"
						type="email"
						name="email"
						bind:value={email}
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
						placeholder="nama@company.com"
						required
						disabled={isLoading}
						autocomplete="email"
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1.5">Password</label>
					<div class="relative">
						<input
							id="password"
							type={showPassword ? 'text' : 'password'}
							name="password"
							bind:value={password}
							class="w-full px-4 pr-10 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
							placeholder="••••••••"
							required
							disabled={isLoading}
							autocomplete="current-password"
						/>
						<button type="button" onclick={() => showPassword = !showPassword} class="absolute inset-y-0 right-0 pr-3 flex items-center text-gray-400 hover:text-gray-600 focus:outline-none">
							{#if showPassword}
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
								</svg>
							{:else}
								<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
								</svg>
							{/if}
						</button>
					</div>
				</div>

				<div class="flex items-center justify-between text-sm">
					<label class="flex items-center gap-2 text-gray-600 cursor-pointer">
						<input type="checkbox" bind:checked={remember} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB]" />
						<span>Ingat saya</span>
					</label>
					<button type="button" onclick={goToForgotPassword} class="text-[#1A56DB] hover:underline font-medium cursor-pointer">Lupa password?</button>
				</div>

				<button
					type="submit"
					class="w-full py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center gap-2"
					disabled={isLoading}
				>
					{#if isLoading}
						<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
						</svg>
						Memproses...
					{:else}
						Masuk
					{/if}
				</button>

				<div class="relative my-6">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t border-gray-200"></div>
					</div>
					<div class="relative flex justify-center text-sm">
						<span class="bg-white px-3 text-gray-400">atau</span>
					</div>
				</div>

				<div class="space-y-3">
					<button type="button" class="w-full py-2.5 border border-gray-300 text-gray-700 rounded-lg font-medium text-sm hover:bg-gray-50 transition flex items-center justify-center gap-3 cursor-pointer" disabled={isLoading}>
						<svg class="w-5 h-5" viewBox="0 0 21 21" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
							<rect x="1" y="1" width="18.5" height="18.5" rx="2" fill="#0078D4"/>
							<path d="M12.928 10.5L7.5 16V5L12.928 10.5Z" fill="white"/>
						</svg>
						Masuk dengan Microsoft
					</button>

					<button type="button" class="w-full py-2.5 border border-gray-300 text-gray-700 rounded-lg font-medium text-sm hover:bg-gray-50 transition flex items-center justify-center gap-3 cursor-pointer" disabled={isLoading}>
						<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
							<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
							<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.16v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
							<path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.16C1.43 8.55 1 10.22 1 12s.43 3.45 1.16 4.93l3.68-2.84z" fill="#FBBC05"/>
							<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.16 7.07l3.68 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
						</svg>
						Masuk dengan Google
					</button>
				</div>
			</form>

			<!-- Demo Credentials -->
			<div class="mt-6 p-3 bg-blue-50 border border-blue-100 rounded-lg">
				<p class="text-xs font-semibold text-blue-700 mb-1">🔑 Demo Super Admin</p>
				<p class="text-xs text-blue-600 font-mono">admin@company.com / admin123</p>
			</div>

			<p class="text-xs text-gray-400 text-center mt-6">
				<span class="text-[#1A56DB]">HRMS</span> &mdash; Sistem Informasi Sumber Daya Manusia
			</p>
		</div>
	</div>
</div>

<!-- Mobile View -->
<div class="md:hidden flex flex-col min-h-screen bg-white">
	<!-- Hero Header -->
	<div class="relative bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] px-6 pt-16 pb-12 overflow-hidden rounded-b-[40px] shadow-lg">
		<!-- Decorative blobs -->
		<div class="absolute top-0 right-0 -mr-8 -mt-8 w-32 h-32 rounded-full bg-white/10 blur-xl"></div>
		<div class="absolute bottom-0 left-0 -ml-8 -mb-8 w-24 h-24 rounded-full bg-white/10 blur-lg"></div>
		
		<div class="relative z-10 flex flex-col items-center">
			<div class="w-16 h-16 bg-white/20 backdrop-blur-md rounded-2xl flex items-center justify-center shadow-inner border border-white/20 mb-5">
				<span class="font-bold text-3xl text-white tracking-wider">HR</span>
			</div>
			<h2 class="font-bold text-2xl text-white tracking-tight">Selamat Datang</h2>
			<p class="text-white/80 text-sm mt-1.5 font-medium">Masuk untuk mengelola HR Anda</p>
		</div>
	</div>

	<!-- Form Section -->
	<div class="flex-1 px-6 py-8 flex flex-col justify-center">
		<form onsubmit={handleLogin} class="space-y-5">
			{#if errorMessage}
				<div class="bg-red-50 border-l-4 border-red-500 text-red-700 text-sm rounded-r-lg px-4 py-3 shadow-sm animate-in fade-in slide-in-from-top-2" role="alert">
					<p class="font-medium">{errorMessage}</p>
				</div>
			{/if}

			<div class="space-y-1.5">
				<label for="mobile-email" class="text-xs font-bold text-gray-500 uppercase tracking-wider ml-1">Email</label>
				<div class="relative">
					<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
						<svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 01-2.25 2.25h-15a2.25 2.25 0 01-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0019.5 4.5h-15a2.25 2.25 0 00-2.25 2.25m19.5 0v.243a2.25 2.25 0 01-1.07 1.916l-7.5 4.615a2.25 2.25 0 01-2.36 0L3.32 8.91a2.25 2.25 0 01-1.07-1.916V6.75" /></svg>
					</div>
					<input
						id="mobile-email"
						type="email"
						name="email"
						bind:value={email}
						class="w-full pl-11 pr-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl text-[15px] font-medium text-gray-900 outline-none focus:bg-white focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all shadow-sm"
						placeholder="nama@company.com"
						required
						disabled={isLoading}
						autocomplete="email"
					/>
				</div>
			</div>

			<div class="space-y-1.5">
				<label for="mobile-password" class="text-xs font-bold text-gray-500 uppercase tracking-wider ml-1">Password</label>
				<div class="relative">
					<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
						<svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
					</div>
					<input
						id="mobile-password"
						type={showPassword ? 'text' : 'password'}
						name="password"
						bind:value={password}
						class="w-full pl-11 pr-12 py-3.5 bg-gray-50 border border-gray-200 rounded-xl text-[15px] font-medium text-gray-900 outline-none focus:bg-white focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all shadow-sm"
						placeholder="••••••••"
						required
						disabled={isLoading}
						autocomplete="current-password"
					/>
					<button type="button" onclick={() => showPassword = !showPassword} class="absolute inset-y-0 right-0 pr-4 flex items-center text-gray-400 hover:text-gray-600 focus:outline-none">
						{#if showPassword}
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
							</svg>
						{:else}
							<svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21" />
							</svg>
						{/if}
					</button>
				</div>
			</div>

			<div class="flex items-center justify-between mt-2 mb-4">
				<label class="flex items-center gap-2 text-gray-600 cursor-pointer">
					<input type="checkbox" bind:checked={remember} class="rounded border-gray-300 text-[#1A56DB] focus:ring-[#1A56DB] w-4 h-4" />
					<span class="text-sm font-medium">Ingat saya</span>
				</label>
				<button type="button" onclick={goToForgotPassword} class="text-sm text-[#1A56DB] font-bold hover:underline">Lupa password?</button>
			</div>

			<button
				type="submit"
				class="w-full py-4 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] text-white rounded-xl font-bold text-[15px] hover:shadow-lg hover:shadow-blue-500/30 transition-all active:scale-[0.98] disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-4"
				disabled={isLoading}
			>
				{#if isLoading}
					<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none" aria-hidden="true">
						<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
						<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
					</svg>
					Memproses...
				{:else}
					Masuk ke Sistem
				{/if}
			</button>

			<div class="relative mt-8 mb-6">
				<div class="absolute inset-0 flex items-center">
					<div class="w-full border-t border-gray-200"></div>
				</div>
				<div class="relative flex justify-center text-sm">
					<span class="bg-white px-4 text-gray-400 font-bold text-[10px] tracking-widest uppercase">Atau gunakan</span>
				</div>
			</div>

			<div class="space-y-3">
				<button type="button" class="w-full py-3.5 bg-white border border-gray-200 text-gray-700 rounded-xl font-bold text-[14px] hover:bg-gray-50 hover:border-gray-300 transition-all shadow-sm flex items-center justify-center gap-3 active:scale-[0.98]" disabled={isLoading}>
					<svg class="w-5 h-5" viewBox="0 0 21 21" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
						<rect x="1" y="1" width="18.5" height="18.5" rx="2" fill="#0078D4"/>
						<path d="M12.928 10.5L7.5 16V5L12.928 10.5Z" fill="white"/>
					</svg>
					Masuk dengan Microsoft
				</button>

				<button type="button" class="w-full py-3.5 bg-white border border-gray-200 text-gray-700 rounded-xl font-bold text-[14px] hover:bg-gray-50 hover:border-gray-300 transition-all shadow-sm flex items-center justify-center gap-3 active:scale-[0.98]" disabled={isLoading}>
					<svg class="w-5 h-5" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
						<path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
						<path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.16v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
						<path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.16C1.43 8.55 1 10.22 1 12s.43 3.45 1.16 4.93l3.68-2.84z" fill="#FBBC05"/>
						<path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.16 7.07l3.68 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
					</svg>
					Masuk dengan Google
				</button>
			</div>
		</form>

		<div class="mt-auto pt-8">
			<!-- Demo Credentials Mobile -->
			<div class="p-3 bg-blue-50/80 border border-blue-100 rounded-xl text-center shadow-sm">
				<p class="text-[10px] font-bold text-blue-800 mb-1.5 tracking-wider uppercase">🔑 Akun Demo</p>
				<p class="text-[13px] text-blue-600 font-bold font-mono">admin@company.com</p>
				<p class="text-[13px] text-blue-600 font-medium font-mono mt-0.5">admin123</p>
			</div>
			
			<p class="text-[10px] text-gray-400 font-bold text-center mt-6 uppercase tracking-[0.2em]">
				HRMS &copy; 2026
			</p>
		</div>
	</div>
</div>
