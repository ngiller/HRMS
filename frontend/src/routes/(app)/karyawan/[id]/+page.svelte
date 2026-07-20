<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';

	onMount(async () => {
// eslint-disable-next-line svelte/no-navigation-without-resolve
		await goto('/karyawan', { replaceState: true });
	});
</script>
