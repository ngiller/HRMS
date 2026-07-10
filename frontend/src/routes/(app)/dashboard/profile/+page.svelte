<script lang="ts">
	import { onMount } from 'svelte';
	import { auth, attendance, employeesApi } from '$lib/api.js';

	interface UserData {
		id: string;
		employee_id: string;
		full_name: string;
		email: string;
		role_id: string;
		role_slug: string;
		role_name: string;
		position_name: string;
		department_name: string;
		avatar_initials: string;
		has_face_descriptor?: boolean;
	}

	let user = $state(auth.getUser() as UserData | null);
	let userData = $state<any>(null);
	let isLoading = $state(true);

	// Attendance history
	let attendanceHistory = $state<any[]>([]);
	let historyLoading = $state(false);
	let historyError = $state('');

	// Face registration
	let faceRegLoading = $state(false);
	let faceRegMessage = $state('');
	let faceRegError = $state('');
	let faceRegSuccess = $state(false);
	let cameraMode = $state<'none' | 'register'>('none');
	let videoEl = $state<HTMLVideoElement>();
	let canvasEl = $state<HTMLCanvasElement>();
	let stream: MediaStream | null = null;
	let faceDetected = $state(false);
	let faceDetectionLoading = $state(false);

	onMount(() => {
		loadProfile();
		loadAttendanceHistory();
	});

	async function loadProfile() {
		isLoading = true;
		try {
			const res = await auth.getMe();
			if (res?.success) {
				userData = res.data;
			}
		} catch {
			userData = user;
		} finally {
			isLoading = false;
		}
	}

	async function loadAttendanceHistory() {
		historyLoading = true;
		historyError = '';
		try {
			const res = await attendance.myHistory(1, 5) as any;
			if (res?.success) {
				attendanceHistory = res.data?.records || [];
			}
		} catch (e: any) {
			historyError = e?.message || 'Gagal memuat riwayat absensi';
		} finally {
			historyLoading = false;
		}
	}

	function getInitials(name: string): string {
		if (!name) return 'NA';
		return name.substring(0, 2).toUpperCase();
	}

	const infoItems = $derived([
		{ label: 'Nama Lengkap', value: userData?.full_name || user?.full_name || '-' },
		{ label: 'ID Karyawan', value: userData?.employee_id || user?.employee_id || '-' },
		{ label: 'Email', value: userData?.email || user?.email || '-' },
		{ label: 'Jabatan', value: userData?.position_name || user?.position_name || '-' },
		{ label: 'Departemen', value: userData?.department_name || user?.department_name || '-' },
		{ label: 'Role', value: userData?.role_name || user?.role_name || '-' },
	]);

	// ── Face Registration ──
	async function handleStartFaceRegistration() {
		cameraMode = 'register';
		faceRegError = '';
		faceRegMessage = '';
		faceRegSuccess = false;
		faceDetected = false;
		faceDetectionLoading = true;

		// Load face-api.js if needed
		try {
			if (!(window as any).faceapi) {
				await loadFaceDetectionScript();
			}
			const faceapi = (window as any).faceapi;
			if (faceapi) {
				if (!faceapi.nets.tinyFaceDetector) {
					await faceapi.nets.tinyFaceDetector.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
				}
				if (!faceapi.nets.faceLandmark68Net) {
					await faceapi.nets.faceLandmark68Net.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
				}
				if (!faceapi.nets.faceRecognitionNet) {
					await faceapi.nets.faceRecognitionNet.loadFromUri('https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/weights');
				}
			}
		} catch (e) {
			console.warn('Face-api load error:', e);
		} finally {
			faceDetectionLoading = false;
		}

		// Start camera
		try {
			stream = await navigator.mediaDevices.getUserMedia({
				video: { width: 640, height: 480, facingMode: 'user' }
			});
			if (videoEl) {
				videoEl.srcObject = stream;
				await videoEl.play();
				startFaceDetectionLoop();
			}
		} catch (e) {
			faceRegError = 'Gagal mengakses kamera';
			cameraMode = 'none';
		}
	}

	function startFaceDetectionLoop() {
		const faceapi = (window as any).faceapi;
		if (!faceapi || !videoEl || !canvasEl) return;

		const interval = setInterval(async () => {
			if (!videoEl || !canvasEl || cameraMode !== 'register') {
				clearInterval(interval);
				return;
			}
			try {
				const detections = await faceapi.detectAllFaces(videoEl, new faceapi.TinyFaceDetectorOptions({ inputSize: 320 }));
				const ctx = canvasEl.getContext('2d');
				if (ctx && canvasEl) {
					ctx.clearRect(0, 0, canvasEl.width, canvasEl.height);
					if (detections && detections.length > 0) {
						const dims = { width: videoEl.videoWidth, height: videoEl.videoHeight };
						const resized = faceapi.resizeResults(detections, dims);
						faceapi.draw.drawDetections(canvasEl, resized);
						faceDetected = detections.length > 0;
					} else {
						faceDetected = false;
					}
				}
			} catch {
				// Silent
			}
		}, 200);
	}

	async function handleRegisterFace() {
		if (!videoEl || !canvasEl || !userData?.id) return;
		faceRegLoading = true;
		faceRegError = '';
		faceRegMessage = '';

		try {
			const faceapi = (window as any).faceapi;
			if (!faceapi) {
				faceRegError = 'Library deteksi wajah tidak tersedia';
				return;
			}

			// Compute face descriptor from current video frame
			const detections = await faceapi.detectAllFaces(videoEl, new faceapi.TinyFaceDetectorOptions({ inputSize: 320 }))
				.withFaceLandmarks().withFaceDescriptors();

			if (!detections || detections.length === 0) {
				faceRegError = 'Tidak ada wajah terdeteksi. Pastikan wajah terlihat jelas.';
				return;
			}

			const descriptor = detections[0].descriptor;
			const descriptorJSON = JSON.stringify(Array.from(descriptor));

			const res = await employeesApi.registerFaceDescriptor(userData.id, descriptorJSON);
			if (res?.success) {
				faceRegSuccess = true;
				faceRegMessage = 'Wajah berhasil diregistrasi!';
				stopCamera();
				setTimeout(() => {
					faceRegSuccess = false;
					faceRegMessage = '';
				}, 3000);
			} else {
				faceRegError = res?.error || 'Gagal registrasi wajah';
			}
		} catch (e: any) {
			faceRegError = e?.message || 'Gagal registrasi wajah';
		} finally {
			faceRegLoading = false;
		}
	}

	function stopCamera() {
		if (stream) {
			stream.getTracks().forEach(track => track.stop());
			stream = null;
		}
		cameraMode = 'none';
		faceDetected = false;
	}

	function cancelFaceRegistration() {
		stopCamera();
	}

	async function loadFaceDetectionScript() {
		return new Promise<void>((resolve, reject) => {
			if ((window as any).faceapi) {
				resolve();
				return;
			}
			const script = document.createElement('script');
			script.src = 'https://cdn.jsdelivr.net/gh/justadudewhohacks/face-api.js@master/dist/face-api.min.js';
			script.onload = () => resolve();
			script.onerror = () => reject(new Error('Gagal memuat face-api.js'));
			document.head.appendChild(script);
		});
	}

	function formatDate(dateStr: string) {
		const d = new Date(dateStr);
		return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

	function formatTime(dateStr: string | null | undefined) {
		if (!dateStr) return '-';
		const d = new Date(dateStr);
		return d.toLocaleTimeString('id-ID', { hour: '2-digit', minute: '2-digit' });
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			present: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			late: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			absent: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			leave: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
		};
		return map[status] || 'bg-gray-100 text-gray-700';
	}
</script>

<div class="w-full">
	<!-- Header -->
	<div class="flex items-center gap-4 mb-6">
		<div class="w-16 h-16 bg-gradient-to-br from-[#1A56DB] to-[#1e3a8a] rounded-2xl flex items-center justify-center text-white font-bold text-xl shrink-0">
			{getInitials(userData?.full_name || user?.full_name || '')}
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Profil Saya</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400 mt-0.5">
				{userData?.position_name || user?.position_name || 'Karyawan'}
				{#if userData?.department_name || user?.department_name}
					· {userData?.department_name || user?.department_name}
				{/if}
			</p>
		</div>
	</div>

	{#if isLoading}
		<div class="text-center py-12">
			<div class="animate-spin h-8 w-8 border-4 border-[#1A56DB] border-t-transparent rounded-full mx-auto"></div>
			<p class="text-sm text-gray-400 mt-3">Memuat data...</p>
		</div>
	{:else}
		<div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
			<!-- Left Column: Info Personal -->
			<div class="lg:col-span-2 space-y-6">
				<!-- Data Diri Card -->
				<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 flex items-center justify-between">
						<div class="flex items-center gap-2">
							<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
							<h2 class="font-semibold text-gray-900 dark:text-gray-100">Data Diri</h2>
						</div>
						<span class="text-xs text-gray-400">{userData?.employee_id || user?.employee_id || '-'}</span>
					</div>
					<div class="divide-y divide-gray-100 dark:divide-gray-800">
						{#each infoItems as item}
							<div class="px-6 py-3.5 flex items-center justify-between">
								<span class="text-sm text-gray-500 dark:text-gray-400">{item.label}</span>
								<span class="text-sm font-medium text-gray-900 dark:text-gray-100">{item.value}</span>
							</div>
						{/each}
					</div>
				</div>

				<!-- Riwayat Absensi Card -->
				<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 flex items-center justify-between">
						<div class="flex items-center gap-2">
							<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
							<h2 class="font-semibold text-gray-900 dark:text-gray-100">Riwayat Absensi</h2>
						</div>
						<a href="/absensi" class="text-xs font-medium text-blue-600 dark:text-blue-400 hover:underline">Lihat Semua</a>
					</div>

					{#if historyLoading}
						<div class="flex items-center justify-center py-8">
							<div class="animate-spin h-6 w-6 border-2 border-[#1A56DB] border-t-transparent rounded-full"></div>
						</div>
					{:else if historyError}
						<div class="p-6 text-sm text-red-500 text-center">{historyError}</div>
					{:else if attendanceHistory.length === 0}
						<div class="p-6 text-center">
							<svg class="w-10 h-10 text-gray-300 dark:text-gray-600 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
							<p class="text-sm text-gray-400">Belum ada riwayat absensi</p>
						</div>
					{:else}
						<div class="divide-y divide-gray-100 dark:divide-gray-800">
							{#each attendanceHistory as record}
								<div class="px-6 py-3.5 flex items-center justify-between">
									<div>
										<div class="text-sm font-medium text-gray-900 dark:text-gray-100">{formatDate(record.date)}</div>
										<div class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
											{record.check_in_time ? formatTime(record.check_in_time) : '-'} → {record.check_out_time ? formatTime(record.check_out_time) : '-'}
										</div>
									</div>
									<span class="px-2.5 py-1 rounded-full text-xs font-medium {getStatusBadge(record.status)}">
										{record.status === 'present' ? 'Hadir' : record.status === 'late' ? 'Terlambat' : record.status === 'absent' ? 'Absen' : record.status || '-'}
									</span>
								</div>
							{/each}
						</div>
					{/if}
				</div>
			</div>

			<!-- Right Column: Face Registration & Info -->
			<div class="space-y-6">
				<!-- Registrasi Wajah Card -->
				<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50">
						<div class="flex items-center gap-2">
							<svg class="w-5 h-5 text-gray-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
							<h2 class="font-semibold text-gray-900 dark:text-gray-100">Registrasi Wajah</h2>
						</div>
					</div>
					<div class="p-6">
						{#if cameraMode === 'register'}
							<div class="relative">
								<video
									bind:this={videoEl}
									class="w-full rounded-lg bg-black"
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

								{#if faceDetectionLoading}
									<div class="absolute inset-0 flex items-center justify-center bg-black/50 rounded-lg">
										<div class="text-white text-sm">Memuat detektor wajah...</div>
									</div>
								{/if}

								<div class="mt-3 flex items-center justify-center gap-2">
									{#if faceDetected}
										<span class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-medium bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300">
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
											Wajah Terdeteksi
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 px-3 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300">
											<span class="w-1.5 h-1.5 rounded-full bg-amber-500"></span>
											Arahkan Wajah ke Kamera
										</span>
									{/if}
								</div>

								<div class="mt-4 flex items-center gap-3">
									<button
										onclick={handleRegisterFace}
										disabled={faceRegLoading || !faceDetected}
										class="flex-1 px-4 py-2.5 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition cursor-pointer flex items-center justify-center gap-2">
										{#if faceRegLoading}
											<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" /></svg>
											Menyimpan...
										{:else}
											<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M4.5 12.75l6 6 9-13.5" /></svg>
											Simpan Wajah
										{/if}
									</button>
									<button
										onclick={cancelFaceRegistration}
										class="px-4 py-2.5 bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-lg text-sm font-medium hover:bg-gray-200 dark:hover:bg-gray-700 transition cursor-pointer">
										Batal
									</button>
								</div>

								{#if faceRegError}
									<p class="mt-2 text-xs text-red-500">{faceRegError}</p>
								{/if}
							</div>
						{:else}
							<div class="text-center">
								<div class="w-16 h-16 mx-auto mb-3 rounded-full bg-gradient-to-br from-indigo-100 to-indigo-50 dark:from-indigo-900/30 dark:to-indigo-800/20 flex items-center justify-center">
									<svg class="w-8 h-8 text-indigo-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
								</div>
								<p class="text-sm text-gray-600 dark:text-gray-400 mb-1">Registrasi wajah untuk absensi</p>
								<p class="text-xs text-gray-400 dark:text-gray-500 mb-4">Gunakan kamera untuk mendaftarkan wajah Anda sebagai verifikasi absensi</p>
								<button
									onclick={handleStartFaceRegistration}
									class="px-5 py-2.5 bg-indigo-600 text-white rounded-lg text-sm font-semibold hover:bg-indigo-700 transition cursor-pointer inline-flex items-center gap-2">
									<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z" /></svg>
									Registrasi Wajah
								</button>

								{#if faceRegSuccess}
									<p class="mt-3 text-xs text-emerald-600 font-medium">{faceRegMessage}</p>
								{/if}
							</div>
						{/if}
					</div>
				</div>

				<!-- Info Card -->
				<div class="bg-amber-50 dark:bg-amber-900/20 border border-amber-200 dark:border-amber-800 rounded-xl p-4">
					<div class="flex items-start gap-3">
						<svg class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" />
						</svg>
						<div>
							<p class="text-sm font-medium text-amber-800 dark:text-amber-200">Data profil bersumber dari data karyawan</p>
							<p class="text-xs text-amber-600 dark:text-amber-400 mt-1">Untuk perubahan data, hubungi HR atau buka halaman Karyawan.</p>
						</div>
					</div>
				</div>
			</div>
		</div>
	{/if}
</div>
