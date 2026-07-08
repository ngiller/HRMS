<script lang="ts">
	interface Props {
		show?: boolean;
		type?: 'fade' | 'slide-up' | 'slide-down' | 'slide-left' | 'slide-right' | 'scale';
		duration?: number;
		delay?: number;
		children?: import('svelte').Snippet;
	}

	let {
		show = true,
		type = 'fade',
		duration = 200,
		delay = 0,
		children,
	}: Props = $props();

	let mounted = $state(false);
	let visible = $state(false);

	$effect(() => {
		if (show) {
			mounted = true;
			// Trigger animation on next frame to allow transition to work from initial state
			requestAnimationFrame(() => {
				requestAnimationFrame(() => {
					visible = true;
				});
			});
		} else {
			visible = false;
			// Wait for exit animation then unmount
			const timer = setTimeout(() => {
				mounted = false;
			}, duration + 50);
			return () => clearTimeout(timer);
		}
	});

	const animations: Record<string, string> = {
		'fade': 'opacity',
		'slide-up': 'opacity, transform',
		'slide-down': 'opacity, transform',
		'slide-left': 'opacity, transform',
		'slide-right': 'opacity, transform',
		'scale': 'opacity, transform',
	};

	const enterTransforms: Record<string, string> = {
		'fade': 'none',
		'slide-up': 'translateY(12px)',
		'slide-down': 'translateY(-12px)',
		'slide-left': 'translateX(12px)',
		'slide-right': 'translateX(-12px)',
		'scale': 'scale(0.95)',
	};
</script>

{#if mounted}
	<div
		class="transition-all ease-out"
		style="
			transition-property: {animations[type] || 'opacity'};
			transition-duration: {duration}ms;
			transition-delay: {delay}ms;
			opacity: {visible ? 1 : 0};
			transform: {visible ? 'none' : (enterTransforms[type] || 'none')};
		"
	>
		{@render children?.()}
	</div>
{/if}

<style>
	div {
		will-change: opacity, transform;
	}
</style>
