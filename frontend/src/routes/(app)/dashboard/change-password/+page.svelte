<script lang="ts">
/* eslint-disable svelte/no-navigation-without-resolve */
	import { goto } from '$app/navigation';
	import { auth, ApiError } from '$lib/api.js';

	let currentPassword = $state('');
	let newPassword = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let isSuccess = $state(false);
	let errorMessage = $state('');

	async function handleSubmit(e: Event) {
		e.preventDefault();
		isLoading = true;
		errorMessage = '';

		if (newPassword !== confirmPassword) {
			errorMessage = 'Konfirmasi password tidak cocok';
			isLoading = false;
			return;
		}
		if (newPassword.length < 6) {
			errorMessage = 'Password baru minimal 6 karakter';
			isLoading = false;
			return;
		}
		if (currentPassword === newPassword) {
			errorMessage = 'Password baru harus berbeda dari password saat ini';
			isLoading = false;
			return;
		}

		try {
			const res = await fetch('/api/auth/change-password', {
				method: 'PUT',
				headers: {
					'Content-Type': 'application/json',
					'Authorization': `Bearer ${localStorage.getItem('hrms_access_token')}`,
				},
				body: JSON.stringify({
					current_password: currentPassword,
					new_password: newPassword,
				}),
			});
			const data = await res.json();
			if (!res.ok || !data.success) {
				throw new ApiError(data.message || 'Gagal mengubah password', res.status);
			}
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

	function goToProfile() {
		goto('/dashboard');
	}
</script>

<div class="max-w-full w-full">
	<div class="flex items-center gap-4 mb-8">
		<button onclick={goToProfile} class="p-2 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer" aria-label="Kembali">
			<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" aria-hidden="true">
				<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
			</svg>
		</button>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Ubah Password</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">Ganti password akun Anda</p>
		</div>
	</div>

	{#if isSuccess}
		<div class="bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-xl p-8 text-center">
			<div class="w-16 h-16 bg-green-100 dark:bg-green-900/40 rounded-full flex items-center justify-center mx-auto mb-4">
				<svg class="w-8 h-8 text-green-600 dark:text-green-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" aria-hidden="true">
					<path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
				</svg>
			</div>
			<h3 class="font-semibold text-gray-900 dark:text-gray-100 mb-2">Password Berhasil Diubah!</h3>
			<p class="text-sm text-gray-500 dark:text-gray-400 mb-6">Password akun Anda telah berhasil diperbarui.</p>
			<button onclick={goToProfile} class="px-6 py-2.5 bg-[#1A56DB] text-white rounded-lg font-semibold text-sm hover:bg-[#1e40af] transition cursor-pointer">
				Kembali ke Dashboard
			</button>
		</div>
	{:else}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden">
			<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
				<h2 class="font-semibold text-gray-900 dark:text-gray-100">Form Ubah Password</h2>
			</div>
			<form onsubmit={handleSubmit} class="p-6 space-y-5">
				{#if errorMessage}
					<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-400 text-sm rounded-lg px-4 py-3" role="alert">
						{errorMessage}
					</div>
				{/if}

				<div>
					<label for="current-password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Password Saat Ini</label>
					<input
						id="current-password"
						type="password"
						name="current-password"
						bind:value={currentPassword}
						class="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
						placeholder="Masukkan password saat ini"
						required
						disabled={isLoading}
						autocomplete="current-password"
					/>
				</div>

				<div>
					<label for="new-password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Password Baru</label>
					<input
						id="new-password"
						type="password"
						name="new-password"
						bind:value={newPassword}
						class="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
						placeholder="Minimal 6 karakter"
						required
						minlength={6}
						disabled={isLoading}
						autocomplete="new-password"
					/>
				</div>

				<div>
					<label for="confirm-password" class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1.5">Konfirmasi Password Baru</label>
					<input
						id="confirm-password"
						type="password"
						name="confirm-password"
						bind:value={confirmPassword}
						class="w-full px-4 py-2.5 border border-gray-300 dark:border-gray-600 rounded-lg text-sm focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] outline-none transition bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100"
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
					disabled={isLoading || !currentPassword || !newPassword || !confirmPassword}
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
		</div>
	{/if}
</div>
