<script lang="ts">
	import type { Snippet } from 'svelte';

	type Badge = {
		label: string;
		color?: string; // Tailwind classes for the badge
	};

	interface Props {
		/** Card click handler (optional) */
		onclick?: () => void;
		/** Avatar initials (auto-uppercased) */
		avatar?: string;
		/** Avatar gradient color theme */
		avatarColor?: string;
		/** Main title text */
		title?: string;
		/** Subtitle text (smaller, gray) */
		subtitle?: string;
		/** Badge(s) to show top-right */
		badges?: Badge[];
		/** Whether card is clickable */
		clickable?: boolean;
		/** Custom snippet for body content */
		children?: Snippet;
		/** Custom snippet for footer/actions area */
		footer?: Snippet;
		/** Extra Tailwind classes for the card */
		class?: string;
	}

	let {
		onclick = undefined,
		avatar = '',
		avatarColor = 'from-cyan-50 to-cyan-100 dark:from-cyan-900/30 dark:to-cyan-800/30 text-cyan-600 dark:text-cyan-400 ring-cyan-200 dark:ring-cyan-800',
		title = '',
		subtitle = '',
		badges = [] as Badge[],
		clickable = false,
		children = undefined,
		footer = undefined,
		class: extraClass = '',
	}: Props = $props();

	function getInitials(name: string): string {
		if (!name) return '?';
		const parts = name.trim().split(/\s+/);
		if (parts.length > 1) return (parts[0][0] + parts[1][0]).toUpperCase();
		return name.substring(0, 2).toUpperCase();
	}
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<!-- svelte-ignore a11y_click_events_have_key_events -->
<div
	{...typeof onclick === 'function' ? { onclick, role: 'button', tabindex: 0 } : {}}
	onkeydown={(e) => { if (typeof onclick === 'function' && (e.key === 'Enter' || e.key === ' ')) { e.preventDefault(); onclick(); } }}
	class="bg-white/95 dark:bg-gray-900/95 backdrop-blur-sm rounded-xl border border-gray-200 dark:border-gray-800 p-4 transition-all duration-150 shadow-sm hover:shadow-md {clickable || typeof onclick === 'function' ? 'active:scale-[0.97] cursor-pointer' : ''} {extraClass}"
>
	{#if title || avatar}
		<div class="flex items-center justify-between mb-3">
			<div class="flex items-center gap-3 min-w-0 flex-1">
				{#if avatar}
					<div class="w-9 h-9 rounded-xl bg-gradient-to-br shrink-0 ring-1 flex items-center justify-center text-xs font-semibold {avatarColor}">
						{getInitials(avatar)}
					</div>
				{/if}
				<div class="min-w-0">
					{#if title}
						<div class="text-sm font-semibold text-gray-900 dark:text-white truncate leading-tight">{title}</div>
					{/if}
					{#if subtitle}
						<div class="text-xs text-gray-400 dark:text-gray-500 mt-0.5 truncate">{subtitle}</div>
					{/if}
				</div>
			</div>
			{#if badges && badges.length > 0}
				<div class="flex items-center gap-1.5 shrink-0">
					{#each badges as badge}
						<span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ring-1 {badge.color || 'bg-gray-50 text-gray-600 ring-gray-200 dark:bg-gray-800 dark:text-gray-400 dark:ring-gray-700'}">
							{badge.label}
						</span>
					{/each}
				</div>
			{/if}
		</div>
	{/if}

	{#if children}
		{@render children()}
	{/if}

	{#if footer}
		<div class="pt-2 mt-2 border-t border-gray-100 dark:border-gray-800">
			{@render footer()}
		</div>
	{/if}
</div>
