<script lang="ts">
	import { onDestroy } from 'svelte';

	interface Props {
		onRefresh: () => Promise<void>;
		isRefreshing?: boolean;
		children: import('svelte').Snippet;
	}

	let { onRefresh, isRefreshing = false, children }: Props = $props();

	let pullDistance = $state(0);
	let isPulling = $state(false);
	let isRefreshingState = $state(false);
	let startY = 0;
	let containerEl = $state<HTMLDivElement>();
	let refreshTimer: ReturnType<typeof setTimeout> | undefined;

	onDestroy(() => clearTimeout(refreshTimer));

	const THRESHOLD = 60;
	const MAX_PULL = 100;

	let progress = $derived(Math.min(pullDistance / THRESHOLD, 1));

	// Refresh state sync
	$effect(() => {
		if (isRefreshing) {
			isRefreshingState = true;
		}
	});

	function handleTouchStart(e: TouchEvent) {
		if (containerEl && containerEl.scrollTop <= 0) {
			startY = e.touches[0].clientY;
			isPulling = true;
		}
	}

	function handleTouchMove(e: TouchEvent) {
		if (!isPulling || isRefreshingState) return;
		const currentY = e.touches[0].clientY;
		const diff = currentY - startY;
		if (diff > 0) {
			// Smooth friction: ease out as pull distance increases
			const friction = 0.5 - (diff * 0.001);
			pullDistance = Math.min(diff * Math.max(friction, 0.2), MAX_PULL);
		}
	}

	async function handleTouchEnd() {
		if (!isPulling) return;
		isPulling = false;
		if (pullDistance >= THRESHOLD && !isRefreshingState) {
			isRefreshingState = true;
			try {
				await onRefresh();
			} finally {
				// Delay release for smooth visual
				refreshTimer = setTimeout(() => {
					isRefreshingState = false;
				}, 300);
			}
		}
		pullDistance = 0;
	}

	// Compute arc path for progress circle
	function arcPath(r: number, pct: number): string {
		const circumference = 2 * Math.PI * r;
		const dashLength = circumference * pct;
		return `${dashLength} ${circumference - dashLength}`;
	}
</script>

<div
	bind:this={containerEl}
	class="relative overflow-hidden touch-pan-y"
	ontouchstart={handleTouchStart}
	ontouchmove={handleTouchMove}
	ontouchend={handleTouchEnd}
>
	<!-- Pull indicator -- Talenta style with arc progress -->
	<div
		class="flex items-center justify-center transition-all duration-300 ease-out"
		style="transform: translateY({pullDistance > 0 ? Math.max(pullDistance - 48, 0) : 0}px); opacity: {Math.min(pullDistance / 20, 1)}; height: {pullDistance > 0 ? Math.min(pullDistance, 56) : 0}px"
	>
		{#if isRefreshingState}
			<!-- Spinner -->
			<div class="flex items-center gap-2 text-xs font-medium text-[#1A56DB]">
				<svg class="w-5 h-5 animate-spin" fill="none" viewBox="0 0 24 24">
					<circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
					<path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
				</svg>
				<span>Memuat ulang...</span>
			</div>
		{:else if pullDistance >= THRESHOLD}
			<div class="flex items-center gap-2 text-xs font-semibold text-[#1A56DB]">
				<!-- Arc progress - full -->
				<svg class="w-6 h-6" viewBox="0 0 24 24">
					<circle cx="12" cy="12" r="9" fill="none" stroke="#DBEAFE" stroke-width="2" />
					<circle cx="12" cy="12" r="9" fill="none" stroke="#1A56DB" stroke-width="2"
						stroke-dasharray="56.52 0"
						transform="rotate(-90 12 12)"
						style="transition: stroke-dasharray 0.15s ease"
					/>
				</svg>
				<span>Lepaskan untuk muat ulang</span>
			</div>
		{:else if pullDistance > 5}
			<div class="flex items-center gap-2 text-xs text-gray-400 dark:text-gray-500">
				<!-- Arc progress -->
				<svg class="w-6 h-6" viewBox="0 0 24 24">
					<circle cx="12" cy="12" r="9" fill="none" stroke="#E5E7EB" stroke-width="2" />
					<circle cx="12" cy="12" r="9" fill="none" stroke="#1A56DB" stroke-width="2"
						stroke-dasharray={arcPath(9, progress)}
						stroke-linecap="round"
						transform="rotate(-90 12 12)"
						style="transition: stroke-dasharray 0.1s ease-out"
					/>
					<text x="12" y="12.5" text-anchor="middle" fill="#1A56DB" font-size="7" font-weight="bold">
						{Math.round(progress * 100)}
					</text>
				</svg>
				<span>Tarik ke bawah</span>
			</div>
		{/if}
	</div>

	<!-- Content with spring-back effect -->
	<div
		class="transition-transform duration-300 ease-out"
		style="transform: translateY({isRefreshingState ? 48 : 0}px)"
	>
		{@render children()}
	</div>
</div>
