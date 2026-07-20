<script lang="ts">
/* eslint-disable @typescript-eslint/no-explicit-any */
	import { onMount, onDestroy, tick } from 'svelte';
	import { attendanceLocations as api } from '$lib/api.js';
	import { hasPermission } from '$lib/permissions.js';
	import PullToRefresh from '$lib/components/PullToRefresh.svelte';
	import PulseLoader from '$lib/components/PulseLoader.svelte';
	import { AnimatedPresence } from '$lib';
import MobileCard from '$lib/components/MobileCard.svelte';
import EmptyState from '$lib/components/EmptyState.svelte';
	import { getAvatarTheme, getInitials } from '$lib/avatar-theme.js';
	import type { GridApi, ColDef, GridOptions } from 'ag-grid-community';
	import { getAgGrid } from '$lib/ag-grid.js';
	import type { ApiResponse, AgGridCellParams, AgGridValueParams } from '$lib/types.js';

	/** @type {any} */
	let L: any = null;
	let leafletReady = $state(false);
	type AttendanceLocation = {
		id: string;
		name: string;
		address: string;
		latitude: number;
		longitude: number;
		radius_meters: number;
		is_active: boolean;
		created_at: string;
	};

	type Form = {
		name: string;
		address: string;
		latitude: number;
		longitude: number;
		radius_meters: number;
	};

	let items = $state<AttendanceLocation[]>([]);
	let total = $state(0);
	let page = $state(1);
	let perPage = $state(25);
	let totalPages = $state(0);
	let searchQuery = $state('');
	let isLoading = $state(true);
	let errorMessage = $state('');

	let showForm = $state(false);
	let formTitle = $state('');
	let editingId = $state<string | null>(null);
	let form = $state<Form>({ name: '', address: '', latitude: 0, longitude: 0, radius_meters: 100 });
	let formError = $state('');
	let isSaving = $state(false);

	let showDeleteConfirm = $state(false);
	let showCancelConfirm = $state(false);
	let deletingId = $state<string | null>(null);
	let deletingName = $state('');

	// AG Grid
	let gridContainer = $state<HTMLDivElement>(undefined!);
	let gridApi: GridApi | null = null;
	let agGridModule: typeof import('ag-grid-community') | null = null;

	// Leaflet Map
	let mapContainer: HTMLDivElement;
	let map: any = null;
	let mapMarker: any = null;
	let mapCircle: any = null;

	let mapLayerType = $state<'street' | 'hybrid'>('hybrid');
	let streetLayer: any;
	let hybridLayer: any;

	async function initMap() {
		if (map) {
			map.invalidateSize();
			return;
		}
		if (!leafletReady) {
			const leafletModule = await import('leaflet');
			L = leafletModule.default;
			// Fix Leaflet default icon path for bundlers via CDN
			delete (L.Icon.Default.prototype as any)._getIconUrl;
			L.Icon.Default.mergeOptions({
				iconUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon.png',
				iconRetinaUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-icon-2x.png',
				shadowUrl: 'https://unpkg.com/leaflet@1.9.4/dist/images/marker-shadow.png',
			});
			leafletReady = true;
		}
		if (!mapContainer) return;
		
		streetLayer = L.tileLayer('https://mt1.google.com/vt/lyrs=m&x={x}&y={y}&z={z}', { maxZoom: 20 });
		hybridLayer = L.tileLayer('https://mt1.google.com/vt/lyrs=y&x={x}&y={y}&z={z}', { maxZoom: 20 });
		
		map = L.map(mapContainer, { 
			zoomControl: false, // We'll disable default to add our own or just let leaflet put it top left by default
			attributionControl: false,
			layers: [hybridLayer] // Default to hybrid/structure view
		}).setView([-2.5489, 118.0149], 5);
		
		L.control.zoom({ position: 'bottomright' }).addTo(map);

		map.on('click', (e: any) => {
			form.latitude = parseFloat(e.latlng.lat.toFixed(6));
			form.longitude = parseFloat(e.latlng.lng.toFixed(6));
		});
		setTimeout(() => map?.invalidateSize(), 50);
	}

	function toggleMapLayer() {
		if (!map || !streetLayer || !hybridLayer) return;
		if (mapLayerType === 'hybrid') {
			map.removeLayer(hybridLayer);
			map.addLayer(streetLayer);
			mapLayerType = 'street';
		} else {
			map.removeLayer(streetLayer);
			map.addLayer(hybridLayer);
			mapLayerType = 'hybrid';
		}
	}

	function destroyMap() {
		map?.off();
		map?.remove();
		map = null;
		mapMarker = null;
		mapCircle = null;
	}

	function updateMapOverlay() {
		if (!map) return;
		mapMarker?.remove();
		mapCircle?.remove();
		if (!form.latitude && !form.longitude) return;
		const latlng: any = [form.latitude, form.longitude];
		mapMarker = L.marker(latlng, { draggable: true }).addTo(map);
		mapMarker.on('dragend', () => {
			const pos = mapMarker!.getLatLng();
			form.latitude = parseFloat(pos.lat.toFixed(6));
			form.longitude = parseFloat(pos.lng.toFixed(6));
		});
		mapCircle = L.circle(latlng, { radius: form.radius_meters || 100, color: '#1A56DB', fillColor: '#1A56DB', fillOpacity: 0.1, weight: 2 }).addTo(map);
		map.invalidateSize();
		map.setView(latlng, Math.max(map.getZoom(), 13));
	}

	// Geocoding Map Search
	let mapSearchQuery = $state('');
	let mapSearchResults = $state<{ display_name: string; lat: string; lon: string }[]>([]);
	let mapSearchLoading = $state(false);
	let mapSearchTimeout: ReturnType<typeof setTimeout>;

	function handleMapSearch(e: Event) {
		const target = e.target as HTMLInputElement;
		const query = target.value;
		clearTimeout(mapSearchTimeout);
		if (!query || query.trim().length < 3) {
			mapSearchResults = [];
			return;
		}
		
		mapSearchTimeout = setTimeout(async () => {
			mapSearchLoading = true;
			try {
				const res = await fetch(`https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=5`);
				if (res.ok) {
					mapSearchResults = await res.json();
				}
			} catch (err) {
				console.error('Error fetching address:', err);
			} finally {
				mapSearchLoading = false;
			}
		}, 650);
	}

	function selectMapResult(result: { display_name: string; lat: string; lon: string }) {
		form.latitude = parseFloat(result.lat);
		form.longitude = parseFloat(result.lon);
		if (result.display_name && !form.address) {
			form.address = result.display_name;
		}
		mapSearchResults = [];
		mapSearchQuery = '';
		updateMapOverlay();
	}

	const defaultColDef: ColDef = {
		sortable: true, resizable: true, filter: true, floatingFilter: false,
	};

	function iconEdit(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" /></svg>';
	}
	function iconDelete(): string {
		return '<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m14.74 9-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 0 1-2.244 2.077H8.084a2.25 2.25 0 0 1-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 0 0-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 0 1 3.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 0 0-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 0 0-7.5 0" /></svg>';
	}

	function createActionButton(html: string, className: string, ariaLabel: string, onClick: () => void): HTMLButtonElement {
		const btn = document.createElement('button');
		btn.innerHTML = html;
		btn.className = className;
		btn.setAttribute('aria-label', ariaLabel);
		btn.onclick = (e) => { e.stopPropagation(); onClick(); };
		return btn;
	}

	const columnDefs: ColDef[] = [
		{
			field: 'name', headerName: 'Lokasi', minWidth: 240, flex: 1,
			cellRenderer: (params: AgGridCellParams<AttendanceLocation>) => {
				if (!params.value) return '';
				const initials = getInitials(params.value as string);
				const addr = params.data?.address || '';
				return `<div class="flex items-center gap-3">
					<div class="w-9 h-9 rounded-lg bg-gradient-to-br from-emerald-50 to-emerald-100 flex items-center justify-center text-xs font-semibold text-emerald-600 shrink-0 ring-1 ring-emerald-200">${initials}</div>
					<div><div class="text-sm font-medium text-gray-900">${params.value}</div>${addr ? `<div class="text-xs text-gray-400 truncate max-w-48">${addr}</div>` : ''}</div>
				</div>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'py-1',
		},
		{
			field: 'latitude', headerName: 'Koordinat', minWidth: 180,
			cellRenderer: (params: AgGridCellParams<AttendanceLocation>) => {
				if (params.data?.latitude == null || params.data?.longitude == null) return '';
				return `<span class="text-sm font-mono text-gray-600">${Number(params.data.latitude).toFixed(6)}, ${Number(params.data.longitude).toFixed(6)}</span>`;
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-600',
		},
		{
			field: 'radius_meters', headerName: 'Radius', minWidth: 100,
			valueFormatter: (params: AgGridValueParams) => params.value != null ? `${params.value} m` : '',
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'is_active', headerName: 'Status', minWidth: 110,
			cellRenderer: (params: AgGridCellParams<AttendanceLocation>) => {
				if (params.value == null) return '';
				return params.value
					? '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-emerald-50 text-emerald-700 ring-1 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800">Aktif</span>'
					: '<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-50 text-red-700 ring-1 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800">Nonaktif</span>';
			},
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
		},
		{
			field: 'created_at', headerName: 'Dibuat', minWidth: 130,
			valueFormatter: (params: AgGridValueParams) => formatDate(params.value as string),
			headerClass: 'text-xs font-semibold text-gray-500 uppercase tracking-wider',
			cellClass: 'text-sm text-gray-500',
		},
		{
			field: 'id', headerName: '', minWidth: 100, maxWidth: 100,
			cellRenderer: (params: AgGridCellParams<AttendanceLocation>) => {
				const item = params.data;
				if (!item) return '';
				const container = document.createElement('div');
				container.className = 'flex items-center justify-end gap-1';
				const editBtn = createActionButton(iconEdit(),
					'p-1.5 rounded-lg text-gray-400 hover:text-[#1A56DB] hover:bg-blue-50 transition cursor-pointer',
					'Edit lokasi',
					() => openEditForm(item)
				);
				container.appendChild(editBtn);
				if (hasPermission('attendance_location', 'delete')) {
					const deleteBtn = createActionButton(iconDelete(),
						'p-1.5 rounded-lg text-gray-400 hover:text-red-600 hover:bg-red-50 transition cursor-pointer',
						'Hapus lokasi',
						() => confirmDelete(item.id, item.name)
					);
					container.appendChild(deleteBtn);
				}
				return container;
			},
			sortable: false, filter: false, resizable: false,
		},
	];

	const gridOptions: GridOptions = {
		columnDefs,
		defaultColDef,
		rowHeight: 56,
		headerHeight: 44,
		animateRows: true,
		domLayout: 'autoHeight',
		suppressDragLeaveHidesColumns: true,
		suppressRowHoverHighlight: false,
		enableCellTextSelection: true,
		pagination: false,
		theme: 'legacy',
		onGridReady: (params) => { gridApi = params.api; },
	};

	$effect(() => {
		if (showForm) {
			gridApi?.destroy();
			gridApi = null;
		}
	});

	$effect(() => {
		if (gridContainer && !showForm) {
			if (!gridApi && agGridModule) {
				gridApi = agGridModule.createGrid(gridContainer, gridOptions) as GridApi;
			}
			if (gridApi) { gridApi.updateGridOptions({ rowData: items }); }
		}
	});

	$effect(() => {
		if (showForm) {
			const lat = form.latitude;
			const lng = form.longitude;
			tick().then(() => {
				initMap();
				if (lat || lng) updateMapOverlay();
			});
		}
	});

	onMount(async () => {
		const m = await getAgGrid();
		agGridModule = m;
		load();
	});

	onDestroy(() => {
		gridApi?.destroy();
		gridApi = null;
		destroyMap();
	});
	async function load() {
		gridApi?.destroy();
		gridApi = null;
		isLoading = true;
		errorMessage = '';
		try {
			const response = await api.list(page, perPage, searchQuery) as ApiResponse<AttendanceLocation[]>;
			items = response.data || [];
			total = response.meta?.total || 0;
			page = response.meta?.page || 1;
			perPage = response.meta?.per_page || 25;
			totalPages = Math.ceil(total / perPage);
		} catch (error: unknown) { errorMessage = (error as { message?: string }).message || 'Gagal memuat data'; }
		finally { isLoading = false; }
	}

	function goToPage(p: number) { if (p < 1 || p > totalPages) return; page = p; load(); }

	function openCreateForm() {
		formTitle = 'Tambah Lokasi Absensi';
		editingId = null;
		form = { name: '', address: '', latitude: 0, longitude: 0, radius_meters: 100 };
		formError = '';
		showForm = true;
	}

	function openEditForm(item: AttendanceLocation) {
		formTitle = 'Edit Lokasi Absensi';
		editingId = item.id;
		form = { name: item.name, address: item.address, latitude: item.latitude, longitude: item.longitude, radius_meters: item.radius_meters };
		formError = '';
		showForm = true;
	}

	function cancelForm() {
		showCancelConfirm = true;
	}

	function confirmCancel() {
		showCancelConfirm = false;
		showForm = false;
		formError = '';
		mapSearchQuery = '';
		mapSearchResults = [];
	}

	function abortCancel() {
		showCancelConfirm = false;
	}

	async function handleSave() {
		if (!form.name.trim()) { formError = 'Nama lokasi absensi harus diisi'; return; }
		if (form.latitude === 0 && form.longitude === 0) { formError = 'Koordinat latitude dan longitude harus diisi'; return; }

		isSaving = true;
		formError = '';
		try {
			const payload: Record<string, unknown> = { name: form.name.trim(), address: form.address.trim(), latitude: form.latitude, longitude: form.longitude, radius_meters: form.radius_meters || 100 };
			if (editingId) { await api.update(editingId, payload); }
			else { await api.create(payload); }
			confirmCancel();
			load();
		} catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menyimpan data'; }
		finally { isSaving = false; }
	}

	function confirmDelete(id: string, name: string) { deletingId = id; deletingName = name; showDeleteConfirm = true; }
	function cancelDelete() { showDeleteConfirm = false; deletingId = null; deletingName = ''; }

	async function handleDelete() {
		if (!deletingId) return; isSaving = true;
		try { await api.remove(deletingId); showDeleteConfirm = false; deletingId = null; deletingName = ''; load(); }
		catch (error: unknown) { formError = (error as { message?: string }).message || 'Gagal menghapus data'; showDeleteConfirm = false; }
		finally { isSaving = false; }
	}

	let searchTimeout: ReturnType<typeof setTimeout>;
	function onSearchInput(e: Event) {
		const target = e.target as HTMLInputElement;
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => { searchQuery = target.value; page = 1; load(); }, 400);
	}

	function formatDate(dateStr: string): string {
		if (!dateStr) return '-';
		return new Date(dateStr).toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' });
	}

</script>

<div class="w-full">
	<div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-6">
		<div>
			<h1 class="text-2xl font-bold text-gray-900 tracking-tight">Lokasi Absensi</h1>
			<p class="text-sm text-gray-500 mt-0.5">Kelola lokasi GPS untuk check-in/out absensi karyawan</p>
		</div>
		{#if !showForm && hasPermission('attendance_location', 'create')}
			<button onclick={openCreateForm} class="inline-flex items-center gap-2 px-4 py-2.5 bg-[#1A56DB] text-white rounded-xl text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 cursor-pointer">
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>
				Tambah Lokasi
			</button>
		{/if}
	</div>

	<div class:hidden={showForm}>
		<div class="bg-white border border-gray-200 rounded-xl px-5 py-3.5 mb-4 flex flex-col sm:flex-row sm:items-center sm:justify-between gap-3">
			<div class="relative flex-1 max-w-md">
				<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
				<input type="search" value={searchQuery} placeholder="Cari lokasi absensi..." oninput={onSearchInput} class="w-full pl-9 pr-3 py-2 bg-gray-50 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] focus:bg-white transition placeholder:text-gray-400" />
			</div>
			<div class="text-xs text-gray-400">{total > 0 ? `${total} lokasi ditemukan` : ''}</div>
		</div>
	</div>

	<div class="bg-white border border-gray-200 rounded-xl shadow-sm transition-all duration-300 transform {showForm ? 'opacity-100 translate-y-0 relative' : 'opacity-0 translate-y-4 hidden'}" class:hidden={!showForm}>
			<div class="flex items-center justify-between px-6 py-4 border-b border-gray-200 bg-gray-50/50">
				<h2 class="text-lg font-semibold text-gray-900">{formTitle}</h2>
				<button onclick={cancelForm} aria-label="Tutup" class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 hover:bg-gray-100 transition cursor-pointer">
					<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" /></svg>
				</button>
			</div>
			<div class="px-6 py-5 grid grid-cols-1 lg:grid-cols-2 gap-8">
				<div class="flex flex-col h-full py-2 lg:pr-6 space-y-6">
					{#if formError}<div class="bg-red-50 border border-red-200 text-red-700 text-sm px-4 py-2.5 rounded-lg">{formError}</div>{/if}
					
					<!-- Section 1: Informasi Dasar -->
					<div class="mb-4">
						<h3 class="text-sm font-semibold text-gray-900 mb-4 flex items-center gap-2">
							<svg class="w-4 h-4 text-[#1A56DB]" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z" /></svg>
							Informasi Dasar
						</h3>
						<div class="space-y-4">
							<div>
								<label for="loc-name" class="block text-xs font-medium text-gray-600 mb-1.5">Nama Lokasi <span class="text-red-500">*</span></label>
								<input id="loc-name" type="text" bind:value={form.name} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white" placeholder="Contoh: Kantor Pusat" />
							</div>
							<div>
								<label for="loc-address" class="block text-xs font-medium text-gray-600 mb-1.5">Alamat</label>
								<textarea id="loc-address" bind:value={form.address} rows="2" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition resize-none bg-white" placeholder="Alamat lengkap lokasi (opsional)"></textarea>
							</div>
							<div>
								<label for="loc-radius" class="block text-xs font-medium text-gray-600 mb-1.5">Radius Toleransi (meter)</label>
								<input id="loc-radius" type="number" bind:value={form.radius_meters} min="10" max="10000" class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white" placeholder="100" />
							</div>
						</div>
					</div>

					<!-- Section 2: Koordinat -->
					<div>
						<h3 class="text-sm font-semibold text-gray-900 mb-4 flex items-center gap-2">
							<svg class="w-4 h-4 text-[#1A56DB]" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M15.042 21.672 13.684 16.6m0 0-2.51 2.225.569-9.47 5.227 7.917-3.286-.672ZM12 2.25V4.5m5.834.166-1.591 1.591M20.25 10.5H18M7.757 14.743l-1.59 1.59M6 10.5H3.75m4.007-4.243-1.59-1.59" /></svg>
							Titik Koordinat
						</h3>
						<div class="grid grid-cols-2 gap-4">
							<div>
								<label for="loc-lat" class="block text-xs font-medium text-gray-600 mb-1.5">Latitude <span class="text-red-500">*</span></label>
								<input id="loc-lat" type="number" step="any" bind:value={form.latitude} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white" placeholder="-6.2088" />
							</div>
							<div>
								<label for="loc-lng" class="block text-xs font-medium text-gray-600 mb-1.5">Longitude <span class="text-red-500">*</span></label>
								<input id="loc-lng" type="number" step="any" bind:value={form.longitude} class="w-full px-3 py-2.5 border border-gray-200 rounded-lg text-sm outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition bg-white" placeholder="106.8456" />
							</div>
						</div>
					</div>
				</div>

				<div class="flex flex-col h-full min-h-[400px] lg:min-h-[500px]">
					<label for="loc-map" class="block text-sm font-medium text-gray-700 mb-1.5">Peta Interaktif <span class="text-xs text-gray-400 font-normal">(klik peta untuk menentukan koordinat)</span></label>
					
					<!-- Map Search Box -->
					<div class="relative mb-2">
						<input type="text" bind:value={mapSearchQuery} oninput={handleMapSearch} placeholder="Cari alamat, jalan, atau kota..." class="w-full px-3 py-2.5 pl-9 border border-gray-200 dark:border-gray-700 rounded-lg text-sm bg-white dark:bg-gray-800 text-gray-900 dark:text-gray-100 outline-none focus:ring-2 focus:ring-[#1A56DB]/20 focus:border-[#1A56DB] transition placeholder:text-gray-400" />
						<svg class="w-4 h-4 text-gray-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /></svg>
						
						{#if mapSearchResults.length > 0}
							<div class="absolute left-0 right-0 mt-1 bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg shadow-lg z-[2000] max-h-48 overflow-y-auto divide-y divide-gray-100 dark:divide-gray-750">
								{#each mapSearchResults as result (result)}
									<button type="button" onclick={() => selectMapResult(result)} class="w-full text-left px-3 py-2 hover:bg-gray-50 dark:hover:bg-gray-750 text-xs text-gray-700 dark:text-gray-300 transition-colors truncate block cursor-pointer">
										{result.display_name}
									</button>
								{/each}
							</div>
						{/if}
						
						{#if mapSearchLoading}
							<div class="absolute right-3 top-1/2 -translate-y-1/2 flex items-center">
								<svg class="w-4 h-4 text-gray-400 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>
							</div>
						{/if}
					</div>

					<div class="relative flex-1 rounded-lg border border-gray-200 overflow-hidden" style="min-height: 250px;">
						<div bind:this={mapContainer} class="absolute inset-0" style="z-index: 1;"></div>
						
						<!-- Custom Map Layer Toggle Button -->
						<button type="button" onclick={toggleMapLayer} class="absolute bottom-4 left-4 z-[400] bg-white dark:bg-gray-800 p-1.5 rounded-xl shadow-lg shadow-black/10 border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors flex items-center gap-2.5 cursor-pointer group" aria-label="Ganti tipe peta">
							{#if mapLayerType === 'hybrid'}
								<div class="w-8 h-8 rounded-lg bg-blue-50 text-[#1A56DB] flex items-center justify-center">
									<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25A2.25 2.25 0 0 1 13.5 18v-2.25Z" /></svg>
								</div>
								<span class="text-xs font-semibold text-gray-700 pr-3 group-hover:text-[#1A56DB] transition-colors">Ubah ke Jalan</span>
							{:else}
								<div class="w-8 h-8 rounded-lg bg-emerald-50 text-emerald-600 flex items-center justify-center">
									<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 21a9.004 9.004 0 0 0 8.716-6.747M12 21a9.004 9.004 0 0 1-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 0 1 7.843 4.582M12 3a8.997 8.997 0 0 0-7.843 4.582m15.686 0A11.953 11.953 0 0 1 12 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0 1 21 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0 1 12 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 0 1 3 12c0-1.605.42-3.113 1.157-4.418" /></svg>
								</div>
								<span class="text-xs font-semibold text-gray-700 pr-3 group-hover:text-emerald-600 transition-colors">Ubah ke Satelit</span>
							{/if}
						</button>
					</div>
				</div>
			</div>
			<div class="sticky bottom-0 z-10 flex items-center justify-end gap-3 px-6 py-4 border-t border-gray-200 dark:border-gray-800 bg-gray-50/95 dark:bg-gray-900/95 backdrop-blur-sm">
				<button onclick={cancelForm} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
				<button onclick={handleSave} disabled={isSaving} class="px-5 py-2.5 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition disabled:opacity-50 disabled:cursor-not-allowed inline-flex items-center gap-2 cursor-pointer">
					{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
					{editingId ? 'Simpan Perubahan' : 'Tambah Lokasi'}
				</button>
			</div>
		</div>
	<div class:hidden={showForm}>
		<div class="bg-white border border-gray-200 rounded-xl overflow-hidden shadow-sm">
			{#if isLoading}
				<PulseLoader variant="table-row" count={5} />
			{:else if errorMessage}
				<div class="py-16 text-center">
					<div class="w-14 h-14 mx-auto mb-4 rounded-xl bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
					<p class="text-sm font-medium text-gray-900 mb-1">Gagal memuat data</p>
					<p class="text-sm text-gray-500 mb-4">{errorMessage}</p>
					<button onclick={load} class="px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-medium hover:bg-[#1e40af] transition cursor-pointer">Muat Ulang</button>
				</div>
			{:else if items.length === 0}
				<EmptyState
					variant="empty"
					title="Belum ada data lokasi absensi"
					description={searchQuery ? `Tidak ditemukan dengan kata kunci "${searchQuery}"` : 'Data lokasi absensi akan muncul setelah ditambahkan.'}
				/>
			{:else}
				<!-- Desktop Table — AG Grid -->
				<div class="hidden md:block">
					<div bind:this={gridContainer} class="ag-theme-quartz w-full" style="min-height: 400px"></div>
				</div>
				<div class="md:hidden space-y-3">
					<PullToRefresh onRefresh={load}>
					{#each items as item (item)}
						<MobileCard
							title={item.name}
							subtitle={item.address || ''}
							avatar={getInitials(item.name)}
							avatarColor={getAvatarTheme('attendanceLocation').gradientClasses}
							badges={[{ label: item.is_active ? 'Aktif' : 'Nonaktif', color: item.is_active ? 'bg-emerald-50 text-emerald-700 ring-emerald-200 dark:bg-emerald-900 dark:text-emerald-200 dark:ring-emerald-800' : 'bg-red-50 text-red-700 ring-red-200 dark:bg-red-900 dark:text-red-200 dark:ring-red-800' }]}
						>
							{#snippet footer()}
								<div class="flex items-center gap-2">
									<button onclick={() => openEditForm(item)} class="flex-1 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition cursor-pointer active:scale-95">Edit</button>
									{#if hasPermission('attendance_location', 'delete')}
										<button onclick={() => confirmDelete(item.id, item.name)} class="flex-1 py-2 text-xs font-medium text-red-600 dark:text-red-300 bg-red-50 dark:bg-red-900/30 rounded-lg hover:bg-red-100 dark:hover:bg-red-900/50 transition cursor-pointer active:scale-95">Hapus</button>
									{/if}
								</div>
							{/snippet}
						</MobileCard>
					{/each}
					</PullToRefresh>
				</div>
				<div class="flex items-center justify-between px-5 py-3.5 border-t border-gray-100 bg-gray-50/30">
					<div class="text-xs text-gray-500">Menampilkan {(page - 1) * perPage + 1}-{Math.min(page * perPage, total)} dari <span class="font-medium text-gray-700">{total}</span></div>
					<div class="flex items-center gap-1.5">
						<button onclick={() => goToPage(page - 1)} disabled={page <= 1} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Sebelumnya</button>
						{#each Array.from({ length: Math.min(5, totalPages) }) as _, i (i)}
							{@const pageNum = Math.max(1, Math.min(page - 2, totalPages - 4)) + i}
							{#if pageNum <= totalPages}
								<button onclick={() => goToPage(pageNum)} class="w-8 h-8 text-xs font-medium rounded-lg border transition cursor-pointer {pageNum === page ? 'bg-[#1A56DB] text-white border-[#1A56DB] shadow-sm' : 'border-gray-200 text-gray-600 hover:bg-gray-100'}">{pageNum}</button>
							{/if}
						{/each}
						<button onclick={() => goToPage(page + 1)} disabled={page >= totalPages} class="px-3 py-1.5 text-xs font-medium rounded-lg border border-gray-200 text-gray-600 hover:bg-gray-100 disabled:opacity-40 disabled:cursor-not-allowed transition cursor-pointer">Selanjutnya</button>
					</div>
				</div>
			{/if}
		</div>
	</div>
</div>

<AnimatedPresence show={showDeleteConfirm} type="scale" duration={200}>
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={cancelDelete} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') cancelDelete(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Hapus lokasi absensi" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-red-50 flex items-center justify-center"><svg class="w-7 h-7 text-red-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg></div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Hapus Lokasi Absensi</h3>
				<p class="text-sm text-gray-500 mb-1">Apakah Anda yakin ingin menghapus</p>
				<p class="text-sm font-medium text-gray-900 mb-4">"{deletingName}"?</p>
				<p class="text-xs text-gray-400 mb-6">Data yang sudah dihapus tidak dapat dikembalikan.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={cancelDelete} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Batal</button>
					<button onclick={handleDelete} disabled={isSaving} class="px-5 py-2.5 bg-red-600 text-white rounded-lg text-sm font-semibold hover:bg-red-700 transition disabled:opacity-50 inline-flex items-center gap-2 cursor-pointer">
						{#if isSaving}<svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" /><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" /></svg>{/if}
						Ya, Hapus
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>

<AnimatedPresence show={showCancelConfirm} type="scale" duration={200}>
	<!-- svelte-ignore a11y_click_events_have_key_events -->
	<div onclick={abortCancel} onkeydown={(e) => { if (e.key === 'Escape' || e.key === 'Enter') abortCancel(); }}
		role="presentation" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4">
		<div onclick={(e) => e.stopPropagation()} role="dialog" tabindex="-1" aria-modal="true" aria-label="Konfirmasi Batal" class="bg-white rounded-2xl shadow-2xl w-full max-w-sm transform transition-all">
			<div class="px-6 py-6 text-center">
				<div class="w-14 h-14 mx-auto mb-4 rounded-full bg-amber-50 flex items-center justify-center">
					<svg class="w-7 h-7 text-amber-500" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>
				</div>
				<h3 class="text-lg font-semibold text-gray-900 mb-2">Batalkan Perubahan?</h3>
				<p class="text-sm text-gray-500 mb-6">Semua data yang sudah diisi akan hilang dan tidak disimpan.</p>
				<div class="flex items-center justify-center gap-3">
					<button onclick={abortCancel} class="px-4 py-2.5 border border-gray-200 rounded-lg text-sm font-medium text-gray-700 hover:bg-gray-100 transition cursor-pointer">Lanjutkan Mengisi</button>
					<button onclick={confirmCancel} class="px-5 py-2.5 bg-amber-500 text-white rounded-lg text-sm font-semibold hover:bg-amber-600 transition inline-flex items-center gap-2 cursor-pointer">
						Ya, Batalkan
					</button>
				</div>
			</div>
		</div>
	</div>
</AnimatedPresence>
