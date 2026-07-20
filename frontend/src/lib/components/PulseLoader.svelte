<script lang="ts">
	interface Props {
		variant?: 'card' | 'text' | 'table-row' | 'custom';
		count?: number;
		height?: string;
		width?: string;
		rounded?: string;
		containerClass?: string;
	}

	let props: Props = $props();
	let variant = $derived(props.variant ?? 'text');
	let count = $derived(props.count ?? 1);
	let height = $derived(props.height ?? 'h-4');
	let width = $derived(props.width ?? 'w-full');
	let rounded = $derived(props.rounded ?? 'rounded-lg');
	let containerClass = $derived(props.containerClass ?? '');
</script>

{#each {length: count} as _, i (i)}
	{#if variant === 'card'}
		<div class="bg-white dark:bg-gray-900 rounded-xl border border-gray-200 dark:border-gray-800 p-4 overflow-hidden {containerClass}">
			<div class="flex items-center gap-3 mb-4">
				<div class="w-10 h-10 rounded-full shimmer shrink-0"></div>
				<div class="flex-1 space-y-2">
					<div class="h-3 shimmer rounded w-3/4"></div>
					<div class="h-2.5 shimmer rounded w-1/2"></div>
				</div>
			</div>
			<div class="space-y-2">
				<div class="h-2.5 shimmer rounded w-full"></div>
				<div class="h-2.5 shimmer rounded w-2/3"></div>
			</div>
			<div class="flex items-center gap-2 mt-4 pt-3 border-t border-gray-100 dark:border-gray-800">
				<div class="h-8 shimmer rounded-lg flex-1"></div>
				<div class="h-8 shimmer rounded-lg flex-1"></div>
			</div>
		</div>
	{:else if variant === 'table-row'}
		<div class="flex items-center gap-4 py-3 px-4 {containerClass}">
			<div class="w-9 h-9 rounded-lg shimmer shrink-0"></div>
			<div class="flex-1 space-y-1.5">
				<div class="h-3 shimmer rounded w-44"></div>
				<div class="h-2.5 shimmer rounded w-28"></div>
			</div>
			<div class="h-6 w-20 shimmer rounded-full shrink-0"></div>
			<div class="h-8 w-24 shimmer rounded-lg shrink-0"></div>
		</div>
	{:else if variant === 'text'}
		<div class="space-y-2 {containerClass}">
			<div class="h-3 shimmer rounded w-full"></div>
			<div class="h-3 shimmer rounded w-5/6"></div>
			<div class="h-3 shimmer rounded w-2/3"></div>
		</div>
	{:else}
		<div class="{height} {width} {rounded} shimmer {containerClass}"></div>
	{/if}
{/each}


