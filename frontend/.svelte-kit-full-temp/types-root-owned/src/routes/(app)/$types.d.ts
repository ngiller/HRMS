import type * as Kit from '@sveltejs/kit';

type Expand<T> = T extends infer O ? { [K in keyof O]: O[K] } : never;
type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;
type RouteParams = {  };
type RouteId = '/(app)';
type MaybeWithVoid<T> = {} extends T ? T | void : T;
export type RequiredKeys<T> = { [K in keyof T]-?: {} extends { [P in K]: T[K] } ? never : K; }[keyof T];
type OutputDataShape<T> = MaybeWithVoid<Omit<App.PageData, RequiredKeys<T>> & Partial<Pick<App.PageData, keyof T & keyof App.PageData>> & Record<string, any>>
type EnsureDefined<T> = T extends null | undefined ? {} : T;
type OptionalUnion<U extends Record<string, any>, A extends keyof U = U extends U ? keyof U : never> = U extends unknown ? { [P in Exclude<A, keyof U>]?: never } & U : never;
export type Snapshot<T = any> = Kit.Snapshot<T>;
type LayoutRouteId = RouteId | "/(app)/[...rest]" | "/(app)/absensi" | "/(app)/absensi-manual" | "/(app)/audit-trail" | "/(app)/cuti" | "/(app)/dashboard" | "/(app)/dashboard/change-password" | "/(app)/dashboard/pengaturan" | "/(app)/dashboard/profile" | "/(app)/departemen" | "/(app)/dokumen" | "/(app)/golongan-jabatan" | "/(app)/hari-libur" | "/(app)/jadwal-karyawan" | "/(app)/jadwal-kerja" | "/(app)/jadwal-templates" | "/(app)/jurnal-harian" | "/(app)/karyawan" | "/(app)/karyawan/[id]" | "/(app)/karyawan/detail" | "/(app)/kpi" | "/(app)/laporan" | "/(app)/lembur" | "/(app)/lokasi-absensi" | "/(app)/notifikasi" | "/(app)/pengaturan" | "/(app)/pengaturan/roles" | "/(app)/penggajian" | "/(app)/penggajian/[id]" | "/(app)/penggajian/payslip/[periodId]/[employeeId]" | "/(app)/penggajian/slip-saya" | "/(app)/pengumuman" | "/(app)/permintaan-shift" | "/(app)/persetujuan" | "/(app)/pinjaman" | "/(app)/posisi-jabatan" | "/(app)/reimbursement" | "/(app)/resign" | "/(app)/struktur-organisasi" | "/(app)/surat-peringatan"
type LayoutParams = RouteParams & { rest?: string | undefined; id?: string | undefined; periodId?: string | undefined; employeeId?: string | undefined }
type LayoutParentData = EnsureDefined<import('../$types.js').LayoutData>;

export type LayoutServerData = null;
export type LayoutData = Expand<LayoutParentData>;
export type LayoutProps = { params: LayoutParams; data: LayoutData; children: import("svelte").Snippet }