<script lang="ts">
	import { fly } from 'svelte/transition';

	type ToastItem = { id: string; message: string; type: 'success' | 'error' | 'warning' | 'info' };

	let {
		toasts: initialToasts = [] as ToastItem[],
		duration = 4000
	}: {
		toasts?: ToastItem[];
		duration?: number;
	} = $props();

	let toasts = $state<ToastItem[]>(initialToasts);

	const typeStyles: Record<string, string> = {
		success: 'bg-green-600 text-white shadow-lg shadow-green-600/30',
		error: 'bg-red-600 text-white shadow-lg shadow-red-600/30',
		warning: 'bg-amber-500 text-white shadow-lg shadow-amber-500/30',
		info: 'bg-[#1A56DB] text-white shadow-lg shadow-blue-600/30'
	};

	const typeIcons: Record<string, string> = {
		success: `<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12.75 11.25 15 15 9.75M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Z"/></svg>`,
		error: `<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z"/></svg>`,
		warning: `<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"/></svg>`,
		info: `<svg class="w-5 h-5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11.25 11.25l.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z"/></svg>`
	};

	function removeToast(id: string) {
		toasts = toasts.filter(t => t.id !== id);
	}

	function addToast(message: string, type: 'success' | 'error' | 'warning' | 'info' = 'info') {
		const id = Math.random().toString(36).slice(2, 9);
		toasts = [...toasts, { id, message, type }];
		setTimeout(() => removeToast(id), duration);
	}

	// Expose addToast globally for easy access
	$effect(() => {
		if (typeof window !== 'undefined') {
			(window as unknown as Record<string, unknown>).__toast = { add: addToast };
		}
	});
</script>

{#if toasts.length > 0}
	<div class="fixed bottom-6 left-1/2 -translate-x-1/2 z-[9999] flex flex-col gap-2 w-[calc(100%-2rem)] max-w-md pointer-events-none">
		{#each toasts as toast, i (toast.id)}
			<div
				transition:fly={{ y: 24, duration: 300, opacity: 0 }}
				class="pointer-events-auto flex items-start gap-3 px-4 py-3.5 rounded-2xl {typeStyles[toast.type]} cursor-pointer shadow-xl"
				role="alert"
				onclick={() => removeToast(toast.id)}
				onkeydown={(e) => e.key === 'Enter' && removeToast(toast.id)}
				tabindex="0"
			>
				{@html typeIcons[toast.type]}
				<p class="text-sm font-medium flex-1 leading-relaxed">{toast.message}</p>
				<button
					class="shrink-0 opacity-60 hover:opacity-100 transition-opacity p-0.5 -mr-1"
					onclick={(e) => { e.stopPropagation(); removeToast(toast.id); }}
					aria-label="Tutup"
				>
					<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
					</svg>
				</button>
			</div>
			<!-- Progress bar -->
			<div class="pointer-events-auto h-1 -mt-1.5 mx-3 rounded-full overflow-hidden bg-white/20 opacity-80">
				<div
					class="toast-progress h-full bg-white/70 rounded-full"
					style="--duration: {duration}ms"
				></div>
			</div>
		{/each}
	</div>
{/if}

<style>
	.toast-progress {
		animation: toast-progress-kf var(--duration) linear forwards;
	}

	@keyframes toast-progress-kf {
		from { width: 100%; }
		to { width: 0%; }
	}
</style>
