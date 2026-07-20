<script lang="ts">
	import { onMount } from 'svelte';
	import { company as companyApi } from '$lib/api.js';
	import PulseLoader from '$lib/components/PulseLoader.svelte';

	type BPJSComponent = {
		enabled?: boolean;
		employee_rate?: number;
		company_rate?: number;
		ceiling?: number;
	};

	type BPJSConfig = {
		kesehatan?: BPJSComponent;
		jht?: BPJSComponent;
		jp?: BPJSComponent;
		jkk?: BPJSComponent;
		jkm?: BPJSComponent;
	};

	type CompanySettings = {
		id: string;
		name: string;
		legal_name?: string;
		address?: string;
		city?: string;
		province?: string;
		postal_code?: string;
		phone?: string;
		email?: string;
		npwp?: string;
		bpjs_ks_number?: string;
		bpjs_jht_number?: string;
		bpjs_jp_number?: string;
		bpjs_jkk_rate?: number;
		hr_settings?: {
			bpjs?: BPJSConfig;
			face_match_threshold?: number;
			cutoff_start_day?: number;
			cutoff_end_day?: number;
		};
	};

	let settings = $state<CompanySettings | null>(null);
let isLoading = $state(true);
let isSaving = $state(false);
let loadError = $state('');
	let errorMessage = $state('');
	let successMessage = $state('');
	let activeTab = $state<'company' | 'bpjs' | 'attendance'>('bpjs');

	let edit = $state({
		name: '',
		legal_name: '',
		address: '',
		city: '',
		province: '',
		phone: '',
		email: '',
		npwp: '',
		bpjs_ks_number: '',
		bpjs_jht_number: '',
		bpjs_jp_number: '',
		bpjs_jkk_rate: 0.54,
		bpjs: {
			kesehatan: { enabled: true, employee_rate: 1, company_rate: 4, ceiling: 12000000 },
			jht: { enabled: true, employee_rate: 2, company_rate: 3.7, ceiling: 0 },
			jp: { enabled: true, employee_rate: 1, company_rate: 2, ceiling: 10000000 },
			jkk: { enabled: true, company_rate: 0 },
			jkm: { enabled: true, company_rate: 0.3, ceiling: 0 },
		} as BPJSConfig,
		face_match_threshold: 0.6,
		cutoff_start_day: 26,
		cutoff_end_day: 25,
	});

	onMount(async () => {
		try {
			const res = await companyApi.getSettings() as { success: boolean; data?: CompanySettings };
			if (res.data) {
				settings = res.data;
				edit.name = res.data.name || '';
				edit.legal_name = res.data.legal_name || '';
				edit.address = res.data.address || '';
				edit.city = res.data.city || '';
				edit.province = res.data.province || '';
				edit.phone = res.data.phone || '';
				edit.email = res.data.email || '';
				edit.npwp = res.data.npwp || '';
				edit.bpjs_ks_number = res.data.bpjs_ks_number || '';
				edit.bpjs_jht_number = res.data.bpjs_jht_number || '';
				edit.bpjs_jp_number = res.data.bpjs_jp_number || '';
				edit.bpjs_jkk_rate = res.data.bpjs_jkk_rate || 0.54;

				// Load face_match_threshold from hr_settings
				if (res.data.hr_settings?.face_match_threshold != null) {
					edit.face_match_threshold = res.data.hr_settings.face_match_threshold;
				}

				// Load cutoff days from hr_settings
				if (res.data.hr_settings?.cutoff_start_day != null) {
					edit.cutoff_start_day = res.data.hr_settings.cutoff_start_day;
				} else {
					edit.cutoff_start_day = 26;
				}
				if (res.data.hr_settings?.cutoff_end_day != null) {
					edit.cutoff_end_day = res.data.hr_settings.cutoff_end_day;
				} else {
					edit.cutoff_end_day = 25;
				}

				if (res.data.hr_settings?.bpjs) {
					const b = res.data.hr_settings.bpjs;
					if (b.kesehatan) {
						edit.bpjs.kesehatan = {
							enabled: b.kesehatan.enabled ?? true,
							employee_rate: b.kesehatan.employee_rate != null ? b.kesehatan.employee_rate * 100 : 1,
							company_rate: b.kesehatan.company_rate != null ? b.kesehatan.company_rate * 100 : 4,
							ceiling: b.kesehatan.ceiling ?? 12000000,
						};
					}
					if (b.jht) {
						edit.bpjs.jht = {
							enabled: b.jht.enabled ?? true,
							employee_rate: b.jht.employee_rate != null ? b.jht.employee_rate * 100 : 2,
							company_rate: b.jht.company_rate != null ? b.jht.company_rate * 100 : 3.7,
						};
					}
					if (b.jp) {
						edit.bpjs.jp = {
							enabled: b.jp.enabled ?? true,
							employee_rate: b.jp.employee_rate != null ? b.jp.employee_rate * 100 : 1,
							company_rate: b.jp.company_rate != null ? b.jp.company_rate * 100 : 2,
							ceiling: b.jp.ceiling ?? 10000000,
						};
					}
					if (b.jkk) {
						edit.bpjs.jkk = {
							enabled: b.jkk.enabled ?? true,
						};
					}
					if (b.jkm) {
						edit.bpjs.jkm = {
							enabled: b.jkm.enabled ?? true,
							company_rate: b.jkm.company_rate != null ? b.jkm.company_rate * 100 : 0.3,
						};
					}
				}
			}
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal memuat pengaturan';
		} finally {
			isLoading = false;
		}
	});

	async function saveSettings() {
		isSaving = true;
		errorMessage = '';
		successMessage = '';
		try {
			const bpjsPayload: BPJSConfig = {
				kesehatan: {
					enabled: edit.bpjs.kesehatan?.enabled ?? true,
					employee_rate: (edit.bpjs.kesehatan?.employee_rate ?? 1) / 100,
					company_rate: (edit.bpjs.kesehatan?.company_rate ?? 4) / 100,
					ceiling: edit.bpjs.kesehatan?.ceiling ?? 12000000,
				},
				jht: edit.bpjs.jht?.enabled ? {
					enabled: true,
					employee_rate: (edit.bpjs.jht?.employee_rate ?? 2) / 100,
					company_rate: (edit.bpjs.jht?.company_rate ?? 3.7) / 100,
				} : { enabled: false },
				jp: edit.bpjs.jp?.enabled ? {
					enabled: true,
					employee_rate: (edit.bpjs.jp?.employee_rate ?? 1) / 100,
					company_rate: (edit.bpjs.jp?.company_rate ?? 2) / 100,
					ceiling: edit.bpjs.jp?.ceiling ?? 10000000,
				} : { enabled: false },
				jkk: edit.bpjs.jkk?.enabled ? { enabled: true } : { enabled: false },
				jkm: edit.bpjs.jkm?.enabled ? {
					enabled: true,
					company_rate: (edit.bpjs.jkm?.company_rate ?? 0.3) / 100,
				} : { enabled: false },
			};

			const payload = {
				name: edit.name || undefined,
				legal_name: edit.legal_name || undefined,
				address: edit.address || undefined,
				city: edit.city || undefined,
				province: edit.province || undefined,
				phone: edit.phone || undefined,
				email: edit.email || undefined,
				npwp: edit.npwp || undefined,
				bpjs_ks_number: edit.bpjs_ks_number || undefined,
				bpjs_jht_number: edit.bpjs_jht_number || undefined,
				bpjs_jp_number: edit.bpjs_jp_number || undefined,
				bpjs_jkk_rate: edit.bpjs_jkk_rate || undefined,
				hr_settings: {
					bpjs: Object.keys(bpjsPayload).length > 0 ? bpjsPayload : undefined,
					face_match_threshold: edit.face_match_threshold,
					cutoff_start_day: edit.cutoff_start_day,
					cutoff_end_day: edit.cutoff_end_day,
				},
			};

			const res = await companyApi.updateSettings(payload) as { success: boolean; data?: CompanySettings };
			if (res.data) {
				settings = res.data;
				successMessage = 'Pengaturan berhasil disimpan';
				setTimeout(() => successMessage = '', 3000);
			}
		} catch (err: unknown) {
			errorMessage = (err as { message?: string }).message || 'Gagal menyimpan pengaturan';
		} finally {
			isSaving = false;
		}
	}

	function formatCurrency(val: number): string {
		return new Intl.NumberFormat('id-ID', { style: 'currency', currency: 'IDR', minimumFractionDigits: 0 }).format(val);
	}

	function bpjsHelp(component: string): string {
		const help: Record<string, string> = {
			kesehatan: 'Iuran Jaminan Kesehatan untuk layanan BPJS Kesehatan.',
			jht: 'Jaminan Hari Tua — tabungan hari tua yang dibayarkan perusahaan & pekerja.',
			jp: 'Jaminan Pensiun — program pensiun BPJS Ketenagakerjaan.',
			jkk: 'Jaminan Kecelakaan Kerja — rate berdasarkan tingkat risiko perusahaan.',
			jkm: 'Jaminan Kematian — santunan untuk ahli waris jika pekerja meninggal.',
		};
		return help[component] || '';
	}
</script>

<div class="w-full">
	<div class="flex items-center gap-4 mb-8">
		<div class="w-12 h-12 bg-[#1A56DB] rounded-2xl flex items-center justify-center text-white font-bold text-lg shrink-0">
			<svg class="w-6 h-6" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.325.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 0 1 1.37.49l1.296 2.247a1.125 1.125 0 0 1-.26 1.431l-1.003.827c-.293.241-.438.613-.43.992a7.723 7.723 0 0 1 0 .255c-.008.378.137.75.43.991l1.004.827c.424.35.534.955.26 1.43l-1.298 2.247a1.125 1.125 0 0 1-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.47 6.47 0 0 1-.22.128c-.331.183-.581.495-.644.869l-.213 1.281c-.09.543-.56.94-1.11.94h-2.594c-.55 0-1.019-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 0 1-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 0 1-1.369-.49l-1.297-2.247a1.125 1.125 0 0 1 .26-1.431l1.004-.827c.292-.24.437-.613.43-.991a6.932 6.932 0 0 1 0-.255c.007-.38-.138-.751-.43-.992l-1.004-.827a1.125 1.125 0 0 1-.26-1.43l1.297-2.247a1.125 1.125 0 0 1 1.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.086.22-.128.332-.183.582-.495.644-.869l.214-1.28Z" />
			</svg>
		</div>
		<div>
			<h1 class="text-2xl font-bold text-gray-900 dark:text-gray-100">Pengaturan Perusahaan</h1>
			<p class="text-sm text-gray-500 dark:text-gray-400">Konfigurasi BPJS, profil perusahaan, dan pengaturan HR</p>
		</div>
	</div>

	{#if isLoading}
		<PulseLoader variant="card" count={2} />
	{:else}
		<div class="flex flex-wrap sm:flex-nowrap sm:overflow-x-auto gap-1 mb-6 bg-gray-100 dark:bg-gray-800 rounded-xl p-1 w-full sm:w-fit hide-scrollbar">
			<button onclick={() => activeTab = 'bpjs'}
				class="px-5 py-2 rounded-lg text-sm font-medium transition cursor-pointer {activeTab === 'bpjs' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500 hover:text-gray-700'}">
				Konfigurasi BPJS
			</button>
			<button onclick={() => activeTab = 'company'}
				class="px-5 py-2 rounded-lg text-sm font-medium transition cursor-pointer {activeTab === 'company' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500 hover:text-gray-700'}">
				Profil Perusahaan
			</button>
			<button onclick={() => activeTab = 'attendance'}
				class="px-5 py-2 rounded-lg text-sm font-medium transition cursor-pointer {activeTab === 'attendance' ? 'bg-white dark:bg-gray-700 text-gray-900 dark:text-gray-100 shadow-sm' : 'text-gray-500 hover:text-gray-700'}">
				Konfigurasi Absensi
			</button>
		</div>

		{#if successMessage}
			<div class="mb-4 bg-emerald-50 dark:bg-emerald-900/20 border border-emerald-200 dark:border-emerald-800 rounded-xl px-5 py-3.5">
				<div class="flex items-center gap-2.5">
					<svg class="w-5 h-5 text-emerald-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
					<p class="text-sm font-medium text-emerald-800 dark:text-emerald-200">{successMessage}</p>
				</div>
			</div>
		{/if}
		{#if errorMessage}
			<div class="mb-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-xl px-5 py-3.5">
				<div class="flex items-center gap-2.5">
					<svg class="w-5 h-5 text-red-500 shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>
					<p class="text-sm font-medium text-red-800 dark:text-red-200">{errorMessage}</p>
				</div>
			</div>
		{/if}

		<form onsubmit={(e) => { e.preventDefault(); saveSettings(); }}>
			{#if activeTab === 'bpjs'}
				<div class="space-y-6">
					<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm hover:shadow-md transition-all duration-200">
						<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gradient-to-r from-blue-50 to-white dark:from-gray-800 dark:to-gray-900">
							<div class="flex items-center gap-3">
								<svg class="w-5 h-5 text-blue-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 0 0 6 16.5h2.25M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0 1 18 16.5h-2.25m-7.5 0h7.5m-7.5 0-1 3m8.5-3 1 3m0 0 .5 1.5m-.5-1.5h-9.5m0 0-.5 1.5m.75-9 3-3 2.148 2.148A12.061 12.061 0 0 1 16.5 7.605" /></svg>
								<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Konfigurasi BPJS</h2>
							</div>
							<p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-8">Atur persentase iuran BPJS Kesehatan, JHT, JP, JKK, dan JKM sesuai kebijakan perusahaan.</p>
						</div>

						<div class="p-6 space-y-6">
							<!-- BPJS Kesehatan -->
						<!-- svelte-ignore a11y_label_has_associated_control -->
						<div class="bg-blue-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-blue-100 dark:border-gray-700">
							<div class="flex items-center justify-between mb-4">
								<div>
									<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">BPJS Kesehatan</h3>
									<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{bpjsHelp('kesehatan')}</p>
								</div>
								<label class="relative inline-flex items-center cursor-pointer">
									<input type="checkbox" bind:checked={edit.bpjs.kesehatan!.enabled} class="sr-only peer" />
									<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"></div>
								</label>
							</div>
								{#if edit.bpjs.kesehatan?.enabled}
									<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Karyawan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.kesehatan!.employee_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
										</div>
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Perusahaan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.kesehatan!.company_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
										</div>
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Plafon (Rp)</label>
											<input type="number" step="100000" min="0" bind:value={edit.bpjs.kesehatan!.ceiling}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
										</div>
									</div>
									<div class="mt-3 text-xs text-gray-400">
										Default: Karyawan 1%, Perusahaan 4%, Plafon {formatCurrency(12000000)}
									</div>
								{/if}
							</div>

							<!-- BPJS JHT -->
						<!-- svelte-ignore a11y_label_has_associated_control -->
						<div class="bg-purple-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-purple-100 dark:border-gray-700">
							<div class="flex items-center justify-between mb-4">
								<div>
									<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">BPJS JHT (Jaminan Hari Tua)</h3>
									<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{bpjsHelp('jht')}</p>
								</div>
								<label class="relative inline-flex items-center cursor-pointer">
									<input type="checkbox" bind:checked={edit.bpjs.jht!.enabled} class="sr-only peer" />
									<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-purple-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-purple-600"></div>
								</label>
							</div>
								{#if edit.bpjs.jht?.enabled}
									<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Karyawan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.jht!.employee_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-purple-500" />
										</div>
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Perusahaan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.jht!.company_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-purple-500" />
										</div>
									</div>
									<div class="mt-3 text-xs text-gray-400">Default: Karyawan 2%, Perusahaan 3,7%</div>
								{/if}
							</div>

							<!-- BPJS JP -->
							<!-- svelte-ignore a11y_label_has_associated_control -->
							<div class="bg-amber-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-amber-100 dark:border-gray-700">
								<div class="flex items-center justify-between mb-4">
									<div>
										<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">BPJS JP (Jaminan Pensiun)</h3>
										<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{bpjsHelp('jp')}</p>
									</div>
									<label class="relative inline-flex items-center cursor-pointer">
										<input type="checkbox" bind:checked={edit.bpjs.jp!.enabled} class="sr-only peer" />
										<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-amber-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-amber-600"></div>
									</label>
								</div>
								{#if edit.bpjs.jp?.enabled}
									<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Karyawan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.jp!.employee_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-amber-500 focus:border-amber-500" />
										</div>
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Perusahaan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.jp!.company_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-amber-500 focus:border-amber-500" />
										</div>
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Plafon (Rp)</label>
											<input type="number" step="100000" min="0" bind:value={edit.bpjs.jp!.ceiling}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-amber-500 focus:border-amber-500" />
										</div>
									</div>
									<div class="mt-3 text-xs text-gray-400">
										Default: Karyawan 1%, Perusahaan 2%, Plafon {formatCurrency(10000000)}
									</div>
								{/if}
							</div>

							<!-- BPJS JKK & JKM -->
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div class="bg-rose-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-rose-100 dark:border-gray-700">
									<div class="flex items-center justify-between mb-3">
										<div>
											<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">BPJS JKK</h3>
											<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{bpjsHelp('jkk')}</p>
										</div>
										<label class="relative inline-flex items-center cursor-pointer">
											<input type="checkbox" bind:checked={edit.bpjs.jkk!.enabled} class="sr-only peer" />
											<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-rose-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-rose-600"></div>
										</label>
									</div>
									{#if edit.bpjs.jkk?.enabled}
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Rate Risiko (%)</label>
											<input type="number" step="0.01" min="0.24" max="1.74" bind:value={edit.bpjs_jkk_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-rose-500 focus:border-rose-500" />
											<p class="text-xs text-gray-400 mt-1">Berdasarkan tingkat risiko perusahaan (0.24% - 1.74%). Default 0.54%.</p>
										</div>
									{/if}
								</div>

								<div class="bg-teal-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-teal-100 dark:border-gray-700">
									<div class="flex items-center justify-between mb-3">
										<div>
											<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">BPJS JKM</h3>
											<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">{bpjsHelp('jkm')}</p>
										</div>
										<label class="relative inline-flex items-center cursor-pointer">
											<input type="checkbox" bind:checked={edit.bpjs.jkm!.enabled} class="sr-only peer" />
											<div class="w-9 h-5 bg-gray-200 peer-focus:outline-none peer-focus:ring-2 peer-focus:ring-teal-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-teal-600"></div>
										</label>
									</div>
									{#if edit.bpjs.jkm?.enabled}
										<div>
											<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Perusahaan (%)</label>
											<input type="number" step="0.01" min="0" max="100" bind:value={edit.bpjs.jkm!.company_rate}
												class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-teal-500 focus:border-teal-500" />
										</div>
									{/if}
								</div>
							</div>

							<!-- BPJS Numbers -->
							<div class="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-5 border border-gray-200 dark:border-gray-700">
								<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-4">Nomor BPJS Perusahaan</h3>
								<div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
									<div>
										<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">BPJS Kesehatan</label>
										<input type="text" bind:value={edit.bpjs_ks_number} placeholder="00123456789"
											class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
									</div>
									<div>
										<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">BPJS JHT</label>
										<input type="text" bind:value={edit.bpjs_jht_number} placeholder="00987654321"
											class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-purple-500 focus:border-purple-500" />
									</div>
									<div>
										<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">BPJS JP</label>
										<input type="text" bind:value={edit.bpjs_jp_number} placeholder="00543218765"
											class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-amber-500 focus:border-amber-500" />
									</div>
								</div>
							</div>

						</div>
					</div>
				</div>

			{:else if activeTab === 'company'}
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm hover:shadow-md transition-all duration-200">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gradient-to-r from-gray-50 to-white dark:from-gray-800 dark:to-gray-900">
						<div class="flex items-center gap-3">
							<svg class="w-5 h-5 text-gray-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3.75h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Zm0 3h.008v.008h-.008v-.008Z" /></svg>
							<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Profil Perusahaan</h2>
						</div>
					</div>
					<div class="p-6 space-y-5">
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Nama Perusahaan</label>
								<input type="text" bind:value={edit.name}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Nama Legal</label>
								<input type="text" bind:value={edit.legal_name}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white/95 dark:bg-gray-800/95 backdrop-blur-sm text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">NPWP</label>
								<input type="text" bind:value={edit.npwp}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Email</label>
								<input type="email" bind:value={edit.email}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Telepon</label>
								<input type="text" bind:value={edit.phone}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Kota</label>
								<input type="text" bind:value={edit.city}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
							<div>
								<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Provinsi</label>
								<input type="text" bind:value={edit.province}
									class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500" />
							</div>
						</div>
						<div>
							<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Alamat</label>
							<textarea bind:value={edit.address} rows="3"
								class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-blue-500 focus:border-blue-500"></textarea>
						</div>
					</div>
				</div>

			{:else if activeTab === 'attendance'}
				<div class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 overflow-hidden shadow-sm hover:shadow-md transition-all duration-200">
					<div class="px-6 py-4 border-b border-gray-200 dark:border-gray-800 bg-gradient-to-r from-teal-50 to-white dark:from-gray-800 dark:to-gray-900">
						<div class="flex items-center gap-3">
							<svg class="w-5 h-5 text-teal-600" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z" /></svg>
							<h2 class="text-base font-semibold text-gray-900 dark:text-gray-100">Konfigurasi Absensi</h2>
						</div>
						<p class="text-xs text-gray-500 dark:text-gray-400 mt-1 ml-8">Atur threshold verifikasi wajah dan pengaturan absensi lainnya.</p>
					</div>

					<div class="p-6">
						<div class="bg-teal-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-teal-100 dark:border-gray-700">
							<div class="flex items-center justify-between mb-6">
								<div>
									<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Threshold Verifikasi Wajah</h3>
									<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Batas kemiripan wajah untuk verifikasi absensi. Semakin kecil nilai, semakin ketat verifikasi.</p>
								</div>
							</div>

							<div class="px-2">
								<input
									type="range"
									min="0.1"
									max="1.0"
									step="0.05"
									bind:value={edit.face_match_threshold}
									class="w-full h-2 bg-gray-200 dark:bg-gray-700 rounded-lg appearance-none cursor-pointer accent-teal-600"
								/>
								<div class="flex items-center justify-between mt-2">
									<div class="flex items-center gap-2">
										<span class="text-xs text-gray-400">Ketat</span>
										<div class="w-20 h-1.5 rounded-full bg-gradient-to-r from-teal-600 via-amber-400 to-red-400"></div>
										<span class="text-xs text-gray-400">Longgar</span>
									</div>
									<span class="text-sm font-semibold text-teal-700 dark:text-teal-300 tabular-nums">{edit.face_match_threshold.toFixed(2)}</span>
								</div>
							</div>

							<div class="mt-6 grid grid-cols-1 sm:grid-cols-3 gap-4">
								<div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 text-center">
									<div class="text-xs text-gray-400 mb-1">Sangat Ketat</div>
									<div class="text-sm font-semibold text-gray-900 dark:text-gray-100">0.10 - 0.30</div>
									<div class="text-xs text-teal-600 dark:text-teal-400 mt-1">Cocok untuk keamanan tinggi</div>
								</div>
								<div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 text-center">
									<div class="text-xs text-gray-400 mb-1">Normal</div>
									<div class="text-sm font-semibold text-gray-900 dark:text-gray-100">0.35 - 0.60</div>
									<div class="text-xs text-amber-600 dark:text-amber-400 mt-1">Default: 0.60 — rekomendasi</div>
								</div>
								<div class="bg-white dark:bg-gray-800 rounded-lg p-4 border border-gray-200 dark:border-gray-700 text-center">
									<div class="text-xs text-gray-400 mb-1">Longgar</div>
									<div class="text-sm font-semibold text-gray-900 dark:text-gray-100">0.65 - 1.00</div>
									<div class="text-xs text-red-600 dark:text-red-400 mt-1">Risiko false positive tinggi</div>
								</div>
							</div>
						</div>

						<!-- Periode Absensi (Buka & Tutup Buku) -->
						<div class="bg-indigo-50/30 dark:bg-gray-800/50 rounded-xl p-5 border border-indigo-100 dark:border-gray-700 mt-6">
							<div class="mb-4">
								<h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">Siklus Periode Absensi (Buka & Tutup Buku)</h3>
								<p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">Tentukan tanggal mulai (buka buku) dan tanggal selesai (tutup buku/cutoff) untuk perhitungan absensi bulanan.</p>
							</div>

							<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
								<div>
									<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Tanggal Buka Buku (Mulai Siklus)</label>
									<select bind:value={edit.cutoff_start_day}
										class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500">
										{#each Array.from({ length: 31 }, (_, i) => i + 1) as day}
											<option value={day}>Tanggal {day}</option>
										{/each}
									</select>
									<span class="text-[10px] text-gray-400 mt-1 block">Biasanya tanggal 26 bulan sebelumnya.</span>
								</div>
								<div>
									<label class="block text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">Tanggal Tutup Buku (Cutoff/Akhir Siklus)</label>
									<select bind:value={edit.cutoff_end_day}
										class="w-full px-3 py-2 rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 text-sm text-gray-900 dark:text-gray-100 focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500">
										{#each Array.from({ length: 31 }, (_, i) => i + 1) as day}
											<option value={day}>Tanggal {day}</option>
										{/each}
									</select>
									<span class="text-[10px] text-gray-400 mt-1 block">Biasanya tanggal 25 bulan berjalan.</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			{/if}

			<div class="mt-6 flex items-center justify-between">
				<p class="text-xs text-gray-400 dark:text-gray-500">Perubahan akan diterapkan pada perhitungan payroll berikutnya.</p>
				<button type="submit" disabled={isSaving}
					class="inline-flex items-center gap-2 px-6 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-blue-700 transition disabled:opacity-50 cursor-pointer">
					{#if isSaving}
						<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
						Menyimpan...
					{:else}
						<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>
						Simpan Pengaturan
					{/if}
				</button>
			</div>
		</form>
	{/if}
</div>
