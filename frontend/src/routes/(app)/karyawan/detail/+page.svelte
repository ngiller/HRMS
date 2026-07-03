<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import EmployeeDetail from '$lib/components/EmployeeDetail.svelte';

	let state = $derived($page.state as Record<string, unknown>);
	let empId = $derived(typeof state?.employeeId === 'string' ? state.employeeId as string : '');

	$effect(() => {
		if (!empId) {
			goto('/karyawan', { replaceState: true });
		}
	});
</script>

{#if empId}
	<div class="w-full">
		<EmployeeDetail employeeId={empId} onclose={() => goto('/karyawan')} />
	</div>
{:else}
	<div class="py-16 text-center text-gray-400 text-sm">Redirecting...</div>
{/if}
