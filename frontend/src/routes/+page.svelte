<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';

	onMount(async () => {
// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto('/login', { replaceState: true });
	});
</script>
