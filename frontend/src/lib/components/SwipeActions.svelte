<script lang="ts">
	interface Props {
		onApprove?: () => void;
		onReject?: () => void;
		approveLabel?: string;
		rejectLabel?: string;
		children: import('svelte').Snippet;
	}

	let {
		onApprove,
		onReject,
		approveLabel = 'Setujui',
		rejectLabel = 'Tolak',
		children,
	}: Props = $props();

	let translateX = $state(0);
	let isSwiping = $state(false);
	let startX = 0;
	let showSwipeActions = $state(false);
	let isAnimating = $state(false);
	let pointerEventsEnabled = $state(true);

	function handleTouchStart(e: TouchEvent) {
		if (isAnimating || !pointerEventsEnabled) return;
		startX = e.touches[0].clientX;
		isSwiping = true;
	}

	function handleTouchMove(e: TouchEvent) {
		if (!isSwiping || isAnimating || !pointerEventsEnabled) return;
		const currentX = e.touches[0].clientX;
		const deltaX = currentX - startX;

		if (showSwipeActions) {
			if (deltaX > 0) {
				translateX = Math.min(0, -80 + deltaX);
			} else {
				translateX = Math.max(-120, -80 + deltaX * 0.3);
			}
			return;
		}

		translateX = Math.max(-80, Math.min(80, deltaX * 0.8));
	}

	function handleTouchEnd() {
		if (!isSwiping) return;
		isSwiping = false;

		if (showSwipeActions) {
			if (translateX > -30) {
				hideActions();
			} else {
				snapToActions();
			}
			return;
		}

		if (translateX > 40 && onApprove) {
			// Swiped right — approve
			blockClicks();
			translateX = 0;
			showSwipeActions = false;
			setTimeout(() => {
				unblockClicks();
				onApprove?.();
			}, 200);
		} else if (translateX < -40 && onReject) {
			// Swiped left — show reject button
			snapToActions();
		} else {
			translateX = 0;
		}
	}

	function snapToActions() {
		showSwipeActions = true;
		translateX = -80;
	}

	function hideActions() {
		isAnimating = true;
		showSwipeActions = false;
		translateX = 0;
		setTimeout(() => {
			isAnimating = false;
		}, 200);
	}

	function blockClicks() {
		isAnimating = true;
		pointerEventsEnabled = false;
	}

	function unblockClicks() {
		isAnimating = false;
		pointerEventsEnabled = true;
	}

	function handleApproveClick(e: MouseEvent) {
		e.stopPropagation();
		blockClicks();
		translateX = 0;
		setTimeout(() => {
			unblockClicks();
			if (onApprove) onApprove();
		}, 200);
	}

	function handleRejectClick(e: MouseEvent) {
		e.stopPropagation();
		blockClicks();
		translateX = 0;
		setTimeout(() => {
			unblockClicks();
			if (onReject) onReject();
		}, 200);
	}
</script>

<div class="relative overflow-hidden rounded-xl">
	<!-- Approve action behind the card (left side) -->
	<div class="absolute inset-y-0 left-0 flex items-center" style="pointer-events: none;">
		{#if onApprove}
			<button
				onclick={handleApproveClick}
				class="h-full px-4 bg-emerald-500 text-white text-xs font-semibold flex items-center gap-1.5 transition-colors hover:bg-emerald-600 cursor-pointer active:scale-95"
				style="border-radius: 0.75rem 0 0 0.75rem; pointer-events: auto;"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" />
				</svg>
				{approveLabel}
			</button>
		{/if}
	</div>

	<!-- Reject action behind the card (right side) -->
	<div class="absolute inset-y-0 right-0 flex items-center" style="pointer-events: none;">
		{#if onReject}
			<button
				onclick={handleRejectClick}
				class="h-full px-4 bg-red-500 text-white text-xs font-semibold flex items-center gap-1.5 transition-colors hover:bg-red-600 cursor-pointer active:scale-95"
				style="border-radius: 0 0.75rem 0.75rem 0; pointer-events: auto;"
			>
				<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
				</svg>
				{rejectLabel}
			</button>
		{/if}
	</div>

	<!-- Card content -->
	<div
		class="relative bg-white dark:bg-gray-900 z-10"
		class:touch-pan-y={pointerEventsEnabled}
		style="transform: translateX({translateX}px); transition: {isAnimating ? 'transform 0.2s ease-out' : isSwiping ? 'none' : 'transform 0.2s ease-out'}; pointer-events: {pointerEventsEnabled ? 'auto' : 'none'};"
		ontouchstart={handleTouchStart}
		ontouchmove={handleTouchMove}
		ontouchend={handleTouchEnd}
	>
		{@render children()}
	</div>
</div>
