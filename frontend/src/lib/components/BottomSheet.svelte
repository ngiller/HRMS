<script lang="ts">
	let {
		open = $bindable(false),
		title = '',
		children,
		footer,
	}: {
		open: boolean;
		title?: string;
		children?: import('svelte').Snippet;
		footer?: import('svelte').Snippet;
	} = $props();

	let startY = $state(0);
	let currentY = $state(0);
	let isDragging = $state(false);
	let sheetEl = $state<HTMLDivElement>(undefined!);

	function close() {
		open = false;
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) close();
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Escape') close();
	}

	function handleTouchStart(e: TouchEvent) {
		// Only enable swipe-down if scrolled to top
		if (sheetEl && sheetEl.scrollTop > 0) return;
		startY = e.touches[0].clientY;
		currentY = startY;
		isDragging = true;
	}

	function handleTouchMove(e: TouchEvent) {
		if (!isDragging) return;
		currentY = e.touches[0].clientY;
		const delta = currentY - startY;
		if (delta > 0) {
			// Apply friction: translate sheet by delta with diminishing effect
			const translate = Math.min(delta * 0.5, 200);
			if (sheetEl) {
				sheetEl.style.transform = `translateY(${translate}px)`;
				sheetEl.style.transition = 'none';
			}
		}
	}

	function handleTouchEnd(_e: TouchEvent) {
		if (!isDragging) return;
		isDragging = false;
		const delta = currentY - startY;
		if (delta > 80) {
			// Swipe down past threshold — close
			if (sheetEl) {
				sheetEl.style.transition = 'transform 0.2s ease-out';
				sheetEl.style.transform = 'translateY(100%)';
			}
			setTimeout(() => { close(); }, 150);
		} else {
			// Snap back
			if (sheetEl) {
				sheetEl.style.transition = 'transform 0.25s cubic-bezier(0.16, 1, 0.3, 1)';
				sheetEl.style.transform = 'translateY(0)';
			}
		}
	}

	$effect(() => {
		if (open) {
			document.body.style.overflow = 'hidden';
		} else {
			document.body.style.overflow = '';
		}
		return () => { document.body.style.overflow = ''; };
	});

	// Reset transform when sheet opens
	$effect(() => {
		if (open && sheetEl) {
			sheetEl.style.transform = 'translateY(0)';
		}
	});
</script>

{#if open}
	<div
		class="fixed inset-0 z-50 flex items-end justify-center"
		onclick={handleBackdropClick}
		onkeydown={handleKeydown}
		role="presentation"
	>
		<!-- Backdrop -->
		<div
			class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity duration-300"
			class:opacity-100={open}
		></div>

		<!-- Sheet -->
		<div
			bind:this={sheetEl}
			role="dialog"
			aria-modal="true"
			aria-label={title || 'Bottom sheet'}
			ontouchstart={handleTouchStart}
			ontouchmove={handleTouchMove}
			ontouchend={handleTouchEnd}
			class="relative w-full max-h-[90dvh] bg-white dark:bg-gray-900 rounded-t-2xl shadow-2xl flex flex-col"
			class:animate-slide-up={!isDragging}
		>
			<!-- Handle bar (drag indicator) -->
			<div class="flex justify-center pt-2 pb-1 shrink-0 cursor-grab active:cursor-grabbing touch-none">
				<div class="w-10 h-1 rounded-full bg-gray-300 dark:bg-gray-600"></div>
			</div>

			<!-- Header -->
			{#if title}
				<div class="flex items-center justify-between px-5 py-3 border-b border-gray-100 dark:border-gray-800 shrink-0">
					<h2 class="text-base font-semibold text-gray-900 dark:text-white">{title}</h2>
					<button
						onclick={close}
						class="p-1.5 rounded-lg text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 transition cursor-pointer"
						aria-label="Tutup"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
						</svg>
					</button>
				</div>
			{/if}

			<!-- Body (scrollable) -->
			<div class="flex-1 overflow-y-auto px-5 py-4">
				{@render children?.()}
			</div>

			<!-- Footer (sticky) -->
			{#if footer}
				<div class="shrink-0 px-5 py-4 border-t border-gray-100 dark:border-gray-800 bg-gray-50/50 dark:bg-gray-800/50">
					{@render footer?.()}
				</div>
			{/if}
		</div>
	</div>
{/if}

<style>
	@keyframes slide-up {
		from { transform: translateY(100%); }
		to { transform: translateY(0); }
	}
	.animate-slide-up {
		animation: slide-up 0.3s cubic-bezier(0.16, 1, 0.3, 1);
	}
</style>
