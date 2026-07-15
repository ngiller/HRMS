<script lang="ts">
	/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount } from 'svelte';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import config from '$lib/config.js';
	import { attendance } from '$lib/api.js';

	let record = $state<any>(null);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const id = $page.params.id;
		if (!id) {
			error = 'ID absensi tidak valid';
			loading = false;
			return;
		}
		try {
			const res = await attendance.getRecord(id);
			if (res?.success) {
				record = res.data;
			} else {
				error = res?.error || 'Gagal memuat detail absensi';
			}
		} catch (e: any) {
			error = e?.message || 'Gagal memuat data';
		} finally {
			loading = false;
		}
	});

	function getPhotoUrl(url: string | null | undefined): string {
		if (!url) return '';
		if (url.startsWith('http')) return url;
		return `${config.API_BASE_URL}${url}`;
	}

	function formatTime(d: string | null | undefined) {
		if (!d) return '-';
		return new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' });
	}

	function getDayName(d: string) {
		return new Date(d).toLocaleDateString('id-ID', { weekday: 'long' });
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			present: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			late: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			hadir: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			terlambat: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			absent: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			izin: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
			sakit: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
			tanpa_keterangan: 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300',
			cuti: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
			libur: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
		};
		return map[status] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300';
	}

	function getStatusText(status: string) {
		const map: Record<string, string> = {
			present: 'Hadir',
			late: 'Terlambat',
			hadir: 'Hadir',
			terlambat: 'Terlambat',
			absent: 'Absen',
			izin: 'Izin',
			sakit: 'Sakit',
			tanpa_keterangan: 'Tanpa Keterangan',
			cuti: 'Cuti',
			libur: 'Libur',
		};
		return map[status] || status;
	}



	function formatDuration(hours: number | null | undefined): string {
		if (!hours) return '-';
		const h = Math.floor(hours);
		const m = Math.round((hours - h) * 60);
		if (h > 0) return `${h}j ${m}m`;
		return `${m}m`;
	}

	function goToCorrection() {
		if (!record) return;
		const datePart = record.date.split('T')[0];
		let checkIn = '';
		if (record.check_in_time) {
			const d = new Date(record.check_in_time);
			checkIn = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
		}
		let checkOut = '';
		if (record.check_out_time) {
			const d = new Date(record.check_out_time);
			checkOut = `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`;
		}
		goto(`/absensi-manual?action=create&date=${datePart}&check_in=${checkIn}&check_out=${checkOut}`);
	}

	function getHour(d: string | null | undefined): number | null {
		if (!d) return null;
		return new Date(d).getHours();
	}

	function isLate(record: any): boolean {
		if (!record) return false;
		if (record.is_late === true || record.status === 'terlambat') return true;
		if (!record.check_in_time) return false;
		const d = new Date(record.check_in_time);
		const hour = d.getHours();
		const min = d.getMinutes();
		if (hour > 8 || (hour === 8 && min > 0)) return true;
		return false;
	}

	function isEarlyLeave(record: any): boolean {
		if (!record) return false;
		if (record.is_early_leave === true) return true;
		if (!record.check_out_time) return false;
		const d = new Date(record.check_out_time);
		const hour = d.getHours();
		if (hour < 17) return true;
		return false;
	}
</script>

<svelte:head>
	<title>{record ? `Detail Absensi - ${formatDate(record.date)}` : 'Detail Absensi'} - HRMS</title>
</svelte:head>

<div class="w-full animate-in fade-in slide-in-from-bottom-4 duration-500">
	<!-- Back button -->
	<div class="flex items-center gap-3 mb-6 px-4 sm:px-6 lg:px-8">
		<button onclick={() => goto('/absensi')}
			class="p-2 -ml-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors cursor-pointer">
			<svg class="w-5 h-5 text-gray-600 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18" />
			</svg>
		</button>
		<h1 class="text-xl font-bold text-gray-900 dark:text-white tracking-tight">Detail Absensi</h1>
	</div>

	{#if loading}
		<div class="flex items-center justify-center py-20">
			<div class="animate-spin h-8 w-8 border-2 border-blue-600 border-t-transparent rounded-full"></div>
		</div>
	{:else if error}
		<div class="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-4">
			<p class="text-sm font-medium text-red-800 dark:text-red-200">{error}</p>
		</div>
		<div class="mt-4 text-center">
			<button onclick={() => goto('/absensi')}
				class="px-5 py-2.5 bg-blue-600 text-white rounded-xl text-sm font-semibold hover:bg-blue-700 transition cursor-pointer">
				Kembali ke Absensi
			</button>
		</div>
	{:else if record}
		<!-- Header Card -->
		<div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden mb-5">
			<div class="p-5 sm:p-6">
				<div class="flex items-start justify-between mb-4">
					<div>
						<div class="text-lg font-bold text-gray-900 dark:text-white">{formatDate(record.date)}</div>
						<div class="text-sm text-gray-500 dark:text-gray-400">{getDayName(record.date)}</div>
					</div>
					<div class="flex items-center gap-2">
						<span class="px-3 py-1 rounded-full text-xs font-semibold {getStatusBadge(record.status)}">
							{getStatusText(record.status)}
						</span>
					</div>
				</div>

				<!-- Employee Info -->
				<div class="flex items-center gap-3 p-3 bg-gray-50 dark:bg-gray-900/50 rounded-xl">
					<div class="w-10 h-10 rounded-full bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400 font-bold text-sm shrink-0">
						{record.employee_name?.charAt(0) || '?'}
					</div>
					<div>
						<div class="text-sm font-semibold text-gray-900 dark:text-white">{record.employee_name || 'Karyawan'}</div>
						{#if record.department_name}
							<div class="text-xs text-gray-500 dark:text-gray-400">{record.department_name}</div>
						{/if}
					</div>
				</div>
			</div>

			<!-- Work Hours Summary -->
			<div class="grid grid-cols-3 divide-x divide-gray-100 dark:divide-gray-700 border-t border-gray-100 dark:border-gray-700">
				<div class="p-4 text-center">
					<div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Check In</div>
					<div class="text-sm font-bold {isLate(record) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(record.check_in_time)}</div>
				</div>
				<div class="p-4 text-center">
					<div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Check Out</div>
					<div class="text-sm font-bold {isEarlyLeave(record) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(record.check_out_time)}</div>
				</div>
				<div class="p-4 text-center">
					<div class="text-xs text-gray-500 dark:text-gray-400 mb-1">Total Jam</div>
					<div class="text-sm font-bold text-emerald-600 dark:text-emerald-400">{formatDuration(record.total_work_hours)}</div>
				</div>
			</div>
		</div>

		<!-- Warning Banner if Incomplete -->
		{#if !record.check_in_time || !record.check_out_time || isLate(record) || isEarlyLeave(record)}
			<div class="mx-4 sm:mx-6 lg:mx-8 {isLate(record) || isEarlyLeave(record) ? 'bg-red-50 dark:bg-red-950/20 border-red-200 dark:border-red-900' : 'bg-amber-50 dark:bg-amber-950/20 border-amber-200 dark:border-amber-900'} border rounded-2xl p-4 mb-5 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3 animate-in fade-in duration-300">
				<div class="flex items-start gap-3">
					<svg class="w-5 h-5 {isLate(record) || isEarlyLeave(record) ? 'text-red-600 dark:text-red-400' : 'text-amber-600 dark:text-amber-400'} mt-0.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
					</svg>
					<div>
						<h4 class="text-sm font-semibold {isLate(record) || isEarlyLeave(record) ? 'text-red-800 dark:text-red-200' : 'text-amber-800 dark:text-amber-200'}">
							{#if isLate(record) && isEarlyLeave(record)}
								Telat & Pulang Awal
							{:else if isLate(record)}
								Terlambat Masuk
							{:else if isEarlyLeave(record)}
								Pulang Lebih Awal
							{:else}
								Absensi Belum Lengkap
							{/if}
						</h4>
						<p class="text-xs {isLate(record) || isEarlyLeave(record) ? 'text-red-600 dark:text-red-400' : 'text-amber-600 dark:text-amber-400'} mt-0.5">
							{#if !record.check_in_time || !record.check_out_time}
								Ada data check-in atau check-out yang kosong untuk tanggal ini.
							{:else if isLate(record) && isEarlyLeave(record)}
								Anda terlambat masuk dan pulang lebih awal. Perbaiki data absensi Anda.
							{:else if isLate(record)}
								Anda terlambat masuk hari ini. Perbaiki data absensi Anda.
							{:else}
								Anda pulang lebih awal hari ini. Perbaiki data absensi Anda.
							{/if}
						</p>
					</div>
				</div>
				<button onclick={goToCorrection} class="px-4 py-2 {isLate(record) || isEarlyLeave(record) ? 'bg-red-600 hover:bg-red-700' : 'bg-amber-600 hover:bg-amber-700'} text-white rounded-xl text-xs font-semibold transition active:scale-[0.97] cursor-pointer shrink-0 shadow-sm">
					Perbaiki Absen
				</button>
			</div>
		{/if}

		<!-- Content in full-width responsive grid -->
		<div class="px-4 sm:px-6 lg:px-8 grid grid-cols-1 lg:grid-cols-2 gap-5 mb-5">
			<!-- Check In Section -->
			<div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
				<div class="px-5 py-4 bg-blue-50/50 dark:bg-blue-900/10 border-b border-gray-200 dark:border-gray-700">
					<div class="flex items-center gap-2.5">
						<div class="w-9 h-9 rounded-xl bg-blue-100 dark:bg-blue-900/30 flex items-center justify-center text-blue-600 dark:text-blue-400">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9"/>
							</svg>
						</div>
						<div>
							<h3 class="text-sm font-bold text-gray-900 dark:text-white">Check In</h3>
							<p class="text-xs text-gray-500 dark:text-gray-400">Absensi Masuk</p>
						</div>
						<div class="ml-auto text-right">
							<div class="text-lg font-bold {isLate(record) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(record.check_in_time)}</div>
							{#if isLate(record)}
								<div class="text-[10px] font-semibold text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 px-2 py-0.5 rounded-full inline-block mt-0.5">
									{#if record.late_minutes && record.late_minutes > 0}
										Telat {record.late_minutes} mnt
									{:else}
										Telat
									{/if}
								</div>
							{/if}
						</div>
					</div>
				</div>
				<div class="p-5 space-y-4">
					<!-- Photo & Map side by side -->
					<div class="grid gap-4 {record.check_in_photo_url && record.check_in_lat ? 'grid-cols-2' : 'grid-cols-1'}">
						{#if record.check_in_photo_url}
							<div class="space-y-1.5 flex flex-col">
								<p class="text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1.5">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z"/></svg>
									Foto Selfie
								</p>
								<div class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700 shadow-sm bg-gray-50 dark:bg-gray-900/50 flex-1 aspect-[4/3]">
									<img src={getPhotoUrl(record.check_in_photo_url)} alt="Foto Check In"
										class="w-full h-full object-cover hover:scale-105 transition-transform duration-500 cursor-pointer"
										onclick={() => window.open(getPhotoUrl(record.check_in_photo_url), '_blank')} />
								</div>
							</div>
						{/if}

						{#if record.check_in_lat && record.check_in_lng}
							<div class="space-y-1.5 flex flex-col">
								<p class="text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1.5 truncate">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"/></svg>
									Lokasi: {record.check_in_location_name || 'Koordinat GPS'}
								</p>
								<div class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700 shadow-sm flex-1 aspect-[4/3] min-h-[220px]">
									<iframe title="Lokasi Check In" width="100%" height="100%" frameborder="0" style="border:0"
										src={`https://maps.google.com/maps?q=${record.check_in_lat},${record.check_in_lng}&t=&z=15&ie=UTF8&iwloc=&output=embed`}
										allowfullscreen loading="lazy"></iframe>
								</div>
							</div>
						{/if}
					</div>

					<!-- Warning Banners (Below Photo & Map) -->
					{#if isLate(record) || record.check_in_location_name === 'Luar area absensi'}
						<div class="space-y-2">
							{#if isLate(record)}
								<div class="flex items-center gap-2.5 p-3 bg-rose-50 dark:bg-rose-950/20 text-rose-700 dark:text-rose-350 rounded-xl border border-rose-100 dark:border-rose-900/50 text-xs font-semibold">
									<svg class="w-4 h-4 shrink-0 text-rose-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
									</svg>
									<span>Terlambat masuk.
										{#if record.late_minutes && record.late_minutes > 0}
											Selama {record.late_minutes} menit.
										{/if}
									</span>
								</div>
							{/if}
							{#if record.check_in_location_name === 'Luar area absensi'}
								<div class="flex items-center gap-2.5 p-3 bg-amber-50 dark:bg-amber-950/20 text-amber-700 dark:text-amber-350 rounded-xl border border-amber-100 dark:border-amber-900/50 text-xs font-semibold">
									<svg class="w-4 h-4 shrink-0 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
									</svg>
									<span>Absen dilakukan di luar koordinat radius lokasi kantor.</span>
								</div>
							{/if}
						</div>
					{/if}

					{#if !record.check_in_photo_url && !record.check_in_location_name}
						<div class="text-center py-6 text-sm text-gray-400 col-span-full">
							<svg class="w-10 h-10 mx-auto mb-2 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"/></svg>
							Tidak ada data foto atau lokasi untuk check-in
						</div>
					{/if}
				</div>
			</div>

			<!-- Check Out Section -->
			<div class="bg-white dark:bg-gray-800 rounded-2xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
				<div class="px-5 py-4 bg-amber-50/50 dark:bg-amber-900/10 border-b border-gray-200 dark:border-gray-700">
					<div class="flex items-center gap-2.5">
						<div class="w-9 h-9 rounded-xl bg-amber-100 dark:bg-amber-900/30 flex items-center justify-center text-amber-600 dark:text-amber-400">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15"/>
							</svg>
						</div>
						<div>
							<h3 class="text-sm font-bold text-gray-900 dark:text-white">Check Out</h3>
							<p class="text-xs text-gray-500 dark:text-gray-400">Absensi Pulang</p>
						</div>
						<div class="ml-auto text-right">
							<div class="text-lg font-bold {isEarlyLeave(record) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(record.check_out_time)}</div>
							{#if isEarlyLeave(record)}
								<div class="text-[10px] font-semibold text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 px-2 py-0.5 rounded-full inline-block mt-0.5">Pulang Cepat</div>
							{/if}
						</div>
					</div>
				</div>
				<div class="p-5 space-y-4">
					<!-- Photo & Map side by side -->
					<div class="grid gap-4 {record.check_out_photo_url && record.check_out_lat ? 'grid-cols-2' : 'grid-cols-1'}">
						{#if record.check_out_photo_url}
							<div class="space-y-1.5 flex flex-col">
								<p class="text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1.5">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z"/></svg>
									Foto Selfie
								</p>
								<div class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700 shadow-sm bg-gray-50 dark:bg-gray-900/50 flex-1 aspect-[4/3]">
									<img src={getPhotoUrl(record.check_out_photo_url)} alt="Foto Check Out"
										class="w-full h-full object-cover hover:scale-105 transition-transform duration-500 cursor-pointer"
										onclick={() => window.open(getPhotoUrl(record.check_out_photo_url), '_blank')} />
								</div>
							</div>
						{/if}

						{#if record.check_out_lat && record.check_out_lng}
							<div class="space-y-1.5 flex flex-col">
								<p class="text-xs font-semibold text-gray-500 uppercase tracking-wider flex items-center gap-1.5 truncate">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"/></svg>
									Lokasi: {record.check_out_location_name || 'Koordinat GPS'}
								</p>
								<div class="rounded-xl overflow-hidden border border-gray-200 dark:border-gray-700 shadow-sm flex-1 aspect-[4/3] min-h-[220px]">
									<iframe title="Lokasi Check Out" width="100%" height="100%" frameborder="0" style="border:0"
										src={`https://maps.google.com/maps?q=${record.check_out_lat},${record.check_out_lng}&t=&z=15&ie=UTF8&iwloc=&output=embed`}
										allowfullscreen loading="lazy"></iframe>
								</div>
							</div>
						{/if}
					</div>

					<!-- Warning Banners (Below Photo & Map) -->
					{#if isEarlyLeave(record) || record.check_out_location_name === 'Luar area absensi'}
						<div class="space-y-2">
							{#if isEarlyLeave(record)}
								<div class="flex items-center gap-2.5 p-3 bg-rose-50 dark:bg-rose-950/20 text-rose-700 dark:text-rose-350 rounded-xl border border-rose-100 dark:border-rose-900/50 text-xs font-semibold">
									<svg class="w-4 h-4 shrink-0 text-rose-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" />
									</svg>
									<span>Pulang lebih awal.
										{#if record.early_leave_minutes && record.early_leave_minutes > 0}
											Selama {record.early_leave_minutes} menit.
										{/if}
									</span>
								</div>
							{/if}
							{#if record.check_out_location_name === 'Luar area absensi'}
								<div class="flex items-center gap-2.5 p-3 bg-amber-50 dark:bg-amber-950/20 text-amber-700 dark:text-amber-350 rounded-xl border border-amber-100 dark:border-amber-900/50 text-xs font-semibold">
									<svg class="w-4 h-4 shrink-0 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" />
									</svg>
									<span>Absen dilakukan di luar koordinat radius lokasi kantor.</span>
								</div>
							{/if}
						</div>
					{/if}

					{#if !record.check_out_photo_url && !record.check_out_location_name && record.check_out_time}
						<div class="text-center py-6 text-sm text-gray-400 col-span-full">
							<svg class="w-10 h-10 mx-auto mb-2 text-gray-300 dark:text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5"/></svg>
							Tidak ada data foto atau lokasi untuk check-out
						</div>
					{/if}
				</div>
			</div>
		</div>

		<!-- Manual Entry Info -->
		{#if record.is_manual_entry}
			<div class="mx-4 sm:mx-6 lg:mx-8 bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl p-4 mb-5">
				<div class="flex items-center gap-2.5">
					<svg class="w-5 h-5 text-amber-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"/></svg>
					<p class="text-sm text-amber-800 dark:text-amber-200"><span class="font-semibold">Entry Manual</span> — Data ini diinput oleh HR/Admin</p>
				</div>
			</div>
		{/if}

		<!-- Bottom Navigation -->
		<div class="px-4 sm:px-6 lg:px-8 pb-8">
			<button onclick={() => goto('/absensi')}
				class="w-full sm:w-auto inline-flex items-center justify-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-xl text-sm font-semibold hover:bg-blue-700 transition cursor-pointer active:scale-[0.98] shadow-lg shadow-blue-500/20">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M10.5 19.5 3 12m0 0 7.5-7.5M3 12h18"/></svg>
				Kembali ke Absensi
			</button>
		</div>
	{/if}
</div>
