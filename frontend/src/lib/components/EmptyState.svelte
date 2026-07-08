<script lang="ts">
	interface Props {
		/** Empty state variant — determines icon and color */
		variant?: 'empty' | 'search' | 'error' | 'success';
		/** Main title text */
		title?: string;
		/** Description text (smaller, gray) */
		description?: string;
		/** Optional action button label */
		actionLabel?: string;
		/** Optional action button click handler */
		onAction?: () => void;
		/** Optional extra icon (overrides variant default) — raw SVG string */
		icon?: string;
		/** Compact mode: smaller icons and paddings */
		compact?: boolean;
	}

	let {
		variant = 'empty',
		title = '',
		description = '',
		actionLabel = '',
		onAction = undefined,
		icon = '',
		compact = false,
	}: Props = $props();

	const icons: Record<string, string> = {
		empty: '<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 0 1-2.247 2.118H6.622a2.25 2.25 0 0 1-2.247-2.118L3.75 7.5m6 4.125 2.25 2.25m0 0 2.25 2.25M12 11.625l2.25-2.25M12 11.625l-2.25 2.25M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125Z" /></svg>',
		search: '<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-5.197-5.197m0 0A7.5 7.5 0 1 0 5.196 5.196a7.5 7.5 0 0 0 10.607 10.607Z" /><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 10.5a5.25 5.25 0 1 1-10.5 0 5.25 5.25 0 0 1 10.5 0Z" stroke-width="1.5" /></svg>',
		error: '<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>',
		success: '<svg class="w-7 h-7" fill="none" viewBox="0 0 24 24" stroke-width="1.2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z" /></svg>',
	};

	const iconColors: Record<string, string> = {
		empty: 'bg-gray-50 dark:bg-gray-800 text-gray-400 dark:text-gray-500',
		search: 'bg-amber-50 dark:bg-amber-900/10 text-amber-500 dark:text-amber-400',
		error: 'bg-red-50 dark:bg-red-900/10 text-red-500 dark:text-red-400',
		success: 'bg-emerald-50 dark:bg-emerald-900/10 text-emerald-500 dark:text-emerald-400',
	};
</script>

<div class="py-{compact ? '10' : '16'} text-center">
	<!-- Icon -->
	<div class="w-{compact ? '12' : '14'} h-{compact ? '12' : '14'} mx-auto mb-{compact ? '3' : '4'} rounded-xl flex items-center justify-center {iconColors[variant] || iconColors.empty}">
		{#if icon}
			{@html icon}
		{:else}
			{@html icons[variant] || icons.empty}
		{/if}
	</div>

	<!-- Title -->
	{#if title}
		<h3 class="text-sm font-semibold text-gray-900 dark:text-white mb-1">{title}</h3>
	{/if}

	<!-- Description -->
	{#if description}
		<p class="text-sm text-gray-500 dark:text-gray-400 mb-{actionLabel ? '4' : '0'}">{description}</p>
	{/if}

	<!-- Action Button -->
	{#if actionLabel && typeof onAction === 'function'}
		<button
			onclick={onAction}
			class="inline-flex items-center gap-2 px-5 py-2 bg-[#1A56DB] text-white rounded-lg text-sm font-semibold hover:bg-[#1e40af] transition-all active:scale-[0.97] shadow-sm shadow-blue-200 dark:shadow-none cursor-pointer"
		>
			{actionLabel}
		</button>
	{/if}
</div>
