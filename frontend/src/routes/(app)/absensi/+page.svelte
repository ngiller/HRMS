<script lang="ts">
	import { onMount } from 'svelte';
	import { attendance, employeesApi, auth } from '$lib/api.js';
	import AnimatedPresence from '$lib/components/AnimatedPresence.svelte';
	import EmptyState from '$lib/components/EmptyState.svelte';
	import { hasPermission } from '$lib/permissions';

	// Camera detection config
	const DETECTION_INTERVAL_MS = 300;
	const PHOTO_CAPTURE_DELAY_MS = 500;

	// State
	let todayStatus = $state<any>(null);
	let items = $state<any[]>([]);
	let total = $state(0);
	let pageNum = $state(1);
	const perPage = 10;
	let loadingStatus = $state(true);
	let loadingHistory = $state(true);
	let checkInLoading = $state(false);
	let checkOutLoading = $state(false);
	let pageError = $state('');
	let showCamera = $state(false);
	let showCheckInConfirm = $state(false);
	let showCheckOutConfirm = $state(false);

	// Camera
	let videoEl = $state<HTMLVideoElement>();
	let canvasEl = $state<HTMLCanvasElement>();
	let photoCanvasEl = $state<HTMLCanvasElement>();
	let stream: MediaStream | null = null;
	let faceDetectionInterval: ReturnType<typeof setInterval> | null = null;
	let faceDetected = $state(false);
	let faceDetectionLoading = $state(false);
	let lastDescriptor: Float32Array | null = null;
	let photoDataUrl: string | null = null;

	// Check-in/out result
	let checkResult = $state<{ success: boolean; message: string } | null>(null);

	// Face registration
	let showFaceRegistration = $state(false);
	let faceRegLoading = $state(false);
	let faceRegError = $state('');
	let faceRegSuccess = $state(false);

	onMount(async () => {
		await loadFaceDetectionScript();
		loadTodayStatus();
		loadHistory();
	});

	async function loadTodayStatus() {
		loadingStatus = true;
		try {
			const res = await attendance.getTodayStatus();
			todayStatus = res.data || null;
		} catch {
			todayStatus = null;
		} finally {
			loadingStatus = false;
		}
	}

	async function loadHistory() {
		loadingHistory = true;
		try {
			const res = await attendance.myHistory(pageNum, perPage);
			if (res?.success) {
				items = res.data?.records || [];
				total = res.data?.total || 0;
			}
		} catch {
			// Silent
		} finally {
			loadingHistory = false;
		}
	}

	async function loadFaceDetectionScript() {
		return new Promise<void>((resolve, reject) => {
			if ((window as any).faceapi) {
				resolve();
				return;
			}
			const script = document.createElement('script');
			script.src = 'https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/dist/face-api.min.js';
			script.onload = () => {
				loadFaceDetectionModels().then(resolve).catch(reject);
			};
			script.onerror = () => reject(new Error('Gagal memuat face-api.js'));
			document.head.appendChild(script);
		});
	}

	async function loadFaceDetectionModels() {
		const faceapi = (window as any).faceapi;
		if (!faceapi) return;
		try {
			await faceapi.nets.tinyFaceDetector.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
			await faceapi.nets.faceLandmark68Net.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
			await faceapi.nets.faceRecognitionNet.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
		} catch (e) {
			console.warn('Face model load warning:', e);
		}
	}

	async function startCamera() {
		try {
			stream = await navigator.mediaDevices.getUserMedia({
				video: { width: 640, height: 480, facingMode: 'user' },
			});
			if (videoEl) {
				videoEl.srcObject = stream;
				await videoEl.play();
			}
			setTimeout(startFaceDetection, PHOTO_CAPTURE_DELAY_MS);
		} catch {
			pageError = 'Gagal mengakses kamera. Pastikan izin kamera diberikan.';
		}
	}

	function startFaceDetection() {
		if (faceDetectionInterval) clearInterval(faceDetectionInterval);
		faceDetectionInterval = setInterval(async () => {
			const faceapi = (window as any).faceapi;
			if (!faceapi || !videoEl || !canvasEl) return;
			try {
				const detections = await faceapi.detectAllFaces(videoEl, new faceapi.TinyFaceDetectorOptions({ inputSize: 320 }));
				const ctx = canvasEl.getContext('2d');
				if (ctx) {
					ctx.clearRect(0, 0, 640, 480);
					if (detections.length > 0) {
						const dims = { width: 640, height: 480 };
						const resized = faceapi.resizeResults(detections, dims);
						faceapi.draw.drawDetections(canvasEl, resized);
						faceDetected = true;
					} else {
						faceDetected = false;
					}
				}
			} catch {
				// Silent
			}
		}, DETECTION_INTERVAL_MS);
	}

	async function computeFaceDescriptor(video: HTMLVideoElement): Promise<Float32Array | null> {
		const faceapi = (window as any).faceapi;
		if (!faceapi) return null;
		try {
			const detections = await faceapi
				.detectAllFaces(video, new faceapi.TinyFaceDetectorOptions({ inputSize: 320 }))
				.withFaceLandmarks()
				.withFaceDescriptors();
			if (!detections || detections.length === 0) return null;
			return detections[0].descriptor;
		} catch {
			return null;
		}
	}

	function capturePhoto(video: HTMLVideoElement): string | null {
		if (!photoCanvasEl) return null;
		const ctx = photoCanvasEl.getContext('2d');
		if (!ctx) return null;
		photoCanvasEl.width = video.videoWidth;
		photoCanvasEl.height = video.videoHeight;
		ctx.drawImage(video, 0, 0);
		return photoCanvasEl.toDataURL('image/jpeg', 0.7);
	}

	function stopCamera() {
		if (faceDetectionInterval) {
			clearInterval(faceDetectionInterval);
			faceDetectionInterval = null;
		}
		if (stream) {
			stream.getTracks().forEach((t) => t.stop());
			stream = null;
		}
		faceDetected = false;
		lastDescriptor = null;
		photoDataUrl = null;
		showCamera = false;
		showCheckInConfirm = false;
		showCheckOutConfirm = false;
	}

	async function handleCheckIn() {
		showCamera = true;
		showCheckInConfirm = true;
		showCheckOutConfirm = false;
		await startCamera();
	}

	async function handleCheckOut() {
		showCamera = true;
		showCheckInConfirm = false;
		showCheckOutConfirm = true;
		await startCamera();
	}

	async function confirmCheckIn() {
		if (!videoEl) return;
		checkInLoading = true;
		checkResult = null;
		try {
			photoDataUrl = capturePhoto(videoEl);
			const descriptor = await computeFaceDescriptor(videoEl);
			lastDescriptor = descriptor;

			const payload: Record<string, any> = {
				photo: photoDataUrl,
			};
			if (descriptor) {
				payload.face_descriptor = JSON.stringify(Array.from(descriptor));
			}

			const res = await attendance.checkIn(payload);
			if (res?.success) {
				checkResult = { success: true, message: 'Check-in berhasil!' };
				stopCamera();
				loadTodayStatus();
				loadHistory();
			} else {
				checkResult = { success: false, message: res?.error || 'Check-in gagal' };
			}
		} catch (e: any) {
			checkResult = { success: false, message: e?.message || 'Check-in gagal' };
		} finally {
			checkInLoading = false;
			setTimeout(() => (checkResult = null), 4000);
		}
	}

	async function confirmCheckOut() {
		if (!videoEl) return;
		checkOutLoading = true;
		checkResult = null;
		try {
			photoDataUrl = capturePhoto(videoEl);
			const descriptor = await computeFaceDescriptor(videoEl);
			lastDescriptor = descriptor;

			const payload: Record<string, any> = {
				photo: photoDataUrl,
			};
			if (descriptor) {
				payload.face_descriptor = JSON.stringify(Array.from(descriptor));
			}

			const res = await attendance.checkOut(payload);
			if (res?.success) {
				checkResult = { success: true, message: 'Check-out berhasil!' };
				stopCamera();
				loadTodayStatus();
				loadHistory();
			} else {
				checkResult = { success: false, message: res?.error || 'Check-out gagal' };
			}
		} catch (e: any) {
			checkResult = { success: false, message: e?.message || 'Check-out gagal' };
		} finally {
			checkOutLoading = false;
			setTimeout(() => (checkResult = null), 4000);
		}
	}

	// ── Face Registration ──
	async function handleStartFaceRegistration() {
		showFaceRegistration = true;
		showCamera = true;
		faceRegError = '';
		faceRegSuccess = false;
		await startCamera();
	}

	async function handleRegisterFace() {
		if (!videoEl || !lastDescriptor) {
			// Try to compute descriptor first
			const descriptor = await computeFaceDescriptor(videoEl!);
			if (!descriptor) {
				faceRegError = 'Tidak ada wajah terdeteksi. Arahkan wajah ke kamera.';
				return;
			}
			lastDescriptor = descriptor;
		}

		faceRegLoading = true;
		faceRegError = '';
		try {
			const user = auth.getUser() as any;
			if (!user?.id) {
				faceRegError = 'Data user tidak ditemukan';
				return;
			}
			const descriptorJSON = JSON.stringify(Array.from(lastDescriptor!));
			const res = await employeesApi.registerFaceDescriptor(user.id, descriptorJSON);
			if (res?.success) {
				faceRegSuccess = true;
				stopCamera();
				showFaceRegistration = false;
				setTimeout(() => (faceRegSuccess = false), 3000);
			} else {
				faceRegError = res?.error || 'Gagal registrasi wajah';
			}
		} catch (e: any) {
			faceRegError = e?.message || 'Gagal registrasi wajah';
		} finally {
			faceRegLoading = false;
		}
	}

	function cancelFaceRegistration() {
		stopCamera();
		showFaceRegistration = false;
		faceRegError = '';
	}

	function paginateHistory(page: number) {
		pageNum = page;
		loadHistory();
	}

	function formatTime(d: string | null | undefined) {
		if (!d) return '-';
		return new Date(d).toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function formatDate(d: string) {
		return new Date(d).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			present: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			late: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			absent: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			leave: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
			holiday: 'bg-purple-100 text-purple-700 dark:bg-purple-900/30 dark:text-purple-300',
		};
		return map[status] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300';
	}

	function getStatusText(status: string) {
		const map: Record<string, string> = {
			present: 'Hadir',
			late: 'Terlambat',
			absent: 'Absen',
			leave: 'Cuti',
			holiday: 'Libur',
		};
		return map[status] || status;
	}
</script>

<svelte:head>
	<title>Absensi - HRMS</title>
</svelte:head>

<div class="w-full">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Absensi</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">Check-in / Check-out kehadiran hari ini</p>
		</div>
		<div class="flex items-center gap-2">
			<button onclick={handleStartFaceRegistration}
				class="px-4 py-2 bg-indigo-50 text-indigo-700 rounded-lg text-sm font-semibold hover:bg-indigo-100 transition cursor-pointer dark:bg-indigo-900/30 dark:text-indigo-300 dark:hover:bg-indigo-900/50 flex items-center gap-1.5">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
				Registrasi Wajah
			</button>
		</div>
	</div>

	{#if checkResult}
		<div class="mb-4 px-5 py-3.5 rounded-xl border {checkResult.success ? 'bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800' : 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'}">
			<div class="flex items-center gap-2.5">
				{#if checkResult.success}
					<svg class="w-5 h-5 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
				{:else}
					<svg class="w-5 h-5 text-red-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
				{/if}
				<p class="text-sm font-medium {checkResult.success ? 'text-emerald-800 dark:text-emerald-200' : 'text-red-800 dark:text-red-200'}">{checkResult.message}</p>
			</div>
		</div>
	{/if}

	{#if faceRegSuccess}
		<div class="mb-4 px-5 py-3.5 rounded-xl border bg-emerald-50 dark:bg-emerald-900/20 border-emerald-200 dark:border-emerald-800">
			<div class="flex items-center gap-2.5">
				<svg class="w-5 h-5 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
				<p class="text-sm font-medium text-emerald-800 dark:text-emerald-200">Wajah berhasil diregistrasi!</p>
			</div>
		</div>
	{/if}

	<!-- Camera Overlay -->
	{#if showCamera}
		<div class="fixed inset-0 z-50 bg-black/70 flex items-center justify-center p-4">
			<div class="bg-white dark:bg-gray-900 rounded-2xl max-w-lg w-full overflow-hidden shadow-2xl">
				<div class="relative">
					<video
						bind:this={videoEl}
						class="w-full aspect-[4/3] object-cover bg-gray-900"
						autoplay
						muted
						playsinline
					></video>
					<canvas
						bind:this={canvasEl}
						class="absolute inset-0 w-full h-full"
						width={640}
						height={480}
					></canvas>

					<!-- Registration Badge -->
					{#if showFaceRegistration}
						<div class="absolute top-4 left-4 bg-indigo-600/90 text-white text-xs font-semibold px-3 py-1.5 rounded-full backdrop-blur-sm">
							Registrasi Wajah
						</div>
					{/if}

					<!-- Close button -->
					<button onclick={showFaceRegistration ? cancelFaceRegistration : stopCamera}
						class="absolute top-4 right-4 w-8 h-8 bg-black/50 text-white rounded-full flex items-center justify-center hover:bg-black/70 transition cursor-pointer">
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
					</button>

					<!-- Face detection status -->
					<div class="absolute bottom-4 left-4">
						{#if faceDetected}
							<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-emerald-500/80 text-white backdrop-blur-sm">
								<span class="w-1.5 h-1.5 rounded-full bg-white"></span>
								Wajah Terdeteksi
							</span>
						{:else}
							<span class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-full text-xs font-medium bg-amber-500/80 text-white backdrop-blur-sm">
								<span class="w-1.5 h-1.5 rounded-full bg-white"></span>
								Arahkan Wajah ke Kamera
							</span>
						{/if}
					</div>
				</div>

				<div class="p-4 space-y-3">
					{#if showFaceRegistration}
						<!-- Registration mode controls -->
						{#if faceRegError}
							<p class="text-xs text-red-500 text-center">{faceRegError}</p>
						{/if}
						<div class="flex items-center gap-3">
							<button onclick={handleRegisterFace}
								disabled={faceRegLoading || !faceDetected}
								class="flex-1 px-5 py-2.5 bg-indigo-600 text-white rounded-xl text-sm font-semibold hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer flex items-center justify-center gap-2">
								{#if faceRegLoading}
									<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
									Menyimpan Wajah...
								{:else}
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
									Simpan Wajah
								{/if}
							</button>
							<button onclick={cancelFaceRegistration}
								class="px-5 py-2.5 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-xl text-sm font-medium hover:bg-gray-200 dark:hover:bg-gray-700 transition cursor-pointer">
								Batal
							</button>
						</div>
					{:else if showCheckInConfirm}
						<button onclick={confirmCheckIn}
							disabled={checkInLoading || !faceDetected}
							class="w-full px-5 py-3 bg-blue-600 text-white rounded-xl text-sm font-semibold hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer flex items-center justify-center gap-2">
							{#if checkInLoading}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
								Memproses Check-in...
							{:else}
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
								Konfirmasi Check-in
							{/if}
						</button>
					{:else if showCheckOutConfirm}
						<button onclick={confirmCheckOut}
							disabled={checkOutLoading || !faceDetected}
							class="w-full px-5 py-3 bg-amber-600 text-white rounded-xl text-sm font-semibold hover:bg-amber-700 disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer flex items-center justify-center gap-2">
							{#if checkOutLoading}
								<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
								Memproses Check-out...
							{:else}
								<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>
								Konfirmasi Check-out
							{/if}
						</button>
					{/if}

					<p class="text-xs text-gray-400 text-center">
						{faceDetected ? 'Wajah terdeteksi, siap melakukan absensi.' : 'Arahkan wajah Anda ke kamera.'}
					</p>
				</div>
			</div>
		</div>
	{/if}

	<canvas bind:this={photoCanvasEl} class="hidden"></canvas>

	<!-- Status & Action Cards -->
	<div class="grid grid-cols-1 md:grid-cols-2 gap-4 mb-6">
		<!-- Check-in Card -->
		<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm transition-all hover:shadow-md {todayStatus?.has_checked_in ? 'border-l-4 border-l-emerald-500' : ''}">
			<div class="flex items-center justify-between mb-3">
				<div class="flex items-center gap-2.5">
					<div class="w-10 h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center {todayStatus?.has_checked_in ? 'text-emerald-600' : 'text-blue-600'}">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
					</div>
					<div>
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white">Check In</h3>
						<p class="text-xs text-gray-500 dark:text-gray-400">Absensi Masuk</p>
					</div>
				</div>
				<div class="text-right">
					<div class="text-lg font-bold {todayStatus?.has_checked_in ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'}">
						{todayStatus?.has_checked_in ? formatTime(todayStatus?.record?.check_in_time) : '--:--'}
					</div>
				</div>
			</div>
			<button onclick={handleCheckIn}
				disabled={todayStatus?.has_checked_in || loadingStatus}
				class="w-full py-2.5 rounded-xl text-sm font-semibold transition cursor-pointer {todayStatus?.has_checked_in ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed' : 'bg-blue-600 text-white hover:bg-blue-700'}">
				{todayStatus?.has_checked_in ? 'Sudah Check In' : 'Check In'}
			</button>
		</div>

		<!-- Check-out Card -->
		<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-5 shadow-sm transition-all hover:shadow-md {todayStatus?.has_checked_out ? 'border-l-4 border-l-amber-500' : ''}">
			<div class="flex items-center justify-between mb-3">
				<div class="flex items-center gap-2.5">
					<div class="w-10 h-10 rounded-xl bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center {todayStatus?.has_checked_out ? 'text-amber-600' : 'text-amber-600'}">
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15" /></svg>
					</div>
					<div>
						<h3 class="text-sm font-semibold text-gray-900 dark:text-white">Check Out</h3>
						<p class="text-xs text-gray-500 dark:text-gray-400">Absensi Pulang</p>
					</div>
				</div>
				<div class="text-right">
					<div class="text-lg font-bold {todayStatus?.has_checked_out ? 'text-amber-600 dark:text-amber-400' : 'text-gray-400'}">
						{todayStatus?.has_checked_out ? formatTime(todayStatus?.record?.check_out_time) : '--:--'}
					</div>
				</div>
			</div>
			<button onclick={handleCheckOut}
				disabled={todayStatus?.has_checked_out || !todayStatus?.has_checked_in || loadingStatus || todayStatus?.has_checked_in === false}
				class="w-full py-2.5 rounded-xl text-sm font-semibold transition cursor-pointer {todayStatus?.has_checked_out ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed' : 'bg-amber-600 text-white hover:bg-amber-700'}">
				{todayStatus?.has_checked_out ? 'Sudah Check Out' : 'Check Out'}
			</button>
		</div>
	</div>

	<!-- Schedule Info -->
	{#if todayStatus?.schedule_name}
		<div class="mb-6 px-5 py-3 bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-xl">
			<div class="flex items-center justify-between text-sm">
				<div class="flex items-center gap-2">
					<svg class="w-4 h-4 text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					<span class="text-gray-500 dark:text-gray-400">Jadwal: <strong class="font-semibold text-gray-700 dark:text-gray-300">{todayStatus.schedule_name}</strong></span>
				</div>
				<span class="text-gray-500 dark:text-gray-400">
					{todayStatus.schedule_start || '-'} - {todayStatus.schedule_end || '-'}
				</span>
			</div>
		</div>
	{/if}

	<!-- Attendance History -->
	<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-hidden">
		<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 flex items-center justify-between">
			<h2 class="text-sm font-semibold text-gray-900 dark:text-white">Riwayat Absensi</h2>
			<span class="text-xs text-gray-400">Total {total} record</span>
		</div>

		{#if loadingHistory}
			<div class="flex items-center justify-center py-10">
				<div class="animate-spin h-6 w-6 border-2 border-blue-600 border-t-transparent rounded-full"></div>
			</div>
		{:else if items.length === 0}
			<div class="p-6 text-center">
				<svg class="w-10 h-10 text-gray-300 dark:text-gray-600 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
					<p class="text-sm text-gray-400">Belum ada riwayat absensi</p>
				</div>
		{:else}
			<div class="divide-y divide-gray-100 dark:divide-gray-700/50">
				{#each items as item}
					<div class="px-6 py-3.5 flex items-center justify-between hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors">
						<div class="flex items-center gap-3 min-w-0">
							<div class="w-9 h-9 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center shrink-0">
								<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
							</div>
							<div class="min-w-0">
								<div class="text-sm font-medium text-gray-900 dark:text-white truncate">{formatDate(item.date)}</div>
								<div class="text-xs text-gray-500 dark:text-gray-400">
									{formatTime(item.check_in_time)} → {formatTime(item.check_out_time)}
								</div>
							</div>
						</div>
						<div class="flex items-center gap-2 shrink-0">
							{#if item.is_late && item.late_minutes && item.late_minutes > 0}
								<span class="text-xs text-amber-600 dark:text-amber-400 font-medium">{item.late_minutes} menit</span>
							{/if}
							<span class="px-2.5 py-1 rounded-full text-xs font-medium {getStatusBadge(item.status)}">
								{getStatusText(item.status)}
							</span>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
