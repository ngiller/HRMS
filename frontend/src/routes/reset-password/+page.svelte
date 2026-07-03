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
		<div class="w-full max-w-sm">
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
<div class="md:hidden min-h-screen bg-gradient-to-b from-[#1A56DB] to-[#1e3a8a] flex items-center justify-center p-4">
	<div class="w-full max-w-sm bg-white rounded-2xl shadow-xl overflow-hidden">
		<div class="bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] px-6 pt-10 pb-8 text-white text-center">
			<div class="w-14 h-14 bg-white/20 rounded-2xl flex items-center justify-center font-bold text-2xl mx-auto mb-4">
				HR
			</div>
			<h2 class="font-bold text-xl">Reset Password</h2>
			<p class="text-white/70 text-xs mt-1">Buat password baru</p>
		</div>

		<div class="p-6">
			{#if isSuccess}
				<div class="text-center py-4">
					<div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
						<svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
						</svg>
					</div>
					<p class="text-sm text-gray-600 font-medium">Password berhasil diubah!</p>
					<button onclick={goToLogin} class="mt-4 w-full py-3 bg-[#1A56DB] text-white rounded-xl font-semibold text-sm cursor-pointer">Login Sekarang</button>
				</div>
			{:else if !token}
				<div class="text-center py-4">
					<div class="w-12 h-12 bg-amber-100 rounded-full flex items-center justify-center mx-auto mb-3">
						<svg class="w-6 h-6 text-amber-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
						</svg>
					</div>
					<p class="text-sm text-gray-600">Token tidak valid</p>
					<button onclick={goToLogin} class="mt-3 text-xs text-[#1A56DB] font-medium cursor-pointer">Kembali</button>
				</div>
			{:else}
				<form onsubmit={handleSubmit} class="space-y-4">
					{#if errorMessage}
						<div class="bg-red-50 border border-red-200 text-red-700 text-xs rounded-lg px-3 py-2.5" role="alert">
							{errorMessage}
						</div>
					{/if}
					<div>
					<input
						id="mobile-new-password"
						type="password"
						name="new-password"
						bind:value={newPassword}
						class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition"
						placeholder="Password baru"
						required
						minlength={6}
						autocomplete="new-password"
					/>
					</div>
					<div>
					<input
						id="mobile-confirm-password"
						type="password"
						name="confirm-password"
						bind:value={confirmPassword}
						class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition"
						placeholder="Konfirmasi password"
						required
						minlength={6}
						autocomplete="new-password"
					/>
					</div>
					<button type="submit" class="w-full py-3 bg-[#1A56DB] text-white rounded-xl font-semibold text-sm hover:bg-[#1e40af] transition cursor-pointer" disabled={isLoading}>
						{isLoading ? 'Menyimpan...' : 'Simpan'}
					</button>
				</form>
			{/if}
		</div>
	</div>
</div>
