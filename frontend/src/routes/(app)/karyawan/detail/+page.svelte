<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/stores';
	import EmployeeDetail from '$lib/components/EmployeeDetail.svelte';

	let state = $derived($page.state as Record<string, unknown>);
	let empId = $derived(typeof state?.employeeId === 'string' ? state.employeeId as string : '');
	let initialTab = $derived(typeof state?.initialTab === 'string' ? state.initialTab as 'profile' | 'salary' | 'bpjs' | 'tax' | 'overtime' : 'profile');

	$effect(() => {
		if (!empId) {
// eslint-disable-next-line svelte/no-navigation-without-resolve
			goto('/karyawan', { replaceState: true });
		}
	});
</script>

{#if empId}
	<div class="w-full">
<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -->
		<EmployeeDetail employeeId={empId} {initialTab} onclose={async () => await goto('/karyawan')} />
	</div>
{:else}
	<div class="py-16 text-center text-gray-400 text-sm">Redirecting...</div>
{/if}
