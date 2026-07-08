<script lang="ts">
	import { fly } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	interface Props {
		/** Each item's unique key for transitions */
		items: unknown[];
		/** Stagger delay in ms between each item */
		stagger?: number;
		/** Fly transition y offset */
		y?: number;
		/** Fly transition duration in ms */
		duration?: number;
		/** Children snippet receives the current item */
		children?: Snippet<[any]>;
	}

	let {
		items = [],
		stagger = 60,
		y = 16,
		duration = 300,
		children,
	}: Props = $props();
</script>

{#if children}
	{#each items as item, i (item)}
		<div
			transition:fly={{ y, duration, delay: i * stagger, opacity: 0 }}
		>
			{@render children(item)}
		</div>
	{/each}
{/if}
