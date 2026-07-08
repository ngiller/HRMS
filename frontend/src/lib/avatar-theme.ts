/**
 * Avatar Theme Configuration
 * Centralized color/gradient mapping per entity module for consistent
 * frosted glass avatar styling across all pages.
 *
 * Usage:
 *   import { getAvatarTheme } from '$lib/avatar-theme';
 *   const { gradient, bg, text, ring } = getAvatarTheme('leave');
 *   // → { gradient: 'from-sky-50 to-sky-100...', bg: 'bg-sky-50...', ... }
 */

export interface AvatarTheme {
	/** Gradient classes for the avatar background */
	gradient: string;
	/** Solid background (for non-gradient use) */
	bg: string;
	/** Text/icon color */
	text: string;
	/** Ring/border color */
	ring: string;
	/** Dark mode variants */
	darkGradient: string;
	darkBg: string;
	darkText: string;
	darkRing: string;
	/** Combined classes for typical w-9 h-9 rounded-xl gradient avatar */
	gradientClasses: string;
	/** Combined classes for typical w-9 h-9 rounded-xl solid avatar */
	solidClasses: string;
}

/**
 * Theme definitions — each module has a unique color identity.
 * Colors are chosen for clear visual distinction and accessibility.
 */
const themeMap: Record<string, AvatarTheme> = {
	// Karyawan / Employee
	employee: {
		gradient: 'from-cyan-50 to-cyan-100',
		bg: 'bg-cyan-50',
		text: 'text-cyan-600',
		ring: 'ring-cyan-200',
		darkGradient: 'dark:from-cyan-900/30 dark:to-cyan-800/30',
		darkBg: 'dark:bg-cyan-900/20',
		darkText: 'dark:text-cyan-400',
		darkRing: 'dark:ring-cyan-800',
		gradientClasses: 'from-cyan-50 to-cyan-100 dark:from-cyan-900/30 dark:to-cyan-800/30 text-cyan-600 dark:text-cyan-400 ring-cyan-200 dark:ring-cyan-800',
		solidClasses: 'bg-cyan-50 dark:bg-cyan-900/20 text-cyan-600 dark:text-cyan-400 ring-cyan-200 dark:ring-cyan-800',
	},
	// Cuti / Leave
	leave: {
		gradient: 'from-sky-50 to-sky-100',
		bg: 'bg-sky-50',
		text: 'text-sky-600',
		ring: 'ring-sky-200',
		darkGradient: 'dark:from-sky-900/30 dark:to-sky-800/30',
		darkBg: 'dark:bg-sky-900/20',
		darkText: 'dark:text-sky-400',
		darkRing: 'dark:ring-sky-800',
		gradientClasses: 'from-sky-50 to-sky-100 dark:from-sky-900/30 dark:to-sky-800/30 text-sky-600 dark:text-sky-400 ring-sky-200 dark:ring-sky-800',
		solidClasses: 'bg-sky-50 dark:bg-sky-900/20 text-sky-600 dark:text-sky-400 ring-sky-200 dark:ring-sky-800',
	},
	// Lembur / Overtime
	overtime: {
		gradient: 'from-amber-50 to-amber-100',
		bg: 'bg-amber-50',
		text: 'text-amber-600',
		ring: 'ring-amber-200',
		darkGradient: 'dark:from-amber-900/30 dark:to-amber-800/30',
		darkBg: 'dark:bg-amber-900/20',
		darkText: 'dark:text-amber-400',
		darkRing: 'dark:ring-amber-800',
		gradientClasses: 'from-amber-50 to-amber-100 dark:from-amber-900/30 dark:to-amber-800/30 text-amber-600 dark:text-amber-400 ring-amber-200 dark:ring-amber-800',
		solidClasses: 'bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 ring-amber-200 dark:ring-amber-800',
	},
	// Reimbursement
	reimbursement: {
		gradient: 'from-violet-50 to-violet-100',
		bg: 'bg-violet-50',
		text: 'text-violet-600',
		ring: 'ring-violet-200',
		darkGradient: 'dark:from-violet-900/30 dark:to-violet-800/30',
		darkBg: 'dark:bg-violet-900/20',
		darkText: 'dark:text-violet-400',
		darkRing: 'dark:ring-violet-800',
		gradientClasses: 'from-violet-50 to-violet-100 dark:from-violet-900/30 dark:to-violet-800/30 text-violet-600 dark:text-violet-400 ring-violet-200 dark:ring-violet-800',
		solidClasses: 'bg-violet-50 dark:bg-violet-900/20 text-violet-600 dark:text-violet-400 ring-violet-200 dark:ring-violet-800',
	},
	// Pinjaman / Loan
	loan: {
		gradient: 'from-emerald-50 to-emerald-100',
		bg: 'bg-emerald-50',
		text: 'text-emerald-600',
		ring: 'ring-emerald-200',
		darkGradient: 'dark:from-emerald-900/30 dark:to-emerald-800/30',
		darkBg: 'dark:bg-emerald-900/20',
		darkText: 'dark:text-emerald-400',
		darkRing: 'dark:ring-emerald-800',
		gradientClasses: 'from-emerald-50 to-emerald-100 dark:from-emerald-900/30 dark:to-emerald-800/30 text-emerald-600 dark:text-emerald-400 ring-emerald-200 dark:ring-emerald-800',
		solidClasses: 'bg-emerald-50 dark:bg-emerald-900/20 text-emerald-600 dark:text-emerald-400 ring-emerald-200 dark:ring-emerald-800',
	},
	// KPI / Performance
	kpi: {
		gradient: 'from-rose-50 to-rose-100',
		bg: 'bg-rose-50',
		text: 'text-rose-600',
		ring: 'ring-rose-200',
		darkGradient: 'dark:from-rose-900/30 dark:to-rose-800/30',
		darkBg: 'dark:bg-rose-900/20',
		darkText: 'dark:text-rose-400',
		darkRing: 'dark:ring-rose-800',
		gradientClasses: 'from-rose-50 to-rose-100 dark:from-rose-900/30 dark:to-rose-800/30 text-rose-600 dark:text-rose-400 ring-rose-200 dark:ring-rose-800',
		solidClasses: 'bg-rose-50 dark:bg-rose-900/20 text-rose-600 dark:text-rose-400 ring-rose-200 dark:ring-rose-800',
	},
	// Penggajian / Payroll
	payroll: {
		gradient: 'from-blue-50 to-blue-100',
		bg: 'bg-blue-50',
		text: 'text-blue-600',
		ring: 'ring-blue-200',
		darkGradient: 'dark:from-blue-900/30 dark:to-blue-800/30',
		darkBg: 'dark:bg-blue-900/20',
		darkText: 'dark:text-blue-400',
		darkRing: 'dark:ring-blue-800',
		gradientClasses: 'from-blue-50 to-blue-100 dark:from-blue-900/30 dark:to-blue-800/30 text-blue-600 dark:text-blue-400 ring-blue-200 dark:ring-blue-800',
		solidClasses: 'bg-blue-50 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400 ring-blue-200 dark:ring-blue-800',
	},
	// Absensi / Attendance
	attendance: {
		gradient: 'from-teal-50 to-teal-100',
		bg: 'bg-teal-50',
		text: 'text-teal-600',
		ring: 'ring-teal-200',
		darkGradient: 'dark:from-teal-900/30 dark:to-teal-800/30',
		darkBg: 'dark:bg-teal-900/20',
		darkText: 'dark:text-teal-400',
		darkRing: 'dark:ring-teal-800',
		gradientClasses: 'from-teal-50 to-teal-100 dark:from-teal-900/30 dark:to-teal-800/30 text-teal-600 dark:text-teal-400 ring-teal-200 dark:ring-teal-800',
		solidClasses: 'bg-teal-50 dark:bg-teal-900/20 text-teal-600 dark:text-teal-400 ring-teal-200 dark:ring-teal-800',
	},
	// Surat Peringatan / Reprimand
	reprimand: {
		gradient: 'from-red-50 to-red-100',
		bg: 'bg-red-50',
		text: 'text-red-600',
		ring: 'ring-red-200',
		darkGradient: 'dark:from-red-900/30 dark:to-red-800/30',
		darkBg: 'dark:bg-red-900/20',
		darkText: 'dark:text-red-400',
		darkRing: 'dark:ring-red-800',
		gradientClasses: 'from-red-50 to-red-100 dark:from-red-900/30 dark:to-red-800/30 text-red-600 dark:text-red-400 ring-red-200 dark:ring-red-800',
		solidClasses: 'bg-red-50 dark:bg-red-900/20 text-red-600 dark:text-red-400 ring-red-200 dark:ring-red-800',
	},
	// Dokumen / Document
	document: {
		gradient: 'from-indigo-50 to-indigo-100',
		bg: 'bg-indigo-50',
		text: 'text-indigo-600',
		ring: 'ring-indigo-200',
		darkGradient: 'dark:from-indigo-900/30 dark:to-indigo-800/30',
		darkBg: 'dark:bg-indigo-900/20',
		darkText: 'dark:text-indigo-400',
		darkRing: 'dark:ring-indigo-800',
		gradientClasses: 'from-indigo-50 to-indigo-100 dark:from-indigo-900/30 dark:to-indigo-800/30 text-indigo-600 dark:text-indigo-400 ring-indigo-200 dark:ring-indigo-800',
		solidClasses: 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 ring-indigo-200 dark:ring-indigo-800',
	},
	// Jurnal Harian / Daily Journal
	dailyJournal: {
		gradient: 'from-orange-50 to-orange-100',
		bg: 'bg-orange-50',
		text: 'text-orange-600',
		ring: 'ring-orange-200',
		darkGradient: 'dark:from-orange-900/30 dark:to-orange-800/30',
		darkBg: 'dark:bg-orange-900/20',
		darkText: 'dark:text-orange-400',
		darkRing: 'dark:ring-orange-800',
		gradientClasses: 'from-orange-50 to-orange-100 dark:from-orange-900/30 dark:to-orange-800/30 text-orange-600 dark:text-orange-400 ring-orange-200 dark:ring-orange-800',
		solidClasses: 'bg-orange-50 dark:bg-orange-900/20 text-orange-600 dark:text-orange-400 ring-orange-200 dark:ring-orange-800',
	},
	// Pengumuman / Announcement
	announcement: {
		gradient: 'from-pink-50 to-pink-100',
		bg: 'bg-pink-50',
		text: 'text-pink-600',
		ring: 'ring-pink-200',
		darkGradient: 'dark:from-pink-900/30 dark:to-pink-800/30',
		darkBg: 'dark:bg-pink-900/20',
		darkText: 'dark:text-pink-400',
		darkRing: 'dark:ring-pink-800',
		gradientClasses: 'from-pink-50 to-pink-100 dark:from-pink-900/30 dark:to-pink-800/30 text-pink-600 dark:text-pink-400 ring-pink-200 dark:ring-pink-800',
		solidClasses: 'bg-pink-50 dark:bg-pink-900/20 text-pink-600 dark:text-pink-400 ring-pink-200 dark:ring-pink-800',
	},
	// Lokasi Absensi / Attendance Location
	attendanceLocation: {
		gradient: 'from-lime-50 to-lime-100',
		bg: 'bg-lime-50',
		text: 'text-lime-600',
		ring: 'ring-lime-200',
		darkGradient: 'dark:from-lime-900/30 dark:to-lime-800/30',
		darkBg: 'dark:bg-lime-900/20',
		darkText: 'dark:text-lime-400',
		darkRing: 'dark:ring-lime-800',
		gradientClasses: 'from-lime-50 to-lime-100 dark:from-lime-900/30 dark:to-lime-800/30 text-lime-600 dark:text-lime-400 ring-lime-200 dark:ring-lime-800',
		solidClasses: 'bg-lime-50 dark:bg-lime-900/20 text-lime-600 dark:text-lime-400 ring-lime-200 dark:ring-lime-800',
	},
	// Notifikasi / Notification
	notification: {
		gradient: 'from-gray-50 to-gray-100',
		bg: 'bg-gray-50',
		text: 'text-gray-600',
		ring: 'ring-gray-200',
		darkGradient: 'dark:from-gray-800/30 dark:to-gray-700/30',
		darkBg: 'dark:bg-gray-800/20',
		darkText: 'dark:text-gray-400',
		darkRing: 'dark:ring-gray-700',
		gradientClasses: 'from-gray-50 to-gray-100 dark:from-gray-800/30 dark:to-gray-700/30 text-gray-600 dark:text-gray-400 ring-gray-200 dark:ring-gray-700',
		solidClasses: 'bg-gray-50 dark:bg-gray-800/20 text-gray-600 dark:text-gray-400 ring-gray-200 dark:ring-gray-700',
	},
	// Shift / Permintaan Shift
	shift: {
		gradient: 'from-yellow-50 to-yellow-100',
		bg: 'bg-yellow-50',
		text: 'text-yellow-600',
		ring: 'ring-yellow-200',
		darkGradient: 'dark:from-yellow-900/30 dark:to-yellow-800/30',
		darkBg: 'dark:bg-yellow-900/20',
		darkText: 'dark:text-yellow-400',
		darkRing: 'dark:ring-yellow-800',
		gradientClasses: 'from-yellow-50 to-yellow-100 dark:from-yellow-900/30 dark:to-yellow-800/30 text-yellow-600 dark:text-yellow-400 ring-yellow-200 dark:ring-yellow-800',
		solidClasses: 'bg-yellow-50 dark:bg-yellow-900/20 text-yellow-600 dark:text-yellow-400 ring-yellow-200 dark:ring-yellow-800',
	},
	// Posisi Jabatan / Position
	position: {
		gradient: 'from-amber-50 to-amber-100',
		bg: 'bg-amber-50',
		text: 'text-amber-600',
		ring: 'ring-amber-200',
		darkGradient: 'dark:from-amber-900/30 dark:to-amber-800/30',
		darkBg: 'dark:bg-amber-900/20',
		darkText: 'dark:text-amber-400',
		darkRing: 'dark:ring-amber-800',
		gradientClasses: 'from-amber-50 to-amber-100 dark:from-amber-900/30 dark:to-amber-800/30 text-amber-600 dark:text-amber-400 ring-amber-200 dark:ring-amber-800',
		solidClasses: 'bg-amber-50 dark:bg-amber-900/20 text-amber-600 dark:text-amber-400 ring-amber-200 dark:ring-amber-800',
	},
	// Golongan Jabatan / Position Grade
	positionGrade: {
		gradient: 'from-purple-50 to-purple-100',
		bg: 'bg-purple-50',
		text: 'text-purple-600',
		ring: 'ring-purple-200',
		darkGradient: 'dark:from-purple-900/30 dark:to-purple-800/30',
		darkBg: 'dark:bg-purple-900/20',
		darkText: 'dark:text-purple-400',
		darkRing: 'dark:ring-purple-800',
		gradientClasses: 'from-purple-50 to-purple-100 dark:from-purple-900/30 dark:to-purple-800/30 text-purple-600 dark:text-purple-400 ring-purple-200 dark:ring-purple-800',
		solidClasses: 'bg-purple-50 dark:bg-purple-900/20 text-purple-600 dark:text-purple-400 ring-purple-200 dark:ring-purple-800',
	},
	// Jadwal / Schedule
	schedule: {
		gradient: 'from-cyan-50 to-cyan-100',
		bg: 'bg-cyan-50',
		text: 'text-cyan-600',
		ring: 'ring-cyan-200',
		darkGradient: 'dark:from-cyan-900/30 dark:to-cyan-800/30',
		darkBg: 'dark:bg-cyan-900/20',
		darkText: 'dark:text-cyan-400',
		darkRing: 'dark:ring-cyan-800',
		gradientClasses: 'from-cyan-50 to-cyan-100 dark:from-cyan-900/30 dark:to-cyan-800/30 text-cyan-600 dark:text-cyan-400 ring-cyan-200 dark:ring-cyan-800',
		solidClasses: 'bg-cyan-50 dark:bg-cyan-900/20 text-cyan-600 dark:text-cyan-400 ring-cyan-200 dark:ring-cyan-800',
	},
	// Hari Libur / Holiday
	holiday: {
		gradient: 'from-green-50 to-green-100',
		bg: 'bg-green-50',
		text: 'text-green-600',
		ring: 'ring-green-200',
		darkGradient: 'dark:from-green-900/30 dark:to-green-800/30',
		darkBg: 'dark:bg-green-900/20',
		darkText: 'dark:text-green-400',
		darkRing: 'dark:ring-green-800',
		gradientClasses: 'from-green-50 to-green-100 dark:from-green-900/30 dark:to-green-800/30 text-green-600 dark:text-green-400 ring-green-200 dark:ring-green-800',
		solidClasses: 'bg-green-50 dark:bg-green-900/20 text-green-600 dark:text-green-400 ring-green-200 dark:ring-green-800',
	},
	// Departemen / Department
	department: {
		gradient: 'from-indigo-50 to-indigo-100',
		bg: 'bg-indigo-50',
		text: 'text-indigo-600',
		ring: 'ring-indigo-200',
		darkGradient: 'dark:from-indigo-900/30 dark:to-indigo-800/30',
		darkBg: 'dark:bg-indigo-900/20',
		darkText: 'dark:text-indigo-400',
		darkRing: 'dark:ring-indigo-800',
		gradientClasses: 'from-indigo-50 to-indigo-100 dark:from-indigo-900/30 dark:to-indigo-800/30 text-indigo-600 dark:text-indigo-400 ring-indigo-200 dark:ring-indigo-800',
		solidClasses: 'bg-indigo-50 dark:bg-indigo-900/20 text-indigo-600 dark:text-indigo-400 ring-indigo-200 dark:ring-indigo-800',
	},
	// Resign / Exit
	resign: {
		gradient: 'from-rose-50 to-rose-100',
		bg: 'bg-rose-50',
		text: 'text-rose-600',
		ring: 'ring-rose-200',
		darkGradient: 'dark:from-rose-900/30 dark:to-rose-800/30',
		darkBg: 'dark:bg-rose-900/20',
		darkText: 'dark:text-rose-400',
		darkRing: 'dark:ring-rose-800',
		gradientClasses: 'from-rose-50 to-rose-100 dark:from-rose-900/30 dark:to-rose-800/30 text-rose-600 dark:text-rose-400 ring-rose-200 dark:ring-rose-800',
		solidClasses: 'bg-rose-50 dark:bg-rose-900/20 text-rose-600 dark:text-rose-400 ring-rose-200 dark:ring-rose-800',
	},
	// Default / Fallback
	default: {
		gradient: 'from-gray-50 to-gray-100',
		bg: 'bg-gray-50',
		text: 'text-gray-600',
		ring: 'ring-gray-200',
		darkGradient: 'dark:from-gray-800/30 dark:to-gray-700/30',
		darkBg: 'dark:bg-gray-800/20',
		darkText: 'dark:text-gray-400',
		darkRing: 'dark:ring-gray-700',
		gradientClasses: 'from-gray-50 to-gray-100 dark:from-gray-800/30 dark:to-gray-700/30 text-gray-600 dark:text-gray-400 ring-gray-200 dark:ring-gray-700',
		solidClasses: 'bg-gray-50 dark:bg-gray-800/20 text-gray-600 dark:text-gray-400 ring-gray-200 dark:ring-gray-700',
	},
} as const;

/** Entity type keys for autocomplete */
export type AvatarEntity = keyof typeof themeMap;

/**
 * Get avatar theme classes for a given entity type.
 * Falls back to "default" if the entity is unknown.
 */
export function getAvatarTheme(entity: string): AvatarTheme {
	return themeMap[entity] || themeMap.default;
}

/**
 * Get the default avatar color for MobileCard component.
 * Returns the full class string ready to use in the avatarColor prop.
 */
export function getAvatarColorForEntity(entity: string): string {
	return getAvatarTheme(entity).gradientClasses;
}

/**
 * Hash-based deterministic avatar color for dynamic entities
 * (e.g., department names, position names, employee names).
 */
export function getHashAvatarColor(name: string): string {
	if (!name) return themeMap.default.gradientClasses;

	let hash = 0;
	for (let i = 0; i < name.length; i++) {
		hash = name.charCodeAt(i) + ((hash << 5) - hash);
	}

	const palette: (keyof typeof themeMap)[] = [
		'employee', 'leave', 'overtime', 'reimbursement',
		'loan', 'payroll', 'attendance', 'shift',
		'holiday', 'document', 'notification',
	];

	const index = Math.abs(hash) % palette.length;
	return themeMap[palette[index]].gradientClasses;
}

/**
 * Get initials from a name string (max 2 characters).
 */
export function getInitials(name: string): string {
	if (!name) return '?';
	const parts = name.trim().split(/\s+/);
	if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
	return name.substring(0, 2).toUpperCase();
}
