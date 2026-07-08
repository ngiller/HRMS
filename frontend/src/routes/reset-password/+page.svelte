<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import { auth, ApiError } from '$lib/api.js';

	let token = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let isSuccess = $state(false);
	let errorMessage = $state('');

	// Extract token from URL query params on mount
	$effect(() => {
		const urlToken = $page.url.searchParams.get('token');
		if (urlToken) {
			token = urlToken;
		}
	});

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';

		if (newPassword !== confirmPassword) {
			errorMessage = 'Password tidak cocok';
			isLoading = false;
			return;
		}

		if (newPassword.length < 6) {
			errorMessage = 'Password minimal 6 karakter';
			isLoading = false;
			return;
		}

		try {
			await auth.resetPassword(token, newPassword);
			isSuccess = true;
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

	function goToLogin() {
		goto('/login');
	}
</script>

<!-- Desktop View -->
<div class="hidden md:flex min-h-screen">
	<!-- Left Panel -->
	<div class="w-2/5 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] p-12 flex flex-col justify-between text-white relative overflow-hidden">
		<div class="relative z-10">
			<div class="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center font-bold text-xl mb-10">
				HR
			</div>
			<h2 class="text-3xl font-bold leading-tight">
				Buat Password Baru
			</h2>
			<p class="text-white/70 text-sm mt-4 max-w-xs">
				Password baru Anda harus berbeda dari password sebelumnya.
			</p>
		</div>
		<div class="absolute -bottom-20 -right-20 w-80 h-80 rounded-full bg-white/5"></div>
		<div class="absolute -bottom-10 -right-10 w-40 h-40 rounded-full bg-white/5"></div>
		<div class="relative z-10 text-white/40 text-xs">
			&copy; 2026 HRMS Application
		</div>
	</div>

	<!-- Right Panel -->
	<div class="w-3/5 flex items-center justify-center p-8">
		<div class="w-full max-w-md">
			<div class="mb-8">
				<h1 class="text-2xl font-bold text-gray-900">Reset Password</h1>
				<p class="text-gray-500 text-sm mt-2">Masukkan password baru Anda.</p>
			</div>

			{#if isSuccess}
				<div class="bg-green-50 border border-green-200 rounded-xl p-6 text-center">
					<div class="w-14 h-14 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-7 h-7 text-green-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
						</svg>
					</div>
					<h3 class="font-semibold text-gray-900 mb-1">Password Berhasil Diubah!</h3>
					<p class="text-sm text-gray-600">Silakan login dengan password baru Anda.</p>
					<button onclick={goToLogin} class="mt-6 px-6 py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition cursor-pointer">
						Login Sekarang
					</button>
				</div>
			{:else if !token}
				<div class="bg-amber-50 border border-amber-200 rounded-xl p-6 text-center">
					<div class="w-14 h-14 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-7 h-7 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
						</svg>
					</div>
					<h3 class="font-semibold text-gray-900 mb-1">Token Tidak Ditemukan</h3>
					<p class="text-sm text-gray-600">Link reset password tidak valid. Silakan minta tautan baru.</p>
					<button onclick={goToLogin} class="mt-4 text-sm text-[#1A56DB] hover:underline font-medium cursor-pointer">Kembali ke Login</button>
				</div>
			{:else}
				<form onsubmit={handleSubmit} class="space-y-5">
					{#if errorMessage}
						<div class="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3" role="alert">
							{errorMessage}
						</div>
					{/if}

					<div>
						<label for="new-password" class="block text-sm font-medium text-gray-700 mb-1.5">Password Baru</label>
						<input
							id="new-password"
							type="password"
							name="new-password"
							bind:value={newPassword}
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
							placeholder="Minimal 6 karakter"
							required
							minlength={6}
							disabled={isLoading}
							autocomplete="new-password"
						/>
					</div>

					<div>
						<label for="confirm-password" class="block text-sm font-medium text-gray-700 mb-1.5">Konfirmasi Password Baru</label>
						<input
							id="confirm-password"
							type="password"
							name="confirm-password"
							bind:value={confirmPassword}
							class="w-full px-4 py-2.5 border border-gray-300 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition"
							placeholder="Ketik ulang password baru"
							required
							minlength={6}
							disabled={isLoading}
							autocomplete="new-password"
						/>
					</div>

					<button
						type="submit"
						class="w-full py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center gap-2"
						disabled={isLoading || !newPassword || !confirmPassword}
					>
						{#if isLoading}
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
							</svg>
							Menyimpan...
						{:else}
							Simpan Password Baru
						{/if}
					</button>
				</form>
			{/if}
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
				<svg class="w-8 h-8 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
					<path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
				</svg>
			</div>
			<h2 class="font-bold text-2xl text-white tracking-tight">Buat Password Baru</h2>
			<p class="text-white/80 text-sm mt-1.5 font-medium text-center px-4">Pastikan password baru Anda aman dan berbeda.</p>
		</div>
	</div>

	<!-- Form Section -->
	<div class="flex-1 px-6 py-8 flex flex-col justify-center">
		{#if isSuccess}
			<div class="text-center py-6 animate-in fade-in slide-in-from-bottom-4">
				<div class="w-20 h-20 bg-green-50 rounded-full flex items-center justify-center mx-auto mb-6 shadow-sm border border-green-100">
					<svg class="w-10 h-10 text-green-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
					</svg>
				</div>
				<h3 class="text-xl font-bold text-gray-900 mb-2">Berhasil Diubah!</h3>
				<p class="text-[15px] text-gray-500 mb-8 px-4">
					Password Anda telah diperbarui. Silakan login kembali.
				</p>
				<button onclick={goToLogin} class="w-full py-4 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] text-white rounded-xl font-bold text-[15px] hover:shadow-lg transition-all active:scale-[0.98] shadow-sm">
					Login Sekarang
				</button>
			</div>
		{:else if !token}
			<div class="text-center py-6 animate-in fade-in slide-in-from-bottom-4">
				<div class="w-20 h-20 bg-amber-50 rounded-full flex items-center justify-center mx-auto mb-6 shadow-sm border border-amber-100">
					<svg class="w-10 h-10 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
					</svg>
				</div>
				<h3 class="text-xl font-bold text-gray-900 mb-2">Tautan Kedaluwarsa</h3>
				<p class="text-[15px] text-gray-500 mb-8 px-4">
					Tautan reset password ini tidak valid atau sudah tidak berlaku.
				</p>
				<button onclick={goToLogin} class="w-full py-4 bg-gray-50 border border-gray-200 text-gray-700 rounded-xl font-bold text-[15px] hover:bg-gray-100 transition-all active:scale-[0.98] shadow-sm">
					Kembali ke Login
				</button>
			</div>
		{:else}
			<form onsubmit={handleSubmit} class="space-y-5">
				{#if errorMessage}
					<div class="bg-red-50 border-l-4 border-red-500 text-red-700 text-sm rounded-r-lg px-4 py-3 shadow-sm animate-in fade-in slide-in-from-top-2" role="alert">
						<p class="font-medium">{errorMessage}</p>
					</div>
				{/if}
				
				<div class="space-y-1.5">
					<label for="mobile-new-password" class="text-xs font-bold text-gray-500 uppercase tracking-wider ml-1">Password Baru</label>
					<div class="relative">
						<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
							<svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
						</div>
						<input
							id="mobile-new-password"
							type="password"
							name="new-password"
							bind:value={newPassword}
							class="w-full pl-11 pr-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl text-[15px] font-medium text-gray-900 outline-none focus:bg-white focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all shadow-sm"
							placeholder="Minimal 6 karakter"
							required
							minlength={6}
							disabled={isLoading}
							autocomplete="new-password"
						/>
					</div>
				</div>

				<div class="space-y-1.5">
					<label for="mobile-confirm-password" class="text-xs font-bold text-gray-500 uppercase tracking-wider ml-1">Konfirmasi Password</label>
					<div class="relative">
						<div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
							<svg class="w-5 h-5 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5"><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" /></svg>
						</div>
						<input
							id="mobile-confirm-password"
							type="password"
							name="confirm-password"
							bind:value={confirmPassword}
							class="w-full pl-11 pr-4 py-3.5 bg-gray-50 border border-gray-200 rounded-xl text-[15px] font-medium text-gray-900 outline-none focus:bg-white focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition-all shadow-sm"
							placeholder="Ketik ulang password baru"
							required
							minlength={6}
							disabled={isLoading}
							autocomplete="new-password"
						/>
					</div>
				</div>

				<button
					type="submit"
					class="w-full py-4 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] text-white rounded-xl font-bold text-[15px] hover:shadow-lg hover:shadow-blue-500/30 transition-all active:scale-[0.98] disabled:opacity-70 disabled:cursor-not-allowed flex items-center justify-center gap-2 mt-4"
					disabled={isLoading || !newPassword || !confirmPassword}
				>
					{#if isLoading}
						<svg class="animate-spin h-5 w-5" viewBox="0 0 24 24" fill="none" aria-hidden="true">
							<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
							<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
						</svg>
						Menyimpan...
					{:else}
						Simpan Password Baru
					{/if}
				</button>
			</form>
		{/if}

		<div class="mt-auto pt-8">
			<p class="text-[10px] text-gray-400 font-bold text-center mt-6 uppercase tracking-[0.2em]">
				HRMS &copy; 2026
			</p>
		</div>
	</div>
</div>
