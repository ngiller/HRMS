<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { attendance, employeesApi, auth, company as companyApi, holidays } from '$lib/api.js';
	import config from '$lib/config.js';
	import { hasPermission } from '$lib/permissions.js';

	// Camera detection config
	const DETECTION_INTERVAL_MS = 300;
	const PHOTO_CAPTURE_DELAY_MS = 500;

	// State
	let todayStatus = $state<any>(null);
	let items = $state<any[]>([]);
	let total = $state(0);
	let pageNum = $state(1);
	const perPage = 50;
	let loadingStatus = $state(true);
	let loadingHistory = $state(true);
	let checkInLoading = $state(false);
	let checkOutLoading = $state(false);
	let pageError = $state('');
	let showCamera = $state(false);
	let showCheckInConfirm = $state(false);
	let showCheckOutConfirm = $state(false);

	// Date Filters for History
	let filterDateFrom = $state('');
	let filterDateTo = $state('');
	let holidayList = $state<any[]>([]);

	// Permission status
	let camGranted = $state<boolean | null>(null);
	let locGranted = $state<boolean | null>(null);
	let checkingPerms = $state(true);

	onMount(async () => {
		checkPermissions();
		await loadFaceDetectionScript();
		loadTodayStatus();
		
		// Load company settings and holidays concurrently
		try {
			const [settingsRes, holidayRes] = await Promise.all([
				companyApi.getSettings(),
				holidays.getByYear(new Date().getFullYear())
			]) as any;

			if (holidayRes?.success && holidayRes?.data) {
				holidayList = holidayRes.data.holidays || [];
			}

			if (settingsRes?.success && settingsRes?.data?.hr_settings) {
				const hrs = settingsRes.data.hr_settings;
				const startDay = hrs.cutoff_start_day || 26;
				const endDay = hrs.cutoff_end_day || 25;
				const period = calculateActiveCutoffPeriod(startDay, endDay);
				filterDateFrom = period.start;
				filterDateTo = period.end;
			} else {
				// Default fallback
				const period = calculateActiveCutoffPeriod(26, 25);
				filterDateFrom = period.start;
				filterDateTo = period.end;
			}
		} catch {
			// Default fallback
			const period = calculateActiveCutoffPeriod(26, 25);
			filterDateFrom = period.start;
			filterDateTo = period.end;
		}

		loadHistory();
	});

	function calculateActiveCutoffPeriod(startDay: number, endDay: number) {
		const now = new Date();
		const year = now.getFullYear();
		const month = now.getMonth(); // 0-indexed: 0=Jan, 11=Dec
		const day = now.getDate();

		let fromDate: Date;
		let toDate: Date;

		if (!startDay || !endDay) {
			// Calendar month default
			fromDate = new Date(year, month, 1);
			toDate = new Date(year, month + 1, 0); // Last day of month
		} else {
			// If endDay is 25:
			// If today is <= 25:
			//   Period runs from M-1 26th to M 25th
			// If today is > 25:
			//   Period runs from M 26th to M+1 25th
			if (day <= endDay) {
				fromDate = new Date(year, month - 1, startDay);
				toDate = new Date(year, month, endDay);
			} else {
				fromDate = new Date(year, month, startDay);
				toDate = new Date(year, month + 1, endDay);
			}
		}

		// Format to YYYY-MM-DD
		const formatYYYYMMDD = (d: Date) => {
			const y = d.getFullYear();
			const m = String(d.getMonth() + 1).padStart(2, '0');
			const dayStr = String(d.getDate()).padStart(2, '0');
			return `${y}-${m}-${dayStr}`;
		};

		return {
			start: formatYYYYMMDD(fromDate),
			end: formatYYYYMMDD(toDate)
		};
	}

	function generateDateRange(startStr: string, endStr: string) {
		const list: string[] = [];
		if (!startStr || !endStr) return list;
		const start = new Date(startStr);
		const end = new Date(endStr);
		const curr = new Date(start);
		while (curr <= end) {
			list.push(curr.toISOString().split('T')[0]);
			curr.setDate(curr.getDate() + 1);
		}
		return list.reverse(); // Newest dates first
	}

	let timesheetItems = $derived.by(() => {
		if (!filterDateFrom || !filterDateTo) return [];
		
		const dates = generateDateRange(filterDateFrom, filterDateTo);
		const recordMap = new Map<string, any>();
		
		items.forEach(item => {
			if (item.date) {
				const dateStr = item.date.split('T')[0];
				recordMap.set(dateStr, item);
			}
		});

		const todayStr = new Date().toISOString().split('T')[0];

		return dates.map(dateStr => {
			if (recordMap.has(dateStr)) {
				const rec = recordMap.get(dateStr);
				const matchingHoliday = holidayList.find(h => h.date === dateStr);
				if (matchingHoliday) {
					rec.holiday_name = matchingHoliday.name;
				}
				return rec;
			}

			const dateObj = new Date(dateStr);
			const dayOfWeek = dateObj.getDay(); // 0=Sunday, 6=Saturday
			const isWeekend = dayOfWeek === 0 || dayOfWeek === 6;
			
			const matchingHoliday = holidayList.find(h => h.date === dateStr);
			
			let status = 'absent';
			let holidayName = '';
			if (matchingHoliday) {
				status = 'holiday';
				holidayName = matchingHoliday.name;
			} else if (isWeekend) {
				status = 'holiday';
				holidayName = 'Libur Akhir Pekan';
			} else if (dateStr > todayStr) {
				status = 'scheduled';
			} else if (dateStr === todayStr) {
				status = 'scheduled';
			}

			return {
				id: '',
				date: dateStr + 'T00:00:00Z',
				day_name: dateObj.toLocaleDateString('id-ID', { weekday: 'long' }),
				status: status,
				holiday_name: holidayName,
				check_in_time: null,
				check_out_time: null,
				is_virtual: true
			};
		});
	});

	async function checkPermissions() {
		checkingPerms = true;
		// Check camera
		try {
			const camResult = await navigator.permissions.query({ name: 'camera' as PermissionName });
			camGranted = camResult.state === 'granted';
			camResult.onchange = () => { camGranted = camResult.state === 'granted'; };
		} catch {
			camGranted = null;
		}
		// Check location
		try {
			const locResult = await navigator.permissions.query({ name: 'geolocation' as PermissionName });
			locGranted = locResult.state === 'granted';
			locResult.onchange = () => { locGranted = locResult.state === 'granted'; };
		} catch {
			locGranted = null;
		}
		checkingPerms = false;
	}

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
	let locationWarning = $state<string | null>(null);

	function openDetail(item: any) {
		if (item?.id) {
			goto(`/absensi/${item.id}`);
		}
	}

	function getPhotoUrl(url: string | null | undefined): string {
		if (!url) return '';
		if (url.startsWith('http')) return url;
		return `${config.API_BASE_URL}${url}`;
	}

	// Face registration
	let showFaceRegistration = $state(false);
	let faceRegLoading = $state(false);
	let faceRegError = $state('');
	let faceRegSuccess = $state(false);

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
			const res = await attendance.myHistory(pageNum, perPage, filterDateFrom, filterDateTo);
			if (res?.success) {
				items = res.data || [];
				total = (res.meta as any)?.total || 0;
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

	
	function getCoordinates(): Promise<{ lat: number; lng: number } | null> {
		return new Promise((resolve) => {
			if (typeof window === 'undefined' || !navigator.geolocation) {
				console.warn('Geolocation not supported');
				resolve(null);
				return;
			}
			navigator.geolocation.getCurrentPosition(
				(position) => {
					resolve({
						lat: position.coords.latitude,
						lng: position.coords.longitude
					});
				},
				(error) => {
					console.warn('Geolocation error:', error);
					resolve(null);
				},
				{
					enableHighAccuracy: true,
					timeout: 5000,
					maximumAge: 0
				}
			);
		});
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

			const coords = await getCoordinates();

			const payload: Record<string, any> = {
				photo: photoDataUrl,
			};
			if (descriptor) {
				payload.face_descriptor = JSON.stringify(Array.from(descriptor));
			}
			if (coords) {
				payload.lat = coords.lat;
				payload.lng = coords.lng;
			}
			const res = await attendance.checkIn(payload);
			if (res?.success) {
				checkResult = { success: true, message: 'Check-in berhasil!' };
				locationWarning = res?.data?.location_warning || null;
				stopCamera();
				loadTodayStatus();
				loadHistory();
			} else {
				checkResult = { success: false, message: res?.error || 'Check-in gagal' };
				locationWarning = null;
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

			const coords = await getCoordinates();

			const payload: Record<string, any> = {
				photo: photoDataUrl,
			};
			if (descriptor) {
				payload.face_descriptor = JSON.stringify(Array.from(descriptor));
			}
			if (coords) {
				payload.lat = coords.lat;
				payload.lng = coords.lng;
			}

			const res = await attendance.checkOut(payload);
			if (res?.success) {
				checkResult = { success: true, message: 'Check-out berhasil!' };
				locationWarning = res?.data?.location_warning || null;
				stopCamera();
				loadTodayStatus();
				loadHistory();
			} else {
				checkResult = { success: false, message: res?.error || 'Check-out gagal' };
				locationWarning = null;
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

	function getHour(timeStr: string | null | undefined): number | null {
		if (!timeStr) return null;
		const d = new Date(timeStr);
		if (isNaN(d.getTime())) return null;
		return d.getHours();
	}

	/**
	 * Deteksi apakah record ini telat — pake data backend, plus fallback time-based heuristic
	 * buat record lama yg is_late-nya salah di database.
	 * Normal jam kerja mulai 7-9 pagi. Lewat jam 10 = pasti telat.
	 */
	function isLateRecord(item: any): boolean {
		if (item.is_late === true || item.status === 'terlambat') return true;
		if (!item.check_in_time) return false;
		const d = new Date(item.check_in_time);
		const hour = d.getHours();
		const min = d.getMinutes();
		if (hour > 8 || (hour === 8 && min > 0)) return true;
		return false;
	}

	function isEarlyLeaveRecord(item: any): boolean {
		if (item.is_early_leave === true) return true;
		if (!item.check_out_time) return false;
		const d = new Date(item.check_out_time);
		const hour = d.getHours();
		if (hour < 17) return true;
		return false;
	}

	function getStatusBadge(status: string) {
		const map: Record<string, string> = {
			present: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-300',
			late: 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-300',
			absent: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			leave: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300',
			holiday: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300',
			scheduled: 'bg-gray-100 text-gray-500 dark:bg-gray-800/40 dark:text-gray-400',
		};
		return map[status] || 'bg-gray-100 text-gray-700 dark:bg-gray-700 dark:text-gray-300';
	}

	function getStatusText(status: string) {
		const map: Record<string, string> = {
			present: 'Hadir',
			late: 'Terlambat',
			absent: 'Alpa / Tidak Hadir',
			leave: 'Cuti',
			holiday: 'Libur',
			scheduled: 'Belum Absen',
		};
		return map[status] || status;
	}
</script>

<svelte:head>
	<title>Absensi - HRMS</title>
</svelte:head>

<div class="w-full">
	<!-- Header -->
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-2 sm:gap-4 mb-4 sm:mb-6">
		<div>
			<h1 class="text-xl sm:text-2xl font-bold text-gray-900 dark:text-white tracking-tight">Absensi</h1>
			<p class="text-xs sm:text-sm text-gray-500 dark:text-gray-400 mt-0.5">Check-in / Check-out kehadiran hari ini</p>
		</div>
		<div class="flex items-center gap-2">
		</div>
	</div>

	<!-- Device Permission Status -->
	{#if !checkingPerms}
		<div class="mb-3 flex items-center gap-3 px-4 py-2 bg-white dark:bg-gray-800 rounded-xl border border-gray-100 dark:border-gray-700/50 shadow-sm">
			<div class="flex items-center gap-1.5 text-xs">
				<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.827 6.175A2.31 2.31 0 0 1 5.186 7.23c-.38.054-.757.112-1.134.175C2.999 7.58 2.25 8.507 2.25 9.574V18a2.25 2.25 0 0 0 2.25 2.25h15A2.25 2.25 0 0 0 21.75 18V9.574c0-1.067-.75-1.994-1.802-2.169a47.865 47.865 0 0 0-1.134-.175 2.31 2.31 0 0 1-1.64-1.055l-.822-1.316a2.192 2.192 0 0 0-1.736-1.039 48.774 48.774 0 0 0-5.232 0 2.192 2.192 0 0 0-1.736 1.039l-.821 1.316Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M16.5 12.75a4.5 4.5 0 1 1-9 0 4.5 4.5 0 0 1 9 0Z" /></svg>
				<span class="text-gray-500 dark:text-gray-400">Kamera:</span>
				<span class="font-medium {camGranted === true ? 'text-emerald-600 dark:text-emerald-400' : camGranted === false ? 'text-red-500 dark:text-red-400' : 'text-gray-400'}">
					{camGranted === true ? '✓ Aktif' : camGranted === false ? '✗ Diblokir' : '...'}
				</span>
			</div>
			<div class="w-px h-4 bg-gray-200 dark:bg-gray-700"></div>
			<div class="flex items-center gap-1.5 text-xs">
				<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
				<span class="text-gray-500 dark:text-gray-400">Lokasi:</span>
				<span class="font-medium {locGranted === true ? 'text-emerald-600 dark:text-emerald-400' : locGranted === false ? 'text-red-500 dark:text-red-400' : 'text-gray-400'}">
					{locGranted === true ? '✓ Aktif' : locGranted === false ? '✗ Diblokir' : '...'}
				</span>
			</div>
			{#if camGranted === false || locGranted === false}
				<div class="ml-auto">
					<a href="/dashboard/pengaturan" class="text-[10px] text-blue-600 dark:text-blue-400 hover:underline font-medium">Atur Izin</a>
				</div>
			{/if}
		</div>
	{/if}

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

	{#if locationWarning}
		<div class="mb-4 px-5 py-3.5 rounded-xl border bg-amber-50 dark:bg-amber-900/20 border-amber-200 dark:border-amber-800">
			<div class="flex items-start gap-2.5">
				<svg class="w-5 h-5 text-amber-500 shrink-0 mt-0.5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>
				<div>
					<p class="text-sm font-medium text-amber-800 dark:text-amber-200">⚠️ Peringatan Lokasi</p>
					<p class="text-xs text-amber-700 dark:text-amber-300 mt-1">{locationWarning}</p>
				</div>
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
					{:else if showCheckInConfirm || showCheckOutConfirm}
						{#if showCheckInConfirm}
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
						{:else}
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
					{/if}

					<p class="text-xs text-gray-400 text-center">
						{faceDetected ? 'Wajah terdeteksi, siap melakukan absensi.' : 'Arahkan wajah Anda ke kamera.'}
					</p>
				</div>
			</div>
		</div>
	{/if}

	<canvas bind:this={photoCanvasEl} class="hidden"></canvas>

	<!-- Day Off / Attendance cards conditional display -->
	{#if todayStatus?.is_day_off && !todayStatus?.has_checked_in}
		<div class="bg-red-50 dark:bg-red-900/10 rounded-2xl border border-red-100 dark:border-red-900/30 p-8 text-center mb-4 sm:mb-6 flex flex-col items-center justify-center min-h-[160px] shadow-sm">
			<div class="w-12 h-12 rounded-2xl bg-red-100/50 dark:bg-red-900/40 flex items-center justify-center text-red-600 dark:text-red-400 mb-3 shrink-0">
				<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
			</div>
			<p class="text-base font-bold text-red-800 dark:text-red-200">Hari ini adalah hari libur</p>
			<p class="text-xs text-red-500/70 dark:text-red-400/60 mt-1 font-medium">Tidak perlu melakukan absensi.</p>
		</div>
	{:else}
		<!-- Status & Action Cards -->
		<div class="grid grid-cols-1 md:grid-cols-2 gap-3 sm:gap-4 mb-4 sm:mb-6">
			<!-- Check-in Card -->
			<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-5 shadow-sm transition-all hover:shadow-md active:scale-[0.99] {todayStatus?.has_checked_in ? 'border-l-[3px] sm:border-l-4 border-l-emerald-500' : ''}">
				<div class="flex items-center justify-between mb-2 sm:mb-3">
					<div class="flex items-center gap-2.5 sm:gap-3">
						<div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center shrink-0 {todayStatus?.has_checked_in ? 'text-emerald-600' : 'text-blue-600'}">
							<svg class="w-[18px] h-[18px] sm:w-5 sm:h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
						</div>
						<div>
							<h3 class="text-sm sm:text-base font-semibold text-gray-900 dark:text-white">Check In</h3>
							<p class="text-[10px] sm:text-xs text-gray-500 dark:text-gray-400">Absensi Masuk</p>
						</div>
					</div>
					<div class="text-right shrink-0">
						<div class="text-base sm:text-lg font-bold {todayStatus?.has_checked_in ? 'text-emerald-600 dark:text-emerald-400' : 'text-gray-400'}">
							{todayStatus?.has_checked_in ? formatTime(todayStatus?.record?.check_in_time) : '--:--'}
						</div>
					</div>
				</div>
				{#if todayStatus?.has_checked_in && todayStatus?.record?.check_in_location_name}
				<div class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700 space-y-2">
					<div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
						<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
						<span>{todayStatus.record.check_in_location_name}</span>
					</div>
				</div>
				{/if}

				<button onclick={handleCheckIn}
					disabled={todayStatus?.has_checked_in || loadingStatus}
					class="w-full py-3 sm:py-2.5 rounded-xl text-sm font-semibold transition cursor-pointer active:scale-[0.97] {todayStatus?.has_checked_in ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed' : 'bg-blue-600 text-white hover:bg-blue-700'}">
					{todayStatus?.has_checked_in ? 'Sudah Check In' : 'Check In'}
				</button>
			</div>

			<!-- Check-out Card -->
			<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 p-4 sm:p-5 shadow-sm transition-all hover:shadow-md active:scale-[0.99] {todayStatus?.has_checked_out ? 'border-l-[3px] sm:border-l-4 border-l-amber-500' : ''}">
				<div class="flex items-center justify-between mb-2 sm:mb-3">
					<div class="flex items-center gap-2.5 sm:gap-3">
						<div class="w-9 h-9 sm:w-10 sm:h-10 rounded-xl bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center shrink-0 text-amber-600">
							<svg class="w-[18px] h-[18px] sm:w-5 sm:h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15" /></svg>
						</div>
						<div>
							<h3 class="text-sm sm:text-base font-semibold text-gray-900 dark:text-white">Check Out</h3>
							<p class="text-[10px] sm:text-xs text-gray-500 dark:text-gray-400">Absensi Pulang</p>
						</div>
					</div>
					<div class="text-right shrink-0">
						<div class="text-base sm:text-lg font-bold {todayStatus?.has_checked_out ? 'text-amber-600 dark:text-amber-400' : 'text-gray-400'}">
							{todayStatus?.has_checked_out ? formatTime(todayStatus?.record?.check_out_time) : '--:--'}
						</div>
					</div>
				</div>
				{#if todayStatus?.has_checked_out && todayStatus?.record?.check_out_location_name}
				<div class="mt-3 pt-3 border-t border-gray-100 dark:border-gray-700 space-y-2">
					<div class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
						<svg class="w-3.5 h-3.5 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
						<span>{todayStatus.record.check_out_location_name}</span>
					</div>
				</div>
				{/if}
				<button onclick={handleCheckOut}
					disabled={todayStatus?.has_checked_out || !todayStatus?.has_checked_in || loadingStatus || todayStatus?.has_checked_in === false}
					class="w-full py-3 sm:py-2.5 rounded-xl text-sm font-semibold transition cursor-pointer active:scale-[0.97] {todayStatus?.has_checked_out ? 'bg-gray-100 dark:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed' : 'bg-amber-600 text-white hover:bg-amber-700'}">
					{todayStatus?.has_checked_out ? 'Sudah Check Out' : 'Check Out'}
				</button>
			</div>
		</div>
	{/if}

	<!-- Schedule Info -->
	{#if todayStatus?.schedule_name}
		<div class="mb-4 sm:mb-6 px-4 sm:px-5 py-2.5 sm:py-3 bg-gray-50 dark:bg-gray-800/50 border border-gray-200 dark:border-gray-700 rounded-xl">
			<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-1 sm:gap-0 text-xs sm:text-sm">
				<div class="flex items-center gap-2">
					<svg class="w-3.5 h-3.5 sm:w-4 sm:h-4 text-gray-400 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					<span class="text-gray-500 dark:text-gray-400">Jadwal: <strong class="font-semibold text-gray-700 dark:text-gray-300">{todayStatus.schedule_name}</strong></span>
				</div>
				<span class="text-gray-500 dark:text-gray-400 ml-5 sm:ml-0">
					{todayStatus.schedule_start || '-'} - {todayStatus.schedule_end || '-'}
				</span>
			</div>
		</div>
	{/if}

	<!-- Attendance History -->
	<div class="bg-white dark:bg-gray-800 rounded-xl border border-gray-200 dark:border-gray-700 shadow-sm overflow-y-auto max-h-[calc(100dvh-24rem)] sm:max-h-[calc(100dvh-20rem)]">
		<div class="sticky top-0 z-10 bg-white dark:bg-gray-800 rounded-t-xl shadow-[0_1px_3px_-1px_rgba(0,0,0,0.08)]">
			<div class="px-4 sm:px-6 py-3 sm:py-4 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800/50 flex items-center justify-between">
				<h2 class="text-xs sm:text-sm font-semibold text-gray-900 dark:text-white">Riwayat Absensi</h2>
				<span class="text-[10px] sm:text-xs text-gray-400">Total {total} record</span>
			</div>

			<!-- Date Filters (Read Only) -->
			<div class="px-4 sm:px-6 py-2.5 sm:py-3 border-b border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/20 flex items-center gap-2">
				<span class="text-[10px] font-bold text-gray-400 dark:text-gray-500 uppercase tracking-wider">Periode Aktif:</span>
				{#if filterDateFrom && filterDateTo}
					<span class="text-xs font-semibold text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/30 px-2.5 py-1 rounded-lg">
						📅 {new Date(filterDateFrom).toLocaleDateString('id-ID', { day: 'numeric', month: 'short' })} - {new Date(filterDateTo).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })}
					</span>
				{/if}
			</div>
		</div>

		{#if loadingHistory}
			<div class="flex items-center justify-center py-8 sm:py-10">
				<div class="animate-spin h-5 w-5 sm:h-6 sm:w-6 border-2 border-blue-600 border-t-transparent rounded-full"></div>
			</div>
		{:else if timesheetItems.length === 0}
			<div class="p-5 sm:p-6 text-center">
				<svg class="w-8 h-8 sm:w-10 sm:h-10 text-gray-300 dark:text-gray-600 mx-auto mb-2" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
				<p class="text-xs sm:text-sm text-gray-400">Belum ada riwayat absensi</p>
			</div>
		{:else}
			<div class="p-3 sm:p-4 space-y-2.5">
				{#each timesheetItems as item (item)}
					<div onclick={() => openDetail(item)} class="group bg-white dark:bg-gray-800/50 rounded-xl border border-gray-100 dark:border-gray-700/50 hover:border-gray-200 dark:hover:border-gray-600 transition-all duration-200 overflow-hidden {item.is_virtual ? 'opacity-80' : 'hover:shadow-md active:shadow-sm cursor-pointer active:scale-[0.99]'}">
						<!-- Colored top bar based on status -->
						<div class="h-1 {item.status === 'late' || item.status === 'terlambat' ? 'bg-red-500' : item.status === 'present' || item.status === 'hadir' ? 'bg-emerald-500' : item.status === 'leave' || item.status === 'cuti' ? 'bg-blue-500' : item.status === 'sick' || item.status === 'sakit' ? 'bg-purple-500' : item.status === 'holiday' || item.status === 'libur' ? 'bg-red-500' : item.status === 'scheduled' ? 'bg-gray-200 dark:bg-gray-700' : 'bg-gray-400'}"></div>
						
						<div class="p-3.5 sm:p-4">
							<!-- Top row: Date + Status -->
							<div class="flex items-center justify-between mb-3">
								<div class="flex items-center gap-2.5">
									<div class="w-9 h-9 rounded-xl bg-gray-50 dark:bg-gray-800 flex items-center justify-center shrink-0">
										<svg class="w-4 h-4 text-gray-500 dark:text-gray-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6.75 3v2.25M17.25 3v2.25M3 18.75V7.5a2.25 2.25 0 0 1 2.25-2.25h13.5A2.25 2.25 0 0 1 21 7.5v11.25m-18 0A2.25 2.25 0 0 0 5.25 21h13.5A2.25 2.25 0 0 0 21 18.75m-18 0v-7.5A2.25 2.25 0 0 1 5.25 9h13.5A2.25 2.25 0 0 1 21 11.25v7.5" /></svg>
									</div>
									<div>
										<p class="text-sm font-semibold text-gray-900 dark:text-white">{formatDate(item.date)}</p>
										{#if item.day_name}
											<p class="text-[10px] text-gray-400 dark:text-gray-500">
												{item.day_name.trim()}
												{#if item.holiday_name}
													<span class="text-red-500 dark:text-red-400 font-semibold ml-1">({item.holiday_name})</span>
												{/if}
											</p>
										{/if}
									</div>
								</div>
								<div class="flex items-center gap-2">
									{#if isLateRecord(item)}
										<span class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-red-100 dark:bg-red-900/30 text-red-600 dark:text-red-400">
											{item.late_minutes && item.late_minutes > 0 ? `Telat ${item.late_minutes}m` : 'Telat'}
										</span>
									{/if}
									<span class="px-2 py-0.5 rounded-full text-[10px] font-semibold {getStatusBadge(item.status)}">{getStatusText(item.status)}</span>
									{#if !item.is_virtual}
										<svg class="w-3.5 h-3.5 text-gray-300 dark:text-gray-600 group-hover:text-gray-400 transition-colors" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 8.25l-7.5 7.5-7.5-7.5" /></svg>
									{/if}
								</div>
							</div>

							<!-- Time row -->
							<div class="grid grid-cols-2 gap-2">
								<div class="flex items-center gap-2 p-2.5 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
									<div class="w-7 h-7 rounded-lg bg-blue-50 dark:bg-blue-900/30 flex items-center justify-center shrink-0">
										<svg class="w-3.5 h-3.5 text-blue-600 dark:text-blue-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
									</div>
									<div class="min-w-0">
										<p class="text-[10px] text-gray-400 dark:text-gray-500">Check-In</p>
										<p class="text-sm font-bold tabular-nums {isLateRecord(item) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(item.check_in_time)}</p>
									</div>
								</div>
								<div class="flex items-center gap-2 p-2.5 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
									<div class="w-7 h-7 rounded-lg bg-amber-50 dark:bg-amber-900/30 flex items-center justify-center shrink-0">
										<svg class="w-3.5 h-3.5 text-amber-600 dark:text-amber-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M8.25 9V5.25A2.25 2.25 0 0 1 10.5 3h6a2.25 2.25 0 0 1 2.25 2.25v13.5A2.25 2.25 0 0 1 16.5 21h-6a2.25 2.25 0 0 1-2.25-2.25V15m-3 0-3-3m0 0 3-3m-3 3H15" /></svg>
									</div>
									<div class="min-w-0">
										<p class="text-[10px] text-gray-400 dark:text-gray-500">Check-Out</p>
										<p class="text-sm font-bold tabular-nums {isEarlyLeaveRecord(item) ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'}">{formatTime(item.check_out_time)}</p>
									</div>
								</div>
							</div>

							<!-- Location row -->
							{#if item.check_in_location_name || item.check_out_location_name}
								<div class="mt-2 flex items-center gap-1.5 text-[10px] text-gray-400 dark:text-gray-500">
									<svg class="w-3 h-3 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
									<span class="truncate">{item.check_in_location_name || item.check_out_location_name || '-'}</span>
								</div>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</div>
</div>
