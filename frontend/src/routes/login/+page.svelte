<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api.js';

	let email = $state('admin@company.com');
	let password = $state('admin123');
	let remember = $state(true);
	let isLoading = $state(false);
	let errorMessage = $state('');

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
		<div class="w-full max-w-sm">
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
						bind:value={email}
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
						placeholder="nama@company.com"
						required
						disabled={isLoading}
					/>
				</div>

				<div>
					<label for="password" class="block text-sm font-medium text-gray-700 mb-1.5">Password</label>
					<input
						id="password"
						type="password"
						bind:value={password}
						class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
						placeholder="••••••••"
						required
						disabled={isLoading}
					/>
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

				<button type="button" class="w-full py-2.5 border border-gray-300 text-gray-700 rounded-lg font-medium text-sm hover:bg-gray-50 transition flex items-center justify-center gap-3 cursor-pointer" disabled={isLoading}>
					<svg class="w-5 h-5" viewBox="0 0 21 21" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
						<rect x="1" y="1" width="18.5" height="18.5" rx="2" fill="#0078D4"/>
						<path d="M12.928 10.5L7.5 16V5L12.928 10.5Z" fill="white"/>
					</svg>
					Masuk dengan Microsoft
				</button>
			</form>

			<p class="text-xs text-gray-400 text-center mt-8">
				<span class="text-[#1A56DB]">HRMS</span> &mdash; Sistem Informasi Sumber Daya Manusia
			</p>
		</div>
	</div>
</div>

<!-- Mobile View -->
<div class="md:hidden min-h-screen bg-gradient-to-b from-[#1A56DB] to-[#1e3a8a] flex items-center justify-center p-4">
	<div class="w-full max-w-sm bg-white rounded-2xl shadow-xl overflow-hidden">
		<!-- Mobile Header -->
		<div class="bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] px-6 pt-10 pb-8 text-white text-center">
			<div class="w-14 h-14 bg-white/20 rounded-2xl flex items-center justify-center font-bold text-2xl mx-auto mb-4">
				HR
			</div>
			<h2 class="font-bold text-xl">HRMS</h2>
			<p class="text-white/70 text-xs mt-1">Masuk ke akun Anda</p>
		</div>

		<!-- Mobile Form -->
		<form onsubmit={handleLogin} class="p-6 space-y-4">
			{#if errorMessage}
				<div class="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg px-3 py-2.5" role="alert">
					{errorMessage}
				</div>
			{/if}
			<div>
				<input
					type="email"
					bind:value={email}
					class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition disabled:opacity-50"
					placeholder="Email"
					required
					disabled={isLoading}
				/>
			</div>
			<div>
				<input
					type="password"
					bind:value={password}
					class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition disabled:opacity-50"
					placeholder="Password"
					required
					disabled={isLoading}
				/>
			</div>
			<button
				type="submit"
				class="w-full py-3 bg-[#1A56DB] text-white rounded-xl font-semibold text-sm hover:bg-[#1e40af] transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center gap-2"
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
			<button type="button" onclick={goToForgotPassword} class="block text-center text-xs text-[#1A56DB] font-medium w-full cursor-pointer">Lupa password?</button>
		</form>
	</div>
</div>
