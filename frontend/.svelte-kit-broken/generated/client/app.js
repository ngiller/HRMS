export { matchers } from './matchers.js';

export const nodes = [
	() => import('./nodes/0'),
	() => import('./nodes/1'),
	() => import('./nodes/2'),
	() => import('./nodes/3'),
	() => import('./nodes/4'),
	() => import('./nodes/5'),
	() => import('./nodes/6'),
	() => import('./nodes/7'),
	() => import('./nodes/8'),
	() => import('./nodes/9'),
	() => import('./nodes/10'),
	() => import('./nodes/11'),
	() => import('./nodes/12'),
	() => import('./nodes/13'),
	() => import('./nodes/14'),
	() => import('./nodes/15'),
	() => import('./nodes/16'),
	() => import('./nodes/17'),
	() => import('./nodes/18'),
	() => import('./nodes/19'),
	() => import('./nodes/20'),
	() => import('./nodes/21'),
	() => import('./nodes/22'),
	() => import('./nodes/23'),
	() => import('./nodes/24'),
	() => import('./nodes/25'),
	() => import('./nodes/26'),
	() => import('./nodes/27'),
	() => import('./nodes/28'),
	() => import('./nodes/29'),
	() => import('./nodes/30'),
	() => import('./nodes/31'),
	() => import('./nodes/32'),
	() => import('./nodes/33'),
	() => import('./nodes/34'),
	() => import('./nodes/35'),
	() => import('./nodes/36'),
	() => import('./nodes/37'),
	() => import('./nodes/38'),
	() => import('./nodes/39'),
	() => import('./nodes/40'),
	() => import('./nodes/41'),
	() => import('./nodes/42')
];

export const server_loads = [];

export const dictionary = {
		"/": [3],
		"/(app)/absensi": [5,[2]],
		"/(app)/audit-trail": [6,[2]],
		"/(app)/cuti": [7,[2]],
		"/(app)/dashboard": [8,[2]],
		"/(app)/dashboard/change-password": [9,[2]],
		"/(app)/dashboard/pengaturan": [10,[2]],
		"/(app)/dashboard/profile": [11,[2]],
		"/(app)/departemen": [12,[2]],
		"/(app)/dokumen": [13,[2]],
		"/forgot-password": [40],
		"/(app)/golongan-jabatan": [14,[2]],
		"/(app)/hari-libur": [15,[2]],
		"/(app)/jadwal-karyawan": [16,[2]],
		"/(app)/jadwal-kerja": [17,[2]],
		"/(app)/jadwal-templates": [18,[2]],
		"/(app)/jurnal-harian": [19,[2]],
		"/(app)/karyawan": [20,[2]],
		"/(app)/karyawan/detail": [22,[2]],
		"/(app)/karyawan/[id]": [21,[2]],
		"/(app)/kpi": [23,[2]],
		"/(app)/laporan": [24,[2]],
		"/(app)/lembur": [25,[2]],
		"/login": [41],
		"/(app)/lokasi-absensi": [26,[2]],
		"/(app)/notifikasi": [27,[2]],
		"/(app)/pengaturan/roles": [28,[2]],
		"/(app)/penggajian": [29,[2]],
		"/(app)/penggajian/payslip/[periodId]/[employeeId]": [31,[2]],
		"/(app)/penggajian/slip-saya": [32,[2]],
		"/(app)/penggajian/[id]": [30,[2]],
		"/(app)/pengumuman": [33,[2]],
		"/(app)/permintaan-shift": [34,[2]],
		"/(app)/pinjaman": [35,[2]],
		"/(app)/posisi-jabatan": [36,[2]],
		"/(app)/reimbursement": [37,[2]],
		"/reset-password": [42],
		"/(app)/struktur-organisasi": [38,[2]],
		"/(app)/surat-peringatan": [39,[2]],
		"/(app)/[...rest]": [4,[2]]
	};

export const hooks = {
	handleError: (({ error }) => { console.error(error) }),
	
	reroute: (() => {}),
	transport: {}
};

export const decoders = Object.fromEntries(Object.entries(hooks.transport).map(([k, v]) => [k, v.decode]));
export const encoders = Object.fromEntries(Object.entries(hooks.transport).map(([k, v]) => [k, v.encode]));

export const hash = false;

export const decode = (type, value) => decoders[type](value);

export { default as root } from '../root.js';