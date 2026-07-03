<script lang="ts">
	import { onMount } from 'svelte';
	import { attendance } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';

	interface TodayStatus {
		has_checked_in: boolean;
		has_checked_out: boolean;
		schedule_name: string;
		schedule_start: string;
		schedule_end: string;
		record?: {
			id: string;
			check_in_time: string | null;
			check_out_time: string | null;
			status: string;
			is_late: boolean;
			late_minutes: number;
			total_work_hours: number | null;
			check_in_location_name: string | null;
		};
	}

	interface HistoryRecord {
		id: string;
		date: string;
		day_name: string;
		check_in_time: string | null;
		check_out_time: string | null;
		status: string;
		is_late: boolean;
		late_minutes: number;
		total_work_hours: number | null;
		check_in_location_name: string | null;
		check_in_photo_url?: string | null;
		check_in_lat?: number | null;
		check_in_lng?: number | null;
		check_out_location_name?: string | null;
		check_out_photo_url?: string | null;
		check_out_lat?: number | null;
		check_out_lng?: number | null;
	}

	let status = $state<TodayStatus | null>(null);
	let history = $state<HistoryRecord[]>([]);
	let loading = $state(true);
	let checkingIn = $state(false);
	let checkingOut = $state(false);
	let error = $state('');
	let gpsStatus = $state('');
	let gpsCoords = $state<{ lat: number; lng: number } | null>(null);
	let isExporting = $state(false);
	
	let expandedRecordId = $state<string | null>(null);
	
	function toggleExpand(id: string) {
		expandedRecordId = expandedRecordId === id ? null : id;
	}

	async function handleExportReport() {
		isExporting = true;
		try {
			const blob = await attendance.exportReport();
			const url = URL.createObjectURL(blob);
			const a = document.createElement('a');
			a.href = url;
			a.download = 'laporan-absensi.xlsx';
			document.body.appendChild(a);
			a.click();
			document.body.removeChild(a);
			URL.revokeObjectURL(url);
		} catch (err: any) {
			console.error('Export error:', err);
		} finally {
			isExporting = false;
		}
	}

	onMount(() => {
		loadTodayStatus();
		loadHistory();
	});

	async function loadTodayStatus() {
		try {
			const res = await attendance.getTodayStatus();
			if (res.success) {
				status = res.data;
			}
		} catch (e: any) {
			error = 'Gagal memuat status absensi';
		} finally {
			loading = false;
		}
	}

	async function loadHistory() {
		try {
			const res = await attendance.myHistory(1, 5);
			if (res.success) {
				history = res.data || [];
			}
		} catch (e: any) {
			// silently fail
		}
	}

	function getGPS(): Promise<{ lat: number; lng: number }> {
		return new Promise((resolve, reject) => {
			if (!navigator.geolocation) {
				reject(new Error('Geolocation tidak didukung browser ini'));
				return;
			}
			gpsStatus = 'Mendapatkan lokasi...';
			navigator.geolocation.getCurrentPosition(
				(pos) => {
					const coords = { lat: pos.coords.latitude, lng: pos.coords.longitude };
					gpsCoords = coords;
					gpsStatus = `Lokasi: ${coords.lat.toFixed(4)}, ${coords.lng.toFixed(4)}`;
					resolve(coords);
				},
				(err) => {
					gpsStatus = 'Gagal mendapatkan lokasi';
					reject(err);
				},
				{ enableHighAccuracy: true, timeout: 10000 }
			);
		});
	}

	async function handleCheckIn() {
		checkingIn = true;
		error = '';
		try {
			const coords = await getGPS();
			const res = await attendance.checkIn({
				lat: coords.lat,
				lng: coords.lng,
			});
			if (res.success) {
				await loadTodayStatus();
			}
		} catch (e: any) {
			error = e.message || 'Gagal check-in';
		} finally {
			checkingIn = false;
		}
	}

	async function handleCheckOut() {
		checkingOut = true;
		error = '';
		try {
			const coords = await getGPS();
			const res = await attendance.checkOut({
				lat: coords.lat,
				lng: coords.lng,
			});
			if (res.success) {
				await loadTodayStatus();
				await loadHistory();
			}
		} catch (e: any) {
			error = e.message || 'Gagal check-out';
		} finally {
			checkingOut = false;
		}
	}

	function formatTime(iso: string | null): string {
		if (!iso) return '—';
		return new Date(iso).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function statusBadge(s: string): string {
		const map: Record<string, string> = {
			hadir: 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200',
			terlambat: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200',
			izin: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
			sakit: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
			tanpa_keterangan: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
		};
		return map[s] || 'bg-gray-100 text-gray-800';
	}
</script>

<div class="max-w-full space-y-6">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<h1 class="text-2xl font-bold text-gray-900 dark:text-white">Absensi</h1>
		{#if hasPermission('attendance', 'read')}
			<button onclick={handleExportReport} disabled={isExporting}
				class="inline-flex items-center gap-2 px-4 py-2.5 border border-gray-200 dark:border-gray-600 text-gray-700 dark:text-gray-300 rounded-xl text-sm font-semibold hover:bg-gray-50 dark:hover:bg-gray-700 transition shadow-sm cursor-pointer disabled:opacity-50">
				{#if isExporting}
					<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
				{:else}
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3 16.5v2.25A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75V16.5M16.5 12 12 16.5m0 0L7.5 12m4.5 4.5V3" /></svg>
				{/if}
				{isExporting ? 'Mengexport...' : 'Export Excel'}
			</button>
		{/if}
	</div>

	{#if loading}
		<div class="text-center py-12 text-gray-500 dark:text-gray-400">Memuat...</div>
	{:else if error}
		<div class="bg-red-50 dark:bg-red-900/30 border border-red-200 dark:border-red-800 text-red-700 dark:text-red-300 px-4 py-3 rounded-lg">{error}</div>
	{/if}

	<!-- Today's Status Card -->
	<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-gray-900 dark:text-white">Hari Ini</h2>
			<span class="text-sm text-gray-500 dark:text-gray-400">
				{new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
			</span>
		</div>

		{#if status}
			<div class="space-y-4">
				<!-- Schedule Info -->
				<div class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
					</svg>
					<span>{status.schedule_name || 'Tidak ada jadwal'}</span>
					{#if status.schedule_start}
						<span class="text-gray-400">|</span>
						<span>{status.schedule_start} - {status.schedule_end}</span>
					{/if}
				</div>

				<!-- Check In Status -->
				<div class="grid grid-cols-2 gap-4">
					<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4 text-center">
						<div class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">Check In</div>
						<div class="text-xl font-bold" class:text-green-600={status.has_checked_in} class:text-gray-400={!status.has_checked_in}>
							{status.has_checked_in ? formatTime(status.record?.check_in_time ?? null) : '—'}
						</div>
						{#if status.record?.is_late}
							<div class="text-xs text-yellow-600 dark:text-yellow-400 mt-1">Terlambat {status.record.late_minutes} menit</div>
						{/if}
						{#if status.record?.check_in_location_name}
							<div class="text-xs text-gray-400 mt-1">{status.record.check_in_location_name}</div>
						{/if}
					</div>
					<div class="bg-gray-50 dark:bg-gray-700/50 rounded-lg p-4 text-center">
						<div class="text-xs text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-1">Check Out</div>
						<div class="text-xl font-bold" class:text-red-600={status.has_checked_out} class:text-gray-400={!status.has_checked_out}>
							{status.has_checked_out ? formatTime(status.record?.check_out_time ?? null) : '—'}
						</div>
						{#if status.record?.total_work_hours}
							<div class="text-xs text-gray-500 dark:text-gray-400 mt-1">{status.record.total_work_hours.toFixed(1)} jam</div>
						{/if}
					</div>
				</div>

				<!-- Action Buttons -->
				<div class="flex gap-3">
					{#if !status.has_checked_in}
						<button
							onclick={handleCheckIn}
							disabled={checkingIn}
							class="flex-1 bg-blue-600 hover:bg-blue-700 disabled:bg-blue-400 text-white font-semibold py-3 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z" />
							</svg>
							{checkingIn ? 'Memproses...' : 'Check In'}
						</button>
					{:else if !status.has_checked_out}
						<button
							onclick={handleCheckOut}
							disabled={checkingOut}
							class="flex-1 bg-orange-500 hover:bg-orange-600 disabled:bg-orange-300 text-white font-semibold py-3 px-6 rounded-lg transition-colors flex items-center justify-center gap-2"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14l-4.553 2.276A1 1 0 013 15.382V8.618a1 1 0 011.447-.894L9 10m0 0V7a2 2 0 012-2h6a2 2 0 012 2v10a2 2 0 01-2 2h-6a2 2 0 01-2-2v-3z" />
							</svg>
							{checkingOut ? 'Memproses...' : 'Check Out'}
						</button>
					{:else}
						<div class="flex-1 bg-green-50 dark:bg-green-900/30 text-green-700 dark:text-green-300 font-semibold py-3 px-6 rounded-lg text-center">
							Selesai — {status.record?.total_work_hours?.toFixed(1) ?? '—'} jam
						</div>
					{/if}
				</div>

				<!-- GPS Status -->
				{#if gpsStatus}
					<div class="text-xs text-gray-400 dark:text-gray-500 text-center">{gpsStatus}</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Recent History -->
	<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Riwayat Terbaru</h2>

		{#if history.length === 0}
			<p class="text-gray-500 dark:text-gray-400 text-sm">Belum ada riwayat absensi</p>
		{:else}
			<div class="space-y-2">
				{#each history as r}
					<div class="border-b border-gray-100 dark:border-gray-700 last:border-0">
						<!-- svelte-ignore a11y_click_events_have_key_events -->
						<!-- svelte-ignore a11y_no_static_element_interactions -->
						<div 
							onclick={() => toggleExpand(r.id)}
							class="flex items-center justify-between py-3 cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg px-2 -mx-2 transition-colors"
						>
							<div>
								<div class="text-sm font-medium text-gray-900 dark:text-white flex items-center gap-2">
									{formatDate(r.date)}
									<svg class="w-4 h-4 text-gray-400 transition-transform {expandedRecordId === r.id ? "rotate-180" : ""}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
								</div>
								<div class="text-xs text-gray-500 dark:text-gray-400">{r.day_name}</div>
							</div>
							<div class="text-right">
								<div class="text-sm text-gray-700 dark:text-gray-300">
									{formatTime(r.check_in_time)} — {formatTime(r.check_out_time)}
								</div>
								<div class="flex items-center gap-2 justify-end mt-1">
									<span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium {statusBadge(r.status)}">
										{r.status}
									</span>
									{#if r.total_work_hours}
										<span class="text-xs text-gray-400">{r.total_work_hours.toFixed(1)}j</span>
									{/if}
								</div>
							</div>
						</div>
						
						{#if expandedRecordId === r.id}
							<div class="px-2 pb-4 pt-1 animate-in slide-in-from-top-2 duration-200">
								<div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-gray-50 dark:bg-gray-900/50 p-4 rounded-xl border border-gray-100 dark:border-gray-800">
									<!-- Check In Detail -->
									<div class="space-y-3">
										<h4 class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2 text-sm">
											<svg class="w-4 h-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"/></svg>
											Check In: <span class="text-gray-900 dark:text-white">{formatTime(r.check_in_time)}</span>
										</h4>
										{#if r.check_in_photo_url}
											<div class="aspect-video bg-gray-200 dark:bg-gray-800 rounded-lg overflow-hidden relative border border-gray-200 dark:border-gray-700">
												<img src={r.check_in_photo_url} alt="Check In" class="w-full h-full object-cover" />
											</div>
										{/if}
										<div class="text-xs">
											<div class="text-gray-500 dark:text-gray-400 mb-0.5">Lokasi</div>
											<div class="font-medium text-gray-900 dark:text-white">{r.check_in_location_name || "Tidak diketahui"}</div>
											{#if r.check_in_lat && r.check_in_lng}
												<a href={`https://www.google.com/maps?q=${r.check_in_lat},${r.check_in_lng}`} target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline mt-1 inline-flex items-center gap-1">
													<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
													Buka di Maps
												</a>
											{/if}
										</div>
									</div>

									<!-- Check Out Detail -->
									<div class="space-y-3 border-t md:border-t-0 md:border-l border-gray-200 dark:border-gray-700 pt-4 md:pt-0 md:pl-4">
										<h4 class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2 text-sm">
											<svg class="w-4 h-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
											Check Out: <span class="text-gray-900 dark:text-white">{formatTime(r.check_out_time)}</span>
										</h4>
										{#if r.check_out_time}
											{#if r.check_out_photo_url}
												<div class="aspect-video bg-gray-200 dark:bg-gray-800 rounded-lg overflow-hidden relative border border-gray-200 dark:border-gray-700">
													<img src={r.check_out_photo_url} alt="Check Out" class="w-full h-full object-cover" />
												</div>
											{/if}
											<div class="text-xs">
												<div class="text-gray-500 dark:text-gray-400 mb-0.5">Lokasi</div>
												<div class="font-medium text-gray-900 dark:text-white">{r.check_out_location_name || "Tidak diketahui"}</div>
												{#if r.check_out_lat && r.check_out_lng}
													<a href={`https://www.google.com/maps?q=${r.check_out_lat},${r.check_out_lng}`} target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline mt-1 inline-flex items-center gap-1">
														<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/></svg>
														Buka di Maps
													</a>
												{/if}
											</div>
										{:else}
											<div class="text-xs text-gray-500 italic mt-2">Belum melakukan Check Out</div>
										{/if}
									</div>
								</div>
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
