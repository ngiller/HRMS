<script lang="ts">
	import { onMount, onDestroy, tick } from 'svelte';
	import { attendance } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import config from '$lib/config.js';

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
			check_in_photo_url?: string | null;
			check_out_photo_url?: string | null;
			check_out_location_name?: string | null;
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
		is_early_leave?: boolean;
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

	// Camera Face Attendance
	let videoElement = $state<HTMLVideoElement | undefined>();
	let stream = $state<MediaStream | null>(null);
	let showCamera = $state(false);

	// Face Detection (face-api.js)
	let faceDetector = $state<any>(null);
	let faceDetected = $state(false);
	let faceDetectionLoading = $state(false);
	let faceDetectionCanvas = $state<HTMLCanvasElement | undefined>();
	let faceDetectionInterval: number | null = null;
	let isFaceApiLoaded = $state(false);

	// Mobile-specific state
	let currentTime = $state(new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' }));
	
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
		} catch (err: unknown) {
			console.error('Export error:', err);
		} finally {
			isExporting = false;
		}
	}

	onMount(() => {
		loadTodayStatus();
		loadHistory();
		// Real-time clock update every 30s
		const interval = setInterval(() => {
			currentTime = new Date().toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
		}, 30000);
		return () => clearInterval(interval);
	});

	async function loadTodayStatus() {
		try {
			const res = await attendance.getTodayStatus();
			if (res.success) {
				status = res.data;
			}
		} catch (_e: unknown) {
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
		} catch (_e: unknown) {
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

	async function loadFaceDetectionScript() {
		if ((window as any).faceapi) {
			isFaceApiLoaded = true;
			return;
		}
		return new Promise<void>((resolve, reject) => {
			const script = document.createElement('script');
			script.src = 'https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/dist/face-api.min.js';
			script.onload = () => {
				isFaceApiLoaded = true;
				resolve();
			};
			script.onerror = () => reject(new Error('Gagal memuat face-api.js'));
			document.head.appendChild(script);
		});
	}

	async function startFaceDetection() {
		if (!videoElement || !(window as any).faceapi) return;
		const faceapi = (window as any).faceapi;			try {
				faceDetectionLoading = true;
				await faceapi.nets.tinyFaceDetector.loadFromUri('https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/weights');
				await faceapi.nets.faceLandmark68Net.loadFromUri('https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/weights');
				// Pre-load face recognition model for descriptor computation during check-in
				await faceapi.nets.faceRecognitionNet.loadFromUri('https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/weights');
				faceDetectionLoading = false;

			const canvas = faceDetectionCanvas;
			if (!canvas) return;

			const displaySize = { width: videoElement.videoWidth || 320, height: videoElement.videoHeight || 400 };
			faceapi.matchDimensions(canvas, displaySize);

			faceDetectionInterval = window.setInterval(async () => {
				if (!videoElement || !videoElement.videoWidth) return;
				
				try {
					const detections = await faceapi
						.detectAllFaces(videoElement, new faceapi.TinyFaceDetectorOptions({ inputSize: 320, scoreThreshold: 0.5 }))
						.withFaceLandmarks();

					const ctx = canvas.getContext('2d');
					if (ctx) {
						ctx.clearRect(0, 0, canvas.width, canvas.height);
					}

					if (detections && detections.length > 0) {
						faceDetected = true;
						const resized = faceapi.resizeResults(detections, displaySize);
						if (ctx) {
							// Draw face bounding box
							faceapi.draw.drawDetections(canvas, resized);
							// Draw face landmarks
							faceapi.draw.drawFaceLandmarks(canvas, resized);
						}
					} else {
						faceDetected = false;
					}
				} catch {
					// Silently continue if detection fails on this frame
				}
			}, 200);
		} catch (err) {
			console.error('Face detection init error:', err);
			faceDetectionLoading = false;
		}
	}

	function stopFaceDetection() {
		if (faceDetectionInterval) {
			clearInterval(faceDetectionInterval);
			faceDetectionInterval = null;
		}
		faceDetected = false;
		// Clear canvas
		const canvas = faceDetectionCanvas;
		if (canvas) {
			const ctx = canvas.getContext('2d');
			if (ctx) ctx.clearRect(0, 0, canvas.width, canvas.height);
		}
	}

	async function startCamera() {
		try {
			stream = await navigator.mediaDevices.getUserMedia({ video: { facingMode: 'user', width: { ideal: 320 }, height: { ideal: 400 } } });
			showCamera = true;
			await tick();
			if (videoElement) {
				videoElement.srcObject = stream;
			}
			// Load face-api.js and start face detection
			if (!isFaceApiLoaded) {
				await loadFaceDetectionScript();
			}
			// Wait a bit for video to be ready
			setTimeout(async () => {
				if (videoElement && videoElement.videoWidth > 0) {
					await startFaceDetection();
				}
			}, 500);
		} catch (err) {
			error = 'Gagal mengakses kamera. Pastikan izin kamera diberikan.';
		}
	}

	function stopCamera() {
		stopFaceDetection();
		if (stream) {
			stream.getTracks().forEach(track => track.stop());
			stream = null;
		}
		showCamera = false;
	}

	function capturePhoto(): string {
		if (!videoElement) return '';
		const canvas = document.createElement('canvas');
		canvas.width = videoElement.videoWidth;
		canvas.height = videoElement.videoHeight;
		const ctx = canvas.getContext('2d');
		if (ctx) {
			// Mirror the canvas to match the mirrored video
			ctx.translate(canvas.width, 0);
			ctx.scale(-1, 1);
			ctx.drawImage(videoElement, 0, 0, canvas.width, canvas.height);
			return canvas.toDataURL('image/jpeg', 0.8);
		}
		return '';
	}

	onDestroy(() => {
		stopCamera();
	});

	function isBlockedByFaceDetection(): boolean {
		return showCamera && !faceDetectionLoading && !faceDetected;
	}

	async function computeFaceDescriptor(photoDataUrl: string): Promise<number[] | null> {
		const faceapi = (window as any).faceapi;
		if (!faceapi) return null;

		try {
			// Load face recognition model if not loaded
			if (!faceapi.nets.faceRecognitionNet) {
				await faceapi.nets.faceRecognitionNet.loadFromUri('https://cdn.jsdelivr.net/npm/face-api.js@0.22.2/weights');
			}

			const img = new Image();
			img.src = photoDataUrl;
			await new Promise((resolve) => { img.onload = resolve; });

			const result = await faceapi
				.detectSingleFace(img, new faceapi.TinyFaceDetectorOptions({ inputSize: 320, scoreThreshold: 0.5 }))
				.withFaceLandmarks()
				.withFaceDescriptor();

			if (result && result.descriptor) {
				return Array.from(result.descriptor as Float32Array);
			}
			return null;
		} catch (err) {
			console.error('Face descriptor error:', err);
			return null;
		}
	}

	async function handleCheckIn() {
		if (!showCamera) {
			await startCamera();
			return;
		}

		// Block if face not detected
		if (isBlockedByFaceDetection()) {
			error = 'Wajah tidak terdeteksi. Silakan posisikan wajah Anda di depan kamera.';
			return;
		}

		checkingIn = true;
		error = '';
		try {
			const photo = capturePhoto();
			
			// Compute face descriptor for verification
			const descriptor = await computeFaceDescriptor(photo);
			
			const coords = await getGPS();
			const res = await attendance.checkIn({
				lat: coords.lat,
				lng: coords.lng,
				photo: photo,
				face_descriptor: descriptor ? JSON.stringify(descriptor) : undefined
			});
			if (res.success) {
				await loadTodayStatus();
				stopCamera();
			}
		} catch (e: unknown) {
			error = (e as { message?: string }).message || 'Gagal check-in';
		} finally {
			checkingIn = false;
		}
	}

	async function handleCheckOut() {
		if (!showCamera) {
			await startCamera();
			return;
		}

		// Block if face not detected
		if (isBlockedByFaceDetection()) {
			error = 'Wajah tidak terdeteksi. Silakan posisikan wajah Anda di depan kamera.';
			return;
		}

		checkingOut = true;
		error = '';
		try {
			const photo = capturePhoto();
			
			// Compute face descriptor for verification
			const descriptor = await computeFaceDescriptor(photo);
			
			const coords = await getGPS();
			const res = await attendance.checkOut({
				lat: coords.lat,
				lng: coords.lng,
				photo: photo,
				face_descriptor: descriptor ? JSON.stringify(descriptor) : undefined
			});
			if (res.success) {
				await loadTodayStatus();
				await loadHistory();
				stopCamera();
			}
		} catch (e: unknown) {
			error = (e as { message?: string }).message || 'Gagal check-out';
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
			terlambat: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
			izin: 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200',
			sakit: 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200',
			tanpa_keterangan: 'bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200',
		};
		return map[s] || 'bg-gray-100 text-gray-800';
	}

	function getPhotoUrl(url: string | null | undefined): string {
		if (!url) return '';
		if (url.startsWith('http')) return url;
		return config.API_BASE_URL + url;
	}

	function isEarlyCheckOut(statusObj: any): boolean {
		if (!statusObj?.has_checked_out || !statusObj?.record?.check_out_time || !statusObj?.schedule_end) return false;
		const [h, m] = statusObj.schedule_end.split(':').map(Number);
		if (isNaN(h) || isNaN(m)) return false;
		const coTime = new Date(statusObj.record.check_out_time);
		const scheduleEndTime = new Date(coTime);
		scheduleEndTime.setHours(h, m, 0, 0);
		return coTime < scheduleEndTime;
	}

	function isLateCheckIn(statusObj: any): boolean {
		if (!statusObj?.has_checked_in || !statusObj?.record?.check_in_time || !statusObj?.schedule_start) return false;
		const [h, m] = statusObj.schedule_start.split(':').map(Number);
		if (isNaN(h) || isNaN(m)) return false;
		const ciTime = new Date(statusObj.record.check_in_time);
		const scheduleStartTime = new Date(ciTime);
		scheduleStartTime.setHours(h, m, 0, 0);
		return ciTime > scheduleStartTime;
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

	<!-- Today's Status Card -- Talenta Style -->
	<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 overflow-hidden">
		<!-- Header dengan jam real-time -->
		<div class="bg-gradient-to-r from-[#1A56DB] to-[#1e40af] px-5 py-4">
			<div class="flex items-center justify-between text-white">
				<div>
					<p class="text-xs text-blue-200 font-medium">Hari Ini</p>
					<p class="text-sm font-semibold mt-0.5">
						{new Date().toLocaleDateString('id-ID', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
					</p>
				</div>
				<div class="text-right">
					<p class="text-2xl font-bold tabular-nums tracking-tight">{currentTime}</p>
				</div>
			</div>
			{#if status}
				<div class="flex items-center gap-2 mt-3 text-xs text-blue-100">
					<svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/></svg>
					<span>{status.schedule_name || 'Tidak ada jadwal'}</span>
					{#if status.schedule_start}
						<span class="opacity-50">|</span>
						<span>{status.schedule_start} - {status.schedule_end}</span>
					{/if}
				</div>
			{/if}
		</div>

		{#if status}
			<div class="p-5 space-y-4">
				<!-- Check In/Out Status -- Talenta style -->
				<div class="flex items-center gap-3">
					<div class="flex-1 bg-gray-50 dark:bg-gray-700/50 rounded-xl p-4 text-center border border-gray-100 dark:border-gray-700">
						<div class="flex items-center justify-center gap-2 mb-2">
							<div class="w-2.5 h-2.5 rounded-full {status.has_checked_in ? (isLateCheckIn(status) ? 'bg-red-500' : 'bg-green-500') : 'bg-gray-300'}"></div>
							<span class="text-[10px] font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Check In</span>
						</div>
						<div class="text-2xl font-bold tabular-nums {status.has_checked_in ? (isLateCheckIn(status) ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400') : 'text-gray-300 dark:text-gray-600'}">
							{status.has_checked_in ? formatTime(status.record?.check_in_time ?? null) : '—'}
						</div>
						{#if status.has_checked_in && isLateCheckIn(status)}
							<div class="mt-1 inline-flex items-center gap-1 px-2 py-0.5 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-full text-[10px] font-medium">
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"/></svg>
								Datang terlambat
							</div>
						{/if}
						{#if status.record?.check_in_photo_url}
							<div class="mt-2 flex justify-center">
								<img src={getPhotoUrl(status.record.check_in_photo_url)} alt="Check In Photo" class="w-16 h-16 rounded-lg object-cover border border-gray-200 dark:border-gray-600 shadow-sm" />
							</div>
						{/if}
						{#if status.record?.check_in_location_name}
							<div class="text-[10px] text-gray-400 mt-1.5 flex items-center justify-center gap-1">
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"/></svg>
								{status.record.check_in_location_name}
							</div>
						{/if}
					</div>
					<div class="flex-1 bg-gray-50 dark:bg-gray-700/50 rounded-xl p-4 text-center border border-gray-100 dark:border-gray-700">
						<div class="flex items-center justify-center gap-2 mb-2">
							<div class="w-2.5 h-2.5 rounded-full {status.has_checked_out ? (isEarlyCheckOut(status) ? 'bg-red-500' : 'bg-green-500') : status.has_checked_in ? 'bg-amber-400' : 'bg-gray-300'}"></div>
							<span class="text-[10px] font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider">Check Out</span>
						</div>
						<div class="text-2xl font-bold tabular-nums {status.has_checked_out ? (isEarlyCheckOut(status) ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400') : status.has_checked_in ? 'text-amber-500' : 'text-gray-300 dark:text-gray-600'}">
							{status.has_checked_out ? formatTime(status.record?.check_out_time ?? null) : '—'}
						</div>
						{#if status.has_checked_out && isEarlyCheckOut(status)}
							<div class="mt-1 inline-flex items-center gap-1 px-2 py-0.5 bg-red-50 dark:bg-red-900/30 text-red-600 dark:text-red-400 rounded-full text-[10px] font-medium">
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"/></svg>
								Pulang lebih awal
							</div>
						{/if}
						{#if status.record?.total_work_hours}
							<div class="mt-1 inline-flex items-center gap-1 text-[10px] text-gray-500 dark:text-gray-400">
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
								{status.record.total_work_hours.toFixed(1)} jam
							</div>
						{/if}
						{#if status.record?.check_out_photo_url}
							<div class="mt-2 flex justify-center">
								<img src={getPhotoUrl(status.record.check_out_photo_url)} alt="Check Out Photo" class="w-16 h-16 rounded-lg object-cover border border-gray-200 dark:border-gray-600 shadow-sm" />
							</div>
						{/if}
						{#if status.record?.check_out_location_name}
							<div class="text-[10px] text-gray-400 mt-1.5 flex items-center justify-center gap-1">
								<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"/><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"/></svg>
								{status.record.check_out_location_name}
							</div>
						{/if}
					</div>
				</div>

				<!-- Face Attendance Camera -->
				{#if showCamera}
					<div class="relative rounded-2xl overflow-hidden bg-black aspect-[3/4] max-w-sm mx-auto shadow-inner border border-gray-200 dark:border-gray-700 animate-in fade-in duration-300">
						<!-- svelte-ignore a11y_media_has_caption -->
						<video
							bind:this={videoElement}
							autoplay
							playsinline
							class="w-full h-full object-cover transform scale-x-[-1]"
						></video>
						
						<!-- Face Detection Overlay Canvas (mirrored to match video) -->
						<canvas
							bind:this={faceDetectionCanvas}
							class="absolute inset-0 w-full h-full pointer-events-none transform scale-x-[-1]"
						></canvas>
						
						<!-- Face Detection Status -->
						<div class="absolute top-4 left-4">
							{#if faceDetectionLoading}
								<div class="flex items-center gap-2 bg-black/50 backdrop-blur-sm text-white/80 text-xs px-3 py-1.5 rounded-full">
									<svg class="w-3.5 h-3.5 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"/><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/></svg>
									Memuat detektor wajah...
								</div>
							{:else if faceDetected}
								<div class="flex items-center gap-1.5 bg-green-500/80 backdrop-blur-sm text-white text-xs px-3 py-1.5 rounded-full">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5"/></svg>
									Wajah terdeteksi
								</div>
							{:else}
								<div class="flex items-center gap-1.5 bg-red-500/80 backdrop-blur-sm text-white text-xs px-3 py-1.5 rounded-full">
									<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"/></svg>
									Wajah tidak terdeteksi
								</div>
							{/if}
						</div>
						
						<!-- Cancel button -->
						<button 
							onclick={(e) => { e.stopPropagation(); stopCamera(); }}
							class="absolute top-4 right-4 w-8 h-8 bg-black/50 text-white rounded-full flex items-center justify-center backdrop-blur-sm hover:bg-black/70 transition-colors"
						>
							<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12" /></svg>
						</button>
						
						<div class="absolute bottom-4 left-0 w-full text-center text-white text-xs font-medium drop-shadow-md tracking-wide uppercase">
							{faceDetected ? 'WAJAH TERDETEKSI ✓' : 'POSISIKAN WAJAH ANDA'}
						</div>
					</div>
				{/if}

				<!-- Action Buttons -- Premium -->
				<div class="flex gap-3">
					{#if !status.has_checked_in}
						<button
							onclick={handleCheckIn}
							disabled={checkingIn || isBlockedByFaceDetection()}
							class="flex-1 bg-gradient-to-r from-[#1A56DB] to-[#1e40af] hover:from-[#1e40af] hover:to-[#1e3a8a] disabled:opacity-60 text-white font-bold py-3.5 px-6 rounded-xl transition-all duration-200 flex items-center justify-center gap-2 shadow-lg shadow-blue-200/50 dark:shadow-blue-900/30 active:scale-[0.97] cursor-pointer"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 10l4.553-2.276A1 1 0 0121 8.618v6.764a1 1 0 01-1.447.894L15 14M5 18h8a2 2 0 002-2V8a2 2 0 00-2-2H5a2 2 0 00-2 2v8a2 2 0 002 2z"/></svg>
							{showCamera ? (checkingIn ? 'Memproses...' : (isBlockedByFaceDetection() ? 'Arahkan Wajah ke Kamera' : 'Ambil Foto & Check In')) : 'Check In'}
						</button>
					{:else if !status.has_checked_out}
						<button
							onclick={handleCheckOut}
							disabled={checkingOut || isBlockedByFaceDetection()}
							class="flex-1 bg-gradient-to-r from-orange-500 to-orange-600 hover:from-orange-600 hover:to-orange-700 disabled:opacity-60 text-white font-bold py-3.5 px-6 rounded-xl transition-all duration-200 flex items-center justify-center gap-2 shadow-lg shadow-orange-200/50 dark:shadow-orange-900/30 active:scale-[0.97] cursor-pointer"
						>
							<svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 14l-4.553 2.276A1 1 0 013 15.382V8.618a1 1 0 011.447-.894L9 10m0 0V7a2 2 0 012-2h6a2 2 0 012 2v10a2 2 0 01-2 2h-6a2 2 0 01-2-2v-3z"/></svg>
							{showCamera ? (checkingOut ? 'Memproses...' : (isBlockedByFaceDetection() ? 'Arahkan Wajah ke Kamera' : 'Ambil Foto & Check Out')) : 'Check Out'}
						</button>
					{:else}
						<div class="flex-1 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 text-green-700 dark:text-green-300 font-bold py-3.5 px-6 rounded-xl text-center flex items-center justify-center gap-2">
							<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>
							Selesai — {status.record?.total_work_hours?.toFixed(1) ?? '—'} jam
						</div>
					{/if}
				</div>

				<!-- GPS Status -->
				{#if gpsStatus}
					<div class="text-xs text-gray-400 dark:text-gray-500 text-center flex items-center justify-center gap-1.5">
						<svg class="w-3 h-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 6.75V15m6-6v8.25m.503 3.498 4.875-2.437c.381-.19.622-.58.622-1.006V4.82c0-.836-.88-1.38-1.628-1.006l-3.869 1.934c-.317.159-.69.159-1.006 0L9.503 3.252a1.125 1.125 0 0 0-1.006 0L3.622 5.689C3.24 5.88 3 6.27 3 6.695V19.18c0 .836.88 1.38 1.628 1.006l3.869-1.934c.317-.159.69-.159 1.006 0l4.994 2.497c.317.158.69.158 1.006 0Z"/></svg>
						{gpsStatus}
					</div>
				{/if}
			</div>
		{/if}
	</div>

	<!-- Recent History -->
	<PullToRefresh onRefresh={async () => { await Promise.all([loadTodayStatus(), loadHistory()]); }}>
	<div class="bg-white dark:bg-gray-800 rounded-xl shadow-sm border border-gray-200 dark:border-gray-700 p-6">
		<h2 class="text-lg font-semibold text-gray-900 dark:text-white mb-4">Riwayat Terbaru</h2>

		{#if history.length === 0}
			<p class="text-gray-500 dark:text-gray-400 text-sm">Belum ada riwayat absensi</p>
		{:else}
			<div class="space-y-2">
				{#each history as r}
					<!-- svelte-ignore a11y_click_events_have_key_events -->
					<!-- svelte-ignore a11y_no_static_element_interactions -->
					<div 
						onclick={() => toggleExpand(r.id)}
						class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 active:scale-[0.98] transition-all duration-150 shadow-sm cursor-pointer"
					>
						<div class="flex items-center justify-between mb-2">
							<div class="min-w-0 flex-1">
								<div class="text-sm font-semibold text-gray-900 dark:text-white flex items-center gap-2">
									{formatDate(r.date)}
									<svg class="w-4 h-4 text-gray-400 transition-transform shrink-0 {expandedRecordId === r.id ? 'rotate-180' : ''}" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/></svg>
								</div>
								<div class="text-xs text-gray-500 dark:text-gray-400">{r.day_name}</div>
							</div>
							<span class="inline-block px-2 py-0.5 rounded-full text-xs font-medium shrink-0 {statusBadge(r.status)}">
								{r.status}
							</span>
						</div>
						<div class="flex items-center justify-between text-xs text-gray-500 dark:text-gray-400">
							<span class="tabular-nums">
								<span class={r.is_late || r.status === 'terlambat' ? "text-red-500 font-semibold" : ""}>{formatTime(r.check_in_time)}</span> 
								— 
								<span class={r.is_early_leave || r.status === 'pulang_awal' ? "text-red-500 font-semibold" : ""}>{formatTime(r.check_out_time)}</span>
							</span>
							{#if r.total_work_hours}
								<span class="tabular-nums">{r.total_work_hours.toFixed(1)} jam</span>
							{/if}
						</div>
						
						{#if expandedRecordId === r.id}
							<div class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-800 animate-in slide-in-from-top-2 duration-200">
								<div class="grid grid-cols-1 md:grid-cols-2 gap-4 bg-gray-50 dark:bg-gray-900/50 p-4 rounded-xl border border-gray-100 dark:border-gray-800">
									<div class="space-y-3">
										<h4 class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2 text-sm">
											<svg class="w-4 h-4 text-green-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1"/></svg>
											Check In: <span class={r.is_late || r.status === 'terlambat' ? "text-red-500 font-bold" : "text-gray-900 dark:text-white"}>{formatTime(r.check_in_time)}</span>
										</h4>
										{#if r.check_in_photo_url}
											<div class="aspect-video bg-gray-200 dark:bg-gray-800 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
												<img src={getPhotoUrl(r.check_in_photo_url)} alt="Check In Photo" class="w-full h-full object-cover" />
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

									<div class="space-y-3 border-t md:border-t-0 md:border-l border-gray-200 dark:border-gray-700 pt-4 md:pt-0 md:pl-4">
										<h4 class="font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2 text-sm">
											<svg class="w-4 h-4 text-red-500" fill="none" viewBox="0 0 24 24" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1"/></svg>
											Check Out: <span class={r.is_early_leave || r.status === 'pulang_awal' ? "text-red-500 font-bold" : "text-gray-900 dark:text-white"}>{formatTime(r.check_out_time)}</span>
										</h4>
										{#if r.check_out_time}
											{#if r.check_out_photo_url}
												<div class="aspect-video bg-gray-200 dark:bg-gray-800 rounded-lg overflow-hidden border border-gray-200 dark:border-gray-700">
													<img src={getPhotoUrl(r.check_out_photo_url)} alt="Check Out Photo" class="w-full h-full object-cover" />
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
								{#if r.is_late || r.status === 'terlambat' || r.is_early_leave || r.status === 'pulang_awal' || !r.check_in_time || !r.check_out_time}
									<div class="mt-4 flex justify-end px-4 pb-4">
										<a href="/absensi-manual" class="inline-flex items-center gap-1.5 px-4 py-2 bg-white dark:bg-gray-800 text-blue-600 dark:text-blue-400 border border-blue-200 dark:border-blue-800 rounded-lg text-sm font-medium hover:bg-blue-50 dark:hover:bg-blue-900/30 transition-colors">
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125"/></svg>
											Perbaiki Absensi
										</a>
									</div>
								{/if}
							</div>
						{/if}
					</div>
				{/each}
			</div>
		{/if}
	</div>
	</PullToRefresh>
</div>
