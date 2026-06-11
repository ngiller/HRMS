<script lang="ts">
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api.js';

	let email = $state('');
	let isLoading = $state(false);
	let isSuccess = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';

		try {
			await auth.forgotPassword(email);
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

	function goBack() {
		goto('/login');
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
				Lupa Password?
			</h2>
			<p class="text-white/70 text-sm mt-4 max-w-xs">
				Jangan khawatir, kami akan mengirimkan tautan reset password ke email Anda.
			</p>
		</div>
		<div class="absolute -bottom-20 -right-20 w-80 h-80 rounded-full bg-white/5"></div>
		<div class="absolute -bottom-10 -right-10 w-40 h-40 rounded-full bg-white/5"></div>
		<div class="relative z-10 text-white/40 text-xs">
			&copy; 2026 HRMS Application
		</div>
	</div>

	<!-- Right Panel - Form -->
	<div class="w-3/5 flex items-center justify-center p-8">
		<div class="w-full max-w-sm">
			<div class="mb-8">
				<div class="flex items-center gap-3 mb-4">
					<button onclick={goBack} class="p-1.5 rounded-lg hover:bg-gray-100 transition cursor-pointer" aria-label="Kembali ke login">
						<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
						</svg>
					</button>
					<h1 class="text-2xl font-bold text-gray-900">Reset Password</h1>
				</div>
				<p class="text-gray-500 text-sm">Masukkan email Anda yang terdaftar untuk menerima tautan reset password.</p>
			</div>

			{#if isSuccess}
				<div class="bg-green-50 border border-green-200 rounded-xl p-6 text-center">
					<div class="w-14 h-14 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
						<svg class="w-7 h-7 text-green-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="M21.75 6.75v10.5a2.25 2.25 0 0 1-2.25 2.25h-15a2.25 2.25 0 0 1-2.25-2.25V6.75m19.5 0A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25m19.5 0v.243a2.25 2.25 0 0 1-1.07 1.916l-7.5 4.615a2.25 2.25 0 0 1-2.36 0L3.32 8.91a2.25 2.25 0 0 1-1.07-1.916V6.75" />
						</svg>
					</div>
					<h3 class="font-semibold text-gray-900 mb-1">Email Terkirim!</h3>
					<p class="text-sm text-gray-600">
						Jika email <strong>{email}</strong> terdaftar, Anda akan menerima tautan reset password.
					</p>
					<button onclick={goBack} class="mt-6 text-sm text-[#1A56DB] hover:underline font-medium cursor-pointer">
						Kembali ke Login
					</button>
				</div>
			{:else}
				<form onsubmit={handleSubmit} class="space-y-5">
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

					<button
						type="submit"
						class="w-full py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition active:scale-[0.98] disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer flex items-center justify-center gap-2"
						disabled={isLoading || !email}
					>
						{#if isLoading}
							<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none" aria-hidden="true">
								<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
								<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
							</svg>
							Mengirim...
						{:else}
							Kirim Tautan Reset
						{/if}
					</button>
				</form>

				<p class="text-center mt-6">
					<button onclick={goBack} class="text-sm text-[#1A56DB] hover:underline font-medium cursor-pointer">Kembali ke Login</button>
				</p>
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
			<p class="text-white/70 text-xs mt-1">Masukkan email Anda</p>
		</div>

		<div class="p-6">
			{#if isSuccess}
				<div class="text-center py-4">
					<div class="w-12 h-12 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-3">
						<svg class="w-6 h-6 text-green-600" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
							<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
						</svg>
					</div>
					<p class="text-sm text-gray-600">Cek email Anda untuk tautan reset password.</p>
					<button onclick={goBack} class="mt-4 text-xs text-[#1A56DB] font-medium cursor-pointer">Kembali ke Login</button>
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
							type="email"
							bind:value={email}
							class="w-full px-4 py-3 bg-gray-50 border border-gray-200 rounded-xl text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition disabled:opacity-50"
							placeholder="Email"
							required
							disabled={isLoading}
						/>
					</div>
					<button
						type="submit"
						class="w-full py-3 bg-[#1A56DB] text-white rounded-xl font-semibold text-sm hover:bg-[#1e40af] transition disabled:opacity-50 cursor-pointer"
						disabled={isLoading || !email}
					>
						{#if isLoading}Mengirim...{:else}Kirim Tautan Reset{/if}
					</button>
				</form>
				<button onclick={goBack} class="block text-center text-xs text-[#1A56DB] font-medium w-full mt-4 cursor-pointer">Kembali</button>
			{/if}
		</div>
	</div>
</div>
